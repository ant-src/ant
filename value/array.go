package value

import (
	"sync"
	"sync/atomic"
)

// Array is a dense, growable sequence. Reads and element stores go through an immutable header
// behind an atomic pointer; push and pop take the mutex and publish a new header. A concurrent
// push or pop leaves a reader the state before or after it, never a corrupted structure.
type Array struct {
	h    atomic.Pointer[arrayHeader]
	mu   sync.Mutex
	self cell // holds the array, for slots that store it
}

type arrayHeader struct {
	elems []Slot // capacity; only the first n are live
	n     int
}

// NewArray creates an array holding elems.
func NewArray(elems []Value) *Array {
	a := &Array{}
	a.self.v = a
	h := &arrayHeader{elems: make([]Slot, len(elems)), n: len(elems)}
	for i, v := range elems {
		h.elems[i].Store(v)
	}
	a.h.Store(h)
	return a
}

// Len is the current length.
func (a *Array) Len() int { return a.h.Load().n }

// Get reads the element at i, or reports false when i is out of range.
func (a *Array) Get(i int) (Value, bool) {
	h := a.h.Load()
	if i < 0 || i >= h.n {
		return nil, false
	}
	return h.elems[i].Load(), true
}

// Set stores v at i, or reports false when i is out of range.
func (a *Array) Set(i int, v Value) bool {
	h := a.h.Load()
	if i < 0 || i >= h.n {
		return false
	}
	h.elems[i].Store(v)
	return true
}

// Push appends v.
func (a *Array) Push(v Value) {
	a.mu.Lock()
	defer a.mu.Unlock()
	h := a.h.Load()
	if h.n < len(h.elems) {
		h.elems[h.n].Store(v)
		a.h.Store(&arrayHeader{elems: h.elems, n: h.n + 1})
		return
	}
	elems := make([]Slot, growCap(len(h.elems)))
	for i := 0; i < h.n; i++ {
		elems[i].Store(h.elems[i].Load())
	}
	elems[h.n].Store(v)
	a.h.Store(&arrayHeader{elems: elems, n: h.n + 1})
}

func growCap(c int) int {
	if c < 4 {
		return 4
	}
	return c * 2
}

// Pop removes and returns the last element, or reports false when the array is empty.
func (a *Array) Pop() (Value, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	h := a.h.Load()
	if h.n == 0 {
		return nil, false
	}
	v := h.elems[h.n-1].Load()
	h.elems[h.n-1].Store(nil)
	a.h.Store(&arrayHeader{elems: h.elems, n: h.n - 1})
	return v, true
}

// Values copies the elements out, from one consistent header.
func (a *Array) Values() []Value {
	h := a.h.Load()
	out := make([]Value, h.n)
	for i := 0; i < h.n; i++ {
		out[i] = h.elems[i].Load()
	}
	return out
}
