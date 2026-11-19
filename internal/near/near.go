// Package near finds the names close to a misspelled one, for the help of the checker and of
// ant doc.
package near

import (
	"slices"
	"strings"
	"unicode/utf8"
)

// Nearest is the candidates within a small edit distance of name, nearest first: one edit for a
// name of up to four characters, two for a longer one, where swapping two neighbors is one edit.
func Nearest(name string, candidates []string) []string {
	most := 1
	if utf8.RuneCountInString(name) > 4 {
		most = 2
	}
	type hit struct {
		name string
		d    int
	}
	var hits []hit
	seen := map[string]bool{}
	for _, cand := range candidates {
		if cand == name || seen[cand] {
			continue
		}
		seen[cand] = true
		if d := Distance(strings.ToLower(name), strings.ToLower(cand)); d <= most {
			hits = append(hits, hit{cand, d})
		}
	}
	slices.SortStableFunc(hits, func(a, b hit) int {
		if a.d != b.d {
			return a.d - b.d
		}
		return strings.Compare(a.name, b.name)
	})
	var out []string
	for k, h := range hits {
		if k == 3 {
			break
		}
		out = append(out, h.name)
	}
	return out
}

// Distance is the optimal string alignment distance of a and b: insertions, deletions,
// substitutions, and swaps of two neighbors, each one edit.
func Distance(a, b string) int {
	x, y := []rune(a), []rune(b)
	d := make([][]int, len(x)+1)
	for i := range d {
		d[i] = make([]int, len(y)+1)
		d[i][0] = i
	}
	for j := range y {
		d[0][j+1] = j + 1
	}
	for i := 1; i <= len(x); i++ {
		for j := 1; j <= len(y); j++ {
			cost := 1
			if x[i-1] == y[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && x[i-1] == y[j-2] && x[i-2] == y[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[len(x)][len(y)]
}
