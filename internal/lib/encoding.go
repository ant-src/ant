package lib

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"ant-lang.org/internal/interp"
	"ant-lang.org/value"
)

func boolArg(g *interp.Goroutine, who string, v value.Value) bool {
	return typedArg[bool](g, who, "a bool", v)
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
			fn("Encoding.toBase64", "bytes, url?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				enc := base64Alphabet(g, who, a[1]).WithPadding(base64.StdPadding)
				return enc.EncodeToString(view(bytesArg(g, who, a[0])))
			}),
			fn("Encoding.fromBase64", "s, url?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				// The padding may be present or absent: a text that ends in = is decoded with it,
				// any other without. Go skips line breaks and names offsets in the text as given.
				s := stringArg(g, who, a[0])
				enc := base64Alphabet(g, who, a[1])
				if strings.HasSuffix(strings.TrimRight(s, "\r\n"), "=") {
					enc = enc.WithPadding(base64.StdPadding)
				}
				data, err := enc.DecodeString(s)
				if err != nil {
					return errResult(err)
				}
				return value.Ok(bytesValue(data))
			}),
			fn("Encoding.toHex", "bytes", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return hex.EncodeToString(view(bytesArg(g, who, a[0])))
			}),
			fn("Encoding.fromHex", "s", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				text := stringArg(g, who, a[0])
				data, err := hex.DecodeString(text)
				if _, bad := err.(hex.InvalidByteError); bad {
					// Go names the offending byte; name the character it begins instead, so a
					// non-ASCII character reads as itself.
					for _, r := range text {
						if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
							err = fmt.Errorf("invalid byte: %#U", r)
							break
						}
					}
				}
				if err != nil {
					// Go's messages start with its package's name, which means nothing here.
					return errResult(errors.New(strings.TrimPrefix(err.Error(), "encoding/hex: ")))
				}
				return value.Ok(bytesValue(data))
			}),
		)
	})
}
