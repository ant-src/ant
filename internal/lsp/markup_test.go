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
