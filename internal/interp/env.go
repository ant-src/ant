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

// Environments with up to four plain slots are allocated together with their slots, sized
// exactly, so that opening one is a single allocation of no more than it needs.
type (
	env1 struct {
		Env
		slots [1]value.Value
	}
	env2 struct {
		Env
		slots [2]value.Value
	}
	env3 struct {
		Env
		slots [3]value.Value
	}
	env4 struct {
		Env
		slots [4]value.Value
	}
)

func allocEnv(plain int) *Env {
	switch plain {
	case 0:
		return &Env{}
	case 1:
		x := &env1{}
		x.plain = x.slots[:]
		return &x.Env
	case 2:
		x := &env2{}
		x.plain = x.slots[:]
		return &x.Env
	case 3:
		x := &env3{}
		x.plain = x.slots[:]
		return &x.Env
	case 4:
		x := &env4{}
		x.plain = x.slots[:]
		return &x.Env
	}
	return &Env{plain: make([]value.Value, plain)}
}

func newEnv(scope *ast.BlockScope, parent *Env) *Env {
	var e *Env
	if scope == nil {
		e = &Env{}
	} else {
		e = allocEnv(scope.Plain)
		for i := range e.plain {
			e.plain[i] = value.Uninitialized
		}
		if scope.Shared > 0 {
			e.shared = make([]value.Slot, scope.Shared)
		}
	}
	e.parent = parent
	if parent != nil {
		e.mod = parent.mod
	}
	return e
}

// enter opens a block or loop iteration's scope in env: a new environment, or, for an inline
// scope, its range of env's slots reset to unbound.
func enter(scope *ast.BlockScope, env *Env) *Env {
	if !scope.Inline {
		return newEnv(scope, env)
	}
	slots := env.plain[scope.Base : scope.Base+scope.Plain]
	for i := range slots {
		slots[i] = value.Uninitialized
	}
	return env
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
// len(Params) of them. A call may omit the last Optional parameters, and an omitted one arrives
// as nil, for the builtin to replace by its default. The slice may live on the goroutine's
// argument stack, so Fn must not keep it, or a subslice of it, after it returns.
type Builtin struct {
	Name     string
	Params   []string
	Optional int
	Fn       func(g *Goroutine, args []value.Value) value.Value
	// Each is set on a sequence the library implements in Go. It runs the sequence, pushing
	// each element until push reports false. A for loop or a library stage drives it through
	// Each rather than calling Fn with a yield function, which no Ant code could observe.
	Each func(g *Goroutine, push func(value.Value) bool)
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
