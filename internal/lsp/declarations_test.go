package lsp

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ant-src/ant/internal/prelude"
)

// TestPreludeDocument: the prelude's document is served as a file is (spec/lsp.md section 4):
// hover, definition, and semantic tokens work inside it, as in any declaration file.
func TestPreludeDocument(t *testing.T) {
	c := start(t, workspace())
	src := prelude.Source()
	var h *Hover
	c.call("textDocument/hover", c.pos(preludeURI, at(src, "trace: Array[Panic.Frame]", 1, 19)), &h)
	if h == nil || !strings.Contains(h.Contents.Value, "type Panic.Frame = {") {
		t.Errorf("hover on Panic.Frame in the prelude: %+v", h)
	}
	var locs []Location
	c.call("textDocument/definition", c.pos(preludeURI, at(src, "trace: Array[Panic.Frame]", 1, 19)), &locs)
	if len(locs) != 1 || locs[0].URI != preludeURI || locs[0].Range.Start != at(src, "type Panic.Frame", 1, 11) {
		t.Errorf("definition of Panic.Frame in the prelude: %+v", locs)
	}
	h = nil
	c.call("textDocument/hover", c.pos(preludeURI, at(src, "func Array.map[T, U]", 1, 11)), &h)
	if h == nil || !strings.Contains(h.Contents.Value, "func Array.map[T, U](xs: Array[T], f: func(T): U): Array[U]") {
		t.Errorf("hover on Array.map in the prelude: %+v", h)
	}
	var toks *SemanticTokens
	c.call("textDocument/semanticTokens/full", SemanticTokensParams{TextDocument: TextDocumentIdentifier{URI: preludeURI}}, &toks)
	if toks == nil || len(toks.Data) == 0 {
		t.Errorf("semantic tokens of the prelude: %+v", toks)
	}
	h = nil
	c.call("textDocument/hover", c.pos(preludeURI, at(src, "func Array.map[T, U](xs", 1, 21)), &h)
	if h == nil || !strings.Contains(h.Contents.Value, "xs: Array[T]") {
		t.Errorf("hover on a parameter in the prelude: %+v", h)
	}
	// The outline lists each namespace once, where its first member is, with its members under
	// it, and a type's qualified types and functions under the type.
	var syms []DocumentSymbol
	c.call("textDocument/documentSymbol", DocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: preludeURI}}, &syms)
	byName := map[string]DocumentSymbol{}
	for _, s := range syms {
		if _, dup := byName[s.Name]; dup {
			t.Errorf("symbol %s listed twice", s.Name)
		}
		byName[s.Name] = s
	}
	if arr := byName["Array"]; arr.Kind != SymbolNamespace || len(arr.Children) < 20 || arr.Children[0].Kind != SymbolMethod ||
		before(arr.Children[len(arr.Children)-1].Range.End, arr.Range.End) || arr.SelectionRange != arr.Children[0].SelectionRange {
		t.Errorf("the namespace Array: %+v", arr)
	}
	var frame, catch bool
	for _, ch := range byName["Panic"].Children {
		frame = frame || ch.Name == "Frame" && ch.Kind == SymbolStruct
		catch = catch || ch.Name == "catch" && ch.Kind == SymbolMethod
	}
	if byName["Panic"].Kind != SymbolStruct || !frame || !catch {
		t.Errorf("the type Panic: %+v", byName["Panic"])
	}
	if math := byName["Math"]; math.Kind != SymbolNamespace || !slices.ContainsFunc(math.Children, func(s DocumentSymbol) bool {
		return s.Name == "PI" && s.Kind == SymbolVariable && s.Detail == "float64"
	}) {
		t.Errorf("the namespace Math: %+v", math)
	}
	// Every request answers on the prelude's document without harming the server.
	id := TextDocumentIdentifier{URI: preludeURI}
	p := c.pos(preludeURI, at(src, "func Array.map[T, U]", 1, 11))
	var raw json.RawMessage
	for _, r := range []struct {
		method string
		params any
	}{
		{"textDocument/completion", CompletionParams{TextDocumentPositionParams: p}},
		{"textDocument/signatureHelp", p},
		{"textDocument/references", ReferenceParams{TextDocumentPositionParams: p}},
		{"textDocument/documentHighlight", p},
		{"textDocument/prepareRename", p},
		{"textDocument/formatting", DocumentFormattingParams{TextDocument: id}},
		{"textDocument/foldingRange", FoldingRangeParams{TextDocument: id}},
		{"textDocument/selectionRange", SelectionRangeParams{TextDocument: id, Positions: []Position{p.Position}}},
		{"textDocument/inlayHint", InlayHintParams{TextDocument: id, Range: Range{End: Position{Line: 2000}}}},
		{"textDocument/codeAction", CodeActionParams{TextDocument: id}},
	} {
		err := c.call(r.method, r.params, &raw)
		if r.method == "textDocument/prepareRename" {
			if err == nil || err.Message != "a name of the prelude cannot be renamed" {
				t.Errorf("prepareRename on the prelude: %v", err)
			}
		} else if err != nil {
			t.Errorf("%s on the prelude: %v", r.method, err.Message)
		}
	}
	// A constant's name and a declaration's owner resolve.
	h = nil
	c.call("textDocument/hover", c.pos(preludeURI, at(src, "var Math.PI", 1, 9)), &h)
	if h == nil || !strings.Contains(h.Contents.Value, "var Math.PI: float64") {
		t.Errorf("hover on the constant Math.PI: %+v", h)
	}
	h = nil
	c.call("textDocument/hover", c.pos(preludeURI, at(src, "type Panic.Frame", 1, 6)), &h)
	if h == nil || !strings.Contains(h.Contents.Value, "type Panic = {") {
		t.Errorf("hover on the owner Panic of Panic.Frame: %+v", h)
	}
	locs = nil
	c.call("textDocument/definition", c.pos(preludeURI, at(src, "type Panic.Frame", 1, 6)), &locs)
	if len(locs) != 1 || locs[0].Range.Start != at(src, "type Panic = ", 1, 5) {
		t.Errorf("definition of the owner Panic: %+v", locs)
	}
	var ws []SymbolInformation
	c.call("workspace/symbol", WorkspaceSymbolParams{Query: "Frame"}, &ws)
	for _, w := range ws {
		if w.Name == "Frame" && w.ContainerName != "Panic" {
			t.Errorf("workspace symbol Frame: %+v", w)
		}
	}
	h = nil
	c.call("textDocument/hover", c.pos(preludeURI, at(src, "trace: Array[Panic.Frame]", 1, 19)), &h)
	if h == nil {
		t.Error("hover after every request on the prelude")
	}
	// The client may open the prelude under its URI; the server still serves its own copy.
	c.notify("textDocument/didOpen", DidOpenTextDocumentParams{TextDocument: TextDocumentItem{URI: preludeURI, LanguageID: "ant", Version: 1, Text: src}})
	h = nil
	c.call("textDocument/hover", c.pos(preludeURI, at(src, "trace: Array[Panic.Frame]", 1, 19)), &h)
	if h == nil || !strings.Contains(h.Contents.Value, "type Panic.Frame = {") {
		t.Errorf("hover after the client opened the prelude: %+v", h)
	}
}

// TestDeclarationFileDocument: a .d.ant file is a program of its own (spec/lsp.md section 2),
// checked as a declaration file, with hover and definition inside it.
func TestDeclarationFileDocument(t *testing.T) {
	c := start(t, workspace())
	src := "// Db.Row is a row.\ntype Db.Row = Map[string, string]\nfunc Db.query(sql: string): Result[Array[Db.Row], Error]\nfunc Db.count(): Count\n"
	uri := c.open("db.d.ant", src)
	diags := c.diagnostics(uri)
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "unknown type Count") || diags[0].Range.Start != at(src, "Count", 1, 0) {
		t.Errorf("diagnostics of a declaration file: %+v", diags)
	}
	var h *Hover
	c.call("textDocument/hover", c.pos(uri, at(src, "Array[Db.Row]", 1, 9)), &h)
	if h == nil || !strings.Contains(h.Contents.Value, "type Db.Row = Map[string, string]") || !strings.Contains(h.Contents.Value, "Db.Row is a row.") {
		t.Errorf("hover on Db.Row: %+v", h)
	}
	var locs []Location
	c.call("textDocument/definition", c.pos(uri, at(src, "Array[Db.Row]", 1, 9)), &locs)
	if len(locs) != 1 || locs[0].URI != uri || locs[0].Range.Start != at(src, "type Db.Row", 1, 8) {
		t.Errorf("definition of Db.Row: %+v", locs)
	}
	locs = nil
	c.call("textDocument/definition", c.pos(uri, at(src, "], Error]", 1, 3)), &locs)
	if len(locs) != 1 || locs[0].URI != preludeURI || locs[0].Range.Start != at(prelude.Source(), "type Error = ", 1, 5) {
		t.Errorf("definition of Error from a declaration file: %+v", locs)
	}
	var syms []DocumentSymbol
	c.call("textDocument/documentSymbol", DocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: uri}}, &syms)
	if len(syms) != 1 || syms[0].Name != "Db" || syms[0].Kind != SymbolNamespace || len(syms[0].Children) != 3 ||
		syms[0].Children[0].Name != "Row" || syms[0].Children[1].Name != "query" || syms[0].Children[1].Detail != "(sql: string): Result[Array[Db.Row], Error]" {
		t.Errorf("symbols of a declaration file: %+v", syms)
	}
	h = nil
	c.call("textDocument/hover", c.pos(uri, at(src, "type Db.Row", 1, 5)), &h)
	if h == nil || !strings.Contains(h.Contents.Value, "Db") {
		t.Errorf("hover on the namespace Db: %+v", h)
	}
	// A namespace whose members are interrupted is listed once per run, each holding only its own.
	split := "func Db.a(): int\ntype Point = { x: int }\nvar Point.ORIGIN: Point\nfunc Db.b(): int\n"
	uri2 := c.open("split.d.ant", split)
	syms = nil
	c.call("textDocument/documentSymbol", DocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: uri2}}, &syms)
	var names []string
	for _, s := range syms {
		names = append(names, s.Name)
	}
	if strings.Join(names, " ") != "Db Point Db" || syms[0].Range.End.Line != 0 || len(syms[1].Children) != 2 || syms[1].Children[1].Name != "ORIGIN" {
		t.Errorf("symbols of a split namespace: %v %+v", names, syms)
	}
}
