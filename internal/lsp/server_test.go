package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// client drives a server over pipes in a test.
type client struct {
	t        *testing.T
	w        io.Writer
	msgs     chan *message
	nextID   int
	done     chan int
	workdir  string
	utf16    bool
	pending  []*message // notifications read while waiting for a response
	received []PublishDiagnosticsParams
}

// start runs a server and a client over pipes, with a workspace of the given files.
func start(t *testing.T, files map[string]string) *client {
	t.Helper()
	dir := t.TempDir()
	for name, text := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	toServer, clientW := io.Pipe()
	serverR, fromServer := io.Pipe()
	_ = serverR
	c := &client{t: t, w: clientW, msgs: make(chan *message, 100), done: make(chan int, 1), workdir: dir, utf16: true}
	go func() {
		c.done <- Serve(context.Background(), toServer, fromServer, io.Discard)
		fromServer.Close()
	}()
	go func() {
		conn := &conn{r: bufio.NewReader(serverR)}
		for {
			m, err := conn.read()
			if err != nil {
				close(c.msgs)
				return
			}
			c.msgs <- m
		}
	}()
	var result InitializeResult
	c.call("initialize", InitializeParams{RootURI: fileURI(dir), WorkspaceFolders: []WorkspaceFolder{{URI: fileURI(dir), Name: "ws"}}}, &result)
	if result.Capabilities.PositionEncoding != "utf-16" || !result.Capabilities.HoverProvider {
		t.Fatalf("capabilities %+v", result.Capabilities)
	}
	c.notify("initialized", struct{}{})
	t.Cleanup(func() {
		c.call("shutdown", nil, nil)
		c.notify("exit", nil)
		select {
		case code := <-c.done:
			if code != 0 {
				t.Errorf("exit status %d", code)
			}
		case <-time.After(5 * time.Second):
			t.Error("the server did not exit")
		}
	})
	return c
}

func (c *client) send(v any) {
	body, _ := json.Marshal(v)
	fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n%s", len(body), body)
}

func (c *client) notify(method string, params any) {
	c.send(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

// call sends a request and decodes its result into result; notifications read meanwhile are
// kept.
func (c *client) call(method string, params any, result any) *rpcError {
	c.t.Helper()
	c.nextID++
	id := c.nextID
	c.send(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	for {
		select {
		case m, ok := <-c.msgs:
			if !ok {
				c.t.Fatalf("%s: the server ended", method)
			}
			if m.Method != "" {
				c.note(m)
				continue
			}
			var got int
			json.Unmarshal(m.ID, &got)
			if got != id {
				continue
			}
			if m.Error != nil {
				return m.Error
			}
			if result != nil && len(m.Result) > 0 && string(m.Result) != "null" {
				if err := json.Unmarshal(m.Result, result); err != nil {
					c.t.Fatalf("%s: %v in %s", method, err, m.Result)
				}
			}
			return nil
		case <-time.After(10 * time.Second):
			c.t.Fatalf("%s: no response", method)
		}
	}
}

func (c *client) note(m *message) {
	if m.Method == "textDocument/publishDiagnostics" {
		var p PublishDiagnosticsParams
		json.Unmarshal(m.Params, &p)
		c.received = append(c.received, p)
	}
	c.pending = append(c.pending, m)
}

// diagnostics waits for the diagnostics of a document, published after the last change.
func (c *client) diagnostics(uri string) []Diagnostic {
	c.t.Helper()
	// a round trip makes sure the notifications of the last change have arrived
	c.call("textDocument/hover", TextDocumentPositionParams{TextDocument: TextDocumentIdentifier{URI: uri}}, nil)
	for i := len(c.received) - 1; i >= 0; i-- {
		if c.received[i].URI == uri {
			return c.received[i].Diagnostics
		}
	}
	return nil
}

func (c *client) uri(name string) string {
	return fileURI(filepath.Join(c.workdir, filepath.FromSlash(name)))
}

func (c *client) open(name, text string) string {
	uri := c.uri(name)
	c.notify("textDocument/didOpen", DidOpenTextDocumentParams{TextDocument: TextDocumentItem{URI: uri, LanguageID: "ant", Version: 1, Text: text}})
	return uri
}

func (c *client) change(uri string, version int, text string) {
	c.notify("textDocument/didChange", DidChangeTextDocumentParams{TextDocument: VersionedTextDocumentIdentifier{URI: uri, Version: version},
		ContentChanges: []TextDocumentContentChangeEvent{{Text: text}}})
}

// at finds the position of the nth occurrence of needle in text, plus skip characters.
func at(text, needle string, n, skip int) Position {
	off := -1
	for ; n > 0; n-- {
		i := strings.Index(text[off+1:], needle)
		if i < 0 {
			panic("no " + needle)
		}
		off += i + 1
	}
	d := newDocument("", 0, text)
	return d.position(off+skip, true)
}

func (c *client) pos(uri string, p Position) TextDocumentPositionParams {
	return TextDocumentPositionParams{TextDocument: TextDocumentIdentifier{URI: uri}, Position: p}
}

const geometry = `//! Geometry has points on the plane.

// Point is a position on the plane.
export type Point = { x: float64, y: float64 }

export var origin = Point{ x: 0.0, y: 0.0 }

// Point.add adds two points.
export func Point.add(a: Point, b: Point): Point { Point{ x: a.x + b.x, y: a.y + b.y } }

// distance returns the straight-line distance from a to b.
export func distance(a: Point, b: Point): float64 {
  var dx = a.x - b.x
  Math.sqrt(dx * dx)
}
`

const mainSrc = `import "./lib/geometry.ant" { Point, distance: dist }
import "./lib/geometry.ant" Geo

type Shape = { name: string, area: float64 }

// describe names a shape.
func describe(s: Shape, times: int = 1): string {
  var label = s.name
  var p = Point{ x: 1.0, y: 2.0 }
  var d = dist(p, Geo.origin)
  var xs = [1, 2, 3] | Array.map({ x * times })
  if s.area > 0.0 { label } else { "none" }
}

func first(v: int | null): int {
  if v != null { v } else { 0 }
}
`

func workspace() map[string]string {
	return map[string]string{"lib/geometry.ant": geometry, "main.ant": mainSrc}
}

func TestDiagnostics(t *testing.T) {
	c := start(t, workspace())
	uri := c.open("main.ant", mainSrc)
	if diags := c.diagnostics(uri); len(diags) != 0 {
		t.Fatalf("diagnostics on a clean program: %v", diags)
	}
	// a type error and a syntax error, with analysis past the latter
	broken := strings.Replace(mainSrc, `var label = s.name`, "var label: int = s.name\n  var q = ", 1)
	c.change(uri, 2, broken)
	diags := c.diagnostics(uri)
	var msgs []string
	for _, d := range diags {
		msgs = append(msgs, fmt.Sprintf("%d:%d-%d:%d %s", d.Range.Start.Line, d.Range.Start.Character, d.Range.End.Line, d.Range.End.Character, d.Message))
	}
	// The lexer joins the line after = with the next, so the syntax error names its var.
	want := []string{
		"7:19-7:25 cannot use string where int is expected",
		"9:2-9:5 unexpected 'var'",
		"12:20-12:25 cannot use int where string is expected",
	}
	if strings.Join(msgs, "\n") != strings.Join(want, "\n") {
		t.Errorf("diagnostics:\n%s\nwant:\n%s", strings.Join(msgs, "\n"), strings.Join(want, "\n"))
	}
	// an error in the imported module shows on the import
	c.change(uri, 3, mainSrc)
	libURI := c.open("lib/geometry.ant", strings.Replace(geometry, "var dx = a.x - b.x", "var dx = a.x - b.z", 1))
	diags = c.diagnostics(uri)
	if len(diags) != 2 || !strings.Contains(diags[0].Message, "./lib/geometry.ant has errors") || diags[0].RelatedInformation == nil {
		t.Errorf("import diagnostics: %+v", diags)
	}
	if ld := c.diagnostics(libURI); len(ld) != 1 || !strings.Contains(ld[0].Message, "no field z") {
		t.Errorf("library diagnostics: %+v", ld)
	}
}

func TestHover(t *testing.T) {
	c := start(t, workspace())
	uri := c.open("main.ant", mainSrc)
	cases := []struct {
		name string
		pos  Position
		want []string
	}{
		{"imported function", at(mainSrc, "dist(p", 1, 0), []string{"func distance(a: Point, b: Point): float64", "distance returns the straight-line"}},
		{"module namespace", at(mainSrc, "Geo.origin", 1, 0), []string{`import "./lib/geometry.ant" Geo`, "Geometry has points"}},
		{"record field", at(mainSrc, "s.name", 1, 2), []string{"name: string", "A field of Shape"}},
		{"narrowed parameter", at(mainSrc, "{ v }", 1, 2), []string{"v: int | null", "// here: int"}},
		{"local variable", at(mainSrc, "label }", 1, 0), []string{"var label: string"}},
		{"prelude function", at(mainSrc, "Array.map", 1, 6), []string{"func Array.map[T, U]", "Array.map returns"}},
		{"type", at(mainSrc, "s: Shape", 1, 3), []string{"type Shape = { name: string, area: float64 }"}},
		{"lambda x", at(mainSrc, "x * times", 1, 0), []string{"var x: int"}},
		{"import path", at(mainSrc, `"./lib/geometry.ant" Geo`, 1, 3), []string{`import "./lib/geometry.ant"`, "Geometry has points"}},
	}
	for _, tc := range cases {
		var h Hover
		c.call("textDocument/hover", c.pos(uri, tc.pos), &h)
		for _, w := range tc.want {
			if !strings.Contains(h.Contents.Value, w) {
				t.Errorf("%s: hover %q lacks %q", tc.name, h.Contents.Value, w)
			}
		}
	}
	var none *Hover
	c.call("textDocument/hover", c.pos(uri, at(mainSrc, "1.0", 1, 0)), &none)
	if none != nil {
		t.Errorf("hover on a literal: %+v", none)
	}
}

func TestDefinition(t *testing.T) {
	c := start(t, workspace())
	uri := c.open("main.ant", mainSrc)
	var locs []Location
	c.call("textDocument/definition", c.pos(uri, at(mainSrc, "dist(p", 1, 0)), &locs)
	if len(locs) != 1 || locs[0].URI != c.uri("lib/geometry.ant") || locs[0].Range.Start != at(geometry, "distance(a", 1, 0) {
		t.Errorf("definition of dist: %+v", locs)
	}
	locs = nil
	c.call("textDocument/definition", c.pos(uri, at(mainSrc, "label }", 1, 0)), &locs)
	if len(locs) != 1 || locs[0].URI != uri || locs[0].Range.Start != at(mainSrc, "label =", 1, 0) {
		t.Errorf("definition of label: %+v", locs)
	}
	locs = nil
	c.call("textDocument/definition", c.pos(uri, at(mainSrc, "Array.map", 1, 7)), &locs)
	if len(locs) != 1 || locs[0].URI != preludeURI {
		t.Errorf("definition of Array.map: %+v", locs)
	}
	var src string
	c.call("ant/source", SourceParams{URI: preludeURI}, &src)
	if !strings.Contains(src, "func Array.map[T, U]") {
		t.Errorf("prelude source: %d bytes", len(src))
	}
	locs = nil
	c.call("textDocument/definition", c.pos(uri, at(mainSrc, `"./lib/geometry.ant" Geo`, 1, 3)), &locs)
	if len(locs) != 1 || locs[0].URI != c.uri("lib/geometry.ant") {
		t.Errorf("definition of an import path: %+v", locs)
	}
}

func TestReferencesAndRename(t *testing.T) {
	c := start(t, workspace())
	uri := c.open("main.ant", mainSrc)
	var locs []Location
	p := ReferenceParams{TextDocumentPositionParams: c.pos(uri, at(mainSrc, "s: Shape", 1, 0))}
	p.Context.IncludeDeclaration = true
	c.call("textDocument/references", p, &locs)
	if len(locs) != 3 {
		t.Errorf("references of s: %d, want 3: %+v", len(locs), locs)
	}
	var hl []DocumentHighlight
	c.call("textDocument/documentHighlight", c.pos(uri, at(mainSrc, "Point{", 1, 0)), &hl)
	if len(hl) != 2 {
		t.Errorf("highlights of Point: %+v", hl)
	}
	libURI := c.open("lib/geometry.ant", geometry)
	var edit WorkspaceEdit
	c.call("textDocument/rename", RenameParams{TextDocumentPositionParams: c.pos(libURI, at(geometry, "distance(a", 1, 0)), NewName: "dist2"}, &edit)
	if len(edit.Changes[libURI]) != 1 || len(edit.Changes[uri]) != 1 {
		// the declaration, and the import name in main, whose alias dist stays
		t.Errorf("renaming distance: %+v", edit.Changes)
	}
	edit = WorkspaceEdit{}
	c.call("textDocument/rename", RenameParams{TextDocumentPositionParams: c.pos(uri, at(mainSrc, "dist(p", 1, 0)), NewName: "far"}, &edit)
	if n := len(edit.Changes[uri]); n != 2 {
		t.Errorf("renaming the alias dist: %d edits in main, want 2: %+v", n, edit.Changes)
	}
	if _, touched := edit.Changes[c.uri("lib/geometry.ant")]; touched {
		t.Errorf("renaming the alias touched the library")
	}
	edit = WorkspaceEdit{}
	c.call("textDocument/rename", RenameParams{TextDocumentPositionParams: c.pos(uri, at(mainSrc, "Point{", 1, 0)), NewName: "Pt"}, &edit)
	if len(edit.Changes[uri]) != 2 || len(edit.Changes[c.uri("lib/geometry.ant")]) < 5 {
		t.Errorf("renaming Point: %d in main, %d in the library", len(edit.Changes[uri]), len(edit.Changes[c.uri("lib/geometry.ant")]))
	}
	if err := c.call("textDocument/rename", RenameParams{TextDocumentPositionParams: c.pos(uri, at(mainSrc, "label }", 1, 0)), NewName: "type"}, nil); err == nil || !strings.Contains(err.Message, "keyword") {
		t.Errorf("renaming to a keyword: %v", err)
	}
	if err := c.call("textDocument/rename", RenameParams{TextDocumentPositionParams: c.pos(uri, at(mainSrc, "Array.map", 1, 7)), NewName: "mapp"}, nil); err == nil {
		t.Error("renaming a prelude function was allowed")
	}
}

func TestCompletion(t *testing.T) {
	c := start(t, workspace())
	uri := c.open("main.ant", mainSrc)
	labels := func(text string, p Position) []string {
		c.change(uri, 2, text)
		var list CompletionList
		c.call("textDocument/completion", CompletionParams{TextDocumentPositionParams: c.pos(uri, p)}, &list)
		var out []string
		for _, it := range list.Items {
			out = append(out, it.Label)
		}
		return out
	}
	has := func(name string, labels []string, want ...string) {
		t.Helper()
		for _, w := range want {
			found := false
			for _, l := range labels {
				if l == w {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: %q is not offered among %d items: %q", name, w, len(labels), labels)
			}
		}
	}
	lacks := func(name string, labels []string, unwanted ...string) {
		t.Helper()
		for _, u := range unwanted {
			for _, l := range labels {
				if l == u {
					t.Errorf("%s: %q is offered", name, u)
				}
			}
		}
	}
	// record fields after a dot, while the name is typed
	src := strings.Replace(mainSrc, "var label = s.name", "var label = s.na", 1)
	got := labels(src, at(src, "s.na", 1, 4))
	has("fields", got, "name", "area")
	lacks("fields", got, "describe")
	// namespace members after a dot at the end of a line
	src = strings.Replace(mainSrc, "var label = s.name", "var label = Math.\n  var z = 1", 1)
	got = labels(src, at(src, "Math.", 1, 5))
	has("namespace members", got, "sqrt", "PI")
	// module members
	src = strings.Replace(mainSrc, "Geo.origin", "Geo.", 1)
	got = labels(src, at(src, "Geo.", 1, 4))
	has("module members", got, "origin", "distance", "Point")
	// functions on a type
	src = strings.Replace(mainSrc, "var label = s.name", "var label = Point.", 1)
	got = labels(src, at(src, "Point.", 1, 6))
	has("functions on a type", got, "add")
	// names in scope
	src = strings.Replace(mainSrc, "var label = s.name", "var label = ", 1)
	got = labels(src, at(src, "var label = ", 1, 12))
	has("scope", got, "s", "times", "describe", "first", "Shape", "Point", "dist", "Geo", "Array", "println", "if")
	lacks("scope", got, "distance", "p", "xs")
	// after a pipe: functions whose first parameter takes an array
	src = strings.Replace(mainSrc, "| Array.map({ x * times })", "| ", 1)
	got = labels(src, at(src, "| ", 1, 2))
	has("pipe", got, "Array.map", "Array.filter", "Array.length")
	lacks("pipe", got, "String.trim", "Math.sqrt")
	// after a pipe and a dot: the namespace's functions that take the left side
	src = strings.Replace(mainSrc, "| Array.map({ x * times })", "| Array.", 1)
	got = labels(src, at(src, "| Array.", 1, 8))
	has("piped members", got, "map", "filter")
	lacks("piped members", got, "filled", "repeat")
	// named arguments
	src = strings.Replace(mainSrc, "dist(p, Geo.origin)", "dist(p, ", 1)
	got = labels(src, at(src, "dist(p, ", 1, 8))
	has("named arguments", got, "b:")
	lacks("named arguments", got, "a:")
	// record literal fields
	src = strings.Replace(mainSrc, "Point{ x: 1.0, y: 2.0 }", "Point{ x: 1.0, ", 1)
	got = labels(src, at(src, "Point{ x: 1.0, ", 1, 15))
	has("record fields", got, "y")
	lacks("record fields", got, "x")
	// types after a colon
	src = strings.Replace(mainSrc, "times: int = 1", "times: ", 1)
	got = labels(src, at(src, "times: ", 1, 7))
	has("types", got, "int", "string", "Shape", "Point", "Geo", "Array", "Time", "Result")
	lacks("types", got, "describe", "println")
	// is targets
	src = strings.Replace(mainSrc, "if s.area > 0.0", "if s is ", 1)
	got = labels(src, at(src, "if s is ", 1, 8))
	has("is", got, "null", "Shape", "string")
	// import names
	src = strings.Replace(mainSrc, "{ Point, distance: dist }", "{ Point, ", 1)
	got = labels(src, at(src, "{ Point, ", 1, 9))
	has("import names", got, "distance", "origin")
	lacks("import names", got, "Point")
	// import paths
	src = strings.Replace(mainSrc, `import "./lib/geometry.ant" Geo`, `import "./lib/" Geo`, 1)
	got = labels(src, at(src, `"./lib/`, 1, 7))
	has("import paths", got, "geometry.ant")
	// markup: elements, attributes, end tags
	src = strings.Replace(mainSrc, "if s.area > 0.0 { label } else { \"none\" }", "var h = <div><sp", 1)
	got = labels(src, at(src, "<sp", 1, 3))
	has("elements", got, "span", "section")
	src = strings.Replace(mainSrc, "if s.area > 0.0 { label } else { \"none\" }", "var h = <input ty", 1)
	got = labels(src, at(src, "<input ty", 1, 9))
	has("attributes", got, "type", "class", "value")
	lacks("attributes", got, "onclick")
	src = strings.Replace(mainSrc, "if s.area > 0.0 { label } else { \"none\" }", "var h = <div><p>hi</", 1)
	got = labels(src, at(src, "hi</", 1, 4))
	has("end tag", got, "p>")
	src = strings.Replace(mainSrc, "if s.area > 0.0 { label } else { \"none\" }", "var h = <p>{s.}</p>", 1)
	got = labels(src, at(src, "{s.", 1, 3))
	has("markup interpolation", got, "name", "area")
	// nothing in a comment or a string
	src = strings.Replace(mainSrc, "var label = s.name", "var label = \"s.\"", 1)
	got = labels(src, at(src, "\"s.", 1, 3))
	if len(got) != 0 {
		t.Errorf("completion in a string: %q", got)
	}
}

func TestSignatureHelp(t *testing.T) {
	c := start(t, workspace())
	uri := c.open("main.ant", mainSrc)
	var sh SignatureHelp
	c.call("textDocument/signatureHelp", SignatureHelpParams{TextDocumentPositionParams: c.pos(uri, at(mainSrc, "dist(p, Geo", 1, 8))}, &sh)
	if len(sh.Signatures) != 1 || sh.Signatures[0].Label != "distance(a: Point, b: Point): float64" || sh.ActiveParameter != 1 {
		t.Errorf("signature help: %+v", sh)
	}
	if p := sh.Signatures[0].Parameters; len(p) != 2 || p[1].Label != [2]int{19, 27} {
		t.Errorf("parameters: %+v", p)
	}
	// through a pipe, the active parameter counts from the second
	sh = SignatureHelp{}
	c.call("textDocument/signatureHelp", SignatureHelpParams{TextDocumentPositionParams: c.pos(uri, at(mainSrc, "Array.map(", 1, 10))}, &sh)
	if len(sh.Signatures) != 1 || !strings.HasPrefix(sh.Signatures[0].Label, "Array.map[T, U](xs: Array[T], f: func(T): U)") || sh.ActiveParameter != 1 {
		t.Errorf("piped signature help: %+v", sh)
	}
	// a call being typed
	src := strings.Replace(mainSrc, "dist(p, Geo.origin)", "dist(", 1)
	c.change(uri, 2, src)
	sh = SignatureHelp{}
	c.call("textDocument/signatureHelp", SignatureHelpParams{TextDocumentPositionParams: c.pos(uri, at(src, "dist(", 1, 5))}, &sh)
	if len(sh.Signatures) != 1 || sh.ActiveParameter != 0 {
		t.Errorf("signature help in an open call: %+v", sh)
	}
}

func TestSymbolsFoldingSelection(t *testing.T) {
	c := start(t, workspace())
	uri := c.open("main.ant", mainSrc)
	var syms []DocumentSymbol
	c.call("textDocument/documentSymbol", DocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: uri}}, &syms)
	var names []string
	for _, s := range syms {
		names = append(names, s.Name)
	}
	if strings.Join(names, " ") != "./lib/geometry.ant Geo Shape describe first" {
		t.Errorf("symbols: %v", names)
	}
	if len(syms) < 3 || len(syms[2].Children) != 2 || syms[2].Children[0].Name != "name" {
		t.Errorf("Shape's children: %+v", syms[2])
	}
	var ws []SymbolInformation
	c.call("workspace/symbol", WorkspaceSymbolParams{Query: "dist"}, &ws)
	if len(ws) != 1 || ws[0].Name != "distance" {
		t.Errorf("workspace symbols: %+v", ws)
	}
	var folds []FoldingRange
	c.call("textDocument/foldingRange", FoldingRangeParams{TextDocument: TextDocumentIdentifier{URI: uri}}, &folds)
	if len(folds) < 3 || folds[0].Kind != "imports" || folds[0].StartLine != 0 || folds[0].EndLine != 1 {
		t.Errorf("folds: %+v", folds)
	}
	var sel []SelectionRange
	c.call("textDocument/selectionRange", SelectionRangeParams{TextDocument: TextDocumentIdentifier{URI: uri}, Positions: []Position{at(mainSrc, "s.name", 1, 2)}}, &sel)
	if len(sel) != 1 || sel[0].Parent == nil || sel[0].Parent.Parent == nil {
		t.Errorf("selection ranges: %+v", sel)
	}
	var toks SemanticTokens
	c.call("textDocument/semanticTokens/full", SemanticTokensParams{TextDocument: TextDocumentIdentifier{URI: uri}}, &toks)
	if len(toks.Data)%5 != 0 || len(toks.Data) < 50 {
		t.Errorf("semantic tokens: %d values", len(toks.Data))
	}
	var hints []InlayHint
	c.call("textDocument/inlayHint", InlayHintParams{TextDocument: TextDocumentIdentifier{URI: uri}, Range: Range{End: Position{Line: 100}}}, &hints)
	var labels []string
	for _, h := range hints {
		labels = append(labels, h.Label)
	}
	if strings.Join(labels, " ") != ": string : Point : float64 a: b: : Array[int] f:" {
		t.Errorf("inlay hints: %q", labels)
	}
}

func TestFormatting(t *testing.T) {
	c := start(t, workspace())
	ugly := "func f( a:int ):int{\n    a+1\n}\n"
	uri := c.open("main.ant", ugly)
	var edits []TextEdit
	c.call("textDocument/formatting", DocumentFormattingParams{TextDocument: TextDocumentIdentifier{URI: uri}}, &edits)
	if len(edits) != 1 || apply(ugly, edits[0]) != "func f(a: int): int {\n  a + 1\n}\n" {
		t.Errorf("formatting: %+v", edits)
	}
	c.change(uri, 2, "func f( {")
	edits = nil
	c.call("textDocument/formatting", DocumentFormattingParams{TextDocument: TextDocumentIdentifier{URI: uri}}, &edits)
	if edits != nil {
		t.Errorf("formatting a broken document: %+v", edits)
	}
}

// apply applies one edit to text.
func apply(text string, e TextEdit) string {
	d := newDocument("", 0, text)
	start, end := d.offset(e.Range.Start, true), d.offset(e.Range.End, true)
	return text[:start] + e.NewText + text[end:]
}

func TestCodeActions(t *testing.T) {
	c := start(t, workspace())
	src := strings.Replace(mainSrc, "Point{ x: 1.0, y: 2.0 }", "Point{ x: 1.0 }", 1)
	uri := c.open("main.ant", src)
	diags := c.diagnostics(uri)
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "missing the field y") {
		t.Fatalf("diagnostics: %+v", diags)
	}
	var acts []CodeAction
	p := CodeActionParams{TextDocument: TextDocumentIdentifier{URI: uri}, Range: diags[0].Range}
	p.Context.Diagnostics = diags
	c.call("textDocument/codeAction", p, &acts)
	if len(acts) != 1 || acts[0].Title != "Add the missing field y" {
		t.Fatalf("code actions: %+v", acts)
	}
	fixed := apply(src, acts[0].Edit.Changes[uri][0])
	if !strings.Contains(fixed, "Point{ x: 1.0, y: 0.0 }") {
		t.Errorf("after the fix: %s", fixed)
	}
	// an exported function without a return type
	lib := strings.Replace(geometry, "export func distance(a: Point, b: Point): float64 {", "export func distance(a: Point, b: Point) {", 1)
	libURI := c.open("lib/geometry.ant", lib)
	diags = c.diagnostics(libURI)
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "must state its return type") {
		t.Fatalf("library diagnostics: %+v", diags)
	}
	acts = nil
	p = CodeActionParams{TextDocument: TextDocumentIdentifier{URI: libURI}, Range: diags[0].Range}
	c.call("textDocument/codeAction", p, &acts)
	if len(acts) != 1 || acts[0].Title != "State the return type float64" || !strings.Contains(apply(lib, acts[0].Edit.Changes[libURI][0]), "b: Point): float64 {") {
		t.Errorf("code actions: %+v", acts)
	}
	// a rename that would clash
	if err := c.call("textDocument/rename", RenameParams{TextDocumentPositionParams: c.pos(uri, at(src, "label }", 1, 0)), NewName: "p"}, nil); err == nil || !strings.Contains(err.Message, "already declared") {
		t.Errorf("renaming label to p: %v", err)
	}
}
