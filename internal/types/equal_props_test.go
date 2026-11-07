package types

import (
	"math/rand"
	"testing"

	"github.com/ant-src/ant/value"
)

// TestEqualProperties checks on random types that Equal is an equivalence, that Hash agrees with
// it, and that Equal implies Assignable both ways (section 4, Types as values).
func TestEqualProperties(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	// Declarations: a generic record Box[T], a phantom alias K[T] = Array[int], an identity
	// alias Id[T] = T, an alias over a union U = int | string, a recursive alias
	// J = null | float64 | Array[J], its twin J2, and a generic alias N[T] = Array[NBox[T]] | Box[T]
	// that recurs through the field of the record NBox[T] = { items: N[T] }.
	p := &ParamDecl{Name: "T"}
	box := &Decl{Name: "Box", Kind: RecordDecl, Params: []*ParamDecl{p}, Fields: []Field{{Name: "v", Type: Param{Decl: p}}}}
	kp := &ParamDecl{Name: "T"}
	k := &Decl{Name: "K", Kind: AliasDecl, Params: []*ParamDecl{kp}, Body: Array{Elem: Prim{Kind: value.KindInt64}}}
	ip := &ParamDecl{Name: "T"}
	id := &Decl{Name: "Id", Kind: AliasDecl, Params: []*ParamDecl{ip}, Body: Param{Decl: ip}}
	uu, _ := MakeUnion([]Type{Prim{Kind: value.KindInt64}, Prim{Kind: value.KindString}})
	u := &Decl{Name: "U", Kind: AliasDecl, Body: uu}
	j := &Decl{Name: "J", Kind: AliasDecl}
	jb, _ := MakeUnion([]Type{Null{}, Prim{Kind: value.KindFloat64}, Array{Elem: Named{Decl: j}}})
	j.Body = jb
	j2 := &Decl{Name: "J2", Kind: AliasDecl}
	j2b, _ := MakeUnion([]Type{Null{}, Prim{Kind: value.KindFloat64}, Array{Elem: Named{Decl: j2}}})
	j2.Body = j2b
	np := &ParamDecl{Name: "T"}
	n := &Decl{Name: "N", Kind: AliasDecl, Params: []*ParamDecl{np}}
	nbp := &ParamDecl{Name: "T"}
	nbox := &Decl{Name: "NBox", Kind: RecordDecl, Params: []*ParamDecl{nbp}}
	nbox.Fields = []Field{{Name: "items", Type: Named{Decl: n, Args: []Type{Param{Decl: nbp}}}}}
	nb, _ := MakeUnion([]Type{Array{Elem: Named{Decl: nbox, Args: []Type{Param{Decl: np}}}}, Named{Decl: box, Args: []Type{Param{Decl: np}}}})
	n.Body = nb
	prims := []value.Kind{value.KindInt64, value.KindString, value.KindBool, value.KindFloat64}

	var gen func(depth int) Type
	gen = func(depth int) Type {
		c := r.Intn(14)
		if depth > 3 {
			c = r.Intn(3)
		}
		switch c {
		case 0:
			return Prim{Kind: prims[r.Intn(len(prims))]}
		case 1:
			return Null{}
		case 2:
			return Named{Decl: u}
		case 3:
			return Array{Elem: gen(depth + 1)}
		case 4:
			return Map{Key: Prim{Kind: value.KindString}, Val: gen(depth + 1)}
		case 5:
			return Set{Elem: gen(depth + 1)}
		case 6:
			return Named{Decl: box, Args: []Type{gen(depth + 1)}}
		case 7:
			return Named{Decl: k, Args: []Type{gen(depth + 1)}}
		case 8:
			return Named{Decl: id, Args: []Type{gen(depth + 1)}}
		case 9:
			if r.Intn(2) == 0 {
				return Named{Decl: j}
			}
			return Named{Decl: j2}
		case 10:
			return Named{Decl: n, Args: []Type{Prim{Kind: prims[r.Intn(2)]}}}
		case 11:
			if r.Intn(2) == 0 {
				return Record{Fields: []Field{{Name: "a", Type: gen(depth + 1)}, {Name: "b", Type: gen(depth + 1)}}}
			}
			return Func{Params: []Type{gen(depth + 1)}, Return: gen(depth + 1), Required: 1}
		default:
			m, err := MakeUnion([]Type{gen(depth + 1), gen(depth + 1)})
			if err != nil {
				return Null{}
			}
			return m
		}
	}
	var ts []Type
	for i := 0; i < 400; i++ {
		ts = append(ts, gen(0))
	}
	// Add expanded and re-written twins so that equal pairs are common.
	for i := 0; i < 400; i++ {
		ts = append(ts, Expand(ts[i]))
		ts = append(ts, Named{Decl: id, Args: []Type{ts[i]}})
		ts = append(ts, Named{Decl: k, Args: []Type{ts[i]}})
	}
	eq := make([][]bool, len(ts))
	fails := 0
	report := func(format string, a ...any) {
		fails++
		if fails < 25 {
			t.Errorf(format, a...)
		}
	}
	for a := range ts {
		eq[a] = make([]bool, len(ts))
		for b := range ts {
			eq[a][b] = Equal(ts[a], ts[b])
		}
	}
	for a := range ts {
		if !eq[a][a] {
			report("not reflexive: %s", String(ts[a]))
		}
		for b := range ts {
			if eq[a][b] != eq[b][a] {
				report("not symmetric: %s / %s", String(ts[a]), String(ts[b]))
			}
			if eq[a][b] {
				if Hash(ts[a]) != Hash(ts[b]) {
					report("hash differs: %s / %s", String(ts[a]), String(ts[b]))
				}
				if !Assignable(ts[a], ts[b]) {
					report("equal but not assignable: %s -> %s", String(ts[a]), String(ts[b]))
				}
			}
		}
	}
	for a := 0; a < len(ts) && fails < 25; a++ {
		for b := range ts {
			if !eq[a][b] {
				continue
			}
			for c := range ts {
				if eq[b][c] && !eq[a][c] {
					report("not transitive: %s = %s = %s", String(ts[a]), String(ts[b]), String(ts[c]))
				}
			}
		}
	}
	t.Logf("%d types, %d failures", len(ts), fails)
}
