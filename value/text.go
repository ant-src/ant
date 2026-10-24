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

// TextIn returns v with Text applied to every string in it: v itself, the elements of arrays, the
// keys and values of maps, the members of sets, the fields of records, and the payloads of tagged
// primitives. Containers change in place, keeping their order; a map or a set whose keys become
// equal keeps the last. It is how a value a Go host hands to a program becomes one that holds only
// scalar values.
func TextIn(v Value) Value {
	out, _ := textIn(v, map[any]bool{})
	return out
}

// textIn is TextIn, reporting whether anything changed; seen holds the containers visited, so a
// cyclic value terminates.
func textIn(v Value, seen map[any]bool) (Value, bool) {
	switch x := v.(type) {
	case string:
		if utf8.ValidString(x) {
			return x, false
		}
		return Text(x), true
	case *Tagged:
		if p, changed := textIn(x.Payload, seen); changed {
			return &Tagged{Tag: x.Tag, Payload: p}, true
		}
	case *Array:
		if seen[x] {
			return x, false
		}
		seen[x] = true
		for i, e := range x.Values() {
			if n, changed := textIn(e, seen); changed {
				x.Set(i, n)
			}
		}
	case *Record:
		if seen[x] {
			return x, false
		}
		seen[x] = true
		for i := 0; i < x.Len(); i++ {
			if n, changed := textIn(x.Get(i), seen); changed {
				x.Set(i, n)
			}
		}
	case *Map:
		if seen[x] {
			return x, false
		}
		seen[x] = true
		entries := x.Entries()
		dirty := false
		for i, e := range entries {
			k, kc := textIn(e.Key, seen)
			val, vc := textIn(e.Value, seen)
			entries[i] = KeyValue{Key: k, Value: val}
			dirty = dirty || kc || vc
		}
		if dirty {
			for _, k := range x.Keys() {
				x.Remove(k)
			}
			for _, e := range entries {
				x.Set(e.Key, e.Value)
			}
		}
	case *Set:
		if seen[x] {
			return x, false
		}
		seen[x] = true
		members := x.Values()
		dirty := false
		for i, m := range members {
			n, changed := textIn(m, seen)
			members[i] = n
			dirty = dirty || changed
		}
		if dirty {
			for _, m := range x.Values() {
				x.Remove(m)
			}
			for _, m := range members {
				x.Add(m)
			}
		}
	}
	return v, false
}
