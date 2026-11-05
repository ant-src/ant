package types

import (
	"runtime/debug"
	"testing"
	"time"
)

// TestDeepTypes relates types nested far deeper than one small Go stack holds, as a walk of a
// type such as type R[T] = { next: R[Array[T]] | null } builds at run time, so the walks must
// move to fresh goroutines as they go down (section 9: no crash), and with no depth limit
// (section 4, Assignability).
func TestDeepTypes(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(64 << 20))
	n := 300_000
	nest := func(leaf Type, wrap func(Type) Type) Type {
		for i := 0; i < n; i++ {
			leaf = wrap(leaf)
		}
		return leaf
	}
	array := func(e Type) Type { return Array{Elem: e} }
	thunk := func(r Type) Type { return Func{Return: r} }
	if !Equal(nest(IntType, array), nest(IntType, array)) {
		t.Error("two equal deep array types are unequal")
	}
	if Equal(nest(IntType, array), nest(StringType, array)) {
		t.Error("deep array types with different leaves are equal")
	}
	wide := Nullable(IntType)
	if !Assignable(nest(IntType, thunk), nest(wide, thunk)) {
		t.Error("a deep function type is not assignable to one with a wider innermost result")
	}
	if Assignable(nest(wide, thunk), nest(IntType, thunk)) {
		t.Error("a deep function type is assignable to one with a narrower innermost result")
	}
	// An alias at every level, as Opt[Opt[...]] is, goes through identical and the assumed pairs.
	p := &ParamDecl{Name: "T"}
	box := &Decl{Name: "Box", Kind: AliasDecl, Params: []*ParamDecl{p}, Body: Array{Elem: Param{Decl: p}}}
	boxed := func(e Type) Type { return Named{Decl: box, Args: []Type{e}} }
	if !Equal(nest(IntType, boxed), nest(IntType, array)) {
		t.Error("a deep alias type is unequal to its expansion")
	}
}

// TestNestedAliasArgs compares and forms unions of an alias nested in its own arguments, which
// once took time exponential in the nesting, and keeps both instantiations of one alias in a
// union.
func TestNestedAliasArgs(t *testing.T) {
	p := &ParamDecl{Name: "T"}
	body, _ := MakeUnion([]Type{Param{Decl: p}, Null{}})
	opt := &Decl{Name: "Opt", Kind: AliasDecl, Params: []*ParamDecl{p}, Body: body}
	nest := func(n int, leaf Type) Type {
		for i := 0; i < n; i++ {
			leaf = Named{Decl: opt, Args: []Type{leaf}}
		}
		return leaf
	}
	start := time.Now()
	if Equal(nest(200, IntType), nest(200, StringType)) {
		t.Error("Opt^200[int] equals Opt^200[string]")
	}
	if !Equal(nest(200, IntType), Nullable(IntType)) {
		t.Error("Opt^200[int] is not int | null")
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("comparing nested aliases took %v", d)
	}
	u, err := MakeUnion([]Type{nest(1, IntType), nest(1, StringType)})
	if err != nil || len(Members(u)) != 3 {
		t.Errorf("Opt[int] | Opt[string] is %s (%v), want three members", String(u), err)
	}
}

// TestRecursiveAliasAssignable relates two recursive aliases coinductively: type F =
// func(): F | null is assignable to type G = func(): G | int | null, and not the other way.
func TestRecursiveAliasAssignable(t *testing.T) {
	f := &Decl{Name: "F", Kind: AliasDecl}
	f.Body = Func{Return: Nullable(Named{Decl: f})}
	g := &Decl{Name: "G", Kind: AliasDecl}
	gr, _ := MakeUnion([]Type{Named{Decl: g}, IntType, Null{}})
	g.Body = Func{Return: gr}
	if !Assignable(Named{Decl: f}, Named{Decl: g}) {
		t.Error("F is not assignable to G")
	}
	if Assignable(Named{Decl: g}, Named{Decl: f}) {
		t.Error("G is assignable to F")
	}
}
