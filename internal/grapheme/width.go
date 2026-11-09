package grapheme

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/width"
)

// Width is the number of columns s takes in a monospace font: two for each grapheme cluster that
// starts with an East Asian Wide or Fullwidth character, is a flag (a pair of regional
// indicators), or holds the emoji presentation selector U+FE0F, and one for every other cluster.
// So a combining mark adds nothing, and an emoji built with joiners counts once.
func Width(s string) int {
	n := 0
	for i := 0; i < len(s); {
		k := max(First(s[i:]), 1)
		n += ClusterWidth(s[i : i+k])
		i += k
	}
	return n
}

// ClusterWidth is the columns one grapheme cluster takes, as Width counts them.
func ClusterWidth(c string) int {
	r, size := utf8.DecodeRuneInString(c)
	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	}
	if lookup(r).breakValue() == regional && len(c) > size {
		return 2 // a flag
	}
	if strings.ContainsRune(c, '️') {
		return 2
	}
	return 1
}
