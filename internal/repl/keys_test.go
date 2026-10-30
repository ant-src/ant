package repl

import "testing"

func TestDecodeKey(t *testing.T) {
	tests := []struct {
		in    string
		final bool
		want  keyKind
		r     rune
		n     int
	}{
		{"a", false, keyRune, 'a', 1},
		{"é", false, keyRune, 'é', 2},
		{"\xc3", false, keyNone, 0, 0}, // half a rune: wait
		{"\xc3", true, keyNone, 0, 1},  // or drop it
		{"\r", false, keyEnter, 0, 1},
		{"\n", false, keyEnter, 0, 1},
		{"\t", false, keyTab, 0, 1},
		{"\x7f", false, keyBackspace, 0, 1},
		{"\x01", false, keyHome, 0, 1},
		{"\x03", false, keyInterrupt, 0, 1},
		{"\x04", false, keyEOF, 0, 1},
		{"\x1b", false, keyNone, 0, 0},
		{"\x1b", true, keyEscape, 0, 1},
		{"\x1b[", false, keyNone, 0, 0},
		{"\x1b[A", false, keyUp, 0, 3},
		{"\x1b[Dx", false, keyLeft, 0, 3},
		{"\x1b[1;5C", false, keyWordRight, 0, 6},
		{"\x1b[1;3D", false, keyWordLeft, 0, 6},
		{"\x1b[3~", false, keyDelete, 0, 4},
		{"\x1b[1~", false, keyHome, 0, 4},
		{"\x1b[F", false, keyEnd, 0, 3},
		{"\x1b[Z", false, keyShiftTab, 0, 3},
		{"\x1b[5~", false, keyNone, 0, 4},
		{"\x1bOH", false, keyHome, 0, 3},
		{"\x1bb", false, keyWordLeft, 0, 2},
		{"\x1bf", false, keyWordRight, 0, 2},
		{"\x1b\x7f", false, keyDeleteWordLeft, 0, 2},
		{"\x1bq", false, keyEscape, 0, 1},
	}
	for _, tc := range tests {
		k, n := decodeKey([]byte(tc.in), tc.final)
		if k.kind != tc.want || k.r != tc.r || n != tc.n {
			t.Errorf("decodeKey(%q, %v) = %v %q %d, want %v %q %d", tc.in, tc.final, k.kind, k.r, n, tc.want, tc.r, tc.n)
		}
	}
}
