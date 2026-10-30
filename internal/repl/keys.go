package repl

import "unicode/utf8"

// A key is one keystroke as the editor sees it: a rune typed, or a control key.
type key struct {
	kind keyKind
	r    rune // for keyRune
}

type keyKind uint8

const (
	keyNone keyKind = iota
	keyRune
	keyEnter
	keyTab
	keyShiftTab
	keyBackspace
	keyDelete
	keyLeft
	keyRight
	keyUp
	keyDown
	keyHome
	keyEnd
	keyWordLeft
	keyWordRight
	keyDeleteWordLeft
	keyDeleteWordRight
	keyKillToEnd   // Ctrl-K
	keyKillToStart // Ctrl-U
	keyClear       // Ctrl-L
	keyInterrupt   // Ctrl-C
	keyEOF         // Ctrl-D
	keyEscape
)

// decodeKey reads one key from the start of b and returns it with the bytes it took. It returns
// n == 0 when b holds the start of a sequence that is not complete yet: the caller waits for
// more bytes, or, after a pause, takes the escape alone (decodeKey with final true).
func decodeKey(b []byte, final bool) (k key, n int) {
	if len(b) == 0 {
		return key{}, 0
	}
	switch c := b[0]; {
	case c == 0x1b:
		return decodeEscape(b, final)
	case c == '\r' || c == '\n':
		return key{kind: keyEnter}, 1
	case c == '\t':
		return key{kind: keyTab}, 1
	case c == 0x7f || c == 0x08:
		return key{kind: keyBackspace}, 1
	case c < 0x20:
		return controlKey(c), 1
	}
	r, size := utf8.DecodeRune(b)
	if r == utf8.RuneError && size == 1 {
		if !utf8.FullRune(b) && !final {
			return key{}, 0 // the rest of the rune has not arrived yet
		}
		return key{kind: keyNone}, 1 // a byte that is not UTF-8 is dropped
	}
	return key{kind: keyRune, r: r}, size
}

func controlKey(c byte) key {
	switch c {
	case 0x01: // Ctrl-A
		return key{kind: keyHome}
	case 0x02: // Ctrl-B
		return key{kind: keyLeft}
	case 0x03:
		return key{kind: keyInterrupt}
	case 0x04:
		return key{kind: keyEOF}
	case 0x05: // Ctrl-E
		return key{kind: keyEnd}
	case 0x06: // Ctrl-F
		return key{kind: keyRight}
	case 0x0b:
		return key{kind: keyKillToEnd}
	case 0x0c:
		return key{kind: keyClear}
	case 0x0e: // Ctrl-N
		return key{kind: keyDown}
	case 0x10: // Ctrl-P
		return key{kind: keyUp}
	case 0x15:
		return key{kind: keyKillToStart}
	case 0x17: // Ctrl-W
		return key{kind: keyDeleteWordLeft}
	}
	return key{kind: keyNone}
}

// decodeEscape reads an escape sequence: an arrow or editing key as CSI (ESC [) or SS3 (ESC O),
// or a key with Alt held as ESC and the key.
func decodeEscape(b []byte, final bool) (key, int) {
	if len(b) == 1 {
		if final {
			return key{kind: keyEscape}, 1
		}
		return key{}, 0
	}
	switch b[1] {
	case '[':
		// CSI: parameters, then a final byte in 0x40..0x7e
		i := 2
		for i < len(b) && (b[i] >= '0' && b[i] <= '9' || b[i] == ';') {
			i++
		}
		if i >= len(b) {
			if final {
				return key{kind: keyEscape}, 1
			}
			return key{}, 0
		}
		params, fin := string(b[2:i]), b[i]
		n := i + 1
		mod := ""
		if j := indexByte(params, ';'); j >= 0 {
			params, mod = params[:j], params[j+1:]
		}
		switch fin {
		case 'A':
			return key{kind: keyUp}, n
		case 'B':
			return key{kind: keyDown}, n
		case 'C':
			if mod != "" && mod != "1" {
				return key{kind: keyWordRight}, n
			}
			return key{kind: keyRight}, n
		case 'D':
			if mod != "" && mod != "1" {
				return key{kind: keyWordLeft}, n
			}
			return key{kind: keyLeft}, n
		case 'H':
			return key{kind: keyHome}, n
		case 'F':
			return key{kind: keyEnd}, n
		case 'Z':
			return key{kind: keyShiftTab}, n
		case '~':
			switch params {
			case "1", "7":
				return key{kind: keyHome}, n
			case "4", "8":
				return key{kind: keyEnd}, n
			case "3":
				return key{kind: keyDelete}, n
			}
		}
		return key{kind: keyNone}, n
	case 'O':
		if len(b) < 3 {
			if final {
				return key{kind: keyEscape}, 1
			}
			return key{}, 0
		}
		switch b[2] {
		case 'A':
			return key{kind: keyUp}, 3
		case 'B':
			return key{kind: keyDown}, 3
		case 'C':
			return key{kind: keyRight}, 3
		case 'D':
			return key{kind: keyLeft}, 3
		case 'H':
			return key{kind: keyHome}, 3
		case 'F':
			return key{kind: keyEnd}, 3
		}
		return key{kind: keyNone}, 3
	case 'b':
		return key{kind: keyWordLeft}, 2
	case 'f':
		return key{kind: keyWordRight}, 2
	case 'd':
		return key{kind: keyDeleteWordRight}, 2
	case 0x7f, 0x08:
		return key{kind: keyDeleteWordLeft}, 2
	case 0x1b:
		return key{kind: keyEscape}, 1
	}
	// Alt with another key: the key alone
	return key{kind: keyEscape}, 1
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
