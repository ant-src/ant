package value

import (
	"hash/maphash"
	"math"
)

// Equal is == (section 12): structural, never panics, and finite on cyclic values.
func Equal(a, b Value) bool {
	return equal(a, b, nil)
}

type pair struct{ a, b any }

func equal(a, b Value, visiting map[pair]bool) bool {
	switch x := a.(type) {
	case int64:
		y, ok := b.(int64)
		return ok && x == y
	case float64:
		y, ok := b.(float64)
		return ok && x == y
	case string:
		y, ok := b.(string)
		return ok && x == y
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case nullType:
		_, ok := b.(nullType)
		return ok
	case voidType:
		_, ok := b.(voidType)
		return ok
	case int8:
		y, ok := b.(int8)
		return ok && x == y
	case int16:
		y, ok := b.(int16)
		return ok && x == y
	case int32:
		y, ok := b.(int32)
		return ok && x == y
	case uint8:
		y, ok := b.(uint8)
		return ok && x == y
	case uint16:
		y, ok := b.(uint16)
		return ok && x == y
	case uint32:
		y, ok := b.(uint32)
		return ok && x == y
	case uint64:
		y, ok := b.(uint64)
		return ok && x == y
	case float32:
		y, ok := b.(float32)
		return ok && x == y
	case *Tagged:
		y, ok := b.(*Tagged)
		return ok && x.Tag == y.Tag && equal(x.Payload, y.Payload, visiting)
	case *Record:
		y, ok := b.(*Record)
		if !ok || x.Tag != y.Tag {
			return false
		}
		if x == y {
			return true
		}
		if visiting == nil {
			visiting = map[pair]bool{}
		}
		p := pair{x, y}
		if visiting[p] {
			return true
		}
		visiting[p] = true
		defer delete(visiting, p)
		for i := 0; i < x.Len(); i++ {
			if !equal(x.Get(i), y.Get(i), visiting) {
				return false
			}
		}
		return true
	case *Array:
		y, ok := b.(*Array)
		if !ok {
			return false
		}
		if x == y {
			return true
		}
		if visiting == nil {
			visiting = map[pair]bool{}
		}
		p := pair{x, y}
		if visiting[p] {
			return true
		}
		visiting[p] = true
		defer delete(visiting, p)
		xs, ys := x.Values(), y.Values()
		if len(xs) != len(ys) {
			return false
		}
		for i := range xs {
			if !equal(xs[i], ys[i], visiting) {
				return false
			}
		}
		return true
	case *Map:
		y, ok := b.(*Map)
		if !ok {
			return false
		}
		if x == y {
			return true
		}
		if visiting == nil {
			visiting = map[pair]bool{}
		}
		p := pair{x, y}
		if visiting[p] {
			return true
		}
		visiting[p] = true
		defer delete(visiting, p)
		xs := x.Entries()
		if len(xs) != y.Len() {
			return false
		}
		for _, e := range xs {
			v, ok := y.Get(e.Key)
			if !ok || !equal(e.Value, v, visiting) {
				return false
			}
		}
		return true
	case *Set:
		y, ok := b.(*Set)
		if !ok {
			return false
		}
		if x == y {
			return true
		}
		if visiting == nil {
			visiting = map[pair]bool{}
		}
		p := pair{x, y}
		if visiting[p] {
			return true
		}
		visiting[p] = true
		defer delete(visiting, p)
		xs := x.Values()
		if len(xs) != y.Len() {
			return false
		}
		for _, e := range xs {
			if !y.Has(e) {
				return false
			}
		}
		return true
	case *Tag:
		y, ok := b.(*Tag)
		return ok && x == y
	case Callable:
		return a == b
	case Sync:
		return a == b
	case Time:
		y, ok := b.(Time)
		return ok && x.TimeType() == y.TimeType() && x.Key() == y.Key()
	case *Regex:
		y, ok := b.(*Regex)
		return ok && x.Source == y.Source && x.Flags == y.Flags
	}
	return false
}

var seed = maphash.MakeSeed()

// Hash agrees with Equal: equal values hash alike. It walks a bounded prefix of a container, so
// it terminates on cyclic values.
func Hash(v Value) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	hashInto(&h, v, 0)
	return h.Sum64()
}

const (
	hashDepth = 3
	hashWidth = 8
)

func hashInto(h *maphash.Hash, v Value, depth int) {
	k := KindOf(v)
	h.WriteByte(byte(k))
	switch x := v.(type) {
	case int64:
		writeUint(h, uint64(x))
	case int8:
		writeUint(h, uint64(x))
	case int16:
		writeUint(h, uint64(x))
	case int32:
		writeUint(h, uint64(x))
	case uint8:
		writeUint(h, uint64(x))
	case uint16:
		writeUint(h, uint64(x))
	case uint32:
		writeUint(h, uint64(x))
	case uint64:
		writeUint(h, x)
	case float64:
		if x == 0 {
			x = 0 // -0.0 and 0.0 are equal
		}
		writeUint(h, math.Float64bits(x))
	case float32:
		if x == 0 {
			x = 0
		}
		writeUint(h, uint64(math.Float32bits(x)))
	case string:
		h.WriteString(x)
	case bool:
		if x {
			h.WriteByte(1)
		} else {
			h.WriteByte(0)
		}
	case *Tagged:
		writeUint(h, uint64(uintptr(tagID(x.Tag))))
		hashInto(h, x.Payload, depth)
	case *Record:
		writeUint(h, uint64(uintptr(tagID(x.Tag))))
		if depth < hashDepth {
			n := x.Len()
			if n > hashWidth {
				n = hashWidth
			}
			for i := 0; i < n; i++ {
				hashInto(h, x.Get(i), depth+1)
			}
		}
	case *Array:
		xs := x.Values()
		writeUint(h, uint64(len(xs)))
		if depth < hashDepth {
			n := len(xs)
			if n > hashWidth {
				n = hashWidth
			}
			for i := 0; i < n; i++ {
				hashInto(h, xs[i], depth+1)
			}
		}
	case *Map:
		// Insertion order is not part of equality, so only the size contributes.
		writeUint(h, uint64(x.Len()))
	case *Set:
		writeUint(h, uint64(x.Len())) // likewise
	case *Tag:
		writeUint(h, uint64(uintptr(tagID(x))))
	case Time:
		h.WriteString(x.TimeType())
		h.WriteString(x.Key())
	case *Regex:
		h.WriteString("regex")
		h.WriteString(x.Source)
		h.WriteString(x.Flags)
	default:
		// Functions and sync objects compare by identity; hash their identity.
		writeUint(h, identity(v))
	}
}

func writeUint(h *maphash.Hash, u uint64) {
	var b [8]byte
	for i := range b {
		b[i] = byte(u >> (8 * i))
	}
	h.Write(b[:])
}
