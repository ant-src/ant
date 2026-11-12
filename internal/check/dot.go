package check

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// ---- calls through .

// receiver is the value a call through a dot passes first (section 5, Calls through .): its
// type, where it is written, and whether the call is positional, as it is through an
// interface, a type parameter, or a union of several namespaces. A receiver whose namespace is
// known without its type, such as [[1], ["x"]] or Set.of([]), is not typed yet: typ is nil, and
// expr is checked as the first argument of the direct call, which waits for the other arguments
// when it fixes nothing.
type receiver struct {
	typ        types.Type
	expr       ast.Expr
	pos        ast.Pos
	at         ast.Expr // the receiver as written, what an error about it covers
	positional bool
}

// untyped reports a receiver checked as the direct call's first argument.
func (r *receiver) untyped() bool {
	return r != nil && r.expr != nil
}

// dotCall checks a call x.f(...) whose x is a value rather than a namespace, a type or a module
// (section 5, Calls through .): the field f of x when its type has one, else the function f of
// the namespace of x's type, with x first. It reports false for a call that call resolves as a
// direct one, or whose x.f is written in parentheses and so is a field read.
func (c *Checker) dotCall(x *ast.Call, f *ast.Field, sc *scope, expected types.Type) (types.Type, bool) {
	if f.Grouped {
		return nil, false
	}
	if base := c.entityPath(f.X, sc); base != nil {
		switch base.kind {
		case entType, entNamespace, entModule, entMissing:
			return nil, false
		}
	}
	if !f.Optional && (onlyEmpty(f.X) || isCollectionLit(f.X)) {
		if t, ok := c.waitingCall(x, f, sc, expected); ok {
			return c.endLink(x, x.Grouped, t, false), true
		}
	}
	recv, short := c.chainRecv(f.X, sc)
	if f.Optional {
		if !types.Includes(recv, types.Null{}) {
			c.errorf(f.Pos, "?. needs a nullable left side, got %s%s", types.String(recv), chainNote(short, "."))
			c.synthArgs(x.Args, sc)
			return types.Invalid{}, true
		}
		recv = types.Without(recv, types.Null{})
		short = true
	}
	inner := c.dotCallOn(recv, x, f, sc, expected, false)
	return c.endLink(x, x.Grouped, inner, short), true
}

// dotCallOn checks the call on a receiver of type recv: a field, a function of an interface,
// the function of the type's namespace, or one call per member of a union whose members do not
// share a namespace, whose results join.
func (c *Checker) dotCallOn(recv types.Type, x *ast.Call, f *ast.Field, sc *scope, expected types.Type, positional bool) types.Type {
	if types.IsInvalid(recv) {
		c.synthArgs(x.Args, sc)
		return types.Invalid{}
	}
	t := types.Expand(recv)
	switch r := t.(type) {
	case types.Union:
		if ns := c.sharedNamespace(r.Members); ns != nil && !c.anyField(r.Members, f.Name) {
			return c.namespaceCall(recv, ns, x, f, sc, expected, positional)
		}
		for _, m := range r.Members {
			if !c.offers(m, f.Name) {
				if alias := c.aliasFunc(recv, f.Name); alias != "" {
					c.errorf(f.Pos, "%s.%s is declared on the alias %s, which a value does not reach; call it as %s.%s(...)", alias, f.Name, alias, alias, f.Name)
					c.synthArgs(x.Args, sc)
					return types.Invalid{}
				}
				c.errorf(f.Pos, "%s has no field or function %s on every member: narrow the value first%s", types.String(recv), f.Name, c.outerVarHint(f.X, sc))
				c.synthArgs(x.Args, sc)
				return types.Invalid{}
			}
		}
		for _, a := range x.Args {
			if a.Name != "" {
				// one error for the call, not one per member (section 5, Calls through .)
				c.errorf(a.Pos, "a call through %s is positional, since each member's %s names its parameters as it likes; pass %s positionally", types.String(recv), f.Name, a.Name)
				c.synthArgs(x.Args, sc)
				return types.Invalid{}
			}
		}
		var results []types.Type
		for _, m := range r.Members {
			results = append(results, c.dotCallOn(m, x, f, sc, expected, true))
		}
		u, err := types.MakeUnion(results)
		if err != nil {
			c.errorIn(x, "the results of %s through %s do not form a valid type: %s; narrow first or give the whole an expected type", f.Name, types.String(recv), err)
			return types.Invalid{}
		}
		return u
	case types.Param:
		// as for its bound, with the parameter itself as the receiver
		if r.Decl.Bound == nil {
			c.errorf(f.Pos, "%s has no field or function %s: it has no bound", r.Decl.Name, f.Name)
			c.synthArgs(x.Args, sc)
			return types.Invalid{}
		}
		if ft, ok := c.fieldType(t, f.Name); ok {
			return c.fieldCall(ft, recv, x, f, sc)
		}
		if b, ok := types.Expand(r.Decl.Bound).(types.Named); ok && b.Decl.Kind == types.InterfaceDecl {
			return c.interfaceCall(recv, b, x, f, sc, expected)
		}
		ns := c.sharedNamespace(types.Members(r.Decl.Bound))
		if ns == nil {
			c.errorf(f.Pos, "%s has no field or function %s on every member of its bound %s", r.Decl.Name, f.Name, types.String(r.Decl.Bound))
			c.synthArgs(x.Args, sc)
			return types.Invalid{}
		}
		return c.namespaceCall(recv, ns, x, f, sc, expected, true)
	}
	if ft, ok := c.fieldType(t, f.Name); ok {
		return c.fieldCall(ft, recv, x, f, sc)
	}
	if n, ok := t.(types.Named); ok && n.Decl.Kind == types.InterfaceDecl {
		return c.interfaceCall(recv, n, x, f, sc, expected)
	}
	ns := c.namespaceOf(t)
	if ns == nil {
		c.errorf(f.Pos, "%s has no field or function %s", types.String(recv), f.Name)
		c.synthArgs(x.Args, sc)
		return types.Invalid{}
	}
	return c.namespaceCall(recv, ns, x, f, sc, expected, positional)
}

// offers reports whether a call through a dot finds a field or a function name on a value of
// type t, before the function's first parameter is checked.
func (c *Checker) offers(t types.Type, name string) bool {
	t = types.Expand(t)
	if _, ok := c.fieldType(t, name); ok {
		return true
	}
	switch x := t.(type) {
	case types.Named:
		if x.Decl.Kind == types.InterfaceDecl {
			_, ok := x.Decl.Funcs[name]
			return ok
		}
	case types.Param:
		return x.Decl.Bound != nil && c.offers(x.Decl.Bound, name)
	case types.Union:
		for _, m := range x.Members {
			if !c.offers(m, name) {
				return false
			}
		}
		return true
	}
	ns := c.namespaceOf(t)
	if ns == nil {
		return false
	}
	m, ok := ns.members[name]
	return ok && m.kind == entFunc && !c.foreignPrivate(m)
}

// fieldType is the type of the field name of t, when t is a record type that declares it.
func (c *Checker) fieldType(t types.Type, name string) (types.Type, bool) {
	for _, f := range types.FieldsOf(t) {
		if f.Name == name {
			return f.Type, true
		}
	}
	return nil, false
}

// fieldCall checks a call of the field f of x, as the call of a function value.
func (c *Checker) fieldCall(ft, recv types.Type, x *ast.Call, f *ast.Field, sc *scope) types.Type {
	c.info.recordMember(f, c.info.fieldOf(recv, f.Name))
	c.info.recordType(f, ft)
	return c.callValue(ft, x, sc)
}

// anyField reports whether a member of a union has a field name.
func (c *Checker) anyField(members []types.Type, name string) bool {
	for _, m := range members {
		if _, ok := c.fieldType(types.Expand(m), name); ok {
			return true
		}
	}
	return false
}

// namespaceOf is the namespace whose functions a call through a dot reaches on a value of type
// t (section 5, Calls through .): the entity of a user type, which holds the functions declared
// on it, or the prelude namespace the table of section 5 names for a built-in type. It is nil
// for a number, bool, null, void, an anonymous record type, a type parameter, an interface,
// and a union.
func (c *Checker) namespaceOf(t types.Type) *entity {
	switch x := types.Expand(t).(type) {
	case types.Prim:
		if x.Kind == value.KindString {
			return c.prelude["String"]
		}
	case types.Array:
		return c.prelude["Array"]
	case types.Map:
		return c.prelude["Map"]
	case types.Set:
		return c.prelude["Set"]
	case types.Func:
		return c.prelude["Iter"] // a sequence (Decision 7)
	case types.Named:
		d := x.Decl
		switch {
		case d.Kind == types.InterfaceDecl:
			return nil
		case d == c.bytes:
			return c.prelude["Bytes"]
		case d == c.html:
			return c.prelude["Html"]
		case d == c.okDecl || d == c.errDecl:
			return c.prelude["Result"]
		}
		if i := strings.IndexByte(d.Name, '.'); i > 0 {
			return c.prelude[d.Name[:i]] // a type a prelude namespace declares, Time.Instant or File.Handle
		}
		if ti := c.typeOf[d]; ti != nil {
			return ti.ent
		}
	}
	return nil
}

// sharedNamespace is the one namespace every member of a union has, or nil.
func (c *Checker) sharedNamespace(members []types.Type) *entity {
	var ns *entity
	for _, m := range members {
		n := c.namespaceOf(m)
		if n == nil || ns != nil && n != ns {
			return nil
		}
		ns = n
	}
	return ns
}

// namespaceCall checks a call of the function f.Name of the namespace ns on a receiver of type
// recv: the function must be visible here and take recv first, and the call is then checked as
// the direct call ns.f(recv, ...).
func (c *Checker) namespaceCall(recv types.Type, ns *entity, x *ast.Call, f *ast.Field, sc *scope, expected types.Type, positional bool) types.Type {
	m, ok := ns.members[f.Name]
	if !ok || m.kind != entFunc {
		c.errorf(f.Pos, "%s has no field or function %s", types.String(recv), f.Name)
		c.synthArgs(x.Args, sc)
		return types.Invalid{}
	}
	if c.foreignPrivate(m) {
		c.errorf(f.Pos, "%s.%s is not exported", ns.name, f.Name)
		c.synthArgs(x.Args, sc)
		return types.Invalid{}
	}
	s := c.sigOf(m, f.Pos)
	if s == nil {
		c.synthArgs(x.Args, sc)
		return types.Invalid{}
	}
	if !c.accepts(s, recv) {
		if len(s.fn.Params) == 0 {
			c.errorf(f.Pos, "%s does not take a %s first, so a value does not reach it; call it as %s()", s.fullName(), types.String(recv), s.fullName())
		} else {
			// The direct call would fail the same way, so the hint is to narrow, and only a
			// union with a member the function takes can.
			hint := ""
			if ms := types.Members(recv); len(ms) > 1 && slices.ContainsFunc(ms, func(m types.Type) bool { return c.accepts(s, m) }) {
				hint = "; narrow it first"
			}
			n := c.typeNames(recv, s.fn.Params[0])
			c.errorf(f.Pos, "%s takes %s first, so a %s does not reach it%s", s.fullName(), n[1], n[0], hint)
		}
		c.synthArgs(x.Args, sc)
		return types.Invalid{}
	}
	c.info.recordMember(f, c.info.object(m))
	return c.applySig(s, x, sc, expected, c.info.object(m), &receiver{typ: recv, pos: f.X.Position(), at: f.X, positional: positional})
}

// aliasFunc is the name of the alias recv when a function name is declared on it: a value of
// the alias carries a member's tag, so x.name does not reach it (section 5, Calls through .).
func (c *Checker) aliasFunc(recv types.Type, name string) string {
	n, ok := recv.(types.Named)
	if !ok || n.Decl.Kind != types.AliasDecl {
		return ""
	}
	ti := c.typeOf[n.Decl]
	if ti == nil || ti.ent == nil {
		return ""
	}
	if m, ok := ti.ent.members[name]; ok && m.kind == entFunc {
		return n.Decl.Name
	}
	return ""
}

// waitingCall checks x.f(…) whose receiver is typed by the call, as Set.of([]).isSubset(s),
// [].concat([1]) or [[1], ["x"]].flatten(). The namespace comes from the receiver's kind alone,
// and the call is checked as the direct call Set.isSubset(Set.of([]), s): the receiver is its
// first argument, with the expected type of the call and the other arguments to go by.
// It reports false when the namespace has no such function, so that the usual path reports it.
func (c *Checker) waitingCall(x *ast.Call, f *ast.Field, sc *scope, expected types.Type) (types.Type, bool) {
	ns := c.kindNamespace(f.X, sc)
	if ns == nil {
		return nil, false
	}
	m, ok := ns.members[f.Name]
	if !ok || m.kind != entFunc || c.foreignPrivate(m) {
		return nil, false
	}
	s := c.sigOf(m, f.Pos)
	if s == nil || len(s.fn.Params) == 0 || c.namespaceOf(s.fn.Params[0]) != ns {
		return nil, false // f does not take a value of this kind first
	}
	c.info.recordMember(f, c.info.object(m))
	return c.applySig(s, x, sc, expected, c.info.object(m), &receiver{expr: f.X, pos: f.X.Position(), at: f.X}), true
}

// isCollectionLit reports an array or map literal, whose element types an expected type can
// decide.
func isCollectionLit(e ast.Expr) bool {
	switch e.(type) {
	case *ast.ArrayLit, *ast.MapLit:
		return true
	}
	return false
}

// kindNamespace is the namespace of a value that fixes nothing, known without typing it: Array
// for an array literal, Map for Map{}, and for a call such as Set.of([]) the namespace of its
// declared result type.
func (c *Checker) kindNamespace(e ast.Expr, sc *scope) *entity {
	switch x := e.(type) {
	case *ast.ArrayLit:
		return c.prelude["Array"]
	case *ast.MapLit:
		return c.prelude["Map"]
	case *ast.Call:
		if ent := c.entityPath(x.Fn, sc); ent != nil && ent.kind == entFunc {
			if s := c.sigOf(ent, ast.Start(x)); s != nil {
				return c.namespaceOf(s.fn.Return)
			}
		}
	}
	return nil
}

// accepts reports whether the first parameter of s takes a value of type recv: any value when
// it is a type parameter (Decision 10).
func (c *Checker) accepts(s *sig, recv types.Type) bool {
	if len(s.fn.Params) == 0 {
		return false
	}
	if len(s.tparams) == 0 {
		return types.Assignable(recv, s.fn.Params[0])
	}
	return c.unify(s.fn.Params[0], recv, map[*types.ParamDecl]types.Type{})
}

// interfaceCall checks a call of the function f.Name of the interface n on a receiver of type
// self: the interface itself, or a type parameter bounded by it.
func (c *Checker) interfaceCall(self types.Type, n types.Named, x *ast.Call, f *ast.Field, sc *scope, expected types.Type) types.Type {
	s := c.interfaceSig(n, f.Name, self)
	if s == nil {
		c.errorf(f.Pos, "%s has no field or function %s", types.String(self), f.Name)
		c.synthArgs(x.Args, sc)
		return types.Invalid{}
	}
	callee := c.info.interfaceFunc(n.Decl, f.Name)
	c.info.recordMember(f, callee)
	return c.applySig(s, x, sc, expected, callee, &receiver{typ: self, pos: f.X.Position(), at: f.X, positional: true})
}

// interfaceSig is the signature of the function name of the interface n read for the satisfying
// type self, with the interface's type parameters and Self substituted; nil when n has no such
// function.
func (c *Checker) interfaceSig(n types.Named, name string, self types.Type) *sig {
	fn, ok := n.Decl.Funcs[name]
	if !ok {
		return nil
	}
	m := types.Bindings(n.Decl.Params, n.Args)
	m[n.Decl.Self] = self
	t, err := types.Subst(*fn, m)
	if err != nil {
		return nil
	}
	return &sig{fn: t.(types.Func), name: name, ownerName: n.Decl.Name}
}

// ---- interfaces

// satisfier is the Satisfies hook of an interface's declaration (types.Decl), which decides
// with the namespaces the checker knows, visible where it asks.
func (c *Checker) satisfier(decl *types.Decl) func(s types.Type, args []types.Type, rel func(a, b types.Type) bool) bool {
	return func(s types.Type, args []types.Type, rel func(a, b types.Type) bool) bool {
		return c.satisfies(s, decl, args, rel)
	}
}

// satisfies reports whether a type s satisfies the interface decl instantiated with args
// (section 4, Interfaces): for each function of the interface, the call reaches a namespace
// function on s whose type, instantiated for s, rel relates to the signature with Self as s.
func (c *Checker) satisfies(s types.Type, decl *types.Decl, args []types.Type, rel func(a, b types.Type) bool) bool {
	if _, isUnion := types.Expand(s).(types.Union); isUnion {
		return false
	}
	m := types.Bindings(decl.Params, args)
	m[decl.Self] = s
	for _, name := range decl.Order {
		wt, err := types.Subst(*decl.Funcs[name], m)
		if err != nil {
			return false
		}
		want := wt.(types.Func)
		got := c.offered(s, name, want, rel)
		if got == nil || !rel(*got, want) {
			return false
		}
	}
	return true
}

// unsatisfied explains why s does not satisfy the interface n, for an error: the first function
// of n that s does not offer, or offers with another type. It is "" when there is none to name.
func (c *Checker) unsatisfied(s types.Type, n types.Named) string {
	if _, isUnion := types.Expand(s).(types.Union); isUnion {
		return ""
	}
	m := types.Bindings(n.Decl.Params, n.Args)
	m[n.Decl.Self] = s
	for _, name := range n.Decl.Order {
		wt, err := types.Subst(*n.Decl.Funcs[name], m)
		if err != nil {
			return ""
		}
		want := wt.(types.Func)
		got := c.offered(s, name, want, types.Assignable)
		names := c.typeNames(s, want)
		switch {
		case got == nil:
			if _, isField := c.fieldType(types.Expand(s), name); isField {
				return fmt.Sprintf("%s has a field %s, not a function declared on it; the interface needs func %s.%s", names[0], name, names[0], name)
			}
			if ns := c.namespaceOf(types.Expand(s)); ns != nil {
				if m, ok := ns.members[name]; ok && m.kind == entFunc && c.foreignPrivate(m) {
					return fmt.Sprintf("%s.%s is not exported, so it does not satisfy %s here", names[0], name, types.String(n))
				}
			}
			return fmt.Sprintf("%s has no function %s of type %s", names[0], name, names[1])
		case !types.Assignable(*got, want):
			return fmt.Sprintf("%s.%s has type %s, not %s", names[0], name, c.typeNames(*got)[0], names[1])
		}
	}
	return ""
}

// offered is the function name that a call through a dot reaches on a value of type s, as a
// function type taking the receiver first, instantiated for s and, where s alone does not fix
// a type parameter, for the signature want it is to match. It is nil when the call reaches a
// field, nothing, or a function s does not instantiate. Whether the first parameter takes s is
// decided by rel, the assignability under way: a first parameter that is the interface itself
// holds for the pair under test instead of asking again without end.
func (c *Checker) offered(s types.Type, name string, want types.Func, rel func(a, b types.Type) bool) *types.Func {
	t := types.Expand(s)
	var ns *entity
	switch x := t.(type) {
	case types.Named:
		if x.Decl.Kind == types.InterfaceDecl {
			if sg := c.interfaceSig(x, name, s); sg != nil {
				return &sg.fn
			}
			return nil
		}
		ns = c.namespaceOf(t)
	case types.Param:
		if x.Decl.Bound == nil {
			return nil
		}
		if b, ok := types.Expand(x.Decl.Bound).(types.Named); ok && b.Decl.Kind == types.InterfaceDecl {
			if sg := c.interfaceSig(b, name, s); sg != nil {
				return &sg.fn
			}
			return nil
		}
		if _, isField := c.fieldType(t, name); isField {
			return nil
		}
		ns = c.sharedNamespace(types.Members(x.Decl.Bound))
	default:
		ns = c.namespaceOf(t)
	}
	if _, isField := c.fieldType(t, name); isField || ns == nil {
		return nil
	}
	m, ok := ns.members[name]
	if !ok || m.kind != entFunc || c.foreignPrivate(m) {
		return nil
	}
	sg := c.sigReady(m)
	if sg == nil || len(sg.fn.Params) == 0 {
		return nil
	}
	if len(sg.tparams) == 0 {
		if !rel(s, sg.fn.Params[0]) {
			return nil
		}
		fn := sg.fn
		return &fn
	}
	subst := map[*types.ParamDecl]types.Type{}
	if !c.unify(sg.fn.Params[0], s, subst) {
		return nil
	}
	for i := 1; i < len(sg.fn.Params) && i < len(want.Params); i++ {
		c.unify(sg.fn.Params[i], want.Params[i], subst)
	}
	c.unify(sg.fn.Return, want.Return, subst)
	for _, p := range sg.tparams {
		if _, done := subst[p]; !done || !boundFits(p, subst[p], subst, sg.tparams) {
			return nil
		}
	}
	inst, err := types.Subst(sg.fn, subst)
	if err != nil {
		return nil
	}
	fn := inst.(types.Func)
	return &fn
}

// sigReady is a function entity's signature with its return type inferred, or nil when it
// cannot be known here: the function is being inferred, as a recursive one is.
func (c *Checker) sigReady(ent *entity) *sig {
	if ent.sig == nil && ent.fn != nil {
		c.resolveSignature(ent.fn)
	}
	s := ent.sig
	if s == nil {
		return nil
	}
	if s.fn.Return == nil {
		if ent.fn == nil || s.inferring || ent.fn.state == 1 {
			return nil
		}
		c.checkFunc(ent.fn)
		if s.fn.Return == nil {
			return nil
		}
	}
	return s
}

// narrowPattern is the type an is test on an interface narrows to for a target (section 5, is),
// with the type arguments the target leaves open as free type parameters: a named type with its
// own, an array, a map, or a set with fresh ones, a primitive, Bytes, or Html with none. The
// name is the target's, for messages. A function target has no type to narrow to.
func narrowPattern(target types.Type) (pattern types.Type, free []*types.ParamDecl, name string, ok bool) {
	fresh := func(name string) types.Type {
		p := &types.ParamDecl{Name: name}
		free = append(free, p)
		return types.Param{Decl: p}
	}
	switch t := target.(type) {
	case types.Named:
		switch t.Decl.Kind {
		case types.RecordDecl, types.PrimitiveDecl, types.KindOpaqueDecl, types.TaggedOpaqueDecl:
		default:
			return nil, nil, "", false
		}
		var args []types.Type
		for _, p := range t.Decl.Params {
			args = append(args, types.Param{Decl: p})
		}
		return types.Named{Decl: t.Decl, Args: args}, t.Decl.Params, t.Decl.Name, true
	case types.Prim:
		return t, nil, types.String(t), true
	case types.Array:
		return types.Array{Elem: fresh("T")}, free, "Array", true
	case types.Map:
		return types.Map{Key: fresh("K"), Val: fresh("V")}, free, "Map", true
	case types.Set:
		return types.Set{Elem: fresh("T")}, free, "Set", true
	}
	return nil, nil, "", false
}

// instantiateSatisfying finds the type an is test narrows a value of the interface iface to
// (section 5, is): the pattern with its free type parameters solved, the one instantiation that
// satisfies the interface. Each function of the interface, matched against the function of that
// name the pattern's namespace offers, fixes the parameters its receiver mentions, the functions
// must agree, and together they must fix every parameter; a pattern without parameters is
// simply checked. name names the target in messages.
func (c *Checker) instantiateSatisfying(pattern types.Type, free []*types.ParamDecl, name string, iface types.Named, pos ast.Pos) (types.Type, bool) {
	if len(free) == 0 {
		if !types.Assignable(pattern, iface) {
			c.errorf(pos, "%s does not satisfy %s, so the test is always false", name, types.String(iface))
			return nil, false
		}
		return pattern, true
	}
	ns := c.namespaceOf(pattern)
	if ns == nil {
		c.errorf(pos, "%s does not satisfy %s, so the test is always false", name, types.String(iface))
		return nil, false
	}
	m := types.Bindings(iface.Decl.Params, iface.Args)
	fixed := map[*types.ParamDecl]types.Type{} // the free parameters some function has fixed
	for _, fname := range iface.Decl.Order {
		wt, err := types.Subst(*iface.Decl.Funcs[fname], m)
		if err != nil {
			return nil, false
		}
		want := wt.(types.Func)
		ent, ok := ns.members[fname]
		if !ok || ent.kind != entFunc || c.foreignPrivate(ent) {
			c.errorf(pos, "%s has no function %s, so it does not satisfy %s and the test is always false", name, fname, types.String(iface))
			return nil, false
		}
		sg := c.sigReady(ent)
		if sg == nil {
			return nil, false
		}
		if len(sg.fn.Params) == 0 {
			c.errorf(pos, "%s takes no argument, so %s does not satisfy %s and the test is always false", sg.fullName(), name, types.String(iface))
			return nil, false
		}
		subst := map[*types.ParamDecl]types.Type{}
		for i := 1; i < len(sg.fn.Params) && i < len(want.Params); i++ {
			c.unify(sg.fn.Params[i], want.Params[i], subst)
		}
		c.unify(sg.fn.Return, want.Return, subst)
		r, err := types.Subst(sg.fn.Params[0], subst)
		if err != nil {
			c.errorf(pos, "%s", err)
			return nil, false
		}
		open := map[*types.ParamDecl]bool{}
		for _, p := range sg.tparams {
			if _, done := subst[p]; !done {
				open[p] = true
			}
		}
		bind := map[*types.ParamDecl]types.Type{}
		if !c.unify(pattern, r, bind) {
			c.errorf(pos, "%s takes %s first, not %s, so %s does not satisfy %s and the test is always false", sg.fullName(), types.String(r), withArticle(name), name, types.String(iface))
			return nil, false
		}
		for _, p := range free {
			t, ok := bind[p]
			if !ok || types.Mentions(t, open) {
				continue // this function leaves the parameter open
			}
			if prev, done := fixed[p]; !done {
				fixed[p] = t
			} else if !types.Equal(prev, t) {
				c.errorf(pos, "%s fixes the type parameter %s of %s to %s through one function and to %s through %s, so the test cannot narrow it", types.String(iface), p.Name, name, types.String(prev), types.String(t), fname)
				return nil, false
			}
		}
	}
	for _, p := range free {
		if _, ok := fixed[p]; !ok {
			c.errorf(pos, "%s does not determine the type parameter %s of %s, so the test cannot narrow to it", types.String(iface), p.Name, name)
			return nil, false
		}
	}
	inst, err := types.Subst(pattern, fixed)
	if err != nil {
		c.errorf(pos, "%s", err)
		return nil, false
	}
	if n, ok := inst.(types.Named); ok {
		if err := c.checkArgs(n); err != nil {
			c.errorf(pos, "%s: %s", types.String(n), err)
			return nil, false
		}
		if err := types.CheckBounds(n); err != nil {
			c.errorf(pos, "%s: %s", types.String(n), err)
			return nil, false
		}
	}
	if !types.Assignable(inst, iface) {
		c.errorf(pos, "%s does not satisfy %s, so the test is always false", types.String(inst), types.String(iface))
		return nil, false
	}
	return inst, true
}

// withArticle prefixes a type's name with a or an.
func withArticle(name string) string {
	switch name[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an " + name
	}
	return "a " + name
}

// reachesInterface is the first interface a type reaches, through the fields and bodies of the
// declarations it names, or nil (section 4, Types as values).
func reachesInterface(t types.Type) *types.Decl {
	seen := map[*types.Decl]bool{}
	var found *types.Decl
	var visit func(t types.Type)
	visit = func(t types.Type) {
		types.Walk(t, func(t types.Type) {
			n, ok := t.(types.Named)
			if !ok || found != nil || seen[n.Decl] {
				return
			}
			if n.Decl.Kind == types.InterfaceDecl {
				found = n.Decl
				return
			}
			seen[n.Decl] = true
			for _, f := range n.Decl.Fields {
				visit(f.Type)
			}
			if n.Decl.Body != nil {
				visit(n.Decl.Body)
			}
		})
	}
	visit(t)
	return found
}

// mentionsSelf reports whether a written type names Self.
func mentionsSelf(t ast.Type) bool {
	found := false
	if t != nil {
		ast.Inspect(t, func(n ast.Node) bool {
			if tn, ok := n.(*ast.TypeName); ok && len(tn.Path) == 1 && tn.Path[0] == "Self" {
				found = true
			}
			return !found
		})
	}
	return found
}

// resolveInterfaces resolves the signatures of a module's interfaces, once its types are
// declared: each function's parameters and return type, in a scope where Self and the
// interface's own type parameters are bound.
func (c *Checker) resolveInterfaces(mi *moduleInfo) {
	for _, ii := range mi.ifaces {
		d, decl := ii.decl, ii.t
		sc := &scope{tparams: map[string]*types.ParamDecl{}, typeDecl: d.Name}
		for i, tp := range d.TypeParams {
			sc.tparams[tp.Name] = decl.Params[i]
		}
		sc.tparams["Self"] = decl.Self
		for i, tp := range d.TypeParams {
			if tp.Bound != nil {
				decl.Params[i].Bound = c.resolveBound(tp.Bound, sc)
			}
		}
		c.bareBounds(decl.Params, d.TypeParams)
		decl.Funcs = map[string]*types.Func{}
		for i := range d.Funcs {
			f := &d.Funcs[i]
			if _, dup := decl.Funcs[f.Name]; dup {
				c.errorf(f.Pos, "%s declares %s twice", d.Name, f.Name)
				continue
			}
			fn := &types.Func{Required: len(f.Params)}
			for _, p := range f.Params {
				fn.Params = append(fn.Params, c.resolveType(p.Type, sc))
				fn.Names = append(fn.Names, p.Name)
			}
			fn.Return = c.resolveType(f.Return, sc)
			decl.Funcs[f.Name] = fn
			decl.Order = append(decl.Order, f.Name)
			c.info.interfaceFunc(decl, f.Name)
		}
	}
}
