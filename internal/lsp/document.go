package lsp

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/ant-src/ant/ast"
)

// document is an open document: its text and the line table that maps the protocol's positions
// to byte offsets and the parser's positions.
type document struct {
	uri     string
	path    string // the file's path, for the loader
	version int
	text    string
	lines   []int // the byte offset at which each line starts
}

func newDocument(uri string, version int, text string) *document {
	d := &document{uri: uri, path: uriPath(uri), version: version}
	d.set(text)
	return d
}

func (d *document) set(text string) {
	d.text = text
	d.lines = lineStarts(text)
}

// lineStarts returns the offset of each line's start. Lines end at the line breaks of section 2:
// \n, \r\n, or a lone \r.
func lineStarts(text string) []int {
	lines := []int{0}
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\n':
			lines = append(lines, i+1)
		case '\r':
			if i+1 < len(text) && text[i+1] == '\n' {
				continue // the \n ends the line
			}
			lines = append(lines, i+1)
		}
	}
	return lines
}

// lineEnd is the offset at which line i's text ends, before its line break.
func (d *document) lineEnd(i int) int {
	if i+1 < len(d.lines) {
		end := d.lines[i+1]
		if end > 0 && d.text[end-1] == '\n' {
			end--
		}
		if end > 0 && d.text[end-1] == '\r' {
			end--
		}
		return end
	}
	return len(d.text)
}

// offset converts a protocol position to a byte offset, clamping a position past the end of its
// line or of the document.
func (d *document) offset(p Position, utf16Units bool) int {
	if p.Line < 0 {
		return 0
	}
	if p.Line >= len(d.lines) {
		return len(d.text)
	}
	start, end := d.lines[p.Line], d.lineEnd(p.Line)
	line := d.text[start:end]
	if !utf16Units {
		return start + min(max(p.Character, 0), len(line))
	}
	units := 0
	for i, r := range line {
		if units >= p.Character {
			return start + i
		}
		units += utf16.RuneLen(r)
		if units > p.Character {
			return start + i // inside a surrogate pair: the character's start
		}
	}
	return end
}

// position converts a byte offset to a protocol position.
func (d *document) position(off int, utf16Units bool) Position {
	off = min(max(off, 0), len(d.text))
	line := d.lineOf(off)
	start := d.lines[line]
	if !utf16Units {
		return Position{Line: line, Character: off - start}
	}
	units := 0
	for _, r := range d.text[start:off] {
		units += utf16.RuneLen(r)
	}
	return Position{Line: line, Character: units}
}

// lineOf is the line holding the byte at off.
func (d *document) lineOf(off int) int {
	lo, hi := 0, len(d.lines)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if d.lines[mid] <= off {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo
}

// offsetOf converts a parser position, a 1-based line and byte column, to a byte offset.
func (d *document) offsetOf(p ast.Pos) int {
	if p.Line < 1 || p.Line > len(d.lines) {
		return len(d.text)
	}
	return min(d.lines[p.Line-1]+p.Col-1, len(d.text))
}

// posOf converts a byte offset to a parser position.
func (d *document) posOf(off int) ast.Pos {
	off = min(max(off, 0), len(d.text))
	line := d.lineOf(off)
	return ast.Pos{Line: line + 1, Col: off - d.lines[line] + 1}
}

// apply applies the client's changes in order.
func (d *document) apply(changes []TextDocumentContentChangeEvent, utf16Units bool) {
	for _, ch := range changes {
		if ch.Range == nil {
			d.set(ch.Text)
			continue
		}
		start, end := d.offset(ch.Range.Start, utf16Units), d.offset(ch.Range.End, utf16Units)
		if end < start {
			start, end = end, start
		}
		d.set(d.text[:start] + ch.Text + d.text[end:])
	}
}

// wordAt returns the identifier around off, and where it starts.
func (d *document) wordAt(off int) (string, int) {
	start := off
	for start > 0 && isIdentByte(d.text[start-1]) {
		start--
	}
	end := off
	for end < len(d.text) && isIdentByte(d.text[end]) {
		end++
	}
	return d.text[start:end], start
}

func isIdentByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}

// uriPath turns a file URI into a path, or "" for another scheme, whose document the server
// does not analyze.
func uriPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return ""
	}
	p := u.Path
	if runtime.GOOS == "windows" {
		p = strings.TrimPrefix(p, "/")
		p = filepath.FromSlash(p)
	}
	return filepath.Clean(p)
}

// fileURI turns a path into a file URI.
func fileURI(path string) string {
	if strings.Contains(path, "://") || strings.HasPrefix(path, "ant:") {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = filepath.ToSlash(abs)
	if !strings.HasPrefix(abs, "/") {
		abs = "/" + abs
	}
	u := url.URL{Scheme: "file", Path: abs}
	return u.String()
}

// validUTF8 says whether the text is UTF-8, which a document from the client always is.
func validUTF8(s string) error {
	if !utf8.ValidString(s) {
		return fmt.Errorf("the document is not UTF-8")
	}
	return nil
}
