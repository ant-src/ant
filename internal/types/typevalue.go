package types

import (
	"hash/maphash"
	"unsafe"

	"github.com/ant-src/ant/value"
)

// TypeValue is a type as a run-time value (section 4, Types as values): what a type expression
// evaluates to and what the Reflect functions pass around. It implements value.TypeValue, so
// the runtime compares, hashes, and renders type values with the checker's own relations.
type TypeValue struct {
	T Type
}

// EqualType reports whether two type values denote the same type.
func (v *TypeValue) EqualType(o value.TypeValue) bool {
	w, ok := o.(*TypeValue)
	return ok && Equal(v.T, w.T)
}

// HashType agrees with EqualType.
func (v *TypeValue) HashType() uint64 { return Hash(v.T) }

// String renders the type in annotation syntax.
func (v *TypeValue) String() string { return String(v.T) }

// DeclValue is a declaration as a value, Reflect.Decl, compared by the declaration's identity.
func DeclValue(d *Decl) value.Declaration { return value.Declaration{Key: d, Name: d.Name} }

var hashSeed = maphash.MakeSeed()

// Hash agrees with Equal: equal types hash alike. It expands aliases and hashes a bounded depth
// of the structure, so a recursive alias such as Json.Value terminates.
func Hash(t Type) uint64 {
	var h maphash.Hash
	h.SetSeed(hashSeed)
	hashType(&h, t, 0)
	return h.Sum64()
}

const hashTypeDepth = 3

func hashType(h *maphash.Hash, t Type, depth int) {
	t = Expand(t)
	switch x := t.(type) {
	case Prim:
		h.WriteByte(1)
		h.WriteByte(byte(x.Kind))
	case Null:
		h.WriteByte(2)
	case Void:
		h.WriteByte(3)
	case Never:
		h.WriteByte(4)
	case Invalid:
		h.WriteByte(5)
	case Array:
		h.WriteByte(6)
		if depth < hashTypeDepth {
			hashType(h, x.Elem, depth+1)
		}
	case Map:
		h.WriteByte(7)
		if depth < hashTypeDepth {
			hashType(h, x.Key, depth+1)
			hashType(h, x.Val, depth+1)
		}
	case Set:
		h.WriteByte(8)
		if depth < hashTypeDepth {
			hashType(h, x.Elem, depth+1)
		}
	case Func:
		h.WriteByte(9)
		h.WriteByte(byte(len(x.Params)))
		if depth < hashTypeDepth {
			for _, p := range x.Params {
				hashType(h, p, depth+1)
			}
			hashType(h, x.Return, depth+1)
		}
	case Record:
		// Field order is not part of an anonymous record's shape, so the
		// fields' hashes combine in any order, as union members do.
		h.WriteByte(10)
		h.WriteByte(byte(len(x.Fields)))
		var sum uint64
		for _, f := range x.Fields {
			var fh maphash.Hash
			fh.SetSeed(hashSeed)
			fh.WriteString(f.Name)
			fh.WriteByte(0)
			if depth < hashTypeDepth {
				hashType(&fh, f.Type, depth+1)
			}
			sum += fh.Sum64()
		}
		writeUint(h, sum)
	case Named:
		h.WriteByte(11)
		writeUint(h, uint64(uintptr(unsafe.Pointer(x.Decl))))
		if depth < hashTypeDepth {
			for _, a := range x.Args {
				hashType(h, a, depth+1)
			}
		}
	case Param:
		h.WriteByte(12)
		writeUint(h, uint64(uintptr(unsafe.Pointer(x.Decl))))
	case Union:
		// Members are a set, so their hashes combine in any order.
		h.WriteByte(13)
		var sum uint64
		for _, m := range x.Members {
			var mh maphash.Hash
			mh.SetSeed(hashSeed)
			hashType(&mh, m, depth+1)
			sum += mh.Sum64()
		}
		writeUint(h, sum)
	}
}

func writeUint(h *maphash.Hash, u uint64) {
	var b [8]byte
	for i := range b {
		b[i] = byte(u >> (8 * i))
	}
	h.Write(b[:])
}
