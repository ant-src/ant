package types

// Subst replaces type parameters by their bindings. Unions are formed again after
// substitution, which can fail where the arguments are written (section 4, Unions).
func Subst(t Type, m map[*ParamDecl]Type) (Type, error) {
	if len(m) == 0 {
		return t, nil
	}
	return (&substituter{m: m}).subst(t)
}

// substituter carries the bindings and the aliases whose bodies are being validated.
type substituter struct {
	m    map[*ParamDecl]Type
	seen map[*Decl]bool
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
		if x.Decl.Kind == AliasDecl && x.Decl.Body != nil && !s.seen[x.Decl] {
			// The arguments may form an invalid union inside the alias's body, as
			// Either[Box[int], Box[string]] does. An alias that reaches itself is checked once,
			// since it passes its own parameters (section 4, Type declarations).
			if s.seen == nil {
				s.seen = map[*Decl]bool{}
			}
			s.seen[x.Decl] = true
			_, err := (&substituter{m: Bindings(x.Decl.Params, args), seen: s.seen}).subst(x.Decl.Body)
			delete(s.seen, x.Decl)
			if err != nil {
				return nil, err
			}
		}
		return Named{Decl: x.Decl, Args: args}, nil
	case Union:
		ms := make([]Type, len(x.Members))
		for i, mem := range x.Members {
			n, err := s.subst(mem)
			if err != nil {
				return nil, err
			}
			ms[i] = n
		}
		return MakeUnion(ms)
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
