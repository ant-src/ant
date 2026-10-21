package lib

import (
	"math/bits"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// Bitwise and wrapping operations on the eight integer kinds (section 11). Every operand is
// taken to its bits as a uint64, operated on, and truncated back to its kind.

func intBits(g *interp.Goroutine, who string, v value.Value) (uint64, value.Kind) {
	switch x := v.(type) {
	case int64:
		return uint64(x), value.KindInt64
	case int8:
		return uint64(x), value.KindInt8
	case int16:
		return uint64(x), value.KindInt16
	case int32:
		return uint64(x), value.KindInt32
	case uint8:
		return uint64(x), value.KindUint8
	case uint16:
		return uint64(x), value.KindUint16
	case uint32:
		return uint64(x), value.KindUint32
	case uint64:
		return x, value.KindUint64
	}
	g.Panic("%s needs an integer, got %s", who, value.Describe(v))
	return 0, value.KindInvalid
}

// fromBits truncates u to k's width and reinterprets it in k.
func fromBits(k value.Kind, u uint64) value.Value {
	switch k {
	case value.KindInt8:
		return int8(u)
	case value.KindInt16:
		return int16(u)
	case value.KindInt32:
		return int32(u)
	case value.KindInt64:
		return int64(u)
	case value.KindUint8:
		return uint8(u)
	case value.KindUint16:
		return uint16(u)
	case value.KindUint32:
		return uint32(u)
	}
	return u
}

func pairBits(g *interp.Goroutine, who string, a, b value.Value) (uint64, uint64, value.Kind) {
	x, kx := intBits(g, who, a)
	y, ky := intBits(g, who, b)
	if kx != ky {
		g.Panic("%s needs two integers of one kind, got %s and %s", who, kx, ky)
	}
	return x, y, kx
}

func binaryBits(name string, op func(x, y uint64) uint64) *interp.Builtin {
	return fn(name, "a, b", func(g *interp.Goroutine, a []value.Value) value.Value {
		x, y, k := pairBits(g, name, a[0], a[1])
		return fromBits(k, op(x, y))
	})
}

func shift(name string, right bool) *interp.Builtin {
	return fn(name, "a, n", func(g *interp.Goroutine, a []value.Value) value.Value {
		x, k := intBits(g, name, a[0])
		n := intArg(g, name, a[1])
		width := int64(k.Bits())
		if n < 0 || n >= width {
			g.Panic("%s needs a count from 0 up to but excluding %d, got %d", name, width, n)
		}
		if !right {
			return fromBits(k, x<<uint(n))
		}
		if k.IsSigned() {
			// arithmetic: sign-extend from the kind's width first
			s := int64(x) << uint(64-width) >> uint(64-width)
			return fromBits(k, uint64(s>>uint(n)))
		}
		return fromBits(k, x>>uint(n))
	})
}

func init() {
	register("Bits", func() value.Value {
		return namespace("Bits",
			binaryBits("Bits.and", func(x, y uint64) uint64 { return x & y }),
			binaryBits("Bits.or", func(x, y uint64) uint64 { return x | y }),
			binaryBits("Bits.xor", func(x, y uint64) uint64 { return x ^ y }),
			fn("Bits.not", "a", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, k := intBits(g, "Bits.not", a[0])
				return fromBits(k, ^x)
			}),
			shift("Bits.shiftLeft", false),
			shift("Bits.shiftRight", true),
			fn("Bits.popCount", "a", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, k := intBits(g, "Bits.popCount", a[0])
				mask := uint64(1)<<uint(k.Bits()) - 1
				if k.Bits() == 64 {
					mask = ^uint64(0)
				}
				return int64(bits.OnesCount64(x & mask))
			}),
			binaryBits("Bits.wrapAdd", func(x, y uint64) uint64 { return x + y }),
			binaryBits("Bits.wrapSub", func(x, y uint64) uint64 { return x - y }),
			binaryBits("Bits.wrapMul", func(x, y uint64) uint64 { return x * y }),
		)
	})
}
