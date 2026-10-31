// Package diag finds what a diagnostic is about in its source, and prints it as ant check does:
// the position and the message, then the lines it covers with the span underlined.
package diag

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/width"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/grapheme"
	"github.com/ant-src/ant/internal/highlight"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/parser"
)

// LineStarts returns the offset of each line's start. Lines end at the line breaks of section 2:
// \n, \r\n, or a lone \r.
func LineStarts(text string) []int {
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

// Offset converts a parser position, a 1-based line and byte column, to a byte offset into text,
// whose line starts are lines.
func Offset(text string, lines []int, p ast.Pos) int {
	if p.Line < 1 || p.Line > len(lines) {
		return len(text)
	}
	return min(lines[p.Line-1]+p.Col-1, len(text))
}

// Span is the byte span of what a diagnostic at pos is about (spec/lsp.md section 3): the token
// there when it is a keyword or an operator, as in a syntax error; else the innermost expression
// or declaration that starts there, or whose operator stands there; else the token or word
// there. f and toks are the file's tree and tokens, either of which may be nil.
func Span(f *ast.File, toks []token.Token, text string, lines []int, pos ast.Pos) (from, to int) {
	if pos.Line == 0 {
		return 0, 0
	}
	var tok *token.Token
	for i := range toks {
		t := &toks[i]
		if t.Pos.Line == pos.Line && t.Pos.Col == pos.Col && t.Kind != token.Newline && t.Kind != token.EOF {
			tok = t
			break
		}
	}
	// The innermost node that starts at the position, or whose operator stands there: an error
	// on a field read or a binary operation is reported at the dot or the operator and covers
	// the whole expression.
	var found ast.Node
	if f != nil {
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
		return Offset(text, lines, ast.Start(found)), Offset(text, lines, found.EndPosition())
	}
	off := Offset(text, lines, pos)
	if tok != nil {
		return off, max(off, tok.End)
	}
	// the word at the position, as for an error inside markup or a comment
	end := off
	for end < len(text) && token.IsIdentByte(text[end]) {
		end++
	}
	if end == off && off < len(text) && !strings.ContainsRune("\r\n", rune(text[off])) {
		_, n := utf8.DecodeRuneInString(text[off:])
		end = off + n
	}
	return off, end
}

// operatorHeaded reports whether a node's Position is its operator rather than its start.
func operatorHeaded(n ast.Node) bool {
	switch n.(type) {
	case *ast.Field, *ast.Index, *ast.Call, *ast.Binary, *ast.Is, *ast.Assign:
		return true
	}
	return false
}

// A Source is the text of a file, parsed for spans and lexed for colors when first needed.
type Source struct {
	Path, Text string
	lines      []int
	parsed     bool
	file       *ast.File
	toks       []token.Token
	cls        []highlight.Class
}

// NewSource returns the source of the file at path.
func NewSource(path, text string) *Source {
	return &Source{Path: path, Text: text, lines: LineStarts(text)}
}

// Span is the byte span of what a diagnostic at pos is about.
func (s *Source) Span(pos ast.Pos) (from, to int) {
	if !s.parsed {
		s.parsed = true
		var layout *parser.Layout
		s.file, layout, _ = parser.ParseRecover(s.Path, s.Text)
		if layout != nil {
			s.toks = layout.Tokens
		}
	}
	return Span(s.file, s.toks, s.Text, s.lines, pos)
}

func (s *Source) classes() []highlight.Class {
	if s.cls == nil {
		s.cls = highlight.FileClasses(s.Text)
	}
	return s.cls
}

// line is the text of the 0-based line i, without its line break, and its start.
func (s *Source) line(i int) (string, int) {
	if i < 0 || i >= len(s.lines) {
		return "", len(s.Text)
	}
	start, end := s.lines[i], len(s.Text)
	if i+1 < len(s.lines) {
		end = s.lines[i+1]
	}
	return strings.TrimRight(s.Text[start:end], "\r\n"), start
}

// lineOf is the 0-based line holding the byte at off.
func (s *Source) lineOf(off int) int {
	lo, hi := 0, len(s.lines)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if s.lines[mid] <= off {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo
}

// A Printer prints diagnostics with the code they are about.
type Printer struct {
	// Color colors the output with ANSI escapes: the code as the REPL highlights it.
	Color bool
	// Read returns the text of the file at path; a diagnostic in a file it cannot read is
	// printed without code.
	Read    func(path string) (string, error)
	sources map[string]*Source
}

const (
	bold  = "\x1b[1m"
	red   = "\x1b[1;31m"
	faint = "\x1b[90m"
	reset = highlight.Reset
	tab   = 4 // the columns between tab stops
	// most lines a span shows in full; a longer one shows its first two lines and its last
	most = 4
)

func (p *Printer) style(sb *strings.Builder, style, s string) {
	if p.Color {
		sb.WriteString(style + s + reset)
	} else {
		sb.WriteString(s)
	}
}

func (p *Printer) source(path string) *Source {
	if s, ok := p.sources[path]; ok {
		return s
	}
	var s *Source
	if p.Read != nil {
		if text, err := p.Read(path); err == nil {
			s = NewSource(path, text)
		}
	}
	if p.sources == nil {
		p.sources = map[string]*Source{}
	}
	p.sources[path] = s
	return s
}

// Print writes the diagnostic msg at pos in the file at path: a first line as the compilers of
// the terminal print one, path:line:col: msg, then the lines of the code it is about, numbered,
// with what it covers underlined, then an empty line.
func (p *Printer) Print(w io.Writer, path string, pos ast.Pos, msg string) {
	var sb strings.Builder
	switch {
	case path == "":
	case pos.Line == 0:
		p.style(&sb, bold, path+":")
		sb.WriteString(" ")
	default:
		p.style(&sb, bold, fmt.Sprintf("%s:%d:%d:", path, pos.Line, pos.Col))
		sb.WriteString(" ")
	}
	p.style(&sb, bold, msg)
	sb.WriteString("\n")
	if s := p.source(path); s != nil && pos.Line > 0 && pos.Line <= len(s.lines) {
		from, to := s.Span(pos)
		p.snippet(&sb, s, from, to)
	}
	sb.WriteString("\n")
	io.WriteString(w, sb.String())
}

// Summary writes how many errors there were.
func (p *Printer) Summary(w io.Writer, n int) {
	var sb strings.Builder
	noun := "errors"
	if n == 1 {
		noun = "error"
	}
	p.style(&sb, red, fmt.Sprintf("%d %s", n, noun))
	sb.WriteString("\n")
	io.WriteString(w, sb.String())
}

// snippet writes the lines that the span [from, to) of s covers, each underlined where the span
// covers it; an empty span marks the one column where it stands. A span on a blank line, as at the
// end of the file, shows the last line before it with code, as context.
func (p *Printer) snippet(sb *strings.Builder, s *Source, from, to int) {
	first, last := s.lineOf(from), s.lineOf(max(from, to-1))
	var shown []int
	if text, _ := s.line(first); first == last && strings.TrimSpace(text) == "" {
		for i := first - 1; i >= 0; i-- {
			if text, _ := s.line(i); strings.TrimSpace(text) != "" {
				shown = append(shown, i)
				break
			}
		}
	}
	for i := first; i <= last; i++ {
		if last-first+1 > most && i == first+2 {
			shown = append(shown, -1) // the elided lines
			i = last
		}
		shown = append(shown, i)
	}
	gutter := len(strconv.Itoa(last + 1))
	margin := func(label string) {
		p.style(sb, faint, fmt.Sprintf(" %*s │", gutter, label))
	}
	margin("")
	sb.WriteString("\n")
	for _, i := range shown {
		if i < 0 {
			margin("⋮")
			sb.WriteString("\n")
			continue
		}
		text, start := s.line(i)
		margin(strconv.Itoa(i + 1))
		if text != "" {
			sb.WriteString(" ")
			p.code(sb, s, text, start)
		}
		sb.WriteString("\n")
		if i < first {
			continue // context
		}
		// The part of this line the span covers: on a line after the first, from the code after
		// the indentation; on a line before the last, to the end of the code.
		a, b := max(from-start, 0), min(to-start, len(text))
		if i > first {
			a = len(text) - len(strings.TrimLeft(text, " \t"))
		}
		if i < last {
			b = len(strings.TrimRight(text, " \t"))
		}
		if a > len(text) {
			a = len(text)
		}
		if b <= a && from != to && first != last {
			continue // a blank line inside the span
		}
		margin("")
		sb.WriteString(" " + strings.Repeat(" ", columns(text[:a])))
		p.style(sb, red, strings.Repeat("^", max(columns(text[a:max(a, b)]), 1)))
		sb.WriteString("\n")
	}
}

// code writes a line of the source, at byte start of it, with its tabs expanded, highlighted
// when the printer colors.
func (p *Printer) code(sb *strings.Builder, s *Source, text string, start int) {
	for i := 0; i <= len(text); {
		j := strings.IndexByte(text[i:], '\t')
		if j < 0 {
			j = len(text)
		} else {
			j += i
		}
		if p.Color {
			highlight.Paint(sb, text[i:j], s.classes(), start+i, highlight.ANSI[:])
		} else {
			sb.WriteString(text[i:j])
		}
		if j < len(text) {
			sb.WriteString(strings.Repeat(" ", tab-columns(text[:j])%tab))
		}
		i = j + 1
	}
}

// columns is how many terminal columns text takes at the start of a line, its tabs expanded.
func columns(text string) int {
	col := 0
	for i := 0; i < len(text); {
		if text[i] == '\t' {
			col += tab - col%tab
			i++
			continue
		}
		n := max(grapheme.First(text[i:]), 1)
		col += clusterWidth(text[i : i+n])
		i += n
	}
	return col
}

// clusterWidth is the columns a grapheme cluster takes: two for a wide character or an emoji
// shown as one, else one.
func clusterWidth(c string) int {
	r, _ := utf8.DecodeRuneInString(c)
	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	}
	if strings.ContainsRune(c, '️') {
		return 2
	}
	return 1
}
