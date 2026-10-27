package check

import (
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// htmlExpr checks an HTML expression (section 5, HTML expressions), whose type is Html.
func (c *Checker) htmlExpr(x *ast.Html, sc *scope) types.Type {
	c.markup(x.Content, sc, false)
	return types.Named{Decl: c.html}
}

func (c *Checker) markup(content []ast.HtmlNode, sc *scope, textOnly bool) {
	for _, n := range content {
		switch {
		case n.Element != nil:
			for _, a := range n.Element.Attrs {
				c.htmlAttr(a, sc)
			}
			c.markup(n.Element.Content, sc, n.Element.TextOnly)
		case n.Expr != nil:
			t := c.synth(n.Expr, sc)
			if !c.renders(t, !textOnly, nil) {
				switch {
				case textOnly && c.renders(t, true, nil):
					c.errorf(n.Expr.Position(), "<title> and <textarea> hold text only, so they cannot hold Html")
				default:
					c.errorf(n.Expr.Position(), "markup takes a string, a number, void, Html, or an array of them, got %s%s", types.String(t), contentHint(t))
				}
			}
		}
	}
}

// contentHint says how to write a value that markup does not take.
func contentHint(t types.Type) string {
	for _, m := range types.Members(t) {
		switch m := types.Expand(m).(type) {
		case types.Null:
			return "; decide what null shows, as {x ?? \"none\"}, or use an if"
		case types.Named:
			if m.Decl.Kind == types.PrimitiveDecl {
				return "; convert it, as {string(x)}"
			}
		}
	}
	return ""
}

// renders reports whether every value of t can be inserted into content: a string, a number,
// void, Html when html is set, or an array of these. seen holds the aliases being checked, so
// that a recursive one, as type Nested = string | Array[Nested], is checked once.
func (c *Checker) renders(t types.Type, html bool, seen []types.Type) bool {
	if n, ok := t.(types.Named); ok && n.Decl.Kind == types.AliasDecl {
		for _, s := range seen {
			if types.Equal(s, t) {
				return true
			}
		}
		seen = append(seen, t)
	}
	switch x := types.Expand(t).(type) {
	case types.Invalid, types.Never, types.Void:
		return true
	case types.Prim:
		return x.Kind == value.KindString || x.Kind.IsNumber()
	case types.Named:
		return html && x.Decl == c.html
	case types.Array:
		return c.renders(x.Elem, html, seen)
	case types.Param:
		return x.Decl.Bound != nil && c.renders(x.Decl.Bound, html, seen)
	case types.Union:
		members := x.Members
		if x.Written != nil {
			// the members as written keep the aliases, which seen recognizes
			members = x.Written
		}
		for _, m := range members {
			if !c.renders(m, html, seen) {
				return false
			}
		}
		return true
	}
	return false
}

// attrValue reports whether every value of t is one an attribute takes: a string, a number, or
// null, a bool when boolean is set, or an array of string, null and void when class is set.
func attrValue(t types.Type, boolean, class bool, depth int) bool {
	if depth > 64 {
		return false
	}
	switch x := types.Expand(t).(type) {
	case types.Invalid, types.Never, types.Null:
		return true
	case types.Prim:
		return x.Kind == value.KindString || x.Kind.IsNumber() || boolean && x.Kind == value.KindBool
	case types.Array:
		return class && classPart(x.Elem, 0)
	case types.Param:
		return x.Decl.Bound != nil && attrValue(x.Decl.Bound, boolean, class, depth+1)
	case types.Union:
		for _, m := range x.Members {
			if !attrValue(m, boolean, class, depth+1) {
				return false
			}
		}
		return true
	}
	return false
}

func classPart(t types.Type, depth int) bool {
	if depth > 64 {
		return false
	}
	switch x := types.Expand(t).(type) {
	case types.Invalid, types.Never, types.Null, types.Void:
		return true
	case types.Prim:
		return x.Kind == value.KindString
	case types.Param:
		return x.Decl.Bound != nil && classPart(x.Decl.Bound, depth+1)
	case types.Union:
		for _, m := range x.Members {
			if !classPart(m, depth+1) {
				return false
			}
		}
		return true
	}
	return false
}

func (c *Checker) htmlAttr(a ast.HtmlAttr, sc *scope) {
	switch {
	case a.Spread:
		valueType, _ := types.MakeUnion([]types.Type{types.StringType, types.BoolType, types.Null{}})
		c.check(a.Expr, sc, types.Map{Key: types.StringType, Val: valueType})
	case a.Expr != nil:
		lower := strings.ToLower(a.Name)
		t := c.synth(a.Expr, sc)
		boolean, class := value.HtmlBoolean[lower], lower == "class"
		if attrValue(t, boolean, class, 0) {
			return
		}
		switch {
		case attrValue(t, true, class, 0):
			c.errorf(a.Expr.Position(), "%s is not a boolean attribute, so it takes a string, not a bool; write the words, as {if on { \"true\" } else { \"false\" }}", a.Name)
		case class:
			c.errorf(a.Expr.Position(), "class takes a string, a number, null, or an array of string, null, and void, got %s", types.String(t))
		default:
			c.errorf(a.Expr.Position(), "attribute %s takes a string, a number, or null, got %s%s", a.Name, types.String(t), contentHint(t))
		}
	}
}
