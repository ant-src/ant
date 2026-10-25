package lib

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// PartsTag is Url.Parts: a reference as RFC 3986 divides it.
var PartsTag = value.NewTag("Url.Parts", value.RecordTag, []string{"scheme", "opaque", "user", "host", "port", "path", "query", "fragment"})

// errNotUTF8 is the error of a %XX sequence that decodes to bytes that are not UTF-8, which no
// string can hold (section 11, String).
var errNotUTF8 = errors.New("the %XX escapes do not decode to UTF-8 text")

// queryUnescape decodes one query key or value as url.QueryUnescape does, and is an error when
// the escapes decode to bytes that are not UTF-8.
func queryUnescape(s string) (string, error) {
	s, err := url.QueryUnescape(s)
	if err == nil && !utf8.ValidString(s) {
		err = errNotUTF8
	}
	return s, err
}

// keptEscape reports whether s starts with an escape that a path keeps: %2F, a slash that is not a
// separator, or %25, so that a path's text tells the two apart from what it decodes to.
func keptEscape(s string) bool {
	return len(s) >= 3 && s[0] == '%' && (s[1:3] == "25" || strings.EqualFold(s[1:3], "2F"))
}

// decodePath decodes the escapes of an escaped path but the ones it keeps.
func decodePath(escaped string) (string, error) {
	var sb strings.Builder
	for escaped != "" {
		i := strings.IndexByte(escaped, '%')
		for i >= 0 && !keptEscape(escaped[i:]) {
			j := strings.IndexByte(escaped[i+1:], '%')
			if j < 0 {
				i = -1
			} else {
				i += 1 + j
			}
		}
		run := escaped
		if i >= 0 {
			run = escaped[:i]
		}
		d, err := url.PathUnescape(run)
		if err != nil {
			return "", err
		}
		sb.WriteString(d)
		if i < 0 {
			break
		}
		sb.WriteString(escaped[i : i+3])
		escaped = escaped[i+3:]
	}
	return sb.String(), nil
}

// partsRecord builds a Url.Parts from a parsed URL. What follows "scheme:" is opaque when it does
// not start with //, a single / included, so a reference with no authority is told apart from
// one whose authority is empty.
func partsRecord(u *url.URL) (value.Value, error) {
	user := ""
	if u.User != nil {
		user = u.User.String()
	}
	path, err := decodePath(u.EscapedPath())
	if err != nil {
		return nil, err
	}
	opaque := u.Opaque
	if u.OmitHost && u.Host == "" && u.User == nil {
		opaque, path = u.EscapedPath(), ""
	}
	return value.NewRecord(PartsTag, []value.Value{u.Scheme, opaque, user, u.Hostname(), u.Port(), path, u.RawQuery, u.Fragment}), nil
}

// escapePath escapes a path for a URL, keeping the escapes decodePath keeps and escaping every
// other %.
func escapePath(path string) string {
	var sb strings.Builder
	for {
		i := 0
		for i < len(path) && !keptEscape(path[i:]) {
			i++
		}
		sb.WriteString((&url.URL{Path: path[:i]}).EscapedPath())
		if i == len(path) {
			return sb.String()
		}
		sb.WriteString(path[i : i+3])
		path = path[i+3:]
	}
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
	u := &url.URL{Scheme: field("scheme"), Opaque: field("opaque"), RawQuery: field("query"), Fragment: field("fragment")}
	u.RawPath = escapePath(field("path"))
	u.Path, _ = url.PathUnescape(u.RawPath)
	if user := field("user"); user != "" {
		name, pass, hasPass := strings.Cut(user, ":")
		// An escape that does not decode stays as written, and is escaped as a whole.
		if n, err := url.PathUnescape(name); err == nil {
			name = n
		}
		if p, err := url.PathUnescape(pass); err == nil {
			pass = p
		}
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
			fn("Url.parse", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				u, err := url.Parse(stringArg(g, who, a[0]))
				if err == nil && !(utf8.ValidString(u.Host) && utf8.ValidString(u.Path) && utf8.ValidString(u.Fragment)) {
					err = errNotUTF8
				}
				var parts value.Value
				if err == nil {
					parts, err = partsRecord(u)
				}
				if err != nil {
					return errResult(err)
				}
				return value.Ok(parts)
			}),
			fn("Url.stringify", "parts", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return partsURL(g, who, a[0]).String()
			}),
			fn("Url.escape", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return url.QueryEscape(stringArg(g, who, a[0]))
			}),
			fn("Url.unescape", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				s, err := queryUnescape(stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(s)
			}),
			fn("Url.parseQuery", "query", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				// url.ParseQuery gives a Go map, which has no order, so the pairs are read here:
				// split at &, each at the first =, both sides unescaped, keys in order of first
				// appearance and a repeated key's values in order.
				m := value.NewMap()
				for _, pair := range strings.Split(stringArg(g, who, a[0]), "&") {
					if pair == "" {
						continue
					}
					k, v, _ := strings.Cut(pair, "=")
					k, err := queryUnescape(k)
					if err != nil {
						return errResult(err)
					}
					v, err = queryUnescape(v)
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
			fn("Url.stringifyQuery", "params", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var sb strings.Builder
				for _, e := range mapArg(g, who, a[0]).Entries() {
					k := url.QueryEscape(stringArg(g, who, e.Key))
					for _, v := range stringsArg(g, who, e.Value) {
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
