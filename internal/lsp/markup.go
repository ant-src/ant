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
				n := name.String()
				if !value.HtmlVoid[strings.ToLower(n)] {
					ctx.open = append(ctx.open, n)
				}
				state = inText
				if l := strings.ToLower(n); value.HtmlRawText[l] {
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
				i = skipString(src, i)
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

// skipString returns the index of the quote or backtick that closes the string or raw string
// opening at i, or len(src) when none does. A string's escapes and its interpolations, with the
// strings and braces inside them, are stepped over.
func skipString(src string, i int) int {
	if src[i] == '`' {
		if j := strings.IndexByte(src[i+1:], '`'); j >= 0 {
			return i + 1 + j
		}
		return len(src)
	}
	for j := i + 1; j < len(src); j++ {
		switch {
		case src[j] == '\\':
			j++
		case src[j] == '"':
			return j
		case strings.HasPrefix(src[j:], "${"):
			j++ // at the {
			for depth := 1; depth > 0 && j+1 < len(src); {
				j++
				switch src[j] {
				case '{':
					depth++
				case '}':
					depth--
				case '"', '`':
					j = skipString(src, j)
				}
			}
		}
	}
	return len(src)
}
