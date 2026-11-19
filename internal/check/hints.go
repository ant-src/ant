package check

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/near"
	"github.com/ant-src/ant/internal/types"
)

// This file adds help and fixes to errors whose repair the checker can name (spec/check.md
// section 2, Help and fixes).

// receiverHint adds to the error at i, that recv has no function f.Name to call as recv.f(...),
// the prelude functions of that name that take recv first. With one, the help names it and the
// fix calls it through its namespace, as String.join(parts, ", ") for parts.join(", "); else
// the help offers the names of recv's own members nearest to f.Name.
func (c *Checker) receiverHint(i int, recv types.Type, x *ast.Call, f *ast.Field, own []string) {
	if i < 0 {
		return
	}
	var found []string
	for name, ns := range c.prelude {
		if ns.kind != entNamespace {
			continue
		}
		if m, ok := ns.members[f.Name]; ok && m.kind == entFunc && m.sig != nil && c.accepts(m.sig, recv) && takes(m.sig, x) {
			found = append(found, name)
		}
	}
	slices.Sort(found)
	switch {
	case len(found) == 1:
		ns := found[0]
		c.help(i, "%s is %s.%s, which takes the %s first; call it through its namespace", f.Name, ns, f.Name, describe(recv))
		c.callThrough(i, ns, x, f)
	case len(found) > 1:
		var names []string
		for _, ns := range found {
			names = append(names, ns+"."+f.Name)
		}
		c.help(i, "%s takes the %s first; call one through its namespace", strings.Join(names, " or "), describe(recv))
		for _, ns := range found {
			c.callThrough(i, ns, x, f)
		}
	default:
		c.nearHint(i, f.Name, own, "")
	}
}

// callThrough adds to the error at i the fix that rewrites recv.name(args) as ns.name(recv,
// args), when the source shows the . right after recv and the call's ( right after the name:
// recv is left as written, so the fix inserts the namespace before it and turns .name( into a
// comma, or into nothing when the call has no arguments. A receiver in parentheses, (xs).name,
// gets no fix, since its start is inside them. It is auto when it is the only fix.
func (c *Checker) callThrough(i int, ns string, x *ast.Call, f *ast.Field) {
	paren := ast.Pos{Line: f.NamePos.Line, Col: f.NamePos.Col + len(f.Name)}
	if f.Optional || x == nil || f.X.EndPosition() != f.Pos || c.srcByte(paren) != '(' {
		return
	}
	sep, rest := ", ", ", ..."
	if len(x.Args) == 0 {
		sep, rest = "", ""
	}
	call := ns + "." + f.Name
	c.fix(i, fmt.Sprintf("%s(%s%s)", call, exprText(f.X), rest), false,
		loader.Edit{Start: ast.Start(f.X), End: ast.Start(f.X), Text: call + "("},
		loader.Edit{Start: f.Pos, End: ast.Pos{Line: paren.Line, Col: paren.Col + 1}, Text: sep})
	if fixes := c.diags[i].Fixes; len(fixes) == 1 {
		fixes[0].Auto = true
	} else {
		for k := range fixes {
			fixes[k].Auto = false
		}
	}
}

// nearHint adds to the error at i the names among candidates nearest to name, as a misspelling
// would be, or else says where the list is: the namespace's documentation, when listed is set.
func (c *Checker) nearHint(i int, name string, candidates []string, listed string) {
	if i < 0 {
		return
	}
	if near := near.Nearest(name, candidates); len(near) > 0 {
		c.help(i, "did you mean %s?", strings.Join(near, " or "))
		return
	}
	if listed != "" {
		c.help(i, "ant doc %s lists its members", listed)
	}
}

// describe names a type for a help line: the type itself, or "array" and the like for the
// built-in kinds a reader thinks of by name.
func describe(t types.Type) string {
	switch types.Expand(t).(type) {
	case types.Array:
		return "array"
	case types.Map:
		return "map"
	case types.Set:
		return "set"
	}
	return types.String(t)
}

// srcByte is the byte of the current module's source at pos, or 0 when the source is not kept
// (loader.Config Recover) or pos is outside it.
func (c *Checker) srcByte(pos ast.Pos) byte {
	if c.mod == nil || c.mod.m == nil || c.mod.m.Src == "" || pos.Line < 1 || pos.Col < 1 {
		return 0
	}
	src := c.mod.m.Src
	for line := 1; line < pos.Line; line++ {
		nl := strings.IndexByte(src, '\n')
		if nl < 0 {
			return 0
		}
		src = src[nl+1:]
	}
	if pos.Col-1 >= len(src) {
		return 0
	}
	return src[pos.Col-1]
}

// exprText is a short label for e in a fix's title: its name when it is one, else "...".
func exprText(e ast.Expr) string {
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return "..."
}

// memberNames is what recv.name may name: the fields of recv's record type and the functions of
// its namespace ns, for nearHint.
func (c *Checker) memberNames(recv types.Type, ns *entity) []string {
	var out []string
	if r, ok := types.Expand(recv).(types.Record); ok {
		for _, f := range r.Fields {
			out = append(out, f.Name)
		}
	} else if n, ok := types.Expand(recv).(types.Named); ok && n.Decl.Kind == types.RecordDecl {
		for _, f := range n.Decl.Fields {
			out = append(out, f.Name)
		}
	}
	if ns != nil {
		for name, m := range ns.members {
			if m.kind == entFunc {
				out = append(out, name)
			}
		}
	}
	return out
}

// takes reports whether s takes the receiver and the arguments of the call x, by count.
func takes(s *sig, x *ast.Call) bool {
	n := 1
	if x != nil {
		n += len(x.Args)
	}
	return n >= s.fn.Required && n <= len(s.fn.Params)
}

// visibleNames is every value name visible from sc: its locals, the current module's names, and
// the prelude's, for nearHint.
func (c *Checker) visibleNames(sc *scope) []string {
	var out []string
	for s := sc; s != nil; s = s.parent {
		for name := range s.names {
			out = append(out, name)
		}
	}
	if c.mod != nil {
		for name := range c.mod.names {
			out = append(out, name)
		}
	}
	for name := range c.prelude {
		out = append(out, name)
	}
	return out
}

// okVoidHint adds to the error at i, that Result.ok() has no value, that a Result[void, E] holds
// void, with the fix that passes it. The fix is not auto: the author may have left out a value.
func (c *Checker) okVoidHint(i int, x *ast.Call) {
	c.help(i, "Result.ok takes the value; the value of a Result[void, E] is void, as in Result.ok(void)")
	paren := ast.Pos{Line: x.End.Line, Col: x.End.Col - 1}
	if c.srcByte(paren) == ')' {
		c.fix(i, "Result.ok(void)", false, loader.Edit{Start: paren, End: paren, Text: "void"})
	}
}

// pathTest is a field path, such as r.error, tested with is in the function fn.
type pathTest struct {
	fn   *fnCtx
	path string
}

// pathKey is e written as a path of names, such as "json.value", or "" when e is not one.
func pathKey(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.Field:
		if x.Optional {
			return ""
		}
		if base := pathKey(x.X); base != "" {
			return base + "." + x.Name
		}
	}
	return ""
}

// notePathTest records that an is test at pos tested e, when e is a field path: a test narrows
// only a local or a parameter (section 7), so a later error about e may come from expecting it to.
func (c *Checker) notePathTest(e ast.Expr, pos ast.Pos) {
	key := pathKey(e)
	if _, isField := e.(*ast.Field); !isField || key == "" {
		return
	}
	if c.pathTests == nil {
		c.pathTests = map[pathTest]ast.Pos{}
	}
	k := pathTest{c.fn, key}
	if _, seen := c.pathTests[k]; !seen {
		c.pathTests[k] = pos
	}
}

// pathHint adds to the error at i, about the field path e, that an is test of e earlier in the
// function did not narrow it, and where that test is.
func (c *Checker) pathHint(i int, e ast.Expr) {
	key := pathKey(e)
	if i < 0 || key == "" {
		return
	}
	pos, ok := c.pathTests[pathTest{c.fn, key}]
	if !ok || !pos.Before(c.diags[i].Pos) {
		return
	}
	c.help(i, "an is test narrows a local or a parameter, not the field path %s; bind it first, as let v = %s, and test v", key, key)
	c.related(i, pos, "the test of "+key)
}

// pathHintSince adds pathHint to the first error reported since before, the number of errors
// there were: the caller checked one use of e in between.
func (c *Checker) pathHintSince(before int, e ast.Expr) {
	if before < len(c.diags) {
		c.pathHint(before, e)
	}
}

// coverHint adds to the error at i, an if without else whose condition is an is test of a local,
// the members of its type that the chain of tests leaves out: those the local still has when the
// test is false.
func (c *Checker) coverHint(i int, cond ast.Expr, sc *scope, whenFalse flow) {
	test, ok := cond.(*ast.Is)
	if !ok || i < 0 {
		return
	}
	id, ok := test.X.(*ast.Ident)
	if !ok {
		return
	}
	l := c.narrowable(sc, id)
	if l == nil {
		return
	}
	left, ok := whenFalse[l]
	if !ok || types.IsNever(left) || types.IsInvalid(left) {
		return
	}
	var names []string
	for _, m := range types.Members(left) {
		names = append(names, types.String(m))
	}
	if len(names) == 1 {
		c.help(i, "the tests leave out %s: add else if %s is %s { ... }, or an else", names[0], id.Name, names[0])
		return
	}
	c.help(i, "the tests leave out %s: test each, or add an else", strings.Join(names, ", "))
}
