package interp

import (
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/value"
)

// Env is an environment: the slots of one block, loop iteration, or function body, with its
// parent. Plain slots belong to one goroutine; shared slots may be captured by a nested body
// and are published through value.Slot.
type Env struct {
	plain  []value.Value
	shared []value.Slot
	parent *Env
	mod    *module
}

func newEnv(scope *ast.BlockScope, parent *Env) *Env {
	e := &Env{parent: parent}
	if parent != nil {
		e.mod = parent.mod
	}
	if scope != nil {
		if scope.Plain > 0 {
			e.plain = make([]value.Value, scope.Plain)
			for i := range e.plain {
				e.plain[i] = value.Uninitialized
			}
		}
		if scope.Shared > 0 {
			e.shared = make([]value.Slot, scope.Shared)
		}
	}
	return e
}

func (e *Env) up(depth int) *Env {
	for ; depth > 0; depth-- {
		e = e.parent
	}
	return e
}

func (e *Env) get(slot ast.Slot) value.Value {
	if slot.Shared {
		return e.shared[slot.Index].Load()
	}
	return e.plain[slot.Index]
}

func (e *Env) set(slot ast.Slot, v value.Value) {
	if slot.Shared {
		e.shared[slot.Index].Store(v)
		return
	}
	e.plain[slot.Index] = v
}

// ---- functions

// Func is a declared or anonymous function, or a lambda, with the environment it captured.
type Func struct {
	Name   string
	Params []ast.Param
	Body   *ast.Block
	Scope  *ast.FuncScope
	Env    *Env
	Lambda bool
}

// FuncName returns the function's name, "func" for anonymous functions and lambdas.
func (f *Func) FuncName() string {
	if f.Name == "" {
		return "func"
	}
	return f.Name
}

// Builtin is a function implemented in Go. It receives its arguments positionally, exactly
// len(Params) of them.
type Builtin struct {
	Name   string
	Params []string
	Fn     func(g *Goroutine, args []value.Value) value.Value
}

// FuncName returns the builtin's name.
func (b *Builtin) FuncName() string { return b.Name }

// deferred is a registered deferred block.
type deferred struct {
	body  *ast.Block
	scope *ast.FuncScope
	env   *Env
}

type deferFrame struct {
	list []deferred
}
