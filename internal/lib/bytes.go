package lib

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

func bytesArg(g *interp.Goroutine, who string, v value.Value) *value.Bytes {
	return typedArg[*value.Bytes](g, who, "Bytes", v)
}

// view is b's live bytes without copying them, for a callee that only reads them during the call:
// a hash, an encoder, a synchronous write (the ownership rule of section 11, Bytes).
func view(b *value.Bytes) []byte { return b.View() }

// bytesValue makes Bytes of b, adopting b's array. Pass only a slice nothing else holds, such as
// what a read or a hash returned: what the library gives is the caller's.
func bytesValue(b []byte) *value.Bytes { return value.NewBytes(b) }

// textOrBytes reads a string | Bytes argument as the bytes to send: a string as its UTF-8, Bytes
// copied, so that no goroutine of the runtime reads the program's bytes after the call returns.
func textOrBytes(g *interp.Goroutine, who string, v value.Value) string {
	switch x := v.(type) {
	case string:
		return x
	case *value.Bytes:
		return string(x.View())
	}
	g.Panic("%s needs a string or Bytes, got %s", who, value.Describe(v))
	return ""
}

func byteArg(g *interp.Goroutine, who string, v value.Value) byte {
	return typedArg[uint8](g, who, "a byte", v)
}

// optTo is an optional to: the length when absent or null.
func optTo(g *interp.Goroutine, who string, v value.Value, length int) int64 {
	if !given(v) {
		return int64(length)
	}
	return intArg(g, who, v)
}

func optOffset(g *interp.Goroutine, who string, v value.Value) int64 {
	if !given(v) {
		return 0
	}
	return intArg(g, who, v)
}

// orderArg reads a Bytes.Order: true for little-endian.
func orderArg(g *interp.Goroutine, who string, v value.Value) binary.ByteOrder {
	if r, ok := v.(*value.Record); ok {
		switch r.Tag.Name {
		case "Bytes.BigEndian":
			return binary.BigEndian
		case "Bytes.LittleEndian":
			return binary.LittleEndian
		}
	}
	g.Panic("%s needs a Bytes.Order, got %s", who, value.Describe(v))
	return nil
}

// numberKind reads a type value that must be one number type, and its width in bytes.
func numberKind(g *interp.Goroutine, who string, v value.Value) (value.Kind, int) {
	if p, ok := types.Expand(typeArg(g, who, v).T).(types.Prim); ok {
		switch p.Kind {
		case value.KindInt8, value.KindUint8:
			return p.Kind, 1
		case value.KindInt16, value.KindUint16:
			return p.Kind, 2
		case value.KindInt32, value.KindUint32, value.KindFloat32:
			return p.Kind, 4
		case value.KindInt64, value.KindUint64, value.KindFloat64:
			return p.Kind, 8
		}
	}
	g.Panic("%s needs the type of one number kind, got %s", who, value.Render(v))
	return 0, 0
}

// decode reads a number of kind k from p, which holds exactly its width.
func decode(k value.Kind, p []byte, o binary.ByteOrder) value.Value {
	switch k {
	case value.KindUint8:
		return p[0]
	case value.KindInt8:
		return int8(p[0])
	case value.KindUint16:
		return o.Uint16(p)
	case value.KindInt16:
		return int16(o.Uint16(p))
	case value.KindUint32:
		return o.Uint32(p)
	case value.KindInt32:
		return int32(o.Uint32(p))
	case value.KindFloat32:
		return math.Float32frombits(o.Uint32(p))
	case value.KindUint64:
		return o.Uint64(p)
	case value.KindInt64:
		return int64(o.Uint64(p))
	}
	return math.Float64frombits(o.Uint64(p))
}

// encode writes x, a number of kind k, into p, which holds exactly its width.
func encode(g *interp.Goroutine, who string, k value.Kind, x value.Value, p []byte, o binary.ByteOrder) {
	if value.KindOf(x) != k {
		g.Panic("%s needs a %s, got %s", who, k, value.Describe(x))
	}
	switch v := x.(type) {
	case uint8:
		p[0] = v
	case int8:
		p[0] = byte(v)
	case uint16:
		o.PutUint16(p, v)
	case int16:
		o.PutUint16(p, uint16(v))
	case uint32:
		o.PutUint32(p, v)
	case int32:
		o.PutUint32(p, uint32(v))
	case float32:
		o.PutUint32(p, math.Float32bits(v))
	case uint64:
		o.PutUint64(p, v)
	case int64:
		o.PutUint64(p, uint64(v))
	case float64:
		o.PutUint64(p, math.Float64bits(v))
	}
}

// within panics unless at and at + size are within 0 ..= length.
func spanWithin(g *interp.Goroutine, who string, at int64, size, length int) {
	if at < 0 || at > int64(length) || int64(size) > int64(length)-at {
		g.Panic("%s: %d bytes at %d are not within 0 ..= %d", who, size, at, length)
	}
}

func init() {
	register("Bytes", func() value.Value {
		return namespace("Bytes",
			constant("BigEndian", value.NewTag("Bytes.BigEndian", value.RecordTag, nil)),
			constant("LittleEndian", value.NewTag("Bytes.LittleEndian", value.RecordTag, nil)),
			fn("Bytes.new", "size?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				n := optOffset(g, who, a[0])
				if n < 0 {
					g.Panic("%s needs a size of at least 0, got %d", who, n)
				}
				boundCount(g, who, "size", n)
				return bytesValue(make([]byte, n))
			}),
			fn("Bytes.clone", "b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return bytesValue(bytesArg(g, who, a[0]).Snapshot())
			}),
			fn("Bytes.same", "a, b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return bytesArg(g, who, a[0]) == bytesArg(g, who, a[1])
			}),
			fn("Bytes.length", "b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(bytesArg(g, who, a[0]).Len())
			}),
			fn("Bytes.push", "b, x", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				b := bytesArg(g, who, a[0])
				b.Push(byteArg(g, who, a[1]))
				return value.Void
			}),
			fn("Bytes.extend", "b, src", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				b, src := bytesArg(g, who, a[0]), bytesArg(g, who, a[1])
				b.Append(src.View())
				return value.Void
			}),
			fn("Bytes.pop", "b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if c, ok := bytesArg(g, who, a[0]).Pop(); ok {
					return c
				}
				return value.Null
			}),
			fn("Bytes.truncate", "b, n", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				b, n := bytesArg(g, who, a[0]), intArg(g, who, a[1])
				if n < 0 || n > int64(b.Len()) || !b.Truncate(int(n)) {
					g.Panic("%s needs a length within 0 ..= %d, got %d", who, b.Len(), n)
				}
				return value.Void
			}),
			fn("Bytes.fill", "b, x, from?, to?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				b, x := bytesArg(g, who, a[0]), byteArg(g, who, a[1])
				v := b.View()
				from, to := optOffset(g, who, a[2]), optTo(g, who, a[3], len(v))
				sliceBounds(g, who, from, to, len(v))
				for i := from; i < to; i++ {
					v[i] = x
				}
				return value.Void
			}),
			fn("Bytes.copy", "dst, at, src, from?, to?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				dst, at, src := bytesArg(g, who, a[0]), intArg(g, who, a[1]), bytesArg(g, who, a[2])
				d, s := dst.View(), src.View()
				from, to := optOffset(g, who, a[3]), optTo(g, who, a[4], len(s))
				sliceBounds(g, who, from, to, len(s))
				spanWithin(g, who, at, int(to-from), len(d))
				copy(d[at:], s[from:to]) // copy handles overlap as memmove does
				return value.Void
			}),
			fn("Bytes.slice", "b, from, to", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v := bytesArg(g, who, a[0]).View()
				from, to := intArg(g, who, a[1]), intArg(g, who, a[2])
				sliceBounds(g, who, from, to, len(v))
				return value.CopyBytes(v[from:to])
			}),
			fn("Bytes.concat", "parts", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				parts := arrayArg(g, who, a[0]).Values()
				var out []byte
				for _, v := range parts {
					out = append(out, bytesArg(g, who, v).View()...)
				}
				if out == nil {
					out = []byte{}
				}
				return bytesValue(out)
			}),
			fn("Bytes.indexOf", "b, sub, from?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v, sub := bytesArg(g, who, a[0]).View(), bytesArg(g, who, a[1]).View()
				from := optOffset(g, who, a[2])
				if from < 0 || from > int64(len(v)) {
					g.Panic("%s needs a from within 0 ..= %d, got %d", who, len(v), from)
				}
				if i := bytes.Index(v[from:], sub); i >= 0 {
					return from + int64(i)
				}
				return value.Null
			}),
			fn("Bytes.startsWith", "b, prefix, at?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v, p := bytesArg(g, who, a[0]).View(), bytesArg(g, who, a[1]).View()
				at := optOffset(g, who, a[2])
				if at < 0 || at > int64(len(v)) {
					g.Panic("%s needs an at within 0 ..= %d, got %d", who, len(v), at)
				}
				return bytes.HasPrefix(v[at:], p)
			}),
			fn("Bytes.endsWith", "b, suffix", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return bytes.HasSuffix(bytesArg(g, who, a[0]).View(), bytesArg(g, who, a[1]).View())
			}),
			fn("Bytes.fromArray", "xs", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				vals := arrayArg(g, who, a[0]).Values()
				out := make([]byte, len(vals))
				for i, v := range vals {
					if !value.KindOf(v).IsInteger() {
						g.Panic("%s needs an array of integers, got an element of kind %s", who, value.Describe(v))
					}
					b, err := value.Convert(value.KindUint8, v)
					if err != nil {
						g.Panic("%s needs integers from 0 to 255, got %s at index %d", who, value.Render(v), i)
					}
					out[i] = b.(uint8)
				}
				return bytesValue(out)
			}),
			fn("Bytes.toArray", "b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v := bytesArg(g, who, a[0]).View()
				out := make([]value.Value, len(v))
				for i, c := range v {
					out[i] = c
				}
				return value.NewArray(out)
			}),
			fn("Bytes.read", "b, t, at, order", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v := bytesArg(g, who, a[0]).View()
				k, size := numberKind(g, who, a[1])
				at := intArg(g, who, a[2])
				o := orderArg(g, who, a[3])
				spanWithin(g, who, at, size, len(v))
				return decode(k, v[at:at+int64(size)], o)
			}),
			fn("Bytes.write", "b, t, at, x, order", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v := bytesArg(g, who, a[0]).View()
				k, size := numberKind(g, who, a[1])
				at := intArg(g, who, a[2])
				o := orderArg(g, who, a[4])
				spanWithin(g, who, at, size, len(v))
				encode(g, who, k, a[3], v[at:at+int64(size)], o)
				return value.Void
			}),
			fn("Bytes.append", "b, t, x, order", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				b := bytesArg(g, who, a[0])
				k, size := numberKind(g, who, a[1])
				o := orderArg(g, who, a[3])
				var p [8]byte
				encode(g, who, k, a[2], p[:size], o)
				b.Append(p[:size])
				return value.Void
			}),
			fn("Bytes.sizeOf", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				_, size := numberKind(g, who, a[0])
				return int64(size)
			}),
		)
	})
}
