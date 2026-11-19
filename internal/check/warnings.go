package check

import (
	"fmt"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/codes"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// This file reports warnings: code the checker accepts that is almost always a mistake
// (spec/check.md section 4). Each is silenced by writing the intent, never by a comment.

// warnf reports a warning at pos and returns its index in c.warns, for warnHelp and warnFix.
func (c *Checker) warnf(code codes.Code, pos, end ast.Pos, format string, args ...any) int {
	c.warns = append(c.warns, Diagnostic{Path: c.mod.path, Pos: pos, End: end, Msg: fmt.Sprintf(format, args...), Code: code, Warning: true})
	return len(c.warns) - 1
}

func (c *Checker) warnHelp(i int, format string, args ...any) {
	c.warns[i].Help = fmt.Sprintf(format, args...)
}

func (c *Checker) warnFix(i int, title string, auto bool, edits ...loader.Edit) {
	for k := range edits {
		if edits[k].Path == "" {
			edits[k].Path = c.mod.path
		}
	}
	c.warns[i].Fixes = append(c.warns[i].Fixes, loader.Fix{Title: title, Auto: auto, Edits: edits})
}

// discard marks b as a block whose value the language discards: the body of a function or
// lambda that returns void, of a test, a loop, or a go block.
func (c *Checker) discard(b *ast.Block) {
	if b == nil {
		return
	}
	if c.discarded == nil {
		c.discarded = map[*ast.Block]bool{}
	}
	c.discarded[b] = true
}

// discardBranches marks the blocks of an if or select whose value is discarded, so that what
// their last statements leave is discarded too.
func (c *Checker) discardBranches(e ast.Node) {
	switch x := e.(type) {
	case *ast.If:
		c.discard(x.Then)
		c.discardBranches(x.Else)
	case *ast.Block:
		c.discard(x)
	case *ast.Select:
		for _, cl := range x.Clauses {
			c.discard(cl.Body)
		}
	}
}

// droppedResult warns when the statement s, whose value is discarded, is a Result or a union
// with one: its Err would be lost (unused-result). A deferred block's results are discarded on
// purpose (section 8), so nothing in one is reported.
func (c *Checker) droppedResult(s ast.Expr, t types.Type) {
	if c.deferred > 0 || !c.isResultish(t) {
		return
	}
	switch s.(type) {
	case *ast.If, *ast.Select, *ast.Block:
		return // its branches are reported where they end
	}
	what := "this expression"
	if call, ok := s.(*ast.Call); ok {
		if name := calleeName(call); name != "" {
			what = name
		}
	}
	i := c.warnf(codes.UnusedResult, ast.Start(s), s.EndPosition(), "the Result of %s is discarded, so its Err is lost", what)
	c.warnHelp(i, "propagate it with try, or discard it on purpose with _ = ...")
	c.warnFix(i, "_ = ...", false, loader.Edit{Start: ast.Start(s), End: ast.Start(s), Text: "_ = "})
}

// isResultish reports whether t is a Result, or a union with a Result among its members.
func (c *Checker) isResultish(t types.Type) bool {
	if t == nil || types.IsInvalid(t) {
		return false
	}
	ok, err := c.prelude["Ok"], c.prelude["Err"]
	for _, m := range types.Members(types.Expand(t)) {
		if n, isNamed := types.Expand(m).(types.Named); isNamed && (n.Decl == ok.decl || n.Decl == err.decl) {
			return true
		}
	}
	return false
}

// calleeName is how a warning names the function a call calls: f, Ns.f, or x.f.
func calleeName(x *ast.Call) string {
	switch f := x.Fn.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.Field:
		if base := pathKey(f.X); base != "" {
			return base + "." + f.Name
		}
		return f.Name
	}
	return ""
}

// namedInterpolation warns when a part of a string template is a named primitive type, which
// renders as its constructor, UserId("u1"), not as the value (named-interpolation). Markup
// rejects one for that reason (section 5); a string is left to say what it means.
func (c *Checker) namedInterpolation(e ast.Expr, t types.Type) {
	n, ok := types.Expand(t).(types.Named)
	if !ok || n.Decl.Kind != types.PrimitiveDecl {
		return
	}
	conv := baseConversion(n.Decl.Base)
	i := c.warnf(codes.NamedInterpolation, ast.Start(e), e.EndPosition(), "%s is a %s, which a string shows as %s(...), not as its value", exprLabel(e), n.Decl.Name, n.Decl.Name)
	c.warnHelp(i, "write ${%s(...)} for the value, or ${String.from(...)} for the %s(...) form", conv, n.Decl.Name)
	c.warnFix(i, conv+"(...)", false,
		loader.Edit{Start: ast.Start(e), End: ast.Start(e), Text: conv + "("},
		loader.Edit{Start: e.EndPosition(), End: e.EndPosition(), Text: ")"})
}

// baseConversion is the conversion that unwraps a named primitive over k.
func baseConversion(k value.Kind) string {
	return k.String()
}

// exprLabel is e in a warning: its path of names, or "the value".
func exprLabel(e ast.Expr) string {
	if p := pathKey(e); p != "" {
		return p
	}
	return "the value"
}

// warnUnused reports, in a module that parsed, an imported name the module never uses
// (unused-import) and a let or var in a block that nothing reads (unused-variable), from what
// the resolver found each name refers to.
func (c *Checker) warnUnused(mi *moduleInfo) {
	info := mi.m.Resolve
	if info == nil {
		return
	}
	c.mod = mi
	f := mi.m.File
	for _, imp := range f.Imports {
		if imp.Namespace != "" {
			if len(info.Uses[imp]) == 0 && !mi.typeRefs[imp.Namespace] {
				i := c.warnf(codes.UnusedImport, imp.Pos, imp.End, "%s is imported but never used", imp.Namespace)
				c.warnHelp(i, "remove the import")
			}
			continue
		}
		for j := range imp.Names {
			n := &imp.Names[j]
			if len(info.Uses[n]) > 0 || mi.typeRefs[n.Alias] {
				continue
			}
			i := c.warnf(codes.UnusedImport, n.Pos, ast.Pos{}, "%s is imported but never used", n.Alias)
			c.removeImportName(i, imp, j)
		}
	}
	ast.Inspect(f, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.FuncDecl, *ast.FuncLit, *ast.Lambda, *ast.Test, *ast.Block:
			_ = x
			return true
		case *ast.VarDecl:
			if c.topLevel(f, x) {
				return true
			}
			ast.Inspect(x.Target, func(t ast.Node) bool {
				if nt, ok := t.(*ast.NameTarget); ok && nt.Name != "_" && len(info.Uses[nt]) == 0 {
					kind := "var"
					if x.Let {
						kind = "let"
					}
					i := c.warnf(codes.UnusedVariable, nt.Pos, ast.Pos{}, "%s is declared with %s but never used", nt.Name, kind)
					c.warnHelp(i, "remove it, or bind _ to keep what its initializer does")
				}
				return true
			})
		}
		return true
	})
}

// topLevel reports whether d is a declaration of f's top level, which a later module or the
// script itself may use, and so is not an unused local.
func (c *Checker) topLevel(f *ast.File, d *ast.VarDecl) bool {
	for _, x := range f.Decls {
		if x == d {
			return true
		}
	}
	for _, s := range f.Stmts {
		if s == d {
			return true
		}
	}
	return false
}

// removeImportName adds to the warning at i the fix that removes the j-th name of imp with the
// comma beside it. It is auto when another name of the import remains, so that the module stays
// imported and its initializers still run; removing the last name removes the whole import.
func (c *Checker) removeImportName(i int, imp *ast.Import, j int) {
	src := c.mod.m.Src
	if src == "" {
		return
	}
	n := imp.Names[j]
	start, end := c.offset(n.Pos), c.offset(n.End)
	if start < 0 || end < 0 {
		return
	}
	if len(imp.Names) == 1 {
		c.warnHelp(i, "remove the import")
		return
	}
	c.warnHelp(i, "remove %s from the import", n.Alias)
	// take the comma and the space after the name, or, for the last name, before it
	if after := strings.TrimLeft(src[end:], " \t"); strings.HasPrefix(after, ",") {
		end = len(src) - len(after) + 1
		end += len(src[end:]) - len(strings.TrimLeft(src[end:], " \t"))
	} else {
		before := strings.TrimRight(src[:start], " \t")
		if !strings.HasSuffix(before, ",") {
			return
		}
		start = len(before) - 1
	}
	c.warnFix(i, "remove "+n.Alias, true, loader.Edit{Start: c.pos(start), End: c.pos(end), Text: ""})
}

// offset is the byte offset of pos in the current module's source, or -1.
func (c *Checker) offset(pos ast.Pos) int {
	src := c.mod.m.Src
	off := 0
	for line := 1; line < pos.Line; line++ {
		nl := strings.IndexByte(src[off:], '\n')
		if nl < 0 {
			return -1
		}
		off += nl + 1
	}
	if off+pos.Col-1 > len(src) {
		return -1
	}
	return off + pos.Col - 1
}

// pos is the position of the byte at off in the current module's source.
func (c *Checker) pos(off int) ast.Pos {
	src := c.mod.m.Src[:off]
	line := strings.Count(src, "\n") + 1
	return ast.Pos{Line: line, Col: off - (strings.LastIndexByte(src, '\n') + 1) + 1}
}
