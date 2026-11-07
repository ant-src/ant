package types

import (
	"fmt"
	"slices"
)

// SelfArgs describes how the generic type d refers to itself with other type arguments than
// its own parameters in order, or returns "" when it does not. It follows d's body or fields
// through the generic types they name, tracking which of their parameters stand for which of
// d's.
func SelfArgs(d *Decl) string { return SelfArgsIn(d, nil) }

// SelfArgsIn is SelfArgs following only the declarations in reports, which must hold those of
// d's strongly connected component (Components); nil follows every declaration. The others
// cannot lead back to d.
func SelfArgsIn(d *Decl, in func(*Decl) bool) string {
	if len(DeclBody(d)) == 0 {
		return ""
	}
	want := func() string {
		own := make([]Type, len(d.Params))
		for i, p := range d.Params {
			own[i] = Param{Decl: p}
		}
		return String(Named{Decl: d, Args: own})
	}
	// visited holds, for each generic type followed, the lists of d's parameters its own
	// stood for.
	visited := map[*Decl][][]*ParamDecl{}
	seen := func(decl *Decl, stands []*ParamDecl) bool {
		for _, s := range visited[decl] {
			if slices.Equal(s, stands) {
				return true
			}
		}
		visited[decl] = append(visited[decl], stands)
		return false
	}
	var msg string
	// env maps a parameter of the type being followed to the parameter of d it stands for.
	var follow func(ts []Type, env map[*ParamDecl]*ParamDecl, via *Decl)
	follow = func(ts []Type, env map[*ParamDecl]*ParamDecl, via *Decl) {
		for _, t := range ts {
			Walk(t, func(t Type) {
				n, ok := t.(Named)
				if msg != "" || !ok || len(DeclBody(n.Decl)) == 0 || n.Decl != d && in != nil && !in(n.Decl) {
					return
				}
				stands := make([]*ParamDecl, len(n.Decl.Params))
				for i := range n.Decl.Params {
					if i < len(n.Args) {
						if a, ok := n.Args[i].(Param); ok {
							stands[i] = env[a.Decl]
						}
					}
				}
				if n.Decl == d {
					for i, p := range d.Params {
						if stands[i] != p {
							if via == nil {
								msg = fmt.Sprintf("type %s refers to itself as %s; a generic type must refer to itself as %s, with its own type parameters in order", d.Name, String(n), want())
							} else {
								msg = fmt.Sprintf("type %s refers to itself through %s with other type arguments; a generic type must refer to itself as %s, with its own type parameters in order", d.Name, via.Name, want())
							}
							return
						}
					}
					return
				}
				if seen(n.Decl, stands) {
					return
				}
				next := make(map[*ParamDecl]*ParamDecl, len(stands))
				for i, p := range n.Decl.Params {
					next[p] = stands[i]
				}
				if via == nil {
					follow(DeclBody(n.Decl), next, n.Decl)
				} else {
					follow(DeclBody(n.Decl), next, via)
				}
			})
		}
	}
	env := map[*ParamDecl]*ParamDecl{}
	for _, p := range d.Params {
		env[p] = p
	}
	follow(DeclBody(d), env, nil)
	return msg
}

// AliasCycle reports whether the alias d reaches itself through alias bodies and unions alone,
// without passing through a record, array, map, set, or function (section 4, Type
// declarations), and the other aliases on the way.
func AliasCycle(d *Decl) (path []*Decl, cyclic bool) { return AliasCycleIn(d, nil) }

// AliasCycleIn is AliasCycle following only the bodies of the declarations in reports, which
// must hold those of d's strongly connected component (Components); nil follows every body.
func AliasCycleIn(d *Decl, in func(*Decl) bool) (path []*Decl, cyclic bool) {
	if d.Kind != AliasDecl {
		return nil, false
	}
	return reachesTop(d.Body, d, in)
}

// reachesTop reports whether t reaches the alias d through alias bodies and unions alone, without
// passing through a record, array, map, set, or function, and the other aliases on the way. It
// follows the bodies of the aliases in reports, all if in is nil.
func reachesTop(t Type, d *Decl, in func(*Decl) bool) (path []*Decl, found bool) {
	visited := map[*Decl]bool{}
	var walk func(t Type) bool
	walk = func(t Type) bool {
		switch x := t.(type) {
		case Named:
			if x.Decl.Kind != AliasDecl {
				return false
			}
			if x.Decl == d {
				return true
			}
			if visited[x.Decl] {
				return false
			}
			visited[x.Decl] = true
			path = append(path, x.Decl)
			if (in == nil || in(x.Decl)) && walk(x.Decl.Body) {
				return true
			}
			// An argument the alias puts where its body is, as Id[T] = T does, is reached
			// without a container too: type A[T] = Id[A[T]] is a cycle.
			for i, a := range x.Args {
				if i < len(x.Decl.Params) && unguarded(x.Decl, x.Decl.Params[i], map[*Decl]bool{}) && walk(a) {
					return true
				}
			}
			path = path[:len(path)-1]
		case Union:
			// The members as written first: a union flattens the aliases it names, and the
			// path through them says which declarations are in the cycle.
			for _, m := range append(x.Written, x.Members...) {
				if walk(m) {
					return true
				}
			}
		}
		return false
	}
	if !walk(t) {
		return nil, false
	}
	return path, true
}

// UnionSelf describes how the alias d names itself in a union inside its body, which is an error
// when the body is a union or names a type, or returns "". The outer union would need the inner
// one flattened, and the inner one the alias expanded, which forms the outer union again; a body
// that is an array, map, set, function or record type is not expanded to form a union, so
// type L[T] = Array[L[T] | int] is fine (section 4, Type declarations).
func UnionSelf(d *Decl) string { return UnionSelfIn(d, nil) }

// UnionSelfIn is UnionSelf following only the bodies of the declarations in reports, which must
// hold those of d's strongly connected component (Components); nil follows every body.
func UnionSelfIn(d *Decl, in func(*Decl) bool) string {
	if d.Kind != AliasDecl {
		return ""
	}
	switch d.Body.(type) {
	case Union, Named:
	default:
		return ""
	}
	// The body's own members reaching d are a cycle (AliasCycle), so only the unions below them
	// are looked at.
	parts := []Type{d.Body}
	if top, ok := d.Body.(Union); ok {
		parts = append(top.Written, top.Members...)
	}
	var msg string
	for _, part := range parts {
		Walk(part, func(t Type) {
			u, ok := t.(Union)
			if !ok || msg != "" {
				return
			}
			for _, m := range append(u.Written, u.Members...) {
				if _, found := reachesTop(m, d, in); found {
					msg = fmt.Sprintf("type %s names itself in the union %s; an alias whose body is a union or a type's name may not, so wrap the recursion in a named record type", d.Name, String(u))
					return
				}
			}
		})
	}
	return msg
}

// unguarded reports whether the alias d's body holds its parameter p outside any record, array,
// map, set, or function: as the body, as a member of a union, or as such an argument of another
// alias.
func unguarded(d *Decl, p *ParamDecl, seen map[*Decl]bool) bool {
	if seen[d] {
		return false
	}
	seen[d] = true
	var reach func(t Type) bool
	reach = func(t Type) bool {
		switch x := t.(type) {
		case Param:
			return x.Decl == p
		case Union:
			for _, m := range append(x.Written, x.Members...) {
				if reach(m) {
					return true
				}
			}
		case Named:
			if x.Decl.Kind != AliasDecl {
				return false
			}
			for i, a := range x.Args {
				if i < len(x.Decl.Params) && reach(a) && unguarded(x.Decl, x.Decl.Params[i], seen) {
					return true
				}
			}
		}
		return false
	}
	return reach(d.Body)
}
