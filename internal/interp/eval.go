package interp

import (
	"fmt"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/value"
)

// ctl is how control leaves an expression: normally, or by return, break, or continue.
type ctl int

const (
	ctlNone ctl = iota
	ctlReturn
	ctlBreak
	ctlContinue
)

// loopUnwind carries a return out of a for body that a sequence function is running.
type loopUnwind struct {
	loop *ast.For
	v    value.Value
}

// evalValue evaluates an expression that cannot transfer control, such as a top-level
// initializer.
func (g *Goroutine) evalValue(e ast.Expr, env *Env) value.Value {
	v, c := g.eval(e, env)
	if c != ctlNone {
		g.panicf(e.Position(), "internal error: control left an initializer")
	}
	return v
}

func (g *Goroutine) evalStmts(stmts []ast.Stmt, env *Env) (value.Value, ctl) {
	var last value.Value = value.Void
	for _, s := range stmts {
		switch st := s.(type) {
		case *ast.VarDecl:
			v, c := g.eval(st.Init, env)
			if c != ctlNone {
				return v, c
			}
			g.bind(st.Target, v, env)
			last = value.Void
		case ast.Expr:
			v, c := g.eval(st, env)
			if c != ctlNone {
				return v, c
			}
			last = v
		}
	}
	return last, ctlNone
}

func (g *Goroutine) evalBlock(b *ast.Block, env *Env) (value.Value, ctl) {
	if b.Scope != nil {
		env = newEnv(b.Scope, env)
	}
	if !b.HasDefer {
		return g.evalStmts(b.Stmts, env)
	}
	return g.evalDeferBlock(b, env)
}

// evalDeferBlock runs a block that registers deferred blocks, and runs them in reverse order
// when the block exits, whether normally, by control transfer, or by a panic. A panic in a
// deferred block replaces the one in progress.
func (g *Goroutine) evalDeferBlock(b *ast.Block, env *Env) (v value.Value, c ctl) {
	frame := &deferFrame{}
	g.defers = append(g.defers, frame)
	defer func() {
		g.defers = g.defers[:len(g.defers)-1]
		r := recover()
		if _, exiting := r.(*exitSignal); exiting {
			panic(r)
		}
		for i := len(frame.list) - 1; i >= 0; i-- {
			if p := g.runDeferred(frame.list[i]); p != nil {
				r = p
			}
		}
		if r != nil {
			panic(r)
		}
	}()
	return g.evalStmts(b.Stmts, env)
}

func (g *Goroutine) runDeferred(d deferred) (p any) {
	mark := len(g.args)
	defer func() {
		p = recover()
		g.unwindArgs(mark)
	}()
	g.evalBlock(d.body, newEnv(&d.scope.BlockScope, d.env))
	return nil
}

func (g *Goroutine) eval(e ast.Expr, env *Env) (value.Value, ctl) {
	switch e := e.(type) {
	case *ast.Ident:
		return g.ident(e, env), ctlNone
	case *ast.IntLit:
		return e.Value, ctlNone
	case *ast.FloatLit:
		return e.Value, ctlNone
	case *ast.StringLit:
		return e.Value, ctlNone
	case *ast.BoolLit:
		return e.Value, ctlNone
	case *ast.NullLit:
		return value.Null, ctlNone
	case *ast.VoidLit:
		return value.Void, ctlNone
	case *ast.Template:
		var sb strings.Builder
		for _, p := range e.Parts {
			if p.Expr == nil {
				sb.WriteString(p.Text)
				continue
			}
			v, c := g.eval(p.Expr, env)
			if c != ctlNone {
				return v, c
			}
			sb.WriteString(value.Render(v))
		}
		return sb.String(), ctlNone
	case *ast.RecordLit:
		return g.recordLit(e, env)
	case *ast.MapLit:
		m := value.NewMap()
		for _, en := range e.Entries {
			k, c := g.eval(en.Key, env)
			if c != ctlNone {
				return k, c
			}
			v, c := g.eval(en.Value, env)
			if c != ctlNone {
				return v, c
			}
			m.Set(k, v)
		}
		return m, ctlNone
	case *ast.ArrayLit:
		elems := make([]value.Value, 0, len(e.Elems))
		for _, el := range e.Elems {
			v, c := g.eval(el.Value, env)
			if c != ctlNone {
				return v, c
			}
			if el.Spread {
				arr, ok := v.(*value.Array)
				if !ok {
					g.panicf(el.Value.Position(), "a spread in an array literal needs an array, got %s", value.Describe(v))
				}
				elems = append(elems, arr.Values()...)
			} else {
				elems = append(elems, v)
			}
		}
		return value.NewArray(elems), ctlNone
	case *ast.Lambda:
		return &Func{Body: e.Body, Scope: e.Scope, Env: env, Lambda: true}, ctlNone
	case *ast.FuncLit:
		return &Func{Params: e.Params, Body: e.Body, Scope: e.Scope, Env: env}, ctlNone
	case *ast.Field, *ast.Index, *ast.Call:
		v, c, _ := g.postfix(e, env)
		return v, c
	case *ast.Unary:
		return g.unary(e, env)
	case *ast.Binary:
		return g.binary(e, env)
	case *ast.Is:
		return g.is(e, env)
	case *ast.Assign:
		return g.assign(e, env)
	case *ast.If:
		cond, c := g.eval(e.Cond, env)
		if c != ctlNone {
			return cond, c
		}
		b, ok := cond.(bool)
		if !ok {
			g.panicf(e.Cond.Position(), "an if condition must be a bool, got %s", value.Describe(cond))
		}
		if b {
			return g.evalBlock(e.Then, env)
		}
		switch el := e.Else.(type) {
		case *ast.If:
			return g.eval(el, env)
		case *ast.Block:
			return g.evalBlock(el, env)
		}
		return value.Void, ctlNone
	case *ast.While:
		for {
			g.check()
			cond, c := g.eval(e.Cond, env)
			if c != ctlNone {
				return cond, c
			}
			b, ok := cond.(bool)
			if !ok {
				g.panicf(e.Cond.Position(), "a while condition must be a bool, got %s", value.Describe(cond))
			}
			if !b {
				return value.Void, ctlNone
			}
			v, c := g.evalBlock(e.Body, env)
			switch c {
			case ctlBreak:
				return value.Void, ctlNone
			case ctlReturn:
				return v, c
			}
		}
	case *ast.For:
		return g.forLoop(e, env)
	case *ast.Go:
		body, scope := e.Body, e.Scope
		path, pos := env.mod.m.File.Path, e.Pos
		g.spawn(func(ng *Goroutine) {
			// the block is the goroutine's root frame, so its panics report a position
			ng.push("go block", path, pos)
			defer ng.pop()
			ng.evalBlock(body, newEnv(&scope.BlockScope, env))
		})
		return value.Void, ctlNone
	case *ast.Defer:
		if len(g.defers) == 0 {
			g.panicf(e.Pos, "defer outside a block")
		}
		top := g.defers[len(g.defers)-1]
		top.list = append(top.list, deferred{body: e.Body, scope: e.Scope, env: env})
		return value.Void, ctlNone
	case *ast.Return:
		if e.Value == nil {
			return value.Void, ctlReturn
		}
		v, c := g.eval(e.Value, env)
		if c != ctlNone {
			return v, c
		}
		return v, ctlReturn
	case *ast.Break:
		return nil, ctlBreak
	case *ast.Continue:
		return nil, ctlContinue
	case *ast.Block:
		return g.evalBlock(e, env)
	case *ast.Placeholder:
		g.panicf(e.Pos, "a ... placeholder cannot run")
	}
	g.panicf(e.Position(), "internal error: cannot evaluate %T", e)
	return nil, ctlNone
}

// ---- names

func (g *Goroutine) ident(e *ast.Ident, env *Env) value.Value {
	switch e.Ref.Kind {
	case ast.RefLocal:
		v := env.up(e.Ref.Depth).get(ast.Slot{Index: e.Ref.Index, Shared: e.Ref.Shared})
		if v == value.Uninitialized {
			g.panicf(e.Pos, "%s is read before its initializer has run", e.Name)
		}
		return v
	case ast.RefModule:
		v := env.mod.rec.Get(e.Ref.Index)
		if v == value.Uninitialized {
			g.panicf(e.Pos, "%s is read before its initializer has run", e.Name)
		}
		return v
	case ast.RefImport:
		return g.importValue(env, e.Ref, e.Pos)
	case ast.RefPrelude:
		v, ok := g.in.cfg.Prelude[e.Name]
		if !ok {
			g.panicf(e.Pos, "%s is not available in this run", e.Name)
		}
		return v
	case ast.RefDiscard:
		g.panicf(e.Pos, "_ cannot be read")
	}
	g.panicf(e.Pos, "%s is not defined", e.Name)
	return nil
}

func (g *Goroutine) assignIdent(e *ast.Ident, v value.Value, env *Env) {
	switch e.Ref.Kind {
	case ast.RefLocal:
		env.up(e.Ref.Depth).set(ast.Slot{Index: e.Ref.Index, Shared: e.Ref.Shared}, v)
	case ast.RefModule:
		env.mod.rec.Set(e.Ref.Index, v)
	case ast.RefImport:
		ir := env.mod.imports[e.Ref.Index]
		if ir.field == -1 {
			g.panicf(e.Pos, "cannot assign to the namespace %s", e.Name)
		}
		if !ir.rec.SetField(e.Ref.Name, v) {
			g.panicf(e.Pos, "%s has no export %s", ir.rec.Tag.Name, e.Ref.Name)
		}
	case ast.RefPrelude:
		g.panicf(e.Pos, "cannot assign to the prelude name %s", e.Name)
	case ast.RefDiscard:
	default:
		g.panicf(e.Pos, "%s is not defined", e.Name)
	}
}

// ---- postfix chains

// postfix evaluates a field read, index, or call, skipping the rest of the chain after a ?.
// that met null. The third result reports that skip.
func (g *Goroutine) postfix(e ast.Expr, env *Env) (value.Value, ctl, bool) {
	switch x := e.(type) {
	case *ast.Field:
		recv, c, skip := g.postfix(x.X, env)
		if c != ctlNone {
			return recv, c, false
		}
		if skip || (x.Optional && recv == value.Null) {
			return value.Null, ctlNone, true
		}
		return g.field(recv, x.Name, x.Pos), ctlNone, false
	case *ast.Index:
		recv, c, skip := g.postfix(x.X, env)
		if c != ctlNone {
			return recv, c, false
		}
		if skip || (x.Optional && recv == value.Null) {
			return value.Null, ctlNone, true
		}
		idx, c := g.eval(x.Index, env)
		if c != ctlNone {
			return idx, c, false
		}
		return g.index(recv, idx, x.Pos), ctlNone, false
	case *ast.Call:
		return g.call(x, env)
	default:
		v, c := g.eval(e, env)
		return v, c, false
	}
}

// call evaluates a call. Arguments to an Ant function with no named arguments go on the
// goroutine's argument stack, which the callee copies into its frame before its body runs, so
// such a call allocates no argument slice. Arguments to anything else get their own slice,
// because a builtin may keep it.
func (g *Goroutine) call(x *ast.Call, env *Env) (value.Value, ctl, bool) {
	var pipe value.Value
	rest := x.Args
	if x.Pipe {
		v, c := g.eval(x.Args[0].Value, env)
		if c != ctlNone {
			return v, c, false
		}
		pipe = v
		rest = x.Args[1:]
	}
	fn, c, skip := g.postfix(x.Fn, env)
	if c != ctlNone {
		return fn, c, false
	}
	if skip || (x.Optional && fn == value.Null) {
		return value.Null, ctlNone, true
	}
	if f, ok := fn.(*Func); ok && !namedArgs(rest) {
		base := len(g.args)
		if x.Pipe {
			g.args = append(g.args, pipe)
		}
		for _, a := range rest {
			v, c := g.eval(a.Value, env)
			if c != ctlNone {
				g.args = g.args[:base]
				return v, c, false
			}
			g.args = append(g.args, v)
		}
		g.check()
		return g.callFuncStack(f, base, x.Pos), ctlNone, false
	}
	var args []value.Value
	var names []string
	if x.Pipe {
		args = append(args, pipe)
	}
	for _, a := range rest {
		v, c := g.eval(a.Value, env)
		if c != ctlNone {
			return v, c, false
		}
		if a.Name != "" && names == nil {
			names = make([]string, len(args), len(x.Args))
		}
		args = append(args, v)
		if names != nil {
			names = append(names, a.Name)
		}
	}
	return g.callValue(fn, args, names, x.Pos), ctlNone, false
}

func namedArgs(args []ast.Arg) bool {
	for _, a := range args {
		if a.Name != "" {
			return true
		}
	}
	return false
}

func (g *Goroutine) field(recv value.Value, name string, pos ast.Pos) value.Value {
	switch x := recv.(type) {
	case *value.Record:
		v, ok := x.Field(name)
		if !ok {
			g.panicf(pos, "%s has no field %s", x.Tag.Name, name)
		}
		if v == value.Uninitialized {
			g.panicf(pos, "%s is read before its initializer has run", name)
		}
		return v
	case *value.Tag:
		fn, ok := x.Func(name)
		if !ok {
			g.panicf(pos, "%s has no function %s", x.Name, name)
		}
		return fn
	case *value.Constructor:
		fn, ok := x.Tag.Func(name)
		if !ok {
			g.panicf(pos, "%s has no function %s", x.Tag.Name, name)
		}
		return fn
	}
	g.panicf(pos, "%s is not a record, so it has no field %s", value.Describe(recv), name)
	return nil
}

func (g *Goroutine) index(recv, idx value.Value, pos ast.Pos) value.Value {
	switch x := recv.(type) {
	case *value.Array:
		i, ok := idx.(int64)
		if !ok {
			g.panicf(pos, "an array index must be an int, got %s", value.Describe(idx))
		}
		v, ok := x.Get(int(i))
		if !ok {
			g.panicf(pos, "index %d out of range for an array of length %d", i, x.Len())
		}
		return v
	case *value.Map:
		v, ok := x.Get(idx)
		if !ok {
			return value.Null
		}
		return v
	}
	g.panicf(pos, "%s is not an array or a map, so it cannot be indexed", value.Describe(recv))
	return nil
}

// ---- calls

// Call calls an Ant function value from a builtin.
func (g *Goroutine) Call(fn value.Value, args ...value.Value) value.Value {
	return g.callValue(fn, args, nil, g.pos)
}

func (g *Goroutine) callValue(fn value.Value, args []value.Value, names []string, pos ast.Pos) value.Value {
	g.check()
	switch f := fn.(type) {
	case *Func:
		return g.callFunc(f, args, names, pos)
	case *Builtin:
		return g.callBuiltin(f, args, names, pos)
	case *value.Constructor:
		if len(args) != 1 || names != nil {
			g.panicf(pos, "the constructor %s takes exactly one positional argument", f.Tag.Name)
		}
		if value.KindOf(args[0]) != f.Tag.Base {
			g.panicf(pos, "%s needs a %s, got %s", f.Tag.Name, f.Tag.Base, value.Describe(args[0]))
		}
		return &value.Tagged{Tag: f.Tag, Payload: args[0]}
	}
	g.panicf(pos, "%s is not a function and cannot be called", value.Describe(fn))
	return nil
}

func (g *Goroutine) push(name, path string, pos ast.Pos) {
	if g.depth >= maxDepth {
		g.panicf(pos, "call depth exceeded %d", maxDepth)
	}
	g.depth++
	g.stack = append(g.stack, Frame{Function: name, Path: path, Pos: pos})
}

func (g *Goroutine) pop() {
	g.depth--
	g.stack = g.stack[:len(g.stack)-1]
}

func (g *Goroutine) callFunc(f *Func, args []value.Value, names []string, pos ast.Pos) value.Value {
	g.push(f.FuncName(), f.Env.mod.m.File.Path, pos)
	defer g.pop()
	env := g.frame(f, args, names, pos)
	v, _ := g.evalBlock(f.Body, env)
	return v
}

// callFuncStack calls f with the arguments on the argument stack from base up, and pops them
// once they are bound.
func (g *Goroutine) callFuncStack(f *Func, base int, pos ast.Pos) value.Value {
	g.push(f.FuncName(), f.Env.mod.m.File.Path, pos)
	defer g.pop()
	env := g.frame(f, g.args[base:], nil, pos)
	clear(g.args[base:])
	g.args = g.args[:base]
	v, _ := g.evalBlock(f.Body, env)
	return v
}

// frame opens f's environment and binds the arguments of a call to it.
func (g *Goroutine) frame(f *Func, args []value.Value, names []string, pos ast.Pos) *Env {
	env := newEnv(&f.Scope.BlockScope, f.Env)
	if !f.Lambda && names == nil && len(args) == len(f.Params) {
		for i := range args {
			env.set(f.Params[i].Slot, args[i])
		}
		return env
	}
	if f.Lambda {
		if names != nil {
			g.panicf(pos, "a lambda takes no named arguments")
		}
		if len(args) > 1 {
			g.panicf(pos, "a lambda takes at most one argument, got %d", len(args))
		}
		x := value.Void
		if len(args) == 1 {
			x = args[0]
		}
		env.set(f.Scope.X, x)
	} else {
		filled := make([]bool, len(f.Params))
		np := 0
		for i := range args {
			if names != nil && names[i] != "" {
				break
			}
			np++
		}
		if np > len(f.Params) {
			g.panicf(pos, "%s takes %s, got %d", f.FuncName(), plural(len(f.Params), "argument"), np)
		}
		for i := 0; i < np; i++ {
			env.set(f.Params[i].Slot, args[i])
			filled[i] = true
		}
		for i := np; i < len(args); i++ {
			idx := -1
			for j := range f.Params {
				if f.Params[j].Name == names[i] {
					idx = j
					break
				}
			}
			if idx < 0 {
				g.panicf(pos, "%s has no parameter %s", f.FuncName(), names[i])
			}
			if filled[idx] {
				g.panicf(pos, "the parameter %s is given twice", names[i])
			}
			env.set(f.Params[idx].Slot, args[i])
			filled[idx] = true
		}
		for i := range f.Params {
			if filled[i] {
				continue
			}
			p := &f.Params[i]
			if p.Default == nil {
				g.panicf(pos, "%s needs an argument for %s", f.FuncName(), p.Name)
			}
			v, _ := g.eval(p.Default, env)
			env.set(p.Slot, v)
		}
	}
	return env
}

func (g *Goroutine) callBuiltin(f *Builtin, args []value.Value, names []string, pos ast.Pos) value.Value {
	if names != nil {
		ordered := make([]value.Value, len(f.Params))
		filled := make([]bool, len(f.Params))
		for i, a := range args {
			if names[i] == "" {
				if i >= len(f.Params) {
					g.panicf(pos, "%s takes %s, got %d", f.Name, plural(len(f.Params), "argument"), len(args))
				}
				ordered[i] = a
				filled[i] = true
				continue
			}
			idx := -1
			for j, p := range f.Params {
				if p == names[i] {
					idx = j
					break
				}
			}
			if idx < 0 {
				g.panicf(pos, "%s has no parameter %s", f.Name, names[i])
			}
			if filled[idx] {
				g.panicf(pos, "the parameter %s is given twice", names[i])
			}
			ordered[idx] = a
			filled[idx] = true
		}
		for i, ok := range filled {
			if !ok && i < len(f.Params)-f.Optional {
				g.panicf(pos, "%s needs an argument for %s", f.Name, f.Params[i])
			}
		}
		args = ordered
	} else if len(args) < len(f.Params)-f.Optional || len(args) > len(f.Params) {
		if f.Optional > 0 {
			g.panicf(pos, "%s takes %d to %d arguments, got %d", f.Name, len(f.Params)-f.Optional, len(f.Params), len(args))
		}
		g.panicf(pos, "%s takes %s, got %d", f.Name, plural(len(f.Params), "argument"), len(args))
	} else if len(args) < len(f.Params) {
		full := make([]value.Value, len(f.Params))
		copy(full, args)
		args = full
	}
	saved := g.pos
	g.pos = pos
	defer func() { g.pos = saved }()
	return f.Fn(g, args)
}

// ---- literals

func (g *Goroutine) recordLit(e *ast.RecordLit, env *Env) (value.Value, ctl) {
	head, c := g.eval(e.Head, env)
	if c != ctlNone {
		return head, c
	}
	tag, ok := head.(*value.Tag)
	if !ok || tag.Kind != value.RecordTag {
		name := value.Describe(head)
		if ok {
			name = tag.Name
		}
		g.panicf(e.Pos, "%s is not a record type, so it has no literal", name)
	}
	vals := make([]value.Value, len(tag.Fields))
	for _, en := range e.Entries {
		if en.Spread != nil {
			v, c := g.eval(en.Spread, env)
			if c != ctlNone {
				return v, c
			}
			src, ok := v.(*value.Record)
			if !ok {
				g.panicf(en.Pos, "a spread in a record literal needs a record, got %s", value.Describe(v))
			}
			for i, name := range src.Tag.Fields {
				idx, ok := tag.FieldIndex(name)
				if !ok {
					g.panicf(en.Pos, "%s has no field %s", tag.Name, name)
				}
				vals[idx] = src.Get(i)
			}
			continue
		}
		v, c := g.eval(en.Value, env)
		if c != ctlNone {
			return v, c
		}
		idx, ok := tag.FieldIndex(en.Name)
		if !ok {
			g.panicf(en.Pos, "%s has no field %s", tag.Name, en.Name)
		}
		vals[idx] = v
	}
	for i, v := range vals {
		if v == nil {
			g.panicf(e.Pos, "%s literal is missing the field %s", tag.Name, tag.Fields[i])
		}
	}
	return value.NewRecord(tag, vals), ctlNone
}

// ---- operators

func (g *Goroutine) unary(e *ast.Unary, env *Env) (value.Value, ctl) {
	v, c := g.eval(e.X, env)
	if c != ctlNone {
		return v, c
	}
	switch e.Op {
	case "-":
		r, err := value.Neg(v)
		if err != nil {
			g.panicf(e.Pos, "%s", err)
		}
		return r, ctlNone
	case "!":
		b, ok := v.(bool)
		if !ok {
			g.panicf(e.Pos, "! needs a bool, got %s", value.Describe(v))
		}
		return !b, ctlNone
	case "try":
		rec, ok := v.(*value.Record)
		if ok && rec.Tag == value.OkTag {
			return rec.Get(0), ctlNone
		}
		if ok && rec.Tag == value.ErrTag {
			return value.Err(rec.Get(0)), ctlReturn
		}
		g.panicf(e.Pos, "try needs an Ok or an Err, got %s", value.Describe(v))
	}
	g.panicf(e.Pos, "internal error: unary %s", e.Op)
	return nil, ctlNone
}

func (g *Goroutine) binary(e *ast.Binary, env *Env) (value.Value, ctl) {
	l, c := g.eval(e.X, env)
	if c != ctlNone {
		return l, c
	}
	switch e.Op {
	case "&&", "||":
		lb, ok := l.(bool)
		if !ok {
			g.panicf(e.Pos, "%s needs a bool on the left, got %s", e.Op, value.Describe(l))
		}
		if (e.Op == "&&") != lb {
			return lb, ctlNone
		}
		r, c := g.eval(e.Y, env)
		if c != ctlNone {
			return r, c
		}
		rb, ok := r.(bool)
		if !ok {
			g.panicf(e.Pos, "%s needs a bool on the right, got %s", e.Op, value.Describe(r))
		}
		return rb, ctlNone
	case "??":
		if l != value.Null {
			return l, ctlNone
		}
		return g.eval(e.Y, env)
	}
	r, c := g.eval(e.Y, env)
	if c != ctlNone {
		return r, c
	}
	return g.arith(e.Op, l, r, e.Pos), ctlNone
}

func (g *Goroutine) arith(op string, l, r value.Value, pos ast.Pos) value.Value {
	switch op {
	case "==":
		return value.Equal(l, r)
	case "!=":
		return !value.Equal(l, r)
	case "<", "<=", ">", ">=":
		b, err := value.Compare(op, l, r)
		if err != nil {
			g.panicf(pos, "%s", err)
		}
		return b
	}
	v, err := value.Binary(op, l, r)
	if err != nil {
		g.panicf(pos, "%s", err)
	}
	return v
}

func (g *Goroutine) is(e *ast.Is, env *Env) (value.Value, ctl) {
	v, c := g.eval(e.X, env)
	if c != ctlNone {
		return v, c
	}
	if e.Kind != "" {
		k, _ := value.KindByName(e.Kind)
		return value.KindOf(v) == k, ctlNone
	}
	t, c := g.eval(e.Path, env)
	if c != ctlNone {
		return t, c
	}
	switch tag := t.(type) {
	case *value.Tag:
		if tag.Kind == value.TimeTag {
			tv, ok := v.(value.Time)
			return ok && tv.TimeType() == tag.Name, ctlNone
		}
		if tag.Kind != value.RecordTag {
			g.panicf(e.Pos, "is needs a record type, a named primitive type or a time type; %s is not one", tag.Name)
		}
		r, ok := v.(*value.Record)
		return ok && r.Tag == tag, ctlNone
	case *value.Constructor:
		tv, ok := v.(*value.Tagged)
		return ok && tv.Tag == tag.Tag, ctlNone
	}
	g.panicf(e.Pos, "the right side of is must be a type, got %s", value.Describe(t))
	return nil, ctlNone
}

// ---- assignment

func (g *Goroutine) assign(e *ast.Assign, env *Env) (value.Value, ctl) {
	op := e.Op[:len(e.Op)-1] // "+=" -> "+"
	switch t := e.Target.(type) {
	case *ast.Ident:
		var cur value.Value
		if e.Op != "=" {
			cur = g.ident(t, env)
		}
		v, c := g.eval(e.Value, env)
		if c != ctlNone {
			return v, c
		}
		if e.Op != "=" {
			v = g.arith(op, cur, v, e.Pos)
		}
		g.assignIdent(t, v, env)
	case *ast.Field:
		recv, c := g.eval(t.X, env)
		if c != ctlNone {
			return recv, c
		}
		var cur value.Value
		if e.Op != "=" {
			cur = g.field(recv, t.Name, t.Pos)
		}
		v, c := g.eval(e.Value, env)
		if c != ctlNone {
			return v, c
		}
		if e.Op != "=" {
			v = g.arith(op, cur, v, e.Pos)
		}
		g.setField(recv, t.Name, v, t.Pos)
	case *ast.Index:
		recv, c := g.eval(t.X, env)
		if c != ctlNone {
			return recv, c
		}
		idx, c := g.eval(t.Index, env)
		if c != ctlNone {
			return idx, c
		}
		var cur value.Value
		if e.Op != "=" {
			cur = g.index(recv, idx, t.Pos)
		}
		v, c := g.eval(e.Value, env)
		if c != ctlNone {
			return v, c
		}
		if e.Op != "=" {
			v = g.arith(op, cur, v, e.Pos)
		}
		g.setIndex(recv, idx, v, t.Pos)
	case *ast.AssignArray:
		places, c := g.places(t, env)
		if c != ctlNone {
			return nil, c
		}
		v, c := g.eval(e.Value, env)
		if c != ctlNone {
			return v, c
		}
		g.storePlaces(places, v)
	}
	return value.Void, ctlNone
}

func (g *Goroutine) setField(recv value.Value, name string, v value.Value, pos ast.Pos) {
	rec, ok := recv.(*value.Record)
	if !ok {
		g.panicf(pos, "%s is not a record, so it has no field %s", value.Describe(recv), name)
	}
	if !rec.SetField(name, v) {
		g.panicf(pos, "%s has no field %s", rec.Tag.Name, name)
	}
}

func (g *Goroutine) setIndex(recv, idx, v value.Value, pos ast.Pos) {
	switch x := recv.(type) {
	case *value.Array:
		i, ok := idx.(int64)
		if !ok {
			g.panicf(pos, "an array index must be an int, got %s", value.Describe(idx))
		}
		if !x.Set(int(i), v) {
			g.panicf(pos, "index %d out of range for an array of length %d", i, x.Len())
		}
	case *value.Map:
		x.Set(idx, v)
	default:
		g.panicf(pos, "%s is not an array or a map, so it cannot be indexed", value.Describe(recv))
	}
}

// place is an evaluated assignment target of a destructuring assignment.
type place struct {
	store  func(value.Value)
	nested []place
	rest   func(value.Value)
	pos    ast.Pos
}

// places evaluates the sub-expressions of every position, left to right, before the right
// side is evaluated.
func (g *Goroutine) places(t *ast.AssignArray, env *Env) (place, ctl) {
	p := place{pos: t.Pos}
	for _, el := range t.Elems {
		switch x := el.(type) {
		case *ast.AssignArray:
			np, c := g.places(x, env)
			if c != ctlNone {
				return p, c
			}
			p.nested = append(p.nested, np)
		default:
			st, c := g.placeOf(x.(ast.Expr), env)
			if c != ctlNone {
				return p, c
			}
			p.nested = append(p.nested, place{store: st})
		}
	}
	if t.Rest != nil {
		st, c := g.placeOf(t.Rest, env)
		if c != ctlNone {
			return p, c
		}
		p.rest = st
	}
	return p, ctlNone
}

func (g *Goroutine) placeOf(e ast.Expr, env *Env) (func(value.Value), ctl) {
	switch x := e.(type) {
	case *ast.Ident:
		return func(v value.Value) { g.assignIdent(x, v, env) }, ctlNone
	case *ast.Field:
		recv, c := g.eval(x.X, env)
		if c != ctlNone {
			return nil, c
		}
		return func(v value.Value) { g.setField(recv, x.Name, v, x.Pos) }, ctlNone
	case *ast.Index:
		recv, c := g.eval(x.X, env)
		if c != ctlNone {
			return nil, c
		}
		idx, c := g.eval(x.Index, env)
		if c != ctlNone {
			return nil, c
		}
		return func(v value.Value) { g.setIndex(recv, idx, v, x.Pos) }, ctlNone
	}
	g.panicf(e.Position(), "internal error: not a place")
	return nil, ctlNone
}

func (g *Goroutine) storePlaces(p place, v value.Value) {
	arr, ok := v.(*value.Array)
	if !ok {
		g.panicf(p.pos, "a destructuring assignment needs an array, got %s", value.Describe(v))
	}
	vals := arr.Values()
	if len(vals) < len(p.nested) {
		g.panicf(p.pos, "a destructuring needs %d elements, got %d", len(p.nested), len(vals))
	}
	for i, np := range p.nested {
		if np.store != nil {
			np.store(vals[i])
		} else {
			g.storePlaces(np, vals[i])
		}
	}
	if p.rest != nil {
		p.rest(value.NewArray(vals[len(p.nested):]))
	}
}

// bind binds a var or for target.
func (g *Goroutine) bind(t ast.Target, v value.Value, env *Env) {
	switch t := t.(type) {
	case *ast.NameTarget:
		if t.Name == "_" {
			return
		}
		if t.Slot.Module {
			env.mod.rec.Set(t.Slot.Index, v)
			return
		}
		env.set(t.Slot, v)
	case *ast.RecordTarget:
		rec, ok := v.(*value.Record)
		if !ok {
			g.panicf(t.Pos, "a record destructuring needs a record, got %s", value.Describe(v))
		}
		for _, f := range t.Fields {
			fv, ok := rec.Field(f.Name)
			if !ok {
				g.panicf(f.Pos, "%s has no field %s", rec.Tag.Name, f.Name)
			}
			g.bind(f.Target, fv, env)
		}
	case *ast.ArrayTarget:
		arr, ok := v.(*value.Array)
		if !ok {
			g.panicf(t.Pos, "an array destructuring needs an array, got %s", value.Describe(v))
		}
		vals := arr.Values()
		if len(vals) < len(t.Elems) {
			g.panicf(t.Pos, "a destructuring needs %d elements, got %d", len(t.Elems), len(vals))
		}
		for i, el := range t.Elems {
			g.bind(el, vals[i], env)
		}
		if t.Rest != nil {
			g.bind(t.Rest, value.NewArray(vals[len(t.Elems):]), env)
		}
	}
}

// ---- loops

func (g *Goroutine) forLoop(e *ast.For, env *Env) (value.Value, ctl) {
	iter, c := g.eval(e.Iter, env)
	if c != ctlNone {
		return iter, c
	}
	// run runs the body once; stop reports a break or a return.
	run := func(v value.Value) (value.Value, ctl, bool) {
		ienv := env
		if e.Scope != nil {
			ienv = newEnv(e.Scope, env)
		}
		g.bind(e.Target, v, ienv)
		rv, rc := g.evalBlock(e.Body, ienv)
		switch rc {
		case ctlBreak:
			return nil, ctlNone, true
		case ctlReturn:
			return rv, ctlReturn, true
		}
		return nil, ctlNone, false
	}
	switch x := iter.(type) {
	case *value.Array:
		for i := 0; i < x.Len(); i++ {
			g.check()
			v, ok := x.Get(i)
			if !ok {
				break
			}
			if rv, rc, stop := run(v); stop {
				return rv, rc
			}
		}
	case *value.Map:
		var cursor uint64
		for {
			g.check()
			k, v, seq, ok := x.Next(cursor)
			if !ok {
				break
			}
			cursor = seq
			if rv, rc, stop := run(value.Entry(k, v)); stop {
				return rv, rc
			}
		}
	case string:
		for _, r := range x {
			g.check()
			if rv, rc, stop := run(string(r)); stop {
				return rv, rc
			}
		}
	case value.Callable:
		return g.forSeq(e, x, run)
	default:
		g.panicf(e.Iter.Position(), "for needs an array, a map, a string, or a function, got %s", value.Describe(iter))
	}
	return value.Void, ctlNone
}

func (g *Goroutine) forSeq(e *ast.For, seq value.Value, run func(value.Value) (value.Value, ctl, bool)) (rv value.Value, rc ctl) {
	active := true
	yield := &Builtin{Name: "yield", Params: []string{"x"}, Fn: func(g2 *Goroutine, args []value.Value) value.Value {
		if !active {
			g2.Panic("yield called after the sequence ended")
		}
		v, c, stop := run(args[0])
		if !stop {
			return true
		}
		active = false
		if c == ctlReturn {
			panic(&loopUnwind{loop: e, v: v})
		}
		return false
	}}
	mark := len(g.args)
	defer func() {
		if r := recover(); r != nil {
			if u, ok := r.(*loopUnwind); ok && u.loop == e {
				g.unwindArgs(mark)
				rv, rc = u.v, ctlReturn
				return
			}
			panic(r)
		}
	}()
	g.callValue(seq, []value.Value{yield}, nil, e.Pos)
	active = false
	return value.Void, ctlNone
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
