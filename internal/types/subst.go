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

// substRaw substitutes as substPlain does but forms each union raw (makeUnion), for Equal.
func substRaw(t Type, m map[*ParamDecl]Type) Type {
	if len(m) == 0 {
		return t
	}
	t, _ = (&substituter{m: m, plain: true, raw: true}).subst(t)
	return t
}

// substituter carries the bindings, the declarations whose bodies are being validated, and the
// instantiations already validated in this walk, so that a type reached along many paths, as
// T1 is from T0[X] = { a: T1[X], b: T1[X] }, is validated once.
type substituter struct {
	m     map[*ParamDecl]Type
	seen  map[*Decl]bool
	done  *map[*Decl][][]Type
	plain bool // substPlain: no declaration bodies are validated
	raw   bool // substRaw: unions are formed raw (makeUnion), and plain
	// memo, when set, holds the declarations whose instantiation with distinct type parameters
	// is known valid, and learned collects those this walk finds valid (ArgsMemo).
	memo    *ArgsMemo
	learned *[]*Decl
	// outer holds the alias instantiations whose union bodies are being expanded, for the
	// unions formed inside them (makeUnion)
	outer []Named
}

// substUnionBody substitutes the body of the alias instantiation n under its bindings m without
// validating, and forms a union body inside the expansion of n and of outer (makeUnion), raw
// when raw is set.
func substUnionBody(n Named, m map[*ParamDecl]Type, outer []Named, raw bool) (Type, error) {
	if len(m) == 0 {
		return n.Decl.Body, nil
	}
	s := &substituter{m: m, plain: true, raw: raw}
	if _, isUnion := n.Decl.Body.(Union); isUnion {
		s.outer = append(outer[:len(outer):len(outer)], n)
	}
	return s.subst(n.Decl.Body)
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
		if body := DeclBody(x.Decl); len(body) > 0 && !s.plain && !s.seen[x.Decl] && !s.memo.knows(n) && !s.validated(n) {
			// The arguments may form an invalid union inside the declaration, as
			// Either[Box[int], Box[string]] does in an alias's body and Outer[M] does in the field
			// of the Box[M] it holds, however deep (section 4, Unions). A type that reaches
			// itself is checked once, since it passes its own parameters (section 4, Type
			// declarations).
			if s.seen == nil {
				s.seen = map[*Decl]bool{}
			}
			s.seen[x.Decl] = true
			inner := &substituter{m: Bindings(x.Decl.Params, args), seen: s.seen, done: s.done, memo: s.memo, learned: s.learned}
			for _, t := range body {
				if _, err := inner.subst(t); err != nil {
					delete(s.seen, x.Decl)
					return nil, fmt.Errorf("in %s: %w", String(n), err)
				}
			}
			delete(s.seen, x.Decl)
			(*s.done)[x.Decl] = append((*s.done)[x.Decl], args)
			if s.memo.remembers(n) {
				*s.learned = append(*s.learned, x.Decl)
			}
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
		return makeUnion(ms, s.outer, s.raw)
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
func Walk(t Type, f func(Type)) { walk(t, f, nil) }

// walk calls pre on t before the types written inside it and post after them.
func walk(t Type, pre, post func(Type)) {
	if pre != nil {
		pre(t)
	}
	switch x := t.(type) {
	case Array:
		walk(x.Elem, pre, post)
	case Map:
		walk(x.Key, pre, post)
		walk(x.Val, pre, post)
	case Set:
		walk(x.Elem, pre, post)
	case Func:
		for _, p := range x.Params {
			walk(p, pre, post)
		}
		walk(x.Return, pre, post)
	case Record:
		for _, fl := range x.Fields {
			walk(fl.Type, pre, post)
		}
	case Named:
		for _, a := range x.Args {
			walk(a, pre, post)
		}
	case Union:
		for _, m := range x.Members {
			walk(m, pre, post)
		}
	}
	if post != nil {
		post(t)
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
func CheckArgs(n Named) error { return (*ArgsMemo)(nil).CheckArgs(n) }

// ArgsMemo remembers, from one call of its CheckArgs or CheckWritten to the next, the generic
// declarations whose instantiation with distinct type parameters as arguments was found valid,
// as the body of D[T] = Array[C[T]] instantiates C. Each such instantiation of a declaration is
// valid exactly when every other one is: forming a union tells type parameters apart by
// identity alone, and the bodies an instantiation reaches mention no type parameter but their
// own, which substitution replaces. So a chain of such declarations is validated once, not once
// from each link. A memo may be used only while the declarations it reaches do not change, and
// their type parameters are their own; a nil memo remembers nothing.
type ArgsMemo struct {
	valid map[*Decl]bool
	// Only, when set, limits the memo to the declarations it accepts, those whose bodies, and
	// the bodies of all they reach, no longer change and mention no type parameter but their own.
	Only func(*Decl) bool
}

func (m *ArgsMemo) knows(n Named) bool {
	return m != nil && m.valid[n.Decl] && distinctParams(n)
}

// remembers reports whether m may remember the instantiation n as valid.
func (m *ArgsMemo) remembers(n Named) bool {
	return m != nil && distinctParams(n) && (m.Only == nil || m.Only(n.Decl))
}

// distinctParams reports whether n's arguments are distinct type parameters.
func distinctParams(n Named) bool {
	for i, a := range n.Args {
		p, ok := a.(Param)
		if !ok {
			return false
		}
		for _, b := range n.Args[:i] {
			if b.(Param).Decl == p.Decl {
				return false
			}
		}
	}
	return len(n.Args) > 0
}

// CheckArgs is the function CheckArgs, remembering what it validates when n is valid. One that
// fails may have assumed, along the way, that the instantiation it started from holds.
func (m *ArgsMemo) CheckArgs(n Named) error {
	if m.knows(n) {
		return nil
	}
	b := Bindings(n.Decl.Params, n.Args)
	var learned []*Decl
	for _, t := range DeclBody(n.Decl) {
		if len(b) == 0 {
			break
		}
		if _, err := (&substituter{m: b, memo: m, learned: &learned}).subst(t); err != nil {
			return err
		}
	}
	if m != nil {
		if m.valid == nil {
			m.valid = map[*Decl]bool{}
		}
		for _, d := range learned {
			m.valid[d] = true
		}
		if m.remembers(n) {
			m.valid[n.Decl] = true
		}
	}
	return nil
}

// CheckWritten applies CheckArgs to every instantiation written in t, its arguments before it, as
// the checker resolves them: Opt[Either[Box[int], Box[string]]] fails on the Either, which forming
// the union in the body of Opt would expand.
func CheckWritten(t Type) error { return (*ArgsMemo)(nil).CheckWritten(t) }

// CheckWritten is the function CheckWritten with m's CheckArgs.
func (m *ArgsMemo) CheckWritten(t Type) error {
	var first error
	walk(t, nil, func(t Type) {
		if n, ok := t.(Named); ok && first == nil && len(n.Args) > 0 {
			if err := m.CheckArgs(n); err != nil {
				first = fmt.Errorf("%s: %w", String(n), err)
			}
		}
	})
	return first
}
