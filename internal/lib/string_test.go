package lib

import (
	"testing"
	"unicode"
)

// TestInSomeScript compares the merged runs behind String.inScript(c, "Unknown") with every
// range of every table of unicode.Scripts, at every code point.
func TestInSomeScript(t *testing.T) {
	want := make([]bool, unicode.MaxRune+1)
	for _, table := range unicode.Scripts {
		for _, r := range table.R16 {
			for c := int(r.Lo); c <= int(r.Hi); c += int(r.Stride) {
				want[c] = true
			}
		}
		for _, r := range table.R32 {
			for c := int(r.Lo); c <= int(r.Hi); c += int(r.Stride) {
				want[c] = true
			}
		}
	}
	for r := range want {
		if got := inSomeScript(rune(r)); got != want[r] {
			t.Fatalf("inSomeScript(U+%04X) = %v, want %v", r, got, want[r])
		}
	}
}
