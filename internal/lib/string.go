package lib

import (
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"

	"github.com/ant-src/ant/internal/grapheme"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// CutTag is String.Cut: { before: string, after: string }.
var CutTag = value.NewTag("String.Cut", value.RecordTag, []string{"before", "after"})

// runeIndex converts a byte offset into s to a code point index, or null for a miss.
func runeIndex(s string, i int) value.Value {
	if i < 0 {
		return value.Null
	}
	return int64(utf8.RuneCountInString(s[:i]))
}

// baseArg reads an optional base, 10 when omitted, and panics outside 2 to 36.
func baseArg(g *interp.Goroutine, who string, v value.Value) int {
	if !given(v) {
		return 10
	}
	base := intArg(g, who, v)
	if base < 2 || base > 36 {
		g.Panic("%s needs a base from 2 to 36, got %d", who, base)
	}
	return int(base)
}

// pad pads s with pad to width code points, on the left or the right; the padding is pad
// repeated and cut to the code points s lacks.
func pad(g *interp.Goroutine, who string, a []value.Value, left bool) value.Value {
	s := stringArg(g, who, a[0])
	width := intArg(g, who, a[1])
	boundCount(g, who, "width", width)
	p := " "
	if given(a[2]) {
		p = stringArg(g, who, a[2])
	}
	if p == "" {
		g.Panic("%s needs a pad that is not empty", who)
	}
	missing := int(width) - utf8.RuneCountInString(s)
	if missing <= 0 {
		return s
	}
	prs := []rune(p)
	fill := make([]rune, missing)
	for i := range fill {
		fill[i] = prs[i%len(prs)]
	}
	if left {
		return string(fill) + s
	}
	return s + string(fill)
}

// classTest is String.isDigit and the others: true when every code point of s satisfies f,
// and false on the empty string.
func classTest(name string, f func(rune) bool) *interp.Builtin {
	return fn(name, "s", func(g *interp.Goroutine, _ string, a []value.Value) value.Value {
		s := stringArg(g, name, a[0])
		if s == "" {
			return false
		}
		for _, r := range s {
			if !f(r) {
				return false
			}
		}
		return true
	})
}

// normForms are the four normalization forms of UAX #15 that String.normalize takes.
var normForms = map[string]norm.Form{"NFC": norm.NFC, "NFD": norm.NFD, "NFKC": norm.NFKC, "NFKD": norm.NFKD}

// categories are the two-letter general categories String.category returns, other than Cn,
// which is every code point none of them holds. Cs is not among them: a string holds no
// surrogate.
var categories = []string{
	"Lu", "Ll", "Lt", "Lm", "Lo", "Mn", "Mc", "Me", "Nd", "Nl", "No", "Pc", "Pd", "Ps", "Pe", "Pi",
	"Pf", "Po", "Sm", "Sc", "Sk", "So", "Zs", "Zl", "Zp", "Cc", "Cf", "Co",
}

// scripted lists, in order and merged, the runs of code points that some script of
// unicode.Scripts holds. A code point outside all of them is in the script Unicode calls
// Unknown, which unicode.Scripts does not list.
var scripted = sync.OnceValue(func() [][2]rune {
	var runs [][2]rune
	add := func(lo, hi, stride rune) {
		if stride == 1 {
			runs = append(runs, [2]rune{lo, hi})
			return
		}
		for r := lo; r <= hi; r += stride {
			runs = append(runs, [2]rune{r, r})
		}
	}
	for _, table := range unicode.Scripts {
		for _, r := range table.R16 {
			add(rune(r.Lo), rune(r.Hi), rune(r.Stride))
		}
		for _, r := range table.R32 {
			add(rune(r.Lo), rune(r.Hi), rune(r.Stride))
		}
	}
	slices.SortFunc(runs, func(a, b [2]rune) int { return int(a[0] - b[0]) })
	merged := runs[:1]
	for _, run := range runs[1:] {
		last := &merged[len(merged)-1]
		if run[0] <= last[1]+1 {
			last[1] = max(last[1], run[1])
		} else {
			merged = append(merged, run)
		}
	}
	return merged
})

// inSomeScript reports whether a script of unicode.Scripts holds r.
func inSomeScript(r rune) bool {
	runs := scripted()
	i := sort.Search(len(runs), func(i int) bool { return runs[i][1] >= r })
	return i < len(runs) && runs[i][0] <= r
}

// oneCodePoint reads a string argument of exactly one code point, and panics on any other.
func oneCodePoint(g *interp.Goroutine, who string, v value.Value) rune {
	s := stringArg(g, who, v)
	r, size := utf8.DecodeRuneInString(s)
	if s == "" || size != len(s) {
		g.Panic("%s needs a string of one code point, got %s", who, value.RenderQuoted(s))
	}
	return r
}

// isScalar reports whether n is a Unicode scalar value: at most 10FFFF and not a surrogate.
func isScalar(n int64) bool {
	return n >= 0 && n <= unicode.MaxRune && (n < 0xD800 || n > 0xDFFF)
}

// lines splits s at every \n, drops a \r just before one, and omits an empty final piece.
func lines(s string) []value.Value {
	var out []value.Value
	for {
		i := strings.IndexByte(s, '\n')
		if i < 0 {
			break
		}
		out = append(out, strings.TrimSuffix(s[:i], "\r"))
		s = s[i+1:]
	}
	if s != "" {
		out = append(out, s)
	}
	return out
}

// digitsOf reports whether s is an optional sign followed by one or more digits of base, in
// either letter case, with no prefix, space or underscore.
func digitsOf(s string, base int) bool {
	if s == "" {
		return false
	}
	if s[0] == '+' || s[0] == '-' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		var d int
		switch {
		case c >= '0' && c <= '9':
			d = int(c - '0')
		case c >= 'a' && c <= 'z':
			d = int(c-'a') + 10
		case c >= 'A' && c <= 'Z':
			d = int(c-'A') + 10
		default:
			return false
		}
		if d >= base {
			return false
		}
	}
	return true
}

func init() {
	register("String", func() value.Value {
		return namespace("String",
			constant("Cut", CutTag),
			fn("String.length", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(utf8.RuneCountInString(stringArg(g, who, a[0])))
			}),
			fn("String.isEmpty", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return stringArg(g, who, a[0]) == ""
			}),
			fn("String.slice", "s, from, to", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				rs := []rune(stringArg(g, who, a[0]))
				from, to := intArg(g, who, a[1]), intArg(g, who, a[2])
				sliceBounds(g, who, from, to, len(rs))
				return string(rs[from:to])
			}),
			fn("String.concat", "parts", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.Join(stringsArg(g, who, a[0]), "")
			}),
			fn("String.split", "s, sep", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				s, sep := stringArg(g, who, a[0]), stringArg(g, who, a[1])
				if sep == "" {
					g.Panic("String.split needs a separator that is not empty")
				}
				parts := strings.Split(s, sep)
				out := make([]value.Value, len(parts))
				for i, p := range parts {
					out[i] = p
				}
				return value.NewArray(out)
			}),
			fn("String.join", "parts, sep", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.Join(stringsArg(g, who, a[0]), stringArg(g, who, a[1]))
			}),
			fn("String.repeat", "s, n", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				n := intArg(g, who, a[1])
				if n < 0 {
					g.Panic("String.repeat needs a count of at least 0, got %d", n)
				}
				boundCount(g, who, "count", n)
				return strings.Repeat(stringArg(g, who, a[0]), int(n))
			}),
			fn("String.trim", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.TrimFunc(stringArg(g, who, a[0]), unicode.IsSpace)
			}),
			fn("String.trimStart", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.TrimLeftFunc(stringArg(g, who, a[0]), unicode.IsSpace)
			}),
			fn("String.trimEnd", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.TrimRightFunc(stringArg(g, who, a[0]), unicode.IsSpace)
			}),
			fn("String.trimPrefix", "s, prefix", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.TrimPrefix(stringArg(g, who, a[0]), stringArg(g, who, a[1]))
			}),
			fn("String.trimSuffix", "s, suffix", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.TrimSuffix(stringArg(g, who, a[0]), stringArg(g, who, a[1]))
			}),
			fn("String.padStart", "s, width, pad?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return pad(g, who, a, true)
			}),
			fn("String.padEnd", "s, width, pad?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return pad(g, who, a, false)
			}),
			fn("String.contains", "s, t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.Contains(stringArg(g, who, a[0]), stringArg(g, who, a[1]))
			}),
			fn("String.startsWith", "s, t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.HasPrefix(stringArg(g, who, a[0]), stringArg(g, who, a[1]))
			}),
			fn("String.endsWith", "s, t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.HasSuffix(stringArg(g, who, a[0]), stringArg(g, who, a[1]))
			}),
			fn("String.indexOf", "s, t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				s := stringArg(g, who, a[0])
				i := strings.Index(s, stringArg(g, who, a[1]))
				if i < 0 {
					return value.Null
				}
				return int64(utf8.RuneCountInString(s[:i]))
			}),
			fn("String.lastIndexOf", "s, t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				s := stringArg(g, who, a[0])
				return runeIndex(s, strings.LastIndex(s, stringArg(g, who, a[1])))
			}),
			fn("String.count", "s, t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t := stringArg(g, who, a[1])
				if t == "" {
					g.Panic("String.count needs a substring that is not empty")
				}
				return int64(strings.Count(stringArg(g, who, a[0]), t))
			}),
			fn("String.replace", "s, old, new", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				old := stringArg(g, who, a[1])
				if old == "" {
					g.Panic("String.replace needs a pattern that is not empty")
				}
				return strings.ReplaceAll(stringArg(g, who, a[0]), old, stringArg(g, who, a[2]))
			}),
			fn("String.replaceFirst", "s, old, new", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				old := stringArg(g, who, a[1])
				if old == "" {
					g.Panic("String.replaceFirst needs a pattern that is not empty")
				}
				return strings.Replace(stringArg(g, who, a[0]), old, stringArg(g, who, a[2]), 1)
			}),
			fn("String.cut", "s, sep", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				sep := stringArg(g, who, a[1])
				if sep == "" {
					g.Panic("String.cut needs a separator that is not empty")
				}
				before, after, found := strings.Cut(stringArg(g, who, a[0]), sep)
				if !found {
					return value.Null
				}
				return value.NewRecord(CutTag, []value.Value{before, after})
			}),
			fn("String.fields", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				fields := strings.Fields(stringArg(g, who, a[0]))
				out := make([]value.Value, len(fields))
				for i, f := range fields {
					out[i] = f
				}
				return value.NewArray(out)
			}),
			fn("String.lines", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.NewArray(lines(stringArg(g, who, a[0])))
			}),
			fn("String.toUpper", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return cases.Upper(language.Und).String(stringArg(g, who, a[0]))
			}),
			fn("String.toLower", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return cases.Lower(language.Und).String(stringArg(g, who, a[0]))
			}),
			fn("String.foldCase", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return cases.Fold().String(stringArg(g, who, a[0]))
			}),
			fn("String.equalFold", "a, b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				x, y := stringArg(g, who, a[0]), stringArg(g, who, a[1])
				fold := cases.Fold() // String resets it, so one serves both sides
				return fold.String(x) == fold.String(y)
			}),
			fn("String.normalize", "s, form", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				s, name := stringArg(g, who, a[0]), stringArg(g, who, a[1])
				form, ok := normForms[name]
				if !ok {
					g.Panic("String.normalize needs the form NFC, NFD, NFKC or NFKD, got %s", value.RenderQuoted(name))
				}
				return form.String(s)
			}),
			classTest("String.isDigit", unicode.IsDigit),
			classTest("String.isLetter", unicode.IsLetter),
			classTest("String.isNumber", unicode.IsNumber),
			classTest("String.isSpace", unicode.IsSpace),
			classTest("String.isUpper", unicode.IsUpper),
			classTest("String.isLower", unicode.IsLower),
			classTest("String.isPunct", unicode.IsPunct),
			classTest("String.isSymbol", unicode.IsSymbol),
			classTest("String.isMark", unicode.IsMark),
			classTest("String.isControl", unicode.IsControl),
			classTest("String.isGraphic", unicode.IsGraphic),
			classTest("String.isPrint", unicode.IsPrint),
			fn("String.category", "c", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				r := oneCodePoint(g, who, a[0])
				for _, c := range categories {
					if unicode.Is(unicode.Categories[c], r) {
						return c
					}
				}
				return "Cn"
			}),
			fn("String.inScript", "c, script", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				r := oneCodePoint(g, who, a[0])
				name := stringArg(g, who, a[1])
				if name == "Unknown" {
					return !inSomeScript(r)
				}
				table, ok := unicode.Scripts[name]
				if !ok {
					g.Panic("String.inScript needs the name of a script, got %s", value.RenderQuoted(name))
				}
				return unicode.Is(table, r)
			}),
			fn("String.chars", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var out []value.Value
				for _, r := range stringArg(g, who, a[0]) {
					out = append(out, string(r))
				}
				return value.NewArray(out)
			}),
			fn("String.graphemes", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				clusters := grapheme.Split(stringArg(g, who, a[0]))
				out := make([]value.Value, len(clusters))
				for i, c := range clusters {
					out[i] = c
				}
				return value.NewArray(out)
			}),
			fn("String.toBytes", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.Bytes(stringArg(g, who, a[0]))
			}),
			fn("String.fromBytes", "bytes", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				b := bytesArg(g, who, a[0])
				if !utf8.ValidString(string(b)) {
					return value.Null
				}
				return string(b)
			}),
			fn("String.fromBytesLossy", "bytes", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.Text(string(bytesArg(g, who, a[0])))
			}),
			fn("String.codePoint", "c", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(oneCodePoint(g, who, a[0]))
			}),
			fn("String.fromCodePoint", "n", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				n := intArg(g, who, a[0])
				if !isScalar(n) {
					return value.Null
				}
				return string(rune(n))
			}),
			fn("String.toCodePoints", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var out []value.Value
				for _, r := range stringArg(g, who, a[0]) {
					out = append(out, int64(r))
				}
				return value.NewArray(out)
			}),
			fn("String.fromCodePoints", "points", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var sb strings.Builder
				for _, v := range arrayArg(g, who, a[0]).Values() {
					p, ok := v.(int64)
					if !ok {
						g.Panic("String.fromCodePoints needs an array of ints, got an element of kind %s", value.Describe(v))
					}
					if !isScalar(p) {
						return value.Null
					}
					sb.WriteRune(rune(p))
				}
				return sb.String()
			}),
			fn("String.toInt", "s, base?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				s := stringArg(g, who, a[0])
				base := baseArg(g, who, a[1])
				if !digitsOf(s, base) {
					return value.Null
				}
				n, err := strconv.ParseInt(s, base, 64)
				if err != nil {
					return value.Null
				}
				return n
			}),
			fn("String.toFloat", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				s := stringArg(g, who, a[0])
				if s == "" || strings.ContainsAny(s, "_xXpP") || strings.ContainsAny(strings.ToLower(s), "in") {
					return value.Null
				}
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return value.Null
				}
				return f
			}),
			fn("String.fromInt", "x, base?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				base := baseArg(g, who, a[1])
				switch x := a[0].(type) {
				case int64:
					return strconv.FormatInt(x, base)
				case int8:
					return strconv.FormatInt(int64(x), base)
				case int16:
					return strconv.FormatInt(int64(x), base)
				case int32:
					return strconv.FormatInt(int64(x), base)
				case uint8:
					return strconv.FormatUint(uint64(x), base)
				case uint16:
					return strconv.FormatUint(uint64(x), base)
				case uint32:
					return strconv.FormatUint(uint64(x), base)
				case uint64:
					return strconv.FormatUint(x, base)
				}
				g.Panic("String.fromInt needs an integer, got %s", value.Describe(a[0]))
				return nil
			}),
			fn("String.fromFloat", "x, digits", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				x := floatArg(g, who, a[0])
				digits := intArg(g, who, a[1])
				if digits < 0 {
					g.Panic("String.fromFloat needs a count of digits of at least 0, got %d", digits)
				}
				boundCount(g, who, "count of digits", digits)
				if math.IsNaN(x) || math.IsInf(x, 0) {
					return value.Render(x)
				}
				return strconv.FormatFloat(x, 'f', int(digits), 64)
			}),
			fn("String.from", "value", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.Render(a[0])
			}),
			constant("unicodeVersion", unicode.Version),
		)
	})
}
