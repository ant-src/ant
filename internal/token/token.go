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
	Regex         // a regular expression literal; Text holds the source as written, Flags the sorted flags
	Html          // an HTML expression; Markup holds its tree
	// Bad is source the recovering lexer could not read as a token (lexer.LexRecover), kept so
	// that the parser can step over it; Text holds it as written.
	Bad
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
	case Regex:
		return "regular expression"
	case Html:
		return "HTML expression"
	case Bad:
		return "malformed source"
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
	Flags string  // for Regex
	// Markup is the fragment an Html token holds: an element is a fragment of one node.
	Markup *Markup
	Pos    Pos
	// EndPos is the position just after the token, so Pos and EndPos span it; a Newline is empty.
	EndPos Pos
	// Off and End are the token's byte span in the source: the source between them is the token
	// as written. A Newline is empty, at its line break.
	Off, End int
}

// Part is one piece of a template string: literal text, or the tokens of an interpolation
// (ending with EOF).
type Part struct {
	Text   string
	Tokens []Token
	Pos    Pos
	EndPos Pos // just after the part
	// Off and End are the part's byte span in the source: the text as written, or an
	// interpolation from its $ to its closing brace.
	Off, End int
}

// Markup is the content of an HTML expression as the lexer read it (section 5): its text decoded and
// its white space collapsed, its interpolations as token streams for the parser.
type Markup struct {
	Nodes []MarkupNode
	// Close is the position of the end tag that closed the content, </name> or </>, and zero for
	// the children of a self-closing element.
	Close Pos
}

// MarkupKind is the kind of a MarkupNode.
type MarkupKind uint8

const (
	MarkupText    MarkupKind = iota // Text holds the decoded text
	MarkupElement                   // Name, Attrs, and Children
	MarkupExpr                      // an interpolation; Tokens ends with EOF
)

// MarkupNode is one item of content.
type MarkupNode struct {
	Kind     MarkupKind
	Name     string
	Attrs    []MarkupAttr
	Children *Markup
	Text     string
	Tokens   []Token
	// TextOnly marks a <title> or <textarea> outside <svg> and <math>, whose content may not be Html.
	TextOnly bool
	Pos      Pos
	EndPos   Pos // just after the node
	// Off and End are the node's byte span in the source.
	Off, End int
}

// MarkupAttrKind is the form of a MarkupAttr.
type MarkupAttrKind uint8

const (
	AttrText   MarkupAttrKind = iota // name="text"; Text holds the decoded text
	AttrBare                         // name alone
	AttrExpr                         // name={expr}
	AttrSpread                       // {...expr}
)

// MarkupAttr is an attribute as written; an expression's tokens end with EOF.
type MarkupAttr struct {
	Kind   MarkupAttrKind
	Name   string
	Text   string
	Tokens []Token
	Pos    Pos
	EndPos Pos // just after the attribute
	// Off and End are the byte span of the expression, from its { to its }.
	Off, End int
}

// Comment is a comment as written, kept only by lexer.LexLayout.
type Comment struct {
	Text     string // the comment's source, from // or /* to the end of the line or */
	Pos      Pos
	EndPos   Pos // just after the comment
	Off, End int
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
	case Regex:
		return "regular expression literal"
	case Html:
		return "HTML expression"
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
	"null": true, "return": true, "select": true, "true": true, "try": true, "type": true,
	"var": true, "void": true, "while": true,
}

var reserved = map[string]bool{
	"as": true, "async": true, "await": true, "catch": true, "class": true, "const": true,
	"enum": true, "finally": true, "function": true, "interface": true, "let": true,
	"match": true, "switch": true, "this": true, "throw": true, "undefined": true,
}

// IsKeyword reports whether word is a keyword.
func IsKeyword(word string) bool { return keywords[word] }

// IsIdentByte reports whether c can be part of an identifier: a letter, a digit, or an
// underscore (section 2).
func IsIdentByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}

// IsReserved reports whether word is a reserved word.
func IsReserved(word string) bool { return reserved[word] }
