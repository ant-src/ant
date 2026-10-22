package ant

import (
	"context"
	"fmt"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/lib"
	"github.com/ant-src/ant/parser"
	"github.com/ant-src/ant/value"
)

// Runtime is what a host function sees of the running program.
type Runtime interface {
	// Call calls an Ant function value with positional arguments. An Ant panic raised inside
	// unwinds through the host function as a Go panic, so use defer for any cleanup.
	Call(fn value.Value, args ...value.Value) value.Value
	// Context is the run's context; blocking host functions should observe it.
	Context() context.Context
}

// A HostFunc is a Go function Ant code can call. It receives its arguments positionally, in
// the order its signature lists them. An error it returns becomes an Ant panic with the
// error's message; a nil result is void.
type HostFunc func(rt Runtime, args []value.Value) (value.Value, error)

// Namespace is a set of host functions and types that Ant code reaches as Name.function, the
// way it reaches the standard library.
type Namespace struct {
	// Name is the namespace's name, bound in every module like a prelude name.
	Name string
	// Signatures declares the functions and types in Ant notation, one per line, as the
	// prelude declares the standard library, for example:
	//
	//	func Db.query(sql: string, args: Array[string]): Result[Array[Db.Row], Error]
	//	type Row = Map[string, string]
	//
	// The checker types calls from them; the runtime takes parameter names from them.
	Signatures string
	// Funcs maps each function's name, without the namespace, to its implementation.
	Funcs map[string]HostFunc
}

// WithNamespace adds a host namespace to a program.
func WithNamespace(ns Namespace) Option {
	return func(c *config) { c.namespaces = append(c.namespaces, ns) }
}

// WithoutNamespaces removes prelude namespaces from a program, such as File or Http, so a
// script cannot reach them.
func WithoutNamespaces(names ...string) Option {
	return func(c *config) { c.removed = append(c.removed, names...) }
}

// build parses a namespace's signatures and binds its functions.
func (ns Namespace) build() (value.Value, *ast.File, error) {
	f, err := parser.Parse("namespace "+ns.Name, ns.Signatures, parser.Declarations)
	if err != nil {
		return nil, nil, fmt.Errorf("namespace %s: %v", ns.Name, err)
	}
	var fields []string
	var vals []value.Value
	seen := map[string]bool{}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Owner != ns.Name {
				return nil, nil, fmt.Errorf("namespace %s: func %s must be declared as %s.%s", ns.Name, d.Name, ns.Name, d.Name)
			}
			impl, ok := ns.Funcs[d.Name]
			if !ok {
				return nil, nil, fmt.Errorf("namespace %s: no implementation for %s", ns.Name, d.Name)
			}
			params := make([]string, len(d.Params))
			for i, p := range d.Params {
				params[i] = p.Name
			}
			optional := 0
			for i := len(d.Params) - 1; i >= 0 && d.Params[i].Default != nil; i-- {
				optional++
			}
			fields = append(fields, d.Name)
			vals = append(vals, &interp.Builtin{Name: ns.Name + "." + d.Name, Params: params, Optional: optional, Fn: hostBuiltin(impl)})
			seen[d.Name] = true
		case *ast.TypeDecl:
			fields = append(fields, d.Name)
			vals = append(vals, lib.TypeValue(d))
		default:
			return nil, nil, fmt.Errorf("namespace %s: only func and type declarations are allowed", ns.Name)
		}
	}
	for name := range ns.Funcs {
		if !seen[name] {
			return nil, nil, fmt.Errorf("namespace %s: %s has no signature", ns.Name, name)
		}
	}
	tag := value.NewTag(ns.Name, value.RecordTag, fields)
	return value.NewRecord(tag, vals), f, nil
}

func hostBuiltin(impl HostFunc) func(*interp.Goroutine, []value.Value) value.Value {
	return func(g *interp.Goroutine, args []value.Value) value.Value {
		v, err := impl(g, args)
		if err != nil {
			g.Panic("%s", strings.TrimSpace(err.Error()))
		}
		if v == nil {
			return value.Void
		}
		return v
	}
}

// prelude builds the table of prelude names for a program: the standard library, less the
// removed namespaces, plus the host namespaces.
func (c config) prelude() (map[string]value.Value, []string, []check.HostNamespace, error) {
	table := make(map[string]value.Value, len(lib.Prelude())+len(c.namespaces))
	for k, v := range lib.Prelude() {
		table[k] = v
	}
	for _, name := range c.removed {
		delete(table, name)
	}
	var extra []string
	var decls []check.HostNamespace
	for _, ns := range c.namespaces {
		v, f, err := ns.build()
		if err != nil {
			return nil, nil, nil, err
		}
		table[ns.Name] = v
		extra = append(extra, ns.Name)
		decls = append(decls, check.HostNamespace{Name: ns.Name, Decls: f})
	}
	return table, extra, decls, nil
}

// An Instance is a started program: its modules are initialized and its functions can be
// called from Go until Close.
type Instance struct {
	in *interp.Instance
}

// Start initializes the program's modules without calling main. The options given to Load
// apply, and options given here override them.
func (p *Program) Start(ctx context.Context, opts ...Option) (*Instance, error) {
	cfg := p.cfg.with(opts)
	in, err := interp.Start(ctx, p.prog, interp.Config{
		Prelude: p.prelude, Stdout: cfg.stdout, Stderr: cfg.stderr, Args: cfg.args, Env: cfg.env,
	})
	if err != nil {
		return nil, convertRunError(err)
	}
	return &Instance{in: in}, nil
}

// Func returns a top-level function of the root module, or false when there is none.
func (in *Instance) Func(name string) (*Func, bool) {
	v, ok := in.in.Lookup(name)
	if !ok {
		return nil, false
	}
	if _, ok := v.(value.Callable); !ok {
		return nil, false
	}
	return &Func{in: in.in, v: v, name: name}, true
}

// Value returns a top-level value of the root module, such as a var or a type's tag.
func (in *Instance) Value(name string) (value.Value, bool) { return in.in.Lookup(name) }

// Close stops the instance's goroutines and flushes its output.
func (in *Instance) Close() { in.in.Close() }

// A Func is an Ant function callable from Go.
type Func struct {
	in   *interp.Instance
	v    value.Value
	name string
}

// Call calls the function with positional arguments. An Ant panic comes back as a
// *PanicError, after being reported to the instance's stderr; Os.exit as an *ExitError.
func (f *Func) Call(args ...value.Value) (value.Value, error) {
	v, err := f.in.Call(f.v, args...)
	if err != nil {
		return nil, convertRunError(err)
	}
	return v, nil
}

// Value returns the function as an Ant value, for passing to other calls.
func (f *Func) Value() value.Value { return f.v }

func convertRunError(err error) error {
	switch e := err.(type) {
	case nil:
		return nil
	case *interp.PanicError:
		trace := make([]Frame, len(e.Trace))
		for i, f := range e.Trace {
			trace[i] = Frame{Function: f.Function, Pos: Position{Path: f.Path, Line: f.Pos.Line, Col: f.Pos.Col}}
		}
		return &PanicError{Message: e.Message, Trace: trace}
	case *interp.ExitError:
		return &ExitError{Status: e.Status}
	}
	return err
}
