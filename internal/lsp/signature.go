package lsp

import (
	"strings"
	"unicode/utf16"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/internal/types"
)

// signatureHelp answers textDocument/signatureHelp: the signature of the call the cursor is in
// and its active parameter (spec/lsp.md section 4).
func (s *Server) signatureHelp(p *SignatureHelpParams) *SignatureHelp {
	d := s.doc(p.TextDocument.URI)
	if d == nil {
		return nil
	}
	a := s.analyze(d)
	off := d.offset(p.Position, s.utf16)
	call, index := enclosingCall(a, d, off)
	if call == nil {
		// the call may not parse yet: give it an argument to read
		_, start := d.wordAt(off)
		text := d.text[:start] + placeholder + d.text[off:]
		a = s.analyzeText(d, text)
		call, index = enclosingCall(a, newDocument(d.uri, d.version, text), off)
		if call == nil {
			return nil
		}
	}
	ci := a.info.Calls[call]
	if ci == nil {
		return nil
	}
	label, params := signatureParts(a, ci)
	sig := SignatureInformation{Label: label, Parameters: params}
	if ci.Callee != nil {
		if doc := ci.Callee.Doc(); doc != "" {
			sig.Documentation = markdown(docMarkdown(doc, a.knows))
		}
	}
	active := activeParam(ci, index)
	return &SignatureHelp{Signatures: []SignatureInformation{sig}, ActiveSignature: 0, ActiveParameter: active}
}

// enclosingCall finds the innermost call whose arguments hold the cursor, and which argument
// the cursor is in, by the commas before it at the call's depth.
func enclosingCall(a *analysis, d *document, off int) (*ast.Call, int) {
	pos := d.posOf(off)
	nodes := ast.Path(a.module().File, pos)
	for i := len(nodes) - 1; i >= 0; i-- {
		call, ok := nodes[i].(*ast.Call)
		if !ok || !call.Pos.Before(pos) {
			continue // the cursor is on the function, not in the parentheses
		}
		open := d.offsetOf(call.Pos)
		if off <= open {
			continue
		}
		// count the arguments before the cursor: commas at depth 0 after the (, over the tokens,
		// so that a comma in a comment, a string, a regular expression, or markup is not one
		index := 0
		depth := 0
		for _, t := range streamAt(a.module().Layout.Tokens, open) {
			if t.Off <= open || t.Kind != token.Op {
				continue
			}
			if t.Off >= off {
				break
			}
			switch t.Text {
			case "(", "[", "{":
				depth++
			case ")", "]", "}":
				depth--
			case ",":
				if depth == 0 {
					index++
				}
			}
		}
		if call.Pipe {
			index++ // the left side fills the first argument
		}
		return call, index
	}
	return nil, 0
}

// streamAt is the token stream holding a token that starts at byte offset off: the module's, or
// that of an interpolation in a string or in markup; nil when no token starts there.
func streamAt(toks []token.Token, off int) []token.Token {
	for i := range toks {
		t := &toks[i]
		if t.Off == off {
			return toks
		}
		if t.Off < off && off < t.End {
			switch t.Kind {
			case token.Template:
				for _, p := range t.Parts {
					if p.Tokens != nil && p.Off <= off && off < p.End {
						return streamAt(p.Tokens, off)
					}
				}
			case token.Html:
				return markupStreamAt(t.Markup, off)
			}
			return nil
		}
	}
	return nil
}

// markupStreamAt is streamAt for the interpolations of markup.
func markupStreamAt(m *token.Markup, off int) []token.Token {
	if m == nil {
		return nil
	}
	for _, n := range m.Nodes {
		if off < n.Off || n.End <= off {
			continue
		}
		if s := streamAt(n.Tokens, off); s != nil {
			return s
		}
		for _, at := range n.Attrs {
			if s := streamAt(at.Tokens, off); s != nil {
				return s
			}
		}
		return markupStreamAt(n.Children, off)
	}
	return nil
}

// utf16Len is the length of s in UTF-16 code units.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += utf16.RuneLen(r)
	}
	return n
}

// activeParam is the parameter the argument at index fills: as the checker matched it, or by
// position for an argument not written yet.
func activeParam(ci *check.CallInfo, index int) int {
	if index < len(ci.Slots) && ci.Slots[index] >= 0 {
		return ci.Slots[index]
	}
	if ci.Recv {
		index++ // the receiver fills the first parameter (section 5, Calls through .)
	}
	if index < len(ci.Fn.Params) {
		return index
	}
	return len(ci.Fn.Params)
}

// signatureParts renders a call's signature with the offsets of each parameter's label.
func signatureParts(a *analysis, ci *check.CallInfo) (string, []ParameterInformation) {
	var sb strings.Builder
	switch {
	case ci.Callee != nil && ci.Callee.Kind == check.Func:
		sb.WriteString(ci.Callee.Owner)
		if ci.Callee.Owner != "" {
			sb.WriteString(".")
		}
		sb.WriteString(ci.Callee.Name)
		if tps := ci.Callee.TypeParams(); len(tps) > 0 {
			sb.WriteString("[")
			for i, p := range tps {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(p.Name)
				if p.Bound != nil {
					sb.WriteString(": " + types.String(p.Bound))
				}
			}
			sb.WriteString("]")
		}
	case ci.Callee != nil:
		sb.WriteString(ci.Callee.Name)
	default:
		sb.WriteString("func")
	}
	sb.WriteString("(")
	labels := paramLabels(ci.Fn, ci.Names, ci.Defaults)
	params := make([]ParameterInformation, len(labels))
	for i, l := range labels {
		if i > 0 {
			sb.WriteString(", ")
		}
		// the offsets count UTF-16 code units of the label, as the protocol says
		start := utf16Len(sb.String())
		sb.WriteString(l)
		params[i] = ParameterInformation{Label: [2]int{start, utf16Len(sb.String())}}
	}
	sb.WriteString("): ")
	sb.WriteString(typeString(ci.Fn.Return))
	return sb.String(), params
}
