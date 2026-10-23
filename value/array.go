package value

import (
	"sync"
	"sync/atomic"
)

// Array is a dense, growable sequence. Reads and element stores go through an immutable header
// behind an atomic pointer; push and pop take the mutex and publish a new header. A concurrent
// push or pop leaves a reader the state before or after it, never a corrupted structure.
type Array struct {
	h  atomic.Pointer[arrayHeader]
	mu sync.Mutex
}

type arrayHeader struct {
	elems []Slot // capacity; only the first n are live
	n     int
}

// NewArray creates an array holding elems.
func NewArray(elems []Value) *Array {
	a := &Array{}
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
	// The slot keeps its value: a reader holding the old header may still read it, and must
	// see a value that was stored there, never a cleared one.
	v := h.elems[h.n-1].Load()
	a.h.Store(&arrayHeader{elems: h.elems, n: h.n - 1})
	return v, true
}

// Insert stores v at i and shifts the elements from i on up by one, or reports false when i is
// outside 0 ..= Len(). The elements go into a fresh slice, so a reader of the old header sees
// the array as it was.
func (a *Array) Insert(i int, v Value) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	h := a.h.Load()
	if i < 0 || i > h.n {
		return false
	}
	c := len(h.elems)
	if h.n == c {
		c = growCap(c)
	}
	elems := make([]Slot, c)
	for j := 0; j < i; j++ {
		elems[j].Store(h.elems[j].Load())
	}
	elems[i].Store(v)
	for j := i; j < h.n; j++ {
		elems[j+1].Store(h.elems[j].Load())
	}
	a.h.Store(&arrayHeader{elems: elems, n: h.n + 1})
	return true
}

// Remove removes and returns the element at i, shifting the rest down by one, or reports false
// when i is out of range. As Insert, it publishes a fresh slice.
func (a *Array) Remove(i int) (Value, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	h := a.h.Load()
	if i < 0 || i >= h.n {
		return nil, false
	}
	v := h.elems[i].Load()
	elems := make([]Slot, len(h.elems))
	for j := 0; j < i; j++ {
		elems[j].Store(h.elems[j].Load())
	}
	for j := i + 1; j < h.n; j++ {
		elems[j-1].Store(h.elems[j].Load())
	}
	a.h.Store(&arrayHeader{elems: elems, n: h.n - 1})
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
