package repl

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/internal/token"
)

// fakeTerminal runs an editor over scripted keys on a screen of a given size, with the plain
// theme so that the output can be read.
func fakeTerminal(t *testing.T, cols int, complete func(string, int) []ant.Completion) (*editor, *strings.Builder, chan []byte) {
	t.Helper()
	in := make(chan []byte, 64)
	var out strings.Builder
	e := &editor{
		ctx:      context.Background(),
		in:       in,
		out:      &out,
		size:     func() (int, int) { return cols, 24 },
		th:       &plainTheme,
		complete: complete,
		hist:     loadHistory(""),
	}
	return e, &out, in
}

// typeKeys sends keys as typed: a Tab or a Shift-Tab arrives on its own, the rest in one chunk,
// as pasted text does.
func typeKeys(in chan<- []byte, keys string) {
	for keys != "" {
		i := strings.IndexAny(keys, "\t")
		if j := strings.Index(keys, "\x1b[Z"); j >= 0 && (i < 0 || j < i) {
			if j > 0 {
				in <- []byte(keys[:j])
			}
			in <- []byte("\x1b[Z")
			keys = keys[j+3:]
			continue
		}
		if i < 0 {
			in <- []byte(keys)
			return
		}
		if i > 0 {
			in <- []byte(keys[:i])
		}
		in <- []byte("\t")
		keys = keys[i+1:]
	}
}

func TestEditorLine(t *testing.T) {
	e, out, in := fakeTerminal(t, 80, nil)
	in <- []byte("var x = 1\r")
	line, err := e.readLine("> ", "")
	if err != nil || line != "var x = 1" {
		t.Fatalf("line %q, err %v", line, err)
	}
	// the last drawing is the line as entered, followed by a line break
	if !strings.HasSuffix(out.String(), "\r\x1b[J> var x = 1\r\n") {
		t.Errorf("output %q", out.String())
	}
}

func TestEditorKeys(t *testing.T) {
	tests := []struct{ keys, want string }{
		{"abc\x1b[D\x1b[DX\r", "aXbc"},              // left twice, insert
		{"abc\x01X\x05Y\r", "XabcY"},                // home, end
		{"abc\x7f\x7fX\r", "aX"},                    // backspace
		{"abc\x01\x1b[3~\r", "bc"},                  // delete
		{"abc\x01\x04\r", "bc"},                     // Ctrl-D on a line deletes
		{"hello world\x17X\r", "hello X"},           // Ctrl-W
		{"hello world\x1bbX\r", "hello Xworld"},     // Alt-B
		{"hello world\x01\x1bfX\r", "helloX world"}, // Alt-F
		{"hello world\x1b[1;5DX\r", "hello Xworld"}, // Ctrl-Left
		{"hello world\x01\x0bX\r", "X"},             // Ctrl-K
		{"hello world\x15X\r", "X"},                 // Ctrl-U
		{"héllo\x1b[D\x1b[D\x7f\r", "hélo"},         // runes, not bytes
		{"áb\x1b[D\x1b[DX\r", "Xáb"},              // a combining mark moves with its base
		{"abc\x1b[1;3D\x1b[1~X\r", "Xabc"},          // Alt-Left, Home
		{"abc\x1b\x1b[D\r", "abc"},                  // a lone escape does nothing
		{"\x1bOHX\r", "X"},                          // SS3 home
		{"ab\x1b[200~c\r", "abc"},                   // an unknown CSI sequence is skipped
	}
	for _, tc := range tests {
		e, _, in := fakeTerminal(t, 80, nil)
		in <- []byte(tc.keys)
		line, err := e.readLine("> ", "")
		if err != nil || line != tc.want {
			t.Errorf("%q: line %q, err %v, want %q", tc.keys, line, err, tc.want)
		}
	}
}

func TestEditorEOFAndInterrupt(t *testing.T) {
	e, out, in := fakeTerminal(t, 80, nil)
	in <- []byte("\x04")
	if _, err := e.readLine("> ", ""); err != io.EOF {
		t.Errorf("Ctrl-D on an empty line: %v", err)
	}
	in <- []byte("abc\x03")
	if _, err := e.readLine("> ", ""); err != errInterrupt {
		t.Errorf("Ctrl-C: %v", err)
	}
	if !strings.HasSuffix(out.String(), "\r\x1b[J> abc^C\r\n") {
		t.Errorf("output %q", out.String())
	}
	close(in)
	in2 := make(chan []byte)
	close(in2)
	e.in = in2
	e.closed = false
	in3 := make(chan []byte, 1)
	in3 <- []byte("xy")
	close(in3)
	e.in = in3
	e.closed = false
	if line, err := e.readLine("> ", ""); err != io.EOF || line != "" {
		t.Errorf("input ending mid-line: %q, %v", line, err)
	}
}

func TestEditorHistory(t *testing.T) {
	e, _, in := fakeTerminal(t, 80, nil)
	for _, l := range []string{"first\r", "second\r", "second\r"} {
		in <- []byte(l)
		e.readLine("> ", "")
	}
	if len(e.hist.lines) != 2 {
		t.Fatalf("history %q", e.hist.lines)
	}
	in <- []byte("dra\x1b[A\x1b[A\x1b[A\x1b[B\x1b[B\x1b[B\r") // up past the start, back down to the draft
	if line, _ := e.readLine("> ", ""); line != "dra" {
		t.Errorf("draft: %q", line)
	}
	in <- []byte("\x1b[A\x1b[A\x1b[A!\r") // the draft entered is in the history too
	if line, _ := e.readLine("> ", ""); line != "first!" {
		t.Errorf("edited entry: %q", line)
	}
	in <- []byte("\x10\x0e\r") // Ctrl-P, Ctrl-N
	if line, _ := e.readLine("> ", ""); line != "" {
		t.Errorf("back to the draft: %q", line)
	}
}

// complete offers a few names, as the session would.
func complete(src string, off int) []ant.Completion {
	var items []ant.Completion
	before := src[:off]
	for len(before) > 0 && token.IsIdentByte(before[len(before)-1]) {
		before = before[:len(before)-1]
	}
	switch {
	case strings.HasSuffix(before, "| "):
		for _, n := range []string{"Array.map", "Array.filter", "Iter.map", "twice"} {
			items = append(items, ant.Completion{Label: n, Insert: n, Kind: "function"})
		}
	case strings.HasSuffix(before, "Array."):
		for _, n := range []string{"map", "filter", "flatMap", "first"} {
			items = append(items, ant.Completion{Label: n, Insert: n, Kind: "function", Detail: "(xs: Array[T]): Array[U]"})
		}
	case strings.HasSuffix(before, "twice("):
		items = append(items, ant.Completion{Label: "k:", Insert: "k: ", Kind: "field", Preferred: true})
		items = append(items, ant.Completion{Label: "println", Insert: "println", Kind: "function"})
	default:
		for _, n := range []string{"println", "print", "panic", "Array", "Map", "var", "twice"} {
			kind := "function"
			if n == "var" {
				kind = "keyword"
			} else if n == "Array" || n == "Map" {
				kind = "namespace"
			}
			items = append(items, ant.Completion{Label: n, Insert: n, Kind: kind})
		}
	}
	return items
}

func TestEditorCompletion(t *testing.T) {
	tests := []struct{ keys, want string }{
		{"tw\t\r", "twice"},                       // one reply: written
		{"pr\t\r", "print"},                       // the common prefix (print is a prefix of println)
		{"printl\t\r", "println"},                 // then the only reply
		{"print\t\t\r", "println"},                // or the menu, when the word is a reply itself
		{"Array.f\t\r", "Array.filter"},           // three replies share only the word: the menu opens on the first
		{"Array.f\t\t\r", "Array.first"},          // Tab cycles... (sorted: filter, first, flatMap)
		{"Array.f\t\t\t\r", "Array.flatMap"},      // ...through the replies
		{"Array.f\t\t\t\t\r", "Array.filter"},     // and around
		{"Array.f\t\x1b[Z\r", "Array.flatMap"},    // Shift-Tab goes back
		{"Array.f\t\x1b\r", "Array.filter"},       // Escape leaves the menu with the reply
		{"Array.\x1b[C\r", "Array.filter"},        // Right accepts the ghost after a dot
		{"Array.\x05\r", "Array.filter"},          // so does End
		{"Array.\r", "Array."},                    // Enter does not
		{"twice(\x1b[C\r", "twice(k: "},           // a parameter the call asks for is suggested
		{"pa\x1b[C\r", "panic"},                   // the ghost extends the word
		{"pa\x1b[D\x1b[C\r", "pa"},                // Right in the middle of the line moves the cursor
		{"\t\t\r", "print"},                       // Tab on an empty line opens the menu: functions, then namespaces, then keywords
		{"\x1b[Z\r", "var"},                       // Shift-Tab starts at the last reply
		{"xyz\t\r", "xyz"},                        // no reply
		{"xs | Arra\t\r", "xs | Array."},          // after a pipe the namespace is completed first
		{"xs | A\x1b[C\r", "xs | Array."},         // as the ghost suggests it
		{"xs | Array.f\t\r", "xs | Array.filter"}, // then its members
		{"xs | \t\t\r", "xs | Array."},            // the menu lists each namespace once, after the plain names
	}
	for _, tc := range tests {
		e, _, in := fakeTerminal(t, 80, complete)
		typeKeys(in, tc.keys)
		line, err := e.readLine("> ", "")
		if err != nil || line != tc.want {
			t.Errorf("%q: line %q, err %v, want %q", tc.keys, line, err, tc.want)
		}
	}
}

func TestEditorPaste(t *testing.T) {
	tests := []struct {
		keys  string
		lines int // the lines the keys enter; the last is checked
		want  string
	}{
		{"if x {\r\tprintln(x)\r", 2, "  println(x)"}, // a pasted tab indents instead of completing
		{"a\tb\x1b[Zc\r", 1, "a  bc"},                 // and a pasted Shift-Tab does nothing
		{"var x = 1\r\nx\r\n", 2, "x"},                // CRLF is one line break
	}
	for _, tc := range tests {
		e, _, in := fakeTerminal(t, 80, complete)
		in <- []byte(tc.keys)
		close(in) // so that a line too many ends with io.EOF rather than waiting
		var line string
		var err error
		for i := 0; i < tc.lines && err == nil; i++ {
			line, err = e.readLine("> ", "")
		}
		if err != nil || line != tc.want {
			t.Errorf("%q: line %q, err %v, want %q", tc.keys, line, err, tc.want)
		}
	}
}

func TestEditorGhostAndMenuDrawing(t *testing.T) {
	e, out, in := fakeTerminal(t, 40, complete)
	// each chunk is drawn once, after its last key, as pasted text is
	for _, chunk := range []string{"Array.", "f", "\t"} {
		in <- []byte(chunk)
	}
	go func() { in <- []byte("\r") }()
	e.readLine("> ", "")
	s := out.String()
	// after "Array." the ghost shows the first member after the cursor, and the cursor is put
	// back before it
	if !strings.Contains(s, "> Array.\x1b[2mfilter\x1b[0m\r\x1b[8C") {
		t.Errorf("ghost: %q", s)
	}
	// the menu lists the replies under the line, the selected one reversed, and the selected
	// reply's detail; then the cursor moves back up to the line
	if !strings.Contains(s, "> Array.filter\r\n\x1b[7mfilter\x1b[0m   first    flatMap\r\n\x1b[2mfilter(xs: Array[T]): Array[U]\x1b[0m\r\x1b[2A\x1b[14C") {
		t.Errorf("menu: %q", s)
	}
}

func TestEditorWrapping(t *testing.T) {
	e, out, in := fakeTerminal(t, 10, nil)
	in <- []byte("abcdefghijklmnop\x1b[D\x1b[D") // 18 cells with the prompt: two rows
	go func() { in <- []byte("\r") }()
	e.readLine("> ", "")
	s := out.String()
	// the cursor two cells before the end is on the second row, column 6; the next drawing
	// starts by moving up to the first row
	if !strings.Contains(s, "> abcdefghijklmnop\r\x1b[6C") || !strings.Contains(s, "\x1b[1A\r\x1b[J> abcdefghijklmnop") {
		t.Errorf("wrapped: %q", s)
	}
	// a line that fills its rows exactly goes on at the start of the next row
	e, out, in = fakeTerminal(t, 10, nil)
	in <- []byte("abcdefgh")
	in <- []byte("\r")
	e.readLine("> ", "")
	if !strings.HasSuffix(out.String(), "> abcdefgh\r\n\r\x1b[1A\r\x1b[J> abcdefgh\r\n") {
		t.Errorf("full row: %q", out.String())
	}
}

func TestEditorHighlights(t *testing.T) {
	e, out, in := fakeTerminal(t, 80, nil)
	e.th = &ansiTheme
	in <- []byte("var x = 1\r")
	e.readLine("> ", "")
	if !strings.Contains(out.String(), "> \x1b[35mvar\x1b[0m x = \x1b[33m1\x1b[0m\r\n") {
		t.Errorf("highlighted: %q", out.String())
	}
	// an earlier line gives the context: inside a block comment the line is a comment
	in <- []byte("still\r")
	e.readLine(". ", "/* open\n")
	if !strings.Contains(out.String(), ". \x1b[90mstill\x1b[0m\r\n") {
		t.Errorf("in a comment: %q", out.String())
	}
}
