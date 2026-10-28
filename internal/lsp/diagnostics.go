package lsp

import (
	"fmt"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/token"
)

// publishAll publishes the diagnostics of every open document (spec/lsp.md section 3).
func (s *Server) publishAll() {
	for _, d := range s.docs {
		if d.path != "" {
			s.publish(d)
		}
	}
}

// publish analyzes d as the root of its program and publishes its diagnostics: its own, and one
// on each import whose module has errors of its own.
func (s *Server) publish(d *document) {
	a := s.analyze(d)
	out := []Diagnostic{}
	root := a.module().File
	for _, dg := range a.diags {
		if dg.Path == root.Path {
			out = append(out, Diagnostic{Range: a.diagRange(root.Path, dg.Pos), Severity: 1, Source: "ant", Message: dg.Msg})
		}
	}
	for _, imp := range root.Imports {
		if imp.Module < 0 || imp.Module >= len(a.prog.Modules) {
			continue
		}
		m := a.prog.Modules[imp.Module]
		var first *loader.Diagnostic
		for i := range a.diags {
			if a.diags[i].Path == m.File.Path {
				first = &a.diags[i]
				break
			}
		}
		if first == nil {
			continue
		}
		dg := Diagnostic{Range: s.rng(d, imp.PathPos, imp.PathEnd), Severity: 1, Source: "ant",
			Message: fmt.Sprintf("%s has errors: %s", imp.Path, first.String())}
		if loc, ok := a.location(m.File.Path, first.Pos, first.Pos); ok {
			dg.RelatedInformation = []DiagnosticRelatedInformation{{Location: loc, Message: first.Msg}}
		}
		out = append(out, dg)
	}
	version := d.version
	s.conn.notify("textDocument/publishDiagnostics", PublishDiagnosticsParams{URI: d.uri, Version: &version, Diagnostics: out})
}

// diagRange is the range a diagnostic at pos covers: the token there when it is a keyword or an
// operator, as in a syntax error; else the innermost expression or declaration that starts
// there; else the token or word there.
func (a *analysis) diagRange(path string, pos ast.Pos) Range {
	d := a.doc(path)
	if d == nil || pos.Line == 0 {
		return Range{}
	}
	var tok *token.Token
	if m, ok := a.modules[path]; ok && m.Layout != nil {
		for i := range m.Layout.Tokens {
			t := &m.Layout.Tokens[i]
			if t.Pos.Line == pos.Line && t.Pos.Col == pos.Col && t.Kind != token.Newline && t.Kind != token.EOF {
				tok = t
				break
			}
		}
	}
	tokenRange := func() Range {
		return a.server.rng(d, pos, ast.Pos{Line: tok.EndPos.Line, Col: tok.EndPos.Col})
	}
	// The innermost node that starts at the position, or whose operator stands there: an error
	// on a field read or a binary operation is reported at the dot or the operator and covers
	// the whole expression.
	var found ast.Node
	if f := a.file(path); f != nil {
		for _, n := range ast.Path(f, pos) {
			switch n.(type) {
			case *ast.File, *ast.Block:
				continue
			}
			if (ast.Start(n) == pos || n.Position() == pos && operatorHeaded(n)) && pos.Before(n.EndPosition()) {
				found = n
			}
		}
	}
	if found != nil && (tok == nil || found.Position() == pos && operatorHeaded(found) ||
		tok.Kind != token.Keyword && tok.Kind != token.Op && tok.Kind != token.Reserved && tok.Kind != token.Bad) {
		return a.server.rng(d, ast.Start(found), found.EndPosition())
	}
	if tok != nil {
		return tokenRange()
	}
	// the word at the position, as for an error inside markup or a comment
	off := d.offsetOf(pos)
	end := off
	for end < len(d.text) && isIdentByte(d.text[end]) {
		end++
	}
	if end == off && off < len(d.text) && !strings.ContainsRune("\r\n", rune(d.text[off])) {
		end = off + 1
	}
	return Range{Start: d.position(off, a.server.utf16), End: d.position(end, a.server.utf16)}
}

// operatorHeaded reports whether a node's Position is its operator rather than its start.
func operatorHeaded(n ast.Node) bool {
	switch n.(type) {
	case *ast.Field, *ast.Index, *ast.Call, *ast.Binary, *ast.Is, *ast.Assign:
		return true
	}
	return false
}
