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
			parts = append(parts, f.Name+": "+zeroValue(f.Type))
		}
	}
	if len(parts) == 0 {
		return nil
	}
	// insert before the closing brace, after a comma when entries are there
	closeOff := d.offsetOf(lit.End) - 1
	if closeOff < 0 || closeOff >= len(d.text) || d.text[closeOff] != '}' {
		return nil
	}
	text := strings.Join(parts, ", ")
	insertAt := closeOff
	for insertAt > 0 && (d.text[insertAt-1] == ' ' || d.text[insertAt-1] == '\n' || d.text[insertAt-1] == '\t') {
		insertAt--
	}
	switch {
	case len(lit.Entries) == 0:
		// an empty literal: Point{} becomes Point{ x: 0.0 }
		text = " " + text + " "
		insertAt = closeOff
	case d.text[insertAt-1] == ',':
		text = " " + text
	default:
		text = ", " + text
	}
	title := "Add the missing field " + parts[0][:strings.Index(parts[0], ":")]
	if len(parts) > 1 {
		title = "Add the missing fields"
	}
	pos2 := d.position(insertAt, s.utf16)
	return &CodeAction{Title: title, Kind: "quickfix", Diagnostics: []Diagnostic{dg}, IsPreferred: true,
		Edit: &WorkspaceEdit{Changes: map[string][]TextEdit{d.uri: {{Range: Range{Start: pos2, End: pos2}, NewText: text}}}}}
}

// zeroValue is a placeholder value of a type: its zero, or null for a nullable type.
func zeroValue(t types.Type) string {
	if types.Includes(t, types.Null{}) {
		return "null"
	}
	switch x := types.Expand(t).(type) {
	case types.Prim:
		switch {
		case x.Kind.IsFloat():
			return "0.0"
		case x.Kind.IsNumber():
			return "0"
		case x.Kind == value.KindString:
			return `""`
		case x.Kind == value.KindBool:
			return "false"
		}
	case types.Array:
		return "[]"
	case types.Map:
		return "Map{}"
	case types.Named:
		if x.Decl.Kind == types.RecordDecl {
			return x.Decl.Name + "{}"
		}
	case types.Func:
		return "{ }"
	case types.Void:
		return "void"
	}
	return "null"
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
	if !ok || fn.Return == nil || types.IsInvalid(fn.Return) || types.HasParam(fn.Return) && len(fd.TypeParams) == 0 {
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
	text := ": " + types.String(fn.Return)
	return &CodeAction{Title: "State the return type " + types.String(fn.Return), Kind: "quickfix", Diagnostics: []Diagnostic{dg}, IsPreferred: true,
		Edit: &WorkspaceEdit{Changes: map[string][]TextEdit{d.uri: {{Range: Range{Start: at, End: at}, NewText: text}}}}}
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
	if o == nil || o.Type() == nil || types.IsInvalid(o.Type()) {
		return nil
	}
	at := d.position(d.offsetOf(nt.End), s.utf16)
	return &CodeAction{Title: "State the type " + types.String(o.Type()), Kind: "quickfix", Diagnostics: []Diagnostic{dg},
		Edit: &WorkspaceEdit{Changes: map[string][]TextEdit{d.uri: {{Range: Range{Start: at, End: at}, NewText: ": " + types.String(o.Type())}}}}}
}
