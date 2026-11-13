package value

import (
	"strings"
	"sync"
	"sync/atomic"
)

// Bytes is a mutable, growable sequence of bytes, shared by reference as an array is: Bytes of
// section 11. The storage sits behind an atomic pointer and the length in an atomic word, so that
// a push into spare capacity publishes nothing but the new length. Growth copies into larger
// storage and publishes it before the length passes the old capacity, so a reader that loads the
// length and then the storage always finds the storage covering that length. Functions that change
// the length take the mutex; reads and stores of a byte take nothing.
type Bytes struct {
	s  atomic.Pointer[[]byte] // the storage; its length is the capacity
	n  atomic.Int64           // the live length
	mu sync.Mutex
}

// NewBytes makes Bytes holding b, adopting b's array: the caller gives it up.
func NewBytes(b []byte) *Bytes {
	x := &Bytes{}
	s := b[:cap(b)]
	x.s.Store(&s)
	x.n.Store(int64(len(b)))
	return x
}

// CopyBytes makes Bytes holding a copy of b.
func CopyBytes(b []byte) *Bytes {
	return NewBytes(append([]byte(nil), b...))
}

// View is the live bytes, sharing the storage. A caller only reads it and does not keep it past the
// call it serves; a concurrent change is a data race, which leaves it some bytes that were stored.
func (x *Bytes) View() []byte {
	n := x.n.Load()
	s := *x.s.Load()
	return s[:n]
}

// Snapshot is a copy of the live bytes.
func (x *Bytes) Snapshot() []byte { return append([]byte(nil), x.View()...) }

// Len is the current length.
func (x *Bytes) Len() int { return int(x.n.Load()) }

// Get reads the byte at i, or reports false when i is out of range.
func (x *Bytes) Get(i int) (byte, bool) {
	v := x.View()
	if i < 0 || i >= len(v) {
		return 0, false
	}
	return v[i], true
}

// Set stores c at i, or reports false when i is out of range.
func (x *Bytes) Set(i int, c byte) bool {
	v := x.View()
	if i < 0 || i >= len(v) {
		return false
	}
	v[i] = c
	return true
}

// room makes space for k more bytes and returns the storage and the current length; the caller
// holds the mutex and stores the new length after writing.
func (x *Bytes) room(k int) ([]byte, int) {
	n := int(x.n.Load())
	s := *x.s.Load()
	if n+k <= len(s) {
		return s, n
	}
	c := len(s) * 2
	if c < 8 {
		c = 8
	}
	for c < n+k {
		c *= 2
	}
	ns := make([]byte, c)
	copy(ns, s[:n])
	x.s.Store(&ns)
	return ns, n
}

// Push appends c.
func (x *Bytes) Push(c byte) {
	x.mu.Lock()
	s, n := x.room(1)
	s[n] = c
	x.n.Store(int64(n + 1))
	x.mu.Unlock()
}

// Append appends src. src may be a view of x itself: growth leaves it on the old storage, and
// without growth it ends where the new bytes begin.
func (x *Bytes) Append(src []byte) {
	if len(src) == 0 {
		return
	}
	x.mu.Lock()
	s, n := x.room(len(src))
	copy(s[n:], src)
	x.n.Store(int64(n + len(src)))
	x.mu.Unlock()
}

// Pop removes and returns the last byte, or reports false when x is empty.
func (x *Bytes) Pop() (byte, bool) {
	x.mu.Lock()
	defer x.mu.Unlock()
	n := int(x.n.Load())
	if n == 0 {
		return 0, false
	}
	c := (*x.s.Load())[n-1]
	x.n.Store(int64(n - 1))
	return c, true
}

// Truncate keeps the first n bytes, or reports false when n is outside 0 ..= length. The storage
// stays for reuse; growth overwrites each byte it makes visible again.
func (x *Bytes) Truncate(n int) bool {
	x.mu.Lock()
	defer x.mu.Unlock()
	if n < 0 || int64(n) > x.n.Load() {
		return false
	}
	x.n.Store(int64(n))
	return true
}

// renderBytes writes b as String.from does: Bytes(89 50 4e 47), two lower-case hexadecimal
// digits per byte.
func renderBytes(sb *strings.Builder, x *Bytes) {
	const digits = "0123456789abcdef"
	b := x.View()
	sb.Grow(len("Bytes()") + 3*len(b))
	sb.WriteString("Bytes(")
	for i := 0; i < len(b); i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteByte(digits[b[i]>>4])
		sb.WriteByte(digits[b[i]&0xf])
	}
	sb.WriteByte(')')
}
