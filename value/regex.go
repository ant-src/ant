package value

import (
	"fmt"
	"regexp"
	"strings"
)

// Regex is a compiled regular expression: a Regex.Pattern of section 11, built from a literal
// or by Regex.compile. It is immutable and compared by its normalized source and its sorted
// flags, so it serves as a map key, and its compiled form matches in time linear in the input.
type Regex struct {
	Source string         // normalized: no backslash before a /
	Flags  string         // sorted: some of i, m and s, in that order
	Re     *regexp.Regexp // compiled from (?flags)source
}

// RegexPatternTag is the tag of Regex.Pattern, which is tests.
var RegexPatternTag = NewTag("Regex.Pattern", RegexTag, nil)

// RegexFlags validates a flag string and returns it sorted: each of i, m and s at most once, in
// that order. The error for one of JavaScript's other flags names the Ant way.
func RegexFlags(flags string) (string, error) {
	seen := map[rune]bool{}
	for _, r := range flags {
		switch r {
		case 'i', 'm', 's':
			if seen[r] {
				return "", fmt.Errorf("the flag %c is repeated", r)
			}
			seen[r] = true
		case 'g':
			return "", fmt.Errorf("there is no g flag: Regex.findAll and Regex.replace work on every match")
		case 'y':
			return "", fmt.Errorf("there is no y flag: anchor the pattern with ^ and match a String.slice")
		case 'u', 'v', 'd':
			return "", fmt.Errorf("there is no %c flag: a pattern is Unicode and a match carries its indices already", r)
		default:
			return "", fmt.Errorf("unknown flag %q; the flags are i, m and s", r)
		}
	}
	var sb strings.Builder
	for _, r := range "ims" {
		if seen[r] {
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
}

// NormalizeRegexSource removes the backslash before a / in source, since RE2 reads \/ and /
// alike and a literal needs the backslash only to keep the / from ending it; writes a line
// break, escaped or not, as \n or \r and an empty source as (?:), which RE2 reads the same, so
// that a pattern renders as a literal that reads back to a pattern == to it. Every other escape,
// \\ included, is kept as written, and so is the text of a \Q...\E span (copySpan).
func NormalizeRegexSource(source string) string {
	if source == "" {
		return "(?:)"
	}
	if !strings.ContainsAny(source, "\\\n\r") {
		return source
	}
	var sb strings.Builder
	for i := 0; i < len(source); i++ {
		switch source[i] {
		case '\\':
			if i+1 < len(source) && source[i+1] == 'Q' {
				i = copySpan(&sb, source, i+2) - 1
				continue
			}
			if i+1 < len(source) {
				switch next := source[i+1]; next {
				case '/':
					sb.WriteByte('/')
				case '\n': // an escaped line break is the line break, which \n writes
					sb.WriteString(`\n`)
				case '\r':
					sb.WriteString(`\r`)
				default:
					sb.WriteByte('\\')
					sb.WriteByte(next)
				}
				i++
				continue
			}
			sb.WriteByte('\\')
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		default:
			sb.WriteByte(source[i])
		}
	}
	return sb.String()
}

// copySpan writes the \Q...\E span whose text starts at source[i], as written, since RE2 takes
// every character of it as itself, and returns the offset after its \E. RE2 ends the span at the
// first \E, or at the end of the source, where the copy adds the \E. A span that is closed and
// reopened around a single / or line break, as a rendering writes it, continues with that
// character as itself, so that the rendering reads back to the same source.
func copySpan(sb *strings.Builder, source string, i int) int {
	sb.WriteString(`\Q`)
	for {
		end := strings.Index(source[i:], `\E`)
		if end < 0 {
			sb.WriteString(source[i:])
			sb.WriteString(`\E`)
			return len(source)
		}
		sb.WriteString(source[i : i+end])
		i += end + 2
		c, n := reopened(source[i:])
		if n == 0 {
			sb.WriteString(`\E`)
			return i
		}
		sb.WriteByte(c)
		i += n
	}
}

// reopened reports whether s, which follows the \E of a span, starts with a / or a line break,
// as itself or escaped, and then \Q. It returns that character and the length through the \Q.
func reopened(s string) (byte, int) {
	c, n := byte(0), 0
	switch {
	case len(s) >= 1 && (s[0] == '/' || s[0] == '\n' || s[0] == '\r'):
		c, n = s[0], 1
	case len(s) >= 2 && s[0] == '\\':
		c, n = map[byte]byte{'/': '/', '\n': '\n', '\r': '\r', 'n': '\n', 'r': '\r'}[s[1]], 2
	}
	if c == 0 || !strings.HasPrefix(s[n:], `\Q`) {
		return 0, 0
	}
	return c, n + 2
}

// NewRegex compiles source with flags into a pattern. The source is normalized and the flags
// sorted, so a literal and Regex.compile build equal values from one expression. An error
// names the construct a JavaScript programmer will meet: lookaround, a backreference, or \u.
func NewRegex(source, flags string) (*Regex, error) {
	sorted, err := RegexFlags(flags)
	if err != nil {
		return nil, err
	}
	source = NormalizeRegexSource(source)
	expr := source
	if sorted != "" {
		expr = "(?" + sorted + ")" + source
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, regexError(err)
	}
	return &Regex{Source: source, Flags: sorted, Re: re}, nil
}

// regexError rewords Go's message for the constructs RE2 lacks.
func regexError(err error) error {
	msg := strings.TrimPrefix(err.Error(), "error parsing regexp: ")
	switch {
	case strings.Contains(msg, "`(?=`") || strings.Contains(msg, "`(?!`"):
		return fmt.Errorf("lookahead is not supported: a pattern is RE2, which has no lookahead, lookbehind or backreferences")
	case strings.Contains(msg, "`(?<=") || strings.Contains(msg, "`(?<!"):
		return fmt.Errorf("lookbehind is not supported: a pattern is RE2, which has no lookahead, lookbehind or backreferences")
	case strings.HasPrefix(msg, "invalid escape sequence: `\\") && len(msg) > 27 && msg[27] >= '1' && msg[27] <= '9':
		return fmt.Errorf("backreferences are not supported: a pattern is RE2, which has no lookahead, lookbehind or backreferences")
	case strings.HasPrefix(msg, "invalid escape sequence: `\\u`"):
		return fmt.Errorf("\\u is not an escape in a pattern: write a code point as \\x{0041}")
	}
	return fmt.Errorf("invalid regular expression: %s", msg)
}

// String renders the pattern as a literal that reads back to it.
func (r *Regex) String() string { return RegexLiteral(r.Source, r.Flags) }

// RegexLiteral writes a source and flags as a literal that reads back to them: every / of the
// source escaped, a line break as \n or \r, and an empty source as (?:), as JavaScript has it.
// Inside a \Q...\E span, where an escape would stand for itself, the span is closed before such
// a character and reopened after it: \E\/\Q, \E\n\Q, or \E\r\Q, which copySpan reads back.
func RegexLiteral(source, flags string) string {
	if source == "" {
		source = "(?:)"
	}
	var sb strings.Builder
	sb.WriteByte('/')
	inSpan := false
	for i := 0; i < len(source); i++ {
		c := source[i]
		switch {
		case inSpan && strings.HasPrefix(source[i:], `\E`):
			inSpan = false
			sb.WriteString(`\E`)
			i++
		case !inSpan && c == '\\' && i+1 < len(source):
			inSpan = source[i+1] == 'Q'
			sb.WriteString(source[i : i+2])
			i++
		case c == '/' || c == '\n' || c == '\r':
			esc := map[byte]string{'/': `\/`, '\n': `\n`, '\r': `\r`}[c]
			if inSpan {
				esc = `\E` + esc + `\Q`
			}
			sb.WriteString(esc)
		default:
			sb.WriteByte(c)
		}
	}
	sb.WriteByte('/')
	sb.WriteString(flags)
	return sb.String()
}
