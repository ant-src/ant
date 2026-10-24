package value

import (
	"strings"
	"unicode/utf8"
)

// Text returns s as a string Ant can hold: s itself when it is UTF-8, and otherwise s with each
// maximal subpart of an ill-formed sequence replaced by U+FFFD, the practice Unicode recommends
// in chapter 3 and the WHATWG Encoding standard follows (section 11, String). Every text that
// reaches a program from outside it passes through here.
func Text(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s) + 8)
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r != utf8.RuneError || size > 1 {
			sb.WriteString(s[i : i+size])
			i += size
			continue
		}
		sb.WriteRune(utf8.RuneError)
		i += maximalSubpart(s[i:])
	}
	return sb.String()
}

// maximalSubpart returns the length of the longest prefix of s, which starts with an ill-formed
// sequence, that is the start of some well-formed sequence, and at least 1: the bytes one U+FFFD
// replaces. The ranges are those of Unicode's table 3-7.
func maximalSubpart(s string) int {
	lead := s[0]
	var need int
	lo, hi := byte(0x80), byte(0xBF) // the range of the second byte
	switch {
	case lead >= 0xC2 && lead <= 0xDF:
		need = 1
	case lead == 0xE0:
		need, lo = 2, 0xA0
	case lead >= 0xE1 && lead <= 0xEC, lead == 0xEE, lead == 0xEF:
		need = 2
	case lead == 0xED:
		need, hi = 2, 0x9F
	case lead == 0xF0:
		need, lo = 3, 0x90
	case lead >= 0xF1 && lead <= 0xF3:
		need = 3
	case lead == 0xF4:
		need, hi = 3, 0x8F
	default:
		return 1
	}
	n := 1
	for ; n <= need && n < len(s); n++ {
		b := s[n]
		if n == 1 && (b < lo || b > hi) || n > 1 && (b < 0x80 || b > 0xBF) {
			break
		}
	}
	return n
}
