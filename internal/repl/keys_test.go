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
		{"\r\n", false, keyEnter, 0, 2},
		{"\n\r", false, keyEnter, 0, 1},
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
		{"\x1b[1;5C", false, keyWordRight, 0, 6}, // Ctrl-Right
		{"\x1b[1;3D", false, keyWordLeft, 0, 6},  // Alt-Left
		{"\x1b[1;2C", false, keyRight, 0, 6},     // Shift-Right is Right
		{"\x1b[1;6D", false, keyWordLeft, 0, 6},  // Shift-Ctrl-Left
		{"\x1b[3~", false, keyDelete, 0, 4},
		{"\x1b[1~", false, keyHome, 0, 4},
		{"\x1b[F", false, keyEnd, 0, 3},
		{"\x1b[Z", false, keyShiftTab, 0, 3},
		{"\x1b[5~", false, keyNone, 0, 4},
		{"\x1b[200~", false, keyPasteStart, 0, 6},
		{"\x1b[201~", false, keyPasteEnd, 0, 6},
		{"\x1bOH", false, keyHome, 0, 3},
		{"\x1bb", false, keyWordLeft, 0, 2},
		{"\x1bf", false, keyWordRight, 0, 2},
		{"\x1b\x7f", false, keyDeleteWordLeft, 0, 2},
		{"\x1bq", false, keyNone, 0, 2},    // Alt-Q is not bound: nothing, and no q
		{"\x1bé", false, keyNone, 0, 3},    // nor Alt with a longer rune
		{"\x1b\xc3", false, keyNone, 0, 0}, // whose rest has not arrived yet
		{"\x1b\x1b", false, keyEscape, 0, 1},
		{"\x1b\r", false, keyEscape, 0, 1}, // Escape, then Enter
	}
	for _, tc := range tests {
		k, n := decodeKey([]byte(tc.in), tc.final)
		if k.kind != tc.want || k.r != tc.r || n != tc.n {
			t.Errorf("decodeKey(%q, %v) = %v %q %d, want %v %q %d", tc.in, tc.final, k.kind, k.r, n, tc.want, tc.r, tc.n)
		}
	}
}
