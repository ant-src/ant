package types

import (
	"fmt"
	"runtime/debug"
	"testing"
	"time"
)

// TestDeepTypes relates types nested far deeper than one small Go stack holds, so the walks must
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

// aliasChain declares n aliases, each built by wrap from the one before, from leaf.
func aliasChain(name string, n int, leaf Type, wrap func(prev Type) Type) []*Decl {
	ds := make([]*Decl, n+1)
	for i := range ds {
		ds[i] = &Decl{Name: fmt.Sprint(name, i), Kind: AliasDecl, Body: leaf}
		if i > 0 {
			ds[i].Body = wrap(Named{Decl: ds[i-1]})
		}
	}
	return ds
}

// TestSharedAliases relates types that name each alias many times, which once took time
// exponential or cubic in the number of aliases: a pair met again is not walked again.
func TestSharedAliases(t *testing.T) {
	start := time.Now()
	// F<i> = func(F<i-1>): F<i-1>, and a ring G<i> = func(G<i+1>): G<i+1> | null.
	fn := func(p Type) Type { return Func{Params: []Type{p}, Return: p, Required: 1} }
	f := aliasChain("F", 60, IntType, fn)
	top := Named{Decl: f[60]}
	if !Assignable(top, Nullable(top)) {
		t.Error("F60 is not assignable to F60 | null")
	}
	if !Assignable(top, Union{Members: []Type{IntType, f[60].Body}}) {
		t.Error("F60 is not assignable to (func(F59): F59) | int")
	}
	g := aliasChain("G", 60, nil, func(p Type) Type { return Func{Params: []Type{p}, Return: Nullable(p), Required: 1} })
	g[0].Body = g[60].Body
	if !Assignable(Named{Decl: g[0]}, Nullable(Named{Decl: g[0]})) {
		t.Error("G0 is not assignable to G0 | null")
	}
	// D<i> = Map[D<i-1>, D<i-1>] and its twin E<i>, from int | null and null | int.
	m := func(p Type) Type { return Map{Key: p, Val: p} }
	d := aliasChain("D", 60, Nullable(IntType), m)
	e := aliasChain("E", 60, Union{Members: []Type{Null{}, IntType}}, m)
	if !Equal(Named{Decl: d[60]}, Named{Decl: e[60]}) || !Assignable(Named{Decl: d[60]}, Named{Decl: e[60]}) {
		t.Error("D60 is not E60")
	}
	// Two chains of 20000 aliases D<i> = Array[D<i-1>].
	arr := func(p Type) Type { return Array{Elem: p} }
	a := aliasChain("A", 20000, Nullable(IntType), arr)
	b := aliasChain("B", 20000, Nullable(IntType), arr)
	if !Equal(Named{Decl: a[20000]}, Named{Decl: b[20000]}) {
		t.Error("A20000 is not B20000")
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("relating shared aliases took %v", d)
	}
}

// TestChainComponents checks each link of long chains of declarations the way the checker and
// the runtime do, which once followed the whole chain from every link, in time quadratic in its
// length: the searches stop at the declaration's strongly connected component, and validating
// the instantiations the links write remembers what an earlier link validated.
func TestChainComponents(t *testing.T) {
	start := time.Now()
	const n = 20000
	// D<i>[T] = Array[D<i-1>[T]] from D0[T] = T | null, and a ring R<i> = R<i+1> | null closed
	// at R0, which is a cycle through unions alone.
	d := make([]*Decl, n+1)
	for i := range d {
		p := &ParamDecl{Name: "T"}
		d[i] = &Decl{Name: fmt.Sprint("D", i), Kind: AliasDecl, Params: []*ParamDecl{p}}
		if i == 0 {
			d[i].Body = Nullable(Param{Decl: p})
		} else {
			d[i].Body = Array{Elem: Named{Decl: d[i-1], Args: []Type{Param{Decl: p}}}}
		}
	}
	settled := map[*Decl]map[*Decl]bool{}
	component := func(root *Decl) func(*Decl) bool {
		Components(root, DeclRefs, func(x *Decl) bool { return settled[x] != nil }, func(comp []*Decl, _ bool) {
			set := map[*Decl]bool{}
			for _, x := range comp {
				set[x] = true
			}
			for _, x := range comp {
				settled[x] = set
			}
		})
		return func(x *Decl) bool { return settled[root][x] }
	}
	var memo ArgsMemo
	for _, x := range d {
		in := component(x)
		if msg := SelfArgsIn(x, in); msg != "" {
			t.Fatal(msg)
		}
		if _, cyclic := AliasCycleIn(x, in); cyclic {
			t.Fatalf("%s is a cycle", x.Name)
		}
		if msg := UnionSelfIn(x, in); msg != "" {
			t.Fatal(msg)
		}
		if err := memo.CheckWritten(x.Body); err != nil {
			t.Fatalf("%s: %v", x.Name, err)
		}
	}
	r := make([]*Decl, n)
	for i := range r {
		r[i] = &Decl{Name: fmt.Sprint("R", i), Kind: AliasDecl}
	}
	for i := range r {
		r[i].Body = Union{Written: []Type{Named{Decl: r[(i+1)%n]}, Null{}}, Members: []Type{Null{}}}
	}
	settled = map[*Decl]map[*Decl]bool{}
	in := component(r[0])
	if len(settled[r[0]]) != n {
		t.Fatalf("the ring's component has %d declarations, want %d", len(settled[r[0]]), n)
	}
	if path, cyclic := AliasCycleIn(r[0], in); !cyclic || len(path) != n-1 {
		t.Errorf("R0 reaches itself through %d aliases (%v), want %d", len(path), cyclic, n-1)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("checking the chains took %v", d)
	}
}

// TestFailedAssumption checks that a pair proven only by assuming a pair that then failed is
// not taken as proven. With A = func(D, int): void, B = func(C, string): void, C = func(): A,
// and D = func(): B, relating A to B relates C to D by assuming A to B, which then fails on
// int and string; C is then related to D again, and fails too.
func TestFailedAssumption(t *testing.T) {
	a := &Decl{Name: "A", Kind: AliasDecl}
	b := &Decl{Name: "B", Kind: AliasDecl}
	c := &Decl{Name: "C", Kind: AliasDecl, Body: Func{Return: Named{Decl: a}}}
	d := &Decl{Name: "D", Kind: AliasDecl, Body: Func{Return: Named{Decl: b}}}
	a.Body = Func{Params: []Type{Named{Decl: d}, IntType}, Return: Void{}, Required: 2}
	b.Body = Func{Params: []Type{Named{Decl: c}, StringType}, Return: Void{}, Required: 2}
	thunk := func(d *Decl) Type { return Func{Return: Named{Decl: d}} }
	if Assignable(Union{Members: []Type{thunk(a), thunk(c)}}, Union{Members: []Type{thunk(b), thunk(d), thunk(a)}}) {
		t.Error("func(): C is assignable to func(): D")
	}
	// The same for Equal, with E = Map[G, int], F = Map[H, string], G = Array[E], H = Array[F].
	e := &Decl{Name: "E", Kind: AliasDecl}
	f := &Decl{Name: "F", Kind: AliasDecl}
	g := &Decl{Name: "G", Kind: AliasDecl, Body: Array{Elem: Named{Decl: e}}}
	h := &Decl{Name: "H", Kind: AliasDecl, Body: Array{Elem: Named{Decl: f}}}
	e.Body = Map{Key: Named{Decl: g}, Val: IntType}
	f.Body = Map{Key: Named{Decl: h}, Val: StringType}
	E, F, G, H := Named{Decl: e}, Named{Decl: f}, Named{Decl: g}, Named{Decl: h}
	if Equal(Union{Members: []Type{E, G, F}}, Union{Members: []Type{F, E, H}}) {
		t.Error("G equals H")
	}
}

// TestAliasInItsOwnUnion relates and checks an alias that is a member of a union formed in the
// body it expands to, as A is with A = B[B[B[int]]] and B[T] = func(null): A | null | T. Forming
// that union compares A with B[B[int]], which expands A, which forms the union again; Equal once
// did this without end, each time with none of the pairs it had assumed. A = B[B[B[int]]] is a
// union of two functions (section 4, Unions); with A = B[int], A and B[int] are one type, and
// B[B[int]] forms A | null | B[int] with one function.
func TestAliasInItsOwnUnion(t *testing.T) {
	p := &ParamDecl{Name: "T"}
	a := &Decl{Name: "A", Kind: AliasDecl}
	b := &Decl{Name: "B", Kind: AliasDecl, Params: []*ParamDecl{p}}
	b.Body = Func{Params: []Type{Null{}}, Required: 1, Return: Union{Members: []Type{Named{Decl: a}, Null{}, Param{Decl: p}}}}
	B := func(n int) Type {
		var t Type = IntType
		for i := 0; i < n; i++ {
			t = Named{Decl: b, Args: []Type{t}}
		}
		return t
	}
	A := Named{Decl: a}
	a.Body = B(3)
	if Equal(A, B(2)) || Equal(B(2), A) {
		t.Error("A = B[B[B[int]]] equals B[B[int]]")
	}
	if !Equal(A, B(3)) {
		t.Error("A is unequal to its body")
	}
	if err := CheckArgs(B(3).(Named)); err == nil {
		t.Error("B[B[B[int]]] forms no invalid union")
	}
	if _, err := MakeUnion([]Type{A, B(2)}); err == nil {
		t.Error("A | B[B[int]] is a valid union")
	}
	a.Body = B(1)
	if !Equal(A, B(1)) || Equal(B(2), B(1)) {
		t.Error("with A = B[int], A is not B[int], or B[B[int]] is")
	}
	if err := CheckArgs(B(2).(Named)); err != nil {
		t.Errorf("with A = B[int], B[B[int]] forms an invalid union: %v", err)
	}
	if u, err := MakeUnion([]Type{A, B(1)}); err != nil || len(Members(u)) != 1 {
		t.Errorf("with A = B[int], A | B[int] is %s (%v), want one member", String(u), err)
	}
}

// TestUnionSelfThroughAlias forms the unions of B[T] = (func(): A | int) | null with A = B[int],
// which names B in a union inside B's union body through A. The checker reports it
// (UnionSelf), but only once both bodies are known; until then expanding A inside the union
// being formed for B[int] formed it again, without end. Now the expansion stops at B[int].
func TestUnionSelfThroughAlias(t *testing.T) {
	p := &ParamDecl{Name: "T"}
	a := &Decl{Name: "A", Kind: AliasDecl}
	b := &Decl{Name: "B", Kind: AliasDecl, Params: []*ParamDecl{p}}
	b.Body = Union{Members: []Type{Func{Return: Union{Members: []Type{Named{Decl: a}, IntType}}}, Null{}}}
	a.Body = Named{Decl: b, Args: []Type{IntType}}
	if UnionSelf(b) == "" {
		t.Error("B is not reported")
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = CheckArgs(a.Body.(Named))
		_ = Equal(Named{Decl: a}, a.Body)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("forming the unions of B[int] does not end")
	}
}
