// Package highlight classifies the bytes of Ant source for coloring, as the REPL colors its input
// and ant check its code snippets.
package highlight

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/lexer"
	"github.com/ant-src/ant/internal/token"
)

// A Class is what a byte of source is, for its color.
type Class uint8

const (
	Plain    Class = iota
	Keyword        // a keyword
	Constant       // true, false, null, void
	String         // a string or raw string, including the braces of its interpolations
	Number         // a number literal
	Regex          // a regular expression literal
	Comment        // a comment
	Type           // a namespace, a type, or a module: a capitalized name, or a built-in type
	Function       // a name that is called
	Markup         // an HTML expression, outside its interpolations
	Bad            // what the lexer could not read, or a reserved word
)

// Classes returns the class of every byte of src, which may be unfinished: the lexer goes on
// past what it cannot read.
func Classes(src string) []Class { return classes(src, false) }

// FileClasses is Classes for the whole text of a file, whose first line may be a #! line.
func FileClasses(src string) []Class { return classes(src, true) }

func classes(src string, file bool) []Class {
	cls := make([]Class, len(src))
	toks, comments, _ := lexer.LexRecover(src, file)
	for _, c := range comments {
		fill(cls, c.Off, c.End, Comment)
	}
	classify(cls, toks)
	return cls
}

func fill(cls []Class, from, to int, c Class) {
	for i := max(from, 0); i < to && i < len(cls); i++ {
		cls[i] = c
	}
}

// classify colors a token stream: that of the input, of an interpolation, or of an expression
// in markup.
func classify(cls []Class, toks []token.Token) {
	for i, t := range toks {
		switch t.Kind {
		case token.Keyword:
			switch t.Text {
			case "true", "false", "null", "void":
				fill(cls, t.Off, t.End, Constant)
			default:
				fill(cls, t.Off, t.End, Keyword)
			}
		case token.Reserved, token.Bad:
			fill(cls, t.Off, t.End, Bad)
		case token.Ident:
			switch {
			case next(toks, i).Is("("):
				fill(cls, t.Off, t.End, Function)
			case capitalized(t.Text) || builtinTypes[t.Text]:
				// namespaces, types, and modules are capitalized by convention (section 10)
				fill(cls, t.Off, t.End, Type)
			}
		case token.Int, token.Float:
			fill(cls, t.Off, t.End, Number)
		case token.String:
			fill(cls, t.Off, t.End, String)
		case token.Template:
			fill(cls, t.Off, t.End, String)
			for _, p := range t.Parts {
				if p.Tokens != nil {
					// the expression keeps its own classes between the ${ and the }
					fill(cls, p.Off+2, p.End-1, Plain)
					classify(cls, p.Tokens)
				}
			}
		case token.Regex:
			fill(cls, t.Off, t.End, Regex)
		case token.Html:
			fill(cls, t.Off, t.End, Markup)
			if t.Markup != nil {
				classifyMarkup(cls, t.Markup)
			}
		}
	}
}

// classifyMarkup colors the expressions inside markup: interpolations and attribute values.
func classifyMarkup(cls []Class, m *token.Markup) {
	for _, n := range m.Nodes {
		switch n.Kind {
		case token.MarkupExpr:
			fill(cls, n.Off+1, n.End-1, Plain)
			classify(cls, n.Tokens)
		case token.MarkupElement:
			for _, a := range n.Attrs {
				if a.Tokens != nil {
					fill(cls, a.Off+1, a.End-1, Plain)
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

// ANSI colors each class with the terminal's own sixteen colors, so that it suits a light or a
// dark background alike.
var ANSI = [Bad + 1]string{
	Keyword:  "\x1b[35m", // magenta
	Constant: "\x1b[33m", // yellow
	String:   "\x1b[32m", // green
	Number:   "\x1b[33m", // yellow
	Regex:    "\x1b[31m", // red
	Comment:  "\x1b[90m", // bright black
	Type:     "\x1b[36m", // cyan
	Function: "\x1b[34m", // blue
	Markup:   "\x1b[36m", // cyan
	Bad:      "\x1b[31m", // red
}

// Reset ends a color.
const Reset = "\x1b[0m"

// Paint writes s colored by the class of each of its bytes, which cls gives from byte base of
// the source it was computed for; colors gives the escape sequence that starts each class, or ""
// for none. It resets the color at the end of every run.
func Paint(sb *strings.Builder, s string, cls []Class, base int, colors []string) {
	for i := 0; i < len(s); {
		c := classAt(cls, base+i)
		j := i + 1
		for j < len(s) && classAt(cls, base+j) == c {
			j++
		}
		if int(c) >= len(colors) || colors[c] == "" {
			sb.WriteString(s[i:j])
		} else {
			sb.WriteString(colors[c])
			sb.WriteString(s[i:j])
			sb.WriteString(Reset)
		}
		i = j
	}
}

func classAt(cls []Class, i int) Class {
	if i < 0 || i >= len(cls) {
		return Plain
	}
	return cls[i]
}
