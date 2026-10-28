package lsp

// codeActions answers textDocument/codeAction (spec/lsp.md section 4). The mechanical fixes are
// added with the diagnostics they fix; none yet.
func (s *Server) codeActions(p *CodeActionParams) []CodeAction {
	return []CodeAction{}
}
