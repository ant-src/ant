package value

import (
	"strings"
)

// Html is the value of an HTML expression: Html of section 11, an immutable list of nodes in
// normal form, where no two text nodes are adjacent and no text node is empty. Since a value is
// built from finished values and never changed, it holds no cycle.
type Html struct {
	Nodes []HtmlNode
}

// HtmlNodeKind is the kind of an Html node.
type HtmlNodeKind uint8

const (
	HtmlElement HtmlNodeKind = iota
	HtmlText
	HtmlRaw
)

// HtmlNode is an element, a text node, or a raw node. An element has a Name, its Attrs in their
// final order, and Children; a text node's decoded text and a raw node's markup are in Text.
type HtmlNode struct {
	Kind     HtmlNodeKind
	Name     string
	Attrs    []HtmlAttr
	Children *Html
	Text     string
}

// HtmlAttr is an attribute that is present; a boolean attribute that is on has the empty value.
type HtmlAttr struct {
	Name, Value string
}

// EmptyHtml is the empty Html, the value of <></>.
var EmptyHtml = &Html{}

// HtmlBuilder collects nodes in normal form.
type HtmlBuilder struct {
	nodes []HtmlNode
}

// Text appends a text node, joining it to a text node before it; an empty one adds nothing.
func (b *HtmlBuilder) Text(s string) {
	if s == "" {
		return
	}
	if n := len(b.nodes); n > 0 && b.nodes[n-1].Kind == HtmlText {
		b.nodes[n-1].Text += s
		return
	}
	b.nodes = append(b.nodes, HtmlNode{Kind: HtmlText, Text: s})
}

// Node appends one node, joining a text node to one before it.
func (b *HtmlBuilder) Node(n HtmlNode) {
	if n.Kind == HtmlText {
		b.Text(n.Text)
		return
	}
	b.nodes = append(b.nodes, n)
}

// Splice appends the nodes of h.
func (b *HtmlBuilder) Splice(h *Html) {
	for _, n := range h.Nodes {
		b.Node(n)
	}
}

// Html returns the value built.
func (b *HtmlBuilder) Html() *Html {
	if len(b.nodes) == 0 {
		return EmptyHtml
	}
	return &Html{Nodes: b.nodes}
}

// UnsafeHtml is Html.unsafe: s as one raw node, or the empty Html for an empty s.
func UnsafeHtml(s string) *Html {
	if s == "" {
		return EmptyHtml
	}
	return &Html{Nodes: []HtmlNode{{Kind: HtmlRaw, Text: s}}}
}

func equalHtml(a, b *Html) bool {
	if len(a.Nodes) != len(b.Nodes) {
		return false
	}
	for i := range a.Nodes {
		x, y := &a.Nodes[i], &b.Nodes[i]
		if x.Kind != y.Kind || x.Name != y.Name || x.Text != y.Text || len(x.Attrs) != len(y.Attrs) {
			return false
		}
		for j := range x.Attrs {
			if x.Attrs[j] != y.Attrs[j] {
				return false
			}
		}
		if x.Kind == HtmlElement && !equalHtml(x.Children, y.Children) {
			return false
		}
	}
	return true
}

// ---- rendering

// RenderHtml is String.from of an Html value: its markup.
func RenderHtml(h *Html) string {
	var sb strings.Builder
	renderHtml(&sb, h, false)
	return sb.String()
}

// renderHtml writes the nodes of h; foreign is true inside <svg> or <math> and outside a
// <foreignObject>, where the browser reads <style> as ordinary text.
func renderHtml(sb *strings.Builder, h *Html, foreign bool) {
	for i := range h.Nodes {
		n := &h.Nodes[i]
		switch n.Kind {
		case HtmlText:
			escapeHtml(sb, n.Text)
		case HtmlRaw:
			sb.WriteString(n.Text)
		case HtmlElement:
			renderElement(sb, n, foreign)
		}
	}
}

func renderElement(sb *strings.Builder, n *HtmlNode, foreign bool) {
	name := strings.ToLower(n.Name)
	sb.WriteByte('<')
	sb.WriteString(n.Name)
	for _, a := range n.Attrs {
		sb.WriteByte(' ')
		sb.WriteString(a.Name)
		if a.Value != "" {
			sb.WriteString(`="`)
			escapeHtml(sb, a.Value)
			sb.WriteByte('"')
		}
	}
	sb.WriteByte('>')
	if HtmlVoid[name] {
		return
	}
	switch {
	case name == "svg" || name == "math":
		foreign = true
	case name == "foreignobject":
		foreign = false
	}
	if !foreign && (name == "pre" || name == "textarea") && len(n.Children.Nodes) > 0 &&
		n.Children.Nodes[0].Kind == HtmlText && startsWithBreak(n.Children.Nodes[0].Text) {
		// The browser drops a line break right after the start tag, so one more keeps the text
		// whole.
		sb.WriteByte('\n')
	}
	if !foreign && HtmlRawText[name] {
		// The parser allowed only literal text here, which is written as it is.
		for _, c := range n.Children.Nodes {
			sb.WriteString(c.Text)
		}
	} else {
		renderHtml(sb, n.Children, foreign)
	}
	sb.WriteString("</")
	sb.WriteString(n.Name)
	sb.WriteByte('>')
}

func startsWithBreak(s string) bool { return strings.HasPrefix(s, "\n") || strings.HasPrefix(s, "\r") }

// escapeHtml writes s with &, <, >, ", and ' as references, which is correct in text and in a
// double-quoted attribute value alike.
func escapeHtml(sb *strings.Builder, s string) {
	last := 0
	for i := 0; i < len(s); i++ {
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			esc = "&quot;"
		case '\'':
			esc = "&#39;"
		default:
			continue
		}
		sb.WriteString(s[last:i])
		sb.WriteString(esc)
		last = i + 1
	}
	sb.WriteString(s[last:])
}

// ---- the fixed lists of section 5, matched on lower-case names

func set(words string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(words) {
		m[w] = true
	}
	return m
}

var (
	// HtmlVoid are the void elements, written self-closing and rendered without an end tag.
	HtmlVoid = set("area base br col embed hr img input link meta source track wbr")
	// HtmlRawText are the elements whose content is literal text, rendered unescaped.
	HtmlRawText = set("script style")
	// HtmlTextOnly are the elements that hold text only, outside <svg> and <math>.
	HtmlTextOnly = set("title textarea")
	// HtmlObsolete are the elements that are syntax errors.
	HtmlObsolete = set("xmp plaintext noembed noframes")
	// HtmlBoolean are the boolean attributes, where true gives the empty value and false leaves the
	// attribute out.
	HtmlBoolean = set(`allowfullscreen async autofocus autoplay checked controls default defer disabled
		formnovalidate hidden inert ismap itemscope loop multiple muted nomodule novalidate open
		playsinline readonly required reversed selected`)
	// HtmlURL are the URL attributes, whose interpolated value is checked for its scheme.
	HtmlURL = set(`action archive background cite classid codebase data formaction href icon longdesc
		manifest poster profile src usemap xlink:href`)
	// HtmlSrcset are the attributes that hold a list of URLs.
	HtmlSrcset = set("srcset imagesrcset")
)

// HtmlCodeAttr reports whether the attribute name, in lower case, runs code, so that it takes only
// literal text: every name that starts with on, style, and srcdoc.
func HtmlCodeAttr(lower string) bool {
	return strings.HasPrefix(lower, "on") || lower == "style" || lower == "srcdoc"
}

// HtmlElementName reports whether s is an element name: an ASCII letter, then ASCII letters,
// digits, and -.
func HtmlElementName(s string) bool {
	if s == "" || !isASCIILetter(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if c := s[i]; !isASCIILetter(c) && !isASCIIDigit(c) && c != '-' {
			return false
		}
	}
	return true
}

// HtmlAttrName reports whether s is an attribute name: an ASCII letter or _, then ASCII letters,
// digits, -, _, :, and ..
func HtmlAttrName(s string) bool {
	if s == "" || !isASCIILetter(s[0]) && s[0] != '_' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !HtmlAttrChar(s[i]) {
			return false
		}
	}
	return true
}

// HtmlAttrChar reports whether c may continue an attribute name.
func HtmlAttrChar(c byte) bool {
	return isASCIILetter(c) || isASCIIDigit(c) || c == '-' || c == '_' || c == ':' || c == '.'
}

func isASCIILetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
func isASCIIDigit(c byte) bool  { return c >= '0' && c <= '9' }

// ---- URLs

// HtmlInvalidURL replaces a URL whose scheme is not safe.
const HtmlInvalidURL = "about:invalid#ant"

// HtmlCheckURL returns u when it is relative or its scheme is http, https, mailto, or tel, and
// HtmlInvalidURL otherwise. It reads u as the URL standard does: leading and trailing U+0000 to
// U+0020 removed, and every tab and line break removed.
func HtmlCheckURL(u string) string {
	s := strings.TrimFunc(u, func(r rune) bool { return r <= 0x20 })
	s = strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, s)
	if s == "" || !isASCIILetter(s[0]) {
		return u
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c == ':' {
			switch strings.ToLower(s[:i]) {
			case "http", "https", "mailto", "tel":
				return u
			}
			return HtmlInvalidURL
		}
		if !isASCIILetter(c) && !isASCIIDigit(c) && c != '+' && c != '-' && c != '.' {
			return u
		}
	}
	return u
}

// HtmlCheckSrcset checks each comma-separated candidate of a srcset value: its URL, up to the
// first white space, by HtmlCheckURL, keeping the descriptors and the separators as written.
func HtmlCheckSrcset(v string) string {
	parts := strings.Split(v, ",")
	for i, p := range parts {
		start := 0
		for start < len(p) && isHtmlSpace(p[start]) {
			start++
		}
		end := start
		for end < len(p) && !isHtmlSpace(p[end]) {
			end++
		}
		if end > start {
			if checked := HtmlCheckURL(p[start:end]); checked != p[start:end] {
				parts[i] = p[:start] + checked + p[end:]
			}
		}
	}
	return strings.Join(parts, ",")
}

func isHtmlSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' }
