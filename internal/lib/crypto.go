package lib

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"hash"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// digest is a Crypto function that hashes one byte array.
func digest(name string, h func() hash.Hash) *interp.Builtin {
	return fn(name, "bytes", func(g *interp.Goroutine, _ string, a []value.Value) value.Value {
		d := h()
		d.Write(view(bytesArg(g, name, a[0])))
		return bytesValue(d.Sum(nil))
	})
}

// mac is a Crypto function that computes an HMAC of a message under a key.
func mac(name string, h func() hash.Hash) *interp.Builtin {
	return fn(name, "key, message", func(g *interp.Goroutine, _ string, a []value.Value) value.Value {
		m := hmac.New(h, view(bytesArg(g, name, a[0])))
		m.Write(view(bytesArg(g, name, a[1])))
		return bytesValue(m.Sum(nil))
	})
}

func init() {
	register("Crypto", func() value.Value {
		return namespace("Crypto",
			digest("Crypto.sha256", sha256.New),
			digest("Crypto.sha512", sha512.New),
			digest("Crypto.sha1", sha1.New),
			digest("Crypto.md5", md5.New),
			mac("Crypto.hmacSha1", sha1.New),
			mac("Crypto.hmacSha256", sha256.New),
			mac("Crypto.hmacSha512", sha512.New),
			fn("Crypto.equal", "a, b", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return subtle.ConstantTimeCompare(view(bytesArg(g, who, a[0])), view(bytesArg(g, who, a[1]))) == 1
			}),
			fn("Crypto.randomBytes", "n", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				n := intArg(g, who, a[0])
				if n < 0 {
					g.Panic("Crypto.randomBytes needs a count of at least 0, got %d", n)
				}
				boundCount(g, who, "count", n)
				b := make([]byte, n)
				if _, err := rand.Read(b); err != nil {
					g.Panic("Crypto.randomBytes: %s", err)
				}
				return bytesValue(b)
			}),
			fn("Crypto.uuid", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var b [16]byte
				if _, err := rand.Read(b[:]); err != nil {
					g.Panic("Crypto.uuid: %s", err)
				}
				b[6] = b[6]&0x0f | 0x40 // version 4
				b[8] = b[8]&0x3f | 0x80 // variant 1
				h := hex.EncodeToString(b[:])
				return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
			}),
		)
	})
}
