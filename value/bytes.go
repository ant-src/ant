package value

import "strings"

// Bytes is an immutable sequence of bytes: Bytes of section 11. It is a Go string underneath, a
// distinct type from string so that the two kinds stay apart, and since neither can change, a
// slice of Bytes and a conversion between Bytes and UTF-8 text share their memory.
type Bytes string

// renderBytes writes b as String.from does: Bytes(89 50 4e 47), two lower-case hexadecimal
// digits per byte.
func renderBytes(sb *strings.Builder, b Bytes) {
	const digits = "0123456789abcdef"
	sb.Grow(len("Bytes()") + 3*len(b))
	sb.WriteString("Bytes(")
	for i := 0; i < len(b); i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteByte(digits[b[i]>>4])
		sb.WriteByte(digits[b[i]&0xf])
	}
	sb.WriteByte(')')
}
