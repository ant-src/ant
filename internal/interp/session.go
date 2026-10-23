package interp

import (
	"context"

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
// goroutines stop at their next check, as a run's do, since no wait watches it for them.
func NewSession(ctx context.Context, cfg Config) *Session {
	in := New(ctx, &loader.Program{}, cfg)
	go func() {
		<-in.ctx.Done()
		in.done.Store(true)
	}()
	return &Session{in: in, known: map[string]*module{}, Fields: &ast.ModuleScope{Index: map[string]int{}}}
}

// Instance returns the session's instance, for Close.
func (s *Session) Instance() *Instance { return s.in }

// Lookup returns a name the session's module has bound.
func (s *Session) Lookup(name string) (value.Value, bool) {
	if s.mod == nil {
		return nil, false
	}
	v, ok := s.mod.rec.Field(name)
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
	// the session module: grow its record to the fields known now, keeping earlier values
	if s.mod == nil {
		s.mod = newModule(root)
	} else {
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
	g := s.in.newGoroutine(true)
	var result value.Value = value.Void
	err := s.in.runOn(g, func() {
		for _, m := range fresh {
			s.in.initModule(g, m)
		}
		env := s.mod.root
		for _, st := range root.File.Stmts {
			switch st := st.(type) {
			case *ast.VarDecl:
				v := g.evalValue(st.Init, env)
				g.bind(st.Target, v, env)
				result = value.Void
			case ast.Expr:
				result = g.evalValue(st, env)
			}
		}
	})
	s.in.Flush()
	return result, err
}
