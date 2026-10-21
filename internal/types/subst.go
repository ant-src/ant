package types

// Subst replaces type parameters by their bindings. Unions are formed again after
// substitution, which can fail where the arguments are written (section 4, Unions).
func Subst(t Type, m map[*ParamDecl]Type) (Type, error) {
	if len(m) == 0 {
		return t, nil
	}
	return subst(t, m)
}

func subst(t Type, m map[*ParamDecl]Type) (Type, error) {
	switch x := t.(type) {
	case Param:
		if b, ok := m[x.Decl]; ok {
			return b, nil
		}
		return t, nil
	case Array:
		e, err := subst(x.Elem, m)
		if err != nil {
			return nil, err
		}
		return Array{Elem: e}, nil
	case Map:
		k, err := subst(x.Key, m)
		if err != nil {
			return nil, err
		}
		v, err := subst(x.Val, m)
		if err != nil {
			return nil, err
		}
		return Map{Key: k, Val: v}, nil
	case Func:
		ps := make([]Type, len(x.Params))
		for i, p := range x.Params {
			q, err := subst(p, m)
			if err != nil {
				return nil, err
			}
			ps[i] = q
		}
		r, err := subst(x.Return, m)
		if err != nil {
			return nil, err
		}
		return Func{Params: ps, Return: r, Required: x.Required, Names: x.Names}, nil
	case Record:
		fs := make([]Field, len(x.Fields))
		for i, f := range x.Fields {
			ft, err := subst(f.Type, m)
			if err != nil {
				return nil, err
			}
			fs[i] = Field{Name: f.Name, Type: ft}
		}
		return Record{Fields: fs}, nil
	case Named:
		if len(x.Args) == 0 {
			return t, nil
		}
		args := make([]Type, len(x.Args))
		for i, a := range x.Args {
			b, err := subst(a, m)
			if err != nil {
				return nil, err
			}
			args[i] = b
		}
		return Named{Decl: x.Decl, Args: args}, nil
	case Union:
		ms := make([]Type, len(x.Members))
		for i, mem := range x.Members {
			n, err := subst(mem, m)
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
	switch x := t.(type) {
	case Param:
		return set[x.Decl]
	case Array:
		return Mentions(x.Elem, set)
	case Map:
		return Mentions(x.Key, set) || Mentions(x.Val, set)
	case Func:
		for _, p := range x.Params {
			if Mentions(p, set) {
				return true
			}
		}
		return Mentions(x.Return, set)
	case Record:
		for _, f := range x.Fields {
			if Mentions(f.Type, set) {
				return true
			}
		}
	case Named:
		for _, a := range x.Args {
			if Mentions(a, set) {
				return true
			}
		}
	case Union:
		for _, m := range x.Members {
			if Mentions(m, set) {
				return true
			}
		}
	}
	return false
}
