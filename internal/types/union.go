package types

import (
	"fmt"
	"slices"
	"strings"
)

// MakeUnion forms the union of members: flattened, deduplicated, never dropped. It reports
// an error when two members share a discriminator (section 4, Unions).
func MakeUnion(members []Type) (Type, error) { return makeUnion(members, nil) }

// makeUnion is MakeUnion inside the expansion of the alias instantiations in outer, which a
// union formed in their bodies leaves out rather than expanding again, so that the cycle
// A[T] = B[T] | null with B[T] = A[T] | null, which the checker reports, ends.
func makeUnion(members []Type, outer []Named) (Type, error) {
	var flat []Type
	expanding := map[*Decl]bool{} // aliases being flattened, so that a cyclic alias terminates
	aliased := false              // an alias was expanded, so the written form is kept for messages
	// Each written member is expanded once: the written form below needs its discriminator too,
	// and expanding it again there would expand an alias nested in its own arguments, as in
	// Opt[Opt[Opt[int]]], twice at every level, which is exponential in the nesting.
	written := make([]Type, len(members))
	var collect func(ts []Type, top bool)
	collect = func(ts []Type, top bool) {
		for i, t := range ts {
			alias := false
			if n, ok := t.(Named); ok && n.Decl.Kind == AliasDecl {
				if expanding[n.Decl] || slices.ContainsFunc(outer, func(o Named) bool { return identical(o, n, 0) }) {
					continue
				}
				expanding[n.Decl] = true
				alias, aliased = true, true
			}
			e := expandTopIn(t, outer)
			if top {
				written[i] = e
			}
			switch x := e.(type) {
			case Union:
				collect(x.Members, false)
			case Never:
			default:
				dup := false
				for _, f := range flat {
					if sameMember(f, t) {
						dup = true
						break
					}
				}
				if !dup {
					flat = append(flat, t)
				}
			}
			if alias {
				// Only while its own body is flattened: a sibling may name the same alias with
				// other arguments, as in Opt[int] | Opt[string].
				delete(expanding, t.(Named).Decl)
			}
		}
	}
	collect(members, true)
	switch len(flat) {
	case 0:
		return Never{}, nil
	case 1:
		return flat[0], nil
	}
	if err := checkDiscriminators(flat); err != nil {
		return Invalid{}, err
	}
	u := Union{Members: flat}
	if aliased {
		var seen []Type // the expansions of u.Written
		for i, m := range members {
			if _, never := m.(Never); never {
				continue
			}
			dup := false
			for j, w := range u.Written {
				if sameDiscriminator(seen[j], written[i]) && Equal(w, m) {
					dup = true
					break
				}
			}
			if !dup {
				u.Written = append(u.Written, m)
				seen = append(seen, written[i])
			}
		}
	}
	return u, nil
}

// sameMember reports whether two members of a union are the same type. Only members with one
// discriminator, or two without one, can be, and comparing that first spares Equal the expansion of an alias whose
// body holds a union naming the alias itself, which needs this very union.
func sameMember(a, b Type) bool { return sameDiscriminator(a, b) && Equal(a, b) }

// sameDiscriminator reports whether is cannot tell a from b apart: both have one discriminator,
// or neither has one.
func sameDiscriminator(a, b Type) bool {
	ka, _, oka := Discriminator(a)
	kb, _, okb := Discriminator(b)
	return oka == okb && (!oka || ka == kb)
}

// Join forms the union of branch types where control merges, or returns an error to report.
func Join(ts ...Type) (Type, error) { return MakeUnion(ts) }

// expandTop expands an alias as far as a union needs to see it: a body that is a union, a type
// parameter, or another alias is substituted, and any other body comes back as declared, with
// its own parameters. Its form and discriminator are those of the full expansion, and nothing
// inside an array, map, set, function or record is substituted, which for an alias whose body
// holds a union naming the alias itself, as in type L[T] = Array[L[T] | int], would not end.
func expandTop(t Type) Type { return expandTopIn(t, nil) }

// expandTopIn is expandTop inside the expansion of the alias instantiations in outer. A union
// body is formed with the instantiation added to them (makeUnion).
func expandTopIn(t Type, outer []Named) Type {
	var met []Named
	for i := 0; ; i++ {
		n, ok := t.(Named)
		if !ok || n.Decl.Kind != AliasDecl || n.Decl.Body == nil {
			return t
		}
		if i >= 64 {
			// A chain such as Id[Id[...Id[int]...]] meets one alias many times, each time with
			// other arguments; only the same instantiation again is a cycle.
			if slices.ContainsFunc(met, func(o Named) bool { return identical(o, n, 0) }) {
				panic("internal error: expanding the cyclic alias " + n.Decl.Name + ", which the checker accepted")
			}
			met = append(met, n)
		}
		m := Bindings(n.Decl.Params, n.Args)
		switch b := n.Decl.Body.(type) {
		case Union, Param:
		case Named:
			// Only the arguments: substituting the alias itself checks its body, which may
			// hold a union naming the alias being expanded.
			args := make([]Type, len(b.Args))
			for i, a := range b.Args {
				s, err := substPlain(a, m)
				if err != nil {
					panic("internal error: expanding " + String(n) + ", which the checker accepted: " + err.Error())
				}
				args[i] = s
			}
			t = Named{Decl: b.Decl, Args: args}
			continue
		default:
			return n.Decl.Body
		}
		var body Type
		var err error
		body, err = substUnionBody(n, m, outer)
		if err != nil {
			panic("internal error: expanding " + String(n) + ", which the checker accepted: " + err.Error())
		}
		t = body
	}
}

// Discriminator is what is tests at run time: a kind, a tag, or a shape class, as a key that two
// types share exactly when is cannot tell them apart, and as words for a message. A sync type
// has none, since is does not accept one (section 4, Unions).
func Discriminator(t Type) (key, describe string, ok bool) {
	switch x := expandTop(t).(type) {
	case Prim:
		return "kind:" + x.Kind.String(), x.Kind.String(), true
	case Null:
		return "null", "null", true
	case Void:
		return "void", "void", true
	case Array:
		return "array", "an array", true
	case Map:
		return "map", "a map", true
	case Set:
		return "set", "a set", true
	case Func:
		return "func", "a function", true
	case Named:
		switch x.Decl.Kind {
		case KindOpaqueDecl:
			return "kind:" + x.Decl.Base.String(), x.Decl.Base.String(), true // Bytes or Html: a kind of its own
		case OpaqueDecl:
			return "", "", false // a sync type: is does not accept it
		}
		return fmt.Sprintf("tag:%p", x.Decl), "the tag " + x.Decl.Name, true // a record, primitive or time type
	case Record:
		return "record", "an anonymous record", true
	case Param:
		return fmt.Sprintf("param:%p", x.Decl), "the type parameter " + x.Decl.Name, true
	}
	return "", "", false
}

// CheckUnion reports an error when the members of a formed union, which may name an alias
// whose body was not yet known when the union was formed, share a discriminator.
func CheckUnion(members []Type) error { return checkDiscriminators(members) }

func checkDiscriminators(members []Type) error {
	seen := map[string]Type{}
	var record Type
	var namedRecords, params []Type

	var opaque Type
	for _, m := range members {
		key, desc, ok := Discriminator(m)
		if !ok {
			if n, isNamed := expandTop(m).(Named); isNamed && n.Decl.Kind == OpaqueDecl {
				if opaque != nil {
					return fmt.Errorf("a union may contain at most one type that is cannot test, since it cannot tell %s from %s", String(opaque), String(m))
				}
				opaque = m
			}
			continue
		}
		if prev, dup := seen[key]; dup {
			return fmt.Errorf("a union may contain each discriminator at most once: %s and %s are both %s; narrow first or give the whole an expected type",
				String(prev), String(m), strings.TrimPrefix(desc, "the "))
		}
		seen[key] = m
		switch x := expandTop(m).(type) {
		case Record:
			record = m
		case Param:
			params = append(params, m)
		case Named:
			if x.Decl.Kind == RecordDecl {
				namedRecords = append(namedRecords, m)
			}
		}
	}
	if record != nil && len(namedRecords) > 0 {
		return fmt.Errorf("a union may not contain an anonymous record type %s alongside the named record type %s: is cannot tell them apart",
			String(record), String(namedRecords[0]))
	}
	if len(params) > 0 {
		for _, m := range members {
			switch x := expandTop(m).(type) {
			case Param, Null, Void:
			case Named:
				if len(x.Args) > 0 {
					return fmt.Errorf("a type parameter may not share a union with %s, which takes type arguments", String(m))
				}
			default:
				return fmt.Errorf("a type parameter may share a union only with null, void, other type parameters, and named types without type arguments, not with %s", String(m))
			}
		}
	}
	return nil
}
