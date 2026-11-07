package types

import "unsafe"

// pairs remembers, for one call of Equal or Assignable, the pairs of types the walk has met at
// an alias or a bounded type parameter, and what it found for each. A pair under test is
// assumed to hold, which is what ends the walk through a recursive type (section 4,
// Assignability); a pair already decided is not walked again, which keeps a type that shares
// structure through aliases, as F2 = func(F1): F1 does, from being walked as a tree.
//
// A pair that held may have held only because a pair still under test was assumed to. When
// such a pair fails, every pair decided to hold since it went under test loses its answer and
// is walked again if met. A pair that failed failed under assumptions that only add to what
// holds, so it fails without them.
type pairs struct {
	byKey map[[2]uint64][]*pair
	trail []*pair // the pairs put under test, in order; a pair that fails drops those after it
	rec   map[*Decl]bool
}

type pair struct {
	a, b  Type
	state pairState
}

type pairState uint8

const (
	pairUnknown pairState = iota // never decided, or its answer was dropped
	pairTesting                  // under test, assumed to hold
	pairHeld
	pairFailed
)

// remembered reports whether the walk remembers the pairs that t is part of: those of an alias
// without arguments, which is cheap to find again, and those of an alias that reaches itself,
// which the walk must find again to end. A generic alias that does not reach itself is not
// remembered: its instantiations can nest as deep as a value, as Opt[Opt[...]] does, and are
// rarely met twice.
func (m *pairs) remembered(t Type) bool {
	if !isAlias(t) {
		return false
	}
	n := t.(Named)
	if len(n.Args) == 0 {
		return true
	}
	if m.rec == nil {
		m.rec = map[*Decl]bool{}
	}
	if _, ok := m.rec[n.Decl]; !ok {
		m.findRecursive(n.Decl)
	}
	return m.rec[n.Decl]
}

// findRecursive records in m.rec, for d and every alias its body reaches, whether the alias
// leads back to itself through its body, directly or through other aliases, as Json.Value's
// does. Expanding any other alias ends after as many steps as there are aliases written in it
// and its bodies, so a walk meets the same pair again only through one of these. The aliases
// that lead back to themselves are those of a cycle among the aliases, found by Tarjan's
// method in time linear in the bodies, so a chain of aliases each naming the one before is not
// walked again from each of its links.
func (m *pairs) findRecursive(d *Decl) {
	aliases := func(v *Decl, edge func(*Decl)) {
		Walk(v.Body, func(x Type) {
			if isAlias(x) {
				edge(x.(Named).Decl)
			}
		})
	}
	decided := func(w *Decl) bool {
		_, ok := m.rec[w]
		return ok // decided by an earlier search, so on no cycle through d
	}
	Components(d, aliases, decided, func(comp []*Decl, cyclic bool) {
		for _, w := range comp {
			m.rec[w] = cyclic
		}
	})
}

// enter looks up the pair (a, b). Known reports that it is decided or under test, with holds
// its answer; otherwise the pair goes under test, and leave must be called with its answer.
func (m *pairs) enter(a, b Type) (p *pair, mark int, holds, known bool) {
	if m.byKey == nil {
		m.byKey = map[[2]uint64][]*pair{}
	}
	key := [2]uint64{shape(a, 0), shape(b, 0)}
	for _, q := range m.byKey[key] {
		if identical(q.a, a, 0) && identical(q.b, b, 0) {
			p = q
			break
		}
	}
	if p == nil {
		p = &pair{a: a, b: b}
		m.byKey[key] = append(m.byKey[key], p)
	}
	switch p.state {
	case pairTesting, pairHeld:
		return p, 0, true, true
	case pairFailed:
		return p, 0, false, true
	}
	p.state = pairTesting
	m.trail = append(m.trail, p)
	return p, len(m.trail) - 1, false, false
}

// leave records the answer for a pair that enter put under test.
func (m *pairs) leave(p *pair, mark int, holds bool) {
	if holds {
		p.state = pairHeld
		return
	}
	for _, q := range m.trail[mark+1:] {
		if q.state == pairHeld {
			q.state = pairUnknown
		}
	}
	m.trail = m.trail[:mark]
	p.state = pairFailed
}

// shape hashes the top few levels of t as written, expanding no alias, so that identical types
// hash alike.
func shape(t Type, depth int) uint64 {
	const prime = 1099511628211
	h := uint64(14695981039346656037)
	mix := func(u uint64) { h = (h ^ u) * prime }
	switch x := t.(type) {
	case Prim:
		mix(1)
		mix(uint64(x.Kind))
	case Null:
		mix(2)
	case Void:
		mix(3)
	case Never:
		mix(4)
	case Invalid:
		mix(5)
	case Array:
		mix(6)
		if depth < 3 {
			mix(shape(x.Elem, depth+1))
		}
	case Map:
		mix(7)
		if depth < 3 {
			mix(shape(x.Key, depth+1))
			mix(shape(x.Val, depth+1))
		}
	case Set:
		mix(8)
		if depth < 3 {
			mix(shape(x.Elem, depth+1))
		}
	case Func:
		mix(9)
		mix(uint64(len(x.Params)))
		if depth < 3 {
			for _, p := range x.Params {
				mix(shape(p, depth+1))
			}
			mix(shape(x.Return, depth+1))
		}
	case Record:
		mix(10)
		mix(uint64(len(x.Fields)))
	case Named:
		mix(11)
		mix(uint64(uintptr(unsafe.Pointer(x.Decl))))
		if depth < 3 {
			for _, a := range x.Args {
				mix(shape(a, depth+1))
			}
		}
	case Param:
		mix(12)
		mix(uint64(uintptr(unsafe.Pointer(x.Decl))))
	case Union:
		mix(13)
		mix(uint64(len(x.Members)))
		if depth < 3 {
			for _, m := range x.Members {
				mix(shape(m, depth+1))
			}
		}
	}
	return h
}
