package value

import (
	"hash/maphash"
	"math"
)

// Equal is == (section 12): structural, never panics, and finite on cyclic values. It never
// shortcuts on identity, so a container holding NaN is unequal even to itself.
func Equal(a, b Value) bool {
	return equal(a, b, nil)
}

type pair struct{ a, b any }

func equal(a, b Value, visiting map[pair]bool) bool {
	switch x := a.(type) {
	case int64, int8, int16, int32, uint8, uint16, uint32, uint64, float64, float32, string, bool,
		nullType, voidType, *Tag:
		// Interface comparison checks the dynamic type first, so float32 never equals float64,
		// and it compares floats by value, so NaN is unequal to itself and -0.0 equals 0.0.
		return a == b
	case *Tagged:
		y, ok := b.(*Tagged)
		return ok && x.Tag == y.Tag && equal(x.Payload, y.Payload, visiting)
	case *Record, *Array, *Map, *Set:
		// A pair already under comparison is assumed equal, so cyclic values compare finitely.
		if visiting == nil {
			visiting = map[pair]bool{}
		}
		p := pair{a, b}
		if visiting[p] {
			return true
		}
		visiting[p] = true
		defer delete(visiting, p)
		return equalContainer(a, b, visiting)
	case Callable, Sync:
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

// equalContainer compares a record, array, map or set with b, element by element.
func equalContainer(a, b Value, visiting map[pair]bool) bool {
	switch x := a.(type) {
	case *Record:
		y, ok := b.(*Record)
		if !ok || x.Tag != y.Tag {
			return false
		}
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
		if !ok || x.Len() != y.Len() {
			return false
		}
		for _, e := range x.Entries() {
			v, ok := lookup(y, e.Key, visiting)
			if !ok || !equal(e.Value, v, visiting) {
				return false
			}
		}
		return true
	case *Set:
		y, ok := b.(*Set)
		if !ok || x.Len() != y.Len() {
			return false
		}
		for _, e := range x.Values() {
			if _, ok := lookup(&y.m, e, visiting); !ok {
				return false
			}
		}
		return true
	}
	return false
}

// lookup finds the value under k in m, matching keys with the pairs already under comparison,
// so that keys which contain the maps being compared do not restart the comparison forever.
func lookup(m *Map, k Value, visiting map[pair]bool) (Value, bool) {
	for _, e := range m.bucket(Hash(k)) {
		if equal(e.Key, k, visiting) {
			return e.Value, true
		}
	}
	return nil, false
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
	case int64, int8, int16, int32:
		writeUint(h, uint64(toInt64(v)))
	case uint8, uint16, uint32, uint64:
		writeUint(h, toUint64(v))
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
