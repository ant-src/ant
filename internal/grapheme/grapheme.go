// Package grapheme splits text into extended grapheme clusters by the rules of Unicode's UAX #29,
// for String.graphemes (section 11, String). The tables are generated from the Unicode Character
// Database by gen.go, at the version Version names, which must be the version of Go's unicode
// package and of golang.org/x/text; a test holds the three together.
package grapheme

//go:generate go run gen.go -version 17.0.0

import (
	"sort"
	"unicode/utf8"
)

// span is a run of code points with the same properties: the Grapheme_Cluster_Break value in the
// low four bits, Extended_Pictographic in bit 4, and the Indic_Conjunct_Break value in bits 5 and 6.
type span struct {
	lo, hi rune
	props  uint8
}

// The Grapheme_Cluster_Break values; zero is Other.
const (
	other uint8 = iota
	cr
	lf
	control
	extend
	zwj
	regional
	prepend
	spacingMark
	hangulL
	hangulV
	hangulT
	hangulLV
	hangulLVT
)

// The Indic_Conjunct_Break values; zero is None.
const (
	incbConsonant uint8 = 1 + iota
	incbLinker
	incbExtend
)

type props uint8

func (p props) breakValue() uint8   { return uint8(p) & 0x0F }
func (p props) pictographic() bool  { return p&0x10 != 0 }
func (p props) conjunct() uint8     { return uint8(p) >> 5 & 0x03 }
func (p props) isControlLike() bool { v := p.breakValue(); return v == cr || v == lf || v == control }

// lookup returns the properties of r. A byte that is not UTF-8 decodes as utf8.RuneError,
// U+FFFD, which is Other, so a cluster always makes progress.
func lookup(r rune) props {
	if r < 0x7F && r >= 0x20 {
		return 0
	}
	i := sort.Search(len(ranges), func(i int) bool { return ranges[i].hi >= r })
	if i < len(ranges) && ranges[i].lo <= r {
		return props(ranges[i].props)
	}
	return 0
}

// traceRule, when a test sets it, is told the rule that decided each boundary.
var traceRule func(rule string)

// Split returns the extended grapheme clusters of s, in order; they concatenate to s.
func Split(s string) []string {
	var out []string
	for s != "" {
		n := First(s)
		out = append(out, s[:n])
		s = s[n:]
	}
	return out
}

// First returns the length in bytes of the first extended grapheme cluster of s, which is at
// least one byte unless s is empty.
func First(s string) int {
	if s == "" {
		return 0
	}
	r, size := utf8.DecodeRuneInString(s)
	prev := lookup(r)

	// The state the rules that look back further than one code point need:
	// GB11: 1 after Extended_Pictographic Extend*, 2 after that and a ZWJ.
	// GB12 and GB13: how many Regional_Indicator code points end the text so far.
	// GB9c: 1 after an InCB Consonant followed by InCB Extend or Linker code points, none of
	// them a Linker; 2 when one of them is a Linker.
	emoji, ri, conjunct := 0, 0, 0
	update := func(p props) {
		switch {
		case p.pictographic():
			emoji = 1
		case emoji == 1 && p.breakValue() == extend:
		case emoji == 1 && p.breakValue() == zwj:
			emoji = 2
		default:
			emoji = 0
		}
		if p.breakValue() == regional {
			ri++
		} else {
			ri = 0
		}
		switch c := p.conjunct(); {
		case c == incbConsonant:
			conjunct = 1
		case conjunct != 0 && c == incbLinker:
			conjunct = 2
		case conjunct != 0 && c == incbExtend:
		default:
			conjunct = 0
		}
	}
	update(prev)

	for size < len(s) {
		r, n := utf8.DecodeRuneInString(s[size:])
		cur := lookup(r)
		brk, rule := breaksBetween(prev, cur, emoji, ri, conjunct)
		if traceRule != nil {
			traceRule(rule)
		}
		if brk {
			break
		}
		update(cur)
		prev = cur
		size += n
	}
	return size
}

// breaksBetween applies the rules GB3 to GB999 to the boundary between prev and cur, in their
// order of precedence, and returns whether there is a break and the rule that decided it.
func breaksBetween(prev, cur props, emoji, ri, conjunct int) (bool, string) {
	a, b := prev.breakValue(), cur.breakValue()
	switch {
	case a == cr && b == lf:
		return false, "GB3"
	case prev.isControlLike():
		return true, "GB4"
	case cur.isControlLike():
		return true, "GB5"
	case a == hangulL && (b == hangulL || b == hangulV || b == hangulLV || b == hangulLVT):
		return false, "GB6"
	case (a == hangulLV || a == hangulV) && (b == hangulV || b == hangulT):
		return false, "GB7"
	case (a == hangulLVT || a == hangulT) && b == hangulT:
		return false, "GB8"
	case b == extend || b == zwj:
		return false, "GB9"
	case b == spacingMark:
		return false, "GB9a"
	case a == prepend:
		return false, "GB9b"
	case conjunct == 2 && cur.conjunct() == incbConsonant:
		return false, "GB9c"
	case emoji == 2 && a == zwj && cur.pictographic():
		return false, "GB11"
	case a == regional && b == regional:
		return ri%2 == 0, "GB12"
	}
	return true, "GB999"
}
