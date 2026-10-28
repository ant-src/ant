package lsp

import (
	"github.com/ant-src/ant/internal/format"
)

// formatting answers textDocument/formatting with ant fmt (spec/lsp.md section 4): one edit that
// replaces what changed, or nothing when the document does not parse or is formatted already.
func (s *Server) formatting(p *DocumentFormattingParams) []TextEdit {
	d := s.doc(p.TextDocument.URI)
	if d == nil {
		return nil
	}
	out, err := format.Source(d.path, []byte(d.text))
	if err != nil || out == nil {
		return nil
	}
	formatted := string(out)
	if formatted == d.text {
		return []TextEdit{}
	}
	// Keep the common prefix and suffix, so the cursor and the folds around the change stay.
	prefix := 0
	for prefix < len(d.text) && prefix < len(formatted) && d.text[prefix] == formatted[prefix] {
		prefix++
	}
	// back up to a line start, so the edit replaces whole lines
	for prefix > 0 && d.text[prefix-1] != '\n' {
		prefix--
	}
	suffix := 0
	for suffix < len(d.text)-prefix && suffix < len(formatted)-prefix &&
		d.text[len(d.text)-1-suffix] == formatted[len(formatted)-1-suffix] {
		suffix++
	}
	// the kept suffix begins at a line start, after a line break, so no \r\n is split
	for suffix > 0 && d.text[len(d.text)-suffix-1] != '\n' {
		suffix--
	}
	return []TextEdit{{
		Range:   Range{Start: d.position(prefix, s.utf16), End: d.position(len(d.text)-suffix, s.utf16)},
		NewText: formatted[prefix : len(formatted)-suffix],
	}}
}
