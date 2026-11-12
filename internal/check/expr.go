package check

import (
	"fmt"
	"maps"
	"slices"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// lookup finds a name: a local or a block's type in the scope chain, else a module-level or
// prelude entity.
func (c *Checker) lookup(sc *scope, name string) (*local, *entity) {
	for s := sc; s != nil; s = s.parent {
		if s.names != nil {
			if l, ok := s.names[name]; ok {
				return l, nil
			}
		}
		if e, ok := s.types[name]; ok {
			return nil, e
		}
		if _, ok := s.tparams[name]; ok {
			// A type parameter hides an outer name of its spelling throughout its
			// declaration, and binds nothing a value, a literal head, an is target, or a
			// path can use (section 3).
			return nil, nil
		}
	}
	return nil, c.lookupEntity(name)
}

// hiddenX reports the x of a lambda of arity zero, which is not in scope.
func (c *Checker) hiddenX(l *local, pos ast.Pos) bool {
	if l.hidden {
		c.errorf(pos, "x is not defined: a lambda that takes no argument has no x")
	}
	return l.hidden
}

// typeOfLocal is a local's current type: narrowed in the flow, else declared.
func (c *Checker) typeOfLocal(l *local) types.Type {
	if t, ok := c.flow[l]; ok {
		return t
	}
	return l.typ
}

// check checks e against an expected type and returns e's type.
func (c *Checker) check(e ast.Expr, sc *scope, expected types.Type) types.Type {
	if expected == nil {
		return c.synth(e, sc)
	}
	t := c.checkExpr(e, sc, expected)
	c.info.recordType(e, t)
	return t
}

func (c *Checker) checkExpr(e ast.Expr, sc *scope, expected types.Type) types.Type {
	var t types.Type
	switch x := e.(type) {
	case *ast.Lambda:
		return c.checkLambda(x, sc, expected)
	case *ast.ArrayLit:
		if a, ok := types.MemberOf[types.Array](expected); ok {
			t = c.checkArrayLit(x, sc, a.Elem)
		}
	case *ast.MapLit:
		if m, ok := types.MemberOf[types.Map](expected); ok {
			t = c.checkMapLit(x, sc, m)
		}
	case *ast.RecordLit:
		t = c.recordLit(x, sc, expected)
	case *ast.If:
		t = c.checkIf(x, sc, expected)
	case *ast.Select:
		t = c.checkSelect(x, sc, expected)
	case *ast.Call:
		t = c.call(x, sc, expected)
	case *ast.Block:
		t = c.block(x, sc, expected)
	case *ast.Binary:
		if x.Op == "??" {
			t = c.nullish(x, sc, expected)
		}
	case *ast.Ident, *ast.Field:
		if c.typeNameForTypeValue(x, sc, expected) {
			return types.Invalid{}
		}
		t = c.genericValue(x, sc, expected)
	}
	if t == nil {
		t = c.synth(e, sc)
	}
	c.expectIn(e, t, expected)
	return t
}

// notePath records what the parts of an entity path refer to, which synth never sees.
func (c *Checker) notePath(e ast.Expr, sc *scope) {
	if c.info == nil {
		return
	}
	switch x := e.(type) {
	case *ast.Ident:
		if _, ent := c.lookup(sc, x.Name); ent != nil {
			c.info.recordIdent(x, c.info.object(ent))
		}
	case *ast.Field:
		if ent := c.entityPath(x, sc); ent != nil {
			c.info.recordMember(x, c.info.object(ent))
		}
		c.notePath(x.X, sc)
	}
}

// typeNameForTypeValue reports a type name written where a type value is expected, a primitive's
// conversion function and a named primitive's constructor included: the value is type(T)
// (section 4, Types as values).
func (c *Checker) typeNameForTypeValue(e ast.Expr, sc *scope, expected types.Type) bool {
	want := false
	for _, m := range types.Members(expected) {
		if n, ok := types.Expand(m).(types.Named); ok && n.Decl != nil && (n.Decl == c.reflectDecl("Type") || n.Decl == c.reflectDecl("ErasedType")) {
			want = true
		}
	}
	if !want {
		return false
	}
	ent := c.entityPath(e, sc)
	if ent == nil || !(ent.kind == entType || ent.kind == entFunc && ent.isConv) {
		return false
	}
	c.notePath(e, sc)
	name := pathName(e)
	c.errorf(e.Position(), "%s is a type name; write type(%s) for its value", name, name)
	return true
}

// pathName is the written text of a dotted name.
func pathName(e ast.Expr) string {
	if f, ok := e.(*ast.Field); ok {
		return pathName(f.X) + "." + f.Name
	}
	return e.(*ast.Ident).Name
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
	default:
		if t := convValue(ent, f); t != nil {
			return t
		}
		if t := c.lenValue(ent, f, e.Position()); t != nil {
			return t
		}
		s = c.sigOf(ent, e.Position())
	}
	if s == nil || len(s.tparams) == 0 {
		return nil
	}
	subst := map[*types.ParamDecl]types.Type{}
	if !c.unifyFuncValue(s, f, subst) {
		c.errorf(e.Position(), "cannot use %s where %s is expected", s.fullName(), types.String(expected))
		return types.Invalid{}
	}
	for _, p := range s.tparams {
		if _, bound := subst[p]; !bound {
			c.errorf(e.Position(), "%s does not fix %s of %s", types.String(expected), p.Name, s.fullName())
			return types.Invalid{}
		}
	}
	c.checkBounds(s, subst, e)
	inst, err := types.Subst(s.fn, subst)
	if err != nil {
		c.errorf(e.Position(), "%s", err)
		return types.Invalid{}
	}
	return inst
}

// unifyFuncValue solves a generic function's type parameters from the expected function type:
// from its parameters first, then from its result for those still unsolved, since the result
// need only be assignable to the expected one, which the caller checks. ident as a
// func(Circle): Shape is ident at Circle.
func (c *Checker) unifyFuncValue(s *sig, f types.Func, subst map[*types.ParamDecl]types.Type) bool {
	if len(s.fn.Params) != len(f.Params) {
		return c.unify(s.fn, f, subst)
	}
	for i, p := range s.fn.Params {
		if !c.unify(p, f.Params[i], subst) {
			return false
		}
	}
	unsolved := map[*types.ParamDecl]bool{}
	for _, p := range s.tparams {
		if _, ok := subst[p]; !ok {
			unsolved[p] = true
		}
	}
	if !types.Mentions(s.fn.Return, unsolved) {
		return true
	}
	return c.unify(s.fn.Return, f.Return, subst)
}

// convValue types a conversion used as a value where a function from a named primitive is
// expected: a conversion also takes a named primitive over a kind it accepts, and unwraps it
// (section 11, Prelude), as a value as in a call. It returns nil otherwise.
func convValue(ent *entity, f types.Func) types.Type {
	if !ent.isConv || len(f.Params) != 1 {
		return nil
	}
	n, isNamed := types.Expand(f.Params[0]).(types.Named)
	if !isNamed || n.Decl.Kind != types.PrimitiveDecl {
		return nil
	}
	conv := ent.conversion
	if n.Decl.Base != conv && (conv == value.KindString || conv == value.KindBool || !n.Decl.Base.IsNumber()) {
		return nil
	}
	return types.Func{Params: []types.Type{f.Params[0]}, Return: types.Prim{Kind: conv}, Required: 1}
}

// expect reports a type that is not assignable to the expected one.
func (c *Checker) expect(pos ast.Pos, t, expected types.Type) {
	if msg := c.misfit(t, expected); msg != "" {
		c.errorf(pos, "%s", msg)
	}
}

// expectIn is expect for the expression e, reported at its start and spanning it.
func (c *Checker) expectIn(e ast.Expr, t, expected types.Type) {
	if msg := c.misfit(t, expected); msg != "" {
		c.errorIn(e, "%s", msg)
	}
}

// misfit is the error for a value of type t where expected is expected, or "" when it fits.
func (c *Checker) misfit(t, expected types.Type) string {
	if types.Assignable(t, expected) || hasHole(t) || hasHole(expected) {
		return "" // a type naming an unknown type was reported where it was written
	}
	n := c.typeNames(t, expected)
	why := ""
	if iface, ok := types.Expand(expected).(types.Named); ok && iface.Decl.Kind == types.InterfaceDecl {
		if r := c.unsatisfied(t, iface); r != "" {
			why = ": " + r
		}
	}
	return fmt.Sprintf("cannot use %s where %s is expected%s", n[0], n[1], why)
}

// typeNames renders types for one message. A name that stands for two different types among
// them, as two types of one name declared in different blocks or modules do, says where each is
// declared.
func (c *Checker) typeNames(ts ...types.Type) []string {
	return types.Strings(c.declaredAt, ts...)
}

func (c *Checker) declaredAt(d *types.Decl) string {
	pos, ok := d.Pos.(ast.Pos)
	switch {
	case !ok || pos.Line == 0:
		return ""
	case c.mod != nil && d.Path != c.mod.path:
		return fmt.Sprintf(" (declared at %s:%d:%d)", d.Path, pos.Line, pos.Col)
	}
	return fmt.Sprintf(" (declared at %d:%d)", pos.Line, pos.Col)
}

// synth determines the type of an expression without an expected type.
func (c *Checker) synth(e ast.Expr, sc *scope) types.Type {
	t := c.synthExpr(e, sc)
	c.info.recordType(e, t)
	return t
}

func (c *Checker) synthExpr(e ast.Expr, sc *scope) types.Type {
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
	case *ast.Html:
		return c.htmlExpr(x, sc)
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
	case *ast.TypeExpr:
		return c.typeExpr(x, sc)
	case *ast.Assign:
		return c.assign(x, sc)
	case *ast.If:
		return c.checkIf(x, sc, nil)
	case *ast.Select:
		return c.checkSelect(x, sc, nil)
	case *ast.While:
		return c.whileLoop(x, sc)
	case *ast.For:
		return c.forLoop(x, sc)
	case *ast.Go:
		c.nestedBody(x.Body, sc, "go")
		return types.Void{}
	case *ast.Defer:
		c.nestedBody(x.Body, sc, "defer")
		return types.Void{}
	case *ast.Test:
		c.testBody(x, sc)
		return types.Void{}
	case *ast.Return:
		return c.returnExpr(x, sc)
	case *ast.Break:
		c.markBreak()
		return types.Never{}
	case *ast.Continue:
		return types.Never{}
	case *ast.Block:
		return c.block(x, sc, nil)
	case *ast.Placeholder, *ast.BadExpr:
		return types.Invalid{}
	}
	c.errorf(e.Position(), "internal error: cannot check %T", e)
	return types.Invalid{}
}

// ---- names

func (c *Checker) ident(x *ast.Ident, sc *scope) types.Type {
	if x.Ref.Kind == ast.RefDiscard {
		c.errorf(x.Pos, "_ is write-only and cannot be read")
		return types.Invalid{}
	}
	l, ent := c.lookup(sc, x.Name)
	if l != nil {
		c.info.recordIdent(x, c.info.localObject(l, c.mod.path))
		if c.hiddenX(l, x.Pos) {
			return types.Invalid{}
		}
		if l.unset && l.ctx == c.fn {
			c.errorf(x.Pos, "%s is a parameter whose default has not run yet; a default may read only the parameters before it and those without a default", x.Name)
			return types.Invalid{}
		}
		if l.pending {
			if l.ctx == c.fn {
				c.errorf(x.Pos, "%s is read in its own initializer", x.Name)
				return types.Invalid{}
			}
			if !l.stated {
				if !l.selfReported {
					l.selfReported = true
					c.errorf(x.Pos, "%s is used by a function in its own initializer; state its type, as a type on the binding or a return type on the func", x.Name)
				}
				return types.Invalid{}
			}
		}
		return c.typeOfLocal(l)
	}
	if ent == nil {
		if p, _ := sc.lookupTypeParam(x.Name); p != nil {
			c.errorf(x.Pos, "%s is a type parameter, not a value", x.Name)
			return types.Invalid{}
		}
		c.errorf(x.Pos, "%s is not defined", x.Name)
		return types.Invalid{}
	}
	c.info.recordIdent(x, c.info.object(ent))
	t := c.entityValue(ent, x.Pos)
	if ent.kind == entVar && ent.v.local != nil && !types.IsInvalid(t) {
		// a top-level or imported let narrowed by a test (section 7)
		return c.typeOfLocal(ent.v.local)
	}
	return t
}

// entityValue is the type of an entity used as a value.
func (c *Checker) entityValue(ent *entity, pos ast.Pos) types.Type {
	switch ent.kind {
	case entVar:
		return c.varType(ent.v, pos, c.fn != nil && (c.fn.kind == "init" || c.fn.kind == "script"))
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
			// type Id[T] = string: only an expected function type fixes T (genericValue).
			fn := constructorType(ent.decl)
			if len(ent.decl.Params) == 0 {
				return fn
			}
			args := make([]types.Type, len(ent.decl.Params))
			for i, p := range ent.decl.Params {
				c.errorf(pos, "cannot infer the type parameter %s of %s; annotate the result", p.Name, ent.decl.Name)
				args[i] = types.Invalid{}
			}
			fn.Return = types.Named{Decl: ent.decl, Args: args}
			return fn
		}
		c.errorf(pos, "%s is a type, not a value; a type appears only as a literal head, after is, or before a dot", ent.decl.Name)
		return types.Invalid{}
	case entModule, entNamespace:
		c.errorf(pos, "%s is a namespace, not a value", ent.name)
		return types.Invalid{}
	case entMissing:
		return types.Invalid{}
	}
	return types.Invalid{}
}

// constructorType is the function type of a named primitive type's constructor, with the type's
// own parameters free.
func constructorType(decl *types.Decl) types.Func {
	var args []types.Type
	for _, p := range decl.Params {
		args = append(args, types.Param{Decl: p})
	}
	return types.Func{Params: []types.Type{types.Prim{Kind: decl.Base}}, Return: types.Named{Decl: decl, Args: args}, Required: 1}
}

// ctorSig is the signature of a generic named primitive type's constructor, which is generic in
// the type's parameters, or nil for any other type.
func ctorSig(decl *types.Decl) *sig {
	if decl.Kind != types.PrimitiveDecl || len(decl.Params) == 0 {
		return nil
	}
	return &sig{tparams: decl.Params, fn: constructorType(decl), name: decl.Name}
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
		l, ent := c.lookup(sc, x.Name)
		if l != nil {
			return nil
		}
		return ent
	case *ast.Field:
		if x.Optional {
			return nil
		}
		base := c.entityPath(x.X, sc)
		if base == nil {
			return nil
		}
		if base.kind == entMissing {
			return base
		}
		if base.kind == entModule {
			return base.members[x.Name]
		}
		if base.members == nil {
			return nil
		}
		m, ok := base.members[x.Name]
		if !ok || c.foreignPrivate(m) {
			return nil
		}
		return m
	}
	return nil
}

// foreignPrivate reports a function declared without export on a type of another module,
// which is a private field of that module (section 10) that only a test file sees (Tests).
func (c *Checker) foreignPrivate(m *entity) bool {
	return m.kind == entFunc && !m.exported && m.fn != nil && m.fn.mod != c.mod && !c.mod.test
}

// moduleMember reports a member of an imported module, or the reason it is not visible.
func (c *Checker) moduleMember(base *entity, name string, pos ast.Pos) *entity {
	if m, ok := base.members[name]; ok {
		return m
	}
	if _, ok := base.mod.names[name]; ok {
		c.errorf(pos, "%s is not exported by %s", name, base.name)
	} else {
		c.errorf(pos, "%s has no declaration %s", base.name, name)
	}
	return nil
}

func (c *Checker) field(x *ast.Field, sc *scope) types.Type {
	if base := c.entityPath(x.X, sc); base != nil && !x.Optional {
		switch base.kind {
		case entMissing:
			return types.Invalid{}
		case entModule:
			c.notePath(x.X, sc)
			m := c.moduleMember(base, x.Name, x.Pos)
			if m == nil {
				return types.Invalid{}
			}
			c.info.recordMember(x, c.info.object(m))
			return c.entityValue(m, x.Pos)
		case entNamespace, entType:
			c.notePath(x.X, sc)
			if base.kind == entType && base.decl.Kind == types.InterfaceDecl {
				c.errorf(x.Pos, "%s is an interface, not a namespace: call %s on a value, as x.%s(...)", base.name, x.Name, x.Name)
				return types.Invalid{}
			}
			m, ok := base.members[x.Name]
			if !ok {
				if base.kind == entType && base.decl.Kind == types.PrimitiveDecl {
					break
				}
				c.errorf(x.Pos, "%s has no member %s", base.name, x.Name)
				return types.Invalid{}
			}
			if c.foreignPrivate(m) {
				c.errorf(x.Pos, "%s.%s is not exported", base.name, x.Name)
				return types.Invalid{}
			}
			c.info.recordMember(x, c.info.object(m))
			return c.entityValue(m, x.Pos)
		}
	}
	recv, short := c.chainRecv(x.X, sc)
	if x.Optional {
		if !types.Includes(recv, types.Null{}) {
			c.errorf(x.Pos, "?. needs a nullable left side, got %s%s", types.String(recv), chainNote(short, "."))
			return types.Invalid{}
		}
		c.info.recordMember(x, c.info.fieldOf(types.Without(recv, types.Null{}), x.Name))
		inner := c.fieldOf(types.Without(recv, types.Null{}), x.Name, x.Pos, nil)
		return c.endLink(x, x.Grouped, inner, true)
	}
	c.info.recordMember(x, c.info.fieldOf(recv, x.Name))
	return c.endLink(x, x.Grouped, c.fieldOf(recv, x.Name, x.Pos, func() string { return c.outerVarHint(x.X, sc) }), short)
}

// chainNote explains a ?. on a left side that is not nullable when a ?. earlier in the chain
// already skips the rest of it on null (section 7), so that the plain link, written as next, is
// what is meant.
func chainNote(short bool, next string) string {
	if !short {
		return ""
	}
	return fmt.Sprintf(": an earlier ?. in the chain already skips the rest of it on null, so write %s here", next)
}

// chainRecv synthesizes the left side of a postfix link. When that side is a link a ?. earlier
// in the same chain may skip, a null there skips this link too (section 12), so the link reads
// the type the left side has when nothing was null, and the second result reports that the
// chain's null joins this link's type.
func (c *Checker) chainRecv(e ast.Expr, sc *scope) (types.Type, bool) {
	t := c.synth(e, sc)
	if inner, ok := c.shorted[e]; ok {
		return inner, true
	}
	return t, false
}

// endLink gives a postfix link its type: inner, or inner | null when a ?. in its chain may
// skip it. Parentheses end the chain, so a grouped link passes on only the union.
func (c *Checker) endLink(x ast.Expr, grouped bool, inner types.Type, short bool) types.Type {
	if !short {
		return inner
	}
	if !grouped {
		c.shorted[x] = inner
	}
	return types.Nullable(inner)
}

// outerVarHint explains why x did not narrow when it names a var of an enclosing body read in a
// nested one (section 7), for an error that asks to narrow; it is "" otherwise.
func (c *Checker) outerVarHint(x ast.Expr, sc *scope) string {
	id, ok := x.(*ast.Ident)
	if !ok {
		return ""
	}
	l, _ := c.lookup(sc, id.Name)
	if l == nil || !c.outerVar(l) {
		return ""
	}
	return fmt.Sprintf("; %s is a var of an enclosing body, which does not narrow here, so bind it to a let first", id.Name)
}

// fieldOf is the type of a field read through a type: a record, named or anonymous, or a
// union of records that all declare it.
func (c *Checker) fieldOf(recv types.Type, name string, pos ast.Pos, hint func() string) types.Type {
	note := func() string {
		if hint == nil {
			return ""
		}
		return hint()
	}
	if types.IsInvalid(recv) {
		return types.Invalid{}
	}
	switch r := types.Expand(recv).(type) {
	case types.Union:
		var members []types.Type
		for _, m := range r.Members {
			fields := types.FieldsOf(m)
			if fields == nil {
				c.errorf(pos, "%s has no field %s: narrow the value first%s", types.String(recv), name, note())
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
				c.errorf(pos, "%s has no field %s on every member: narrow the value first%s", types.String(recv), name, note())
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
	recv, short := c.chainRecv(x.X, sc)
	if x.Optional {
		if !types.Includes(recv, types.Null{}) {
			c.errorf(x.Pos, "?.[ needs a nullable left side, got %s%s", types.String(recv), chainNote(short, "["))
			return types.Invalid{}
		}
		inner := c.indexOf(types.Without(recv, types.Null{}), x, sc)
		return c.endLink(x, x.Grouped, inner, true)
	}
	return c.endLink(x, x.Grouped, c.indexOf(recv, x, sc), short)
}

func (c *Checker) indexOf(recv types.Type, x *ast.Index, sc *scope) types.Type {
	switch r := types.Expand(recv).(type) {
	case types.Array:
		c.check(x.Index, sc, types.IntType)
		return r.Elem
	case types.Map:
		c.check(x.Index, sc, r.Key)
		return types.Nullable(r.Val)
	case types.Invalid:
		// The operand was reported. A name inside, as int in Array[int](), may be meant as a
		// type argument, so it is left alone; any other index is still checked.
		if _, isName := x.Index.(*ast.Ident); !isName {
			c.synth(x.Index, sc)
		}
		return types.Invalid{}
	case types.Named:
		if c.isBytes(r) {
			c.check(x.Index, sc, types.IntType)
			return types.ByteType
		}
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

// undefined reports a name that nothing in scope binds, a type parameter included.
func (c *Checker) undefined(name string, sc *scope) bool {
	if p, _ := sc.lookupTypeParam(name); p != nil {
		return false
	}
	l, ent := c.lookup(sc, name)
	return l == nil && ent == nil
}

// namespaceHint suggests the prelude type a missing type name may mean, as Panic.Frame for Frame.
func (c *Checker) namespaceHint(name string) string {
	for _, ns := range slices.Sorted(maps.Keys(c.prelude)) {
		if m, ok := c.prelude[ns].members[name]; ok && m.kind == entType {
			return "; did you mean " + ns + "." + name + "?"
		}
	}
	return ""
}

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
	if head != nil && head.kind == entMissing {
		for _, en := range x.Entries {
			if en.Spread != nil {
				c.synth(en.Spread, sc)
			} else {
				c.synth(en.Value, sc)
			}
		}
		return types.Invalid{}
	}
	if head == nil || head.kind != entType {
		if id, ok := x.Head.(*ast.Ident); ok && head == nil && c.undefined(id.Name, sc) {
			c.errorf(x.Pos, "unknown type %s%s", id.Name, c.namespaceHint(id.Name))
		} else {
			c.errorf(x.Pos, "a record literal needs a record type as its head")
		}
		for _, en := range x.Entries {
			if en.Spread != nil {
				c.synth(en.Spread, sc)
			} else {
				c.synth(en.Value, sc)
			}
		}
		return types.Invalid{}
	}
	c.notePath(x.Head, sc)
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
				subst = types.Bindings(decl.Params, n.Args)
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
	// A later entry wins (section 12), so a spread's field that an entry after it gives again
	// does not constrain the literal's type. Named entries are known up front; the fields of a
	// later spread only once it is checked, so a field a later spread may give is matched
	// after the loop, when the last spread to give it is known.
	later := make([]map[string]bool, len(x.Entries))
	laterSpread := make([]bool, len(x.Entries))
	names := map[string]bool{}
	spreadAfter := false
	for i := len(x.Entries) - 1; i >= 0; i-- {
		later[i] = names
		laterSpread[i] = spreadAfter
		if x.Entries[i].Spread == nil {
			names = maps.Clone(names)
			names[x.Entries[i].Name] = true
		} else {
			spreadAfter = true
		}
	}
	type spreadField struct {
		entry int
		want  types.Type
		got   types.Type
		pos   ast.Pos
		name  string
	}
	var deferred []spreadField
	lastSpread := map[string]int{}
	for i, en := range x.Entries {
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
				lastSpread[f.Name] = i
				switch {
				case later[i][f.Name]:
				case laterSpread[i]:
					deferred = append(deferred, spreadField{i, decl.Fields[idx].Type, f.Type, en.Pos, f.Name})
				default:
					c.matchField(decl.Fields[idx].Type, f.Type, subst, unsolved, en.Pos, f.Name)
				}
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
		if c.info != nil {
			c.info.Entries[&x.Entries[i]] = c.info.field(decl, decl.Fields[idx])
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
	for _, d := range deferred {
		if lastSpread[d.name] == d.entry {
			c.matchField(d.want, d.got, subst, unsolved, d.pos, d.name)
		}
	}
	for _, f := range decl.Fields {
		if !given[f.Name] && !spreadUnknown && !f.Default {
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
	return c.checkInferred(types.Named{Decl: decl, Args: args}, x.Pos)
}

// checkInferred checks the arguments the checker inferred for a generic record as it checks
// written ones, since they may form an invalid union in a field (section 4, Unions); an error
// a field already reported is not repeated.
func (c *Checker) checkInferred(n types.Named, pos ast.Pos) types.Type {
	if err := c.checkArgs(n); err != nil {
		c.errorf(pos, "%s: %s", types.String(n), err)
		return types.Invalid{}
	}
	c.checkWrittenBounds(n, pos)
	return n
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
	pruneSolved(unsolved, subst)
}

func (c *Checker) synthArrayLit(x *ast.ArrayLit, sc *scope) types.Type {
	if len(x.Elems) == 0 {
		c.errorf(x.Pos, "cannot infer the element type of an empty array; annotate it")
		return types.Invalid{}
	}
	// The element type is the union of the elements' types. An element that fixes nothing, such
	// as [] or Set.of([]), waits and is then checked against that union.
	var elems []types.Type
	var empties []ast.Expr
	for _, el := range x.Elems {
		if !el.Spread && onlyEmpty(el.Value) {
			empties = append(empties, el.Value)
			continue
		}
		t := c.synth(el.Value, sc)
		if el.Spread {
			a, ok := types.Expand(t).(types.Array)
			if !ok {
				if !types.IsInvalid(t) && !hasHole(t) { // ? in it: an error was reported
					c.errorf(el.Value.Position(), "a spread in an array literal needs an array, got %s", types.String(t))
				}
				// the literal is invalid: no error about its other elements follows
				elems = append(elems, types.Invalid{})
				continue
			}
			t = a.Elem
		}
		elems = append(elems, t)
	}
	if len(elems) == 0 {
		// Nothing to wait for. A call is typed on its own: ints([]) has a type, and Set.of([])
		// reports the empty literal inside it.
		var literals []ast.Expr
		for _, e := range empties {
			if _, isCall := e.(*ast.Call); isCall {
				elems = append(elems, c.synth(e, sc))
			} else {
				literals = append(literals, e)
			}
		}
		if len(elems) == 0 {
			c.errorf(x.Pos, "cannot infer the element type of %s; annotate it", "an array of empty literals")
			return types.Invalid{}
		}
		empties = literals
	}
	// An element that failed was reported; the literal is invalid too, so later uses stay quiet.
	for _, t := range elems {
		if types.IsInvalid(t) {
			return types.Invalid{}
		}
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
				if !types.IsInvalid(t) && !hasHole(t) { // ? in it: an error was reported
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
	// As in an array literal, a key or value that fixes nothing, such as [] or Set.of([]), waits
	// and is then checked against the union of the other keys or values (section 4, Inference).
	var keys, vals []ast.Expr
	for _, en := range x.Entries {
		keys = append(keys, en.Key)
		vals = append(vals, en.Value)
	}
	// the values are checked after keys that failed too, so that their errors are reported
	k, keysOK := c.waitingUnion(x.Pos, keys, sc, "key", "keys")
	v, valsOK := c.waitingUnion(x.Pos, vals, sc, "value", "values")
	if !keysOK || !valsOK {
		return types.Invalid{}
	}
	return types.Map{Key: k, Val: v}
}

// waitingUnion is the union of the types of es, the keys or the values of a map literal. An
// expression that fixes nothing waits and is checked against the union of the others; when all
// of them wait, a call is typed on its own, as Set.of([]) is. It reports false after an error.
func (c *Checker) waitingUnion(pos ast.Pos, es []ast.Expr, sc *scope, what, plural string) (types.Type, bool) {
	var ts []types.Type
	var empties []ast.Expr
	for _, e := range es {
		if onlyEmpty(e) {
			empties = append(empties, e)
			continue
		}
		ts = append(ts, c.synth(e, sc))
	}
	if len(ts) == 0 {
		var literals []ast.Expr
		for _, e := range empties {
			if _, isCall := e.(*ast.Call); isCall {
				ts = append(ts, c.synth(e, sc))
			} else {
				literals = append(literals, e)
			}
		}
		if len(ts) == 0 {
			c.errorf(pos, "cannot infer the %s type of a map of empty literals; annotate it", what)
			return nil, false
		}
		empties = literals
	}
	for _, t := range ts {
		if types.IsInvalid(t) {
			return nil, false
		}
	}
	u, err := types.MakeUnion(ts)
	if err != nil {
		c.errorf(pos, "the %s do not form a valid type: %s; give the map an expected type", plural, err)
		return nil, false
	}
	for _, e := range empties {
		c.check(e, sc, u)
	}
	return u, true
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
	ft, found := types.MemberOf[types.Func](expected)
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
	t := c.lambda(x, sc, ft)
	if ft.Required < len(ft.Params) {
		// A lambda has no default, so a call that leaves x out would find it unbound.
		c.errorf(x.Pos, "cannot use %s where %s is expected", types.String(t), types.String(expected))
	}
	return t
}

// lambda checks a lambda against the function type it must have. When the expected return
// type is an unsolved type parameter, the body's type stands in for it.
func (c *Checker) lambda(x *ast.Lambda, sc *scope, ft types.Func) types.Type {
	ctx := &fnCtx{parent: c.fn, kind: "lambda", name: "lambda", declared: ft.Return}
	if types.MentionsOther(ft.Return, c.rigid) {
		// The expected result mentions a type parameter the call has not solved yet, as U in
		// Array.map's U or Array.flatMap's Array[U], rather than only the enclosing function's
		// own, which is a fixed type inside its body; the body is synthesized and the call
		// unifies the two to solve it (section 4, Inference).
		ctx.declared = nil
	}
	inner := &scope{parent: sc, names: map[string]*local{}}
	if len(ft.Params) == 1 {
		inner.names["x"] = &local{name: "x", typ: ft.Params[0], slot: x.Scope.X, ctx: ctx, pos: x.Pos, node: x}
		if c.info != nil {
			c.info.Lambdas[x] = ft.Params[0]
		}
	} else {
		// At run time x is bound to void, so it must not reach an outer x either.
		inner.names["x"] = &local{name: "x", typ: types.Invalid{}, slot: x.Scope.X, ctx: ctx, pos: x.Pos, hidden: true, node: x}
		if c.info != nil {
			c.info.Lambdas[x] = nil
		}
	}
	c.info.localObject(inner.names["x"], c.mod.path)
	ret := c.withNested(ctx, func() types.Type { return c.inferBody(x.Body, inner, ctx) })
	return types.Func{Params: ft.Params, Return: ret, Required: len(ft.Params)}
}

// lambdaBodyAnyway checks a lambda's body for its own errors when it has no expected type.
func (c *Checker) lambdaBodyAnyway(x *ast.Lambda, sc *scope) {
	ctx := &fnCtx{parent: c.fn, kind: "lambda", name: "lambda"}
	inner := &scope{parent: sc, names: map[string]*local{}}
	inner.names["x"] = &local{name: "x", typ: types.Invalid{}, slot: x.Scope.X, ctx: ctx, pos: x.Pos, node: x}
	c.info.localObject(inner.names["x"], c.mod.path)
	c.withNested(ctx, func() types.Type { return c.inferBody(x.Body, inner, ctx) })
}

func (c *Checker) funcLit(x *ast.FuncLit, sc *scope) types.Type {
	s := c.litSignature(x, sc)
	delete(c.litSigs, x)
	ctx := &fnCtx{parent: c.fn, kind: "func", name: "func", declared: s.fn.Return}
	inner := &scope{parent: sc, names: map[string]*local{}}
	c.declareParams(x.Params, s, inner, ctx)
	ret := c.withNested(ctx, func() types.Type { return c.inferBody(x.Body, inner, ctx) })
	return types.Func{Params: s.fn.Params, Return: ret, Required: s.fn.Required}
}

// litSignature resolves a func literal's signature once: a binding that reads it to let the
// function call itself (section 6) and the literal's own check share it.
func (c *Checker) litSignature(x *ast.FuncLit, sc *scope) *sig {
	if s, ok := c.litSigs[x]; ok {
		return s
	}
	if c.litSigs == nil {
		c.litSigs = map[*ast.FuncLit]*sig{}
	}
	s := c.signature(&ast.FuncDecl{Pos: x.Pos, Params: x.Params, Return: x.Return, Body: x.Body}, sc, "")
	c.litSigs[x] = s
	return s
}

// withNested runs f as a nested body: a new context whose flow keeps the narrowing of the outer
// locals that cannot be reassigned, a let, a parameter, or a for variable, and drops that of
// every var (section 7).
func (c *Checker) withNested(ctx *fnCtx, f func() types.Type) types.Type {
	savedFn, savedFlow := c.fn, c.flow
	c.fn = ctx
	nf := flow{}
	for l, t := range savedFlow {
		if l.slot.Fixed != ast.Assignable {
			nf[l] = t
		}
	}
	c.flow = nf
	t := f()
	c.fn, c.flow = savedFn, savedFlow
	return t
}

// testBody checks a test or subtest: its name, a string, and its body, a boundary that returns
// void or an Err, as a script body does (section 10, Tests).
func (c *Checker) testBody(t *ast.Test, sc *scope) {
	c.check(t.Name, sc, types.StringType)
	ctx := &fnCtx{parent: c.fn, kind: "test", name: "test"}
	inner := &scope{parent: sc, names: map[string]*local{}}
	c.withNested(ctx, func() types.Type { c.block(t.Body, inner, nil); return types.Void{} })
}

// nestedBody checks a go or defer block.
func (c *Checker) nestedBody(body *ast.Block, sc *scope, kind string) {
	ctx := &fnCtx{parent: c.fn, kind: kind, name: kind + " block"}
	inner := &scope{parent: sc, names: map[string]*local{}}
	c.withNested(ctx, func() types.Type { c.block(body, inner, nil); return types.Void{} })
}

// ---- operators

func (c *Checker) unary(x *ast.Unary, sc *scope) types.Type {
	switch x.Op {
	case "!":
		c.check(x.X, sc, types.BoolType)
		return types.BoolType
	case "-":
		t := c.synth(x.X, sc)
		if !c.operandOK(t, false) {
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
func (c *Checker) operandOK(t types.Type, strings bool) bool {
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
			if !c.operandOK(m, strings) {
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
			n := c.typeNames(l, r)
			c.errorf(x.Pos, "%s and %s cannot be compared with %s: neither is assignable to the other", n[0], n[1], x.Op)
		}
		return types.BoolType
	case "<", "<=", ">", ">=":
		l := c.synth(x.X, sc)
		r := c.synth(x.Y, sc)
		c.sameOperands(l, r, x.Op, true, x.Pos)
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
	if !c.operandOK(l, strings) {
		what := "a single numeric kind or a named primitive type over one"
		if strings {
			what = "a single numeric kind, string, or a named primitive type over one"
		}
		// narrowing helps only a union with a member the operator takes
		hint := ""
		if ms := types.Members(l); len(ms) > 1 && slices.ContainsFunc(ms, func(m types.Type) bool { return c.operandOK(m, strings) }) {
			hint = "; narrow first"
		}
		c.errorf(pos, "%s applies to %s, not to %s%s", op, what, types.String(l), hint)
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
	switch {
	case expected != nil:
		r = c.check(x.Y, sc, expected)
	case onlyEmpty(x.Y) && !types.IsInvalid(nonNull):
		// An empty literal takes the left side's type: m[k] ?? [] (section 4, Inference).
		r = c.check(x.Y, sc, nonNull)
	case onlyEmpty(x.Y):
		// The left side failed and was reported; the empty literal has nothing to add.
		r = types.Invalid{}
	case !types.IsInvalid(nonNull):
		// a call takes the left side's type as a hint: cached ?? Result.ok(0)
		r = c.synthHinted(x.Y, sc, nonNull)
	default:
		r = c.synth(x.Y, sc)
	}
	if types.IsInvalid(r) && expected == nil {
		return types.Invalid{} // reported where it failed
	}
	if types.IsNever(r) {
		if id, ok := x.X.(*ast.Ident); ok {
			if l := c.narrowable(sc, id); l != nil {
				c.narrowLocal(l, nonNull)
			}
		}
		return nonNull
	}
	if expected != nil {
		// As the branches of an if: each side is held to the expected type where it is
		// written, the right one by its check above, so the union is not reported again.
		if !types.IsInvalid(l) {
			c.expect(x.X.Position(), nonNull, expected)
		}
		if types.IsInvalid(expected) {
			return types.Invalid{}
		}
		return expected
	}
	u, err := types.Join(nonNull, r)
	if err != nil {
		c.errorf(x.Pos, "the sides of ?? do not form a valid type: %s; narrow first or give the whole an expected type", err)
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
	if fn.kind == "script" || fn.kind == "test" {
		return okT // a script or a test may return any Err (section 10, Scripts and Tests)
	}
	if fn.declared != nil {
		_, declaredErr, isResult := resultParts(types.Without(fn.declared, types.Null{}))
		if !isResult || !types.Assignable(errT, declaredErr) {
			c.errorf(x.Pos, "try would return %s, but the function returns %s", types.String(errRec), types.String(fn.declared))
		}
	} else {
		fn.returns = append(fn.returns, errRec)
		fn.tries = append(fn.tries, x)
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
		if l := c.narrowable(sc, id); l != nil {
			declared = l.typ
		}
	}
	if c.alwaysTrue(declared, target, x.Pos) {
		return types.BoolType
	}
	c.narrowTo(t, target, x.Pos)
	return types.BoolType
}

// typeExpr types a type expression type(T) as Reflect.Type[T] (section 4, Types as values). T
// is closed: a type parameter of the enclosing function or type has nothing bound to it at run
// time, so the program takes a Reflect.Type[T] parameter instead and builds from it.
func (c *Checker) typeExpr(x *ast.TypeExpr, sc *scope) types.Type {
	if tn, ok := x.Type.(*ast.TypeName); ok && len(tn.Path) == 1 && len(tn.Args) == 0 && !types.BuiltinName(tn.Path[0]) {
		name := tn.Path[0]
		if p, _ := sc.lookupTypeParam(name); p == nil {
			l, ent := c.lookup(sc, name)
			if l != nil || ent != nil && (ent.kind == entVar || ent.kind == entFunc) {
				c.errorf(tn.Pos, "type(...) takes a type, not a value; %s is a value", name)
				return types.Invalid{}
			}
		}
	}
	t := c.resolveType(x.Type, sc)
	if types.IsInvalid(t) {
		return types.Invalid{}
	}
	if d := reachesInterface(t); d != nil {
		// An interface has no run-time identity and no members a format could decode into
		// (section 4, Types as values).
		c.errorf(x.Pos, "type(...) cannot reach an interface: %s is one, and an interface has no run-time identity", d.Name)
		return types.Invalid{}
	}
	var param *types.ParamDecl
	types.Walk(t, func(t types.Type) {
		if p, ok := t.(types.Param); ok && param == nil {
			param = p.Decl
		}
	})
	if param != nil {
		c.errorf(x.Pos, "type(...) cannot mention %s, a type parameter of the enclosing function or type, since nothing is bound to it at run time; take a Reflect.Type[%s] parameter and build the type from it with Reflect.arrayOf, Reflect.setOf, Reflect.mapOf, or Reflect.nullable", param.Name, param.Name)
		return types.Invalid{}
	}
	decl := c.reflectDecl("Type")
	if decl == nil {
		c.errorf(x.Pos, "Reflect is not available in this run, so type(...) cannot be used")
		return types.Invalid{}
	}
	return types.Named{Decl: decl, Args: []types.Type{t}}
}

// reflectDecl is a type of the prelude's Reflect namespace, or nil when a host removed it.
func (c *Checker) reflectDecl(name string) *types.Decl {
	ns, ok := c.prelude["Reflect"]
	if !ok || ns.members == nil {
		return nil
	}
	if e, ok := ns.members[name]; ok && e.kind == entType {
		return e.decl
	}
	return nil
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
		case "Set":
			return types.Set{}
		case "Bytes":
			return types.Named{Decl: c.bytes}
		case "Html":
			return types.Named{Decl: c.html}
		}
		return types.Prim{Kind: builtinKinds[x.Kind]}
	}
	ent := c.entityPath(x.Path, sc)
	c.notePath(x.Path, sc)
	if ent == nil {
		// a private or missing member of an imported module says so
		if f, ok := x.Path.(*ast.Field); ok {
			if base := c.entityPath(f.X, sc); base != nil && base.kind == entModule {
				c.moduleMember(base, f.Name, f.Pos)
				return nil
			}
		}
	}
	if ent != nil && ent.kind == entMissing {
		return nil
	}
	if ent == nil || ent.kind != entType {
		c.errorf(x.Pos, "the right side of is must be a type")
		return nil
	}
	switch ent.decl.Kind {
	case types.AliasDecl:
		c.errorf(x.Pos, "%s is an alias and has no run-time identity; test its members instead", ent.decl.Name)
		return nil
	case types.InterfaceDecl:
		c.errorf(x.Pos, "is cannot test %s, an interface: test a type that satisfies it", ent.decl.Name)
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
			continue
		}
		if types.IsInterface(m) {
			// A value of an interface narrows to a satisfying type, and keeps the interface
			// otherwise, since the set of satisfying types is open (section 4, Interfaces).
			if pattern, free, name, ok := narrowPattern(target); ok {
				inst, found := c.instantiateSatisfying(pattern, free, name, types.Expand(m).(types.Named), pos)
				if !found {
					return nil, nil, false
				}
				yes = append(yes, inst)
				no = append(no, m)
				continue
			}
		}
		no = append(no, m)
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
		n := c.typeNames(t, target)
		c.errorf(pos, "%s is never a %s, so the test is always false", n[0], n[1])
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

// targetName names the target of an is test: a kind word for func, Array, Map and Set, which carry
// no type arguments there, and the type otherwise.
func targetName(target types.Type) string {
	switch target.(type) {
	case types.Func:
		return "func"
	case types.Array:
		return "Array"
	case types.Map:
		return "Map"
	case types.Set:
		return "Set"
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
	case types.Set:
		_, ok := types.Expand(m).(types.Set)
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
	n := types.Named{Decl: decl, Args: args}
	if err := c.checkArgs(n); err != nil {
		c.errorf(pos, "%s: %s", types.String(n), err)
		return nil, false
	}
	if err := types.CheckBounds(n); err != nil {
		c.errorf(pos, "%s: %s", types.String(n), err)
		return nil, false
	}
	return n, true
}

// ---- assignment

func (c *Checker) assign(x *ast.Assign, sc *scope) types.Type {
	switch t := x.Target.(type) {
	case *ast.Ident:
		if t.Ref.Kind == ast.RefDiscard {
			if x.Op != "=" {
				// a compound assignment reads its target
				c.errorf(t.Pos, "_ is write-only and cannot be read")
			}
			c.synth(x.Value, sc)
			return types.Void{}
		}
		l, ent := c.lookup(sc, t.Name)
		switch {
		case l != nil && c.hiddenX(l, t.Pos):
			c.synth(x.Value, sc)
		case c.fixedTarget(t, l, ent):
			c.synth(x.Value, sc)
		case l != nil:
			if x.Op == "=" {
				v := c.check(x.Value, sc, l.typ)
				c.narrowLocal(l, v)
			} else {
				c.assignValue(x, sc, c.typeOfLocal(l))
			}
		case ent == nil:
			c.errorf(t.Pos, "%s is not defined", t.Name)
			c.synth(x.Value, sc)
		case ent.kind == entVar:
			if ent.v.mod != c.mod {
				c.errorf(t.Pos, "cannot assign to the imported variable %s", t.Name)
			}
			c.assignValue(x, sc, c.varType(ent.v, t.Pos, c.fn != nil && (c.fn.kind == "init" || c.fn.kind == "script")))
		default:
			c.errorf(t.Pos, "cannot assign to %s", t.Name)
			c.synth(x.Value, sc)
		}
	case *ast.Field, *ast.Index:
		pt, isMap := c.placeType(t.(ast.Expr), sc)
		switch {
		case pt == nil:
			c.synth(x.Value, sc)
		case x.Op != "=" && isMap:
			// A compound assignment reads its target first, and a map read is V | null
			// (section 11, Map), which no operator takes.
			c.synth(x.Value, sc)
			if !types.IsInvalid(pt) {
				c.errorf(x.Pos, "%s on a map element reads it first, and the read is %s | null; write m[k] = (m[k] ?? init) %s v", x.Op, types.String(pt), x.Op[:len(x.Op)-1])
			}
		default:
			c.assignValue(x, sc, pt)
		}
	case *ast.AssignArray:
		v := c.synth(x.Value, sc)
		c.assignArray(t, v, sc)
	}
	return types.Void{}
}

// fixedTarget reports an assignment to a name that cannot be reassigned: a let, a parameter, a
// lambda's x, or a for variable, local or of the module (section 3). An imported let is
// reported as any import is.
func (c *Checker) fixedTarget(t *ast.Ident, l *local, ent *entity) bool {
	fixed := ast.Assignable
	switch {
	case l != nil:
		fixed = l.slot.Fixed
	case ent != nil && ent.kind == entVar && ent.v.mod == c.mod && ent.v.decl.Let:
		fixed = ast.FixedLet
	}
	switch fixed {
	case ast.Assignable:
		return false
	case ast.FixedLet:
		c.errorf(t.Pos, "cannot assign to %s, a let; declare it with var to reassign it", t.Name)
	default:
		c.errorf(t.Pos, "cannot assign to %s, %s; declare a var to hold a value that changes", t.Name, fixed.What())
	}
	return true
}

// assignValue checks the right side of x against target, the type its place stores: as an
// expected type for =, and as the other operand of a compound assignment.
func (c *Checker) assignValue(x *ast.Assign, sc *scope, target types.Type) {
	if x.Op == "=" {
		c.check(x.Value, sc, target)
		return
	}
	op := x.Op[:len(x.Op)-1] // "+=" -> "+"
	c.sameOperands(target, c.synth(x.Value, sc), op, op == "+", x.Pos)
}

// placeType checks the receiver and index of a field or index place and returns the type the
// place stores, or nil after reporting why it cannot be assigned: an import or a union
// receiver, or a receiver that cannot be indexed. isMap reports a map element, whose read is
// V | null.
func (c *Checker) placeType(p ast.Expr, sc *scope) (_ types.Type, isMap bool) {
	switch t := p.(type) {
	case *ast.Field:
		if base := c.entityPath(t.X, sc); base != nil && base.kind == entModule {
			c.errorf(t.Pos, "cannot assign to %s.%s: imports are read-only", base.name, t.Name)
			return nil, false
		}
		recv := c.synth(t.X, sc)
		if _, isUnion := types.Expand(recv).(types.Union); isUnion {
			c.errorf(t.Pos, "cannot assign a field through the union %s; narrow first", types.String(recv))
			return nil, false
		}
		return c.fieldOf(recv, t.Name, t.Pos, nil), false
	case *ast.Index:
		recv := c.synth(t.X, sc)
		switch r := types.Expand(recv).(type) {
		case types.Array:
			c.check(t.Index, sc, types.IntType)
			return r.Elem, false
		case types.Map:
			c.check(t.Index, sc, r.Key)
			return r.Val, true
		}
		c.synth(t.Index, sc)
		if c.isBytes(recv) {
			c.errorf(t.Pos, "Bytes is immutable, so a byte cannot be assigned; build an Array[byte] and convert it with Bytes.fromArray")
		} else if !types.IsInvalid(recv) {
			c.errorf(t.Pos, "%s cannot be indexed", types.String(recv))
		}
	}
	return nil, false
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
		if t.Ref.Kind == ast.RefDiscard {
			return
		}
		l, ent := c.lookup(sc, t.Name)
		switch {
		case l != nil && c.hiddenX(l, t.Pos):
		case c.fixedTarget(t, l, ent):
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
	case *ast.Field, *ast.Index:
		// the rules of a plain field or index assignment: no import, no union
		if pt, _ := c.placeType(t, sc); pt != nil {
			c.expect(t.Position(), v, pt)
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
