// Package lexer turns Ant source into tokens, applying the line-ending rule of section 2.
package lexer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/value"
)

// Error is a lexical error.
type Error struct {
	Pos token.Pos
	Msg string
}

func (e *Error) Error() string { return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Col, e.Msg) }

type lexer struct {
	src     string
	pos     int
	line    int
	col     int
	toks    []token.Token
	nesting int             // template interpolations entered, bounded by maxNesting
	noRegex bool            // read / as the operator always, for the differential test of the operand rule
	lines   []int           // the byte offset at which each line starts
	layout  bool            // keep the comments, for the formatter
	comment []token.Comment // the comments kept in layout mode, in source order
}

// maxNesting bounds the nesting of template interpolations, so that no input can overflow the
// Go stack in the lexer.
const maxNesting = 1000

// Lex tokenizes src. The token stream ends with a Newline (unless empty) and an EOF.
func Lex(src string) ([]token.Token, error) { return LexWith(src, true) }

// LexWith tokenizes src as Lex does. With regex false, a / never begins a regular expression
// literal, which gives the lexer as it was before the literal existed; the fuzz test compares
// the two on every input in which the literal does not occur.
func LexWith(src string, regex bool) (toks []token.Token, err error) {
	toks, _, err = lex(&lexer{src: src, noRegex: !regex})
	return toks, err
}

// LexLayout tokenizes src as Lex does, and also returns its comments in source order, for the
// formatter, which must print them where they were.
func LexLayout(src string) ([]token.Token, []token.Comment, error) {
	return lex(&lexer{src: src, layout: true})
}

func lex(l *lexer) (toks []token.Token, comments []token.Comment, err error) {
	l.line, l.col, l.lines = 1, 1, []int{0}
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*Error); ok {
				toks, comments, err = nil, nil, e
				return
			}
			panic(r)
		}
	}()
	for l.scan(l.toplevel) {
	}
	if len(l.toks) > 0 && l.toks[len(l.toks)-1].Kind != token.Newline {
		l.emit(token.Token{Kind: token.Newline, Pos: l.here()})
	}
	l.emit(token.Token{Kind: token.EOF, Pos: l.here()})
	return l.toks, l.comment, nil
}

func (l *lexer) fail(pos token.Pos, format string, args ...any) {
	panic(&Error{Pos: pos, Msg: fmt.Sprintf(format, args...)})
}

func (l *lexer) here() token.Pos { return token.Pos{Line: l.line, Col: l.col} }

func (l *lexer) peek(k int) byte {
	if l.pos+k < len(l.src) {
		return l.src[l.pos+k]
	}
	return 0
}

func (l *lexer) has(s string) bool { return strings.HasPrefix(l.src[l.pos:], s) }

func (l *lexer) advance(n int) {
	for i := 0; i < n && l.pos < len(l.src); i++ {
		if l.src[l.pos] == '\n' {
			l.line++
			l.col = 1
			l.lines = append(l.lines, l.pos+1)
		} else {
			l.col++
		}
		l.pos++
	}
}

// emit appends t, which starts at t.Pos and ends where the lexer stands.
func (l *lexer) emit(t token.Token) {
	t.Off, t.End = l.offset(t.Pos), l.pos
	if t.Kind == token.Newline {
		t.End = t.Off
	}
	l.toks = append(l.toks, t)
}

// offset is the byte offset of a position the lexer has passed.
func (l *lexer) offset(p token.Pos) int { return l.lines[p.Line-1] + p.Col - 1 }

// keepComment records the comment from start to where the lexer stands, in layout mode.
func (l *lexer) keepComment(pos token.Pos, start int) {
	if l.layout {
		l.comment = append(l.comment, token.Comment{Text: l.src[start:l.pos], Pos: pos, Off: start, End: l.pos})
	}
}

// Keywords that can end an expression; every other keyword continues the line.
var endingKeywords = map[string]bool{
	"return": true, "break": true, "continue": true, "true": true, "false": true, "null": true,
	"void": true,
}

// lastContinues reports whether the previous token cannot end an expression, so that a line
// break after it is not a statement end.
func (l *lexer) lastContinues() bool {
	if len(l.toks) == 0 {
		return true
	}
	t := l.toks[len(l.toks)-1]
	switch t.Kind {
	case token.Newline:
		return true
	case token.Op:
		return t.Text != ")" && t.Text != "]" && t.Text != "}" && t.Text != ";" && t.Text != "..."
	case token.Keyword:
		return !endingKeywords[t.Text]
	}
	return false
}

// operandKeywords are the keywords that are operands, after which a / is division.
var operandKeywords = map[string]bool{"true": true, "false": true, "null": true, "void": true}

// Operand reports whether t is an operand in the sense of section 2: an identifier, a literal of
// any kind, true, false, null, void, or a closing bracket. After an operand a / is division.
func Operand(t token.Token) bool {
	switch t.Kind {
	case token.Ident, token.Int, token.Float, token.String, token.Template, token.Regex:
		return true
	case token.Keyword:
		return operandKeywords[t.Text]
	case token.Op:
		return t.Text == ")" || t.Text == "]" || t.Text == "}"
	}
	return false
}

// regexAllowed reports whether a / here begins a regular expression literal: when the last
// token, skipping line breaks, is not an Operand (section 2). With no previous token, at the
// start of a file or of an interpolation, a / begins a literal.
func (l *lexer) regexAllowed() bool {
	for i := len(l.toks) - 1; i >= 0; i-- {
		if t := l.toks[i]; t.Kind != token.Newline {
			return !Operand(t)
		}
	}
	return true
}

// nextLineContinues reports whether the next non-blank, non-comment text starts with ".",
// "?.", "|", or "else", which continues the expression (section 2).
func (l *lexer) nextLineContinues() bool {
	i := l.pos
	src := l.src
	for i < len(src) {
		switch {
		case src[i] == ' ' || src[i] == '\t' || src[i] == '\r' || src[i] == '\n':
			i++
		case strings.HasPrefix(src[i:], "//"):
			j := strings.IndexByte(src[i:], '\n')
			if j < 0 {
				return false
			}
			i += j
		case strings.HasPrefix(src[i:], "/*"):
			j := strings.Index(src[i+2:], "*/")
			if j < 0 {
				return false
			}
			i += 2 + j + 2
		default:
			rest := src[i:]
			if strings.HasPrefix(rest, "...") || strings.HasPrefix(rest, "||") {
				return false
			}
			if strings.HasPrefix(rest, ".") || strings.HasPrefix(rest, "?.") || strings.HasPrefix(rest, "|") {
				return true
			}
			if strings.HasPrefix(rest, "else") && (len(rest) == 4 || !isIdentChar(rest[4])) {
				return true
			}
			return false
		}
	}
	return false
}

// toplevel is the default mode: it never stops before the end of the source.
func (l *lexer) toplevel(t token.Token) bool { return false }

// scan lexes one token, or whitespace and comments. It returns false at the end of the source
// or when stop reports that the token just scanned ends the current stream.
func (l *lexer) scan(stop func(token.Token) bool) bool {
	if l.pos >= len(l.src) {
		return false
	}
	c := l.src[l.pos]
	pos := l.here()
	switch {
	case c == '\n':
		if !l.lastContinues() && !l.nextLineContinues() {
			l.emit(token.Token{Kind: token.Newline, Pos: pos})
		}
		l.advance(1)
		return true
	case c == ' ' || c == '\t' || c == '\r':
		l.advance(1)
		return true
	case l.has("//"):
		start := l.pos
		for l.pos < len(l.src) && l.src[l.pos] != '\n' {
			l.advance(1)
		}
		l.keepComment(pos, start)
		return true
	case l.has("/*"):
		end := strings.Index(l.src[l.pos+2:], "*/")
		if end < 0 {
			l.fail(pos, "unterminated block comment")
		}
		start := l.pos
		l.advance(2 + end + 2)
		l.keepComment(pos, start)
		return true
	case isDigit(c):
		l.number()
	case isLetter(c) || c == '_':
		l.identifier()
	case c == '"':
		l.string()
	case c == '`':
		l.template()
	case c == '.' && isDigit(l.peek(1)):
		l.fail(pos, "a float literal needs a digit before the point")
	case c == '/' && !l.noRegex && l.regexAllowed():
		l.regex()
	default:
		l.operator()
	}
	return !stop(l.toks[len(l.toks)-1])
}

// regex scans a regular expression literal at its opening /: the source up to the closing /,
// and the flags after it. A \ escapes the next character, so \/ does not end the literal, and
// neither does a / inside a character class; the literal stays on one line. The source is kept
// as written, and value.NewRegex normalizes it when the parser builds the pattern.
func (l *lexer) regex() {
	pos := l.here()
	l.advance(1)
	start := l.pos
	inClass := false
	for {
		c := l.peek(0)
		switch {
		case c == 0 || c == '\n':
			l.fail(pos, "unterminated regular expression")
		case c == '\\':
			if n := l.peek(1); n == 0 || n == '\n' {
				l.fail(pos, "unterminated regular expression")
			}
			l.advance(2)
			continue
		case c == '[':
			if inClass && l.peek(1) == ':' {
				// A POSIX name such as [:alpha:], whose ] does not end the class.
				l.advance(2)
				for l.peek(0) != 0 && l.peek(0) != '\n' && !(l.peek(0) == ':' && l.peek(1) == ']') {
					l.advance(1)
				}
				if l.peek(0) == ':' {
					l.advance(2)
				}
				continue
			}
			inClass = true
		case c == ']':
			inClass = false
		case c == '/' && !inClass:
			source := l.src[start:l.pos]
			l.advance(1)
			flagPos := l.here()
			flagStart := l.pos
			for isIdentChar(l.peek(0)) {
				l.advance(1)
			}
			flags, err := value.RegexFlags(l.src[flagStart:l.pos])
			if err != nil {
				l.fail(flagPos, "%s", err)
			}
			l.emit(token.Token{Kind: token.Regex, Text: source, Flags: flags, Pos: pos})
			return
		}
		l.advance(1)
	}
}

func (l *lexer) operator() {
	pos := l.here()
	for _, op := range []string{"...", "?.", "??", "==", "!=", "<=", ">=", "&&", "||", "+=", "-=", "*=", "/=", "%="} {
		if l.has(op) {
			l.advance(len(op))
			l.emit(token.Token{Kind: token.Op, Text: op, Pos: pos})
			return
		}
	}
	c := l.src[l.pos]
	if strings.IndexByte("+-*/%<>=!|.,:;()[]{}", c) >= 0 {
		l.advance(1)
		l.emit(token.Token{Kind: token.Op, Text: string(c), Pos: pos})
		return
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.pos:])
	l.fail(pos, "unexpected character %q", r)
}

func isDigit(c byte) bool  { return c >= '0' && c <= '9' }
func isLetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
func isHex(c byte) bool    { return isDigit(c) || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' }
func isBinary(c byte) bool { return c == '0' || c == '1' }
func isIdentChar(c byte) bool {
	return isLetter(c) || isDigit(c) || c == '_'
}

// digits scans digit { ["_"] digit }: at least one digit, underscores only between digits.
func (l *lexer) digits(allowed func(byte) bool, what string) {
	if !allowed(l.peek(0)) {
		l.fail(l.here(), "malformed %s literal", what)
	}
	l.advance(1)
	for {
		switch {
		case allowed(l.peek(0)):
			l.advance(1)
		case l.peek(0) == '_' && allowed(l.peek(1)):
			l.advance(2)
		default:
			return
		}
	}
}

func (l *lexer) number() {
	pos := l.here()
	start := l.pos
	kind := token.Int
	base := 10
	switch {
	case l.has("0x") || l.has("0X"):
		l.advance(2)
		base = 16
		l.digits(isHex, "hexadecimal")
	case l.has("0b") || l.has("0B"):
		l.advance(2)
		base = 2
		l.digits(isBinary, "binary")
	default:
		l.digits(isDigit, "number")
		if l.peek(0) == '.' {
			if !isDigit(l.peek(1)) {
				l.fail(l.here(), "a float literal needs a digit after the point")
			}
			kind = token.Float
			l.advance(1)
			l.digits(isDigit, "number")
		}
		if l.peek(0) == 'e' || l.peek(0) == 'E' {
			kind = token.Float
			l.advance(1)
			if l.peek(0) == '+' || l.peek(0) == '-' {
				l.advance(1)
			}
			l.digits(isDigit, "exponent")
		}
	}
	if isIdentChar(l.peek(0)) {
		l.fail(l.here(), "malformed number literal")
	}
	text := l.src[start:l.pos]
	clean := strings.ReplaceAll(text, "_", "")
	t := token.Token{Kind: kind, Text: text, Pos: pos}
	if kind == token.Int {
		if base != 10 {
			clean = clean[2:]
		}
		v, err := strconv.ParseInt(clean, base, 64)
		if err != nil {
			l.fail(pos, "integer literal %s does not fit an int", text)
		}
		t.Value = v
	} else {
		v, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			l.fail(pos, "float literal %s is out of range", text)
		}
		t.Float = v
	}
	l.emit(t)
}

func (l *lexer) identifier() {
	pos := l.here()
	start := l.pos
	for isIdentChar(l.peek(0)) {
		l.advance(1)
	}
	word := l.src[start:l.pos]
	kind := token.Ident
	afterDot := len(l.toks) > 0 && l.toks[len(l.toks)-1].Kind == token.Op &&
		(l.toks[len(l.toks)-1].Text == "." || l.toks[len(l.toks)-1].Text == "?.")
	switch {
	case afterDot:
		// A field name: any word is a name after a dot (section 2).
	case token.IsReserved(word):
		kind = token.Reserved
	case token.IsKeyword(word):
		kind = token.Keyword
	}
	l.emit(token.Token{Kind: kind, Text: word, Pos: pos})
}

// escape is called at a backslash; it validates the escape, steps over it, and returns the
// character it denotes.
func (l *lexer) escape() string {
	pos := l.here()
	l.advance(1)
	c := l.peek(0)
	switch c {
	case 'n':
		l.advance(1)
		return "\n"
	case 't':
		l.advance(1)
		return "\t"
	case 'r':
		l.advance(1)
		return "\r"
	case '\\', '"', '`', '$':
		l.advance(1)
		return string(c)
	case 'u':
		if l.peek(1) != '{' {
			l.fail(pos, "bad escape: \\u needs a {...} code")
		}
		l.advance(2)
		start := l.pos
		for isHex(l.peek(0)) {
			l.advance(1)
		}
		if l.pos == start || l.peek(0) != '}' {
			l.fail(pos, "bad escape: \\u{...} needs hexadecimal digits")
		}
		v, err := strconv.ParseUint(l.src[start:l.pos], 16, 32)
		if err != nil || v > 0x10FFFF || (v >= 0xD800 && v <= 0xDFFF) {
			l.fail(pos, "bad escape: \\u{%s} is not a Unicode scalar value", l.src[start:l.pos])
		}
		l.advance(1)
		return string(rune(v))
	}
	if c == 0 {
		l.fail(pos, "unterminated string")
	}
	l.fail(pos, "bad escape '\\%c'", c)
	return ""
}

func (l *lexer) string() {
	pos := l.here()
	l.advance(1)
	var sb strings.Builder
	for {
		c := l.peek(0)
		switch {
		case c == 0 || c == '\n':
			l.fail(pos, "unterminated string")
		case c == '"':
			l.advance(1)
			l.emit(token.Token{Kind: token.String, Text: sb.String(), Pos: pos})
			return
		case c == '\\':
			sb.WriteString(l.escape())
		default:
			sb.WriteByte(c)
			l.advance(1)
		}
	}
}

func (l *lexer) template() {
	pos := l.here()
	l.advance(1)
	var parts []token.Part
	var sb strings.Builder
	textPos := l.here()
	flush := func() {
		if sb.Len() > 0 {
			parts = append(parts, token.Part{Text: sb.String(), Pos: textPos, Off: l.offset(textPos), End: l.pos})
			sb.Reset()
		}
	}
	for {
		c := l.peek(0)
		switch {
		case c == 0:
			l.fail(pos, "unterminated template string")
		case c == '`':
			flush()
			l.advance(1)
			l.emit(token.Token{Kind: token.Template, Parts: parts, Pos: pos})
			return
		case c == '\\':
			sb.WriteString(l.escape())
		case c == '$' && l.peek(1) == '{':
			flush()
			exprPos := l.here()
			l.advance(2)
			toks := l.interpolation(exprPos)
			parts = append(parts, token.Part{Tokens: toks, Pos: exprPos, Off: l.offset(exprPos), End: l.pos})
			textPos = l.here()
		default:
			sb.WriteByte(c)
			l.advance(1)
		}
	}
}

// interpolation lexes the tokens of a ${...} up to its matching brace, which it consumes.
func (l *lexer) interpolation(pos token.Pos) []token.Token {
	l.nesting++
	if l.nesting > maxNesting {
		l.fail(pos, "nesting deeper than %d levels", maxNesting)
	}
	defer func() { l.nesting-- }()
	outer := l.toks
	l.toks = nil
	depth := 0
	closed := false
	stop := func(t token.Token) bool {
		if t.Kind != token.Op {
			return false
		}
		switch t.Text {
		case "{":
			depth++
		case "}":
			if depth == 0 {
				closed = true
				return true
			}
			depth--
		}
		return false
	}
	for l.scan(stop) {
	}
	if !closed {
		l.fail(pos, "unterminated ${ in template string")
	}
	toks := l.toks[:len(l.toks)-1] // drop the closing brace
	toks = append(toks, token.Token{Kind: token.EOF, Pos: l.here()})
	l.toks = outer
	return toks
}
