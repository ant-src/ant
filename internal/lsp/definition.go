package lsp

import (
	"github.com/ant-src/ant/ast"
)

// definition answers textDocument/definition (spec/lsp.md section 4).
func (s *Server) definition(p *TextDocumentPositionParams) []Location {
	d, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return nil
	}
	a := s.analyze(d)
	path := a.module().File.Path
	pos := d.posOf(d.offset(p.Position, s.utf16))
	t := a.at(path, pos)
	if t == nil {
		return nil
	}
	if t.imp != nil {
		if t.imp.Module >= 0 && t.imp.Module < len(a.prog.Modules) {
			m := a.prog.Modules[t.imp.Module]
			if loc, ok := a.location(m.File.Path, ast.Pos{Line: 1, Col: 1}, ast.Pos{Line: 1, Col: 1}); ok {
				return []Location{loc}
			}
		}
		return nil
	}
	if t.obj == nil {
		return nil
	}
	declPath, from, to, ok := a.declRange(t.obj)
	if !ok {
		return nil
	}
	if declPath == "" {
		declPath = a.pathOf(t.obj)
	}
	if loc, ok := a.location(declPath, from, to); ok {
		return []Location{loc}
	}
	return nil
}
