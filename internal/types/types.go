// Package types is the checker's type language (section 4): primitives, null, void, never,
// arrays, maps, sets, functions, anonymous records, named types, type parameters, and unions.
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

// Set is Set[T].
type Set struct {
	Elem Type
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
	// Default marks a field of a RecordDecl that declares a default, which a literal may leave
	// out (section 4, Field defaults).
	Default bool
}

// Record is an anonymous record type { name: T, ... }.
type Record struct {
	Fields []Field
}

// DeclKind says what a type declaration declared.
type DeclKind int

const (
	RecordDecl       DeclKind = iota // a named record type
	PrimitiveDecl                    // a named primitive type
	OpaqueDecl                       // a prelude type with no structure, such as Mutex
	KindOpaqueDecl                   // an opaque prelude type whose values carry a kind of their own, Bytes and Html: Base holds the kind, and is tests it by name
	TaggedOpaqueDecl                 // an opaque type whose values carry its tag, as the time types of stdlib.md and Regex.Pattern do: is tests it by the tag
	AliasDecl                        // an alias, interchangeable with its body
)

// Decl is a type declaration.
type Decl struct {
	Name   string
	Kind   DeclKind
	Params []*ParamDecl
	Fields []Field    // RecordDecl: the fields, with Params inside
	Base   value.Kind // PrimitiveDecl: the base kind; KindOpaqueDecl: the kind
	Body   Type       // AliasDecl: the body, with Params inside
	Funcs  map[string]*Func
	Pos    any    // the declaration's position, for messages
	Path   string // the file the declaration is in, for messages
	// Tag is the tag the runtime binds the type to, set by the interpreter on the declarations
	// it builds for reflection (section 4, Types as values); nil in the checker and for a
	// KindOpaqueDecl, whose values carry no tag.
	Tag *value.Tag
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
func (Set) isType()     {}
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

// IsInvalid reports whether t is the error-recovery type, or a type parameter whose bound the
// checker rejected and made invalid, as that of T in func f[T: T] is: the error was reported once,
// at the bound, and what the body does with a T would only repeat it.
func IsInvalid(t Type) bool {
	switch x := t.(type) {
	case Invalid:
		return true
	case Param:
		_, ok := x.Decl.Bound.(Invalid)
		return ok
	}
	return false
}

// IsNever reports whether t is never.
func IsNever(t Type) bool {
	_, ok := t.(Never)
	return ok
}

// Expand replaces an alias by its body, repeatedly, and leaves every other type alone. The
// substitution cannot fail: the checker checks an alias's arguments against its body where they
// are written, and a generic alias never reaches itself (section 4, Type declarations), so no
// deeper instantiation is left unchecked. Nor can the chain loop, since the checker marks a
// cyclic alias invalid as soon as its cycle is resolved; past 64 steps the instantiations met
// are tracked to assert it.
func Expand(t Type) Type {
	t, msg := expand(t)
	if msg != "" {
		panic("internal error: " + msg)
	}
	return t
}

// ExpandOK is Expand for a type the checks of its declarations have not all reached yet, such as
// a bound read while the declarations are resolved: it reports false where Expand would panic,
// on an alias cycle or arguments its body rejects, and expands an alias whose body is not yet
// resolved to itself.
func ExpandOK(t Type) (Type, bool) {
	t, msg := expand(t)
	return t, msg == ""
}

// expand is Expand, reporting what it would panic with.
func expand(t Type) (Type, string) {
	var met cycleCheck
	for i := 0; ; i++ {
		n, ok := t.(Named)
		if !ok || n.Decl.Kind != AliasDecl || n.Decl.Body == nil {
			return t, ""
		}
		if i >= 64 && met.repeats(n) {
			return t, "expanding the cyclic alias " + n.Decl.Name + ", which the checker accepted"
		}
		body, err := substUnionBody(n, Bindings(n.Decl.Params, n.Args))
		if err != nil {
			return t, "expanding " + String(n) + ", which the checker accepted: " + err.Error()
		}
		t = body
	}
}

// cycleCheck finds a repeated instantiation in a chain of alias expansions by Brent's method: it
// compares each step with one saved step, saved again at every power of two, so a cycle is
// found once the saved step is on it and the gap has grown past its length. A chain such as
// Id[Id[...Id[int]...]] meets one alias many times, each time with other arguments, and is
// compared in time linear in its length, not quadratic.
type cycleCheck struct {
	saved      Named
	have       bool
	gap, power int
}

func (c *cycleCheck) repeats(n Named) bool {
	if c.have && identical(c.saved, n, 0) {
		return true
	}
	c.gap++
	if !c.have || c.gap >= c.power {
		c.saved, c.have, c.gap = n, true, 0
		c.power = max(2*c.power, 1)
	}
	return false
}

// Bindings pairs type parameters with their arguments; a parameter past the arguments is left
// out.
func Bindings(params []*ParamDecl, args []Type) map[*ParamDecl]Type {
	m := make(map[*ParamDecl]Type, len(params))
	for i, p := range params {
		if i < len(args) {
			m[p] = args[i]
		}
	}
	return m
}

// FieldsOf returns the fields of a named record type with its arguments substituted. The
// substitution cannot fail: the checker checks every instantiation a written type reaches,
// through fields and alias bodies, where its arguments are written (section 4, Unions), and a
// type value is built only from closed type expressions and builders that keep this (section 4,
// Types as values).
func (n Named) FieldsOf() []Field {
	m := Bindings(n.Decl.Params, n.Args)
	out := make([]Field, len(n.Decl.Fields))
	for i, f := range n.Decl.Fields {
		t, err := Subst(f.Type, m)
		if err != nil {
			panic("internal error: the fields of " + String(n) + ", which the checker accepted: " + err.Error())
		}
		out[i] = Field{Name: f.Name, Type: t, Default: f.Default}
	}
	return out
}

// ---- equality

// Equal reports whether two types are the same type. A recursive alias is compared as the
// infinite type it stands for: a pair of types already under comparison is assumed equal, as
// == does for values, so that type L[T] = Array[L[T] | int] ends.
func Equal(a, b Type) bool { return equal(a, b, &pairs{}, 0) }

// equal compares two types depth levels into a comparison. It moves to a fresh stack every
// value.Deep levels, as == on values does, so that a type value nested a million deep, as a
// host or a long type expression can make, compares without overflowing the Go stack (section 9:
// the interpreter never crashes).
func equal(a, b Type, met *pairs, depth int) bool {
	if value.Deep(depth) {
		return equalOnFreshStack(a, b, met, depth)
	}
	return equalStep(a, b, met, depth)
}

func equalOnFreshStack(a, b Type, met *pairs, depth int) (eq bool) {
	value.OnFreshStack(func() { eq = equalStep(a, b, met, depth) })
	return eq
}

// equalStep is equal on the current stack.
func equalStep(a, b Type, met *pairs, depth int) (eq bool) {
	// An alias with its arguments written alike is the same type. With other arguments it may
	// still be, when its body ignores a parameter or the arguments are equal types written
	// differently, so it is expanded below. Comparing the arguments with equal here as well would
	// compare those of an alias nested in its own arguments, as in Opt[Opt[Opt[int]]], once here
	// and once inside the expansion, at every level: exponential in the nesting.
	if na, ok := a.(Named); ok && isAlias(a) {
		if nb, ok := b.(Named); ok && na.Decl == nb.Decl && identicalAll(na.Args, nb.Args, 0) {
			return true
		}
	}
	if met.remembered(a) || met.remembered(b) {
		p, mark, holds, known := met.enter(a, b)
		if known {
			return holds
		}
		defer func() { met.leave(p, mark, eq) }()
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
		return ok && equal(x.Elem, y.Elem, met, depth+1)
	case Map:
		y, ok := b.(Map)
		return ok && equal(x.Key, y.Key, met, depth+1) && equal(x.Val, y.Val, met, depth+1)
	case Set:
		y, ok := b.(Set)
		return ok && equal(x.Elem, y.Elem, met, depth+1)
	case Func:
		y, ok := b.(Func)
		if !ok || len(x.Params) != len(y.Params) || x.Required != y.Required {
			return false
		}
		for i := range x.Params {
			if !equal(x.Params[i], y.Params[i], met, depth+1) {
				return false
			}
		}
		return equal(x.Return, y.Return, met, depth+1)
	case Record:
		y, ok := b.(Record)
		if !ok || len(x.Fields) != len(y.Fields) {
			return false
		}
		// Field order is not part of an anonymous record's shape.
		for _, f := range x.Fields {
			found := false
			for _, g := range y.Fields {
				if g.Name == f.Name {
					found = equal(f.Type, g.Type, met, depth+1)
					break
				}
			}
			if !found {
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
			if !equal(x.Args[i], y.Args[i], met, depth+1) {
				return false
			}
		}
		return true
	case Param:
		y, ok := b.(Param)
		return ok && x.Decl == y.Decl
	case Union:
		// As sets, since members are not ordered.
		y, ok := b.(Union)
		return ok && within(x, y, met, depth) && within(y, x, met, depth)
	}
	return false
}

// within reports whether every member of x equals a member of y.
func within(x, y Union, met *pairs, depth int) bool {
	for _, m := range x.Members {
		found := false
		for _, n := range y.Members {
			if equal(m, n, met, depth+1) {
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

func isAlias(t Type) bool {
	n, ok := t.(Named)
	return ok && n.Decl.Kind == AliasDecl && n.Decl.Body != nil
}

// identical reports whether two types are written alike, expanding no alias. Like equal, it
// moves to a fresh stack every value.Deep levels.
func identical(a, b Type, depth int) bool {
	if value.Deep(depth) {
		return identicalOnFreshStack(a, b, depth)
	}
	return identicalStep(a, b, depth)
}

func identicalOnFreshStack(a, b Type, depth int) (same bool) {
	value.OnFreshStack(func() { same = identicalStep(a, b, depth) })
	return same
}

// identicalStep is identical on the current stack.
func identicalStep(a, b Type, depth int) bool {
	switch x := a.(type) {
	case Named:
		y, ok := b.(Named)
		return ok && x.Decl == y.Decl && identicalAll(x.Args, y.Args, depth+1)
	case Param:
		y, ok := b.(Param)
		return ok && x.Decl == y.Decl
	case Prim:
		y, ok := b.(Prim)
		return ok && x.Kind == y.Kind
	case Array:
		y, ok := b.(Array)
		return ok && identical(x.Elem, y.Elem, depth+1)
	case Map:
		y, ok := b.(Map)
		return ok && identical(x.Key, y.Key, depth+1) && identical(x.Val, y.Val, depth+1)
	case Set:
		y, ok := b.(Set)
		return ok && identical(x.Elem, y.Elem, depth+1)
	case Func:
		y, ok := b.(Func)
		return ok && x.Required == y.Required && identicalAll(x.Params, y.Params, depth+1) && identical(x.Return, y.Return, depth+1)
	case Record:
		y, ok := b.(Record)
		if !ok || len(x.Fields) != len(y.Fields) {
			return false
		}
		for i, f := range x.Fields {
			if f.Name != y.Fields[i].Name || !identical(f.Type, y.Fields[i].Type, depth+1) {
				return false
			}
		}
		return true
	case Union:
		y, ok := b.(Union)
		return ok && identicalAll(x.Members, y.Members, depth+1)
	case Null, Void, Never, Invalid:
		return a == b
	}
	return false
}

func identicalAll(as, bs []Type, depth int) bool {
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		if !identical(as[i], bs[i], depth) {
			return false
		}
	}
	return true
}

// ---- rendering

// String renders a type as Ant writes it.
func String(t Type) string {
	w := &writer{}
	w.write(t, 0)
	return w.sb.String()
}

// Strings renders several types for one message. When a name stands for two different
// declarations among them, as two types of one name declared in different blocks or modules
// do, each use of it is followed by where(decl).
func Strings(where func(*Decl) string, ts ...Type) []string {
	seen := map[string]*Decl{}
	clash := map[string]bool{}
	probe := &writer{named: func(d *Decl) {
		if old, ok := seen[d.Name]; ok && old != d {
			clash[d.Name] = true
		}
		seen[d.Name] = d
	}}
	for _, t := range ts {
		probe.write(t, 0)
	}
	out := make([]string, len(ts))
	for i, t := range ts {
		w := &writer{}
		if len(clash) > 0 {
			w.suffix = func(d *Decl) string {
				if clash[d.Name] {
					return where(d)
				}
				return ""
			}
		}
		w.write(t, 0)
		out[i] = w.sb.String()
	}
	return out
}

// writer renders types. named, when set, sees every declared type rendered; suffix, when set,
// gives the text that follows its name.
type writer struct {
	sb     strings.Builder
	named  func(*Decl)
	suffix func(*Decl) string
}

func (w *writer) write(t Type, depth int) {
	sb := &w.sb
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
		w.write(x.Elem, depth+1)
		sb.WriteString("]")
	case Map:
		sb.WriteString("Map[")
		w.write(x.Key, depth+1)
		sb.WriteString(", ")
		w.write(x.Val, depth+1)
		sb.WriteString("]")
	case Set:
		sb.WriteString("Set[")
		w.write(x.Elem, depth+1)
		sb.WriteString("]")
	case Func:
		sb.WriteString("func(")
		for i, p := range x.Params {
			if i > 0 {
				sb.WriteString(", ")
			}
			w.write(p, depth+1)
			if i >= x.Required {
				sb.WriteString(" = ...")
			}
		}
		sb.WriteString("): ")
		w.write(x.Return, depth+1)
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
			w.write(f.Type, depth+1)
		}
		sb.WriteString(" }")
	case Named:
		sb.WriteString(x.Decl.Name)
		if w.named != nil {
			w.named(x.Decl)
		}
		if len(x.Args) > 0 {
			sb.WriteString("[")
			for i, a := range x.Args {
				if i > 0 {
					sb.WriteString(", ")
				}
				w.write(a, depth+1)
			}
			sb.WriteString("]")
		}
		if w.suffix != nil {
			sb.WriteString(w.suffix(x.Decl))
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
				w.write(f, depth+1)
				sb.WriteString(")")
				continue
			}
			w.write(m, depth+1)
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

// MemberOf returns the first member of t, or t itself when it is no union, that expands to a T.
func MemberOf[T Type](t Type) (T, bool) {
	for _, m := range Members(t) {
		if x, ok := Expand(m).(T); ok {
			return x, true
		}
	}
	var zero T
	return zero, false
}

// Nullable is t | null.
func Nullable(t Type) Type {
	u, _ := MakeUnion([]Type{t, Null{}})
	return u
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
