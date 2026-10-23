package value

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"
)

// The sync objects of section 9, as small state machines guarded by Go primitives. Each
// reports the misuse section 11 lists as an error for the caller to turn into a panic.

// Mutex is a mutual exclusion lock that any goroutine may unlock.
type Mutex struct {
	mu     sync.Mutex
	cond   *sync.Cond
	locked bool
}

// NewMutex creates an unlocked Mutex.
func NewMutex() *Mutex {
	m := &Mutex{}
	m.cond = sync.NewCond(&m.mu)
	return m
}

// SyncName returns "Mutex".
func (m *Mutex) SyncName() string { return "Mutex" }

// Lock blocks until the mutex is free, then takes it.
func (m *Mutex) Lock() {
	m.mu.Lock()
	for m.locked {
		m.cond.Wait()
	}
	m.locked = true
	m.mu.Unlock()
}

// Unlock releases the mutex; it is an error when the mutex is not locked.
func (m *Mutex) Unlock() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.locked {
		return errors.New("unlock of a Mutex that is not locked")
	}
	m.locked = false
	m.cond.Signal()
	return nil
}

// Locked reports whether the mutex is held.
func (m *Mutex) Locked() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.locked
}

// RwMutex is a reader-writer lock.
type RwMutex struct {
	mu      sync.Mutex
	cond    *sync.Cond
	readers int
	writer  bool
}

// NewRwMutex creates an unlocked RwMutex.
func NewRwMutex() *RwMutex {
	m := &RwMutex{}
	m.cond = sync.NewCond(&m.mu)
	return m
}

// SyncName returns "RwMutex".
func (m *RwMutex) SyncName() string { return "RwMutex" }

// ReadLock takes a shared lock.
func (m *RwMutex) ReadLock() {
	m.mu.Lock()
	for m.writer {
		m.cond.Wait()
	}
	m.readers++
	m.mu.Unlock()
}

// ReadUnlock releases a shared lock.
func (m *RwMutex) ReadUnlock() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.readers == 0 {
		return errors.New("readUnlock of an RwMutex that is not read-locked")
	}
	m.readers--
	if m.readers == 0 {
		m.cond.Broadcast()
	}
	return nil
}

// WriteLock takes the exclusive lock.
func (m *RwMutex) WriteLock() {
	m.mu.Lock()
	for m.writer || m.readers > 0 {
		m.cond.Wait()
	}
	m.writer = true
	m.mu.Unlock()
}

// WriteUnlock releases the exclusive lock.
func (m *RwMutex) WriteUnlock() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.writer {
		return errors.New("writeUnlock of an RwMutex that is not write-locked")
	}
	m.writer = false
	m.cond.Broadcast()
	return nil
}

// WaitGroup counts goroutines to wait for. A channel is closed when the counter reaches zero
// and replaced at the next add from zero, so a wait with a context is a select that leaks
// nothing when it gives up.
type WaitGroup struct {
	mu   sync.Mutex
	n    int64
	zero chan struct{}
}

// NewWaitGroup creates a WaitGroup at zero.
func NewWaitGroup() *WaitGroup {
	w := &WaitGroup{zero: make(chan struct{})}
	close(w.zero)
	return w
}

// SyncName returns "WaitGroup".
func (w *WaitGroup) SyncName() string { return "WaitGroup" }

// Add changes the counter by n; an error when it would go negative.
func (w *WaitGroup) Add(n int64) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.n+n < 0 && n < 0 {
		return errors.New("WaitGroup counter would go negative")
	}
	if n > 0 && w.n > math.MaxInt64-n {
		return errors.New("WaitGroup counter would overflow")
	}
	if w.n == 0 && n > 0 {
		w.zero = make(chan struct{})
	}
	w.n += n
	if w.n == 0 && n != 0 {
		close(w.zero)
	}
	return nil
}

// Done decrements the counter; an error when it is already zero.
func (w *WaitGroup) Done() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.n == 0 {
		return errors.New("WaitGroup.done when the counter is already zero")
	}
	w.n--
	if w.n == 0 {
		close(w.zero)
	}
	return nil
}

// Zero returns a channel that is closed once the counter reaches zero. A counter at zero gives
// a closed channel.
func (w *WaitGroup) Zero() <-chan struct{} {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.zero
}

// Wait blocks until the counter is zero.
func (w *WaitGroup) Wait() { <-w.Zero() }

// The reasons a Context is done, as context causes; Reason names them.
var (
	ErrCancelled = errors.New("cancelled")
	ErrTimeout   = errors.New("timeout")
	ErrInterrupt = errors.New("interrupt")
)

// Context is a Context of section 9: done once, by a call, a deadline or its parent, with a
// reason. It wraps a Go context whose cause is the reason, so every blocking call selects on
// Ctx.Done and reads Reason.
type Context struct {
	Ctx    context.Context
	cancel context.CancelCauseFunc
	stop   context.CancelFunc // releases the timer of a context with a deadline
}

// NewContext makes a child of parent that is done when parent is, when timeout passes if it is
// positive, or when Cancel is called. The timeout layer sits on the parent and the cancel layer
// on it, so that the deadline passing cancels both and unregisters both from the parent; the
// other way round the outer layer would stay in the parent until Cancel, which a program that
// lets its timeouts expire never calls.
func NewContext(parent context.Context, timeout time.Duration) *Context {
	c := &Context{}
	if timeout > 0 {
		parent, c.stop = context.WithTimeoutCause(parent, timeout, ErrTimeout)
	}
	c.Ctx, c.cancel = context.WithCancelCause(parent)
	return c
}

// SyncName returns "Context".
func (c *Context) SyncName() string { return "Context" }

// Cancel makes the context done with cause, unless it is done already.
func (c *Context) Cancel(cause error) {
	c.cancel(cause)
	if c.stop != nil {
		c.stop()
	}
}

// Done reports whether the context is done.
func (c *Context) Done() bool { return c.Ctx.Err() != nil }

// Reason names why the context is done, or "" while it runs: cancelled, timeout or interrupt,
// a parent's reason, and cancelled for the host ending the run.
func (c *Context) Reason() string { return ReasonOf(c.Ctx) }

// ReasonOf names why a Go context is done, as Reason does, for the run's own context too. A
// program's timeout carries ErrTimeout as its cause, so Go's own deadline error can only be the
// host's context expiring, which is the host ending the run: cancelled, as section 9 says.
func ReasonOf(ctx context.Context) string {
	cause := context.Cause(ctx)
	switch {
	case cause == nil:
		return ""
	case errors.Is(cause, ErrTimeout):
		return "timeout"
	case errors.Is(cause, ErrInterrupt):
		return "interrupt"
	}
	return "cancelled"
}

// Once runs one function once. If that function panics, every call, waiting or later, gets the
// same panic.
type Once struct {
	mu    sync.Mutex
	cond  *sync.Cond
	state int // 0 pending, 1 running, 2 done
	err   error
}

// NewOnce creates a pending Once.
func NewOnce() *Once {
	o := &Once{}
	o.cond = sync.NewCond(&o.mu)
	return o
}

// SyncName returns "Once".
func (o *Once) SyncName() string { return "Once" }

// Do runs f the first time it is called; later and concurrent calls wait for that run and
// return the error it returned, which is how a panic in f is shared.
func (o *Once) Do(f func() error) error {
	o.mu.Lock()
	switch o.state {
	case 2:
		o.mu.Unlock()
		return o.err
	case 1:
		for o.state != 2 {
			o.cond.Wait()
		}
		o.mu.Unlock()
		return o.err
	}
	o.state = 1
	o.mu.Unlock()
	err := f()
	o.mu.Lock()
	o.state = 2
	o.err = err
	o.cond.Broadcast()
	o.mu.Unlock()
	return err
}

// Cond is a condition variable over a Mutex.
type Cond struct {
	M    *Mutex
	mu   sync.Mutex
	cond *sync.Cond
}

// NewCond creates a Cond over m.
func NewCond(m *Mutex) *Cond {
	c := &Cond{M: m}
	c.cond = sync.NewCond(&c.mu)
	return c
}

// SyncName returns "Cond".
func (c *Cond) SyncName() string { return "Cond" }

// Wait releases the mutex, waits for a signal, and reacquires the mutex. It is an error when
// the mutex is not locked.
func (c *Cond) Wait() error {
	c.mu.Lock()
	if err := c.M.Unlock(); err != nil {
		c.mu.Unlock()
		return errors.New("Cond.wait when its Mutex is not locked")
	}
	c.cond.Wait()
	c.mu.Unlock()
	c.M.Lock()
	return nil
}

// Signal wakes one waiter.
func (c *Cond) Signal() {
	c.mu.Lock()
	c.cond.Signal()
	c.mu.Unlock()
}

// Broadcast wakes every waiter.
func (c *Cond) Broadcast() {
	c.mu.Lock()
	c.cond.Broadcast()
	c.mu.Unlock()
}

// Atomic is a shared slot whose operations form one total order.
type Atomic struct {
	Slot Slot
}

// NewAtomic creates an Atomic holding v.
func NewAtomic(v Value) *Atomic {
	a := &Atomic{}
	a.Slot.Store(v)
	return a
}

// SyncName returns "Atomic".
func (a *Atomic) SyncName() string { return "Atomic" }

// TryLock takes the mutex if it is free and reports whether it did.
func (m *Mutex) TryLock() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.locked {
		return false
	}
	m.locked = true
	return true
}
