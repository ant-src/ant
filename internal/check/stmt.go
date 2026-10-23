package check

import (
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
)

// flow is the narrowing state: the current type of each local that differs from its declared
// type (section 7).
type flow map[*local]types.Type

func (f flow) copy() flow {
	out := make(flow, len(f))
	for l, t := range f {
		out[l] = t
	}
	return out
}

// joinFlows merges the states of two paths: a local narrowed on both takes the union, one
// narrowed on one path only is joined with its declared type.
func joinFlows(a, b flow) flow {
	out := flow{}
	for l, t := range a {
		if u, ok := b[l]; ok {
			j, err := types.MakeUnion([]types.Type{t, u})
			if err != nil {
				continue
			}
			out[l] = j
			continue
		}
		if j, err := types.MakeUnion([]types.Type{t, l.typ}); err == nil && !types.Equal(j, l.typ) {
			out[l] = j
		}
	}
	for l, u := range b {
		if _, ok := a[l]; ok {
			continue
		}
		if j, err := types.MakeUnion([]types.Type{u, l.typ}); err == nil && !types.Equal(j, l.typ) {
			out[l] = j
		}
	}
	return out
}

// narrowLocal sets a local's current type after a test or an assignment. A local assigned
// inside a nested body never narrows.
func (c *Checker) narrowLocal(l *local, t types.Type) {
	if l.slot.AssignedNested {
		return
	}
	if types.IsInvalid(t) || types.Equal(t, l.typ) {
		delete(c.flow, l)
		return
	}
	c.flow[l] = t
}

func (c *Checker) markBreak() {
	if c.fn != nil && len(c.fn.loops) > 0 {
		c.fn.loops[len(c.fn.loops)-1].hasBreak = true
	}
}

// resetAssigned drops the narrowing of every local a loop body assigns, since the assignment
// may have run before the next test of the condition.
func (c *Checker) resetAssigned(body *ast.Block, sc *scope) {
	walkAssigned(body, func(name string) {
		if l, _ := c.lookup(sc, name); l != nil {
			delete(c.flow, l)
		}
	})
}

// walkAssigned calls visit with the name of every variable a statement list assigns, at any
// depth including nested bodies.
func walkAssigned(n ast.Node, visit func(string)) {
	switch x := n.(type) {
	case nil:
	case *ast.Block:
		for _, s := range x.Stmts {
			walkAssigned(s, visit)
		}
	case *ast.VarDecl:
		walkAssigned(x.Init, visit)
	case *ast.Assign:
		assignedNames(x.Target, visit)
		walkAssigned(x.Target, visit) // the places' own sub-expressions, as m[(u = null)]
		walkAssigned(x.Value, visit)
	case *ast.AssignArray:
		for _, e := range x.Elems {
			walkAssigned(e, visit)
		}
		if x.Rest != nil {
			walkAssigned(x.Rest, visit)
		}
	case *ast.If:
		walkAssigned(x.Cond, visit)
		walkAssigned(x.Then, visit)
		walkAssigned(x.Else, visit)
	case *ast.While:
		walkAssigned(x.Cond, visit)
		walkAssigned(x.Body, visit)
	case *ast.For:
		walkAssigned(x.Iter, visit)
		walkAssigned(x.Body, visit)
	case *ast.Select:
		for _, cl := range x.Clauses {
			if cl.Head != nil {
				walkAssigned(cl.Head, visit)
			}
			walkAssigned(cl.Body, visit)
		}
	case *ast.Go:
		walkAssigned(x.Body, visit)
	case *ast.Defer:
		walkAssigned(x.Body, visit)
	case *ast.Lambda:
		walkAssigned(x.Body, visit)
	case *ast.FuncLit:
		walkAssigned(x.Body, visit)
	case *ast.Call:
		walkAssigned(x.Fn, visit)
		for _, a := range x.Args {
			walkAssigned(a.Value, visit)
		}
	case *ast.Binary:
		walkAssigned(x.X, visit)
		walkAssigned(x.Y, visit)
	case *ast.Unary:
		walkAssigned(x.X, visit)
	case *ast.Field:
		walkAssigned(x.X, visit)
	case *ast.Index:
		walkAssigned(x.X, visit)
		walkAssigned(x.Index, visit)
	case *ast.Return:
		walkAssigned(x.Value, visit)
	case *ast.Template:
		for _, p := range x.Parts {
			walkAssigned(p.Expr, visit)
		}
	case *ast.ArrayLit:
		for _, e := range x.Elems {
			walkAssigned(e.Value, visit)
		}
	case *ast.MapLit:
		for _, e := range x.Entries {
			walkAssigned(e.Key, visit)
			walkAssigned(e.Value, visit)
		}
	case *ast.RecordLit:
		for _, e := range x.Entries {
			walkAssigned(e.Value, visit)
			walkAssigned(e.Spread, visit)
		}
	case *ast.Is:
		walkAssigned(x.X, visit)
	}
}

func assignedNames(t ast.Node, visit func(string)) {
	switch x := t.(type) {
	case *ast.Ident:
		visit(x.Name)
	case *ast.AssignArray:
		for _, e := range x.Elems {
			assignedNames(e, visit)
		}
		if x.Rest != nil {
			assignedNames(x.Rest, visit)
		}
	}
}

// ---- conditions

// cond checks a condition and returns the flow when it holds and when it does not.
func (c *Checker) cond(e ast.Expr, sc *scope) (whenTrue, whenFalse flow) {
	switch x := e.(type) {
	case *ast.Binary:
		switch x.Op {
		case "&&":
			t1, f1 := c.cond(x.X, sc)
			c.flow = t1
			t2, f2 := c.cond(x.Y, sc)
			return t2, joinFlows(f1, f2)
		case "||":
			t1, f1 := c.cond(x.X, sc)
			c.flow = f1
			t2, f2 := c.cond(x.Y, sc)
			return joinFlows(t1, t2), f2
		case "==", "!=":
			if l, ok := c.nullTest(x, sc); ok {
				t := c.typeOfLocal(l)
				c.binary(x, sc)
				if !types.Includes(t, types.Null{}) {
					return c.flow.copy(), c.flow.copy()
				}
				isNull, notNull := c.flow.copy(), c.flow.copy()
				c.setIn(isNull, l, types.Null{})
				c.setIn(notNull, l, types.Without(t, types.Null{}))
				if x.Op == "==" {
					return isNull, notNull
				}
				return notNull, isNull
			}
		}
	case *ast.Unary:
		if x.Op == "!" {
			t, f := c.cond(x.X, sc)
			return f, t
		}
	case *ast.Is:
		if id, ok := x.X.(*ast.Ident); ok {
			if l, _ := c.lookup(sc, id.Name); l != nil {
				t := c.typeOfLocal(l)
				target := c.isTarget(x, sc)
				if target == nil || types.IsInvalid(t) {
					return c.flow.copy(), c.flow.copy()
				}
				wt, wf, ok := c.narrowTo(t, target, x.Pos)
				if !ok {
					return c.flow.copy(), c.flow.copy()
				}
				yes, no := c.flow.copy(), c.flow.copy()
				c.setIn(yes, l, wt)
				c.setIn(no, l, wf)
				return yes, no
			}
		}
	}
	c.check(e, sc, types.BoolType)
	return c.flow.copy(), c.flow.copy()
}

func (c *Checker) setIn(f flow, l *local, t types.Type) {
	if l.slot.AssignedNested {
		return
	}
	f[l] = t
}

// nullTest recognizes x == null and null == x on a local.
func (c *Checker) nullTest(x *ast.Binary, sc *scope) (*local, bool) {
	var id *ast.Ident
	if _, isNull := x.Y.(*ast.NullLit); isNull {
		id, _ = x.X.(*ast.Ident)
	} else if _, isNull := x.X.(*ast.NullLit); isNull {
		id, _ = x.Y.(*ast.Ident)
	}
	if id == nil {
		return nil, false
	}
	l, _ := c.lookup(sc, id.Name)
	if l == nil {
		return nil, false
	}
	return l, true
}

// ---- blocks and statements

func (c *Checker) synthBlock(b *ast.Block, sc *scope) types.Type {
	return c.block(b, sc, nil)
}

func (c *Checker) checkBlock(b *ast.Block, sc *scope, expected types.Type) types.Type {
	return c.block(b, sc, expected)
}

// block checks a block's statements in a new scope. Its type is that of the last expression,
// checked against expected when there is one, or void.
func (c *Checker) block(b *ast.Block, sc *scope, expected types.Type) types.Type {
	inner := &scope{parent: sc, names: map[string]*local{}}
	var last types.Type = types.Void{}
	for i, s := range b.Stmts {
		isLast := i == len(b.Stmts)-1
		switch st := s.(type) {
		case *ast.VarDecl:
			c.varDecl(st, inner)
			last = types.Void{}
		case ast.Expr:
			if isLast && expected != nil {
				last = c.check(st, inner, expected)
			} else {
				last = c.synth(st, inner)
			}
			if _, isAssign := st.(*ast.Assign); isAssign {
				last = types.Void{}
			}
		}
	}
	if expected != nil {
		if len(b.Stmts) == 0 {
			c.expect(b.Pos, types.Void{}, expected)
		} else if _, isVar := b.Stmts[len(b.Stmts)-1].(*ast.VarDecl); isVar {
			c.expect(b.Pos, types.Void{}, expected)
		}
	}
	return last
}

// varDecl checks a local var declaration. The name is in scope inside its own initializer,
// where a direct read is an error and a read from a nested function body needs the type
// stated (section 3).
func (c *Checker) varDecl(d *ast.VarDecl, sc *scope) {
	var declared types.Type
	if d.Type != nil {
		declared = c.resolveType(d.Type, sc)
	}
	var locals []*local
	for _, nt := range targetNames(d.Target) {
		t := declared
		if t == nil {
			t = types.Invalid{}
		}
		l := &local{name: nt.Name, typ: t, slot: nt.Slot, ctx: c.fn, pos: nt.Pos, pending: true, stated: declared != nil}
		sc.names[nt.Name] = l
		locals = append(locals, l)
	}
	if declared == nil && len(locals) == 1 {
		// A func literal with a stated return type gives the var its type before the body is
		// checked, which is how a local function calls itself (section 6).
		if fl, ok := d.Init.(*ast.FuncLit); ok && fl.Return != nil {
			s := c.signature(&ast.FuncDecl{Pos: fl.Pos, Params: fl.Params, Return: fl.Return, Body: fl.Body}, sc, "")
			locals[0].typ = s.fn
			locals[0].stated = true
		}
	}
	var t types.Type
	if declared != nil {
		c.check(d.Init, sc, declared)
		t = declared
	} else {
		t = c.synth(d.Init, sc)
	}
	c.bindTargetTypes(d.Target, t, sc, func(name string, ty types.Type) {
		for _, l := range locals {
			if l.name == name {
				if !l.stated {
					l.typ = ty
				}
				l.pending = false
			}
		}
	})
	for _, l := range locals {
		l.pending = false
	}
}

// bindTargetTypes computes the type each name of a target binds from the source type.
func (c *Checker) bindTargetTypes(t ast.Target, src types.Type, sc *scope, bind func(string, types.Type)) {
	switch t := t.(type) {
	case *ast.NameTarget:
		if t.Name != "_" {
			bind(t.Name, src)
		}
	case *ast.RecordTarget:
		if types.FieldsOf(src) == nil && !types.IsInvalid(src) {
			if _, isUnion := types.Expand(src).(types.Union); !isUnion {
				c.errorf(t.Pos, "a record destructuring needs a record, got %s", types.String(src))
			}
		}
		for _, f := range t.Fields {
			ft := types.Type(types.Invalid{})
			if !types.IsInvalid(src) {
				ft = c.fieldOf(src, f.Name, f.Pos)
			}
			c.bindTargetTypes(f.Target, ft, sc, bind)
		}
	case *ast.ArrayTarget:
		elem := types.Type(types.Invalid{})
		if a, ok := types.Expand(src).(types.Array); ok {
			elem = a.Elem
		} else if !types.IsInvalid(src) {
			c.errorf(t.Pos, "an array destructuring needs an array, got %s", types.String(src))
		}
		for _, e := range t.Elems {
			c.bindTargetTypes(e, elem, sc, bind)
		}
		if t.Rest != nil {
			c.bindTargetTypes(t.Rest, types.Array{Elem: elem}, sc, bind)
		}
	}
}

// declareTarget binds a for target's names as locals in sc.
func (c *Checker) declareTarget(t ast.Target, src types.Type, sc *scope) {
	slots := map[string]ast.Slot{}
	positions := map[string]ast.Pos{}
	for _, nt := range targetNames(t) {
		slots[nt.Name] = nt.Slot
		positions[nt.Name] = nt.Pos
	}
	c.bindTargetTypes(t, src, sc, func(name string, ty types.Type) {
		sc.names[name] = &local{name: name, typ: ty, slot: slots[name], ctx: c.fn, pos: positions[name]}
	})
}

// ---- control flow

func (c *Checker) checkIf(x *ast.If, sc *scope, expected types.Type) types.Type {
	wt, wf := c.cond(x.Cond, sc)
	c.flow = wt
	thenT := c.block(x.Then, sc, expected)
	afterThen := c.flow
	c.flow = wf
	complete := narrowedToNever(wt, wf)
	var elseT types.Type
	switch el := x.Else.(type) {
	case nil:
		if complete {
			elseT = types.Never{}
		} else {
			elseT = types.Void{}
		}
	case *ast.If:
		if complete {
			c.errorf(el.Pos, "this else cannot run: the tests before it cover every case")
		}
		elseT = c.checkIf(el, sc, expected)
	case *ast.Block:
		if complete {
			c.errorf(el.Pos, "this else cannot run: the tests before it cover every case")
		}
		elseT = c.block(el, sc, expected)
	}
	afterElse := c.flow
	switch {
	case types.IsNever(thenT) && types.IsNever(elseT):
		c.flow = afterThen
	case types.IsNever(thenT):
		c.flow = afterElse
	case types.IsNever(elseT):
		c.flow = afterThen
	default:
		c.flow = joinFlows(afterThen, afterElse)
	}
	if x.Else == nil && !complete && expected != nil && !types.IsNever(thenT) {
		if !types.Assignable(types.Void{}, expected) {
			c.errorf(x.Pos, "an if without else has type void when the condition is false, which cannot be used where %s is expected", types.String(expected))
		}
	}
	if expected != nil {
		// Each branch was checked against the expected type and any mismatch is reported
		// there; reporting the union again at the if would be a second error for one mistake.
		if types.IsInvalid(expected) {
			return types.Invalid{}
		}
		return expected
	}
	u, err := types.MakeUnion([]types.Type{thenT, elseT})
	if err != nil {
		c.errorf(x.Pos, "the branches of this if do not form a valid type: %s", err)
		return types.Invalid{}
	}
	return u
}

// selectHeadArity is the argument count each head of a select takes. A head gives no context,
// which is a clause of its own, so an argument beyond these is an error (section 5).
var selectHeadArity = map[string]int{
	"Channel.receive": 1, "Channel.send": 2, "Context.wait": 1, "WaitGroup.wait": 1, "Time.sleep": 1,
}

// selectHeadName is the head's function, as the parser checked it: a two-part path.
func selectHeadName(head *ast.Call) string {
	f := head.Fn.(*ast.Field)
	return f.X.(*ast.Ident).Name + "." + f.Name
}

// checkSelect types a select (section 5): each head as the call it names, with its required
// arguments only, its result bound in the clause's scope when the clause binds; each body
// against the expected type, or the select's type is the union of the bodies' types with never
// dropped, as an if. The heads' operands belong to the enclosing scope, and every body starts
// from the flow before the select, since exactly one runs.
func (c *Checker) checkSelect(x *ast.Select, sc *scope, expected types.Type) types.Type {
	before := c.flow.copy()
	var after []flow
	var bodyTypes []types.Type
	for _, cl := range x.Clauses {
		c.flow = before.copy()
		inner := &scope{parent: sc, names: map[string]*local{}}
		if cl.Head != nil {
			name := selectHeadName(cl.Head)
			if len(cl.Head.Args) > selectHeadArity[name] {
				c.errorf(cl.Head.Pos, "a select clause's head takes no context: the context is a clause of its own, Context.wait(ctx) { ... }")
			}
			t := c.call(cl.Head, sc, nil)
			if cl.Target != nil {
				if _, isVoid := types.Expand(t).(types.Void); isVoid {
					c.errorf(cl.Target.Pos, "%s returns void, so there is nothing to bind to %s", name, cl.Target.Name)
					t = types.Invalid{}
				}
				c.declareTarget(cl.Target, t, inner)
			}
		}
		t := c.block(cl.Body, inner, expected)
		bodyTypes = append(bodyTypes, t)
		if !types.IsNever(t) {
			after = append(after, c.flow)
		}
	}
	if len(after) > 0 {
		f := after[0]
		for _, g := range after[1:] {
			f = joinFlows(f, g)
		}
		c.flow = f
	}
	if expected != nil {
		if types.IsInvalid(expected) {
			return types.Invalid{}
		}
		return expected
	}
	u, err := types.MakeUnion(bodyTypes)
	if err != nil {
		c.errorf(x.Pos, "the clauses of this select do not form a valid type: %s", err)
		return types.Invalid{}
	}
	return u
}

// narrowedToNever reports whether the false path narrowed some local to never, which makes
// a chain of tests complete.
func narrowedToNever(wt, wf flow) bool {
	for l, t := range wf {
		if types.IsNever(t) {
			if u, ok := wt[l]; !ok || !types.IsNever(u) {
				return true
			}
		}
	}
	return false
}

func (c *Checker) whileLoop(x *ast.While, sc *scope) types.Type {
	before := c.flow.copy()
	c.resetAssigned(x.Body, sc)
	loop := &loopCtx{}
	c.fn.loops = append(c.fn.loops, loop)
	wt, wf := c.cond(x.Cond, sc)
	c.flow = wt
	c.synthBlock(x.Body, sc)
	afterBody := c.flow
	c.fn.loops = c.fn.loops[:len(c.fn.loops)-1]
	if loop.hasBreak {
		c.flow = joinFlows(wf, afterBody)
	} else {
		c.flow = wf
	}
	_ = before
	if lit, ok := x.Cond.(*ast.BoolLit); ok && lit.Value && !loop.hasBreak {
		return types.Never{}
	}
	return types.Void{}
}

func (c *Checker) forLoop(x *ast.For, sc *scope) types.Type {
	it := c.synth(x.Iter, sc)
	elem := c.elementType(it, x.Iter.Position())
	inner := &scope{parent: sc, names: map[string]*local{}}
	c.declareTarget(x.Target, elem, inner)
	c.resetAssigned(x.Body, sc)
	before := c.flow.copy()
	loop := &loopCtx{}
	c.fn.loops = append(c.fn.loops, loop)
	c.synthBlock(x.Body, inner)
	c.fn.loops = c.fn.loops[:len(c.fn.loops)-1]
	c.flow = joinFlows(before, c.flow)
	return types.Void{}
}

// elementType is what a for binds from an iterable: an array's element, a map's Entry, a set's
// member, a string's code point, or a sequence's element.
func (c *Checker) elementType(it types.Type, pos ast.Pos) types.Type {
	switch x := types.Expand(it).(type) {
	case types.Invalid:
		return types.Invalid{}
	case types.Array:
		return x.Elem
	case types.Map:
		return types.Named{Decl: c.prelude["Entry"].decl, Args: []types.Type{x.Key, x.Val}}
	case types.Set:
		return x.Elem
	case types.Prim:
		if x.Kind.String() == "string" {
			return types.StringType
		}
	case types.Func:
		if len(x.Params) == 1 {
			if y, ok := types.Expand(x.Params[0]).(types.Func); ok && len(y.Params) == 1 && types.Equal(y.Return, types.BoolType) {
				return y.Params[0]
			}
		}
	}
	c.errorf(pos, "for needs an array, a map, a set, a string, or a sequence, got %s", types.String(it))
	return types.Invalid{}
}

func (c *Checker) returnExpr(x *ast.Return, sc *scope) types.Type {
	fn := c.fn
	if fn == nil || fn.kind == "init" {
		c.errorf(x.Pos, "return outside a function")
		if x.Value != nil {
			c.synth(x.Value, sc)
		}
		return types.Never{}
	}
	if fn.kind == "go" || fn.kind == "defer" {
		return types.Never{}
	}
	if fn.declared != nil {
		if x.Value == nil {
			c.expect(x.Pos, types.Void{}, fn.declared)
		} else {
			c.check(x.Value, sc, fn.declared)
		}
		return types.Never{}
	}
	if x.Value == nil {
		fn.returns = append(fn.returns, types.Void{})
	} else {
		fn.returns = append(fn.returns, c.synth(x.Value, sc))
	}
	return types.Never{}
}
