package types

import "fmt"

// BareBound reports the type parameter a bound is, aliases expanded, when it is one: a bound may
// not itself be a type parameter, its own or another's (section 4, Generics), since reading it
// would lead from the parameter back to a parameter with nothing else to go on, as T: T does.
func BareBound(bound Type) (Param, bool) {
	t, ok := ExpandOK(bound)
	if !ok {
		return Param{}, false
	}
	p, ok := t.(Param)
	return p, ok
}

// BareBoundError is the message for a type parameter p whose bound is the type parameter q.
func BareBoundError(p *ParamDecl, q Param) string {
	return fmt.Sprintf("the bound of %s is the type parameter %s; a bound may not be a type parameter", p.Name, q.Decl.Name)
}

// FitsBound reports whether t may stand for a type parameter whose bound, read with the other
// parameters' solutions in place, is bound (section 4, Generics): a single member of a union
// bound, or a type assignable to any other bound. A bounded type parameter fits any bound its
// own bound is assignable to, since it stands for one member of its own.
func FitsBound(t, bound Type) bool {
	if tp, ok := t.(Param); ok && tp.Decl.Bound != nil && Assignable(tp.Decl.Bound, bound) {
		return true
	}
	members := Members(bound)
	if len(members) > 1 {
		for _, m := range members {
			if Equal(m, t) {
				return true
			}
		}
		return FieldsOf(bound) != nil && Assignable(t, bound)
	}
	return Assignable(t, bound)
}

// CheckBounds checks each argument of n against its parameter's bound, read with n's arguments
// in place of the parameters it names: the bound of a generic type's parameter holds wherever the
// type is written (section 4, Generics). A bound that substituting the arguments makes invalid is
// left to the error that substitution reports elsewhere.
func CheckBounds(n Named) error {
	if len(n.Args) != len(n.Decl.Params) {
		return nil
	}
	var m map[*ParamDecl]Type
	for i, p := range n.Decl.Params {
		arg := n.Args[i]
		if p.Bound == nil || IsInvalid(arg) {
			continue
		}
		if m == nil {
			m = Bindings(n.Decl.Params, n.Args)
		}
		bound, err := Subst(p.Bound, m)
		if err != nil || FitsBound(arg, bound) {
			continue
		}
		if len(Members(bound)) > 1 {
			return fmt.Errorf("%s would be %s, which is not a single member of its bound %s", p.Name, String(arg), String(bound))
		}
		return fmt.Errorf("%s would be %s, which is not assignable to its bound %s", p.Name, String(arg), String(bound))
	}
	return nil
}

// CheckWrittenBounds applies CheckBounds to every instantiation written in t, its arguments
// before it.
func CheckWrittenBounds(t Type) error {
	var first error
	walk(t, nil, func(t Type) {
		if n, ok := t.(Named); ok && first == nil && len(n.Args) > 0 {
			if err := CheckBounds(n); err != nil {
				first = fmt.Errorf("%s: %w", String(n), err)
			}
		}
	})
	return first
}
