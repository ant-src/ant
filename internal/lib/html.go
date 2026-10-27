package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

var (
	// HtmlElementTag is Html.Element: { name, attributes, children }.
	HtmlElementTag = value.NewTag("Html.Element", value.RecordTag, []string{"name", "attributes", "children"})
	// HtmlAttributeTag is Html.Attribute: { name, value }.
	HtmlAttributeTag = value.NewTag("Html.Attribute", value.RecordTag, []string{"name", "value"})
	// HtmlTextTag is Html.Text: { text }.
	HtmlTextTag = value.NewTag("Html.Text", value.RecordTag, []string{"text"})
	// HtmlRawTag is Html.Raw: { markup }.
	HtmlRawTag = value.NewTag("Html.Raw", value.RecordTag, []string{"markup"})
)

func htmlArg(g *interp.Goroutine, who string, v value.Value) *value.Html {
	return typedArg[*value.Html](g, who, "Html", v)
}

// htmlNode is a node of an Html value as a record of Html.Node, a copy.
func htmlNode(n value.HtmlNode) value.Value {
	switch n.Kind {
	case value.HtmlText:
		return value.NewRecord(HtmlTextTag, []value.Value{n.Text})
	case value.HtmlRaw:
		return value.NewRecord(HtmlRawTag, []value.Value{n.Text})
	}
	attrs := make([]value.Value, len(n.Attrs))
	for i, a := range n.Attrs {
		attrs[i] = value.NewRecord(HtmlAttributeTag, []value.Value{a.Name, a.Value})
	}
	return value.NewRecord(HtmlElementTag, []value.Value{n.Name, value.NewArray(attrs), n.Children})
}

func init() {
	register("Html", func() value.Value {
		return namespace("Html",
			constant("Node", value.NewTag("Html.Node", value.AliasTag, nil)),
			constant("Element", HtmlElementTag),
			constant("Attribute", HtmlAttributeTag),
			constant("Text", HtmlTextTag),
			constant("Raw", HtmlRawTag),
			constant("Attributes", value.NewTag("Html.Attributes", value.AliasTag, nil)),
			fn("Html.unsafe", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.UnsafeHtml(stringArg(g, who, a[0]))
			}),
			fn("Html.nodes", "h", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := htmlArg(g, who, a[0])
				nodes := make([]value.Value, len(h.Nodes))
				for i, n := range h.Nodes {
					nodes[i] = htmlNode(n)
				}
				return value.NewArray(nodes)
			}),
		)
	})
}
