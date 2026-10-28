package lsp

import (
	"regexp"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

var (
	missingField   = regexp.MustCompile(`^(\S+) literal is missing the field (\S+)$`)
	stateReturn    = regexp.MustCompile(`^the exported function (\S+) must state its return type$`)
	stateVarType   = regexp.MustCompile(`^(\S+) is used by a function in its own initializer; state its type on the var$`)
	annotateResult = regexp.MustCompile(`annotate the result$`)
)

// codeActions answers textDocument/codeAction with the mechanical fixes for the diagnostics in
// the range (spec/lsp.md section 4): stating a type the checker inferred where a message asks
// for an annotation, and adding the fields a record literal leaves out.
func (s *Server) codeActions(p *CodeActionParams) []CodeAction {
	d := s.doc(p.TextDocument.URI)
	if d == nil {
		return nil
	}
	a := s.analyze(d)
	path := a.module().File.Path
	out := []CodeAction{}
	for _, dg := range a.diags {
		if dg.Path != path {
			continue
		}
		rng := a.diagRange(path, dg.Pos)
		if rng.End.Line < p.Range.Start.Line || rng.Start.Line > p.Range.End.Line {
			continue
		}
		lsp := Diagnostic{Range: rng, Severity: 1, Source: "ant", Message: dg.Msg}
		switch {
		case missingField.MatchString(dg.Msg):
			if act := s.addFields(a, d, dg.Pos, lsp); act != nil {
				out = append(out, *act)
			}
		case stateReturn.MatchString(dg.Msg):
			if act := s.stateReturnType(a, d, dg.Pos, lsp); act != nil {
				out = append(out, *act)
			}
		case stateVarType.MatchString(dg.Msg) || annotateResult.MatchString(dg.Msg):
			if act := s.stateVarType(a, d, dg.Pos, lsp); act != nil {
				out = append(out, *act)
			}
		}
	}
	return dedupeActions(out)
}

func dedupeActions(acts []CodeAction) []CodeAction {
	seen := map[string]bool{}
	out := acts[:0]
	for _, act := range acts {
		if !seen[act.Title] {
			seen[act.Title] = true
			out = append(out, act)
		}
	}
	return out
}

// addFields adds every field a record literal at pos leaves out, each with the zero value of
// its type as a placeholder.
func (s *Server) addFields(a *analysis, d *document, pos ast.Pos, dg Diagnostic) *CodeAction {
	var lit *ast.RecordLit
	for _, n := range ast.Path(a.module().File, pos) {
		if x, ok := n.(*ast.RecordLit); ok && ast.Start(x) == pos {
			lit = x
		}
	}
	if lit == nil {
		return nil
	}
	named, ok := types.Expand(a.info.Types[lit]).(types.Named)
	if !ok || named.Decl.Kind != types.RecordDecl {
		return nil
	}
	given := map[string]bool{}
	for _, e := range lit.Entries {
		if e.Name != "" {
			given[e.Name] = true
		}
	}
	var parts []string
	for _, f := range named.FieldsOf() {
		if !given[f.Name] && !f.Default {
			value, ok := a.zeroValue(d.path, f.Type)
			if !ok {
				return nil
			}
			parts = append(parts, f.Name+": "+value)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	text := strings.Join(parts, ", ")
	var insertAt int
	if len(lit.Entries) == 0 {
		// an empty literal: Point{} becomes Point{ x: 0.0 }, whatever stands between the braces
		open := d.offsetOf(lit.Head.EndPosition())
		for open < len(d.text) && d.text[open] != '{' {
			open++
		}
		if open >= len(d.text) {
			return nil
		}
		insertAt = open + 1
		text = " " + text
		if insertAt < len(d.text) && d.text[insertAt] == '}' {
			text += " "
		}
	} else {
		// after the last entry, before a comment or line break that follows it, keeping a
		// trailing comma last
		last := &lit.Entries[len(lit.Entries)-1]
		insertAt = d.offsetOf(last.End)
		comma := insertAt
		for comma < len(d.text) && (d.text[comma] == ' ' || d.text[comma] == '\t') {
			comma++
		}
		if comma < len(d.text) && d.text[comma] == ',' {
			insertAt = comma + 1
			text = " " + text + ","
		} else {
			text = ", " + text
		}
	}
	title := "Add the missing field " + parts[0][:strings.Index(parts[0], ":")]
	if len(parts) > 1 {
		title = "Add the missing fields"
	}
	pos2 := d.position(insertAt, s.utf16)
	return &CodeAction{Title: title, Kind: "quickfix", Diagnostics: []Diagnostic{dg}, IsPreferred: true,
		Edit: &WorkspaceEdit{Changes: map[string][]TextEdit{d.uri: {{Range: Range{Start: pos2, End: pos2}, NewText: text}}}}}
}

// zeroValue is a placeholder value of a type, written with the names the module at path has:
// its zero, or null for a nullable type; false when the type has a part the checker could not
// determine or a type the module cannot name.
func (a *analysis) zeroValue(path string, t types.Type) (string, bool) {
	if mentionsInvalid(t) {
		return "", false
	}
	if types.Includes(t, types.Null{}) {
		return "null", true
	}
	switch x := types.Expand(t).(type) {
	case types.Prim:
		switch {
		case x.Kind.IsFloat():
			return "0.0", true
		case x.Kind.IsNumber():
			return "0", true
		case x.Kind == value.KindString:
			return `""`, true
		case x.Kind == value.KindBool:
			return "false", true
		}
	case types.Array:
		return "[]", true
	case types.Map:
		return "Map{}", true
	case types.Named:
		if x.Decl.Kind == types.RecordDecl {
			name, ok := a.localTypeName(path, x.Decl)
			if !ok {
				return "", false
			}
			return name + "{}", true
		}
	case types.Func:
		return "{ }", true
	case types.Void:
		return "void", true
	}
	return "null", true
}

// mentionsInvalid reports whether any part of t is a type the checker could not determine.
func mentionsInvalid(t types.Type) bool {
	found := false
	types.Walk(t, func(x types.Type) {
		if types.IsInvalid(x) {
			found = true
		}
	})
	return found
}

// localTypeName is how the module at path names a declared type: by its own name for its own
// and the prelude's types, else by the name an import gives it, or Namespace.Name through a
// namespace import; false when the module has no name for it.
func (a *analysis) localTypeName(path string, decl *types.Decl) (string, bool) {
	if decl.Path == path || decl.Path == "prelude.ant" || decl.Path == "" {
		return decl.Name, true
	}
	m := a.modules[path]
	if m == nil || m.Resolve == nil {
		return "", false
	}
	sc := m.Resolve.Scopes[m.File]
	if sc == nil {
		return "", false
	}
	for _, b := range sc.Names {
		switch n := b.Node.(type) {
		case *ast.ImportName:
			if o := a.info.Decls[n]; o != nil && o.TypeDecl() == decl {
				return b.Name, true
			}
		case *ast.Import:
			if o := a.info.Decls[n]; o != nil {
				for _, e := range o.Members() {
					if e.TypeDecl() == decl {
						return b.Name + "." + e.Name, true
					}
				}
			}
		}
	}
	return "", false
}

// writeType writes a type with the names the module at path has for the declared types in it,
// or false when it has a part the checker could not determine or a type the module cannot
// name.
func (a *analysis) writeType(path string, t types.Type) (string, bool) {
	if mentionsInvalid(t) {
		return "", false
	}
	ok := true
	var write func(t types.Type) string
	write = func(t types.Type) string {
		switch x := t.(type) {
		case types.Array:
			return "Array[" + write(x.Elem) + "]"
		case types.Map:
			return "Map[" + write(x.Key) + ", " + write(x.Val) + "]"
		case types.Set:
			return "Set[" + write(x.Elem) + "]"
		case types.Func:
			var sb strings.Builder
			sb.WriteString("func(")
			for i, p := range x.Params {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(write(p))
			}
			sb.WriteString("): " + write(x.Return))
			return sb.String()
		case types.Record:
			if len(x.Fields) == 0 {
				return "{}"
			}
			var sb strings.Builder
			sb.WriteString("{ ")
			for i, f := range x.Fields {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(f.Name + ": " + write(f.Type))
			}
			sb.WriteString(" }")
			return sb.String()
		case types.Named:
			name, found := a.localTypeName(path, x.Decl)
			if !found {
				ok = false
			}
			if len(x.Args) == 0 {
				return name
			}
			var sb strings.Builder
			sb.WriteString(name + "[")
			for i, arg := range x.Args {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(write(arg))
			}
			sb.WriteString("]")
			return sb.String()
		case types.Union:
			members := x.Members
			if x.Written != nil {
				members = x.Written
			}
			var sb strings.Builder
			for i, m := range members {
				if i > 0 {
					sb.WriteString(" | ")
				}
				if _, isFunc := m.(types.Func); isFunc {
					sb.WriteString("(" + write(m) + ")")
				} else {
					sb.WriteString(write(m))
				}
			}
			return sb.String()
		}
		return types.String(t)
	}
	s := write(t)
	return s, ok
}

// stateReturnType writes the inferred return type on the exported function at pos.
func (s *Server) stateReturnType(a *analysis, d *document, pos ast.Pos, dg Diagnostic) *CodeAction {
	var fd *ast.FuncDecl
	for _, n := range ast.Path(a.module().File, pos) {
		if x, ok := n.(*ast.FuncDecl); ok {
			fd = x
		}
	}
	if fd == nil || fd.Return != nil || fd.Body == nil {
		return nil
	}
	o := a.info.Decls[fd]
	if o == nil {
		return nil
	}
	fn, ok := o.Type().(types.Func)
	if !ok || fn.Return == nil || types.HasParam(fn.Return) && len(fd.TypeParams) == 0 {
		return nil
	}
	ret, ok := a.writeType(d.path, fn.Return)
	if !ok {
		return nil
	}
	// after the ) of the parameters: the last ) before the body
	off := d.offsetOf(fd.Body.Pos)
	for off > 0 && d.text[off-1] != ')' {
		off--
	}
	if off == 0 {
		return nil
	}
	at := d.position(off, s.utf16)
	return &CodeAction{Title: "State the return type " + ret, Kind: "quickfix", Diagnostics: []Diagnostic{dg}, IsPreferred: true,
		Edit: &WorkspaceEdit{Changes: map[string][]TextEdit{d.uri: {{Range: Range{Start: at, End: at}, NewText: ": " + ret}}}}}
}

// stateVarType writes the type of the var at pos, when the checker found it.
func (s *Server) stateVarType(a *analysis, d *document, pos ast.Pos, dg Diagnostic) *CodeAction {
	var decl *ast.VarDecl
	for _, n := range ast.Path(a.module().File, pos) {
		if x, ok := n.(*ast.VarDecl); ok {
			decl = x
		}
	}
	if decl == nil || decl.Type != nil {
		return nil
	}
	nt, ok := decl.Target.(*ast.NameTarget)
	if !ok {
		return nil
	}
	o := a.info.Decls[nt]
	if o == nil || o.Type() == nil {
		return nil
	}
	ty, ok := a.writeType(d.path, o.Type())
	if !ok {
		return nil
	}
	at := d.position(d.offsetOf(nt.End), s.utf16)
	return &CodeAction{Title: "State the type " + ty, Kind: "quickfix", Diagnostics: []Diagnostic{dg},
		Edit: &WorkspaceEdit{Changes: map[string][]TextEdit{d.uri: {{Range: Range{Start: at, End: at}, NewText: ": " + ty}}}}}
}
