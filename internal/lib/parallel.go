package lib

import (
	"sync"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// Parallel of section 11: structured fan-out. Each call makes a scope, a context that is a child
// of the caller's, starts every callback on a goroutine of its own, joins them all, and raises
// the first panic that left a callback in the caller.

// scope is one Parallel call in progress.
type scope struct {
	g   *interp.Goroutine
	ctx *value.Context
	wg  sync.WaitGroup
	mu  sync.Mutex
	p   *interp.AntPanic // the first panic that left a callback
}

func newScope(g *interp.Goroutine, who string, ctx value.Value) *scope {
	return &scope{g: g, ctx: value.NewContext(waitContext(g, who, ctx), 0)}
}

// live reports whether the scope is still running: not cancelled by a panic, a winner, or the
// caller's context.
func (s *scope) live() bool { return s.ctx.Ctx.Err() == nil }

// acquire takes a slot of sem for the next admission, or reports false once the scope has ended.
// The scope is checked before and after the wait so that an ended scope admits nothing, whichever
// way the select chose.
func (s *scope) acquire(sem chan struct{}) bool {
	if !s.live() {
		return false
	}
	select {
	case sem <- struct{}{}:
	case <-s.ctx.Ctx.Done():
		return false
	}
	if !s.live() {
		<-sem
		return false
	}
	return true
}

// start runs call on a new goroutine that the scope joins. A panic that leaves call cancels the
// scope and is kept for the caller, or, when one is kept already, ends the goroutine and is
// reported there. done receives call's value when it returns; release runs once the goroutine is
// finished with its slot.
func (s *scope) start(call func(ng *interp.Goroutine) value.Value, done func(value.Value), release func()) {
	s.wg.Add(1)
	s.g.Spawn(func(ng *interp.Goroutine) {
		var dropped *interp.AntPanic
		func() {
			defer s.wg.Done()
			if release != nil {
				defer release()
			}
			v, p := ng.CatchFor(s.g, func() value.Value { return call(ng) })
			if p != nil {
				dropped = s.fail(p)
				return
			}
			if done != nil {
				done(v)
			}
		}()
		if dropped != nil {
			ng.Repanic(dropped)
		}
	})
}

// fail keeps p as the call's panic and cancels the scope, or returns p when a panic is kept
// already.
func (s *scope) fail(p *interp.AntPanic) *interp.AntPanic {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.p != nil {
		return p
	}
	s.p = p
	s.ctx.Cancel(value.ErrCancelled)
	return nil
}

// join waits for every callback, cancels the scope, and raises the kept panic in the caller.
func (s *scope) join() {
	s.wg.Wait()
	s.ctx.Cancel(value.ErrCancelled)
	s.g.AfterWait()
	if s.p != nil {
		s.g.Repanic(s.p)
	}
}

// limitArg reads the limit of map and each, which must be at least 1.
func limitArg(g *interp.Goroutine, who string, v value.Value) int {
	n := intArg(g, who, v)
	if n < 1 {
		g.Panic("%s needs a limit of at least 1, got %d", who, n)
	}
	if n > maxCount {
		n = maxCount
	}
	return int(n)
}

// funcsArg reads the array of callbacks of run, race, and any.
func funcsArg(g *interp.Goroutine, who string, v value.Value) []value.Value {
	fs := arrayArg(g, who, v).Values()
	for _, f := range fs {
		funcArg(g, who, f)
	}
	return fs
}

// overArray runs f on every element of xs, at most limit at once, and reports how many it
// admitted. each receives the index and the value of every callback that returns.
func (s *scope) overArray(xs []value.Value, f value.Value, limit int, each func(i int, v value.Value)) int {
	sem := make(chan struct{}, limit)
	admitted := 0
	for i, x := range xs {
		if !s.acquire(sem) {
			break
		}
		admitted++
		var done func(value.Value)
		if each != nil {
			done = func(v value.Value) { each(i, v) }
		}
		s.start(func(ng *interp.Goroutine) value.Value { return ng.Call(f, x, s.ctx) }, done, func() { <-sem })
	}
	return admitted
}

// overChannel runs f on every value received from c until it is closed, at most limit at once.
// It receives only with a slot free, and under the scope, so a value it takes always runs.
func (s *scope) overChannel(c *value.Chan, f value.Value, limit int) {
	sem := make(chan struct{}, limit)
	for s.acquire(sem) {
		x, ok, _ := c.Receive(s.ctx.Ctx)
		if !ok {
			<-sem
			return
		}
		s.start(func(ng *interp.Goroutine) value.Value { return ng.Call(f, x, s.ctx) }, nil, func() { <-sem })
	}
}

// all starts every callback of fs together, or none when the scope has ended already.
func (s *scope) all(fs []value.Value, call func(ng *interp.Goroutine, i int, f value.Value) value.Value, each func(i int, v value.Value)) {
	if !s.live() {
		return
	}
	for i, f := range fs {
		var done func(value.Value)
		if each != nil {
			done = func(v value.Value) { each(i, v) }
		}
		s.start(func(ng *interp.Goroutine) value.Value { return call(ng, i, f) }, done, nil)
	}
}

func init() {
	register("Parallel", func() value.Value {
		return namespace("Parallel",
			fn("Parallel.map", "xs, f, limit, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				xs := arrayArg(g, who, a[0]).Values()
				f := funcArg(g, who, a[1])
				limit := limitArg(g, who, a[2])
				s := newScope(g, who, a[3])
				results := make([]value.Value, len(xs))
				admitted := s.overArray(xs, f, limit, func(i int, v value.Value) { results[i] = v })
				s.join()
				if admitted < len(xs) {
					return value.Null
				}
				return value.NewArray(results)
			}),
			fn("Parallel.each", "xs, f, limit, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				f := funcArg(g, who, a[1])
				switch xs := a[0].(type) {
				case *value.Array:
					limit := limitArg(g, who, a[2])
					s := newScope(g, who, a[3])
					s.overArray(xs.Values(), f, limit, nil)
					s.join()
				case *value.Chan:
					limit := limitArg(g, who, a[2])
					s := newScope(g, who, a[3])
					s.overChannel(xs, f, limit)
					s.join()
				default:
					g.Panic("%s needs an array or a Channel, got %s", who, value.Describe(a[0]))
				}
				return value.Void
			}),
			fn("Parallel.run", "fs, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				fs := funcsArg(g, who, a[0])
				s := newScope(g, who, a[1])
				s.all(fs, func(ng *interp.Goroutine, _ int, f value.Value) value.Value { return ng.Call(f, s.ctx) }, nil)
				s.join()
				return value.Void
			}),
			fn("Parallel.race", "fs, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				fs := funcsArg(g, who, a[0])
				if len(fs) == 0 {
					g.Panic("%s needs at least one function", who)
				}
				s := newScope(g, who, a[1])
				var winner value.Value
				s.all(fs, func(ng *interp.Goroutine, _ int, f value.Value) value.Value { return ng.Call(f, s.ctx) },
					func(_ int, v value.Value) {
						s.mu.Lock()
						defer s.mu.Unlock()
						if winner == nil && s.live() {
							winner = v
							s.ctx.Cancel(value.ErrCancelled)
						}
					})
				s.join()
				if winner == nil {
					return value.Null
				}
				return winner
			}),
			fn("Parallel.any", "fs, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				fs := funcsArg(g, who, a[0])
				s := newScope(g, who, a[1])
				var winner value.Value
				errs := make([]value.Value, len(fs))
				failed := 0
				s.all(fs, func(ng *interp.Goroutine, _ int, f value.Value) value.Value {
					v := ng.Call(f, s.ctx)
					if r, ok := v.(*value.Record); !ok || r.Tag != value.OkTag && r.Tag != value.ErrTag {
						ng.Panic("the function given to %s must return a Result, got %s", who, value.Describe(v))
					}
					return v
				}, func(i int, v value.Value) {
					r := v.(*value.Record)
					s.mu.Lock()
					defer s.mu.Unlock()
					if winner != nil || !s.live() {
						return
					}
					if r.Tag == value.OkTag {
						winner = r
						s.ctx.Cancel(value.ErrCancelled)
						return
					}
					errs[i] = r.Get(0)
					failed++
				})
				s.join()
				switch {
				case winner != nil:
					return winner
				case failed == len(fs):
					return value.Err(value.NewArray(errs))
				}
				return value.Null
			}),
		)
	})
}
