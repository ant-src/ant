package lib

import (
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

func boolArg(g *interp.Goroutine, who string, v value.Value) bool {
	b, ok := v.(bool)
	if !ok {
		g.Panic("%s needs a bool, got %s", who, value.Describe(v))
	}
	return b
}

// base64Alphabet picks the standard or the URL-safe alphabet from an optional url argument.
func base64Alphabet(g *interp.Goroutine, who string, url value.Value) *base64.Encoding {
	if given(url) && boolArg(g, who, url) {
		return base64.RawURLEncoding
	}
	return base64.RawStdEncoding
}

func init() {
	register("Encoding", func() value.Value {
		return namespace("Encoding",
			fn("Encoding.toBase64", "bytes, url?", func(g *interp.Goroutine, a []value.Value) value.Value {
				enc := base64Alphabet(g, "Encoding.toBase64", a[1]).WithPadding(base64.StdPadding)
				return enc.EncodeToString(bytesArg(g, "Encoding.toBase64", a[0]))
			}),
			fn("Encoding.fromBase64", "s, url?", func(g *interp.Goroutine, a []value.Value) value.Value {
				// Line breaks are ignored and the padding may be present or absent: the text is
				// decoded without it, by the raw form of the alphabet.
				s := strings.NewReplacer("\r", "", "\n", "").Replace(stringArg(g, "Encoding.fromBase64", a[0]))
				s = strings.TrimRight(s, "=")
				data, err := base64Alphabet(g, "Encoding.fromBase64", a[1]).DecodeString(s)
				if err != nil {
					return errResult(err)
				}
				return value.Ok(bytesValue(data))
			}),
			fn("Encoding.toHex", "bytes", func(g *interp.Goroutine, a []value.Value) value.Value {
				return hex.EncodeToString(bytesArg(g, "Encoding.toHex", a[0]))
			}),
			fn("Encoding.fromHex", "s", func(g *interp.Goroutine, a []value.Value) value.Value {
				data, err := hex.DecodeString(stringArg(g, "Encoding.fromHex", a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(bytesValue(data))
			}),
		)
	})
}
