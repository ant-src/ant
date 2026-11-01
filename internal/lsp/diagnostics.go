package lsp

import (
	"fmt"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/diag"
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
		// The module's own first error, else the first error of a module it imports, at any
		// depth: a broken import is visible from every file that depends on it.
		first := a.firstError(imp.Module, map[int]bool{})
		if first == nil {
			continue
		}
		dg := Diagnostic{Range: s.rng(d, imp.PathPos, imp.PathEnd), Severity: 1, Source: "ant",
			Message: fmt.Sprintf("%s has errors: %s", imp.Path, first.String())}
		if loc, ok := a.location(first.Path, first.Pos, first.Pos); ok {
			dg.RelatedInformation = []DiagnosticRelatedInformation{{Location: loc, Message: first.Msg}}
		}
		out = append(out, dg)
	}
	version := d.version
	s.conn.notify("textDocument/publishDiagnostics", PublishDiagnosticsParams{URI: d.uri, Version: &version, Diagnostics: out})
}

// firstError is the first diagnostic of the module at index i, or of a module it imports, found
// depth first in import order, or nil when the whole subtree is clean.
func (a *analysis) firstError(i int, seen map[int]bool) *loader.Diagnostic {
	if i < 0 || i >= len(a.prog.Modules) || seen[i] {
		return nil
	}
	seen[i] = true
	m := a.prog.Modules[i]
	for k := range a.diags {
		if a.diags[k].Path == m.File.Path {
			return &a.diags[k]
		}
	}
	for _, imp := range m.File.Imports {
		if first := a.firstError(imp.Module, seen); first != nil {
			return first
		}
	}
	return nil
}

// diagRange is the range a diagnostic at pos covers (diag.Span): the token there when it is a
// keyword or an operator, as in a syntax error; else the innermost expression or declaration that
// starts there; else the token or word there.
func (a *analysis) diagRange(path string, pos ast.Pos) Range {
	d := a.doc(path)
	if d == nil || pos.Line == 0 {
		return Range{}
	}
	var toks []token.Token
	if m, ok := a.modules[path]; ok && m.Layout != nil {
		toks = m.Layout.Tokens
	}
	from, to := diag.Span(a.file(path), toks, d.text, d.lines, pos)
	return Range{Start: d.position(from, a.server.utf16), End: d.position(to, a.server.utf16)}
}
