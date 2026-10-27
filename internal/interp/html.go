package interp

import (
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/value"
)

// html evaluates an HTML expression (section 12, Expressions): attributes and then content, left
// to right, into a new Html in normal form.
func (g *Goroutine) html(e *ast.Html, env *Env) (value.Value, ctl) {
	var b value.HtmlBuilder
	if v, c := g.content(&b, e.Content, env, false); c != ctlNone {
		return v, c
	}
	return b.Html(), ctlNone
}

func (g *Goroutine) content(b *value.HtmlBuilder, content []ast.HtmlNode, env *Env, textOnly bool) (value.Value, ctl) {
	for _, n := range content {
		switch {
		case n.Element != nil:
			el, v, c := g.element(n.Element, env)
			if c != ctlNone {
				return v, c
			}
			b.Node(el)
		case n.Expr != nil:
			v, c := g.eval(n.Expr, env)
			if c != ctlNone {
				return v, c
			}
			g.insert(b, v, n.Pos, textOnly, nil)
		default:
			b.Text(n.Text)
		}
	}
	return nil, ctlNone
}

// insert adds an interpolated value to content by its kind. outer holds the arrays being
// inserted, so that an array inside itself panics rather than recursing forever.
func (g *Goroutine) insert(b *value.HtmlBuilder, v value.Value, pos ast.Pos, textOnly bool, outer []*value.Array) {
	switch x := v.(type) {
	case string:
		b.Text(x)
		return
	case *value.Html:
		if textOnly {
			g.panicf(pos, "<title> and <textarea> hold text only, so they cannot hold Html")
		}
		b.Splice(x)
		return
	case *value.Array:
		for _, a := range outer {
			if a == x {
				g.panicf(pos, "an array inside itself cannot be inserted into markup")
			}
		}
		outer = append(outer, x)
		if value.Deep(len(outer)) {
			value.OnFreshStack(func() {
				for _, e := range x.Values() {
					g.insert(b, e, pos, textOnly, outer)
				}
			})
			return
		}
		for _, e := range x.Values() {
			g.insert(b, e, pos, textOnly, outer)
		}
		return
	}
	if v == value.Void {
		return
	}
	if value.KindOf(v).IsNumber() {
		b.Text(value.Render(v))
		return
	}
	g.panicf(pos, "markup takes a string, a number, void, Html, or an array of them, got %s", value.Describe(v))
}

// attr is an attribute while the names of an element are resolved; one that is not present is
// dropped at the end.
type attr struct {
	name, value string
	present     bool
}

func (g *Goroutine) element(e *ast.HtmlElement, env *Env) (value.HtmlNode, value.Value, ctl) {
	var attrs []attr
	index := map[string]int{}
	set := func(name, v string, present bool) {
		lower := strings.ToLower(name)
		if i, ok := index[lower]; ok {
			attrs[i].value, attrs[i].present = v, present
			return
		}
		index[lower] = len(attrs)
		attrs = append(attrs, attr{name, v, present})
	}
	for _, a := range e.Attrs {
		switch {
		case a.Spread:
			v, c := g.eval(a.Expr, env)
			if c != ctlNone {
				return value.HtmlNode{}, v, c
			}
			m, ok := v.(*value.Map)
			if !ok {
				g.panicf(a.Pos, "an attribute spread needs a map, got %s", value.Describe(v))
			}
			for _, kv := range m.Entries() {
				name, ok := kv.Key.(string)
				if !ok {
					g.panicf(a.Pos, "an attribute spread needs string names, got %s", value.Describe(kv.Key))
				}
				lower := strings.ToLower(name)
				switch {
				case !value.HtmlAttrName(name):
					g.panicf(a.Pos, "%q is not an attribute name", name)
				case value.HtmlCodeAttr(lower):
					g.panicf(a.Pos, "%s takes only literal text, so a spread cannot give it", name)
				case lower == "key":
					g.panicf(a.Pos, "key is reserved, so a spread cannot give it")
				}
				switch kv.Value.(type) {
				case string, bool:
				default:
					if kv.Value != value.Null {
						g.panicf(a.Pos, "the value of %s in an attribute spread must be a string, a bool, or null, got %s", name, value.Describe(kv.Value))
					}
				}
				s, present := g.attrValue(name, kv.Value, a.Pos)
				set(name, s, present)
			}
		case a.Expr != nil:
			v, c := g.eval(a.Expr, env)
			if c != ctlNone {
				return value.HtmlNode{}, v, c
			}
			s, present := g.attrValue(a.Name, v, a.Pos)
			set(a.Name, s, present)
		default:
			set(a.Name, a.Text, true)
		}
	}
	var b value.HtmlBuilder
	if v, c := g.content(&b, e.Content, env, e.TextOnly); c != ctlNone {
		return value.HtmlNode{}, v, c
	}
	n := value.HtmlNode{Kind: value.HtmlElement, Name: e.Name, Children: b.Html()}
	for _, a := range attrs {
		if a.present {
			n.Attrs = append(n.Attrs, value.HtmlAttr{Name: a.name, Value: a.value})
		}
	}
	return n, nil, ctlNone
}

// attrValue is the string an interpolated attribute value gives, a URL checked, and whether the
// attribute is present.
func (g *Goroutine) attrValue(name string, v value.Value, pos ast.Pos) (string, bool) {
	lower := strings.ToLower(name)
	var s string
	switch x := v.(type) {
	case string:
		s = x
	case bool:
		if !value.HtmlBoolean[lower] {
			g.panicf(pos, "%s is not a boolean attribute, so it takes a string, not a bool", name)
		}
		return "", x
	case *value.Array:
		if lower != "class" {
			g.panicf(pos, "only class takes an array, not %s", name)
		}
		var parts []string
		for _, e := range x.Values() {
			switch p := e.(type) {
			case string:
				if p != "" {
					parts = append(parts, p)
				}
			default:
				if e != value.Null && e != value.Void {
					g.panicf(pos, "a class array holds strings, null, and void, got %s", value.Describe(e))
				}
			}
		}
		if len(parts) == 0 {
			return "", false
		}
		return strings.Join(parts, " "), true
	default:
		switch {
		case v == value.Null:
			return "", false
		case value.KindOf(v).IsNumber():
			s = value.Render(v)
		default:
			g.panicf(pos, "attribute %s takes a string, a number, or null, got %s", name, value.Describe(v))
		}
	}
	switch {
	case value.HtmlURL[lower]:
		s = value.HtmlCheckURL(s)
	case value.HtmlSrcset[lower]:
		s = value.HtmlCheckSrcset(s)
	}
	return s, true
}
