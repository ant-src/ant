package lib

import (
	"strings"
	"unsafe"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

func bytesArg(g *interp.Goroutine, who string, v value.Value) value.Bytes {
	return typedArg[value.Bytes](g, who, "Bytes", v)
}

// view is b as a byte slice without copying it, for a callee that only reads it: a hash, an
// encoder, a write.
func view(b value.Bytes) []byte {
	return unsafe.Slice(unsafe.StringData(string(b)), len(b))
}

// bytesValue makes Bytes of b, sharing b's array when it is nearly full and copying when more
// than an eighth of it would be held unused, as io.ReadAll and append leave it. b must not change
// afterwards: pass only a slice nothing else holds, such as what a read or a hash returned.
func bytesValue(b []byte) value.Bytes {
	if cap(b)-len(b) > len(b)/8 {
		return value.Bytes(b)
	}
	return value.Bytes(unsafe.String(unsafe.SliceData(b), len(b)))
}

// textOrBytes reads a string | Bytes argument as the bytes to send: a string as its UTF-8.
func textOrBytes(g *interp.Goroutine, who string, v value.Value) string {
	switch x := v.(type) {
	case string:
		return x
	case value.Bytes:
		return string(x)
	}
	g.Panic("%s needs a string or Bytes, got %s", who, value.Describe(v))
	return ""
}

// received is data that arrived from outside as text and as Bytes (stdlib.md, Http): one buffer
// when it is UTF-8, which value.Text then returns as it is.
func received(b []byte) (text, raw value.Value) {
	bs := bytesValue(b)
	return value.Text(string(bs)), bs
}

func init() {
	register("Bytes", func() value.Value {
		return namespace("Bytes",
			fn("Bytes.length", "b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(len(bytesArg(g, who, a[0])))
			}),
			fn("Bytes.slice", "b, from, to", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				b := bytesArg(g, who, a[0])
				from, to := intArg(g, who, a[1]), intArg(g, who, a[2])
				sliceBounds(g, who, from, to, len(b))
				return b[from:to]
			}),
			fn("Bytes.concat", "parts", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				parts := arrayArg(g, who, a[0]).Values()
				size := 0
				for _, v := range parts {
					size += len(bytesArg(g, who, v))
				}
				var sb strings.Builder
				sb.Grow(size)
				for _, v := range parts {
					sb.WriteString(string(v.(value.Bytes)))
				}
				return value.Bytes(sb.String())
			}),
			fn("Bytes.indexOf", "b, sub", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if i := strings.Index(string(bytesArg(g, who, a[0])), string(bytesArg(g, who, a[1]))); i >= 0 {
					return int64(i)
				}
				return value.Null
			}),
			fn("Bytes.startsWith", "b, prefix", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.HasPrefix(string(bytesArg(g, who, a[0])), string(bytesArg(g, who, a[1])))
			}),
			fn("Bytes.endsWith", "b, suffix", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return strings.HasSuffix(string(bytesArg(g, who, a[0])), string(bytesArg(g, who, a[1])))
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
				b := bytesArg(g, who, a[0])
				out := make([]value.Value, len(b))
				for i := 0; i < len(b); i++ {
					out[i] = b[i]
				}
				return value.NewArray(out)
			}),
		)
	})
}
