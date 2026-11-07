package types

import "fmt"

// GenericRecursive describes the generic alias on a cycle of aliases through d, or returns
// cyclic false. An alias is recursive when it lies on a cycle of the graph whose edges lead from
// an alias to every declaration its body or the bounds of its parameters name, type arguments
// included, and from a record to nothing: a record's fields are never expanded, so recursion
// through them is finite, but a record's type arguments are compared, so they do not break a
// cycle. A recursive alias takes no type parameters, and neither does any alias on its cycle,
// since a generic alias that meets its own instantiation would have to be substituted without
// end (section 4, Type declarations). blame is the generic alias of the cycle declared first, so
// that the checker and the runtime blame the same declaration, and the message lists the other
// aliases the cycle passes through from it.
func GenericRecursive(d *Decl) (blame *Decl, msg string, cyclic bool) {
	return GenericRecursiveIn(d, nil)
}

// GenericRecursiveIn is GenericRecursive following only the declarations in reports, which must
// hold those of d's strongly connected component (Components); nil follows every declaration.
func GenericRecursiveIn(d *Decl, in func(*Decl) bool) (blame *Decl, msg string, cyclic bool) {
	if d.Kind != AliasDecl {
		return nil, "", false
	}
	var comp []*Decl
	skip := func(w *Decl) bool { return in != nil && !in(w) }
	Components(d, aliasRefs, skip, func(c []*Decl, isCyclic bool) {
		comp, cyclic = c, isCyclic
	})
	if !cyclic {
		return nil, "", false
	}
	for _, m := range comp {
		if len(m.Params) > 0 && (blame == nil || declaredBefore(m, blame)) {
			blame = m
		}
	}
	if blame == nil {
		return nil, "", false // a cycle of plain aliases, which AliasCycle may report
	}
	// The aliases the cycle passes through from blame back to itself, found along the
	// component's edges alone.
	set := make(map[*Decl]bool, len(comp))
	for _, m := range comp {
		set[m] = true
	}
	var path []*Decl
	visited := map[*Decl]bool{}
	var walk func(v *Decl) bool
	walk = func(v *Decl) bool {
		found := false
		aliasRefs(v, func(w *Decl) {
			if found || !set[w] {
				return
			}
			if w == blame {
				found = true
				return
			}
			if visited[w] {
				return
			}
			visited[w] = true
			path = append(path, w)
			if walk(w) {
				found = true
				return
			}
			path = path[:len(path)-1]
		})
		return found
	}
	walk(blame)
	var through string
	if len(path) > 0 {
		names := make([]string, len(path))
		for i, m := range path {
			names[i] = m.Name
		}
		through = " through " + list(names)
	}
	return blame, fmt.Sprintf("type %s refers to itself%s without passing through a record's field; an alias with type parameters may not, so wrap the recursion in a named record type", blame.Name, through), true
}

// aliasRefs calls f with each alias that the alias d's body or the bounds of its parameters
// name, as written: type arguments included, and the members of a union before it was formed,
// since forming it expands an alias member, whose name the cycle is reported by.
func aliasRefs(d *Decl, f func(*Decl)) {
	if d.Kind != AliasDecl {
		return
	}
	DeclRefs(d, func(e *Decl) {
		if e.Kind == AliasDecl {
			f(e)
		}
	})
}
