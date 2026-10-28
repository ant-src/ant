package lsp

import (
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
)

// inlayHints answers textDocument/inlayHint: the inferred type of a var without an annotation,
// and the parameter name before a positional argument (spec/lsp.md section 4).
func (s *Server) inlayHints(p *InlayHintParams) []InlayHint {
	d, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return nil
	}
	a := s.analyze(d)
	in := a.info
	from, to := d.posOf(d.offset(p.Range.Start, s.utf16)), d.posOf(d.offset(p.Range.End, s.utf16))
	inRange := func(pos ast.Pos) bool { return !pos.Before(from) && !to.Before(pos) }
	out := []InlayHint{}
	ast.Inspect(a.module().File, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.VarDecl:
			if x.Type != nil || x.Init == nil {
				return true
			}
			if _, isLit := x.Init.(*ast.FuncLit); isLit {
				return true
			}
			for _, nt := range targetNames(x.Target) {
				if !inRange(nt.End) {
					continue
				}
				o := in.Decls[nt]
				if o == nil || o.Type() == nil || types.IsInvalid(o.Type()) {
					continue
				}
				out = append(out, InlayHint{Position: d.position(d.offsetOf(nt.End), s.utf16), Label: ": " + types.String(o.Type()), Kind: 1})
			}
		case *ast.Call:
			ci := in.Calls[x]
			if ci == nil || ci.Callee == nil || len(ci.Names) == 0 {
				return true
			}
			for i, arg := range x.Args {
				if arg.Name != "" || x.Pipe && i == 0 || i >= len(ci.Slots) || ci.Slots[i] < 0 || ci.Slots[i] >= len(ci.Names) {
					continue
				}
				name := ci.Names[ci.Slots[i]]
				if id, ok := arg.Value.(*ast.Ident); ok && id.Name == name {
					continue // the argument says it
				}
				start := ast.Start(arg.Value)
				if !inRange(start) {
					continue
				}
				out = append(out, InlayHint{Position: d.position(d.offsetOf(start), s.utf16), Label: name + ":", Kind: 2, PaddingRight: true})
			}
		}
		return true
	})
	return out
}
