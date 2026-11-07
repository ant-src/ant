package lsp

import "testing"

// TestScanMarkupStrings checks that the strings inside an interpolation, with their own
// interpolations, quotes, and braces, do not move the scanner out of step with the markup.
func TestScanMarkupStrings(t *testing.T) {
	for _, c := range []struct {
		src  string
		kind markupKind
	}{
		{`<p>{"a ${b ? "{" : c}"}</p><a hr`, inAttrName},
		{`<p>{"a \" ${"}"} }"}<a `, inAttrs},
		{"<p>{`raw \\ ${ \"` + x}<b", inTagName},
		{`<p>{"open ${f(`, inInterp},
	} {
		if got := scanMarkup(c.src).kind; got != c.kind {
			t.Errorf("%s: kind %d, want %d", c.src, got, c.kind)
		}
	}
}

// TestCompleteNameRanges: an item in markup or an import path replaces the whole name being
// written, its - . and / included, which an editor's word may leave out.
func TestCompleteNameRanges(t *testing.T) {
	c := start(t, map[string]string{"util.ant": "export func f(): int { 1 }\n", "lib/geo.ant": "export var g = 1\n"})
	for _, tc := range []struct {
		src, needle    string // the cursor is after needle
		label, replace string // the item, and the text its edit replaces
	}{
		{"var h = <meta http-e />\n", "http-e", "http-equiv", "http-e"},
		{"var h = <my-widget></my-w\n", "</my-w", "my-widget>", "my-w"},
		{"var h = <di\n", "<di", "div", "di"},
		{"var h = <p cl\n", "<p cl", "class", "cl"},
		{"import \"./util.a\n", "./util.a", "util.ant", "util.a"},
		{"import \"./li\n", "./li", "lib/", "li"},
		{"import \"./lib/ge\n", "./lib/ge", "geo.ant", "ge"},
		{"import \"\n", "\"", "./", ""},
	} {
		uri := c.open("main.ant", tc.src)
		pos := at(tc.src, tc.needle, 1, len(tc.needle))
		var list CompletionList
		c.call("textDocument/completion", CompletionParams{TextDocumentPositionParams: c.pos(uri, pos)}, &list)
		var item *CompletionItem
		for i := range list.Items {
			if list.Items[i].Label == tc.label {
				item = &list.Items[i]
			}
		}
		if item == nil || item.TextEdit == nil {
			t.Errorf("%q: no %s with an edit: %+v", tc.src, tc.label, item)
			continue
		}
		want := Range{Start: at(tc.src, tc.needle, 1, len(tc.needle)-len(tc.replace)), End: pos}
		if item.TextEdit.Range != want || item.TextEdit.NewText != tc.label {
			t.Errorf("%q: %s edits %+v to %q, want %+v", tc.src, tc.label, item.TextEdit.Range, item.TextEdit.NewText, want)
		}
		c.notify("textDocument/didClose", DidCloseTextDocumentParams{TextDocument: TextDocumentIdentifier{URI: uri}})
	}
}
