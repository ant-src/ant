package lib

import (
	"net/url"
	"strings"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// PartsTag is Url.Parts: a reference as RFC 3986 divides it.
var PartsTag = value.NewTag("Url.Parts", value.RecordTag, []string{"scheme", "opaque", "user", "host", "port", "path", "query", "fragment"})

// partsRecord builds a Url.Parts from a parsed URL.
func partsRecord(u *url.URL) value.Value {
	user := ""
	if u.User != nil {
		user = u.User.String()
	}
	return value.NewRecord(PartsTag, []value.Value{u.Scheme, u.Opaque, user, u.Hostname(), u.Port(), u.Path, u.RawQuery, u.Fragment})
}

// partsURL rebuilds a URL from a Url.Parts record.
func partsURL(g *interp.Goroutine, who string, v value.Value) *url.URL {
	r, ok := v.(*value.Record)
	if !ok || r.Tag != PartsTag {
		g.Panic("%s needs a Url.Parts, got %s", who, value.Describe(v))
	}
	field := func(name string) string {
		f, _ := r.Field(name)
		return stringArg(g, who, f)
	}
	u := &url.URL{Scheme: field("scheme"), Opaque: field("opaque"), Path: field("path"), RawQuery: field("query"), Fragment: field("fragment")}
	if user := field("user"); user != "" {
		name, pass, hasPass := strings.Cut(user, ":")
		name, _ = url.PathUnescape(name)
		pass, _ = url.PathUnescape(pass)
		if hasPass {
			u.User = url.UserPassword(name, pass)
		} else {
			u.User = url.User(name)
		}
	}
	host := field("host")
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = "[" + host + "]" // an IPv6 address
	}
	if port := field("port"); port != "" {
		host += ":" + port
	}
	u.Host = host
	return u
}

func init() {
	register("Url", func() value.Value {
		return namespace("Url",
			constant("Parts", PartsTag),
			fn("Url.parse", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				u, err := url.Parse(stringArg(g, "Url.parse", a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(partsRecord(u))
			}),
			fn("Url.stringify", "parts", func(g *interp.Goroutine, a []value.Value) value.Value {
				return partsURL(g, "Url.stringify", a[0]).String()
			}),
			fn("Url.escape", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				return url.QueryEscape(stringArg(g, "Url.escape", a[0]))
			}),
			fn("Url.unescape", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				s, err := url.QueryUnescape(stringArg(g, "Url.unescape", a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(s)
			}),
			fn("Url.parseQuery", "query", func(g *interp.Goroutine, a []value.Value) value.Value {
				// url.ParseQuery gives a Go map, which has no order, so the pairs are read here:
				// split at &, each at the first =, both sides unescaped, keys in order of first
				// appearance and a repeated key's values in order.
				m := value.NewMap()
				for _, pair := range strings.Split(stringArg(g, "Url.parseQuery", a[0]), "&") {
					if pair == "" {
						continue
					}
					k, v, _ := strings.Cut(pair, "=")
					k, err := url.QueryUnescape(k)
					if err != nil {
						return errResult(err)
					}
					v, err = url.QueryUnescape(v)
					if err != nil {
						return errResult(err)
					}
					if vs, ok := m.Get(k); ok {
						vs.(*value.Array).Push(v)
					} else {
						m.Set(k, value.NewArray([]value.Value{v}))
					}
				}
				return value.Ok(m)
			}),
			fn("Url.stringifyQuery", "params", func(g *interp.Goroutine, a []value.Value) value.Value {
				var sb strings.Builder
				for _, e := range mapArg(g, "Url.stringifyQuery", a[0]).Entries() {
					k := url.QueryEscape(stringArg(g, "Url.stringifyQuery", e.Key))
					for _, v := range stringsArg(g, "Url.stringifyQuery", e.Value) {
						if sb.Len() > 0 {
							sb.WriteByte('&')
						}
						sb.WriteString(k)
						sb.WriteByte('=')
						sb.WriteString(url.QueryEscape(v))
					}
				}
				return sb.String()
			}),
		)
	})
}
