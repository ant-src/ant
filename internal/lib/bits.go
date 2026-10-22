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

// widthMask is the mask of a kind's bits.
func widthMask(k value.Kind) uint64 {
	if k.Bits() == 64 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(k.Bits()) - 1
}

// rotate is Bits.rotateLeft and Bits.rotateRight: by n modulo the width, so a count at or
// beyond the width wraps; a negative count panics as the shifts do.
func rotate(name string, right bool) *interp.Builtin {
	return fn(name, "a, n", func(g *interp.Goroutine, a []value.Value) value.Value {
		x, k := intBits(g, name, a[0])
		n := intArg(g, name, a[1])
		if n < 0 {
			g.Panic("%s needs a count of at least 0, got %d", name, n)
		}
		width := int64(k.Bits())
		n %= width
		if right {
			n = (width - n) % width
		}
		x &= widthMask(k)
		if n == 0 {
			return fromBits(k, x)
		}
		return fromBits(k, (x<<uint(n)|x>>uint(width-n))&widthMask(k))
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
			rotate("Bits.rotateLeft", false),
			rotate("Bits.rotateRight", true),
			fn("Bits.popCount", "a", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, k := intBits(g, "Bits.popCount", a[0])
				return int64(bits.OnesCount64(x & widthMask(k)))
			}),
			fn("Bits.leadingZeros", "a", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, k := intBits(g, "Bits.leadingZeros", a[0])
				return int64(bits.LeadingZeros64(x&widthMask(k)) - (64 - k.Bits()))
			}),
			fn("Bits.trailingZeros", "a", func(g *interp.Goroutine, a []value.Value) value.Value {
				x, k := intBits(g, "Bits.trailingZeros", a[0])
				if x&widthMask(k) == 0 {
					return int64(k.Bits())
				}
				return int64(bits.TrailingZeros64(x & widthMask(k)))
			}),
			binaryBits("Bits.wrapAdd", func(x, y uint64) uint64 { return x + y }),
			binaryBits("Bits.wrapSub", func(x, y uint64) uint64 { return x - y }),
			binaryBits("Bits.wrapMul", func(x, y uint64) uint64 { return x * y }),
		)
	})
}
