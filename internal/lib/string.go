package lib

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

func init() {
	register("String", func() value.Value {
		return namespace("String",
			fn("String.length", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return int64(utf8.RuneCountInString(stringArg(g, "String.length", a[0])))
			}),
			fn("String.isEmpty", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return stringArg(g, "String.isEmpty", a[0]) == ""
			}),
			fn("String.slice", "s, from, to", func(g *interp.Goroutine, a []value.Value) value.Value {
				rs := []rune(stringArg(g, "String.slice", a[0]))
				from, to := intArg(g, "String.slice", a[1]), intArg(g, "String.slice", a[2])
				sliceBounds(g, "String.slice", from, to, len(rs))
				return string(rs[from:to])
			}),
			fn("String.concat", "parts", func(g *interp.Goroutine, a []value.Value) value.Value {
				return strings.Join(stringsArg(g, "String.concat", a[0]), "")
			}),
			fn("String.split", "s, sep", func(g *interp.Goroutine, a []value.Value) value.Value {
				s, sep := stringArg(g, "String.split", a[0]), stringArg(g, "String.split", a[1])
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
			fn("String.join", "parts, sep", func(g *interp.Goroutine, a []value.Value) value.Value {
				return strings.Join(stringsArg(g, "String.join", a[0]), stringArg(g, "String.join", a[1]))
			}),
			fn("String.repeat", "s, n", func(g *interp.Goroutine, a []value.Value) value.Value {
				n := intArg(g, "String.repeat", a[1])
				if n < 0 {
					g.Panic("String.repeat needs a count of at least 0, got %d", n)
				}
				return strings.Repeat(stringArg(g, "String.repeat", a[0]), int(n))
			}),
			fn("String.trim", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return strings.TrimFunc(stringArg(g, "String.trim", a[0]), unicode.IsSpace)
			}),
			fn("String.contains", "s, t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return strings.Contains(stringArg(g, "String.contains", a[0]), stringArg(g, "String.contains", a[1]))
			}),
			fn("String.startsWith", "s, t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return strings.HasPrefix(stringArg(g, "String.startsWith", a[0]), stringArg(g, "String.startsWith", a[1]))
			}),
			fn("String.endsWith", "s, t", func(g *interp.Goroutine, a []value.Value) value.Value {
				return strings.HasSuffix(stringArg(g, "String.endsWith", a[0]), stringArg(g, "String.endsWith", a[1]))
			}),
			fn("String.indexOf", "s, t", func(g *interp.Goroutine, a []value.Value) value.Value {
				s := stringArg(g, "String.indexOf", a[0])
				i := strings.Index(s, stringArg(g, "String.indexOf", a[1]))
				if i < 0 {
					return value.Null
				}
				return int64(utf8.RuneCountInString(s[:i]))
			}),
			fn("String.replace", "s, old, new", func(g *interp.Goroutine, a []value.Value) value.Value {
				old := stringArg(g, "String.replace", a[1])
				if old == "" {
					g.Panic("String.replace needs a pattern that is not empty")
				}
				return strings.ReplaceAll(stringArg(g, "String.replace", a[0]), old, stringArg(g, "String.replace", a[2]))
			}),
			fn("String.toUpper", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return strings.ToUpper(stringArg(g, "String.toUpper", a[0]))
			}),
			fn("String.toLower", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return strings.ToLower(stringArg(g, "String.toLower", a[0]))
			}),
			fn("String.chars", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				var out []value.Value
				for _, r := range stringArg(g, "String.chars", a[0]) {
					out = append(out, string(r))
				}
				return value.NewArray(out)
			}),
			fn("String.toBytes", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				s := stringArg(g, "String.toBytes", a[0])
				out := make([]value.Value, len(s))
				for i := 0; i < len(s); i++ {
					out[i] = uint8(s[i])
				}
				return value.NewArray(out)
			}),
			fn("String.fromBytes", "bytes", func(g *interp.Goroutine, a []value.Value) value.Value {
				vals := arrayArg(g, "String.fromBytes", a[0]).Values()
				b := make([]byte, len(vals))
				for i, v := range vals {
					x, ok := v.(uint8)
					if !ok {
						g.Panic("String.fromBytes needs an array of bytes, got an element of kind %s", value.Describe(v))
					}
					b[i] = x
				}
				if !utf8.Valid(b) {
					return value.Null
				}
				return string(b)
			}),
			fn("String.toCodePoints", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				var out []value.Value
				for _, r := range stringArg(g, "String.toCodePoints", a[0]) {
					out = append(out, int64(r))
				}
				return value.NewArray(out)
			}),
			fn("String.fromCodePoints", "points", func(g *interp.Goroutine, a []value.Value) value.Value {
				var sb strings.Builder
				for _, v := range arrayArg(g, "String.fromCodePoints", a[0]).Values() {
					p, ok := v.(int64)
					if !ok {
						g.Panic("String.fromCodePoints needs an array of ints, got an element of kind %s", value.Describe(v))
					}
					if p < 0 || p > unicode.MaxRune || (p >= 0xD800 && p <= 0xDFFF) {
						return value.Null
					}
					sb.WriteRune(rune(p))
				}
				return sb.String()
			}),
			fn("String.toInt", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				s := stringArg(g, "String.toInt", a[0])
				if !decimal(s) {
					return value.Null
				}
				n, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return value.Null
				}
				return n
			}),
			fn("String.toFloat", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				s := stringArg(g, "String.toFloat", a[0])
				if s == "" || strings.ContainsAny(s, "_xXpP") || strings.ContainsAny(strings.ToLower(s), "in") {
					return value.Null
				}
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return value.Null
				}
				return f
			}),
			fn("String.from", "value", func(g *interp.Goroutine, a []value.Value) value.Value {
				return value.Render(a[0])
			}),
		)
	})
}

// decimal reports whether s is an optional sign followed by decimal digits only.
func decimal(s string) bool {
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
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
