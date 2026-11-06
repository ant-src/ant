package types

import "fmt"

// SelfArgs describes how the generic type d refers to itself with other type arguments than
// its own parameters in order, or returns "" when it does not. It follows d's body or fields
// through the generic types they name, tracking which of their parameters stand for which of
// d's.
func SelfArgs(d *Decl) string {
	if len(DeclBody(d)) == 0 {
		return ""
	}
	own := make([]Type, len(d.Params))
	for i, p := range d.Params {
		own[i] = Param{Decl: p}
	}
	want := String(Named{Decl: d, Args: own})
	// env maps a parameter of the alias being followed to the parameter of d it stands for.
	type visit struct {
		decl *Decl
		env  string
	}
	visited := map[visit]bool{}
	var msg string
	var follow func(ts []Type, env map[*ParamDecl]*ParamDecl, via *Decl)
	follow = func(ts []Type, env map[*ParamDecl]*ParamDecl, via *Decl) {
		for _, t := range ts {
			Walk(t, func(t Type) {
				n, ok := t.(Named)
				if msg != "" || !ok || len(DeclBody(n.Decl)) == 0 {
					return
				}
				next := map[*ParamDecl]*ParamDecl{}
				key := ""
				for i, p := range n.Decl.Params {
					var stands *ParamDecl
					if i < len(n.Args) {
						if a, ok := n.Args[i].(Param); ok {
							stands = env[a.Decl]
						}
					}
					next[p] = stands
					key += fmt.Sprintf("%p,", stands)
				}
				if n.Decl == d {
					for i, p := range d.Params {
						if next[n.Decl.Params[i]] != p {
							if via == nil {
								msg = fmt.Sprintf("type %s refers to itself as %s; a generic type must refer to itself as %s, with its own type parameters in order", d.Name, String(n), want)
							} else {
								msg = fmt.Sprintf("type %s refers to itself through %s with other type arguments; a generic type must refer to itself as %s, with its own type parameters in order", d.Name, via.Name, want)
							}
							return
						}
					}
					return
				}
				if visited[visit{n.Decl, key}] {
					return
				}
				visited[visit{n.Decl, key}] = true
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
func AliasCycle(d *Decl) (path []*Decl, cyclic bool) {
	if d.Kind != AliasDecl {
		return nil, false
	}
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
			if walk(x.Decl.Body) {
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
	if !walk(d.Body) {
		return nil, false
	}
	return path, true
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
