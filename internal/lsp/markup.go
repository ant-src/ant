package lsp

import (
	"strings"

	"github.com/ant-src/ant/value"
)

// markupContext is where a cursor stands inside an HTML expression, read from the text between
// the expression's start and the cursor.
type markupContext struct {
	kind    markupKind
	element string   // the element whose start tag the cursor is in, or whose name is being typed
	prefix  string   // the name typed so far
	open    []string // the elements open at the cursor, outermost first
}

type markupKind int

const (
	inText      markupKind = iota // between tags
	inTagName                     // after <, in an element name
	inAttrs                       // in a start tag, between attributes
	inAttrName                    // in an attribute name
	inAttrValue                   // in a quoted attribute value
	inEndTag                      // after </
	inInterp                      // in a {...} interpolation
)

// scanMarkup reads markup from its first < up to the cursor and reports where the cursor is.
func scanMarkup(src string) markupContext {
	var ctx markupContext
	state := inText
	var prev markupKind // the state an interpolation returns to
	depth := 0
	var name, attr strings.Builder
	var raw string // the raw text element being read: script or style
	i := 0
	for i < len(src) {
		c := src[i]
		switch state {
		case inText:
			if raw != "" {
				// literal text up to the end tag
				if strings.HasPrefix(src[i:], "</") && strings.HasPrefix(strings.ToLower(src[i+2:]), raw) {
					raw = ""
					continue
				}
				i++
				continue
			}
			switch {
			case strings.HasPrefix(src[i:], "</"):
				state = inEndTag
				name.Reset()
				i += 2
				continue
			case c == '<':
				state = inTagName
				name.Reset()
			case c == '{':
				prev, state, depth = inText, inInterp, 1
			}
		case inTagName:
			switch {
			case c == '>':
				ctx.open = append(ctx.open, name.String())
				state = inText
				if l := strings.ToLower(name.String()); value.HtmlRawText[l] {
					raw = l
				}
			case strings.HasPrefix(src[i:], "/>"):
				state = inText
				i += 2
				continue
			case c == ' ' || c == '\t' || c == '\n' || c == '\r':
				state = inAttrs
			default:
				name.WriteByte(c)
			}
		case inAttrs:
			switch {
			case c == '>':
				n := name.String()
				if !value.HtmlVoid[strings.ToLower(n)] {
					ctx.open = append(ctx.open, n)
				}
				if l := strings.ToLower(n); value.HtmlRawText[l] {
					raw = l
				}
				state = inText
			case strings.HasPrefix(src[i:], "/>"):
				state = inText
				i += 2
				continue
			case c == '"':
				state = inAttrValue
			case c == '{':
				prev, state, depth = inAttrs, inInterp, 1
			case c == '=' || c == ' ' || c == '\t' || c == '\n' || c == '\r':
			default:
				state = inAttrName
				attr.Reset()
				attr.WriteByte(c)
			}
		case inAttrName:
			switch {
			case c == '=' || c == ' ' || c == '\t' || c == '\n' || c == '\r':
				state = inAttrs
				continue
			case c == '>' || c == '/':
				state = inAttrs
				continue
			default:
				attr.WriteByte(c)
			}
		case inAttrValue:
			if c == '"' {
				state = inAttrs
			}
		case inEndTag:
			switch {
			case c == '>':
				if n := len(ctx.open); n > 0 {
					ctx.open = ctx.open[:n-1]
				}
				state = inText
			default:
				name.WriteByte(c)
			}
		case inInterp:
			switch c {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					state = prev
				}
			case '"', '`':
				// skip a string or template inside the interpolation
				j := i + 1
				for j < len(src) && src[j] != c {
					if src[j] == '\\' {
						j++
					}
					j++
				}
				i = j
			}
		}
		i++
	}
	ctx.kind = state
	switch state {
	case inTagName, inEndTag:
		ctx.prefix = name.String()
	case inAttrName:
		ctx.prefix = attr.String()
		ctx.element = name.String()
	case inAttrs:
		ctx.element = name.String()
	}
	return ctx
}
