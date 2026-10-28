package lsp

import (
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/value"
)

// hover answers textDocument/hover (spec/lsp.md section 4).
func (s *Server) hover(p *TextDocumentPositionParams) *Hover {
	d := s.doc(p.TextDocument.URI)
	if d == nil {
		return nil
	}
	a := s.analyze(d)
	path := a.module().File.Path
	pos := d.posOf(d.offset(p.Position, s.utf16))
	t := a.at(path, pos)
	if t == nil {
		return nil
	}
	rng := s.rng(d, t.from, t.to)
	switch {
	case t.imp != nil:
		if t.imp.Module >= 0 && t.imp.Module < len(a.prog.Modules) {
			m := a.prog.Modules[t.imp.Module]
			text := code("import " + quote(t.imp.Path))
			if m.File.Doc != "" {
				text += "\n\n" + docMarkdown(m.File.Doc, a.knows)
			}
			return &Hover{Contents: *markdown(text), Range: &rng}
		}
		return nil
	case t.element != nil:
		return &Hover{Contents: *markdown(elementHover(t.element)), Range: &rng}
	case t.attr != nil:
		return &Hover{Contents: *markdown(attrHover(t.attr)), Range: &rng}
	case t.obj == nil:
		return nil
	}
	head, doc := a.describe(t.obj)
	// A variable's type at this use, narrowed, when it differs from the declared one.
	if id, ok := t.node.(*ast.Ident); ok && (t.obj.Kind == check.Var || t.obj.Kind == check.Param) {
		if ty, ok := a.info.Types[id]; ok && ty != nil && typeString(ty) != typeString(t.obj.Type()) {
			head += "\n// here: " + typeString(ty)
		}
	}
	text := code(head)
	if doc != "" {
		text += "\n\n" + docMarkdown(doc, a.knows)
	}
	return &Hover{Contents: *markdown(text), Range: &rng}
}

// elementHover describes a markup element (section 5, HTML expressions).
func elementHover(el *ast.HtmlElement) string {
	lower := strings.ToLower(el.Name)
	var notes []string
	switch {
	case value.HtmlVoid[lower]:
		notes = append(notes, "A void element: written self-closing, rendered without an end tag.")
	case value.HtmlRawText[lower]:
		notes = append(notes, "Holds literal text up to its end tag; `{` and `<` are not special inside.")
	case value.HtmlTextOnly[lower]:
		notes = append(notes, "Holds text only outside `<svg>` and `<math>`: strings, numbers, void, and arrays of them, but not Html.")
	}
	text := code("<" + el.Name + ">")
	if len(notes) > 0 {
		text += "\n\n" + strings.Join(notes, " ")
	}
	return text
}

// attrHover describes a markup attribute.
func attrHover(at *ast.HtmlAttr) string {
	lower := strings.ToLower(at.Name)
	var notes []string
	switch {
	case value.HtmlBoolean[lower]:
		notes = append(notes, "A boolean attribute: `true` gives the empty value and `false` leaves it out; a string works too.")
	case value.HtmlCodeAttr(lower):
		notes = append(notes, "Runs code, so it takes only literal text.")
	case value.HtmlURL[lower] || value.HtmlSrcset[lower]:
		notes = append(notes, "A URL attribute: an interpolated value with a scheme other than http, https, mailto, or tel becomes `about:invalid#ant`.")
	case lower == "class":
		notes = append(notes, "Takes a string, a number, null, or an array of string, null, and void, joined with spaces.")
	}
	text := code(at.Name)
	if len(notes) > 0 {
		text += "\n\n" + strings.Join(notes, " ")
	}
	return text
}
