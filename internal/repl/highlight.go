package repl

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/lexer"
	"github.com/ant-src/ant/internal/token"
)

// A class is what a byte of an input is, for its color.
type class uint8

const (
	clsPlain    class = iota
	clsKeyword        // a keyword
	clsConstant       // true, false, null, void
	clsString         // a string or template literal, including the braces of its interpolations
	clsNumber         // a number literal
	clsRegex          // a regular expression literal
	clsComment        // a comment
	clsType           // a namespace, a type, or a module: a capitalized name, or a built-in type
	clsFunction       // a name that is called
	clsMarkup         // an HTML expression, outside its interpolations
	clsBad            // what the lexer could not read, or a reserved word
	clsGhost          // the completion suggested after the cursor
)

// classes returns the class of every byte of src, which may be unfinished: the lexer goes on
// past what it cannot read.
func classes(src string) []class {
	cls := make([]class, len(src))
	toks, comments, _ := lexer.LexRecover(src, false)
	for _, c := range comments {
		fill(cls, c.Off, c.End, clsComment)
	}
	classify(cls, toks)
	return cls
}

func fill(cls []class, from, to int, c class) {
	for i := max(from, 0); i < to && i < len(cls); i++ {
		cls[i] = c
	}
}

// classify colors a token stream: that of the input, of an interpolation, or of an expression
// in markup.
func classify(cls []class, toks []token.Token) {
	for i, t := range toks {
		switch t.Kind {
		case token.Keyword:
			switch t.Text {
			case "true", "false", "null", "void":
				fill(cls, t.Off, t.End, clsConstant)
			default:
				fill(cls, t.Off, t.End, clsKeyword)
			}
		case token.Reserved, token.Bad:
			fill(cls, t.Off, t.End, clsBad)
		case token.Ident:
			switch {
			case next(toks, i).Is("("):
				fill(cls, t.Off, t.End, clsFunction)
			case capitalized(t.Text) || builtinTypes[t.Text]:
				// namespaces, types, and modules are capitalized by convention (section 10)
				fill(cls, t.Off, t.End, clsType)
			}
		case token.Int, token.Float:
			fill(cls, t.Off, t.End, clsNumber)
		case token.String:
			fill(cls, t.Off, t.End, clsString)
		case token.Template:
			fill(cls, t.Off, t.End, clsString)
			for _, p := range t.Parts {
				if p.Tokens != nil {
					// the expression keeps its own classes between the ${ and the }
					fill(cls, p.Off+2, p.End-1, clsPlain)
					classify(cls, p.Tokens)
				}
			}
		case token.Regex:
			fill(cls, t.Off, t.End, clsRegex)
		case token.Html:
			fill(cls, t.Off, t.End, clsMarkup)
			if t.Markup != nil {
				classifyMarkup(cls, t.Markup)
			}
		}
	}
}

// classifyMarkup colors the expressions inside markup: interpolations and attribute values.
func classifyMarkup(cls []class, m *token.Markup) {
	for _, n := range m.Nodes {
		switch n.Kind {
		case token.MarkupExpr:
			fill(cls, n.Off+1, n.End-1, clsPlain)
			classify(cls, n.Tokens)
		case token.MarkupElement:
			for _, a := range n.Attrs {
				if a.Tokens != nil {
					fill(cls, a.Off+1, a.End-1, clsPlain)
					classify(cls, a.Tokens)
				}
			}
			if n.Children != nil {
				classifyMarkup(cls, n.Children)
			}
		}
	}
}

// next is the token after toks[i] that is not a line break, or an EOF token.
func next(toks []token.Token, i int) token.Token {
	for j := i + 1; j < len(toks); j++ {
		if toks[j].Kind != token.Newline {
			return toks[j]
		}
	}
	return token.Token{Kind: token.EOF}
}

// builtinTypes are the type names written in lower case (section 4).
var builtinTypes = map[string]bool{
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint8": true, "uint16": true, "uint32": true, "uint64": true, "byte": true,
	"float32": true, "float64": true, "string": true, "bool": true, "never": true,
}

func capitalized(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

// A theme is the escape sequence that starts each class, or "" for none.
type theme [clsGhost + 1]string

// ansiTheme colors with the terminal's own sixteen colors, so that it suits a light or a dark
// background alike.
var ansiTheme = theme{
	clsKeyword:  "\x1b[35m", // magenta
	clsConstant: "\x1b[33m", // yellow
	clsString:   "\x1b[32m", // green
	clsNumber:   "\x1b[33m", // yellow
	clsRegex:    "\x1b[31m", // red
	clsComment:  "\x1b[90m", // bright black
	clsType:     "\x1b[36m", // cyan
	clsFunction: "\x1b[34m", // blue
	clsMarkup:   "\x1b[36m", // cyan
	clsBad:      "\x1b[31m", // red
	clsGhost:    "\x1b[90m", // bright black
}

// plainTheme colors nothing, for NO_COLOR and a dumb terminal.
var plainTheme theme

const reset = "\x1b[0m"

// paint writes s colored by the class of each of its bytes, which cls gives from byte base of
// the source it was computed for, resetting the color at the end of every run.
func (th *theme) paint(sb *strings.Builder, s string, cls []class, base int) {
	for i := 0; i < len(s); {
		c := classAt(cls, base+i)
		j := i + 1
		for j < len(s) && classAt(cls, base+j) == c {
			j++
		}
		if th[c] == "" {
			sb.WriteString(s[i:j])
		} else {
			sb.WriteString(th[c])
			sb.WriteString(s[i:j])
			sb.WriteString(reset)
		}
		i = j
	}
}

func classAt(cls []class, i int) class {
	if i < 0 || i >= len(cls) {
		return clsPlain
	}
	return cls[i]
}
