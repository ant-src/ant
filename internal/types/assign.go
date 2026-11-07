package types

import "github.com/ant-src/ant/value"

// Assignable reports whether a value of type s can be used where t is expected (section 4,
// Assignability). Equal types are tried first, once: it is the common case, and the walk below
// then compares with Equal only where it reaches a type that must be the same, so that a deep
// function type is related in linear time.
func Assignable(s, t Type) bool { return Equal(s, t) || assignable(s, t, &pairs{}, 0) }

// assignable relates two types depth levels into the walk. A recursive alias is related as
// equal compares it: a pair already under test is assumed assignable (section 4,
// Assignability). That is what ends the walk, with no depth limit: every other step moves to a
// part written inside s or t or expands an alias that does not reach itself, and an alias that
// does reaches itself only with its own parameters (section 4, Type declarations), so the types
// met are finitely many and a walk that does not end meets one pair twice. A bounded type
// parameter, which is related through its bound, is remembered the same way. A pair already
// decided is not walked again, so a function type built from aliases, as F2 = func(F1): F1
// is, is related in time linear in its aliases, not in its expansion. Like equal, the walk
// moves to a fresh stack every value.Deep levels.
func assignable(s, t Type, met *pairs, depth int) bool {
	if value.Deep(depth) {
		return assignableOnFreshStack(s, t, met, depth)
	}
	return assignableStep(s, t, met, depth)
}

func assignableOnFreshStack(s, t Type, met *pairs, depth int) (ok bool) {
	value.OnFreshStack(func() { ok = assignableStep(s, t, met, depth) })
	return ok
}

// assignableStep is assignable on the current stack.
func assignableStep(s, t Type, met *pairs, depth int) (ok bool) {
	if IsInvalid(s) || IsInvalid(t) {
		return true
	}
	if boundedParam(s) || boundedParam(t) || met.remembered(s) || met.remembered(t) {
		p, mark, holds, known := met.enter(s, t)
		if known {
			return holds
		}
		defer func() { met.leave(p, mark, ok) }()
	}
	s, t = Expand(s), Expand(t)
	if _, ok := s.(Never); ok {
		return true
	}
	// A union source needs every member accepted; a union target accepts any member.
	if su, ok := s.(Union); ok {
		for _, m := range su.Members {
			if !assignable(m, t, met, depth+1) {
				return false
			}
		}
		return true
	}
	// A type parameter value is known only to be assignable to its bound, a union bound
	// included, so the bound is tried before the members of a union target.
	if p, ok := s.(Param); ok && p.Decl.Bound != nil && assignable(p.Decl.Bound, t, met, depth+1) {
		return true
	}
	if tu, ok := t.(Union); ok {
		for _, m := range tu.Members {
			if assignable(s, m, met, depth+1) {
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
			if !assignable(x.Params[i], y.Params[i], met, depth+1) {
				return false
			}
		}
		return assignable(y.Return, x.Return, met, depth+1)
	}
	// Everything else, type parameters and the invariant generic types included, only as the
	// same type.
	return Equal(s, t)
}

// boundedParam reports whether t is a type parameter with a bound, which is related through
// the bound and so can lead back to itself, as T: Ord[T] does.
func boundedParam(t Type) bool {
	p, ok := t.(Param)
	return ok && p.Decl.Bound != nil
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
