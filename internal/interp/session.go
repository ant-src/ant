package interp

import (
	"context"
	"sync"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/value"
)

// Session evaluates inputs one at a time against a persistent module: declarations
// accumulate, a name may be declared again, and imported modules are loaded once.
type Session struct {
	in     *Instance
	mod    *module
	known  map[string]*module
	Fields *ast.ModuleScope
}

// NewSession creates a session whose module starts empty. When ctx is done the session's
// goroutines stop at their next check, as a run's do.
func NewSession(ctx context.Context, cfg Config) *Session {
	in := New(ctx, &loader.Program{}, cfg)
	return &Session{in: in, known: map[string]*module{}, Fields: &ast.ModuleScope{Index: map[string]int{}}}
}

// Instance returns the session's instance, for Close.
func (s *Session) Instance() *Instance { return s.in }

// Lookup returns a name the session's module has bound.
func (s *Session) Lookup(name string) (value.Value, bool) {
	if s.mod == nil {
		return nil, false
	}
	s.mod.grow.RLock()
	v, ok := s.mod.rec.Field(name)
	s.mod.grow.RUnlock()
	if !ok || v == value.Uninitialized {
		return nil, false
	}
	return v, true
}

// Eval takes the program of one input, which the loader resolved against the session's
// fields, initializes any module it imports for the first time, binds the input's new
// declarations, and runs its statements. It returns the value of the last expression.
func (s *Session) Eval(prog *loader.Program) (value.Value, error) {
	root := prog.Root
	lookup := func(i int) *module {
		lm := prog.Modules[i]
		if lm == root {
			return s.mod
		}
		return s.known[lm.Path]
	}
	var fresh []*module
	for _, lm := range prog.Modules {
		if lm == root {
			continue
		}
		if _, ok := s.known[lm.Path]; ok {
			continue
		}
		m := newModule(lm)
		s.known[lm.Path] = m
		fresh = append(fresh, m)
	}
	// the session module: grow its record to the fields known now, keeping earlier values,
	// and relink it, under its lock, since goroutines of earlier inputs may be using it
	if s.mod == nil {
		s.mod = newModule(root)
		s.mod.grow = &sync.RWMutex{}
	}
	s.mod.grow.Lock()
	if s.mod.m != root {
		s.mod.m = root
		if len(s.Fields.Fields) > s.mod.rec.Len() {
			tag := value.NewTag(root.File.Path, value.RecordTag, append([]string(nil), s.Fields.Fields...))
			rec := value.NewRecord(tag, nil)
			for i := 0; i < s.mod.rec.Len(); i++ {
				if v := s.mod.rec.Get(i); v != value.Uninitialized {
					rec.Set(i, v)
				}
			}
			s.mod.rec = rec
		}
	}
	for _, m := range fresh {
		s.in.linkModule(m, lookup)
	}
	s.in.linkModule(s.mod, lookup)
	s.mod.grow.Unlock()
	g := s.in.newGoroutine(true)
	var result value.Value = value.Void
	err := s.in.runOn(func() {
		for _, m := range fresh {
			s.in.initModule(g, m)
		}
		var c ctl
		result, c = g.evalStmts(root.File.Stmts, s.mod.root)
		if c != ctlNone {
			// The parser keeps return, break, continue and try inside a function or a loop.
			g.panicf(ast.Pos{}, "internal error: control left a session input")
		}
	})
	s.in.Flush()
	return result, err
}
