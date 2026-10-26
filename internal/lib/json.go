package lib

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// Json (stdlib.md): parse produces Json.Value trees with float64 numbers and maps in document
// order; stringify renders any value, records as objects without their tag.

func init() {
	register("Json", func() value.Value {
		return namespace("Json",
			constant("Value", value.NewTag("Json.Value", value.AliasTag, nil)),
			fn("Json.parse", "text", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v, err := parseJSON(stringArg(g, who, a[0]))
				if err != nil {
					return value.Err(value.ErrorValue(err.Error()))
				}
				return value.Ok(v)
			}),
			fn("Json.stringify", "value", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var sb strings.Builder
				if msg := stringify(&sb, a[0], value.Ancestors{}); msg != "" {
					g.Panic("Json.stringify: %s", msg)
				}
				return sb.String()
			}),
		)
	})
}

// maxJSONDepth is how many arrays and objects Json.parse lets nest (stdlib.md, Json).
const maxJSONDepth = 10000

// jsonParser reads one JSON text (RFC 8259) by recursive descent.
type jsonParser struct {
	src   string
	pos   int
	depth int
}

// jsonError is why a text is not JSON, as Json.parse's Err says it.
type jsonError struct{ msg string }

func (e *jsonError) Error() string { return e.msg }

// errEndOfJSON is the error of a text that ends before its value does.
var errEndOfJSON = &jsonError{"unexpected end of JSON input"}

// parseJSON parses src, which holds one JSON value and white space around it.
func parseJSON(src string) (value.Value, error) {
	p := &jsonParser{src: src}
	v, err := p.value()
	if err != nil {
		return nil, err
	}
	p.space()
	if p.pos < len(p.src) {
		return nil, &jsonError{"unexpected text after the JSON value"}
	}
	return v, nil
}

// fail is the error for what stands at the current position, which is not what was wanted.
func (p *jsonParser) fail(wanted string) error {
	if p.pos >= len(p.src) {
		return errEndOfJSON
	}
	r, _ := utf8.DecodeRuneInString(p.src[p.pos:])
	found := fmt.Sprintf("%q", r)
	if r < ' ' || r == utf8.RuneError {
		found = fmt.Sprintf("U+%04X", r)
	}
	line := 1 + strings.Count(p.src[:p.pos], "\n")
	col := 1 + utf8.RuneCountInString(p.src[strings.LastIndexByte(p.src[:p.pos], '\n')+1:p.pos])
	return &jsonError{fmt.Sprintf("expected %s, found %s at line %d, column %d", wanted, found, line, col)}
}

func (p *jsonParser) space() {
	for p.pos < len(p.src) {
		switch p.src[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

// next skips white space and reports the byte that follows, or 0 at the end.
func (p *jsonParser) next() byte {
	p.space()
	if p.pos < len(p.src) {
		return p.src[p.pos]
	}
	return 0
}

func (p *jsonParser) value() (value.Value, error) {
	switch c := p.next(); {
	case c == '{', c == '[':
		if p.depth == maxJSONDepth {
			return nil, &jsonError{fmt.Sprintf("arrays and objects nested deeper than %d levels", maxJSONDepth)}
		}
		p.depth++
		defer func() { p.depth-- }()
		p.pos++
		if c == '[' {
			return p.array()
		}
		return p.object()
	case c == '"':
		return p.string()
	case c == '-' || '0' <= c && c <= '9':
		return p.number()
	case c == 't':
		return p.literal("true", true)
	case c == 'f':
		return p.literal("false", false)
	case c == 'n':
		return p.literal("null", value.Null)
	}
	return nil, p.fail("a value")
}

func (p *jsonParser) literal(word string, v value.Value) (value.Value, error) {
	for i := 0; i < len(word); i, p.pos = i+1, p.pos+1 {
		if p.pos == len(p.src) || p.src[p.pos] != word[i] {
			return nil, p.fail(fmt.Sprintf("%q", word))
		}
	}
	return v, nil
}

// array reads the elements and the ] after the [.
func (p *jsonParser) array() (value.Value, error) {
	var elems []value.Value
	if p.next() == ']' {
		p.pos++
		return value.NewArray(elems), nil
	}
	for {
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		elems = append(elems, v)
		switch p.next() {
		case ',':
			p.pos++
		case ']':
			p.pos++
			return value.NewArray(elems), nil
		default:
			return nil, p.fail("',' or ']' after an array element")
		}
	}
}

// object reads the members and the } after the {. A repeated key keeps its first place and its
// last value.
func (p *jsonParser) object() (value.Value, error) {
	m := value.NewMap()
	if p.next() == '}' {
		p.pos++
		return m, nil
	}
	for {
		if p.next() != '"' {
			return nil, p.fail("a string key")
		}
		key, err := p.string()
		if err != nil {
			return nil, err
		}
		if p.next() != ':' {
			return nil, p.fail("':' after an object key")
		}
		p.pos++
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		m.Set(key, v)
		switch p.next() {
		case ',':
			p.pos++
		case '}':
			p.pos++
			return m, nil
		default:
			return nil, p.fail("',' or '}' after an object member")
		}
	}
}

// number reads -?(0|[1-9][0-9]*)(.[0-9]+)?([eE][+-]?[0-9]+)? as a float64.
func (p *jsonParser) number() (value.Value, error) {
	start := p.pos
	if p.src[p.pos] == '-' {
		p.pos++
	}
	if p.pos < len(p.src) && p.src[p.pos] == '0' {
		p.pos++
	} else if err := p.digits(); err != nil {
		return nil, err
	}
	if p.pos < len(p.src) && p.src[p.pos] == '.' {
		p.pos++
		if err := p.digits(); err != nil {
			return nil, err
		}
	}
	if p.pos < len(p.src) && (p.src[p.pos] == 'e' || p.src[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.src) && (p.src[p.pos] == '+' || p.src[p.pos] == '-') {
			p.pos++
		}
		if err := p.digits(); err != nil {
			return nil, err
		}
	}
	text := p.src[start:p.pos]
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, &jsonError{fmt.Sprintf("number %s is too large for a float64", text)}
	}
	return f, nil
}

// digits reads one or more decimal digits.
func (p *jsonParser) digits() error {
	start := p.pos
	for p.pos < len(p.src) && '0' <= p.src[p.pos] && p.src[p.pos] <= '9' {
		p.pos++
	}
	if p.pos == start {
		return p.fail("a digit")
	}
	return nil
}

// string reads a string from its opening quote. A \u escape of a surrogate pair is the code point
// the pair encodes, and a surrogate escape outside a pair, like a byte that is not UTF-8, U+FFFD.
func (p *jsonParser) string() (value.Value, error) {
	p.pos++
	var sb strings.Builder
	for {
		if p.pos >= len(p.src) {
			return nil, errEndOfJSON
		}
		c := p.src[p.pos]
		switch {
		case c == '"':
			p.pos++
			return sb.String(), nil
		case c < ' ':
			return nil, p.fail("a character of a string")
		case c == '\\':
			if err := p.escape(&sb); err != nil {
				return nil, err
			}
		case c < utf8.RuneSelf:
			sb.WriteByte(c)
			p.pos++
		default:
			r, size := utf8.DecodeRuneInString(p.src[p.pos:])
			sb.WriteRune(r)
			p.pos += size
		}
	}
}

// escape reads one escape sequence of a string, from its backslash.
func (p *jsonParser) escape(sb *strings.Builder) error {
	p.pos++
	if p.pos >= len(p.src) {
		return errEndOfJSON
	}
	c := p.src[p.pos]
	p.pos++
	switch c {
	case '"', '\\', '/':
		sb.WriteByte(c)
	case 'b':
		sb.WriteByte('\b')
	case 'f':
		sb.WriteByte('\f')
	case 'n':
		sb.WriteByte('\n')
	case 'r':
		sb.WriteByte('\r')
	case 't':
		sb.WriteByte('\t')
	case 'u':
		r, err := p.hex4()
		if err != nil {
			return err
		}
		if utf16.IsSurrogate(r) {
			// The low half follows as another \u escape; one that does not pair is read again.
			if rest := p.src[p.pos:]; len(rest) >= 6 && rest[0] == '\\' && rest[1] == 'u' {
				save := p.pos
				p.pos += 2
				if r2, err := p.hex4(); err == nil {
					if pair := utf16.DecodeRune(r, r2); pair != utf8.RuneError {
						sb.WriteRune(pair)
						return nil
					}
				}
				p.pos = save
			}
			r = utf8.RuneError
		}
		sb.WriteRune(r)
	default:
		p.pos--
		return p.fail(`an escape: one of \" \\ \/ \b \f \n \r \t \u`)
	}
	return nil
}

// hex4 reads the four hex digits of a \u escape.
func (p *jsonParser) hex4() (rune, error) {
	var r rune
	for i := 0; i < 4; i++ {
		if p.pos >= len(p.src) {
			return 0, errEndOfJSON
		}
		c := p.src[p.pos]
		switch {
		case '0' <= c && c <= '9':
			c -= '0'
		case 'a' <= c && c <= 'f':
			c -= 'a' - 10
		case 'A' <= c && c <= 'F':
			c -= 'A' - 10
		default:
			return 0, p.fail("a hex digit")
		}
		r = r<<4 | rune(c)
		p.pos++
	}
	return r, nil
}

// stringify writes v as JSON. It returns why v has no JSON form, a cycle or a part with none, or
// "" when it has one.
func stringify(sb *strings.Builder, v value.Value, stack value.Ancestors) string {
	switch v.(type) {
	case *value.Record, *value.Array, *value.Set, *value.Map:
		if value.Deep(len(stack) + 1) {
			return stringifyOnFreshStack(sb, v, stack)
		}
	}
	return stringifyBody(sb, v, stack)
}

// stringifyOnFreshStack is stringifyBody on a fresh goroutine.
func stringifyOnFreshStack(sb *strings.Builder, v value.Value, stack value.Ancestors) (msg string) {
	value.OnFreshStack(func() { msg = stringifyBody(sb, v, stack) })
	return msg
}

// stringifyBody is stringify once it has decided which goroutine to run on.
func stringifyBody(sb *strings.Builder, v value.Value, stack value.Ancestors) string {
	switch x := v.(type) {
	case int64, int8, int16, int32, uint8, uint16, uint32, uint64:
		sb.WriteString(value.Render(x))
	case float64:
		writeFloat(sb, x, 64)
	case float32:
		writeFloat(sb, float64(x), 32)
	case string:
		writeJSONString(sb, x)
	case bool:
		sb.WriteString(strconv.FormatBool(x))
	case *value.Tagged:
		return stringify(sb, x.Payload, stack)
	case *value.Record:
		if !stack.Enter(x) {
			return "cyclic value"
		}
		defer stack.Leave(x)
		sb.WriteByte('{')
		for i, name := range x.Tag.Fields {
			if i > 0 {
				sb.WriteByte(',')
			}
			writeJSONString(sb, name)
			sb.WriteByte(':')
			if msg := stringify(sb, x.Get(i), stack); msg != "" {
				return msg
			}
		}
		sb.WriteByte('}')
	case *value.Array, *value.Set:
		if !stack.Enter(x) {
			return "cyclic value"
		}
		defer stack.Leave(x)
		var xs []value.Value
		if a, ok := x.(*value.Array); ok {
			xs = a.Values()
		} else {
			xs = x.(*value.Set).Values()
		}
		sb.WriteByte('[')
		for i, e := range xs {
			if i > 0 {
				sb.WriteByte(',')
			}
			if msg := stringify(sb, e, stack); msg != "" {
				return msg
			}
		}
		sb.WriteByte(']')
	case *value.Map:
		if !stack.Enter(x) {
			return "cyclic value"
		}
		defer stack.Leave(x)
		sb.WriteByte('{')
		names := make(map[string]value.Value, x.Len())
		for i, e := range x.Entries() {
			if i > 0 {
				sb.WriteByte(',')
			}
			name, msg := keyName(e.Key)
			if msg != "" {
				return msg
			}
			if first, ok := names[name]; ok {
				return "map keys " + renderKey(first) + " and " + renderKey(e.Key) + " both give the property name " + strconv.Quote(name)
			}
			names[name] = e.Key
			writeJSONString(sb, name)
			sb.WriteByte(':')
			if msg := stringify(sb, e.Value, stack); msg != "" {
				return msg
			}
		}
		sb.WriteByte('}')
	case value.Time:
		writeJSONString(sb, x.String())
	case *value.Regex:
		writeJSONString(sb, x.String())
	default:
		if v == value.Null {
			sb.WriteString("null")
			break
		}
		return noJSON(v)
	}
	return ""
}

// bytesNoJSON is why Json.stringify panics on Bytes: JSON has no form for them (stdlib.md, Json).
const bytesNoJSON = "Bytes has no JSON form; write Encoding.toBase64 of it"

// noJSON is why Json.stringify panics on v, a value that is none of the ones JSON can carry: Bytes,
// a function, void, a type, or a reference such as a channel, a Random or a File.Handle.
func noJSON(v value.Value) string {
	switch x := v.(type) {
	case value.Bytes:
		return bytesNoJSON
	case value.Callable, *value.Constructor:
		return "a function has no JSON form"
	case *value.Tag:
		return "a type has no JSON form"
	case value.Sync:
		return x.SyncName() + " has no JSON form"
	}
	return value.Render(v) + " has no JSON form"
}

// keyName is the property name of a map key, a string or a named primitive over one without its
// tag, or why the key has none (stdlib.md, Json).
func keyName(k value.Value) (string, string) {
	payload := k
	if t, ok := k.(*value.Tagged); ok {
		payload = t.Payload
	}
	if s, ok := payload.(string); ok {
		return s, ""
	}
	return "", "a map key must be a string, got " + describeKey(k)
}

// describeKey names the type of a map key with no JSON form, without rendering it, since a key can
// be a container of any depth.
func describeKey(k value.Value) string {
	if x, ok := k.(value.Sync); ok {
		return x.SyncName()
	}
	if d := value.Describe(k); d != "invalid" {
		return d
	}
	return value.Render(k) // a Random, a File.Handle or a Net.Listener renders as its type's name
}

// renderKey renders a map key as String.from renders it inside a container, a string quoted.
func renderKey(k value.Value) string {
	if s, ok := k.(string); ok {
		return strconv.Quote(s)
	}
	return value.Render(k)
}

func writeFloat(sb *strings.Builder, f float64, bits int) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		sb.WriteString("null")
		return
	}
	sb.WriteString(value.FormatFloat(f, bits))
}

// writeJSONString writes s as a JSON string: a quote and a backslash escaped, a control character
// as \n or its kind or else as \u00XX, U+2028 and U+2029 as \u escapes, since JavaScript once
// read them as line breaks, and every other code point as it is, a byte that is not UTF-8 as
// U+FFFD.
func writeJSONString(sb *strings.Builder, s string) {
	const hex = "0123456789abcdef"
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '\\':
			sb.WriteByte('\\')
			sb.WriteRune(r)
		case '\b':
			sb.WriteString(`\b`)
		case '\f':
			sb.WriteString(`\f`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		case '\u2028', '\u2029':
			sb.WriteString(`\u202`)
			sb.WriteByte(hex[r&0xF])
		default:
			if r < ' ' {
				sb.WriteString(`\u00`)
				sb.WriteByte(hex[r>>4])
				sb.WriteByte(hex[r&0xF])
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
}
