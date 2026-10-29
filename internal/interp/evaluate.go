package interp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/resolve"
	"github.com/ant-src/ant/parser"
	"github.com/ant-src/ant/value"
)

// An expression a debugger evaluates (dap.md section 6) runs as the only statement of a module
// of its own, whose fields are the names the frame sees, holding their values at the stop.

// evalPath is the path of the module an evaluation runs in.
const evalPath = "(evaluate)"

// evalTimeout bounds an evaluation the client asked for, since the other goroutines stay stopped.
const evalTimeout = time.Second

// compiled is an expression resolved against the names a place sees, in visible's order.
type compiled struct {
	file  *ast.File
	expr  ast.Expr
	names []string
}

// visible lists the names env sees, innermost first, with their values: the bindings of its
// environments, then the module's declarations, then its imports.
func visible(env *Env) (names []string, vals []value.Value) {
	seen := map[string]bool{}
	add := func(n string, v value.Value) {
		if n == "" || n == "_" || seen[n] {
			return
		}
		seen[n] = true
		names = append(names, n)
		vals = append(vals, v)
	}
	for e := env; e != nil; e = e.parent {
		if e.scope == nil {
			continue
		}
		for _, n := range e.scope.Names {
			add(n.Name, e.get(ast.Slot{Index: n.Index, Shared: n.Shared}))
		}
	}
	m := env.mod
	if m == nil {
		return names, vals
	}
	for i, f := range m.rec.Tag.Fields {
		add(f, m.get(i))
	}
	for i, imp := range m.m.File.Imports {
		if i >= len(m.imports) {
			break
		}
		ir := m.imports[i]
		if ir.namespace {
			add(imp.Namespace, ir.rec)
			continue
		}
		for _, n := range imp.Names {
			if v, ok := ir.rec.Field(n.Name); ok {
				add(n.Alias, v)
			}
		}
	}
	return names, vals
}

// compileExpr parses and resolves src against names.
func compileExpr(src string, names []string) (*compiled, error) {
	e, err := parser.ParseExpr(src)
	if err != nil {
		return nil, err
	}
	f := &ast.File{Path: evalPath, Script: true, Stmts: []ast.Stmt{e}}
	scope := &ast.ModuleScope{Fields: append([]string(nil), names...), Index: map[string]int{}}
	for i, n := range names {
		scope.Index[n] = i
	}
	if errs := resolve.Resolve(f, resolve.Options{Session: true, Fields: scope}); len(errs) > 0 {
		return nil, errors.New(errs[0].Msg)
	}
	var bad error
	ast.Inspect(e, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if x.Ref.Kind == ast.RefUnbound && bad == nil {
				bad = fmt.Errorf("%s is not defined here", x.Name)
			}
		case *ast.Assign:
			if id, ok := x.Target.(*ast.Ident); ok && id.Ref.Kind == ast.RefModule && bad == nil {
				bad = fmt.Errorf("%s is read as its value at the stop and cannot be assigned", id.Name)
			}
		}
		return bad == nil
	})
	if bad != nil {
		return nil, bad
	}
	return &compiled{file: f, expr: e, names: names}, nil
}

// run evaluates c with the values of its names, on a goroutine of the run that no debugger
// stops. A panic in it is its error.
func (in *Instance) run(c *compiled, vals []value.Value) (v value.Value, err error) {
	rec := value.NewRecord(value.NewTag(evalPath, value.RecordTag, c.names), vals)
	m := &module{m: &loader.Module{Path: evalPath, File: c.file}, path: evalPath, rec: rec, tags: map[string]*value.Tag{}}
	m.root = &Env{mod: m}
	g := &Goroutine{in: in}
	defer func() {
		r := recover()
		switch x := r.(type) {
		case nil:
		case *AntPanic:
			err = errors.New("panic: " + x.Message)
		case *cancelSignal:
			err = errors.New("the run has ended")
		case *exitSignal:
			err = fmt.Errorf("the expression called Os.exit(%d), which does not end the run from here", x.status)
		default:
			err = fmt.Errorf("%v", r)
		}
	}()
	return g.evalValue(c.expr, m.root), nil
}

// Evaluate evaluates src in a frame of the stopped run (dap.md section 6, Expressions).
func (d *Debugger) Evaluate(sf StackFrame, src string) (value.Value, error) {
	d.mu.Lock()
	if !d.stopping || sf.g == nil || sf.i >= len(sf.g.stack) || sf.g.stack[sf.i].env == nil {
		d.mu.Unlock()
		return nil, errors.New("the program is not stopped in this frame")
	}
	names, vals := visible(sf.g.stack[sf.i].env)
	d.mu.Unlock()
	c, err := compileExpr(src, names)
	if err != nil {
		return nil, err
	}
	in := d.in.Load()
	type result struct {
		v   value.Value
		err error
	}
	done := make(chan result, 1)
	go func() {
		v, err := in.run(c, vals)
		done <- result{v, err}
	}()
	select {
	case r := <-done:
		return r.v, r.err
	case <-time.After(evalTimeout):
		return nil, errors.New("the expression did not finish within a second; it may wait for a goroutine, which stays stopped")
	}
}

// ---- breakpoint expressions (dap.md section 6, Breakpoint expressions)

// Breakpoint is a line breakpoint bound to a stop point, with what decides whether it stops.
type Breakpoint struct {
	Pos          ast.Pos
	Condition    string
	HitCondition string
	LogMessage   string
}

// bp is a breakpoint as the run keeps it.
type bp struct {
	Breakpoint
	hitOp string // "==", ">=", ">", or "%"
	hitN  int
	log   []logPart
	// mu keeps one goroutine at a time evaluating and counting the breakpoint
	mu    sync.Mutex
	hits  int
	cache map[string]*compiled // the condition and the log's expressions, by names and source
}

// logPart is a piece of a log message: text, or an expression.
type logPart struct {
	text string
	expr string
}

// newBP checks a breakpoint's expressions, which must parse.
func newBP(b Breakpoint) (*bp, error) {
	p := &bp{Breakpoint: b, cache: map[string]*compiled{}}
	if b.Condition != "" {
		if _, err := parser.ParseExpr(b.Condition); err != nil {
			return nil, fmt.Errorf("the condition: %v", err)
		}
	}
	if h := strings.TrimSpace(b.HitCondition); h != "" {
		op := "=="
		for _, o := range []string{"==", ">=", ">", "%"} {
			if strings.HasPrefix(h, o) {
				op, h = o, strings.TrimSpace(h[len(o):])
				break
			}
		}
		n, err := strconv.Atoi(h)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("the hit condition %q is not N, == N, >= N, > N, or %% N with N above 0", b.HitCondition)
		}
		p.hitOp, p.hitN = op, n
	}
	if b.LogMessage != "" {
		parts, err := parseLog(b.LogMessage)
		if err != nil {
			return nil, err
		}
		p.log = parts
	}
	return p, nil
}

// parseLog splits a log message into text and {expr} parts; {{ and }} are braces.
func parseLog(msg string) ([]logPart, error) {
	var parts []logPart
	var text strings.Builder
	for i := 0; i < len(msg); i++ {
		switch c := msg[i]; {
		case c == '{' && i+1 < len(msg) && msg[i+1] == '{', c == '}' && i+1 < len(msg) && msg[i+1] == '}':
			text.WriteByte(c)
			i++
		case c == '{':
			depth, j := 1, i+1
			for ; j < len(msg) && depth > 0; j++ {
				switch msg[j] {
				case '{':
					depth++
				case '}':
					depth--
				}
			}
			if depth > 0 {
				return nil, errors.New("the log message has a { without its }")
			}
			expr := msg[i+1 : j-1]
			if _, err := parser.ParseExpr(expr); err != nil {
				return nil, fmt.Errorf("the log message's {%s}: %v", expr, err)
			}
			if text.Len() > 0 {
				parts = append(parts, logPart{text: text.String()})
				text.Reset()
			}
			parts = append(parts, logPart{expr: expr})
			i = j - 1
		default:
			text.WriteByte(c)
		}
	}
	if text.Len() > 0 {
		parts = append(parts, logPart{text: text.String()})
	}
	return parts, nil
}

// eval evaluates one of the breakpoint's expressions in env, on the goroutine that reached it.
func (p *bp) eval(in *Instance, src string, env *Env) (value.Value, error) {
	names, vals := visible(env)
	key := strings.Join(names, "\x00") + "\x01" + src
	c := p.cache[key]
	if c == nil {
		var err error
		if c, err = compileExpr(src, names); err != nil {
			return nil, err
		}
		p.cache[key] = c
	}
	return in.run(c, vals)
}

// hit decides what a breakpoint does when g reaches it: whether it stops, and what it logs.
func (d *Debugger) hit(g *Goroutine, p *bp, env *Env) (stop bool, log string) {
	in := g.in
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Condition != "" {
		v, err := p.eval(in, p.Condition, env)
		if err != nil {
			return true, fmt.Sprintf("the breakpoint's condition at %s:%d: %v", env.mod.path, p.Pos.Line, err)
		}
		b, ok := v.(bool)
		if !ok {
			return true, fmt.Sprintf("the breakpoint's condition at %s:%d is %s, not a bool", env.mod.path, p.Pos.Line, value.Describe(v))
		}
		if !b {
			return false, ""
		}
	}
	if p.hitOp != "" {
		p.hits++
		n := p.hits
		switch p.hitOp {
		case "==":
			stop = n == p.hitN
		case ">=":
			stop = n >= p.hitN
		case ">":
			stop = n > p.hitN
		case "%":
			stop = n%p.hitN == 0
		}
		if !stop {
			return false, ""
		}
	}
	if p.log == nil {
		return true, ""
	}
	var sb strings.Builder
	for _, part := range p.log {
		if part.expr == "" {
			sb.WriteString(part.text)
			continue
		}
		v, err := p.eval(in, part.expr, env)
		if err != nil {
			sb.WriteString("<" + err.Error() + ">")
			continue
		}
		sb.WriteString(value.Render(v))
	}
	return false, sb.String() + "\n"
}
