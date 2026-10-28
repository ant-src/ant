package lsp

import (
	"github.com/ant-src/ant/ast"
)

// selectionRanges answers textDocument/selectionRange: the nodes enclosing each position,
// innermost first (spec/lsp.md section 4).
func (s *Server) selectionRanges(p *SelectionRangeParams) []SelectionRange {
	d, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return nil
	}
	a := s.analyze(d)
	f := a.module().File
	out := make([]SelectionRange, 0, len(p.Positions))
	for _, pos := range p.Positions {
		nodes := ast.Path(f, d.posOf(d.offset(pos, s.utf16)))
		var chain *SelectionRange
		var last Range
		for _, n := range nodes {
			r := s.rng(d, ast.Start(n), n.EndPosition())
			if chain != nil && r == last {
				continue
			}
			chain = &SelectionRange{Range: r, Parent: chain}
			last = r
		}
		if chain == nil {
			chain = &SelectionRange{Range: Range{Start: pos, End: pos}}
		}
		out = append(out, *chain)
	}
	return out
}
