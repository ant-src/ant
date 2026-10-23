package types

// Assignable reports whether a value of type s can be used where t is expected (section 4,
// Assignability).
func Assignable(s, t Type) bool { return assignable(s, t, 0) }

func assignable(s, t Type, depth int) bool {
	if depth > 64 {
		return false
	}
	if IsInvalid(s) || IsInvalid(t) {
		return true
	}
	if Equal(s, t) {
		return true
	}
	s, t = Expand(s), Expand(t)
	if _, ok := s.(Never); ok {
		return true
	}
	// A union source needs every member accepted; a union target accepts any member.
	if su, ok := s.(Union); ok {
		for _, m := range su.Members {
			if !assignable(m, t, depth+1) {
				return false
			}
		}
		return true
	}
	// A type parameter value is known only to be assignable to its bound, a union bound
	// included, so the bound is tried before the members of a union target.
	if p, ok := s.(Param); ok && p.Decl.Bound != nil && assignable(p.Decl.Bound, t, depth+1) {
		return true
	}
	if tu, ok := t.(Union); ok {
		for _, m := range tu.Members {
			if assignable(s, m, depth+1) {
				return true
			}
		}
		return false
	}
	switch x := t.(type) {
	case Record:
		fields := fieldsOf(s)
		if fields == nil {
			return false
		}
		for _, want := range x.Fields {
			found := false
			for _, have := range fields {
				if have.Name == want.Name {
					found = Equal(have.Type, want.Type)
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	case Func:
		y, ok := s.(Func)
		if !ok {
			return false
		}
		// y, the value, may have defaults: it must accept every argument count t allows, down to
		// the parameters t requires.
		if x.Required < y.Required || len(x.Params) > len(y.Params) {
			return false
		}
		for i := range x.Params {
			if !assignable(x.Params[i], y.Params[i], depth+1) {
				return false
			}
		}
		return assignable(y.Return, x.Return, depth+1)
	case Param:
		// Only the parameter itself (handled by Equal) or never.
		return false
	}
	return false
}

// fieldsOf returns the fields of a record type, named or anonymous, or nil.
func fieldsOf(t Type) []Field {
	switch x := Expand(t).(type) {
	case Record:
		return x.Fields
	case Named:
		if x.Decl.Kind == RecordDecl {
			return x.FieldsOf()
		}
	case Param:
		if x.Decl.Bound != nil {
			return fieldsOf(x.Decl.Bound)
		}
	}
	return nil
}

// FieldsOf returns the fields a type exposes, or nil when it is not a record type.
func FieldsOf(t Type) []Field { return fieldsOf(t) }
