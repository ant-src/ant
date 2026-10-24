package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// The sync types of section 9 as tags with functions (section 11). A blocking call flushes
// stdout first, so output is not held back while the program waits.

func mutexArg(g *interp.Goroutine, who string, v value.Value) *value.Mutex {
	return typedArg[*value.Mutex](g, who, "a Mutex", v)
}

func rwArg(g *interp.Goroutine, who string, v value.Value) *value.RwMutex {
	return typedArg[*value.RwMutex](g, who, "an RwMutex", v)
}

func wgArg(g *interp.Goroutine, who string, v value.Value) *value.WaitGroup {
	return typedArg[*value.WaitGroup](g, who, "a WaitGroup", v)
}

func onceArg(g *interp.Goroutine, who string, v value.Value) *value.Once {
	return typedArg[*value.Once](g, who, "a Once", v)
}

func condArg(g *interp.Goroutine, who string, v value.Value) *value.Cond {
	return typedArg[*value.Cond](g, who, "a Cond", v)
}

func atomicArg(g *interp.Goroutine, who string, v value.Value) *value.Atomic {
	return typedArg[*value.Atomic](g, who, "an Atomic", v)
}

func lock(g *interp.Goroutine, m *value.Mutex) {
	if m.TryLock() {
		return
	}
	g.Instance().Flush()
	m.Lock()
}

func unlock(g *interp.Goroutine, who string, m *value.Mutex) {
	if err := m.Unlock(); err != nil {
		g.Panic("%s: %s", who, err)
	}
}

// runOnce runs f under the once, sharing a panic in f with every caller. It reports whether f
// returned, which it did not when a return unwound through it.
func runOnce(g *interp.Goroutine, o *value.Once, f value.Value, keep func(value.Value)) bool {
	err := o.Do(func() error {
		var v value.Value
		if p := g.Catch(func() { v = g.Call(f) }); p != nil {
			return p
		}
		keep(v)
		return nil
	}, g.Instance().Flush)
	if p, ok := err.(*interp.AntPanic); ok {
		g.Repanic(p)
	}
	return err == nil
}

func syncTag(name string, funcs ...*interp.Builtin) *value.Tag {
	return withFuncs(value.NewTag(name, value.SyncTag, nil), funcs)
}

func init() {
	register("Mutex", func() value.Value {
		return syncTag("Mutex",
			fn("Mutex.new", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value { return value.NewMutex() }),
			fn("Mutex.lock", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				lock(g, mutexArg(g, who, a[0]))
				return value.Void
			}),
			fn("Mutex.unlock", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				unlock(g, who, mutexArg(g, who, a[0]))
				return value.Void
			}),
			fn("Mutex.withLock", "m, f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				m := mutexArg(g, who, a[0])
				f := funcArg(g, who, a[1])
				lock(g, m)
				defer m.Unlock()
				return g.Call(f)
			}),
		)
	})
	register("RwMutex", func() value.Value {
		return syncTag("RwMutex",
			fn("RwMutex.new", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value { return value.NewRwMutex() }),
			fn("RwMutex.readLock", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				g.Instance().Flush()
				rwArg(g, who, a[0]).ReadLock()
				return value.Void
			}),
			fn("RwMutex.readUnlock", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if err := rwArg(g, who, a[0]).ReadUnlock(); err != nil {
					g.Panic("RwMutex.readUnlock: %s", err)
				}
				return value.Void
			}),
			fn("RwMutex.writeLock", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				g.Instance().Flush()
				rwArg(g, who, a[0]).WriteLock()
				return value.Void
			}),
			fn("RwMutex.writeUnlock", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if err := rwArg(g, who, a[0]).WriteUnlock(); err != nil {
					g.Panic("RwMutex.writeUnlock: %s", err)
				}
				return value.Void
			}),
			fn("RwMutex.withRead", "m, f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				m := rwArg(g, who, a[0])
				f := funcArg(g, who, a[1])
				g.Instance().Flush()
				m.ReadLock()
				defer m.ReadUnlock()
				return g.Call(f)
			}),
			fn("RwMutex.withWrite", "m, f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				m := rwArg(g, who, a[0])
				f := funcArg(g, who, a[1])
				g.Instance().Flush()
				m.WriteLock()
				defer m.WriteUnlock()
				return g.Call(f)
			}),
		)
	})
	register("WaitGroup", func() value.Value {
		return syncTag("WaitGroup",
			fn("WaitGroup.new", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value { return value.NewWaitGroup() }),
			fn("WaitGroup.add", "wg, n", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if err := wgArg(g, who, a[0]).Add(intArg(g, who, a[1])); err != nil {
					g.Panic("WaitGroup.add: %s", err)
				}
				return value.Void
			}),
			fn("WaitGroup.done", "wg", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if err := wgArg(g, who, a[0]).Done(); err != nil {
					g.Panic("WaitGroup.done: %s", err)
				}
				return value.Void
			}),
			fn("WaitGroup.wait", "wg, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				w := wgArg(g, who, a[0])
				ctx := waitContext(g, who, a[1])
				g.Instance().Flush()
				if ctx.Err() == nil {
					select {
					case <-w.Zero():
					case <-ctx.Done():
					}
				}
				g.AfterWait()
				return value.Void
			}),
		)
	})
	register("Once", func() value.Value {
		return syncTag("Once",
			fn("Once.new", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value { return value.NewOnce() }),
			fn("Once.do", "once, f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				o := onceArg(g, who, a[0])
				f := funcArg(g, who, a[1])
				runOnce(g, o, f, func(value.Value) {})
				return value.Void
			}),
			fn("Once.value", "f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				f := funcArg(g, who, a[0])
				o := value.NewOnce()
				var cached value.Value
				return fn(who, "", func(g *interp.Goroutine, _ string, _ []value.Value) value.Value {
					if !runOnce(g, o, f, func(v value.Value) { cached = v }) {
						g.Panic("%s: the function was left by a return before it gave a value", who)
					}
					return cached
				})
			}),
		)
	})
	register("Cond", func() value.Value {
		return syncTag("Cond",
			fn("Cond.new", "m", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.NewCond(mutexArg(g, who, a[0]))
			}),
			fn("Cond.wait", "c", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				c := condArg(g, who, a[0])
				g.Instance().Flush()
				if err := c.Wait(); err != nil {
					g.Panic("%s", err)
				}
				return value.Void
			}),
			fn("Cond.signal", "c", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				condArg(g, who, a[0]).Signal()
				return value.Void
			}),
			fn("Cond.broadcast", "c", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				condArg(g, who, a[0]).Broadcast()
				return value.Void
			}),
		)
	})
	register("Atomic", func() value.Value {
		return syncTag("Atomic",
			fn("Atomic.new", "v", func(g *interp.Goroutine, who string, a []value.Value) value.Value { return value.NewAtomic(a[0]) }),
			fn("Atomic.load", "a", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return atomicArg(g, who, a[0]).Slot.Load()
			}),
			fn("Atomic.store", "a, v", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				atomicArg(g, who, a[0]).Slot.Store(a[1])
				return value.Void
			}),
			fn("Atomic.add", "a, d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				at := atomicArg(g, who, a[0])
				v, err := at.Slot.Modify(func(old value.Value) (value.Value, error) {
					return value.Binary("+", old, a[1])
				})
				if err != nil {
					g.Panic("Atomic.add: %s", err)
				}
				return v
			}),
			fn("Atomic.compareAndSwap", "a, expected, v", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return atomicArg(g, who, a[0]).Slot.CompareAndSwap(a[1], a[2])
			}),
		)
	})
}
