package check

import (
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// lookup finds a name: a local in the scope chain, else a module-level or prelude entity.
func (c *Checker) lookup(sc *scope, name string) (*local, *entity) {
	for s := sc; s != nil; s = s.parent {
		if s.names != nil {
			if l, ok := s.names[name]; ok {
				return l, nil
			}
		}
	}
	return nil, c.lookupEntity(name)
}

// typeOfLocal is a local's current type: narrowed in the flow, else declared.
func (c *Checker) typeOfLocal(l *local) types.Type {
	if t, ok := c.flow[l]; ok {
		return t
	}
	return l.typ
}

// inNested reports whether the current context is a nested body of l's function.
func (c *Checker) inNested(l *local) bool { return l.ctx != c.fn }

// check checks e against an expected type and returns e's type.
func (c *Checker) check(e ast.Expr, sc *scope, expected types.Type) types.Type {
	if expected == nil {
		return c.synth(e, sc)
	}
	var t types.Type
	switch x := e.(type) {
	case *ast.Lambda:
		return c.checkLambda(x, sc, expected)
	case *ast.ArrayLit:
		if a, ok := types.Expand(expected).(types.Array); ok {
			t = c.checkArrayLit(x, sc, a.Elem)
		} else if u, ok := types.Expand(expected).(types.Union); ok {
			for _, m := range u.Members {
				if a, ok := types.Expand(m).(types.Array); ok {
					t = c.checkArrayLit(x, sc, a.Elem)
					break
				}
			}
		}
	case *ast.MapLit:
		if m, ok := types.Expand(expected).(types.Map); ok {
			t = c.checkMapLit(x, sc, m)
		} else if u, ok := types.Expand(expected).(types.Union); ok {
			for _, mem := range u.Members {
				if m, ok := types.Expand(mem).(types.Map); ok {
					t = c.checkMapLit(x, sc, m)
					break
				}
			}
		}
	case *ast.RecordLit:
		t = c.recordLit(x, sc, expected)
	case *ast.If:
		t = c.checkIf(x, sc, expected)
	case *ast.Call:
		t = c.call(x, sc, expected)
	case *ast.Block:
		t = c.checkBlock(x, sc, expected)
	case *ast.Binary:
		if x.Op == "??" {
			t = c.nullish(x, sc, expected)
		}
	case *ast.Ident, *ast.Field:
		t = c.genericValue(x, sc, expected)
	}
	if t == nil {
		t = c.synth(e, sc)
	}
	c.expect(e.Position(), t, expected)
	return t
}

// genericValue instantiates a generic function used as a value from the expected function
// type, which fixes its type parameters (section 4, Generics). It returns nil when e is not a
// generic function or no function type is expected.
func (c *Checker) genericValue(e ast.Expr, sc *scope, expected types.Type) types.Type {
	f, ok := types.Expand(expected).(types.Func)
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
	subst := map[*types.ParamDecl]types.Type{}
	if !c.unify(s.fn, f, subst) {
		c.errorf(e.Position(), "cannot use %s where %s is expected", s.fullName(), types.String(expected))
		return types.Invalid{}
	}
	for _, p := range s.tparams {
		if _, bound := subst[p]; !bound {
			c.errorf(e.Position(), "%s does not fix %s of %s", types.String(expected), p.Name, s.fullName())
			return types.Invalid{}
		}
	}
	inst, err := types.Subst(s.fn, subst)
	if err != nil {
		c.errorf(e.Position(), "%s", err)
		return types.Invalid{}
	}
	return inst
}

// expect reports a type that is not assignable to the expected one.
func (c *Checker) expect(pos ast.Pos, t, expected types.Type) {
	if !types.Assignable(t, expected) {
		c.errorf(pos, "cannot use %s where %s is expected", types.String(t), types.String(expected))
	}
}

// synth determines the type of an expression without an expected type.
func (c *Checker) synth(e ast.Expr, sc *scope) types.Type {
	switch x := e.(type) {
	case *ast.Ident:
		return c.ident(x, sc)
	case *ast.IntLit:
		return types.IntType
	case *ast.FloatLit:
		return types.Float64Type
	case *ast.StringLit:
		return types.StringType
	case *ast.RegexLit:
		return c.regexPattern()
	case *ast.BoolLit:
		return types.BoolType
	case *ast.NullLit:
		return types.Null{}
	case *ast.VoidLit:
		return types.Void{}
	case *ast.Template:
		for _, p := range x.Parts {
			if p.Expr != nil {
				c.synth(p.Expr, sc)
			}
		}
		return types.StringType
	case *ast.RecordLit:
		return c.recordLit(x, sc, nil)
	case *ast.MapLit:
		return c.synthMapLit(x, sc)
	case *ast.ArrayLit:
		return c.synthArrayLit(x, sc)
	case *ast.Lambda:
		c.errorf(x.Pos, "a lambda needs an expected function type; annotate the variable or use func with typed parameters")
		c.lambdaBodyAnyway(x, sc)
		return types.Invalid{}
	case *ast.FuncLit:
		return c.funcLit(x, sc)
	case *ast.Field:
		return c.field(x, sc)
	case *ast.Index:
		return c.index(x, sc)
	case *ast.Call:
		return c.call(x, sc, nil)
	case *ast.Unary:
		return c.unary(x, sc)
	case *ast.Binary:
		return c.binary(x, sc)
	case *ast.Is:
		return c.isExpr(x, sc)
	case *ast.Assign:
		return c.assign(x, sc)
	case *ast.If:
		return c.checkIf(x, sc, nil)
	case *ast.While:
		return c.whileLoop(x, sc)
	case *ast.For:
		return c.forLoop(x, sc)
	case *ast.Go:
		c.nestedBody(x.Body, sc, "go", x.Scope)
		return types.Void{}
	case *ast.Defer:
		c.nestedBody(x.Body, sc, "defer", x.Scope)
		return types.Void{}
	case *ast.Return:
		return c.returnExpr(x, sc)
	case *ast.Break:
		c.markBreak()
		return types.Never{}
	case *ast.Continue:
		return types.Never{}
	case *ast.Block:
		return c.synthBlock(x, sc)
	case *ast.Placeholder:
		return types.Invalid{}
	}
	c.errorf(e.Position(), "internal error: cannot check %T", e)
	return types.Invalid{}
}

// ---- names

func (c *Checker) ident(x *ast.Ident, sc *scope) types.Type {
	if x.Name == "_" {
		c.errorf(x.Pos, "_ is write-only and cannot be read")
		return types.Invalid{}
	}
	l, ent := c.lookup(sc, x.Name)
	if l != nil {
		if l.pending {
			if l.ctx == c.fn {
				c.errorf(x.Pos, "%s is read in its own initializer", x.Name)
				return types.Invalid{}
			}
			if !l.stated {
				c.errorf(x.Pos, "%s is used by a function in its own initializer; state its type on the var", x.Name)
				return types.Invalid{}
			}
		}
		return c.typeOfLocal(l)
	}
	if ent == nil {
		c.errorf(x.Pos, "%s is not defined", x.Name)
		return types.Invalid{}
	}
	return c.entityValue(ent, x.Pos, sc)
}

// entityValue is the type of an entity used as a value.
func (c *Checker) entityValue(ent *entity, pos ast.Pos, sc *scope) types.Type {
	switch ent.kind {
	case entVar:
		return c.varType(ent.v, pos, c.fn != nil && c.fn.kind == "init")
	case entConst:
		return ent.typ
	case entFunc:
		s := c.sigOf(ent, pos)
		if s == nil {
			return types.Invalid{}
		}
		if len(s.tparams) > 0 {
			c.errorf(pos, "%s is generic; call it, or give the expected function type that fixes its type parameters", s.fullName())
			return types.Invalid{}
		}
		return s.fn
	case entType:
		if ent.decl.Kind == types.PrimitiveDecl {
			return c.constructorType(ent.decl)
		}
		c.errorf(pos, "%s is a type, not a value; a type appears only as a literal head, after is, or before a dot", ent.decl.Name)
		return types.Invalid{}
	case entModule, entNamespace:
		c.errorf(pos, "%s is a namespace, not a value", ent.name)
		return types.Invalid{}
	}
	return types.Invalid{}
}

// constructorType is the function type of a named primitive type's constructor.
func (c *Checker) constructorType(decl *types.Decl) types.Type {
	var args []types.Type
	for _, p := range decl.Params {
		args = append(args, types.Param{Decl: p})
	}
	return types.Func{Params: []types.Type{types.Prim{Kind: decl.Base}}, Return: types.Named{Decl: decl, Args: args}, Required: 1}
}

// sigOf returns a function entity's signature with its return type known.
func (c *Checker) sigOf(ent *entity, pos ast.Pos) *sig {
	s := ent.sig
	if s == nil {
		return nil
	}
	if s.fn.Return == nil {
		ret := c.returnTypeOf(ent, pos)
		if s.fn.Return == nil {
			cp := *s
			cp.fn.Return = ret
			return &cp
		}
	}
	return s
}

// ---- fields

// entityPath resolves a dotted path of entities: an import's export, a namespace member, a
// function on a type. It returns nil when the expression is not such a path.
func (c *Checker) entityPath(e ast.Expr, sc *scope) *entity {
	switch x := e.(type) {
	case *ast.Ident:
		if l, _ := c.lookup(sc, x.Name); l != nil {
			return nil
		}
		return c.lookupEntity(x.Name)
	case *ast.Field:
		if x.Optional {
			return nil
		}
		base := c.entityPath(x.X, sc)
		if base == nil {
			return nil
		}
		if base.kind == entModule {
			m, ok := base.mod.names[x.Name]
			if !ok || !m.exported {
				return nil
			}
			return m
		}
		if base.members == nil {
			return nil
		}
		m, ok := base.members[x.Name]
		if !ok {
			return nil
		}
		return m
	}
	return nil
}

// moduleMember reports a member of an imported module, or the reason it is not visible.
func (c *Checker) moduleMember(base *entity, name string, pos ast.Pos) *entity {
	m, ok := base.mod.names[name]
	if !ok {
		c.errorf(pos, "%s has no declaration %s", base.name, name)
		return nil
	}
	if !m.exported {
		c.errorf(pos, "%s is not exported by %s", name, base.name)
		return nil
	}
	return m
}

func (c *Checker) field(x *ast.Field, sc *scope) types.Type {
	if base := c.entityPath(x.X, sc); base != nil && !x.Optional {
		switch base.kind {
		case entModule:
			m := c.moduleMember(base, x.Name, x.Pos)
			if m == nil {
				return types.Invalid{}
			}
			return c.entityValue(m, x.Pos, sc)
		case entNamespace, entType:
			m, ok := base.members[x.Name]
			if !ok {
				if base.kind == entType && base.decl.Kind == types.PrimitiveDecl {
					break
				}
				c.errorf(x.Pos, "%s has no member %s", base.name, x.Name)
				return types.Invalid{}
			}
			return c.entityValue(m, x.Pos, sc)
		}
	}
	recv := c.synth(x.X, sc)
	if x.Optional {
		if !types.Includes(recv, types.Null{}) {
			c.errorf(x.Pos, "?. needs a nullable left side, got %s", types.String(recv))
			return types.Invalid{}
		}
		inner := c.fieldOf(types.Without(recv, types.Null{}), x.Name, x.Pos)
		u, _ := types.MakeUnion([]types.Type{inner, types.Null{}})
		return u
	}
	return c.fieldOf(recv, x.Name, x.Pos)
}

// fieldOf is the type of a field read through a type: a record, named or anonymous, or a
// union of records that all declare it.
func (c *Checker) fieldOf(recv types.Type, name string, pos ast.Pos) types.Type {
	if types.IsInvalid(recv) {
		return types.Invalid{}
	}
	switch r := types.Expand(recv).(type) {
	case types.Union:
		var members []types.Type
		for _, m := range r.Members {
			fields := types.FieldsOf(m)
			if fields == nil {
				c.errorf(pos, "%s has no field %s: narrow the value first", types.String(recv), name)
				return types.Invalid{}
			}
			found := false
			for _, f := range fields {
				if f.Name == name {
					members = append(members, f.Type)
					found = true
				}
			}
			if !found {
				c.errorf(pos, "%s has no field %s on every member: narrow the value first", types.String(recv), name)
				return types.Invalid{}
			}
		}
		u, err := types.MakeUnion(members)
		if err != nil {
			c.errorf(pos, "reading %s through %s: %s", name, types.String(recv), err)
			return types.Invalid{}
		}
		return u
	}
	fields := types.FieldsOf(recv)
	if fields == nil {
		c.errorf(pos, "%s is not a record, so it has no field %s", types.String(recv), name)
		return types.Invalid{}
	}
	for _, f := range fields {
		if f.Name == name {
			return f.Type
		}
	}
	c.errorf(pos, "%s has no field %s", types.String(recv), name)
	return types.Invalid{}
}

func (c *Checker) index(x *ast.Index, sc *scope) types.Type {
	recv := c.synth(x.X, sc)
	if x.Optional {
		if !types.Includes(recv, types.Null{}) {
			c.errorf(x.Pos, "?.[ needs a nullable left side, got %s", types.String(recv))
			return types.Invalid{}
		}
		inner := c.indexOf(types.Without(recv, types.Null{}), x, sc)
		u, _ := types.MakeUnion([]types.Type{inner, types.Null{}})
		return u
	}
	return c.indexOf(recv, x, sc)
}

func (c *Checker) indexOf(recv types.Type, x *ast.Index, sc *scope) types.Type {
	switch r := types.Expand(recv).(type) {
	case types.Array:
		c.check(x.Index, sc, types.IntType)
		return r.Elem
	case types.Map:
		c.check(x.Index, sc, r.Key)
		u, _ := types.MakeUnion([]types.Type{r.Val, types.Null{}})
		return u
	case types.Invalid:
		c.synth(x.Index, sc)
		return types.Invalid{}
	}
	c.synth(x.Index, sc)
	if types.FieldsOf(recv) != nil {
		c.errorf(x.Pos, "a record has no dynamic access; use a Map when the keys are data")
	} else {
		c.errorf(x.Pos, "%s cannot be indexed", types.String(recv))
	}
	return types.Invalid{}
}

// ---- literals

func (c *Checker) recordLit(x *ast.RecordLit, sc *scope, expected types.Type) types.Type {
	head := c.entityPath(x.Head, sc)
	if head == nil {
		if f, ok := x.Head.(*ast.Field); ok {
			if base := c.entityPath(f.X, sc); base != nil && base.kind == entModule {
				c.moduleMember(base, f.Name, f.Pos)
				return types.Invalid{}
			}
		}
	}
	if head == nil || head.kind != entType {
		c.errorf(x.Pos, "a record literal needs a record type as its head")
		for _, en := range x.Entries {
			if en.Spread != nil {
				c.synth(en.Spread, sc)
			} else {
				c.synth(en.Value, sc)
			}
		}
		return types.Invalid{}
	}
	decl := head.decl
	if decl.Kind != types.RecordDecl {
		c.errorf(x.Pos, "%s is not a record type, so it has no literal", decl.Name)
		return types.Invalid{}
	}
	// type arguments: from the expected type, else inferred from the entries
	subst := map[*types.ParamDecl]types.Type{}
	if expected != nil {
		for _, m := range types.Members(expected) {
			if n, ok := types.Expand(m).(types.Named); ok && n.Decl == decl {
				subst = bindingsOf(decl.Params, n.Args)
				break
			}
		}
	}
	unsolved := map[*types.ParamDecl]bool{}
	for _, p := range decl.Params {
		if _, ok := subst[p]; !ok {
			unsolved[p] = true
		}
	}
	given := map[string]bool{}
	spreadUnknown := false // a spread whose fields cannot be known supplies no field
	for _, en := range x.Entries {
		if en.Spread != nil {
			st := c.synth(en.Spread, sc)
			n, ok := types.Expand(st).(types.Named)
			if !ok || n.Decl.Kind != types.RecordDecl {
				if !types.IsInvalid(st) {
					c.errorf(en.Pos, "a spread needs a value of a named record type, got %s", types.String(st))
				}
				spreadUnknown = true
				continue
			}
			for _, f := range n.FieldsOf() {
				idx := fieldIndex(decl, f.Name)
				if idx < 0 {
					c.errorf(en.Pos, "%s has no field %s, which the spread of %s would copy", decl.Name, f.Name, types.String(st))
					continue
				}
				c.matchField(decl.Fields[idx].Type, f.Type, subst, unsolved, en.Pos, f.Name)
				given[f.Name] = true
			}
			continue
		}
		idx := fieldIndex(decl, en.Name)
		if idx < 0 {
			c.errorf(en.Pos, "%s has no field %s", decl.Name, en.Name)
			c.synth(en.Value, sc)
			continue
		}
		want := decl.Fields[idx].Type
		if types.Mentions(want, unsolved) {
			got := c.synth(en.Value, sc)
			c.matchField(want, got, subst, unsolved, en.Pos, en.Name)
		} else {
			w, err := types.Subst(want, subst)
			if err != nil {
				c.errorf(en.Pos, "%s", err)
				w = types.Invalid{}
			}
			c.check(en.Value, sc, w)
		}
		given[en.Name] = true
	}
	for _, f := range decl.Fields {
		if !given[f.Name] && !spreadUnknown {
			c.errorf(x.Pos, "%s literal is missing the field %s", decl.Name, f.Name)
		}
	}
	var args []types.Type
	for _, p := range decl.Params {
		t, ok := subst[p]
		if !ok {
			c.errorf(x.Pos, "cannot infer the type parameter %s of %s; annotate the result", p.Name, decl.Name)
			t = types.Invalid{}
		}
		args = append(args, t)
	}
	return types.Named{Decl: decl, Args: args}
}

func fieldIndex(decl *types.Decl, name string) int {
	for i, f := range decl.Fields {
		if f.Name == name {
			return i
		}
	}
	return -1
}

// matchField unifies a field's declared type with a value's type, solving type parameters.
func (c *Checker) matchField(want, got types.Type, subst map[*types.ParamDecl]types.Type, unsolved map[*types.ParamDecl]bool, pos ast.Pos, name string) {
	if !c.unify(want, got, subst) {
		w, _ := types.Subst(want, subst)
		c.errorf(pos, "cannot use %s for the field %s of type %s", types.String(got), name, types.String(w))
		return
	}
	for p := range unsolved {
		if _, ok := subst[p]; ok {
			delete(unsolved, p)
		}
	}
}

func (c *Checker) synthArrayLit(x *ast.ArrayLit, sc *scope) types.Type {
	if len(x.Elems) == 0 {
		c.errorf(x.Pos, "cannot infer the element type of an empty array; annotate it")
		return types.Invalid{}
	}
	var elems []types.Type
	var empties []ast.Expr
	for _, el := range x.Elems {
		if isEmptyLiteral(el.Value) {
			empties = append(empties, el.Value)
			continue
		}
		t := c.synth(el.Value, sc)
		if el.Spread {
			a, ok := types.Expand(t).(types.Array)
			if !ok {
				c.errorf(el.Value.Position(), "a spread in an array literal needs an array, got %s", types.String(t))
				continue
			}
			t = a.Elem
		}
		elems = append(elems, t)
	}
	if len(elems) == 0 {
		c.errorf(x.Pos, "cannot infer the element type of %s; annotate it", "an array of empty literals")
		return types.Invalid{}
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

// isEmptyLiteral reports an empty array or map literal, whose type only context can give.
func isEmptyLiteral(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.ArrayLit:
		return len(x.Elems) == 0
	case *ast.MapLit:
		return len(x.Entries) == 0
	}
	return false
}

func (c *Checker) checkArrayLit(x *ast.ArrayLit, sc *scope, elem types.Type) types.Type {
	for _, el := range x.Elems {
		if el.Spread {
			t := c.synth(el.Value, sc)
			a, ok := types.Expand(t).(types.Array)
			if !ok {
				if !types.IsInvalid(t) {
					c.errorf(el.Value.Position(), "a spread in an array literal needs an array, got %s", types.String(t))
				}
				continue
			}
			c.expect(el.Value.Position(), a.Elem, elem)
		} else {
			c.check(el.Value, sc, elem)
		}
	}
	return types.Array{Elem: elem}
}

func (c *Checker) synthMapLit(x *ast.MapLit, sc *scope) types.Type {
	if len(x.Entries) == 0 {
		c.errorf(x.Pos, "cannot infer the key and value types of an empty map; annotate it")
		return types.Invalid{}
	}
	var keys, vals []types.Type
	var empties []ast.Expr
	for _, en := range x.Entries {
		keys = append(keys, c.synth(en.Key, sc))
		if isEmptyLiteral(en.Value) {
			empties = append(empties, en.Value)
			continue
		}
		vals = append(vals, c.synth(en.Value, sc))
	}
	if len(vals) == 0 {
		c.errorf(x.Pos, "cannot infer the value type of a map of empty literals; annotate it")
		return types.Invalid{}
	}
	k, err := types.MakeUnion(keys)
	if err != nil {
		c.errorf(x.Pos, "the keys do not form a valid type: %s; give the map an expected type", err)
		return types.Invalid{}
	}
	v, err := types.MakeUnion(vals)
	if err != nil {
		c.errorf(x.Pos, "the values do not form a valid type: %s; give the map an expected type", err)
		return types.Invalid{}
	}
	for _, e := range empties {
		c.check(e, sc, v)
	}
	return types.Map{Key: k, Val: v}
}

func (c *Checker) checkMapLit(x *ast.MapLit, sc *scope, m types.Map) types.Type {
	for _, en := range x.Entries {
		c.check(en.Key, sc, m.Key)
		c.check(en.Value, sc, m.Val)
	}
	return m
}

// ---- functions

func (c *Checker) checkLambda(x *ast.Lambda, sc *scope, expected types.Type) types.Type {
	var ft types.Func
	found := false
	for _, m := range types.Members(expected) {
		if f, ok := types.Expand(m).(types.Func); ok {
			ft = f
			found = true
			break
		}
	}
	if !found {
		if !types.IsInvalid(expected) {
			c.errorf(x.Pos, "a lambda needs an expected function type, but %s is expected", types.String(expected))
		}
		c.lambdaBodyAnyway(x, sc)
		return types.Invalid{}
	}
	if len(ft.Params) > 1 {
		c.errorf(x.Pos, "a lambda has at most one parameter, x; %s expects %d: write func with named parameters", types.String(ft), len(ft.Params))
		c.lambdaBodyAnyway(x, sc)
		return types.Invalid{}
	}
	return c.lambda(x, sc, ft)
}

// lambda checks a lambda against the function type it must have. When the expected return
// type is an unsolved type parameter, the body's type stands in for it.
func (c *Checker) lambda(x *ast.Lambda, sc *scope, ft types.Func) types.Type {
	ctx := &fnCtx{parent: c.fn, kind: "lambda", name: "lambda", declared: ft.Return}
	if types.HasParam(ft.Return) {
		// The expected result mentions a type parameter the call has not solved yet, as U in
		// Array.map's U or Array.flatMap's Array[U]; the body is synthesized and the call
		// unifies the two to solve it (section 4, Inference).
		ctx.declared = nil
	}
	inner := &scope{parent: sc, names: map[string]*local{}}
	if len(ft.Params) == 1 {
		inner.names["x"] = &local{name: "x", typ: ft.Params[0], slot: x.Scope.X, ctx: ctx, pos: x.Pos}
	}
	ret := c.withNested(ctx, func() types.Type { return c.inferBody(x.Body, inner, ctx) })
	return types.Func{Params: ft.Params, Return: ret, Required: len(ft.Params)}
}

// lambdaBodyAnyway checks a lambda's body for its own errors when it has no expected type.
func (c *Checker) lambdaBodyAnyway(x *ast.Lambda, sc *scope) {
	ctx := &fnCtx{parent: c.fn, kind: "lambda", name: "lambda"}
	inner := &scope{parent: sc, names: map[string]*local{}}
	inner.names["x"] = &local{name: "x", typ: types.Invalid{}, slot: x.Scope.X, ctx: ctx, pos: x.Pos}
	c.withNested(ctx, func() types.Type { return c.inferBody(x.Body, inner, ctx) })
}

func (c *Checker) funcLit(x *ast.FuncLit, sc *scope) types.Type {
	d := &ast.FuncDecl{Pos: x.Pos, Params: x.Params, Return: x.Return, Body: x.Body}
	s := c.signature(d, sc, "")
	ctx := &fnCtx{parent: c.fn, kind: "func", name: "func", declared: s.fn.Return}
	inner := &scope{parent: sc, names: map[string]*local{}}
	c.declareParams(x.Params, s, inner, ctx)
	ret := c.withNested(ctx, func() types.Type { return c.inferBody(x.Body, inner, ctx) })
	return types.Func{Params: s.fn.Params, Return: ret, Required: s.fn.Required}
}

// withNested runs f as a nested body: a new context whose flow drops the narrowing of every
// outer local that is ever reassigned (section 7).
func (c *Checker) withNested(ctx *fnCtx, f func() types.Type) types.Type {
	savedFn, savedFlow := c.fn, c.flow
	c.fn = ctx
	nf := flow{}
	for l, t := range savedFlow {
		if !l.slot.Reassigned {
			nf[l] = t
		}
	}
	c.flow = nf
	t := f()
	c.fn, c.flow = savedFn, savedFlow
	return t
}

// nestedBody checks a go or defer block.
func (c *Checker) nestedBody(body *ast.Block, sc *scope, kind string, fs *ast.FuncScope) {
	ctx := &fnCtx{parent: c.fn, kind: kind, name: kind + " block"}
	inner := &scope{parent: sc, names: map[string]*local{}}
	c.withNested(ctx, func() types.Type { c.synthBlock(body, inner); return types.Void{} })
}

// ---- operators

func (c *Checker) unary(x *ast.Unary, sc *scope) types.Type {
	switch x.Op {
	case "!":
		c.check(x.X, sc, types.BoolType)
		return types.BoolType
	case "-":
		t := c.synth(x.X, sc)
		if !c.operandOK(t, "-", false) {
			c.errorf(x.Pos, "unary - needs a number, got %s", types.String(t))
			return types.Invalid{}
		}
		return t
	case "try":
		return c.tryExpr(x, sc)
	}
	return types.Invalid{}
}

// operandOK reports whether an arithmetic operand type is a single numeric kind, string when
// strings are allowed, a named primitive over such a base, or a type parameter whose bound
// admits the operator on every member.
func (c *Checker) operandOK(t types.Type, op string, strings bool) bool {
	switch x := types.Expand(t).(type) {
	case types.Invalid:
		return true
	case types.Prim:
		if x.Kind.IsNumber() {
			return true
		}
		return strings && x.Kind == value.KindString
	case types.Named:
		if x.Decl.Kind != types.PrimitiveDecl {
			return false
		}
		return x.Decl.Base.IsNumber() || strings && x.Decl.Base == value.KindString
	case types.Param:
		if x.Decl.Bound == nil {
			return false
		}
		for _, m := range types.Members(x.Decl.Bound) {
			if !c.operandOK(m, op, strings) {
				return false
			}
		}
		return true
	}
	return false
}

func (c *Checker) binary(x *ast.Binary, sc *scope) types.Type {
	switch x.Op {
	case "&&", "||":
		// The right operand sees the narrowing of the left one, as in a condition (section 7),
		// but the expression's value narrows nothing after it.
		saved := c.flow
		c.cond(x, sc)
		c.flow = saved
		return types.BoolType
	case "??":
		return c.nullish(x, sc, nil)
	case "==", "!=":
		l := c.synth(x.X, sc)
		r := c.synth(x.Y, sc)
		if !types.Assignable(l, r) && !types.Assignable(r, l) {
			c.errorf(x.Pos, "%s and %s cannot be compared with %s: neither is assignable to the other", types.String(l), types.String(r), x.Op)
		}
		return types.BoolType
	case "<", "<=", ">", ">=":
		l := c.synth(x.X, sc)
		r := c.synth(x.Y, sc)
		if !c.sameOperands(l, r, x.Op, true, x.Pos) {
			return types.BoolType
		}
		return types.BoolType
	}
	l := c.synth(x.X, sc)
	r := c.synth(x.Y, sc)
	if !c.sameOperands(l, r, x.Op, x.Op == "+", x.Pos) {
		return types.Invalid{}
	}
	if types.IsInvalid(l) {
		return r
	}
	return l
}

// sameOperands checks that two operands have one type that the operator applies to.
func (c *Checker) sameOperands(l, r types.Type, op string, strings bool, pos ast.Pos) bool {
	if types.IsInvalid(l) || types.IsInvalid(r) {
		return true
	}
	if !types.Equal(l, r) {
		c.errorf(pos, "%s needs both operands of one type, got %s and %s", op, types.String(l), types.String(r))
		return false
	}
	if !c.operandOK(l, op, strings) {
		what := "a single numeric kind or a named primitive type over one"
		if strings {
			what = "a single numeric kind, string, or a named primitive type over one"
		}
		c.errorf(pos, "%s applies to %s, not to %s; narrow first", op, what, types.String(l))
		return false
	}
	return true
}

// nullish checks a ?? b: the left side must be nullable; the result is the left side without
// null joined with the right side. A never on the right narrows a local left side.
func (c *Checker) nullish(x *ast.Binary, sc *scope, expected types.Type) types.Type {
	l := c.synth(x.X, sc)
	if !types.IsInvalid(l) && !types.Includes(l, types.Null{}) {
		c.errorf(x.Pos, "?? needs a nullable left side, got %s", types.String(l))
	}
	nonNull := types.Without(l, types.Null{})
	var r types.Type
	if expected != nil {
		r = c.check(x.Y, sc, expected)
	} else {
		r = c.synth(x.Y, sc)
	}
	if types.IsNever(r) {
		if id, ok := x.X.(*ast.Ident); ok {
			if l, _ := c.lookup(sc, id.Name); l != nil {
				c.narrowLocal(l, nonNull)
			}
		}
		return nonNull
	}
	u, err := types.MakeUnion([]types.Type{nonNull, r})
	if err != nil {
		c.errorf(x.Pos, "the sides of ?? do not form a valid type: %s", err)
		return types.Invalid{}
	}
	return u
}

func (c *Checker) tryExpr(x *ast.Unary, sc *scope) types.Type {
	t := c.synth(x.X, sc)
	if types.IsInvalid(t) {
		return t
	}
	if _, isParam := types.Expand(t).(types.Param); isParam {
		c.errorf(x.Pos, "try needs a Result, not a value of the type parameter %s", types.String(t))
		return types.Invalid{}
	}
	okT, errT, ok := resultParts(t)
	if !ok {
		c.errorf(x.Pos, "try needs a Result, got %s", types.String(t))
		return types.Invalid{}
	}
	fn := c.fn
	if fn == nil || fn.kind == "go" || fn.kind == "defer" || fn.kind == "init" {
		c.errorf(x.Pos, "try needs an enclosing function that returns a Result")
		return okT
	}
	errRec := types.Named{Decl: c.prelude["Err"].decl, Args: []types.Type{errT}}
	if fn.declared != nil {
		_, declaredErr, isResult := resultParts(fn.declared)
		if !isResult || !types.Assignable(errT, declaredErr) {
			c.errorf(x.Pos, "try would return %s, but the function returns %s", types.String(errRec), types.String(fn.declared))
		}
	} else {
		fn.returns = append(fn.returns, errRec)
	}
	return okT
}

// resultParts splits a Result[T, E] into T and E.
func resultParts(t types.Type) (okT, errT types.Type, ok bool) {
	var okTypes, errTypes []types.Type
	for _, m := range types.Members(t) {
		n, isNamed := types.Expand(m).(types.Named)
		if !isNamed || len(n.Args) != 1 {
			return nil, nil, false
		}
		switch n.Decl.Name {
		case "Ok":
			okTypes = append(okTypes, n.Args[0])
		case "Err":
			errTypes = append(errTypes, n.Args[0])
		default:
			return nil, nil, false
		}
	}
	if len(okTypes) == 0 && len(errTypes) == 0 {
		return nil, nil, false
	}
	okT, _ = types.MakeUnion(okTypes)
	errT, _ = types.MakeUnion(errTypes)
	return okT, errT, true
}

// ---- is

func (c *Checker) isExpr(x *ast.Is, sc *scope) types.Type {
	t := c.synth(x.X, sc)
	target := c.isTarget(x, sc)
	if target == nil || types.IsInvalid(t) {
		return types.BoolType
	}
	declared := t
	if id, ok := x.X.(*ast.Ident); ok {
		if l, _ := c.lookup(sc, id.Name); l != nil {
			declared = l.typ
		}
	}
	if c.alwaysTrue(declared, target, x.Pos) {
		return types.BoolType
	}
	c.narrowTo(t, target, x.Pos)
	return types.BoolType
}

// isTarget resolves the right side of is: a kind, or a named record or primitive type.
func (c *Checker) isTarget(x *ast.Is, sc *scope) types.Type {
	if x.Kind != "" {
		switch x.Kind {
		case "null":
			return types.Null{}
		case "void":
			return types.Void{}
		case "func":
			return types.Func{}
		case "Array":
			return types.Array{}
		case "Map":
			return types.Map{}
		}
		return types.Prim{Kind: builtinKinds[x.Kind]}
	}
	ent := c.entityPath(x.Path, sc)
	if ent == nil || ent.kind != entType {
		c.errorf(x.Pos, "the right side of is must be a type")
		return nil
	}
	switch ent.decl.Kind {
	case types.AliasDecl:
		c.errorf(x.Pos, "%s is an alias and has no run-time identity; test its members instead", ent.decl.Name)
		return nil
	case types.OpaqueDecl:
		c.errorf(x.Pos, "is cannot test %s; narrow by testing the other members", ent.decl.Name)
		return nil
	}
	return types.Named{Decl: ent.decl}
}

// narrowTo computes the types of a value when x is T holds and when it does not. target has
// no type arguments; the member of t supplies them.
func (c *Checker) narrowTo(t, target types.Type, pos ast.Pos) (whenTrue, whenFalse types.Type, ok bool) {
	members := types.Members(t)
	var yes, no []types.Type
	for _, m := range members {
		if matchesTarget(m, target) {
			yes = append(yes, m)
		} else {
			no = append(no, m)
		}
	}
	// an anonymous record type admits a named record type assignable to it
	if len(yes) == 0 {
		if n, isNamed := target.(types.Named); isNamed && n.Decl.Kind == types.RecordDecl {
			if r, isRecord := types.Expand(t).(types.Record); isRecord {
				inst, found := c.instantiateFor(n.Decl, r, pos)
				if !found {
					return nil, nil, false
				}
				return inst, t, true
			}
		}
		if p, isParam := types.Expand(t).(types.Param); isParam && p.Decl.Bound != nil {
			for _, m := range types.Members(p.Decl.Bound) {
				if matchesTarget(m, target) {
					return m, t, true
				}
			}
		}
		c.errorf(pos, "%s is never a %s, so the test is always false", types.String(t), types.String(target))
		return nil, nil, false
	}
	wt, _ := types.MakeUnion(yes)
	wf, _ := types.MakeUnion(no)
	return wt, wf, true
}

// alwaysTrue reports a test that cannot fail: the value's declared type has no other member.
func (c *Checker) alwaysTrue(declared, target types.Type, pos ast.Pos) bool {
	for _, m := range types.Members(declared) {
		if !matchesTarget(m, target) {
			return false
		}
	}
	c.errorf(pos, "the test is always true: %s is already %s", types.String(declared), targetName(target))
	return true
}

// targetName names the target of an is test: a kind word for func, Array and Map, which carry
// no type arguments there, and the type otherwise.
func targetName(target types.Type) string {
	switch target.(type) {
	case types.Func:
		return "func"
	case types.Array:
		return "Array"
	case types.Map:
		return "Map"
	}
	return types.String(target)
}

func matchesTarget(m, target types.Type) bool {
	switch tg := target.(type) {
	case types.Named:
		n, ok := types.Expand(m).(types.Named)
		return ok && n.Decl == tg.Decl
	case types.Prim:
		p, ok := types.Expand(m).(types.Prim)
		return ok && p.Kind == tg.Kind
	case types.Null:
		_, ok := types.Expand(m).(types.Null)
		return ok
	case types.Void:
		_, ok := types.Expand(m).(types.Void)
		return ok
	case types.Func:
		_, ok := types.Expand(m).(types.Func)
		return ok
	case types.Array:
		_, ok := types.Expand(m).(types.Array)
		return ok
	case types.Map:
		_, ok := types.Expand(m).(types.Map)
		return ok
	}
	return false
}

// instantiateFor finds the one instantiation of a named record type assignable to an
// anonymous record type, when its fields determine every type argument.
func (c *Checker) instantiateFor(decl *types.Decl, r types.Record, pos ast.Pos) (types.Type, bool) {
	subst := map[*types.ParamDecl]types.Type{}
	for _, want := range r.Fields {
		idx := fieldIndex(decl, want.Name)
		if idx < 0 {
			c.errorf(pos, "%s has no field %s, so it is never a %s", decl.Name, want.Name, types.String(r))
			return nil, false
		}
		if !c.unify(decl.Fields[idx].Type, want.Type, subst) {
			c.errorf(pos, "the field %s of %s does not match %s", want.Name, decl.Name, types.String(r))
			return nil, false
		}
	}
	var args []types.Type
	for _, p := range decl.Params {
		t, ok := subst[p]
		if !ok {
			c.errorf(pos, "%s does not determine the type parameter %s of %s", types.String(r), p.Name, decl.Name)
			return nil, false
		}
		args = append(args, t)
	}
	return types.Named{Decl: decl, Args: args}, true
}

// ---- assignment

func (c *Checker) assign(x *ast.Assign, sc *scope) types.Type {
	op := x.Op[:len(x.Op)-1]
	switch t := x.Target.(type) {
	case *ast.Ident:
		if t.Name == "_" {
			c.synth(x.Value, sc)
			return types.Void{}
		}
		l, ent := c.lookup(sc, t.Name)
		switch {
		case l != nil:
			if x.Op == "=" {
				v := c.check(x.Value, sc, l.typ)
				c.narrowLocal(l, v)
			} else {
				cur := c.typeOfLocal(l)
				v := c.synth(x.Value, sc)
				c.sameOperands(cur, v, op, op == "+", x.Pos)
			}
		case ent == nil:
			c.errorf(t.Pos, "%s is not defined", t.Name)
			c.synth(x.Value, sc)
		case ent.kind == entVar:
			if ent.v.mod != c.mod {
				c.errorf(t.Pos, "cannot assign to the imported variable %s", t.Name)
			}
			vt := c.varType(ent.v, t.Pos, c.fn != nil && c.fn.kind == "init")
			if x.Op == "=" {
				c.check(x.Value, sc, vt)
			} else {
				c.sameOperands(vt, c.synth(x.Value, sc), op, op == "+", x.Pos)
			}
		default:
			c.errorf(t.Pos, "cannot assign to %s", t.Name)
			c.synth(x.Value, sc)
		}
	case *ast.Field:
		if base := c.entityPath(t.X, sc); base != nil && base.kind == entModule {
			c.errorf(t.Pos, "cannot assign to %s.%s: imports are read-only", base.name, t.Name)
			c.synth(x.Value, sc)
			return types.Void{}
		}
		recv := c.synth(t.X, sc)
		if _, isUnion := types.Expand(recv).(types.Union); isUnion {
			c.errorf(t.Pos, "cannot assign a field through the union %s; narrow first", types.String(recv))
			c.synth(x.Value, sc)
			return types.Void{}
		}
		ft := c.fieldOf(recv, t.Name, t.Pos)
		if x.Op == "=" {
			c.check(x.Value, sc, ft)
		} else {
			c.sameOperands(ft, c.synth(x.Value, sc), op, op == "+", x.Pos)
		}
	case *ast.Index:
		recv := c.synth(t.X, sc)
		switch r := types.Expand(recv).(type) {
		case types.Array:
			c.check(t.Index, sc, types.IntType)
			if x.Op == "=" {
				c.check(x.Value, sc, r.Elem)
			} else {
				c.sameOperands(r.Elem, c.synth(x.Value, sc), op, op == "+", x.Pos)
			}
		case types.Map:
			c.check(t.Index, sc, r.Key)
			if x.Op == "=" {
				c.check(x.Value, sc, r.Val)
			} else {
				c.sameOperands(r.Val, c.synth(x.Value, sc), op, op == "+", x.Pos)
			}
		case types.Invalid:
			c.synth(t.Index, sc)
			c.synth(x.Value, sc)
		default:
			c.synth(t.Index, sc)
			c.synth(x.Value, sc)
			c.errorf(t.Pos, "%s cannot be indexed", types.String(recv))
		}
	case *ast.AssignArray:
		v := c.synth(x.Value, sc)
		c.assignArray(t, v, sc)
	}
	return types.Void{}
}

func (c *Checker) assignArray(t *ast.AssignArray, v types.Type, sc *scope) {
	a, ok := types.Expand(v).(types.Array)
	if !ok {
		if !types.IsInvalid(v) {
			c.errorf(t.Pos, "a destructuring assignment needs an array, got %s", types.String(v))
		}
		a = types.Array{Elem: types.Invalid{}}
	}
	for _, el := range t.Elems {
		switch p := el.(type) {
		case *ast.AssignArray:
			c.assignArray(p, a.Elem, sc)
		case ast.Expr:
			c.assignPlace(p, a.Elem, sc)
		}
	}
	if t.Rest != nil {
		c.assignPlace(t.Rest, types.Array{Elem: a.Elem}, sc)
	}
}

// assignPlace checks a store of a value of type v into a place.
func (c *Checker) assignPlace(p ast.Expr, v types.Type, sc *scope) {
	switch t := p.(type) {
	case *ast.Ident:
		if t.Name == "_" {
			return
		}
		l, ent := c.lookup(sc, t.Name)
		switch {
		case l != nil:
			c.expect(t.Pos, v, l.typ)
			c.narrowLocal(l, v)
		case ent != nil && ent.kind == entVar:
			if ent.v.mod != c.mod {
				c.errorf(t.Pos, "cannot assign to the imported variable %s", t.Name)
			}
			c.expect(t.Pos, v, c.varType(ent.v, t.Pos, false))
		default:
			c.errorf(t.Pos, "cannot assign to %s", t.Name)
		}
	case *ast.Field:
		recv := c.synth(t.X, sc)
		c.expect(t.Pos, v, c.fieldOf(recv, t.Name, t.Pos))
	case *ast.Index:
		recv := c.synth(t.X, sc)
		switch r := types.Expand(recv).(type) {
		case types.Array:
			c.check(t.Index, sc, types.IntType)
			c.expect(t.Pos, v, r.Elem)
		case types.Map:
			c.check(t.Index, sc, r.Key)
			c.expect(t.Pos, v, r.Val)
		default:
			c.synth(t.Index, sc)
			if !types.IsInvalid(recv) {
				c.errorf(t.Pos, "%s cannot be indexed", types.String(recv))
			}
		}
	}
}

// regexPattern is the type of a regular expression literal: Regex.Pattern of the prelude.
func (c *Checker) regexPattern() types.Type {
	if ns, ok := c.prelude["Regex"]; ok {
		if p, ok := ns.members["Pattern"]; ok && p.decl != nil {
			return types.Named{Decl: p.decl}
		}
	}
	return types.Invalid{}
}
