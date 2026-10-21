// Package ast declares the syntax tree of an Ant module, as the parser builds it and the
// resolver annotates it.
package ast

// Pos is a position in the module's source: 1-based line and 1-based byte column.
type Pos struct {
	Line, Col int
}

// Node is any syntax node.
type Node interface {
	Position() Pos
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
	Imports []*Import
	Decls   []Decl
	// Stmts holds top-level statements, which only the spec and session modes allow.
	Stmts []Stmt
	// Scope is filled by the resolver: the module record's fields.
	Scope *ModuleScope
}

// Import is one import declaration.
type Import struct {
	Pos       Pos
	Path      string
	Namespace string       // the bare name that binds the whole module, or ""
	Names     []ImportName // the destructured names, when Namespace is ""
	// Module is filled by the loader: the index of the imported module in the program.
	Module int
}

// ImportName is one entry of an import list: the export, and the local name it binds.
type ImportName struct {
	Pos   Pos
	Name  string
	Alias string // the local name; equals Name when not renamed
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
	Pos    Pos
	Export bool
	Owner  string
	Target Target
	Type   Type // nil when inferred; only with a NameTarget
	Init   Expr
}

// FuncDecl is a named function, or a function declared on a type when Owner is set.
type FuncDecl struct {
	Pos        Pos
	Export     bool
	Owner      string // the type for func Type.name, or ""
	Name       string
	TypeParams []TypeParam
	Params     []Param
	Return     Type   // nil when inferred
	Body       *Block // nil only in declarations mode
	// Scope is filled by the resolver.
	Scope *FuncScope
}

// TypeDecl is a type declaration. In declarations mode Owner may name the namespace the type
// belongs to, as in type Json.Value, and Body may be nil for an opaque type such as Mutex.
type TypeDecl struct {
	Pos        Pos
	Export     bool
	Owner      string
	Name       string
	TypeParams []TypeParam
	Body       Type
}

// TypeParam is a type parameter with an optional bound.
type TypeParam struct {
	Pos   Pos
	Name  string
	Bound Type
}

// Param is a function parameter.
type Param struct {
	Pos     Pos
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
	Pos  Pos
	Path []string
	Args []Type
}

// RecordType is an anonymous record type { name: T, ... }.
type RecordType struct {
	Pos    Pos
	Fields []FieldType
}

// FieldType is one field of a record type.
type FieldType struct {
	Pos  Pos
	Name string
	Type Type
}

// FuncType is a function type func(T, U): R.
type FuncType struct {
	Pos    Pos
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
	Members []Type
}

// ---- Targets

// NameTarget binds a name; "_" binds nothing.
type NameTarget struct {
	Pos  Pos
	Name string
	// Slot is filled by the resolver.
	Slot Slot
}

// RecordTarget is a record destructuring { x, y: py, ... }.
type RecordTarget struct {
	Pos    Pos
	Fields []RecordTargetField
}

// RecordTargetField is one entry of a record destructuring: the field, and the target it binds
// (a NameTarget with the field's name for the shorthand).
type RecordTargetField struct {
	Pos    Pos
	Name   string
	Target Target
}

// ArrayTarget is an array destructuring [a, b, ...rest].
type ArrayTarget struct {
	Pos   Pos
	Elems []Target
	Rest  *NameTarget // nil without a rest
}

// ---- Expressions

type (
	// Ident is a variable reference. Ref is filled by the resolver.
	Ident struct {
		Pos  Pos
		Name string
		Ref  Ref
	}
	IntLit struct {
		Pos   Pos
		Value int64
	}
	FloatLit struct {
		Pos   Pos
		Value float64
	}
	StringLit struct {
		Pos   Pos
		Value string
	}
	BoolLit struct {
		Pos   Pos
		Value bool
	}
	NullLit struct{ Pos Pos }
	VoidLit struct{ Pos Pos }
	// Template is a template string; each part is literal text or an interpolated expression.
	Template struct {
		Pos   Pos
		Parts []TemplatePart
	}
	// RecordLit is Head{ entries }, where Head is an identifier path.
	RecordLit struct {
		Pos     Pos
		Head    Expr
		Entries []RecordEntry
	}
	// MapLit is Map{ entries }.
	MapLit struct {
		Pos     Pos
		Entries []MapEntry
	}
	// ArrayLit is [ elements ].
	ArrayLit struct {
		Pos   Pos
		Elems []ArrayElem
	}
	// Lambda is a bare block with the implicit parameter x.
	Lambda struct {
		Pos   Pos
		Body  *Block
		Scope *FuncScope
	}
	// FuncLit is an anonymous func.
	FuncLit struct {
		Pos    Pos
		Params []Param
		Return Type
		Body   *Block
		Scope  *FuncScope
	}
	// Field is a.name or a?.name.
	Field struct {
		Pos      Pos
		X        Expr
		Name     string
		Optional bool
	}
	// Index is a[i] or a?.[i].
	Index struct {
		Pos      Pos
		X        Expr
		Index    Expr
		Optional bool
	}
	// Call is f(args) or f?.(args). A pipe a | f(b) is a Call with Pipe set and a as the first
	// argument; it evaluates that argument before f.
	Call struct {
		Pos      Pos
		Fn       Expr
		Args     []Arg
		Optional bool
		Pipe     bool
	}
	// Unary is -a, !a, or try a.
	Unary struct {
		Pos Pos
		Op  string
		X   Expr
	}
	// Binary is an arithmetic, comparison, equality, logical, or ?? operation.
	Binary struct {
		Pos  Pos
		Op   string
		X, Y Expr
	}
	// Is is x is T. Kind is set for a kind test (null, void, bool, string, func, Array, Map, and
	// the numeric names); otherwise Path is the type's expression.
	Is struct {
		Pos  Pos
		X    Expr
		Kind string
		Path Expr
	}
	// Assign is target op value. Target is an Ident, Field, Index, or ArrayTarget.
	Assign struct {
		Pos    Pos
		Op     string
		Target Node
		Value  Expr
	}
	// AssignArray is the array destructuring target of an assignment: places and nested
	// targets, with an optional rest place.
	AssignArray struct {
		Pos   Pos
		Elems []Node // Ident, Field, Index, or *AssignArray
		Rest  Expr   // Ident, Field, or Index; nil without a rest
	}
	If struct {
		Pos  Pos
		Cond Expr
		Then *Block
		Else Node // *If, *Block, or nil
	}
	While struct {
		Pos  Pos
		Cond Expr
		Body *Block
	}
	For struct {
		Pos    Pos
		Target Target
		Iter   Expr
		Body   *Block
		Scope  *BlockScope // the per-iteration scope, filled by the resolver
	}
	Go struct {
		Pos   Pos
		Body  *Block
		Scope *FuncScope
	}
	Defer struct {
		Pos   Pos
		Body  *Block
		Scope *FuncScope
	}
	Return struct {
		Pos   Pos
		Value Expr // nil for a bare return
	}
	Break    struct{ Pos Pos }
	Continue struct{ Pos Pos }
	// Placeholder is the "..." of spec mode.
	Placeholder struct{ Pos Pos }
	// Block is { statements }.
	Block struct {
		Pos      Pos
		Stmts    []Stmt
		HasDefer bool        // a defer statement registers directly with this block
		Scope    *BlockScope // filled by the resolver; nil when the block declares nothing
	}
)

// TemplatePart is literal text or an interpolated expression.
type TemplatePart struct {
	Text string
	Expr Expr
}

// RecordEntry is a field of a record literal, or a spread.
type RecordEntry struct {
	Pos    Pos
	Name   string
	Value  Expr
	Spread Expr
}

// MapEntry is one key and value of a map literal.
type MapEntry struct {
	Pos   Pos
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
}

// ---- Interface plumbing

func (n *File) Position() Pos         { return Pos{1, 1} }
func (n *Import) Position() Pos       { return n.Pos }
func (n *VarDecl) Position() Pos      { return n.Pos }
func (n *FuncDecl) Position() Pos     { return n.Pos }
func (n *TypeDecl) Position() Pos     { return n.Pos }
func (n *TypeName) Position() Pos     { return n.Pos }
func (n *RecordType) Position() Pos   { return n.Pos }
func (n *FuncType) Position() Pos     { return n.Pos }
func (n *UnionType) Position() Pos    { return n.Pos }
func (n *NameTarget) Position() Pos   { return n.Pos }
func (n *RecordTarget) Position() Pos { return n.Pos }
func (n *ArrayTarget) Position() Pos  { return n.Pos }
func (n *Ident) Position() Pos        { return n.Pos }
func (n *IntLit) Position() Pos       { return n.Pos }
func (n *FloatLit) Position() Pos     { return n.Pos }
func (n *StringLit) Position() Pos    { return n.Pos }
func (n *BoolLit) Position() Pos      { return n.Pos }
func (n *NullLit) Position() Pos      { return n.Pos }
func (n *VoidLit) Position() Pos      { return n.Pos }
func (n *Template) Position() Pos     { return n.Pos }
func (n *RecordLit) Position() Pos    { return n.Pos }
func (n *MapLit) Position() Pos       { return n.Pos }
func (n *ArrayLit) Position() Pos     { return n.Pos }
func (n *Lambda) Position() Pos       { return n.Pos }
func (n *FuncLit) Position() Pos      { return n.Pos }
func (n *Field) Position() Pos        { return n.Pos }
func (n *Index) Position() Pos        { return n.Pos }
func (n *Call) Position() Pos         { return n.Pos }
func (n *Unary) Position() Pos        { return n.Pos }
func (n *Binary) Position() Pos       { return n.Pos }
func (n *Is) Position() Pos           { return n.Pos }
func (n *Assign) Position() Pos       { return n.Pos }
func (n *AssignArray) Position() Pos  { return n.Pos }
func (n *If) Position() Pos           { return n.Pos }
func (n *While) Position() Pos        { return n.Pos }
func (n *For) Position() Pos          { return n.Pos }
func (n *Go) Position() Pos           { return n.Pos }
func (n *Defer) Position() Pos        { return n.Pos }
func (n *Return) Position() Pos       { return n.Pos }
func (n *Break) Position() Pos        { return n.Pos }
func (n *Continue) Position() Pos     { return n.Pos }
func (n *Placeholder) Position() Pos  { return n.Pos }
func (n *Block) Position() Pos        { return n.Pos }

func (*VarDecl) declNode()  {}
func (*FuncDecl) declNode() {}
func (*TypeDecl) declNode() {}

func (n *VarDecl) Exported() bool  { return n.Export }
func (n *FuncDecl) Exported() bool { return n.Export }
func (n *TypeDecl) Exported() bool { return n.Export }

func (*VarDecl) stmtNode()     {}
func (*Ident) stmtNode()       {}
func (*IntLit) stmtNode()      {}
func (*FloatLit) stmtNode()    {}
func (*StringLit) stmtNode()   {}
func (*BoolLit) stmtNode()     {}
func (*NullLit) stmtNode()     {}
func (*VoidLit) stmtNode()     {}
func (*Template) stmtNode()    {}
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
func (*Defer) stmtNode()       {}
func (*Return) stmtNode()      {}
func (*Break) stmtNode()       {}
func (*Continue) stmtNode()    {}
func (*Placeholder) stmtNode() {}
func (*Block) stmtNode()       {}

func (*Ident) exprNode()       {}
func (*IntLit) exprNode()      {}
func (*FloatLit) exprNode()    {}
func (*StringLit) exprNode()   {}
func (*BoolLit) exprNode()     {}
func (*NullLit) exprNode()     {}
func (*VoidLit) exprNode()     {}
func (*Template) exprNode()    {}
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
func (*Defer) exprNode()       {}
func (*Return) exprNode()      {}
func (*Break) exprNode()       {}
func (*Continue) exprNode()    {}
func (*Placeholder) exprNode() {}
func (*Block) exprNode()       {}

func (*TypeName) typeNode()   {}
func (*RecordType) typeNode() {}
func (*FuncType) typeNode()   {}
func (*UnionType) typeNode()  {}

func (*NameTarget) targetNode()   {}
func (*RecordTarget) targetNode() {}
func (*ArrayTarget) targetNode()  {}
