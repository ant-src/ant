package types

import (
	"fmt"
	"strings"

	"github.com/ant-src/ant/ast"
)

// AliasCycleError describes the cycle the alias d is in when it reaches itself without passing
// through a record, array, map, set, or function (AliasCycle), or returns cyclic false. The
// message names the member of the cycle declared first, whichever member was resolved first, so
// the checker and the runtime blame the same declaration, and lists the other members in the
// order the cycle passes through them. blame is that member.
func AliasCycleError(d *Decl) (blame *Decl, msg string, cyclic bool) {
	return AliasCycleErrorIn(d, nil)
}

// AliasCycleErrorIn is AliasCycleError following only the bodies of the declarations in reports,
// which must hold those of d's strongly connected component (Components); nil follows every body.
func AliasCycleErrorIn(d *Decl, in func(*Decl) bool) (blame *Decl, msg string, cyclic bool) {
	members, cyclic := cycleMembers(d, in)
	if !cyclic {
		return nil, "", false
	}
	blame = d
	for _, m := range members {
		if declaredBefore(m, blame) {
			blame = m
		}
	}
	if blame != d {
		members, _ = cycleMembers(blame, in)
	}
	var through string
	if len(members) > 0 {
		names := make([]string, len(members))
		for i, m := range members {
			names[i] = m.Name
		}
		through = " through " + list(names)
	}
	return blame, fmt.Sprintf("type %s refers to itself%s without passing through a record, array, map, set, or function", blame.Name, through), true
}

// cycleMembers lists the aliases other than d on the cycle AliasCycle finds from d, in order: the
// ones whose body reaches d again. An alias the cycle only passes an argument through, as
// Id[T] = T, is left out.
func cycleMembers(d *Decl, in func(*Decl) bool) ([]*Decl, bool) {
	path, cyclic := AliasCycleIn(d, in)
	if !cyclic {
		return nil, false
	}
	var members []*Decl
	for _, m := range path {
		if _, ok := reachesTop(m.Body, d, in); ok {
			members = append(members, m)
		}
	}
	return members, true
}

// declaredBefore reports whether a is declared before b in the source. The members of an alias
// cycle are declared in one module, as imports are acyclic; a declaration without a position
// comes after the others, and the name breaks a tie.
func declaredBefore(a, b *Decl) bool {
	pa, oka := a.Pos.(ast.Pos)
	pb, okb := b.Pos.(ast.Pos)
	switch {
	case oka && okb && pa != pb:
		return pa.Line < pb.Line || pa.Line == pb.Line && pa.Col < pb.Col
	case oka != okb:
		return oka
	}
	return a.Name < b.Name
}

// list joins names as a sentence does: "A", "A and B", "A, B, and C".
func list(names []string) string {
	switch len(names) {
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	}
	return strings.Join(names[:len(names)-1], ", ") + ", and " + names[len(names)-1]
}
