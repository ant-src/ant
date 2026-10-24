package lib

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
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
				// Line breaks are ignored and the padding may be present or absent: the text is
				// decoded without it, by the raw form of the alphabet.
				s := strings.NewReplacer("\r", "", "\n", "").Replace(stringArg(g, who, a[0]))
				trimmed := strings.TrimRight(s, "=")
				if pad := len(s) - len(trimmed); pad > 2 || pad > 0 && len(s)%4 != 0 {
					return errResult(errors.New("illegal base64 padding"))
				}
				data, err := base64Alphabet(g, who, a[1]).DecodeString(trimmed)
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
							err = fmt.Errorf("encoding/hex: invalid byte: %#U", r)
							break
						}
					}
				}
				if err != nil {
					return errResult(err)
				}
				return value.Ok(bytesValue(data))
			}),
		)
	})
}
