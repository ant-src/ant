// Package types is the checker's type language (section 4): primitives, null, void, never,
// arrays, maps, functions, anonymous records, named types, type parameters, and unions.
package types

import (
	"fmt"
	"strings"

	"github.com/ant-src/ant/value"
)

// Type is a static type. Named types are compared by declaration identity and arguments;
// everything else is structural.
type Type interface {
	isType()
}

// Prim is one of the ten numeric kinds, bool, or string.
type Prim struct {
	Kind value.Kind
}

// Null, Void, and Never are the types of null, void, and expressions that do not return.
type (
	Null  struct{}
	Void  struct{}
	Never struct{}
	// Invalid stands for a type that could not be determined; it is assignable both ways so
	// that one error does not cascade.
	Invalid struct{}
)

// Array is Array[T].
type Array struct {
	Elem Type
}

// Map is Map[K, V].
type Map struct {
	Key, Val Type
}

// Func is a function type. Required is the number of parameters without defaults; a
// declared function may be called with fewer arguments than Params down to Required.
type Func struct {
	Params   []Type
	Return   Type
	Required int
	Names    []string // parameter names, for named arguments; nil for a function type
}

// Field is a field of a record type.
type Field struct {
	Name string
	Type Type
}

// Record is an anonymous record type { name: T, ... }.
type Record struct {
	Fields []Field
}

// DeclKind says what a type declaration declared.
type DeclKind int

const (
	RecordDecl    DeclKind = iota // a named record type
	PrimitiveDecl                 // a named primitive type
	OpaqueDecl                    // a prelude type with no structure, such as Mutex
	TimeDecl                      // a time type of stdlib.md: opaque, but is tests it by its tag
	AliasDecl                     // an alias, interchangeable with its body
)

// Decl is a type declaration.
type Decl struct {
	Name   string
	Kind   DeclKind
	Params []*ParamDecl
	Fields []Field    // RecordDecl: the fields, with Params inside
	Base   value.Kind // PrimitiveDecl: the base kind
	Body   Type       // AliasDecl: the body, with Params inside
	Funcs  map[string]*Func
	Pos    any // the declaration's position, for messages
	// resolving marks an alias whose body is being resolved, to catch cycles.
	Resolving bool
}

// Named is a reference to a declared type with type arguments.
type Named struct {
	Decl *Decl
	Args []Type
}

// ParamDecl is a type parameter declaration.
type ParamDecl struct {
	Name  string
	Bound Type // nil when unbounded
}

// Param is a type parameter in use.
type Param struct {
	Decl *ParamDecl
}

// Union is A | B. It is always normalized: flat, without duplicates or never, and with at
// least two members. Written keeps the members as they were written when one of them was an
// alias that normalization expanded, so that messages can name the alias; it is nil otherwise
// and takes no part in equality.
type Union struct {
	Members []Type
	Written []Type
}

func (Prim) isType()    {}
func (Null) isType()    {}
func (Void) isType()    {}
func (Never) isType()   {}
func (Invalid) isType() {}
func (Array) isType()   {}
func (Map) isType()     {}
func (Func) isType()    {}
func (Record) isType()  {}
func (Named) isType()   {}
func (Param) isType()   {}
func (Union) isType()   {}

// Common types.
var (
	IntType     Type = Prim{value.KindInt64}
	Float64Type Type = Prim{value.KindFloat64}
	BoolType    Type = Prim{value.KindBool}
	StringType  Type = Prim{value.KindString}
	ByteType    Type = Prim{value.KindUint8}
)

// IsInvalid reports whether t is the error-recovery type.
func IsInvalid(t Type) bool {
	_, ok := t.(Invalid)
	return ok
}

// IsNever reports whether t is never.
func IsNever(t Type) bool {
	_, ok := t.(Never)
	return ok
}

// Expand replaces an alias by its body, repeatedly, and leaves every other type alone.
func Expand(t Type) Type {
	for i := 0; i < 64; i++ {
		n, ok := t.(Named)
		if !ok || n.Decl.Kind != AliasDecl || n.Decl.Body == nil {
			return t
		}
		body, err := Subst(n.Decl.Body, bindings(n.Decl.Params, n.Args))
		if err != nil {
			return Invalid{}
		}
		t = body
	}
	return Invalid{}
}

func bindings(params []*ParamDecl, args []Type) map[*ParamDecl]Type {
	m := make(map[*ParamDecl]Type, len(params))
	for i, p := range params {
		if i < len(args) {
			m[p] = args[i]
		}
	}
	return m
}

// Fields returns the fields of a named record type with its arguments substituted.
func (n Named) FieldsOf() []Field {
	m := bindings(n.Decl.Params, n.Args)
	out := make([]Field, len(n.Decl.Fields))
	for i, f := range n.Decl.Fields {
		t, err := Subst(f.Type, m)
		if err != nil {
			t = Invalid{}
		}
		out[i] = Field{Name: f.Name, Type: t}
	}
	return out
}

// Field returns a named record type's field.
func (n Named) Field(name string) (Type, bool) {
	for _, f := range n.FieldsOf() {
		if f.Name == name {
			return f.Type, true
		}
	}
	return nil, false
}

// ---- equality

// Equal reports whether two types are the same type.
func Equal(a, b Type) bool { return equal(a, b, 0) }

func equal(a, b Type, depth int) bool {
	if depth > 64 {
		return false
	}
	if na, ok := a.(Named); ok {
		if nb, ok := b.(Named); ok && na.Decl == nb.Decl {
			if len(na.Args) != len(nb.Args) {
				return false
			}
			for i := range na.Args {
				if !equal(na.Args[i], nb.Args[i], depth+1) {
					return false
				}
			}
			return true
		}
	}
	a, b = Expand(a), Expand(b)
	switch x := a.(type) {
	case Prim:
		y, ok := b.(Prim)
		return ok && x.Kind == y.Kind
	case Null:
		_, ok := b.(Null)
		return ok
	case Void:
		_, ok := b.(Void)
		return ok
	case Never:
		_, ok := b.(Never)
		return ok
	case Invalid:
		_, ok := b.(Invalid)
		return ok
	case Array:
		y, ok := b.(Array)
		return ok && equal(x.Elem, y.Elem, depth+1)
	case Map:
		y, ok := b.(Map)
		return ok && equal(x.Key, y.Key, depth+1) && equal(x.Val, y.Val, depth+1)
	case Func:
		y, ok := b.(Func)
		if !ok || len(x.Params) != len(y.Params) || x.Required != y.Required {
			return false
		}
		for i := range x.Params {
			if !equal(x.Params[i], y.Params[i], depth+1) {
				return false
			}
		}
		return equal(x.Return, y.Return, depth+1)
	case Record:
		y, ok := b.(Record)
		if !ok || len(x.Fields) != len(y.Fields) {
			return false
		}
		for i := range x.Fields {
			if x.Fields[i].Name != y.Fields[i].Name || !equal(x.Fields[i].Type, y.Fields[i].Type, depth+1) {
				return false
			}
		}
		return true
	case Named:
		y, ok := b.(Named)
		if !ok || x.Decl != y.Decl || len(x.Args) != len(y.Args) {
			return false
		}
		for i := range x.Args {
			if !equal(x.Args[i], y.Args[i], depth+1) {
				return false
			}
		}
		return true
	case Param:
		y, ok := b.(Param)
		return ok && x.Decl == y.Decl
	case Union:
		y, ok := b.(Union)
		if !ok || len(x.Members) != len(y.Members) {
			return false
		}
		for _, m := range x.Members {
			found := false
			for _, n := range y.Members {
				if equal(m, n, depth+1) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return false
}

// ---- rendering

// String renders a type as Ant writes it.
func String(t Type) string {
	var sb strings.Builder
	write(&sb, t, 0)
	return sb.String()
}

func write(sb *strings.Builder, t Type, depth int) {
	if depth > 8 {
		sb.WriteString("...")
		return
	}
	switch x := t.(type) {
	case Prim:
		sb.WriteString(x.Kind.String())
	case Null:
		sb.WriteString("null")
	case Void:
		sb.WriteString("void")
	case Never:
		sb.WriteString("never")
	case Invalid:
		sb.WriteString("?")
	case Array:
		sb.WriteString("Array[")
		write(sb, x.Elem, depth+1)
		sb.WriteString("]")
	case Map:
		sb.WriteString("Map[")
		write(sb, x.Key, depth+1)
		sb.WriteString(", ")
		write(sb, x.Val, depth+1)
		sb.WriteString("]")
	case Func:
		sb.WriteString("func(")
		for i, p := range x.Params {
			if i > 0 {
				sb.WriteString(", ")
			}
			write(sb, p, depth+1)
			if i >= x.Required {
				sb.WriteString(" = ...")
			}
		}
		sb.WriteString("): ")
		write(sb, x.Return, depth+1)
	case Record:
		if len(x.Fields) == 0 {
			sb.WriteString("{}")
			return
		}
		sb.WriteString("{ ")
		for i, f := range x.Fields {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(f.Name)
			sb.WriteString(": ")
			write(sb, f.Type, depth+1)
		}
		sb.WriteString(" }")
	case Named:
		sb.WriteString(x.Decl.Name)
		if len(x.Args) > 0 {
			sb.WriteString("[")
			for i, a := range x.Args {
				if i > 0 {
					sb.WriteString(", ")
				}
				write(sb, a, depth+1)
			}
			sb.WriteString("]")
		}
	case Param:
		sb.WriteString(x.Decl.Name)
	case Union:
		members := x.Members
		if x.Written != nil {
			members = x.Written
		}
		for i, m := range members {
			if i > 0 {
				sb.WriteString(" | ")
			}
			if f, ok := m.(Func); ok {
				sb.WriteString("(")
				write(sb, f, depth+1)
				sb.WriteString(")")
				continue
			}
			write(sb, m, depth+1)
		}
	default:
		fmt.Fprintf(sb, "%T", t)
	}
}

// Members returns a union's members, or the type itself as one member.
func Members(t Type) []Type {
	if u, ok := Expand(t).(Union); ok {
		return u.Members
	}
	return []Type{t}
}

// Includes reports whether t is a union that has a member equal to m, or is m itself.
func Includes(t, m Type) bool {
	for _, x := range Members(t) {
		if Equal(x, m) {
			return true
		}
	}
	return false
}

// Without returns t with the members equal to m removed; never when nothing remains.
func Without(t Type, m Type) Type {
	var keep []Type
	for _, x := range Members(t) {
		if !Equal(x, m) {
			keep = append(keep, x)
		}
	}
	u, _ := MakeUnion(keep)
	return u
}
