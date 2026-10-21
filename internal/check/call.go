package check

import (
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// call checks a call: a direct call of a declared function, which may be generic and take
// named arguments; a constructor call; or a call through a function value.
func (c *Checker) call(x *ast.Call, sc *scope, expected types.Type) types.Type {
	if x.Optional {
		fn := c.synth(x.Fn, sc)
		if !types.Includes(fn, types.Null{}) {
			c.errorf(x.Pos, "?.( needs a nullable left side, got %s", types.String(fn))
		}
		inner := c.callValue(types.Without(fn, types.Null{}), x, sc)
		u, _ := types.MakeUnion([]types.Type{inner, types.Null{}})
		return u
	}
	ent := c.entityPath(x.Fn, sc)
	if ent != nil {
		switch ent.kind {
		case entFunc:
			s := c.sigOf(ent, x.Pos)
			if s == nil {
				c.synthArgs(x.Args, sc)
				return types.Invalid{}
			}
			if ent.isConv {
				return c.conversion(ent, x, sc)
			}
			return c.applySig(s, x, sc, expected, true)
		case entType:
			if ent.decl.Kind == types.PrimitiveDecl {
				return c.construct(ent.decl, x, sc, expected)
			}
			c.errorf(x.Pos, "%s is a type, not a function", ent.decl.Name)
			c.synthArgs(x.Args, sc)
			return types.Invalid{}
		case entModule, entNamespace:
			c.errorf(x.Pos, "%s is a namespace, not a function", ent.name)
			c.synthArgs(x.Args, sc)
			return types.Invalid{}
		}
	}
	fn := c.synth(x.Fn, sc)
	return c.callValue(fn, x, sc)
}

func (c *Checker) synthArgs(args []ast.Arg, sc *scope) {
	for _, a := range args {
		c.synth(a.Value, sc)
	}
}

// callValue checks a positional call through a function value.
func (c *Checker) callValue(fn types.Type, x *ast.Call, sc *scope) types.Type {
	ft, ok := types.Expand(fn).(types.Func)
	if !ok {
		if !types.IsInvalid(fn) {
			c.errorf(x.Pos, "%s is not a function and cannot be called", types.String(fn))
		}
		c.synthArgs(x.Args, sc)
		return types.Invalid{}
	}
	for _, a := range x.Args {
		if a.Name != "" {
			c.errorf(a.Pos, "a call through a function value is positional; named arguments apply only to a declared function")
			c.synthArgs(x.Args, sc)
			return ft.Return
		}
	}
	if len(x.Args) < ft.Required || len(x.Args) > len(ft.Params) {
		c.errorf(x.Pos, "%s takes %s, got %d", types.String(fn), countParams(ft), len(x.Args))
		c.synthArgs(x.Args, sc)
		return ft.Return
	}
	for i, a := range x.Args {
		c.check(a.Value, sc, ft.Params[i])
	}
	return ft.Return
}

func countParams(ft types.Func) string {
	n := len(ft.Params)
	if ft.Required != n {
		return plural(ft.Required, "argument") + " to " + plural(n, "argument")
	}
	return plural(n, "argument")
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return itoa(n) + " " + noun + "s"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

// construct checks a named primitive type's constructor call.
func (c *Checker) construct(decl *types.Decl, x *ast.Call, sc *scope, expected types.Type) types.Type {
	if len(x.Args) != 1 || x.Args[0].Name != "" {
		c.errorf(x.Pos, "the constructor %s takes exactly one positional argument", decl.Name)
		c.synthArgs(x.Args, sc)
	} else {
		c.check(x.Args[0].Value, sc, types.Prim{Kind: decl.Base})
	}
	if len(decl.Params) == 0 {
		return types.Named{Decl: decl}
	}
	// type Id[T] = string: the body does not mention T, so only an expected type fixes it.
	if expected != nil {
		for _, m := range types.Members(expected) {
			if n, ok := types.Expand(m).(types.Named); ok && n.Decl == decl && len(n.Args) == len(decl.Params) {
				return n
			}
		}
	}
	var args []types.Type
	for _, p := range decl.Params {
		c.errorf(x.Pos, "cannot infer the type parameter %s of %s; annotate the result", p.Name, decl.Name)
		args = append(args, types.Invalid{})
	}
	return types.Named{Decl: decl, Args: args}
}

// conversion checks one of the twelve conversion functions: the argument must be a number, or
// a string or bool for those two, or a named primitive over such a base, which is unwrapped.
func (c *Checker) conversion(ent *entity, x *ast.Call, sc *scope) types.Type {
	result := types.Prim{Kind: ent.conversion}
	if len(x.Args) != 1 || x.Args[0].Name != "" {
		c.errorf(x.Pos, "%s takes exactly one positional argument", ent.name)
		c.synthArgs(x.Args, sc)
		return result
	}
	t := c.synth(x.Args[0].Value, sc)
	base := t
	if n, ok := types.Expand(t).(types.Named); ok && n.Decl.Kind == types.PrimitiveDecl {
		base = types.Prim{Kind: n.Decl.Base}
	}
	p, ok := types.Expand(base).(types.Prim)
	if types.IsInvalid(base) {
		return result
	}
	switch {
	case ent.conversion == value.KindString || ent.conversion == value.KindBool:
		if !ok || p.Kind != ent.conversion {
			c.errorf(x.Pos, "%s converts only a %s or a named primitive type over %s, got %s", ent.name, ent.conversion, ent.conversion, types.String(t))
		}
	default:
		if !ok || !p.Kind.IsNumber() {
			if pp, isParam := types.Expand(base).(types.Param); isParam && pp.Decl.Bound != nil && c.operandOK(base, "+", false) {
				return result
			}
			c.errorf(x.Pos, "%s converts a number or a named primitive type over a number, got %s", ent.name, types.String(t))
		}
	}
	return result
}

// applySig checks a call of a declared function, solving its type parameters (section 4,
// Inference) and matching positional and named arguments to its parameters.
func (c *Checker) applySig(s *sig, x *ast.Call, sc *scope, expected types.Type, named bool) types.Type {
	n := len(s.fn.Params)
	// map arguments to parameters
	slots := make([]*ast.Arg, n)
	ok := true
	positional := 0
	for i := range x.Args {
		a := &x.Args[i]
		if a.Name == "" {
			if positional >= n {
				c.errorf(a.Pos, "%s takes %s, got more", s.fullName(), plural(n, "argument"))
				ok = false
				break
			}
			slots[positional] = a
			positional++
			continue
		}
		idx := -1
		for j, name := range s.fn.Names {
			if name == a.Name {
				idx = j
				break
			}
		}
		if idx < 0 {
			c.errorf(a.Pos, "%s has no parameter %s", s.fullName(), a.Name)
			ok = false
			continue
		}
		if slots[idx] != nil {
			c.errorf(a.Pos, "the parameter %s is given twice", a.Name)
			ok = false
			continue
		}
		slots[idx] = a
	}
	if !ok {
		c.synthArgs(x.Args, sc)
		return c.resultOf(s, nil)
	}
	for i := range slots {
		if slots[i] == nil && !(i < len(s.defaults) && s.defaults[i]) {
			c.errorf(x.Pos, "%s needs an argument for %s", s.fullName(), s.fn.Names[i])
			ok = false
		}
	}
	if !ok {
		c.synthArgs(x.Args, sc)
		return c.resultOf(s, nil)
	}
	if len(s.tparams) == 0 {
		for i, a := range slots {
			if a != nil {
				c.check(a.Value, sc, s.fn.Params[i])
			}
		}
		return s.fn.Return
	}
	// generic: solve from the expected type, then non-lambda arguments, then lambdas
	subst := map[*types.ParamDecl]types.Type{}
	unsolved := map[*types.ParamDecl]bool{}
	for _, p := range s.tparams {
		unsolved[p] = true
	}
	if expected != nil && types.Mentions(s.fn.Return, unsolved) {
		trial := map[*types.ParamDecl]types.Type{}
		if c.unify(s.fn.Return, expected, trial) {
			for p, t := range trial {
				subst[p] = t
				delete(unsolved, p)
			}
		}
	}
	type pending struct {
		i int
		t types.Type
	}
	var synthesized []pending
	for i, a := range slots {
		if a == nil {
			continue
		}
		if _, isLambda := a.Value.(*ast.Lambda); isLambda {
			continue
		}
		pt := s.fn.Params[i]
		if types.Mentions(pt, unsolved) {
			got := c.synthWithPattern(a.Value, pt, subst, unsolved, sc)
			if !c.unify(pt, got, subst) {
				want, _ := types.Subst(pt, subst)
				c.errorf(a.Pos, "cannot use %s where %s is expected", types.String(got), types.String(want))
			}
			for p := range unsolved {
				if _, done := subst[p]; done {
					delete(unsolved, p)
				}
			}
			synthesized = append(synthesized, pending{i, got})
			continue
		}
		want, err := types.Subst(pt, subst)
		if err != nil {
			c.errorf(a.Pos, "%s", err)
			continue
		}
		c.check(a.Value, sc, want)
	}
	for i, a := range slots {
		if a == nil {
			continue
		}
		lam, isLambda := a.Value.(*ast.Lambda)
		if !isLambda {
			continue
		}
		pt, err := types.Subst(s.fn.Params[i], subst)
		if err != nil {
			c.errorf(a.Pos, "%s", err)
			continue
		}
		ft, isFunc := types.Expand(pt).(types.Func)
		if !isFunc {
			c.errorf(a.Pos, "a lambda needs an expected function type, but %s is expected", types.String(pt))
			c.lambdaBodyAnyway(lam, sc)
			continue
		}
		if types.Mentions(ft, unsolved) && anyParamMentions(ft.Params, unsolved) {
			c.errorf(a.Pos, "the lambda's parameter type %s is not fixed yet; pass the other arguments first or annotate", types.String(ft.Params[0]))
			c.lambdaBodyAnyway(lam, sc)
			continue
		}
		if len(ft.Params) > 1 {
			c.errorf(a.Pos, "a lambda has at most one parameter, x; %s expects %d: write func with named parameters", types.String(ft), len(ft.Params))
			c.lambdaBodyAnyway(lam, sc)
			continue
		}
		got := c.lambda(lam, sc, ft)
		if g, isF := got.(types.Func); isF {
			if !c.unify(ft.Return, g.Return, subst) {
				want, _ := types.Subst(ft.Return, subst)
				c.errorf(a.Pos, "the lambda returns %s where %s is expected", types.String(g.Return), types.String(want))
			}
			for p := range unsolved {
				if _, done := subst[p]; done {
					delete(unsolved, p)
				}
			}
		}
	}
	for _, p := range s.tparams {
		if _, done := subst[p]; !done {
			c.errorf(x.Pos, "cannot infer the type parameter %s of %s; annotate the result", p.Name, s.fullName())
			subst[p] = types.Invalid{}
		}
	}
	c.checkBounds(s, subst, x.Pos)
	for _, pe := range synthesized {
		want, err := types.Subst(s.fn.Params[pe.i], subst)
		if err != nil {
			c.errorf(slots[pe.i].Pos, "%s", err)
			continue
		}
		if !types.Assignable(pe.t, want) {
			c.errorf(slots[pe.i].Pos, "cannot use %s where %s is expected", types.String(pe.t), types.String(want))
		}
	}
	return c.resultOf(s, subst)
}

func anyParamMentions(ps []types.Type, set map[*types.ParamDecl]bool) bool {
	for _, p := range ps {
		if types.Mentions(p, set) {
			return true
		}
	}
	return false
}

func (c *Checker) resultOf(s *sig, subst map[*types.ParamDecl]types.Type) types.Type {
	if s.fn.Return == nil {
		return types.Invalid{}
	}
	if subst == nil {
		if len(s.tparams) > 0 {
			return types.Invalid{}
		}
		return s.fn.Return
	}
	t, err := types.Subst(s.fn.Return, subst)
	if err != nil {
		return types.Invalid{}
	}
	return t
}

// checkBounds checks each solution against its bound: assignable to it, and when the bound is
// a union, one member of it rather than a union.
func (c *Checker) checkBounds(s *sig, subst map[*types.ParamDecl]types.Type, pos ast.Pos) {
	for _, p := range s.tparams {
		if p.Bound == nil {
			continue
		}
		t := subst[p]
		if types.IsInvalid(t) {
			continue
		}
		if tp, ok := t.(types.Param); ok && tp.Decl.Bound != nil && types.Assignable(tp.Decl.Bound, p.Bound) {
			// A bounded parameter stands for one member of its bound, so it satisfies any bound
			// that contains its own (section 4, Generics).
			continue
		}
		members := types.Members(p.Bound)
		if len(members) > 1 {
			single := false
			for _, m := range members {
				if types.Equal(m, t) {
					single = true
					break
				}
			}
			if !single {
				if fields := types.FieldsOf(p.Bound); fields != nil && types.Assignable(t, p.Bound) {
					continue
				}
				c.errorf(pos, "%s would be %s, which is not a single member of its bound %s; narrow first", p.Name, types.String(t), types.String(p.Bound))
			}
			continue
		}
		if !types.Assignable(t, p.Bound) {
			c.errorf(pos, "%s would be %s, which is not assignable to its bound %s", p.Name, types.String(t), types.String(p.Bound))
		}
	}
}

// unify matches a pattern that may mention type parameters against an actual type, binding
// the parameters it meets. A parameter already bound must match by equality.
func (c *Checker) unify(pattern, actual types.Type, subst map[*types.ParamDecl]types.Type) bool {
	if types.IsInvalid(actual) {
		bindInvalid(pattern, subst)
		return true
	}
	if types.IsInvalid(pattern) {
		return true
	}
	if p, ok := pattern.(types.Param); ok {
		if bound, done := subst[p.Decl]; done {
			if types.IsNever(bound) {
				// bound by a never argument: any other argument fixes it for real
				subst[p.Decl] = actual
				return true
			}
			return types.Equal(bound, actual) || types.Assignable(actual, bound)
		}
		subst[p.Decl] = actual
		return true
	}
	switch pt := pattern.(type) {
	case types.Named:
		if an, ok := actual.(types.Named); ok && an.Decl == pt.Decl {
			for i := range pt.Args {
				if i >= len(an.Args) || !c.unify(pt.Args[i], an.Args[i], subst) {
					return false
				}
			}
			return true
		}
		if pt.Decl.Kind == types.AliasDecl {
			return c.unify(types.Expand(pattern), actual, subst)
		}
		return types.Equal(pattern, actual)
	case types.Array:
		a, ok := types.Expand(actual).(types.Array)
		return ok && c.unify(pt.Elem, a.Elem, subst)
	case types.Map:
		a, ok := types.Expand(actual).(types.Map)
		return ok && c.unify(pt.Key, a.Key, subst) && c.unify(pt.Val, a.Val, subst)
	case types.Func:
		a, ok := types.Expand(actual).(types.Func)
		if !ok || len(a.Params) < pt.Required || len(a.Params) > len(pt.Params) && len(pt.Params) < a.Required {
			return false
		}
		if len(a.Params) != len(pt.Params) {
			if len(pt.Params) < a.Required || len(pt.Params) > len(a.Params) {
				return false
			}
		}
		for i := range pt.Params {
			if i >= len(a.Params) {
				break
			}
			if !c.unify(pt.Params[i], a.Params[i], subst) {
				return false
			}
		}
		return c.unify(pt.Return, a.Return, subst)
	case types.Record:
		fields := types.FieldsOf(actual)
		if fields == nil {
			return false
		}
		for _, want := range pt.Fields {
			found := false
			for _, have := range fields {
				if have.Name == want.Name {
					if !c.unify(want.Type, have.Type, subst) {
						return false
					}
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	case types.Union:
		return c.unifyUnion(pt, actual, subst)
	}
	if an, ok := types.Expand(actual).(types.Named); ok && an.Decl.Kind == types.AliasDecl {
		return c.unify(pattern, types.Expand(actual), subst)
	}
	return types.Assignable(actual, pattern)
}

// unifyUnion matches the members of a union pattern against the actual type's members by
// discriminator; a type parameter member absorbs whatever is left.
func (c *Checker) unifyUnion(pt types.Union, actual types.Type, subst map[*types.ParamDecl]types.Type) bool {
	remaining := append([]types.Type(nil), types.Members(actual)...)
	var params []types.Param
	for _, pm := range pt.Members {
		if p, ok := pm.(types.Param); ok {
			params = append(params, p)
			continue
		}
		matched := false
		for i, am := range remaining {
			if sameDiscriminator(pm, am) {
				if !c.unify(pm, am, subst) {
					return false
				}
				remaining = append(remaining[:i], remaining[i+1:]...)
				matched = true
				break
			}
		}
		_ = matched
	}
	if len(params) == 0 {
		return len(remaining) == 0
	}
	if len(params) > 1 && len(remaining) > 0 {
		// Either[A, B] against a union: pair them in order
		for i, p := range params {
			if i < len(remaining) {
				if !c.unify(p, remaining[i], subst) {
					return false
				}
			}
		}
		return true
	}
	rest, err := types.MakeUnion(remaining)
	if err != nil {
		return false
	}
	if _, isNever := rest.(types.Never); isNever {
		if _, done := subst[params[0].Decl]; done {
			return true
		}
		return true
	}
	return c.unify(params[0], rest, subst)
}

func sameDiscriminator(a, b types.Type) bool {
	ka, _, oka := discriminatorOf(a)
	kb, _, okb := discriminatorOf(b)
	return oka && okb && ka == kb
}

func discriminatorOf(t types.Type) (string, string, bool) {
	switch x := types.Expand(t).(type) {
	case types.Prim:
		return "kind:" + x.Kind.String(), "", true
	case types.Null:
		return "null", "", true
	case types.Void:
		return "void", "", true
	case types.Array:
		return "array", "", true
	case types.Map:
		return "map", "", true
	case types.Func:
		return "func", "", true
	case types.Named:
		return "tag:" + x.Decl.Name + itoa(len(x.Decl.Fields)), "", true
	case types.Record:
		return "record", "", true
	}
	return "", "", false
}

// synthWithPattern synthesizes an argument whose parameter type still mentions unsolved type
// parameters. An array or map literal is taken apart along the pattern, so that an empty
// literal inside takes its type from the pattern and a nested literal is checked against the
// solved part.
func (c *Checker) synthWithPattern(e ast.Expr, pattern types.Type, subst map[*types.ParamDecl]types.Type, unsolved map[*types.ParamDecl]bool, sc *scope) types.Type {
	switch x := e.(type) {
	case *ast.Ident, *ast.Field:
		if t := c.genericArg(x, pattern, subst, unsolved, sc); t != nil {
			return t
		}
	case *ast.ArrayLit:
		if a, ok := types.Expand(pattern).(types.Array); ok {
			var elems []types.Type
			var empties []ast.Expr
			for _, el := range x.Elems {
				if el.Spread {
					t := c.synthWithPattern(el.Value, pattern, subst, unsolved, sc)
					if arr, ok := types.Expand(t).(types.Array); ok {
						elems = append(elems, arr.Elem)
					}
					continue
				}
				if isEmptyLiteral(el.Value) {
					empties = append(empties, el.Value)
					continue
				}
				elems = append(elems, c.synthWithPattern(el.Value, a.Elem, subst, unsolved, sc))
			}
			if len(elems) == 0 {
				solved, err := types.Subst(a.Elem, subst)
				if err != nil || types.Mentions(solved, unsolved) {
					return types.Array{Elem: types.Never{}}
				}
				for _, e := range empties {
					c.check(e, sc, solved)
				}
				return types.Array{Elem: solved}
			}
			u, err := types.MakeUnion(elems)
			if err != nil {
				c.errorf(x.Pos, "the elements do not form a valid type: %s", err)
				return types.Invalid{}
			}
			for _, e := range empties {
				c.check(e, sc, u)
			}
			return types.Array{Elem: u}
		}
	case *ast.MapLit:
		if m, ok := types.Expand(pattern).(types.Map); ok {
			var keys, vals []types.Type
			for _, en := range x.Entries {
				keys = append(keys, c.synthWithPattern(en.Key, m.Key, subst, unsolved, sc))
				vals = append(vals, c.synthWithPattern(en.Value, m.Val, subst, unsolved, sc))
			}
			if len(keys) == 0 {
				k, errK := types.Subst(m.Key, subst)
				v, errV := types.Subst(m.Val, subst)
				if errK != nil || errV != nil || types.Mentions(k, unsolved) || types.Mentions(v, unsolved) {
					return types.Map{Key: types.Never{}, Val: types.Never{}}
				}
				return types.Map{Key: k, Val: v}
			}
			k, errK := types.MakeUnion(keys)
			v, errV := types.MakeUnion(vals)
			if errK != nil || errV != nil {
				c.errorf(x.Pos, "the entries do not form a valid type")
				return types.Invalid{}
			}
			return types.Map{Key: k, Val: v}
		}
	case *ast.RecordLit:
		expected, err := types.Subst(pattern, subst)
		if err == nil && !types.Mentions(expected, unsolved) {
			return c.recordLit(x, sc, expected)
		}
		return c.recordLit(x, sc, nil)
	case *ast.Lambda:
		return c.synth(e, sc)
	}
	solved, err := types.Subst(pattern, subst)
	if err == nil && !types.Mentions(solved, unsolved) {
		return c.check(e, sc, solved)
	}
	return c.synth(e, sc)
}

// bindInvalid binds every unsolved type parameter a pattern mentions to the invalid type.
func bindInvalid(pattern types.Type, subst map[*types.ParamDecl]types.Type) {
	switch x := pattern.(type) {
	case types.Param:
		if _, done := subst[x.Decl]; !done {
			subst[x.Decl] = types.Invalid{}
		}
	case types.Array:
		bindInvalid(x.Elem, subst)
	case types.Map:
		bindInvalid(x.Key, subst)
		bindInvalid(x.Val, subst)
	case types.Func:
		for _, p := range x.Params {
			bindInvalid(p, subst)
		}
		bindInvalid(x.Return, subst)
	case types.Record:
		for _, f := range x.Fields {
			bindInvalid(f.Type, subst)
		}
	case types.Named:
		for _, a := range x.Args {
			bindInvalid(a, subst)
		}
	case types.Union:
		for _, m := range x.Members {
			bindInvalid(m, subst)
		}
	}
}

// genericArg instantiates a generic function passed as an argument from the parts of the
// parameter's pattern the call has solved so far: in Array.map(xs, byte) the element type fixes
// byte's own type parameter, and the call then solves its U from byte's result. It returns nil
// when e is not a generic function, no function type is expected, or too little is known yet,
// and the plain path reports the problem.
func (c *Checker) genericArg(e ast.Expr, pattern types.Type, subst map[*types.ParamDecl]types.Type, unsolved map[*types.ParamDecl]bool, sc *scope) types.Type {
	pat, err := types.Subst(pattern, subst)
	if err != nil {
		return nil
	}
	pf, ok := types.Expand(pat).(types.Func)
	if !ok {
		return nil
	}
	ent := c.entityPath(e, sc)
	if ent == nil || ent.kind != entFunc {
		return nil
	}
	s := c.sigOf(ent, e.Position())
	if s == nil || len(s.tparams) == 0 {
		return nil
	}
	argSubst := map[*types.ParamDecl]types.Type{}
	for i := range pf.Params {
		if i >= len(s.fn.Params) {
			break
		}
		if types.Mentions(pf.Params[i], unsolved) {
			continue
		}
		if !c.unify(s.fn.Params[i], pf.Params[i], argSubst) {
			return nil
		}
	}
	if !types.Mentions(pf.Return, unsolved) {
		c.unify(s.fn.Return, pf.Return, argSubst)
	}
	for _, p := range s.tparams {
		if _, bound := argSubst[p]; !bound {
			return nil
		}
	}
	c.checkBounds(s, argSubst, e.Position())
	inst, err := types.Subst(s.fn, argSubst)
	if err != nil {
		return nil
	}
	return inst
}
