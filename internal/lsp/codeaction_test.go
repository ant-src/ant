package lsp

import (
	"strings"
	"testing"
)

// TestCodeActionsInRange: the fixes are those of the diagnostics the range touches, so that on
// the second of two literals on a line the fix edits the second.
func TestCodeActionsInRange(t *testing.T) {
	c := start(t, map[string]string{})
	src := "type P = { x: int, y: int }\nvar ps = [P{ x: 1 }, P{ x: 2 }]\nprintln(ps)\n"
	uri := c.open("main.ant", src)
	c.diagnostics(uri)
	cur := at(src, "P{ x: 2", 1, 0)
	var acts []CodeAction
	c.call("textDocument/codeAction", CodeActionParams{TextDocument: TextDocumentIdentifier{URI: uri}, Range: Range{Start: cur, End: cur}}, &acts)
	if len(acts) != 1 {
		t.Fatalf("code actions: %+v", acts)
	}
	if got := apply(src, acts[0].Edit.Changes[uri][0]); !strings.Contains(got, "[P{ x: 1 }, P{ x: 2, y: 0 }]") {
		t.Errorf("after the fix:\n%s", got)
	}
	// the whole line: a fix for each literal, though the titles are the same
	line := at(src, "var ps", 1, 0)
	acts = nil
	c.call("textDocument/codeAction", CodeActionParams{TextDocument: TextDocumentIdentifier{URI: uri}, Range: Range{Start: line, End: Position{Line: line.Line + 1}}}, &acts)
	if len(acts) != 2 {
		t.Errorf("code actions on the line: %+v", acts)
	}
}

// TestAddFieldsZeroValues: the placeholder of a set, of Bytes, and of Html is a value of the
// type, so that the fixed literal checks.
func TestAddFieldsZeroValues(t *testing.T) {
	c := start(t, map[string]string{})
	src := "type R = { s: Set[int], b: Bytes, h: Html, n: int }\nvar r = R{ n: 1 }\nprintln(r)\n"
	uri := c.open("main.ant", src)
	c.diagnostics(uri)
	cur := at(src, "R{ n", 1, 0)
	var acts []CodeAction
	c.call("textDocument/codeAction", CodeActionParams{TextDocument: TextDocumentIdentifier{URI: uri}, Range: Range{Start: cur, End: cur}}, &acts)
	if len(acts) != 1 {
		t.Fatalf("code actions: %+v", acts)
	}
	fixed := apply(src, acts[0].Edit.Changes[uri][0])
	if !strings.Contains(fixed, "R{ n: 1, s: Set.of([]), b: Bytes.concat([]), h: <></> }") {
		t.Errorf("after the fix:\n%s", fixed)
	}
	c.change(uri, 2, fixed)
	if diags := c.diagnostics(uri); len(diags) != 0 {
		t.Errorf("the fixed literal does not check: %+v", diags)
	}
}
