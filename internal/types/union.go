package types

import (
	"fmt"
	"strings"
)

// MakeUnion forms the union of members: flattened, deduplicated, never dropped. It reports
// an error when two members share a discriminator (section 4, Unions).
func MakeUnion(members []Type) (Type, error) {
	var flat []Type
	expanding := map[*Decl]bool{} // aliases being flattened, so that a cyclic alias terminates
	aliased := false              // an alias was expanded, so the written form is kept for messages
	var collect func(ts []Type)
	collect = func(ts []Type) {
		for _, t := range ts {
			if n, ok := t.(Named); ok && n.Decl.Kind == AliasDecl {
				if expanding[n.Decl] {
					continue
				}
				expanding[n.Decl] = true
				aliased = true
			}
			switch x := Expand(t).(type) {
			case Union:
				collect(x.Members)
			case Never:
			default:
				dup := false
				for _, f := range flat {
					if Equal(f, t) {
						dup = true
						break
					}
				}
				if !dup {
					flat = append(flat, t)
				}
			}
		}
	}
	collect(members)
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
		for _, m := range members {
			if _, never := m.(Never); never {
				continue
			}
			dup := false
			for _, w := range u.Written {
				if Equal(w, m) {
					dup = true
					break
				}
			}
			if !dup {
				u.Written = append(u.Written, m)
			}
		}
	}
	return u, nil
}

// Join forms the union of branch types where control merges, or returns an error to report.
func Join(ts ...Type) (Type, error) { return MakeUnion(ts) }

// Discriminator is what is tests at run time: a kind, a tag, or a shape class, as a key that two
// types share exactly when is cannot tell them apart, and as words for a message. A sync type
// has none, since is does not accept one (section 4, Unions).
func Discriminator(t Type) (key, describe string, ok bool) {
	switch x := Expand(t).(type) {
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
		if x.Decl.Kind == OpaqueDecl {
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

func checkDiscriminators(members []Type) error {
	seen := map[string]Type{}
	var record Type
	var namedRecords, params []Type

	var opaque Type
	for _, m := range members {
		key, desc, ok := Discriminator(m)
		if !ok {
			if n, isNamed := Expand(m).(Named); isNamed && n.Decl.Kind == OpaqueDecl {
				if opaque != nil {
					return fmt.Errorf("a union may contain at most one sync type, since is cannot tell %s from %s", String(opaque), String(m))
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
		switch x := Expand(m).(type) {
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
			switch x := Expand(m).(type) {
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
