package repl

import "unicode"

// runeWidth is the columns a rune takes on the terminal: none for a combining mark or a joiner,
// two for an East Asian wide or an emoji presentation, one otherwise.
func runeWidth(r rune) int {
	switch {
	case r == 0 || r == 0x200b || r == 0x200d || r == 0xfe0f || r >= 0xfe00 && r <= 0xfe0f:
		return 0
	case r < 0x20 || r >= 0x7f && r < 0xa0:
		return 0
	case unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r):
		return 0
	case r >= 0x1100 && (r <= 0x115f ||
		r == 0x2329 || r == 0x232a ||
		r >= 0x2e80 && r <= 0xa4cf && r != 0x303f ||
		r >= 0xac00 && r <= 0xd7a3 ||
		r >= 0xf900 && r <= 0xfaff ||
		r >= 0xfe30 && r <= 0xfe4f ||
		r >= 0xff00 && r <= 0xff60 ||
		r >= 0xffe0 && r <= 0xffe6 ||
		r >= 0x1f300 && r <= 0x1f64f ||
		r >= 0x1f900 && r <= 0x1f9ff ||
		r >= 0x20000 && r <= 0x3fffd):
		return 2
	}
	return 1
}

// stringWidth is the columns s takes.
func stringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

// runesWidth is the columns rs take.
func runesWidth(rs []rune) int {
	w := 0
	for _, r := range rs {
		w += runeWidth(r)
	}
	return w
}

// truncate cuts s to at most width columns, ending it with tail when it was cut.
func truncate(s string, width int, tail string) string {
	if stringWidth(s) <= width {
		return s
	}
	room := width - stringWidth(tail)
	w := 0
	for i, r := range s {
		w += runeWidth(r)
		if w > room {
			return s[:i] + tail
		}
	}
	return s
}

// isWordRune says whether r belongs to a word, for the word movements: a letter, a digit, or
// an underscore.
func isWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
