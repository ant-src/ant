package types

// Components finds the strongly connected components of the declarations reached from root
// along the edges succ lists, by Tarjan's method in time linear in the edges, and calls done
// with each, a component before any that reaches it; the last is root's. A declaration skip
// reports, when first reached, is left out with its edges: one decided by an earlier search,
// or one that cannot be on a cycle through root. Cyclic reports that the component has more
// than one declaration, or one with an edge to itself.
//
// A declaration refers to itself only through the declarations of its own component, so a
// search for such a reference can stop at every other one: following each declaration a chain
// of aliases reaches, from each of its links, would take time quadratic in its length.
func Components(root *Decl, succ func(d *Decl, edge func(*Decl)), skip func(*Decl) bool, done func(comp []*Decl, cyclic bool)) {
	index := map[*Decl]int{}
	low := map[*Decl]int{}
	onStack := map[*Decl]bool{}
	self := map[*Decl]bool{}
	var stack []*Decl
	var visit func(v *Decl)
	visit = func(v *Decl) {
		index[v], low[v] = len(index), len(index)
		stack = append(stack, v)
		onStack[v] = true
		succ(v, func(w *Decl) {
			if w == v {
				self[v] = true
			}
			if _, ok := index[w]; !ok {
				if skip != nil && skip(w) {
					return
				}
				visit(w)
				low[v] = min(low[v], low[w])
			} else if onStack[w] {
				low[v] = min(low[v], index[w])
			}
		})
		if low[v] != index[v] {
			return
		}
		i := len(stack) - 1
		for stack[i] != v {
			i--
		}
		comp := append([]*Decl(nil), stack[i:]...)
		stack = stack[:i]
		for _, w := range comp {
			onStack[w] = false
		}
		done(comp, len(comp) > 1 || self[v])
	}
	visit(root)
}

// DeclRefs calls f with each declaration that d's alias body or field types name, the members
// of a union as written included.
func DeclRefs(d *Decl, f func(*Decl)) {
	if d.Kind == AliasDecl && d.Body != nil {
		namedIn(d.Body, f)
	}
	for _, fl := range d.Fields {
		namedIn(fl.Type, f)
	}
}

func namedIn(t Type, f func(*Decl)) {
	Walk(t, func(t Type) {
		switch x := t.(type) {
		case Named:
			f(x.Decl)
		case Union:
			for _, w := range x.Written {
				namedIn(w, f)
			}
		}
	})
}
