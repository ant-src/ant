package lexer

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/value"
)

// markupAllowed reports whether a < here begins an HTML expression (section 2): when the last
// token, skipping line breaks, is not an Operand, or when the < is the first token of a line
// whose line break ended a statement outside ( and [. Such a < must be followed by an ASCII
// letter or by >. After no operand anything else is an error, since a comparison needs a left
// operand. At the start of a line it is a comparison: the lexer cannot tell a block's { from a
// record or map literal's, inside which a line break ends nothing and < n continues the line.
func (l *lexer) markupAllowed() bool {
	afterOperand := false
	for i := len(l.toks) - 1; i >= 0; i-- {
		if t := l.toks[i]; t.Kind != token.Newline {
			afterOperand = Operand(t)
			break
		}
	}
	tag := isLetter(l.peek(1)) || l.peek(1) == '>'
	if afterOperand {
		lineStart := len(l.toks) > 0 && l.toks[len(l.toks)-1].Kind == token.Newline &&
			(len(l.brackets) == 0 || l.brackets[len(l.brackets)-1] == '{')
		return lineStart && tag
	}
	if !tag {
		l.fail(l.here(), "a comparison needs a left operand; an HTML element starts with a letter after <")
	}
	return true
}

// markupContext is what the content being read is inside of.
type markupContext struct {
	foreign  bool   // inside <svg> or <math>, outside <foreignObject>
	preserve bool   // inside <pre> or <textarea>, where white space is kept
	textOnly string // the <title> or <textarea> outside foreign content that the content is inside of
}

// markup scans an HTML expression at its <: an element or a fragment, as one Html token.
func (l *lexer) markup() {
	pos := l.here()
	var m *token.Markup
	if l.peek(1) == '>' {
		l.advance(2)
		m = l.content(pos, "", markupContext{})
	} else {
		m = &token.Markup{Nodes: []token.MarkupNode{l.element(markupContext{})}}
	}
	l.emit(token.Token{Kind: token.Html, Markup: m, Pos: pos})
}

// element scans an element at its <.
func (l *lexer) element(ctx markupContext) token.MarkupNode {
	pos := l.here()
	start := l.pos
	l.nesting++
	if l.nesting > maxNesting {
		l.fail(pos, "nesting deeper than %d levels", maxNesting)
	}
	defer func() { l.nesting-- }()
	l.advance(1)
	name := l.elementName()
	lower := strings.ToLower(name)
	if c := l.peek(0); !l.atEnd(0) && !isMarkupSpace(c) && c != '/' && c != '>' {
		l.soft(l.here(), "an element name is ASCII letters, digits, and -; unexpected %s after <%s", l.describe(), name)
	}
	switch {
	case value.HtmlObsolete[lower]:
		l.soft(pos, "<%s> is obsolete and not allowed", name)
	case lower == "script" && ctx.foreign:
		l.soft(pos, "<script> is not allowed inside <svg> or <math>")
	case ctx.textOnly != "":
		l.soft(pos, "<%s> holds text only, so it cannot contain <%s>", ctx.textOnly, name)
	}
	l.open = append(l.open, name)
	defer func() { l.open = l.open[:len(l.open)-1] }()
	n := token.MarkupNode{Kind: token.MarkupElement, Name: name, Pos: pos}
	n.Attrs = l.attributes(name)
	inner := ctx
	switch lower {
	case "svg", "math":
		inner.foreign = true
	case "foreignobject":
		inner.foreign = false
	}
	if !ctx.foreign && value.HtmlTextOnly[lower] {
		inner.textOnly = name
		n.TextOnly = true
	}
	if l.has("/>") {
		l.advance(2)
		n.Children = &token.Markup{}
		n.Off, n.End, n.EndPos = start, l.pos, l.here()
		return n
	}
	if l.atEnd(0) {
		// an unterminated start tag, reported by attributes when recovering: the element ends here
		n.Children = &token.Markup{}
		n.Off, n.End, n.EndPos = start, l.pos, l.here()
		return n
	}
	if l.peek(0) == '>' {
		l.advance(1)
	}
	// else the start tag was never closed, reported by attributes when recovering, and what
	// follows is read as the element's content, so a tag typed above its content keeps the tree
	if value.HtmlVoid[lower] {
		l.soft(pos, "<%s> is a void element; write <%s />", name, name)
	}
	switch {
	case lower == "script" || lower == "style":
		n.Children = l.rawText(pos, name, ctx.foreign)
	default:
		if lower == "pre" || lower == "textarea" {
			inner.preserve = true
		}
		n.Children = l.contentOf(pos, name, inner, lower == "pre" || lower == "textarea")
		if lower == "iframe" && len(n.Children.Nodes) > 0 {
			l.soft(pos, "<iframe> takes no content; write <iframe ... />")
		}
	}
	n.Off, n.End, n.EndPos = start, l.pos, l.here()
	return n
}

// elementName scans an element name: ASCII letters, digits, and -. The caller has checked the
// first character, a letter, where one is required.
func (l *lexer) elementName() string {
	start := l.pos
	for isLetter(l.peek(0)) || isDigit(l.peek(0)) || l.peek(0) == '-' {
		l.advance(1)
	}
	return l.src[start:l.pos]
}

// describe names the character at the lexer's position for an error message.
func (l *lexer) describe() string {
	if l.atEnd(0) {
		return "end of file"
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.pos:])
	return strconv.QuoteRune(r)
}

func isMarkupSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

// skipPast advances past the next s, or to the end of the source, when recovering.
func (l *lexer) skipPast(s string) {
	i := strings.Index(l.src[l.pos:], s)
	if i < 0 {
		l.advance(len(l.src) - l.pos)
		return
	}
	l.advance(i + len(s))
}

func (l *lexer) skipMarkupSpace() {
	for !l.atEnd(0) && isMarkupSpace(l.peek(0)) {
		l.advance(1)
	}
}

// attributes scans the attributes of a start tag up to its > or />, which it does not consume.
func (l *lexer) attributes(element string) []token.MarkupAttr {
	var attrs []token.MarkupAttr
	seen := map[string]bool{}
	for {
		l.skipMarkupSpace()
		pos := l.here()
		switch c := l.peek(0); {
		case l.atEnd(0):
			if l.soft(pos, "unterminated start tag <%s>", element) {
				return attrs
			}
		case c == '>' || l.has("/>"):
			return attrs
		case c == '{':
			off := l.pos
			l.advance(1)
			l.skipMarkupSpace()
			if l.has("...") {
				l.advance(3)
			} else {
				l.soft(pos, "an attribute is name=\"text\", name={expr}, or a spread {...attrs}")
			}
			toks := l.interpolationOf(pos, "unterminated { in a start tag")
			attrs = append(attrs, token.MarkupAttr{Kind: token.AttrSpread, Tokens: toks, Pos: pos, EndPos: l.here(), Off: off, End: l.pos})
			continue
		case !isLetter(c) && c != '_':
			if l.soft(pos, "unexpected %s in the start tag <%s>", l.describe(), element) {
				if c == '<' {
					return attrs // the tag was never closed; the next tag belongs to the content
				}
				_, size := utf8.DecodeRuneInString(l.src[l.pos:])
				l.advance(size)
				continue
			}
		}
		start := l.pos
		for !l.atEnd(0) && value.HtmlAttrChar(l.peek(0)) {
			l.advance(1)
		}
		name := l.src[start:l.pos]
		lower := strings.ToLower(name)
		if c := l.peek(0); !l.atEnd(0) && !isMarkupSpace(c) && c != '=' && c != '>' && c != '/' {
			l.soft(l.here(), "an attribute name is ASCII letters, digits, -, _, :, and .; unexpected %s after %s", l.describe(), name)
		}
		if seen[lower] {
			l.soft(pos, "attribute %s is written twice", name)
		}
		seen[lower] = true
		if lower == "key" {
			l.soft(pos, "key is reserved for a later client renderer")
		}
		a := token.MarkupAttr{Kind: token.AttrBare, Name: name, Pos: pos}
		eq := l.pos
		for eq < len(l.src) && isMarkupSpace(l.src[eq]) {
			eq++
		}
		if eq == len(l.src) || l.src[eq] != '=' {
			a.EndPos = l.here()
			attrs = append(attrs, a)
			continue
		}
		l.advance(eq - l.pos + 1)
		l.skipMarkupSpace()
		switch l.peek(0) {
		case '"':
			a.Kind = token.AttrText
			a.Text = l.attrText(name, '"')
		case '{':
			if value.HtmlCodeAttr(lower) {
				l.soft(pos, "%s takes only literal text, since it runs code; pass data in a data- attribute that a script reads, or use a class", name)
			}
			a.Kind = token.AttrExpr
			a.Off = l.pos
			ipos := l.here()
			l.advance(1)
			a.Tokens = l.interpolationOf(ipos, "unterminated { in an attribute")
			a.End = l.pos
		case '\'':
			if l.soft(l.here(), "an attribute value is written in double quotes") {
				a.Kind = token.AttrText
				a.Text = l.attrText(name, '\'')
			}
		default:
			l.soft(l.here(), "an attribute value is \"text\" or {expr}")
		}
		a.EndPos = l.here()
		attrs = append(attrs, a)
	}
}

// attrText scans a quoted attribute value at its opening quote and returns it decoded.
func (l *lexer) attrText(name string, quote byte) string {
	pos := l.here()
	l.advance(1)
	start := l.pos
	for {
		switch c := l.peek(0); {
		case l.atEnd(0):
			if l.soft(pos, "unterminated attribute value") {
				return decodeReferences(l.src[start:l.pos])
			}
		case c == quote:
			raw := l.src[start:l.pos]
			l.advance(1)
			return decodeReferences(raw)
		case c == '{' || c == '}':
			l.soft(l.here(), "%c in a quoted attribute value; interpolate with %s={...}, or write &#%d;", c, name, c)
			l.advance(1)
		case c == '&':
			l.reference()
		default:
			l.advance(1)
		}
	}
}

// reference checks a character reference at its &, and steps over it, or over the & alone when
// it does not begin one.
func (l *lexer) reference() {
	pos := l.here()
	rest := l.src[l.pos+1:]
	if strings.HasPrefix(rest, "#") {
		digits, base := rest[1:], 10
		if strings.HasPrefix(digits, "x") || strings.HasPrefix(digits, "X") {
			digits, base = digits[1:], 16
		}
		n := 0
		for n < len(digits) && (base == 10 && isDigit(digits[n]) || base == 16 && isHex(digits[n])) {
			n++
		}
		if n == 0 || n == len(digits) || digits[n] != ';' {
			l.advance(1)
			return
		}
		v, err := strconv.ParseUint(digits[:n], base, 32)
		if err != nil || v == 0 || v > 0x10FFFF || v >= 0xD800 && v <= 0xDFFF || v >= 0x80 && v <= 0x9F {
			l.soft(pos, "&%s is not a character a page can hold", rest[:len(rest)-len(digits)+n+1])
		}
		l.advance(1 + len(rest) - len(digits) + n + 1)
		return
	}
	n := 0
	for n < len(rest) && (isLetter(rest[n]) || isDigit(rest[n])) {
		n++
	}
	if n == 0 || n == len(rest) || rest[n] != ';' {
		l.advance(1)
		return
	}
	if _, ok := namedReferences[rest[:n]]; !ok {
		l.soft(pos, "&%s; is not one of the named references &amp; &lt; &gt; &quot; &apos; &nbsp;; write the character itself or a numeric reference", rest[:n])
	}
	l.advance(1 + n + 1)
}

var namedReferences = map[string]string{"amp": "&", "lt": "<", "gt": ">", "quot": `"`, "apos": "'", "nbsp": " "}

// decodeReferences replaces each character reference in s, which reference has checked, by the
// character it names.
func decodeReferences(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}
	var sb strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '&' {
			sb.WriteByte(s[i])
			i++
			continue
		}
		semi := strings.IndexByte(s[i:], ';')
		if semi < 0 {
			sb.WriteString(s[i:])
			break
		}
		body := s[i+1 : i+semi]
		if r, ok := referenceValue(body); ok {
			sb.WriteString(r)
			i += semi + 1
			continue
		}
		sb.WriteByte('&')
		i++
	}
	return sb.String()
}

func referenceValue(body string) (string, bool) {
	if r, ok := namedReferences[body]; ok {
		return r, true
	}
	if !strings.HasPrefix(body, "#") || len(body) < 2 {
		return "", false
	}
	digits, base := body[1:], 10
	if digits[0] == 'x' || digits[0] == 'X' {
		digits, base = digits[1:], 16
	}
	v, err := strconv.ParseUint(digits, base, 32)
	if err != nil || digits == "" {
		return "", false
	}
	return string(rune(v)), true
}

// rawText scans the content of a <script> or <style> up to its end tag: literal text, kept
// exactly.
func (l *lexer) rawText(pos token.Pos, name string, foreign bool) *token.Markup {
	start := l.pos
	textPos := l.here()
	closed := true
	for {
		if l.atEnd(0) {
			if l.soft(pos, "<%s> is never closed", name) {
				closed = false
				break
			}
		}
		if l.has("</") && len(l.src)-l.pos >= 2+len(name) && strings.EqualFold(l.src[l.pos+2:l.pos+2+len(name)], name) {
			after := l.pos + 2 + len(name)
			if after >= len(l.src) || l.src[after] == '>' || isMarkupSpace(l.src[after]) || l.src[after] == '/' {
				break
			}
		}
		if foreign && l.has("<![CDATA[") {
			l.soft(l.here(), "<![CDATA[ inside <%s> would reach the CSS as text; remove the CDATA wrapper", name)
		}
		l.advance(1)
	}
	text := l.src[start:l.pos]
	textEnd := l.here()
	m := &token.Markup{}
	if closed {
		m.Close = l.endTag(pos, name)
	}
	if text != "" {
		m.Nodes = append(m.Nodes, token.MarkupNode{Kind: token.MarkupText, Text: text, Pos: textPos, EndPos: textEnd, Off: start, End: start + len(text)})
	}
	return m
}

// endTag scans the end tag of the element name at its </, or of a fragment when name is "", and
// returns where the tag starts. When recovering, an end tag that names an enclosing element is
// left for that element, and the one being read ends before it, as a browser would close it; any
// other mismatched end tag closes the element being read.
func (l *lexer) endTag(open token.Pos, name string) token.Pos {
	pos := l.here()
	mark := l.mark()
	l.advance(2)
	got := l.elementName()
	l.skipMarkupSpace()
	if l.peek(0) != '>' {
		if l.soft(l.here(), "expected > to end the end tag </%s", got) {
			return pos
		}
	}
	l.advance(1)
	if got == name {
		return pos
	}
	switch {
	case name == "":
		l.soft(pos, "</%s> closes nothing; the fragment opened at %d:%d ends with </>", got, open.Line, open.Col)
	case got == "":
		l.soft(pos, "</> closes no fragment; <%s> opened at %d:%d ends with </%s>", name, open.Line, open.Col, name)
	default:
		l.soft(pos, "</%s> does not close <%s>, opened at %d:%d", got, name, open.Line, open.Col)
	}
	for _, outer := range l.open[:max(len(l.open)-1, 0)] {
		if outer == got {
			l.reset(mark) // the outer element closes on this tag
			break
		}
	}
	return pos
}

// A position the lexer can return to, for recovery; nothing is emitted between mark and reset.
type position struct {
	pos, line, col, lines int
}

func (l *lexer) mark() position { return position{l.pos, l.line, l.col, len(l.lines)} }

func (l *lexer) reset(m position) {
	l.pos, l.line, l.col, l.lines = m.pos, m.line, m.col, l.lines[:m.lines]
}

// contentOf scans an element's content up to its end tag. With first set, a line break right
// after the start tag of a <pre> or <textarea> is dropped.
func (l *lexer) contentOf(pos token.Pos, name string, ctx markupContext, first bool) *token.Markup {
	m := l.contentItems(pos, name, ctx)
	if first && len(m.Nodes) > 0 && m.Nodes[0].Kind == token.MarkupText {
		t := &m.Nodes[0]
		switch {
		case strings.HasPrefix(t.Text, "\r\n"):
			t.Text = t.Text[2:]
		case strings.HasPrefix(t.Text, "\n"), strings.HasPrefix(t.Text, "\r"):
			t.Text = t.Text[1:]
		}
	}
	return finishContent(m, ctx)
}

// content scans a fragment's content up to </>.
func (l *lexer) content(pos token.Pos, name string, ctx markupContext) *token.Markup {
	return finishContent(l.contentItems(pos, name, ctx), ctx)
}

// contentItems scans content up to the end tag of name, or </> when name is "", and consumes
// the end tag. Text comes back as written, with an interpolation of only comments as a
// MarkupExpr with no tokens.
func (l *lexer) contentItems(pos token.Pos, name string, ctx markupContext) *token.Markup {
	m := &token.Markup{}
	textStart := -1
	var textPos token.Pos
	flush := func() {
		if textStart >= 0 {
			m.Nodes = append(m.Nodes, token.MarkupNode{Kind: token.MarkupText, Text: l.src[textStart:l.pos], Pos: textPos, EndPos: l.here(), Off: textStart, End: l.pos})
			textStart = -1
		}
	}
	for {
		here := l.here()
		switch c := l.peek(0); {
		case l.atEnd(0):
			flush()
			if name == "" {
				l.soft(pos, "the fragment is never closed; end it with </>")
			} else {
				l.soft(pos, "<%s> is never closed", name)
			}
			return m
		case c == '<':
			switch n := l.peek(1); {
			case n == '/':
				flush()
				before := l.pos
				m.Close = l.endTag(pos, name)
				if l.pos == before {
					m.Close = token.Pos{} // left for the enclosing element
				}
				return m
			case isLetter(n):
				flush()
				m.Nodes = append(m.Nodes, l.element(ctx))
			case l.has("<!--"):
				if l.soft(here, "an HTML comment is written {/* ... */} in markup") {
					l.skipPast("-->")
				}
			case n == '!':
				if l.soft(here, "<! is not allowed in markup; a doctype is written where the page becomes a response") {
					l.skipPast(">")
				}
			case n == '?':
				if l.soft(here, "<? is not allowed in markup") {
					l.skipPast(">")
				}
			default:
				if l.soft(here, "a < in text is written &lt;") {
					if textStart < 0 {
						textStart, textPos = l.pos, here
					}
					l.advance(1)
				}
			}
		case c == '{':
			flush()
			off := l.pos
			l.advance(1)
			toks := l.interpolationOf(here, "unterminated { in markup")
			if len(toks) == 1 {
				// only an EOF: empty, or only comments
				if strings.TrimSpace(l.src[off+1:max(off+1, l.pos-1)]) == "" {
					l.soft(here, "an empty {} in markup")
				}
				toks = nil
			}
			m.Nodes = append(m.Nodes, token.MarkupNode{Kind: token.MarkupExpr, Tokens: toks, Pos: here, EndPos: l.here(), Off: off, End: l.pos})
		case c == '}':
			if l.soft(here, "a } in text is written &#125; or {\"}\"}") {
				if textStart < 0 {
					textStart, textPos = l.pos, here
				}
				l.advance(1)
			}
		default:
			if textStart < 0 {
				textStart, textPos = l.pos, here
			}
			if c == '&' {
				l.reference()
			} else {
				l.advance(1)
			}
		}
	}
}

// finishContent drops the interpolations that hold only comments, joins the text around them,
// collapses white space outside <pre> and <textarea> (section 5, White space), and decodes the
// references. Each piece of text was checked on its own, so the references are decoded piece by
// piece: a reference split by a comment is not one.
func finishContent(m *token.Markup, ctx markupContext) *token.Markup {
	var items []token.MarkupNode
	var joins [][]int // for each item of text, the offsets in it where a dropped comment was
	for _, n := range m.Nodes {
		if n.Kind == token.MarkupExpr && n.Tokens == nil {
			continue
		}
		if k := len(items); n.Kind == token.MarkupText && k > 0 && items[k-1].Kind == token.MarkupText {
			joins[k-1] = append(joins[k-1], len(items[k-1].Text))
			items[k-1].Text += n.Text
			items[k-1].End, items[k-1].EndPos = n.End, n.EndPos
			continue
		}
		items = append(items, n)
		joins = append(joins, nil)
	}
	out := &token.Markup{Close: m.Close}
	for i, n := range items {
		if n.Kind == token.MarkupText {
			text, bounds := n.Text, joins[i]
			if !ctx.preserve {
				text, bounds = collapse(text, bounds, i == 0, i == len(items)-1)
			}
			var sb strings.Builder
			at := 0
			for _, b := range bounds {
				sb.WriteString(decodeReferences(text[at:b]))
				at = b
			}
			sb.WriteString(decodeReferences(text[at:]))
			n.Text = sb.String()
			if n.Text == "" {
				continue
			}
		}
		out.Nodes = append(out.Nodes, n)
	}
	return out
}

// collapse applies the white space rule to text: a run with a line break is removed at the start
// of the content (first) or its end (last), and becomes one space elsewhere. It returns the
// offsets bounds in text moved to where they fall in the result; one inside a run moves past it.
func collapse(s string, bounds []int, first, last bool) (string, []int) {
	var sb strings.Builder
	moved := make([]int, len(bounds))
	b := 0
	for i := 0; i < len(s); {
		if !isMarkupSpace(s[i]) {
			for b < len(bounds) && bounds[b] <= i {
				moved[b] = sb.Len()
				b++
			}
			sb.WriteByte(s[i])
			i++
			continue
		}
		j := i
		for j < len(s) && isMarkupSpace(s[j]) {
			j++
		}
		run := s[i:j]
		switch {
		case !strings.ContainsAny(run, "\r\n"):
			sb.WriteString(run)
		case first && i == 0, last && j == len(s):
		default:
			sb.WriteByte(' ')
		}
		for b < len(bounds) && bounds[b] <= j {
			moved[b] = sb.Len()
			b++
		}
		i = j
	}
	for ; b < len(bounds); b++ {
		moved[b] = sb.Len()
	}
	return sb.String(), moved
}
