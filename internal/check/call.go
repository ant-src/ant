package check

import (
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// call checks a call: a direct call of a declared function, which may be generic and take
// named arguments; a constructor call; or a call through a function value.
func (c *Checker) call(x *ast.Call, sc *scope, expected types.Type) types.Type {
	if x.Optional {
		fn, short := c.chainRecv(x.Fn, sc)
		if !types.Includes(fn, types.Null{}) {
			c.errorf(x.Pos, "?.( needs a nullable left side, got %s%s", types.String(fn), chainNote(short, "("))
		}
		inner := c.callValue(types.Without(fn, types.Null{}), x, sc)
		return c.endLink(x, x.Grouped, inner, true)
	}
	ent := c.entityPath(x.Fn, sc)
	if ent != nil {
		c.notePath(x.Fn, sc)
		switch ent.kind {
		case entFunc:
			s := c.sigOf(ent, x.Pos)
			if s == nil {
				c.synthArgs(x.Args, sc)
				return types.Invalid{}
			}
			if ent.isConv || ent.isLen || ent.isPrintln && len(x.Args) == 0 {
				c.info.recordCall(x, &CallInfo{Fn: s.fn, Names: s.fn.Names, Defaults: s.defaults, Callee: c.info.object(ent), Slots: positional(x)})
			}
			if ent.isConv {
				return c.conversion(ent, x, sc)
			}
			if ent.isLen {
				return c.length(x, sc)
			}
			if ent.isPrintln && len(x.Args) == 0 {
				return types.Void{}
			}
			if ent.isAssert && x.Assert != nil {
				return c.assertCall(s, x, sc, c.info.object(ent))
			}
			return c.applySig(s, x, sc, expected, c.info.object(ent), nil)
		case entType:
			if ent.decl.Kind == types.PrimitiveDecl {
				c.info.recordCall(x, &CallInfo{Fn: constructorType(ent.decl), Callee: c.info.object(ent), Slots: positional(x)})
				return c.construct(ent.decl, x, sc, expected)
			}
			c.errorf(x.Pos, "%s is a type, not a function", ent.decl.Name)
			c.synthArgs(x.Args, sc)
			return types.Invalid{}
		case entModule, entNamespace:
			c.errorf(x.Pos, "%s is a namespace, not a function", ent.name)
			c.synthArgs(x.Args, sc)
			return types.Invalid{}
		case entMissing:
			c.synthArgs(x.Args, sc)
			return types.Invalid{}
		}
	}
	if f, ok := x.Fn.(*ast.Field); ok {
		if t, done := c.dotCall(x, f, sc, expected); done {
			return t
		}
	}
	fn, short := c.chainRecv(x.Fn, sc)
	return c.endLink(x, x.Grouped, c.callValue(fn, x, sc), short)
}

func (c *Checker) synthArgs(args []ast.Arg, sc *scope) {
	for _, a := range args {
		// The call failed, so a lambda has no parameter type to take; that is the call's error,
		// and the lambda's body is checked on its own.
		if lam, ok := a.Value.(*ast.Lambda); ok {
			c.lambdaBodyAnyway(lam, sc)
			continue
		}
		c.synth(a.Value, sc)
	}
}

// positional maps each argument of a call to the parameter of its position.
func positional(x *ast.Call) []int {
	slots := make([]int, len(x.Args))
	for i := range slots {
		slots[i] = i
	}
	return slots
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
	c.info.recordCall(x, &CallInfo{Fn: ft, Slots: positional(x)})
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

// arity says how many arguments a call passes after the first skip parameters: "2 arguments",
// or "1 to 2 arguments" when some have defaults.
func (s *sig) arity(skip int) string {
	hi := len(s.fn.Params) - skip
	lo := hi
	for i := skip; i < len(s.fn.Params); i++ {
		if i < len(s.defaults) && s.defaults[i] || s.noArgs {
			lo--
		}
	}
	if lo == hi {
		return plural(hi, "argument")
	}
	return fmt.Sprintf("%d to %d arguments", lo, hi)
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
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
	name := types.String(result) // int for int64, byte for uint8, as everywhere (section 4)
	if len(x.Args) != 1 || x.Args[0].Name != "" {
		c.errorf(x.Pos, "%s takes exactly one positional argument", name)
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
			c.errorf(x.Pos, "%s converts only a %s or a named primitive type over %s, got %s", name, ent.conversion, ent.conversion, types.String(t))
		}
	default:
		if !ok || !p.Kind.IsNumber() {
			if pp, isParam := types.Expand(base).(types.Param); isParam && pp.Decl.Bound != nil && c.operandOK(base, false) {
				return result
			}
			c.errorf(x.Pos, "%s converts a number or a named primitive type over a number, got %s", name, types.String(t))
		}
	}
	return result
}

// length checks len: one positional argument that has a size (section 11, Prelude).
func (c *Checker) length(x *ast.Call, sc *scope) types.Type {
	if len(x.Args) != 1 || x.Args[0].Name != "" {
		c.errorf(x.Pos, "len takes exactly one positional argument")
		c.synthArgs(x.Args, sc)
		return types.IntType
	}
	t := c.synth(x.Args[0].Value, sc)
	if !c.sized(t) {
		c.errorf(x.Pos, "len needs an array, a map, a set, a string, Bytes, or a channel, got %s", types.String(t))
	}
	return types.IntType
}

// sized reports whether len takes a value of type t: an array, a map, a set, a string, Bytes, a
// channel, a named primitive over string, or a union or bounded type parameter of these.
func (c *Checker) sized(t types.Type) bool {
	switch x := types.Expand(t).(type) {
	case types.Invalid, types.Array, types.Map, types.Set:
		return true
	case types.Prim:
		return x.Kind == value.KindString
	case types.Named:
		if c.isBytes(x) {
			return true
		}
		if x.Decl.Kind == types.PrimitiveDecl {
			return x.Decl.Base == value.KindString
		}
		ch, ok := c.prelude["Channel"]
		return ok && x.Decl == ch.decl
	case types.Union:
		for _, m := range x.Members {
			if !c.sized(m) {
				return false
			}
		}
		return true
	case types.Param:
		return x.Decl.Bound != nil && c.sized(x.Decl.Bound)
	}
	return false
}

// lenValue types len used as a value from the function type expected: its one parameter must
// be a type len takes. It returns nil when ent is not len.
func (c *Checker) lenValue(ent *entity, f types.Func, pos ast.Pos) types.Type {
	if !ent.isLen {
		return nil
	}
	if len(f.Params) != 1 || !c.sized(f.Params[0]) {
		c.errorf(pos, "cannot use len where %s is expected", types.String(f))
		return types.Invalid{}
	}
	return types.Func{Params: []types.Type{f.Params[0]}, Return: types.IntType, Required: 1}
}

// applySig checks a call of a declared function, solving its type parameters (section 4,
// Inference) and matching positional and named arguments to its parameters. A call through a
// dot passes its receiver, typed already, as the first argument (section 5, Calls through .).
func (c *Checker) applySig(s *sig, x *ast.Call, sc *scope, expected types.Type, callee *Object, recv *receiver) types.Type {
	slots, ok := c.argSlots(s, x, recv)
	c.recordSlots(s, x, slots, ok, callee, recv != nil)
	if ok && recv.untyped() {
		// checked as the first argument of the direct call
		slots[0] = &ast.Arg{Pos: recv.pos, Value: recv.expr}
	}
	if !ok {
		if recv.untyped() {
			c.synth(recv.expr, sc)
		}
		c.synthArgs(x.Args, sc)
		return c.resultOf(s, nil, x.Pos)
	}
	if len(s.tparams) == 0 {
		if recv != nil && !recv.untyped() {
			c.expect(recv.pos, recv.typ, s.fn.Params[0])
		}
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
	// misfit: the result can be no instance of the expected type, whatever the arguments solve.
	misfit := false
	if expected != nil && types.Mentions(s.fn.Return, unsolved) {
		trial := map[*types.ParamDecl]types.Type{}
		misfit = true
		if c.unify(s.fn.Return, expectedPart(s.fn.Return, expected), trial) {
			misfit = false
			for p, t := range trial {
				if !boundFits(p, t, trial, s.tparams) {
					// An expected union such as Time.Dated is no solution for a bounded T,
					// which stands for one member; the arguments solve it instead.
					continue
				}
				subst[p] = t
				delete(unsolved, p)
			}
		}
	}
	type pending struct {
		i   int
		t   types.Type
		pos ast.Pos
	}
	var synthesized []pending
	mismatch := false
	// An argument that fixes nothing, an empty literal or a call of only those, waits for the
	// other arguments; a receiver can be one too.
	type waiting struct {
		i int
		e ast.Expr
	}
	var empties []waiting
	if recv != nil && !recv.untyped() {
		// the receiver, typed already, solves what it mentions first
		pt := s.fn.Params[0]
		if types.Mentions(pt, unsolved) {
			if c.unify(pt, recv.typ, subst) {
				synthesized = append(synthesized, pending{0, recv.typ, recv.pos})
			} else {
				want, _ := types.Subst(pt, subst)
				n := c.typeNames(recv.typ, want)
				c.errorf(recv.pos, "cannot use %s where %s is expected", n[0], n[1])
				mismatch = true
			}
			pruneSolved(unsolved, subst)
		} else if want, err := types.Subst(pt, subst); err != nil {
			c.errorf(recv.pos, "%s", err)
		} else {
			c.expect(recv.pos, recv.typ, want)
		}
	}
	for i, a := range slots {
		if a == nil {
			continue
		}
		if _, isLambda := a.Value.(*ast.Lambda); isLambda {
			continue
		}
		pt := s.fn.Params[i]
		if types.Mentions(pt, unsolved) && onlyEmpty(a.Value) {
			empties = append(empties, waiting{i, a.Value})
			continue
		}
		if types.Mentions(pt, unsolved) {
			got := c.synthWithPattern(a.Value, pt, subst, unsolved, sc)
			if !c.unify(pt, got, subst) {
				want, _ := types.Subst(pt, subst)
				n := c.typeNames(got, want)
				c.errorf(a.Pos, "cannot use %s where %s is expected", n[0], n[1])
				mismatch = true
				pruneSolved(unsolved, subst)
				continue // reported once, not again against the solved types
			}
			pruneSolved(unsolved, subst)
			synthesized = append(synthesized, pending{i, got, a.Pos})
			continue
		}
		want, err := types.Subst(pt, subst)
		if err != nil {
			c.errorf(a.Pos, "%s", err)
			continue
		}
		c.check(a.Value, sc, want)
	}
	if hint, ok := c.hints[x]; ok && len(unsolved) > 0 {
		c.solveFromHint(s, hint, subst, unsolved)
	}
	for _, w := range empties {
		want, err := types.Subst(s.fn.Params[w.i], subst)
		if err != nil {
			c.errorf(w.e.Position(), "%s", err)
			continue
		}
		if types.Mentions(want, unsolved) {
			// Nothing else fixed it. A call is typed on its own: ints([]) has a type, and
			// Set.of([]) reports the empty literal inside it.
			if _, isCall := w.e.(*ast.Call); isCall && !mismatch {
				got := c.synth(w.e, sc)
				if !types.IsInvalid(got) && c.unify(s.fn.Params[w.i], got, subst) {
					pruneSolved(unsolved, subst)
					synthesized = append(synthesized, pending{w.i, got, w.e.Position()})
					continue
				}
			} else if !mismatch {
				c.errorf(w.e.Position(), "%s", emptyError(w.e))
			}
			bindInvalid(s.fn.Params[w.i], subst)
			pruneSolved(unsolved, subst)
			continue
		}
		c.check(w.e, sc, want)
	}
	if mismatch {
		// An argument that does not fit its parameter already explains every type parameter
		// still open, so those are left unsolved silently instead of reported as uninferable.
		for _, p := range s.tparams {
			if _, done := subst[p]; !done {
				subst[p] = types.Invalid{}
			}
		}
		pruneSolved(unsolved, subst)
	}
	for i, a := range slots {
		if a == nil {
			continue
		}
		if lam, isLambda := a.Value.(*ast.Lambda); isLambda {
			c.lambdaArg(a, lam, s.fn.Params[i], sc, subst, unsolved)
		}
	}
	misfitSaid := false
	for _, p := range s.tparams {
		if _, done := subst[p]; !done {
			if !misfit {
				c.errorf(x.Pos, "cannot infer the type parameter %s of %s; annotate the result", p.Name, s.fullName())
			} else if !misfitSaid {
				// Only the expected type could solve p, and the result does not fit it: an
				// annotation would not help, so name the two types, once.
				c.errorf(x.Pos, "cannot use %s where %s is expected: it returns %s", s.fullName(), types.String(expected), types.String(s.fn.Return))
				misfitSaid = true
			}
			subst[p] = types.Invalid{}
		}
	}
	c.checkBounds(s, subst, x.Pos)
	// An argument synthesized while its parameter's type was open is checked against the
	// solved type.
	for _, pe := range synthesized {
		want, err := types.Subst(s.fn.Params[pe.i], subst)
		if err != nil {
			c.errorf(pe.pos, "%s", err)
			continue
		}
		if !types.Assignable(pe.t, want) {
			n := c.typeNames(pe.t, want)
			c.errorf(pe.pos, "cannot use %s where %s is expected", n[0], n[1])
		}
	}
	return c.resultOf(s, subst, x.Pos)
}

// argSlotsLoosely maps each argument of a call that argSlots refused to the parameter it would
// fill: a positional one by its position, after the receiver when the call has one, a named one
// by its name, or -1.
func argSlotsLoosely(s *sig, x *ast.Call, recv bool) []int {
	out := make([]int, len(x.Args))
	positional := 0
	if recv {
		positional = 1
	}
	for i, a := range x.Args {
		out[i] = -1
		if a.Name == "" {
			if positional < len(s.fn.Params) {
				out[i] = positional
			}
			positional++
			continue
		}
		for j, name := range s.fn.Names {
			if name == a.Name {
				out[i] = j
			}
		}
	}
	return out
}

// argSlots matches a call's positional and named arguments to the function's parameters, one
// slot per parameter and nil where a default applies. It reports an extra, unknown, repeated
// or missing argument, and then the second result is false. A receiver fills the first
// parameter, which the slots leave nil, and makes a call through an interface, a type
// parameter or a union of several namespaces positional (section 5, Calls through .).
func (c *Checker) argSlots(s *sig, x *ast.Call, recv *receiver) ([]*ast.Arg, bool) {
	n := len(s.fn.Params)
	slots := make([]*ast.Arg, n)
	ok := true
	positional := 0
	first := 0
	if recv != nil {
		positional, first = 1, 1
	}
	for i := range x.Args {
		a := &x.Args[i]
		if a.Name == "" {
			if positional >= n {
				got := 0
				for _, b := range x.Args {
					if b.Name == "" {
						got++
					}
				}
				if recv != nil {
					c.errorf(a.Pos, "%s takes %s after the receiver, got %d", s.fullName(), s.arity(first), got)
				} else {
					c.errorf(a.Pos, "%s takes %s, got %d", s.fullName(), s.arity(first), got)
				}
				ok = false
				break
			}
			slots[positional] = a
			positional++
			continue
		}
		if recv != nil && recv.positional {
			c.errorf(a.Pos, "a call through %s is positional, since each implementation names its parameters as it likes", types.String(recv.typ))
			ok = false
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
		if idx < first {
			c.errorf(a.Pos, "the parameter %s is the receiver, which the call passes first", a.Name)
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
		return nil, false
	}
	for i := first; i < len(slots); i++ {
		if slots[i] == nil && !(i < len(s.defaults) && s.defaults[i]) {
			c.errorf(x.Pos, "%s needs an argument for %s", s.fullName(), s.fn.Names[i])
			ok = false
		}
	}
	return slots, ok
}

// lambdaArg checks a lambda argument against its parameter's type pt, once the other arguments
// have solved what they can, and solves more type parameters from what the lambda returns.
func (c *Checker) lambdaArg(a *ast.Arg, lam *ast.Lambda, pt types.Type, sc *scope, subst map[*types.ParamDecl]types.Type, unsolved map[*types.ParamDecl]bool) {
	pt, err := types.Subst(pt, subst)
	if err != nil {
		c.errorf(a.Pos, "%s", err)
		return
	}
	ft, isFunc := types.Expand(pt).(types.Func)
	if !isFunc {
		c.errorf(a.Pos, "a lambda needs an expected function type, but %s is expected", types.String(pt))
		c.lambdaBodyAnyway(lam, sc)
		return
	}
	if types.Mentions(ft, unsolved) && anyParamMentions(ft.Params, unsolved) {
		c.errorf(a.Pos, "the lambda's parameter type %s is not fixed yet; pass the other arguments first or annotate", types.String(ft.Params[0]))
		c.lambdaBodyAnyway(lam, sc)
		return
	}
	if len(ft.Params) > 1 {
		c.errorf(a.Pos, "a lambda has at most one parameter, x; %s expects %d: write func with named parameters", types.String(ft), len(ft.Params))
		c.lambdaBodyAnyway(lam, sc)
		return
	}
	got := c.lambda(lam, sc, ft)
	if g, isF := got.(types.Func); isF {
		if invalidResult(g.Return) {
			// The body already reported why; its result says nothing about the type parameters.
			bindInvalid(ft.Return, subst)
			pruneSolved(unsolved, subst)
			return
		}
		if !c.unify(ft.Return, g.Return, subst) {
			want, _ := types.Subst(ft.Return, subst)
			c.errorf(a.Pos, "the lambda returns %s where %s is expected", types.String(g.Return), types.String(want))
			bindInvalid(ft.Return, subst) // the error explains the type parameters it left open
		}
		pruneSolved(unsolved, subst)
	}
}

// invalidResult reports whether a lambda's result is invalid, or a union with an invalid member,
// as when one branch of an if is an error.
func invalidResult(t types.Type) bool {
	if u, ok := t.(types.Union); ok {
		for _, m := range u.Members {
			if types.IsInvalid(m) {
				return true
			}
		}
	}
	return types.IsInvalid(t)
}

// pruneSolved removes from unsolved the type parameters subst now solves.
func pruneSolved(unsolved map[*types.ParamDecl]bool, subst map[*types.ParamDecl]types.Type) {
	for p := range unsolved {
		if _, done := subst[p]; done {
			delete(unsolved, p)
		}
	}
}

func anyParamMentions(ps []types.Type, set map[*types.ParamDecl]bool) bool {
	for _, p := range ps {
		if types.Mentions(p, set) {
			return true
		}
	}
	return false
}

func (c *Checker) resultOf(s *sig, subst map[*types.ParamDecl]types.Type, pos ast.Pos) types.Type {
	if s.fn.Return == nil {
		return types.Invalid{}
	}
	if subst == nil {
		if len(s.tparams) > 0 {
			return types.Invalid{}
		}
		return s.fn.Return
	}
	// A type parameter left invalid was reported where it failed, so a result built from it, such
	// as Set[T] of Set.of([]), is invalid as a whole and reports nothing further.
	failed := map[*types.ParamDecl]bool{}
	for p, t := range subst {
		if types.IsInvalid(t) {
			failed[p] = true
		}
	}
	if len(failed) > 0 && types.Mentions(s.fn.Return, failed) {
		return types.Invalid{}
	}
	t, err := types.Subst(s.fn.Return, subst)
	if err != nil {
		c.errorf(pos, "the result of %s: %s", s.fullName(), err)
		return types.Invalid{}
	}
	return t
}

// checkBounds checks each solution against its bound: assignable to it, and when the bound is
// a union, one member of it rather than a union. A bound may name the call's type parameters,
// its own included, and is read with their solutions in place (section 4, Generics).
func (c *Checker) checkBounds(s *sig, subst map[*types.ParamDecl]types.Type, pos ast.Pos) {
	for _, p := range s.tparams {
		if p.Bound == nil {
			continue
		}
		t := subst[p]
		if types.IsInvalid(t) {
			continue
		}
		bound, ok := solvedBound(p, subst, s.tparams)
		if !ok || boundFits(p, t, subst, s.tparams) {
			continue
		}
		c.errorf(pos, "%s would be %s, which is not %s its bound %s%s", p.Name, types.String(t), boundRelation(t, bound), types.String(bound), c.boundHint(p, t, subst, s.tparams))
	}
}

// boundRelation is what a solution t must be to a bound: a member of a union bound, a single one
// when t is a union itself, else assignable to it.
func boundRelation(t, bound types.Type) string {
	switch {
	case len(types.Members(bound)) > 1 && len(types.Members(t)) > 1:
		return "a single member of"
	case len(types.Members(bound)) > 1:
		return "a member of"
	}
	return "assignable to"
}

// boundHint says how to make a solution t fit p's bound, where something can: narrow a union
// that has a member in the bound, unwrap a named primitive whose base is in it. A type that
// cannot fit gets no hint.
func (c *Checker) boundHint(p *types.ParamDecl, t types.Type, subst map[*types.ParamDecl]types.Type, tparams []*types.ParamDecl) string {
	if ms := types.Members(t); len(ms) > 1 {
		for _, m := range ms {
			if boundFits(p, m, subst, tparams) {
				return "; narrow first"
			}
		}
		return ""
	}
	if n, ok := types.Expand(t).(types.Named); ok && n.Decl.Kind == types.PrimitiveDecl {
		base := types.Prim{Kind: n.Decl.Base}
		if boundFits(p, base, subst, tparams) {
			return fmt.Sprintf("; unwrap it with %s(...)", types.String(base))
		}
	}
	return ""
}

// solvedBound is the bound of p with the solutions in subst put in place of the type
// parameters of tparams it names. The second result is false when the bound still names one
// that subst leaves open, or names an invalid solution, so it cannot be judged yet.
func solvedBound(p *types.ParamDecl, subst map[*types.ParamDecl]types.Type, tparams []*types.ParamDecl) (types.Type, bool) {
	if p.Bound == nil {
		return nil, true
	}
	named := map[*types.ParamDecl]bool{}
	for _, q := range tparams {
		named[q] = true
	}
	if !types.Mentions(p.Bound, named) {
		return p.Bound, true
	}
	for _, q := range tparams {
		if !types.Mentions(p.Bound, map[*types.ParamDecl]bool{q: true}) {
			continue
		}
		if t, done := subst[q]; !done || types.IsInvalid(t) {
			return nil, false
		}
	}
	bound, err := types.Subst(p.Bound, subst)
	if err != nil {
		return nil, false
	}
	return bound, true
}

// boundFits reports whether t is a solution the bound of p admits: a single member of a union
// bound, or a type assignable to any other bound (section 4, Generics). The bound is read with
// the solutions in subst for the type parameters of tparams; one that still names an open
// parameter admits t for now, and checkBounds judges it once every parameter is solved.
func boundFits(p *types.ParamDecl, t types.Type, subst map[*types.ParamDecl]types.Type, tparams []*types.ParamDecl) bool {
	bound, ok := solvedBound(p, subst, tparams)
	if !ok || bound == nil {
		return true
	}
	return types.FitsBound(t, bound)
}

// unify matches a pattern that may mention type parameters against an actual type, binding
// the parameters it meets. A parameter already bound must match by equality.
func (c *Checker) unify(pattern, actual types.Type, subst map[*types.ParamDecl]types.Type) bool {
	// Invalid itself, not a type parameter whose bound is invalid (types.IsInvalid): that one is
	// still solved and solves others, so that a call is checked as usual.
	if _, ok := actual.(types.Invalid); ok {
		bindInvalid(pattern, subst)
		return true
	}
	if _, ok := pattern.(types.Invalid); ok {
		return true
	}
	if p, ok := pattern.(types.Param); ok {
		if c.rigid[p.Decl] {
			// the enclosing function's parameter: a type the caller chooses, fixed here
			return types.Equal(pattern, actual) || types.Assignable(actual, pattern)
		}
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
		if pt.Decl.Kind == types.InterfaceDecl && c.mentionsOpen(pattern) {
			return c.unifyInterface(pt, actual, subst)
		}
		return types.Equal(pattern, actual)
	case types.Array:
		a, ok := types.Expand(actual).(types.Array)
		return ok && c.unify(pt.Elem, a.Elem, subst)
	case types.Map:
		a, ok := types.Expand(actual).(types.Map)
		return ok && c.unify(pt.Key, a.Key, subst) && c.unify(pt.Val, a.Val, subst)
	case types.Set:
		a, ok := types.Expand(actual).(types.Set)
		return ok && c.unify(pt.Elem, a.Elem, subst)
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
			if !c.unifyParam(pt.Params[i], a.Params[i], subst) {
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

// unifyInterface matches an interface pattern whose arguments name open type parameters against
// a type that may satisfy it: each function of the interface, with Self as the actual type, is
// matched against the function a call through a dot reaches, as an argument matches a parameter
// (section 4, Inference). Container[T] against a Stack whose get returns string | null solves T
// as string.
func (c *Checker) unifyInterface(pt types.Named, actual types.Type, subst map[*types.ParamDecl]types.Type) bool {
	if _, isUnion := types.Expand(actual).(types.Union); isUnion {
		return false
	}
	m := types.Bindings(pt.Decl.Params, pt.Args)
	m[pt.Decl.Self] = actual
	trial := maps.Clone(subst)
	for _, name := range pt.Decl.Order {
		wt, err := types.Subst(*pt.Decl.Funcs[name], m)
		if err != nil {
			return false
		}
		want := wt.(types.Func)
		got := c.offered(actual, name, want, types.Assignable)
		if got == nil || !c.unify(want, *got, trial) {
			return false
		}
	}
	maps.Copy(subst, trial)
	return true
}

// unifyParam matches a parameter of a function argument. Matching solves the type parameters it
// mentions; once the expected parameter is known, the argument's need only accept it, as
// assignability has it (section 4): func(int | string): int is a func(int): U.
func (c *Checker) unifyParam(pattern, actual types.Type, subst map[*types.ParamDecl]types.Type) bool {
	trial := maps.Clone(subst)
	if c.unify(pattern, actual, trial) {
		maps.Copy(subst, trial)
		return true
	}
	want, err := types.Subst(pattern, subst)
	return err == nil && !c.mentionsOpen(want) && types.Assignable(want, actual)
}

// mentionsOpen reports whether t names a type parameter still being solved: one that is not
// rigid, a type the enclosing function's caller chooses.
func (c *Checker) mentionsOpen(t types.Type) bool {
	open := false
	types.Walk(t, func(t types.Type) {
		if p, ok := t.(types.Param); ok && !c.rigid[p.Decl] {
			open = true
		}
	})
	return open
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
		for i, am := range remaining {
			if sameDiscriminator(pm, am) {
				if !c.unify(pm, am, subst) {
					return false
				}
				remaining = append(remaining[:i], remaining[i+1:]...)
				break
			}
		}
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
		return true
	}
	return c.unify(params[0], rest, subst)
}

// expectedPart is the part of an expected union that a generic result can be: the members that
// share a discriminator with one of the result's. The result need only be assignable to the
// expected type, so Result[T, E] solves against Result[int, E1] | null by leaving null out. A
// result with a type parameter member could be any member, so it gets the whole union.
func expectedPart(result, expected types.Type) types.Type {
	ems := types.Members(expected)
	if len(ems) < 2 {
		return expected
	}
	rms := types.Members(result)
	for _, r := range rms {
		if _, ok := r.(types.Param); ok {
			return expected
		}
	}
	var keep []types.Type
	for _, e := range ems {
		for _, r := range rms {
			if sameDiscriminator(r, e) {
				keep = append(keep, e)
				break
			}
		}
	}
	if len(keep) == 0 || len(keep) == len(ems) {
		return expected
	}
	u, err := types.MakeUnion(keep)
	if err != nil {
		return expected
	}
	return u
}

// sameDiscriminator reports whether a and b are members that is tells apart by the same kind,
// tag or shape. A sync type has no discriminator, but a union holds at most one, so it pairs with
// its own declaration.
func sameDiscriminator(a, b types.Type) bool {
	ka, _, oka := types.Discriminator(a)
	kb, _, okb := types.Discriminator(b)
	if oka || okb {
		return oka && okb && ka == kb
	}
	na, isA := types.Expand(a).(types.Named)
	nb, isB := types.Expand(b).(types.Named)
	return isA && isB && na.Decl == nb.Decl
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
			if slices.ContainsFunc(elems, types.IsInvalid) {
				return types.Invalid{} // an element's error was reported; the array would only repeat it
			}
			u, err := types.MakeUnion(elems)
			if err != nil {
				c.errorf(x.Pos, "the elements do not form a valid type: %s; give the array an expected type", err)
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
			if slices.ContainsFunc(keys, types.IsInvalid) || slices.ContainsFunc(vals, types.IsInvalid) {
				return types.Invalid{}
			}
			k, errK := types.MakeUnion(keys)
			v, errV := types.MakeUnion(vals)
			if errK != nil || errV != nil {
				c.errorf(x.Pos, "the entries do not form a valid type; give the map an expected type")
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
		// A lambda inside a literal, as in Parallel.race([{ … }, { … }]), is checked against
		// the element's function type once its parameters are known, and its body solves the
		// result, as a lambda argument's does.
		pat, err := types.Subst(pattern, subst)
		ft, isFunc := types.Expand(pat).(types.Func)
		if err != nil || !isFunc || len(ft.Params) > 1 || anyParamMentions(ft.Params, unsolved) {
			return c.synth(e, sc)
		}
		got := c.lambda(x, sc, ft)
		if g, isF := got.(types.Func); isF && c.unify(ft.Return, g.Return, subst) {
			pruneSolved(unsolved, subst)
			if solved, err := types.Subst(pattern, subst); err == nil && !types.Mentions(solved, unsolved) {
				return solved
			}
		}
		return got
	}
	solved, err := types.Subst(pattern, subst)
	if err == nil && !types.Mentions(solved, unsolved) {
		return c.check(e, sc, solved)
	}
	if err == nil {
		// partly solved, as Result[T, Error]: a call still takes what is known
		return c.synthHinted(e, sc, solved)
	}
	return c.synth(e, sc)
}

// synthHinted synthesizes e. When e is a call, hint is a type its result should fit, which
// solves only the type parameters its arguments leave open (section 4, Inference): passed where
// Result[T, Error] is expected, Result.ok(1) solves E to Error and T to int from its argument.
// Unlike an expected type, a hint that does not fit is no error, and a part of it that names a
// type parameter still being solved, as T there, solves nothing.
func (c *Checker) synthHinted(e ast.Expr, sc *scope, hint types.Type) types.Type {
	if x, ok := e.(*ast.Call); ok {
		if c.hints == nil {
			c.hints = map[*ast.Call]types.Type{}
		}
		c.hints[x] = hint
		defer delete(c.hints, x)
	}
	return c.synth(e, sc)
}

// solveFromHint solves the type parameters of s still open from the hint for its result, as the
// expected type does but after the arguments.
func (c *Checker) solveFromHint(s *sig, hint types.Type, subst map[*types.ParamDecl]types.Type, unsolved map[*types.ParamDecl]bool) {
	// A type parameter of an enclosing call, still open, is a hole: it matches anything and
	// solves nothing. Invalid does that in unify.
	holes := map[*types.ParamDecl]types.Type{}
	types.Walk(hint, func(t types.Type) {
		if p, ok := t.(types.Param); ok && !c.rigid[p.Decl] {
			holes[p.Decl] = types.Invalid{}
		}
	})
	hint, err := types.Subst(hint, holes)
	if err != nil {
		return
	}
	trial := maps.Clone(subst)
	if !c.unify(s.fn.Return, expectedPart(s.fn.Return, hint), trial) {
		return
	}
	for p := range unsolved {
		if t, ok := trial[p]; ok && !hasHole(t) && boundFits(p, t, trial, s.tparams) {
			subst[p] = t
		}
	}
	pruneSolved(unsolved, subst)
}

// hasHole reports whether t holds Invalid anywhere, as Result[?, Error] does.
func hasHole(t types.Type) bool {
	found := false
	types.Walk(t, func(t types.Type) {
		if _, ok := t.(types.Invalid); ok {
			found = true
		}
	})
	return found
}

// onlyEmpty reports an empty array or map literal, an array literal of nothing but those, or a
// call whose arguments are only those and lambdas, such as Set.of([]): what only an expected type
// can give a type.
func onlyEmpty(e ast.Expr) bool {
	if isEmptyLiteral(e) {
		return true
	}
	if call, ok := e.(*ast.Call); ok {
		// A call whose arguments fix nothing, as Set.of([]), takes its type from context too.
		empty := false
		for _, a := range call.Args {
			if _, isLambda := a.Value.(*ast.Lambda); isLambda {
				continue
			}
			if !onlyEmpty(a.Value) {
				return false
			}
			empty = true
		}
		return empty
	}
	x, ok := e.(*ast.ArrayLit)
	if !ok {
		return false
	}
	for _, el := range x.Elems {
		if el.Spread || !onlyEmpty(el.Value) {
			return false
		}
	}
	return true
}

// emptyError is the error for what onlyEmpty accepts when nothing gives it a type, worded as
// for the same literal in a var.
func emptyError(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.MapLit:
		return "cannot infer the key and value types of an empty map; annotate it"
	case *ast.ArrayLit:
		if len(x.Elems) > 0 {
			return "cannot infer the element type of an array of empty literals; annotate it"
		}
	}
	return "cannot infer the element type of an empty array; annotate it"
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
	case types.Set:
		bindInvalid(x.Elem, subst)
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
// and the plain path reports the problem; when a known parameter does not fit, it reports the
// mismatch and returns Invalid.
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
	if ent == nil {
		return nil
	}
	c.notePath(e, sc)
	var s *sig
	switch {
	case ent.kind == entType:
		s = ctorSig(ent.decl)
	case ent.kind != entFunc:
		return nil
	case convValue(ent, pf) != nil:
		return convValue(ent, pf)
	case ent.isLen:
		if len(pf.Params) == 1 && types.Mentions(pf.Params[0], unsolved) {
			return nil
		}
		return c.lenValue(ent, pf, e.Position())
	default:
		s = c.sigOf(ent, e.Position())
	}
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
			// The expected parameter is known and does not fit: a mismatch, not a missing
			// expected type, so say so rather than leave it to the plain path.
			c.errorf(e.Position(), "cannot use %s where %s is expected: it takes %s, not %s", s.fullName(), types.String(pat), types.String(s.fn.Params[i]), types.String(pf.Params[i]))
			return types.Invalid{}
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

// isBytes reports whether t is the prelude's Bytes, which len, for and indexing take (section 11).
func (c *Checker) isBytes(t types.Type) bool {
	n, ok := types.Expand(t).(types.Named)
	return ok && n.Decl == c.bytes
}

// recordSlots records for an editor which parameter each argument of a call fills: by slots
// when the arguments matched, and as near as they go when they did not.
func (c *Checker) recordSlots(s *sig, x *ast.Call, slots []*ast.Arg, ok bool, callee *Object, recv bool) {
	if c.info == nil {
		return
	}
	ci := &CallInfo{Fn: s.fn, Names: s.fn.Names, Defaults: s.defaults, Callee: callee, Slots: make([]int, len(x.Args)), Recv: recv}
	for i := range ci.Slots {
		ci.Slots[i] = -1
	}
	for param, a := range slots {
		if a != nil {
			for i := range x.Args {
				if &x.Args[i] == a {
					ci.Slots[i] = param
				}
			}
		}
	}
	if !ok {
		ci.Slots = argSlotsLoosely(s, x, recv)
	}
	c.info.recordCall(x, ci)
}

// assertCall checks a call of assert: the condition as an if's, the message where the condition
// is false, and after the call the flow where it is true, as after if !cond { panic(...) }
// (section 7, Narrowing).
func (c *Checker) assertCall(s *sig, x *ast.Call, sc *scope, callee *Object) types.Type {
	slots, ok := c.argSlots(s, x, nil)
	c.recordSlots(s, x, slots, ok, callee, false)
	if !ok {
		c.synthArgs(x.Args, sc)
		return types.Void{}
	}
	whenTrue, whenFalse := c.cond(slots[0].Value, sc)
	if slots[1] != nil {
		c.flow = whenFalse
		c.check(slots[1].Value, sc, types.StringType)
	}
	c.flow = whenTrue
	return types.Void{}
}
