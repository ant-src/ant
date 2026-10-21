// Package token defines the tokens the lexer produces and the parser consumes.
package token

import "fmt"

// Kind classifies a token.
type Kind int

const (
	EOF      Kind = iota
	Newline       // a line break that ends a statement (section 2)
	Ident         // an identifier, or any word after "." or "?."
	Keyword       // one of the keywords
	Reserved      // one of the reserved words
	Int           // an integer literal; Value holds the int64
	Float         // a float literal; Float holds the float64
	String        // a string literal; Text holds the unescaped value
	Template      // a template string; Parts holds its pieces
	Op            // an operator or punctuation; Text holds it
)

func (k Kind) String() string {
	switch k {
	case EOF:
		return "end of file"
	case Newline:
		return "newline"
	case Ident:
		return "identifier"
	case Keyword:
		return "keyword"
	case Reserved:
		return "reserved word"
	case Int:
		return "integer"
	case Float:
		return "float"
	case String:
		return "string"
	case Template:
		return "template string"
	case Op:
		return "operator"
	}
	return "token"
}

// Pos is a position in a source file: 1-based line and 1-based byte column.
type Pos struct {
	Line, Col int
}

// Token is one lexical token.
type Token struct {
	Kind  Kind
	Text  string  // the word, operator, or unescaped string value
	Value int64   // for Int
	Float float64 // for Float
	Parts []Part  // for Template
	Pos   Pos
}

// Part is one piece of a template string: literal text, or the tokens of an interpolation
// (ending with EOF).
type Part struct {
	Text   string
	Tokens []Token
	Pos    Pos
}

// String describes the token for error messages.
func (t Token) String() string {
	switch t.Kind {
	case Ident:
		return fmt.Sprintf("identifier %q", t.Text)
	case Keyword, Reserved, Op:
		return fmt.Sprintf("'%s'", t.Text)
	case Int, Float:
		return t.Text
	case String:
		return "string literal"
	case Template:
		return "template string"
	}
	return t.Kind.String()
}

// Is reports whether the token is the keyword or operator with the given text.
func (t Token) Is(text string) bool {
	return (t.Kind == Keyword || t.Kind == Op) && t.Text == text
}

var keywords = map[string]bool{
	"break": true, "continue": true, "defer": true, "else": true, "export": true, "false": true,
	"for": true, "func": true, "go": true, "if": true, "import": true, "in": true, "is": true,
	"null": true, "return": true, "true": true, "try": true, "type": true, "var": true,
	"void": true, "while": true,
}

var reserved = map[string]bool{
	"as": true, "async": true, "await": true, "catch": true, "class": true, "const": true,
	"enum": true, "finally": true, "function": true, "interface": true, "let": true,
	"match": true, "switch": true, "this": true, "throw": true, "undefined": true,
}

// IsKeyword reports whether word is a keyword.
func IsKeyword(word string) bool { return keywords[word] }

// IsReserved reports whether word is a reserved word.
func IsReserved(word string) bool { return reserved[word] }
