package types

import "fmt"

// Subst replaces type parameters by their bindings. Unions are formed again after
// substitution, which can fail where the arguments are written (section 4, Unions).
func Subst(t Type, m map[*ParamDecl]Type) (Type, error) {
	if len(m) == 0 {
		return t, nil
	}
	return (&substituter{m: m}).subst(t)
}

// substPlain substitutes as Subst does but does not validate the bodies of the declarations
// the result names. Expanding an alias needs only its body: its arguments were validated where
// they were written, and validating again from inside a union being formed, as in
// type A[T] = B[T] | null with B[T] = { next: A[T] | int }, would start each time with fresh
// guards and never end.
func substPlain(t Type, m map[*ParamDecl]Type) (Type, error) {
	if len(m) == 0 {
		return t, nil
	}
	return (&substituter{m: m, plain: true}).subst(t)
}

// substituter carries the bindings, the declarations whose bodies are being validated, and the
// instantiations already validated in this walk, so that a type reached along many paths, as
// T1 is from T0[X] = { a: T1[X], b: T1[X] }, is validated once.
type substituter struct {
	m     map[*ParamDecl]Type
	seen  map[*Decl]bool
	done  *map[*Decl][][]Type
	plain bool // substPlain: no declaration bodies are validated
	// outer holds the alias instantiations whose union bodies are being expanded, for the
	// unions formed inside them (makeUnion)
	outer []Named
}

// substUnionBody substitutes the body of the alias instantiation n under its bindings m without
// validating, and forms a union body inside the expansion of n and of outer (makeUnion).
func substUnionBody(n Named, m map[*ParamDecl]Type, outer []Named) (Type, error) {
	if _, isUnion := n.Decl.Body.(Union); !isUnion || len(m) == 0 {
		return substPlain(n.Decl.Body, m)
	}
	return (&substituter{m: m, plain: true, outer: append(outer[:len(outer):len(outer)], n)}).subst(n.Decl.Body)
}

// validated reports whether n's body was already validated with these arguments.
func (s *substituter) validated(n Named) bool {
	for _, args := range (*s.done)[n.Decl] {
		same := true
		for i := range args {
			if !Equal(args[i], n.Args[i]) {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	return false
}

func (s *substituter) subst(t Type) (Type, error) {
	m := s.m
	switch x := t.(type) {
	case Param:
		if b, ok := m[x.Decl]; ok {
			return b, nil
		}
		return t, nil
	case Array:
		e, err := s.subst(x.Elem)
		if err != nil {
			return nil, err
		}
		return Array{Elem: e}, nil
	case Map:
		k, err := s.subst(x.Key)
		if err != nil {
			return nil, err
		}
		v, err := s.subst(x.Val)
		if err != nil {
			return nil, err
		}
		return Map{Key: k, Val: v}, nil
	case Set:
		e, err := s.subst(x.Elem)
		if err != nil {
			return nil, err
		}
		return Set{Elem: e}, nil
	case Func:
		ps := make([]Type, len(x.Params))
		for i, p := range x.Params {
			q, err := s.subst(p)
			if err != nil {
				return nil, err
			}
			ps[i] = q
		}
		r, err := s.subst(x.Return)
		if err != nil {
			return nil, err
		}
		return Func{Params: ps, Return: r, Required: x.Required, Names: x.Names}, nil
	case Record:
		fs := make([]Field, len(x.Fields))
		for i, f := range x.Fields {
			ft, err := s.subst(f.Type)
			if err != nil {
				return nil, err
			}
			fs[i] = Field{Name: f.Name, Type: ft, Default: f.Default}
		}
		return Record{Fields: fs}, nil
	case Named:
		if len(x.Args) == 0 {
			return t, nil
		}
		args := make([]Type, len(x.Args))
		for i, a := range x.Args {
			b, err := s.subst(a)
			if err != nil {
				return nil, err
			}
			args[i] = b
		}
		n := Named{Decl: x.Decl, Args: args}
		if s.done == nil {
			s.done = &map[*Decl][][]Type{}
		}
		if body := DeclBody(x.Decl); len(body) > 0 && !s.plain && !s.seen[x.Decl] && !s.validated(n) {
			// The arguments may form an invalid union inside the declaration, as
			// Either[Box[int], Box[string]] does in an alias's body and Outer[M] does in the field
			// of the Box[M] it holds, however deep (section 4, Unions). A type that reaches
			// itself is checked once, since it passes its own parameters (section 4, Type
			// declarations).
			if s.seen == nil {
				s.seen = map[*Decl]bool{}
			}
			s.seen[x.Decl] = true
			inner := &substituter{m: Bindings(x.Decl.Params, args), seen: s.seen, done: s.done}
			for _, t := range body {
				if _, err := inner.subst(t); err != nil {
					delete(s.seen, x.Decl)
					return nil, fmt.Errorf("in %s: %w", String(n), err)
				}
			}
			delete(s.seen, x.Decl)
			(*s.done)[x.Decl] = append((*s.done)[x.Decl], args)
		}
		return n, nil
	case Union:
		ms := make([]Type, len(x.Members))
		for i, mem := range x.Members {
			n, err := s.subst(mem)
			if err != nil {
				return nil, err
			}
			ms[i] = n
		}
		return makeUnion(ms, s.outer)
	}
	return t, nil
}

// Mentions reports whether t mentions any type parameter in the set.
func Mentions(t Type, set map[*ParamDecl]bool) bool {
	return mentions(t, func(p Param) bool { return set[p.Decl] })
}

// MentionsOther reports whether t mentions a type parameter outside the set.
func MentionsOther(t Type, set map[*ParamDecl]bool) bool {
	return mentions(t, func(p Param) bool { return !set[p.Decl] })
}

// HasParam reports whether t mentions any type parameter at all.
func HasParam(t Type) bool {
	return mentions(t, func(Param) bool { return true })
}

func mentions(t Type, pred func(Param) bool) bool {
	switch x := t.(type) {
	case Param:
		return pred(x)
	case Array:
		return mentions(x.Elem, pred)
	case Map:
		return mentions(x.Key, pred) || mentions(x.Val, pred)
	case Set:
		return mentions(x.Elem, pred)
	case Func:
		for _, p := range x.Params {
			if mentions(p, pred) {
				return true
			}
		}
		return mentions(x.Return, pred)
	case Record:
		for _, f := range x.Fields {
			if mentions(f.Type, pred) {
				return true
			}
		}
	case Named:
		for _, a := range x.Args {
			if mentions(a, pred) {
				return true
			}
		}
	case Union:
		for _, m := range x.Members {
			if mentions(m, pred) {
				return true
			}
		}
	}
	return false
}

// Walk calls f on t and on every type written inside it, the type arguments of a named type
// included, but not on the bodies or fields of the declarations it names.
func Walk(t Type, f func(Type)) {
	f(t)
	switch x := t.(type) {
	case Array:
		Walk(x.Elem, f)
	case Map:
		Walk(x.Key, f)
		Walk(x.Val, f)
	case Set:
		Walk(x.Elem, f)
	case Func:
		for _, p := range x.Params {
			Walk(p, f)
		}
		Walk(x.Return, f)
	case Record:
		for _, fl := range x.Fields {
			Walk(fl.Type, f)
		}
	case Named:
		for _, a := range x.Args {
			Walk(a, f)
		}
	case Union:
		for _, m := range x.Members {
			Walk(m, f)
		}
	}
}

// DeclBody returns the types a generic declaration substitutes its arguments into: an alias's
// body or a record's field types. It is empty for a declaration without type parameters.
func DeclBody(d *Decl) []Type {
	if len(d.Params) == 0 {
		return nil
	}
	switch d.Kind {
	case AliasDecl:
		if d.Body != nil {
			return []Type{d.Body}
		}
	case RecordDecl:
		out := make([]Type, len(d.Fields))
		for i, f := range d.Fields {
			out[i] = f.Type
		}
		return out
	}
	return nil
}

// CheckArgs forms the unions that n's arguments make in its declaration, and in every
// instantiation that declaration reaches, however deep (section 4, Unions): Outer[M] fails when
// the Box[M] in its field does.
func CheckArgs(n Named) error {
	m := Bindings(n.Decl.Params, n.Args)
	for _, t := range DeclBody(n.Decl) {
		if _, err := Subst(t, m); err != nil {
			return err
		}
	}
	return nil
}

// CheckWritten applies CheckArgs to every instantiation written in t.
func CheckWritten(t Type) error {
	var first error
	Walk(t, func(t Type) {
		if n, ok := t.(Named); ok && first == nil && len(n.Args) > 0 {
			if err := CheckArgs(n); err != nil {
				first = fmt.Errorf("%s: %w", String(n), err)
			}
		}
	})
	return first
}
