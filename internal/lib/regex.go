package lib

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// MatchTag is Regex.Match: { text, start, end, groups, named }.
var MatchTag = value.NewTag("Regex.Match", value.RecordTag, []string{"text", "start", "end", "groups", "named"})

func regexArg(g *interp.Goroutine, who string, v value.Value) *value.Regex {
	p, ok := v.(*value.Regex)
	if !ok {
		g.Panic("%s needs a Regex.Pattern, got %s", who, value.Describe(v))
	}
	return p
}

func stringResult(g *interp.Goroutine, who string, v value.Value) string {
	s, ok := v.(string)
	if !ok {
		g.Panic("the function given to %s must return a string, got %s", who, value.Describe(v))
	}
	return s
}

// codePoints converts byte offsets of one string into code point indices, in order. The
// cursor advances through the string once, so a findAll over a long text is linear in it.
type codePoints struct {
	s    string
	byte int // the offset the cursor stands at
	rune int // the code point index there
}

func (c *codePoints) at(offset int) int {
	if offset < c.byte {
		c.byte, c.rune = 0, 0
	}
	c.rune += utf8.RuneCountInString(c.s[c.byte:offset])
	c.byte = offset
	return c.rune
}

// match builds a Regex.Match from a submatch location, as FindStringSubmatchIndex returns it.
func match(p *value.Regex, s string, loc []int, cursor *codePoints) value.Value {
	groups := make([]value.Value, len(loc)/2)
	for i := range groups {
		if loc[2*i] < 0 {
			groups[i] = value.Null
		} else {
			groups[i] = s[loc[2*i]:loc[2*i+1]]
		}
	}
	named := value.NewMap()
	for i, name := range p.Re.SubexpNames() {
		if name != "" {
			named.Set(name, groups[i])
		}
	}
	start := cursor.at(loc[0])
	end := cursor.at(loc[1])
	return value.NewRecord(MatchTag, []value.Value{s[loc[0]:loc[1]], int64(start), int64(end), value.NewArray(groups), named})
}

func init() {
	register("Regex", func() value.Value {
		return namespace("Regex",
			constant("Pattern", value.RegexPatternTag),
			constant("Match", MatchTag),
			fn("Regex.compile", "source, flags?", func(g *interp.Goroutine, a []value.Value) value.Value {
				flags := ""
				if given(a[1]) {
					flags = stringArg(g, "Regex.compile", a[1])
				}
				p, err := value.NewRegex(stringArg(g, "Regex.compile", a[0]), flags)
				if err != nil {
					return errResult(err)
				}
				return value.Ok(p)
			}),
			fn("Regex.test", "p, s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return regexArg(g, "Regex.test", a[0]).Re.MatchString(stringArg(g, "Regex.test", a[1]))
			}),
			fn("Regex.find", "p, s", func(g *interp.Goroutine, a []value.Value) value.Value {
				p, s := regexArg(g, "Regex.find", a[0]), stringArg(g, "Regex.find", a[1])
				loc := p.Re.FindStringSubmatchIndex(s)
				if loc == nil {
					return value.Null
				}
				return match(p, s, loc, &codePoints{s: s})
			}),
			fn("Regex.findAll", "p, s", func(g *interp.Goroutine, a []value.Value) value.Value {
				p, s := regexArg(g, "Regex.findAll", a[0]), stringArg(g, "Regex.findAll", a[1])
				cursor := &codePoints{s: s}
				var out []value.Value
				for _, loc := range p.Re.FindAllStringSubmatchIndex(s, -1) {
					out = append(out, match(p, s, loc, cursor))
				}
				return value.NewArray(out)
			}),
			fn("Regex.replace", "p, s, replacement", func(g *interp.Goroutine, a []value.Value) value.Value {
				p := regexArg(g, "Regex.replace", a[0])
				return p.Re.ReplaceAllString(stringArg(g, "Regex.replace", a[1]), stringArg(g, "Regex.replace", a[2]))
			}),
			fn("Regex.replaceFirst", "p, s, replacement", func(g *interp.Goroutine, a []value.Value) value.Value {
				p, s := regexArg(g, "Regex.replaceFirst", a[0]), stringArg(g, "Regex.replaceFirst", a[1])
				replacement := stringArg(g, "Regex.replaceFirst", a[2])
				loc := p.Re.FindStringSubmatchIndex(s)
				if loc == nil {
					return s
				}
				return s[:loc[0]] + string(p.Re.ExpandString(nil, replacement, s, loc)) + s[loc[1]:]
			}),
			fn("Regex.replaceWith", "p, s, f", func(g *interp.Goroutine, a []value.Value) value.Value {
				p, s := regexArg(g, "Regex.replaceWith", a[0]), stringArg(g, "Regex.replaceWith", a[1])
				f := funcArg(g, "Regex.replaceWith", a[2])
				cursor := &codePoints{s: s}
				var sb strings.Builder
				last := 0
				for _, loc := range p.Re.FindAllStringSubmatchIndex(s, -1) {
					sb.WriteString(s[last:loc[0]])
					sb.WriteString(stringResult(g, "Regex.replaceWith", g.Call(f, match(p, s, loc, cursor))))
					last = loc[1]
				}
				sb.WriteString(s[last:])
				return sb.String()
			}),
			fn("Regex.split", "p, s", func(g *interp.Goroutine, a []value.Value) value.Value {
				pieces := regexArg(g, "Regex.split", a[0]).Re.Split(stringArg(g, "Regex.split", a[1]), -1)
				out := make([]value.Value, len(pieces))
				for i, piece := range pieces {
					out[i] = piece
				}
				return value.NewArray(out)
			}),
			fn("Regex.escape", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return regexp.QuoteMeta(stringArg(g, "Regex.escape", a[0]))
			}),
			fn("Regex.source", "p", func(g *interp.Goroutine, a []value.Value) value.Value {
				return regexArg(g, "Regex.source", a[0]).Source
			}),
			fn("Regex.flags", "p", func(g *interp.Goroutine, a []value.Value) value.Value {
				return regexArg(g, "Regex.flags", a[0]).Flags
			}),
		)
	})
}
