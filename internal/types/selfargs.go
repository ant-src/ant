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
