package near

import (
	"slices"
	"testing"
)

// TestNearest: one edit for a name of up to four characters, counted in characters, not bytes, and
// two for a longer one; nearest first, then by name, at most three.
func TestNearest(t *testing.T) {
	cases := []struct {
		name       string
		candidates []string
		want       []string
	}{
		{"Größ", []string{"Grö", "Gr"}, []string{"Grö"}},
		{"parse", []string{"prase", "pars", "pa"}, []string{"pars", "prase"}},
		{"join", []string{"min", "joins"}, []string{"joins"}},
		{"abcde", []string{"abcdx", "abcdy", "abcdz", "abcdw"}, []string{"abcdw", "abcdx", "abcdy"}},
	}
	for _, c := range cases {
		if got := Nearest(c.name, c.candidates); !slices.Equal(got, c.want) {
			t.Errorf("Nearest(%q, %q) = %q, want %q", c.name, c.candidates, got, c.want)
		}
	}
}
