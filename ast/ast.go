// Package ast declares the syntax tree of an Ant module, as the parser builds it and the
// resolver annotates it.
package ast

import "github.com/ant-src/ant/value"

// Pos is a position in the module's source: 1-based line and 1-based byte column.
type Pos struct {
	Line, Col int
}

// Node is any syntax node. Position is where it starts, as messages report it, and EndPosition
// is just after its last token, so the two span the node; Start is where it begins on the page,
// which for a postfix chain or an operator is its first operand.
type Node interface {
	Position() Pos
	EndPosition() Pos
}

// Expr is an expression. Every construct of section 5 is an expression, including if, loops,
// go, defer, return, break, continue, and assignment.
type Expr interface {
	Stmt
	exprNode()
}

// Stmt is a statement in a block: an expression or a var declaration.
type Stmt interface {
	Node
	stmtNode()
}

// Type is a type expression, read by the checker and discarded by the interpreter except for
// the field names of a record type.
type Type interface {
	Node
	typeNode()
}

// Target is the left side of a var declaration or a for: a name or a destructuring.
type Target interface {
	Node
	targetNode()
}

// ---- Module

// File is a parsed module.
type File struct {
	Path    string
	Doc     string // the module doc, from the //! lines before the first import (section 2), or ""
	Imports []*Import
	Decls   []Decl
	// Stmts holds the top-level var declarations and statements in source order: a script's
	// body (section 10), or a session input's or a spec example's statements.
	Stmts []Stmt
	// Script is set for a program module with a statement at the top level; FirstStmt is where
	// the first one starts.
	Script    bool
	FirstStmt Pos
	// Declarations is set for a declaration file (section 10, Declaration files), parsed in
	// the parser's Declarations mode: its funcs have no bodies and its names join the prelude.
	Declarations bool
	// End is the position after the last token of the file.
	End Pos
	// Scope is filled by the resolver: the module record's fields.
	Scope *ModuleScope
}

// Import is one import declaration.
type Import struct {
	Pos          Pos
	End          Pos
	Path         string
	PathPos      Pos // the path string, from its opening quote
	PathEnd      Pos
	Namespace    string       // the bare name that binds the whole module, or ""
	NamespacePos Pos          // where Namespace is written, or zero
	Names        []ImportName // the destructured names, when Namespace is ""
	// Module is filled by the loader: the index of the imported module in the program, or -1
	// when the module could not be loaded and the loader went on (Config.Recover).
	Module int
}

// ImportName is one entry of an import list: the export, and the local name it binds.
type ImportName struct {
	Pos      Pos
	End      Pos
	Name     string
	Alias    string // the local name; equals Name when not renamed
	AliasPos Pos    // where the alias is written, or zero when not renamed
}

// Decl is a top-level declaration.
type Decl interface {
	Node
	declNode()
	Exported() bool
}

// VarDecl is a var declaration, at the top level or in a block. In declarations mode a
// constant such as var Math.PI: float64 has an Owner and no Init.
type VarDecl struct {
	Pos      Pos
	End      Pos
	Export   bool
	Owner    string // the namespace of a declaration file's constant, var Owner.NAME: T, or ""
	OwnerPos Pos    // where Owner is written, or zero
	Target   Target
	Type     Type // nil when inferred; only with a NameTarget
	Init     Expr
	Doc      string // the doc comment (section 2) of a top-level var, or ""
	Sig      string // the source from var through the type, as written, without the initializer
}

// FuncDecl is a named function, or a function declared on a type when Owner is set.
type FuncDecl struct {
	Pos        Pos
	End        Pos
	Export     bool
	Owner      string // the type for func Type.name, or ""
	OwnerPos   Pos    // where Owner is written, or zero
	Name       string
	NamePos    Pos
	TypeParams []TypeParam
	Params     []Param
	Return     Type   // nil when inferred
	Body       *Block // nil only in declarations mode
	Doc        string // the doc comment (section 2), or ""
	Sig        string // the source from func through the result type, as written
	// Scope is filled by the resolver.
	Scope *FuncScope
}

// TypeDecl is a type declaration. In declarations mode Owner may name the namespace the type
// belongs to, as in type Json.Value, and Body may be nil for an opaque type such as Mutex.
type TypeDecl struct {
	Pos        Pos
	End        Pos
	Export     bool
	Owner      string
	OwnerPos   Pos // where Owner is written, or zero
	Name       string
	NamePos    Pos
	TypeParams []TypeParam
	Body       Type
	Doc        string // the doc comment (section 2), or ""
	Sig        string // the whole declaration from type on, as written
	// Slot is filled by the resolver for a type declared in a block.
	Slot Slot
}

// TypeParam is a type parameter with an optional bound.
type TypeParam struct {
	Pos   Pos
	End   Pos
	Name  string
	Bound Type
}

// Param is a function parameter.
type Param struct {
	Pos     Pos
	End     Pos
	Name    string
	Type    Type
	Default Expr
	// Slot is filled by the resolver.
	Slot Slot
}

// ---- Types

// TypeName is a named type with optional type arguments: int, Array[T], Geometry.Point, and
// also null and void.
type TypeName struct {
	Pos     Pos
	End     Pos
	Path    []string
	PathPos []Pos // where each part of Path is written
	Args    []Type
}

// RecordType is an anonymous record type { name: T, ... }.
type RecordType struct {
	Pos    Pos
	End    Pos
	Fields []FieldType
}

// FieldType is one field of a record type. Default is the expression a literal that leaves the
// field out evaluates, or nil; only the body of a type declaration has defaults (section 4).
type FieldType struct {
	Pos     Pos
	End     Pos
	Name    string
	Type    Type
	Default Expr
}

// FuncType is a function type func(T, U): R.
type FuncType struct {
	Pos    Pos
	End    Pos
	Params []FuncTypeParam
	Return Type
}

// FuncTypeParam is one parameter of a function type, with an optional documenting name.
type FuncTypeParam struct {
	Name string
	Type Type
}

// UnionType is A | B.
type UnionType struct {
	Pos     Pos
	End     Pos
	Members []Type
}

// ---- Targets

// NameTarget binds a name; "_" binds nothing.
type NameTarget struct {
	Pos  Pos
	End  Pos
	Name string
	// Slot is filled by the resolver.
	Slot Slot
}

// RecordTarget is a record destructuring { x, y: py, ... }.
type RecordTarget struct {
	Pos    Pos
	End    Pos
	Fields []RecordTargetField
}

// RecordTargetField is one entry of a record destructuring: the field, and the target it binds
// (a NameTarget with the field's name for the shorthand).
type RecordTargetField struct {
	Pos    Pos
	End    Pos
	Name   string
	Target Target
}

// ArrayTarget is an array destructuring [a, b, ...rest].
type ArrayTarget struct {
	Pos   Pos
	End   Pos
	Elems []Target
	Rest  *NameTarget // nil without a rest
}

// ---- Expressions

type (
	// Ident is a variable reference. Ref is filled by the resolver.
	Ident struct {
		Pos  Pos
		End  Pos
		Name string
		Ref  Ref
	}
	IntLit struct {
		Pos   Pos
		End   Pos
		Value int64
	}
	FloatLit struct {
		Pos   Pos
		End   Pos
		Value float64
	}
	StringLit struct {
		Pos   Pos
		End   Pos
		Value string
	}
	BoolLit struct {
		Pos   Pos
		End   Pos
		Value bool
	}
	NullLit struct{ Pos, End Pos }
	// RegexLit is a regular expression literal (section 2). Pattern is the value it evaluates
	// to, built once when the literal is parsed.
	RegexLit struct {
		Pos     Pos
		End     Pos
		Source  string // normalized
		Flags   string // sorted
		Pattern *value.Regex
	}
	VoidLit struct{ Pos, End Pos }
	// Template is a template string; each part is literal text or an interpolated expression.
	Template struct {
		Pos   Pos
		End   Pos
		Parts []TemplatePart
	}
	// Html is an HTML expression (section 5): an element, or a fragment, as its content.
	Html struct {
		Pos     Pos
		End     Pos
		Content []HtmlNode
		Close   Pos // the </> of a fragment, or zero for an element
	}
	// RecordLit is Head{ entries }, where Head is an identifier path.
	RecordLit struct {
		Pos     Pos
		End     Pos
		Head    Expr
		Entries []RecordEntry
	}
	// MapLit is Map{ entries }.
	MapLit struct {
		Pos     Pos
		End     Pos
		Entries []MapEntry
	}
	// ArrayLit is [ elements ].
	ArrayLit struct {
		Pos   Pos
		End   Pos
		Elems []ArrayElem
		// Off is the byte offset of the [, and SpreadComma reports a comma after a final
		// spread, for the parser when the literal turns out to be a destructuring assignment.
		Off         int
		SpreadComma bool
	}
	// Lambda is a bare block with the implicit parameter x.
	Lambda struct {
		Pos   Pos
		End   Pos
		Body  *Block
		Scope *FuncScope
	}
	// FuncLit is an anonymous func.
	FuncLit struct {
		Pos    Pos
		End    Pos
		Params []Param
		Return Type
		Body   *Block
		Scope  *FuncScope
	}
	// Field is a.name or a?.name. Grouped marks one written in parentheses, which ends its
	// postfix chain: a ?. inside skips nothing outside (section 12).
	Field struct {
		Pos      Pos // the . or ?.
		End      Pos
		X        Expr
		Name     string
		NamePos  Pos
		Optional bool
		Grouped  bool
	}
	// Index is a[i] or a?.[i]; Grouped as for Field.
	Index struct {
		Pos      Pos
		End      Pos
		X        Expr
		Index    Expr
		Optional bool
		Grouped  bool
	}
	// Call is f(args) or f?.(args). A pipe a | f(b) is a Call with Pipe set and a as the first
	// argument; it evaluates that argument before f. Grouped as for Field.
	Call struct {
		Pos      Pos
		End      Pos
		Fn       Expr
		Args     []Arg
		Optional bool
		Pipe     bool
		Grouped  bool
	}
	// Unary is -a, !a, or try a.
	Unary struct {
		Pos Pos
		End Pos
		Op  string
		X   Expr
	}
	// Binary is an arithmetic, comparison, equality, logical, or ?? operation.
	Binary struct {
		Pos  Pos
		End  Pos
		Op   string
		X, Y Expr
	}
	// Is is x is T. Kind is set for a kind test (null, void, bool, string, func, Array, Map, and
	// the numeric names); otherwise Path is the type's expression.
	Is struct {
		Pos     Pos // the is
		End     Pos
		X       Expr
		Kind    string
		KindPos Pos // where Kind is written, or zero
		Path    Expr
	}
	// Assign is target op value. Target is an Ident, Field, Index, or ArrayTarget.
	Assign struct {
		Pos    Pos
		End    Pos
		Op     string
		Target Node
		Value  Expr
	}
	// AssignArray is the array destructuring target of an assignment: places and nested
	// targets, with an optional rest place.
	AssignArray struct {
		Pos   Pos
		End   Pos
		Elems []Node // Ident, Field, Index, or *AssignArray
		Rest  Expr   // Ident, Field, or Index; nil without a rest
	}
	If struct {
		Pos  Pos
		End  Pos
		Cond Expr
		Then *Block
		Else Node // *If, *Block, or nil
	}
	While struct {
		Pos      Pos
		End      Pos
		Cond     Expr
		Body     *Block
		Assigned []string // the locals declared outside the loop that its body assigns, filled by the resolver
	}
	For struct {
		Pos      Pos
		End      Pos
		Target   Target
		Iter     Expr
		Body     *Block
		Scope    *BlockScope // the per-iteration scope, filled by the resolver
		Assigned []string    // the locals declared outside the loop that its body assigns, filled by the resolver
	}
	// Select is a select expression (section 5): clauses with a head each, and at most one
	// else, last.
	Select struct {
		Pos     Pos
		End     Pos
		Clauses []*SelectClause
	}
	// SelectClause is one clause of a select. Head is the call that names the wait, one of
	// Channel.receive, Channel.send, Context.wait, WaitGroup.wait and Time.sleep, which the parser
	// checks, and nil for else. Target binds the head's result when the clause has var x =.
	// Scope is the body's scope with that binding, filled by the resolver, nil when it declares
	// nothing.
	SelectClause struct {
		Pos    Pos
		End    Pos
		Target *NameTarget
		Head   *Call
		Body   *Block
		Scope  *BlockScope
	}
	Go struct {
		Pos   Pos
		End   Pos
		Body  *Block
		Scope *FuncScope
	}
	Defer struct {
		Pos   Pos
		End   Pos
		Body  *Block
		Scope *FuncScope
	}
	Return struct {
		Pos   Pos
		End   Pos
		Value Expr // nil for a bare return
	}
	Break    struct{ Pos, End Pos }
	Continue struct{ Pos, End Pos }
	// Placeholder is the "..." of spec mode.
	Placeholder struct{ Pos, End Pos }
	// BadExpr, BadStmt, and BadType stand for source that did not parse, in a tree the parser
	// built past a syntax error (parser.ParseRecover). They have no meaning: the checker gives a
	// BadExpr the type Invalid and reports nothing about it.
	BadExpr struct{ Pos, End Pos }
	BadStmt struct{ Pos, End Pos }
	BadType struct{ Pos, End Pos }
	// Block is { statements }.
	Block struct {
		Pos      Pos
		End      Pos
		Stmts    []Stmt
		HasDefer bool        // a defer statement registers directly with this block
		Types    []*TypeDecl // the types declared directly in the block, bound when it starts
		Scope    *BlockScope // filled by the resolver; nil when the block declares nothing
	}
)

// HtmlNode is an item of an HTML expression's content: decoded text, an interpolated Expr, or
// an Element.
type HtmlNode struct {
	Pos     Pos
	End     Pos
	Text    string
	Expr    Expr
	Element *HtmlElement
}

// HtmlElement is an element as written. TextOnly marks a <title> or <textarea> outside <svg>
// and <math>, whose content may not hold Html. Close is where its end tag starts, or zero when
// it is written self-closing.
type HtmlElement struct {
	Pos      Pos // the <
	End      Pos
	Name     string
	NamePos  Pos
	Attrs    []HtmlAttr
	Content  []HtmlNode
	Close    Pos
	TextOnly bool
}

// HtmlAttr is an attribute: literal Text when Expr is nil, a value when Expr is set, and every
// entry of a map when Spread is set, in which case Name is empty.
type HtmlAttr struct {
	Pos    Pos
	End    Pos
	Name   string
	Text   string
	Expr   Expr
	Spread bool
}

// TemplatePart is literal text or an interpolated expression; an interpolation spans from its $
// to its closing brace.
type TemplatePart struct {
	Pos  Pos
	End  Pos
	Text string
	Expr Expr
}

// RecordEntry is a field of a record literal, or a spread.
type RecordEntry struct {
	Pos    Pos
	End    Pos
	Name   string
	Value  Expr
	Spread Expr
}

// MapEntry is one key and value of a map literal.
type MapEntry struct {
	Pos   Pos
	End   Pos
	Key   Expr
	Value Expr
}

// ArrayElem is an element of an array literal, or a spread.
type ArrayElem struct {
	Value  Expr
	Spread bool
}

// Arg is a call argument, named when Name is set.
type Arg struct {
	Pos   Pos
	End   Pos
	Name  string
	Value Expr
}

// ---- Resolver annotations

// RefKind says what an identifier refers to.
type RefKind int

const (
	RefUnresolved RefKind = iota
	RefLocal              // a slot Depth environments up
	RefModule             // a field of this module's record
	RefImport             // a field of an imported module's record
	RefPrelude            // a prelude name
	RefUnbound            // no scope binds the name: a panic when evaluated
	RefDiscard            // the write-only _: a panic when read
)

// Ref is where an identifier resolves to.
type Ref struct {
	Kind   RefKind
	Depth  int  // RefLocal: environments to walk up
	Index  int  // RefLocal: slot index; RefModule: field index; RefImport: import index
	Shared bool // RefLocal: the slot is shared (captured by a nested body)
	Name   string
}

// Slot is where a binding lives in its environment, with the facts narrowing needs.
type Slot struct {
	Index  int
	Shared bool
	Module bool // a module field rather than a local slot
	// Reassigned is set when the binding is assigned after its declaration anywhere;
	// AssignedNested when it is assigned inside a nested body (section 7).
	Reassigned     bool
	AssignedNested bool
	// Discard is set for _, which binds nothing: a value stored there is dropped.
	Discard bool
}

// BlockScope describes the environment a block, loop iteration, or function allocates.
type BlockScope struct {
	Plain  int // number of plain slots
	Shared int // number of shared slots
}

// FuncScope describes a function body, lambda, go block, or defer block.
type FuncScope struct {
	BlockScope
	// X is the slot of the implicit lambda parameter.
	X Slot
}

// ModuleScope lists the fields of the module record in declaration order.
type ModuleScope struct {
	Fields []string
	Index  map[string]int
	// Types holds, in a session, the type declared by each field that names one, so a later
	// input can declare functions on it.
	Types map[string]*TypeDecl
}

// ---- Interface plumbing

func (n *File) Position() Pos              { return Pos{1, 1} }
func (n *File) EndPosition() Pos           { return n.End }
func (n *Import) Position() Pos            { return n.Pos }
func (n *VarDecl) Position() Pos           { return n.Pos }
func (n *FuncDecl) Position() Pos          { return n.Pos }
func (n *TypeDecl) Position() Pos          { return n.Pos }
func (n *TypeName) Position() Pos          { return n.Pos }
func (n *RecordType) Position() Pos        { return n.Pos }
func (n *FuncType) Position() Pos          { return n.Pos }
func (n *UnionType) Position() Pos         { return n.Pos }
func (n *NameTarget) Position() Pos        { return n.Pos }
func (n *RecordTarget) Position() Pos      { return n.Pos }
func (n *ArrayTarget) Position() Pos       { return n.Pos }
func (n *Ident) Position() Pos             { return n.Pos }
func (n *IntLit) Position() Pos            { return n.Pos }
func (n *FloatLit) Position() Pos          { return n.Pos }
func (n *StringLit) Position() Pos         { return n.Pos }
func (n *RegexLit) Position() Pos          { return n.Pos }
func (n *BoolLit) Position() Pos           { return n.Pos }
func (n *NullLit) Position() Pos           { return n.Pos }
func (n *VoidLit) Position() Pos           { return n.Pos }
func (n *Template) Position() Pos          { return n.Pos }
func (n *Html) Position() Pos              { return n.Pos }
func (n *RecordLit) Position() Pos         { return n.Pos }
func (n *MapLit) Position() Pos            { return n.Pos }
func (n *ArrayLit) Position() Pos          { return n.Pos }
func (n *Lambda) Position() Pos            { return n.Pos }
func (n *FuncLit) Position() Pos           { return n.Pos }
func (n *Field) Position() Pos             { return n.Pos }
func (n *Index) Position() Pos             { return n.Pos }
func (n *Call) Position() Pos              { return n.Pos }
func (n *Unary) Position() Pos             { return n.Pos }
func (n *Binary) Position() Pos            { return n.Pos }
func (n *Is) Position() Pos                { return n.Pos }
func (n *Assign) Position() Pos            { return n.Pos }
func (n *AssignArray) Position() Pos       { return n.Pos }
func (n *If) Position() Pos                { return n.Pos }
func (n *While) Position() Pos             { return n.Pos }
func (n *For) Position() Pos               { return n.Pos }
func (n *Go) Position() Pos                { return n.Pos }
func (n *Select) Position() Pos            { return n.Pos }
func (n *Defer) Position() Pos             { return n.Pos }
func (n *Return) Position() Pos            { return n.Pos }
func (n *Break) Position() Pos             { return n.Pos }
func (n *Continue) Position() Pos          { return n.Pos }
func (n *Placeholder) Position() Pos       { return n.Pos }
func (n *Block) Position() Pos             { return n.Pos }
func (n *BadExpr) Position() Pos           { return n.Pos }
func (n *BadStmt) Position() Pos           { return n.Pos }
func (n *BadType) Position() Pos           { return n.Pos }
func (n *ImportName) Position() Pos        { return n.Pos }
func (n *TypeParam) Position() Pos         { return n.Pos }
func (n *Param) Position() Pos             { return n.Pos }
func (n *FieldType) Position() Pos         { return n.Pos }
func (n *RecordTargetField) Position() Pos { return n.Pos }
func (n *HtmlNode) Position() Pos          { return n.Pos }
func (n *HtmlElement) Position() Pos       { return n.Pos }
func (n *HtmlAttr) Position() Pos          { return n.Pos }
func (n *TemplatePart) Position() Pos      { return n.Pos }
func (n *RecordEntry) Position() Pos       { return n.Pos }
func (n *MapEntry) Position() Pos          { return n.Pos }
func (n *Arg) Position() Pos               { return n.Pos }
func (n *SelectClause) Position() Pos      { return n.Pos }

func (n *Import) EndPosition() Pos            { return n.End }
func (n *VarDecl) EndPosition() Pos           { return n.End }
func (n *FuncDecl) EndPosition() Pos          { return n.End }
func (n *TypeDecl) EndPosition() Pos          { return n.End }
func (n *TypeName) EndPosition() Pos          { return n.End }
func (n *RecordType) EndPosition() Pos        { return n.End }
func (n *FuncType) EndPosition() Pos          { return n.End }
func (n *UnionType) EndPosition() Pos         { return n.End }
func (n *NameTarget) EndPosition() Pos        { return n.End }
func (n *RecordTarget) EndPosition() Pos      { return n.End }
func (n *ArrayTarget) EndPosition() Pos       { return n.End }
func (n *Ident) EndPosition() Pos             { return n.End }
func (n *IntLit) EndPosition() Pos            { return n.End }
func (n *FloatLit) EndPosition() Pos          { return n.End }
func (n *StringLit) EndPosition() Pos         { return n.End }
func (n *RegexLit) EndPosition() Pos          { return n.End }
func (n *BoolLit) EndPosition() Pos           { return n.End }
func (n *NullLit) EndPosition() Pos           { return n.End }
func (n *VoidLit) EndPosition() Pos           { return n.End }
func (n *Template) EndPosition() Pos          { return n.End }
func (n *Html) EndPosition() Pos              { return n.End }
func (n *RecordLit) EndPosition() Pos         { return n.End }
func (n *MapLit) EndPosition() Pos            { return n.End }
func (n *ArrayLit) EndPosition() Pos          { return n.End }
func (n *Lambda) EndPosition() Pos            { return n.End }
func (n *FuncLit) EndPosition() Pos           { return n.End }
func (n *Field) EndPosition() Pos             { return n.End }
func (n *Index) EndPosition() Pos             { return n.End }
func (n *Call) EndPosition() Pos              { return n.End }
func (n *Unary) EndPosition() Pos             { return n.End }
func (n *Binary) EndPosition() Pos            { return n.End }
func (n *Is) EndPosition() Pos                { return n.End }
func (n *Assign) EndPosition() Pos            { return n.End }
func (n *AssignArray) EndPosition() Pos       { return n.End }
func (n *If) EndPosition() Pos                { return n.End }
func (n *While) EndPosition() Pos             { return n.End }
func (n *For) EndPosition() Pos               { return n.End }
func (n *Go) EndPosition() Pos                { return n.End }
func (n *Select) EndPosition() Pos            { return n.End }
func (n *Defer) EndPosition() Pos             { return n.End }
func (n *Return) EndPosition() Pos            { return n.End }
func (n *Break) EndPosition() Pos             { return n.End }
func (n *Continue) EndPosition() Pos          { return n.End }
func (n *Placeholder) EndPosition() Pos       { return n.End }
func (n *Block) EndPosition() Pos             { return n.End }
func (n *BadExpr) EndPosition() Pos           { return n.End }
func (n *BadStmt) EndPosition() Pos           { return n.End }
func (n *BadType) EndPosition() Pos           { return n.End }
func (n *ImportName) EndPosition() Pos        { return n.End }
func (n *TypeParam) EndPosition() Pos         { return n.End }
func (n *Param) EndPosition() Pos             { return n.End }
func (n *FieldType) EndPosition() Pos         { return n.End }
func (n *RecordTargetField) EndPosition() Pos { return n.End }
func (n *HtmlNode) EndPosition() Pos          { return n.End }
func (n *HtmlElement) EndPosition() Pos       { return n.End }
func (n *HtmlAttr) EndPosition() Pos          { return n.End }
func (n *TemplatePart) EndPosition() Pos      { return n.End }
func (n *RecordEntry) EndPosition() Pos       { return n.End }
func (n *MapEntry) EndPosition() Pos          { return n.End }
func (n *Arg) EndPosition() Pos               { return n.End }
func (n *SelectClause) EndPosition() Pos      { return n.End }

func (*VarDecl) declNode()  {}
func (*FuncDecl) declNode() {}
func (*TypeDecl) declNode() {}

func (n *VarDecl) Exported() bool  { return n.Export }
func (n *FuncDecl) Exported() bool { return n.Export }
func (n *TypeDecl) Exported() bool { return n.Export }

func (*VarDecl) stmtNode()     {}
func (*TypeDecl) stmtNode()    {}
func (*Ident) stmtNode()       {}
func (*IntLit) stmtNode()      {}
func (*FloatLit) stmtNode()    {}
func (*StringLit) stmtNode()   {}
func (*RegexLit) stmtNode()    {}
func (*BoolLit) stmtNode()     {}
func (*NullLit) stmtNode()     {}
func (*VoidLit) stmtNode()     {}
func (*Template) stmtNode()    {}
func (*Html) stmtNode()        {}
func (*RecordLit) stmtNode()   {}
func (*MapLit) stmtNode()      {}
func (*ArrayLit) stmtNode()    {}
func (*Lambda) stmtNode()      {}
func (*FuncLit) stmtNode()     {}
func (*Field) stmtNode()       {}
func (*Index) stmtNode()       {}
func (*Call) stmtNode()        {}
func (*Unary) stmtNode()       {}
func (*Binary) stmtNode()      {}
func (*Is) stmtNode()          {}
func (*Assign) stmtNode()      {}
func (*If) stmtNode()          {}
func (*While) stmtNode()       {}
func (*For) stmtNode()         {}
func (*Go) stmtNode()          {}
func (*Select) stmtNode()      {}
func (*Defer) stmtNode()       {}
func (*Return) stmtNode()      {}
func (*Break) stmtNode()       {}
func (*Continue) stmtNode()    {}
func (*Placeholder) stmtNode() {}
func (*Block) stmtNode()       {}
func (*BadExpr) stmtNode()     {}
func (*BadStmt) stmtNode()     {}

func (*Ident) exprNode()       {}
func (*IntLit) exprNode()      {}
func (*FloatLit) exprNode()    {}
func (*StringLit) exprNode()   {}
func (*RegexLit) exprNode()    {}
func (*BoolLit) exprNode()     {}
func (*NullLit) exprNode()     {}
func (*VoidLit) exprNode()     {}
func (*Template) exprNode()    {}
func (*Html) exprNode()        {}
func (*RecordLit) exprNode()   {}
func (*MapLit) exprNode()      {}
func (*ArrayLit) exprNode()    {}
func (*Lambda) exprNode()      {}
func (*FuncLit) exprNode()     {}
func (*Field) exprNode()       {}
func (*Index) exprNode()       {}
func (*Call) exprNode()        {}
func (*Unary) exprNode()       {}
func (*Binary) exprNode()      {}
func (*Is) exprNode()          {}
func (*Assign) exprNode()      {}
func (*If) exprNode()          {}
func (*While) exprNode()       {}
func (*For) exprNode()         {}
func (*Go) exprNode()          {}
func (*Select) exprNode()      {}
func (*Defer) exprNode()       {}
func (*Return) exprNode()      {}
func (*Break) exprNode()       {}
func (*Continue) exprNode()    {}
func (*Placeholder) exprNode() {}
func (*Block) exprNode()       {}
func (*BadExpr) exprNode()     {}

func (*TypeName) typeNode()   {}
func (*RecordType) typeNode() {}
func (*FuncType) typeNode()   {}
func (*UnionType) typeNode()  {}
func (*BadType) typeNode()    {}

func (*NameTarget) targetNode()   {}
func (*RecordTarget) targetNode() {}
func (*ArrayTarget) targetNode()  {}
