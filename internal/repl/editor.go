package repl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/ant-src/ant"
)

// errInterrupt is what readLine returns when Ctrl-C drops the line.
var errInterrupt = errors.New("interrupt")

// escapeWait is how long a lone escape waits for the rest of its sequence before it is the
// Escape key.
const escapeWait = 50 * time.Millisecond

// An editor reads a line from a terminal in raw mode, drawing it with its syntax highlighted
// and the completion after the cursor, and taking the editing keys of readline.
type editor struct {
	ctx      context.Context
	in       <-chan []byte // what the terminal sends, in chunks; closed at its end
	out      io.Writer
	size     func() (cols, rows int) // the terminal's size; cols 0 when unknown
	th       *theme
	complete func(src string, off int) []ant.Completion // nil completes nothing
	// incomplete reports whether an input needs more lines, so that Enter goes on to the next
	// line rather than ending it; nil ends it at every Enter
	incomplete func(src string) bool
	hist       *history
	raw        func() (restore func(), err error) // puts the terminal in raw mode; nil for a terminal already raw

	// the line being edited
	prompt  string
	pending string // the earlier lines of the input, for highlighting and completion
	buf     []rune
	cur     int // the cursor, in runes
	histIdx int
	draft   []rune // the line left to browse the history
	comp    completion

	// what the last draw left on the screen
	cursorRow int // the row of the drawing the cursor is on
	unread    []byte
	closed    bool

	// mu guards the terminal between the editor and output: a goroutine may print while the
	// prompt waits, which then is cleared, and drawn again under what was printed.
	mu        sync.Mutex
	prompting bool // the prompt and the line are drawn, and waiting for keys
	partial   bool // output last written did not end its line

	// bracketed turns on the terminal's bracketed paste mode while a line is read, in which
	// it marks pasted text; pasting is set between the marks.
	bracketed bool
	pasting   bool
}

// writer returns a writer to w, the terminal stdout or stderr is, that keeps the prompt whole:
// what is written while the prompt waits goes above it, which is then drawn again.
func (e *editor) writer(w io.Writer) io.Writer { return termWriter{e, w} }

type termWriter struct {
	e *editor
	w io.Writer
}

func (t termWriter) Write(p []byte) (int, error) {
	e := t.e
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(p) == 0 {
		return 0, nil
	}
	if !e.prompting {
		e.partial = p[len(p)-1] != '\n'
		return t.w.Write(p)
	}
	var sb strings.Builder
	if e.cursorRow > 0 {
		fmt.Fprintf(&sb, "\x1b[%dA", e.cursorRow)
	}
	sb.WriteString("\r\x1b[J")
	io.WriteString(t.w, sb.String())
	n, err := t.w.Write(p)
	if p[len(p)-1] != '\n' {
		io.WriteString(t.w, "\r\n") // the prompt goes on a line of its own
	}
	e.cursorRow = 0
	e.draw(false)
	return n, err
}

// readLine reads one line after prompt. It returns errInterrupt, with the line, when Ctrl-C
// drops it, io.EOF at Ctrl-D on an empty line or the end of the input, and the context's error
// when it ends; the line is on the screen in every case, with the cursor on the next line.
func (e *editor) readLine(prompt, pending string) (string, error) {
	if e.raw != nil {
		restore, err := e.raw()
		if err != nil {
			return "", err
		}
		defer restore()
	}
	e.mu.Lock()
	if e.partial {
		// what was printed last did not end its line: the prompt starts a new one
		io.WriteString(e.out, "\r\n")
		e.partial = false
	}
	e.prompting = true
	if e.bracketed {
		io.WriteString(e.out, "\x1b[?2004h")
		defer io.WriteString(e.out, "\x1b[?2004l")
	}
	e.pasting = false
	e.prompt, e.pending = prompt, pending
	e.buf, e.cur = e.buf[:0], 0
	e.histIdx = len(e.hist.lines)
	e.draft = nil
	e.comp = completion{}
	e.cursorRow = 0
	e.updateCompletion()
	e.draw(false)
	e.mu.Unlock()
	for {
		k, err := e.readKey()
		e.mu.Lock()
		line, err, done := e.key(k, err)
		e.mu.Unlock()
		if done {
			return line, err
		}
	}
}

// key handles one key, or the error reading it, under the lock. done reports that the line
// ended, with line and err what readLine returns.
func (e *editor) key(k key, err error) (line string, _ error, done bool) {
	if err != nil {
		e.finish()
		return "", err, true
	}
	if k.kind != keyTab && k.kind != keyShiftTab {
		e.closeMenu()
	}
	switch k.kind {
	case keyRune:
		if !e.pasting && (k.r == '}' || k.r == ')' || k.r == ']') {
			e.dedent()
		}
		e.insert([]rune{k.r})
	case keyEnter:
		switch {
		case e.pasting:
			e.insert([]rune{'\n'}) // pasted text keeps its lines, and its own indentation
		case e.incomplete != nil && !isCommand(string(e.buf)) && e.incomplete(e.pending+string(e.buf)+"\n"):
			e.newline()
		default:
			e.cur = len(e.buf)
			e.finish()
			line := string(e.buf)
			e.hist.add(line)
			return line, nil, true
		}
	case keyPasteStart:
		e.pasting = true
	case keyPasteEnd:
		e.pasting = false
	case keyTab:
		if e.pasting || e.moreKeys() {
			e.insert([]rune("  ")) // a tab in pasted text indents
		} else {
			e.completeKey(1)
		}
	case keyShiftTab:
		if !e.pasting && !e.moreKeys() {
			e.completeKey(-1)
		}
	case keyBackspace:
		if e.cur > 0 {
			from := e.prev(e.cur)
			e.buf = append(e.buf[:from], e.buf[e.cur:]...)
			e.cur = from
		}
	case keyDelete:
		e.deleteAt()
	case keyEOF:
		if len(e.buf) == 0 {
			e.finish()
			return "", io.EOF, true
		}
		e.deleteAt()
	case keyInterrupt:
		e.comp = completion{}
		e.draw(true)
		e.prompting = false
		io.WriteString(e.out, "^C\r\n")
		return string(e.buf), errInterrupt, true
	case keyLeft:
		if e.cur > 0 {
			e.cur = e.prev(e.cur)
		}
	case keyRight:
		if e.cur < len(e.buf) {
			e.cur = e.next(e.cur)
		} else {
			e.acceptGhost()
		}
	case keyHome:
		e.cur = e.lineStart(e.cur)
	case keyEnd:
		if e.cur == len(e.buf) {
			e.acceptGhost()
		}
		e.cur = e.lineEnd(e.cur)
	case keyWordLeft:
		e.cur = e.wordStart(e.cur)
	case keyWordRight:
		e.cur = e.wordEnd(e.cur)
	case keyDeleteWordLeft:
		from := e.wordStart(e.cur)
		e.buf = append(e.buf[:from], e.buf[e.cur:]...)
		e.cur = from
	case keyDeleteWordRight:
		to := e.wordEnd(e.cur)
		e.buf = append(e.buf[:e.cur], e.buf[to:]...)
	case keyKillToEnd:
		e.buf = append(e.buf[:e.cur], e.buf[e.lineEnd(e.cur):]...)
	case keyKillToStart:
		from := e.lineStart(e.cur)
		e.buf = append(e.buf[:from], e.buf[e.cur:]...)
		e.cur = from
	case keyUp:
		if e.lineStart(e.cur) == 0 {
			e.browse(-1)
		} else {
			e.moveLine(-1)
		}
	case keyDown:
		if e.lineEnd(e.cur) == len(e.buf) {
			e.browse(1)
		} else {
			e.moveLine(1)
		}
	case keyClear:
		io.WriteString(e.out, "\x1b[H\x1b[2J")
		e.cursorRow = 0
	case keyEscape, keyNone:
	}
	if e.pasting || e.moreKeys() {
		return "", nil, false // pasted text is drawn once, at its end
	}
	e.updateCompletion()
	e.draw(false)
	return "", nil, false

}

// moreKeys reports whether a whole key has arrived already, in what the terminal sent last:
// pasted text comes in chunks, each drawn once.
func (e *editor) moreKeys() bool {
	_, n := decodeKey(e.unread, false)
	return n > 0
}

// readKey reads the next key, waiting for the rest of an escape sequence for a moment.
func (e *editor) readKey() (key, error) {
	for {
		if len(e.unread) > 0 {
			k, n := decodeKey(e.unread, e.closed)
			if n > 0 {
				e.unread = e.unread[n:]
				return k, nil
			}
		} else if e.closed {
			return key{}, io.EOF
		}
		var wait <-chan time.Time
		if len(e.unread) > 0 {
			wait = time.After(escapeWait)
		}
		select {
		case b, ok := <-e.in:
			if !ok {
				e.closed = true
				continue
			}
			e.unread = append(e.unread, b...)
		case <-wait:
			k, n := decodeKey(e.unread, true)
			e.unread = e.unread[n:]
			return k, nil
		case <-e.ctx.Done():
			return key{}, e.ctx.Err()
		}
	}
}

// lineStart is where the line holding position i begins.
func (e *editor) lineStart(i int) int {
	for i > 0 && e.buf[i-1] != '\n' {
		i--
	}
	return i
}

// lineEnd is where the line holding position i ends, before its line break.
func (e *editor) lineEnd(i int) int {
	for i < len(e.buf) && e.buf[i] != '\n' {
		i++
	}
	return i
}

// moveLine moves the cursor to the line above or below, keeping its column where that line
// is long enough.
func (e *editor) moveLine(dir int) {
	start := e.lineStart(e.cur)
	col := e.cur - start
	if dir < 0 {
		start = e.lineStart(start - 1)
	} else {
		start = e.lineEnd(e.cur) + 1
	}
	e.cur = min(start+col, e.lineEnd(start))
}

// newline breaks the line at the cursor, indenting the new line as the one it breaks, and
// two spaces more after an opening brace, bracket, or parenthesis.
func (e *editor) newline() {
	start := e.lineStart(e.cur)
	indent := 0
	for start+indent < e.cur && (e.buf[start+indent] == ' ' || e.buf[start+indent] == '\t') {
		indent++
	}
	ind := append([]rune{'\n'}, e.buf[start:start+indent]...)
	before := strings.TrimRight(string(e.buf[start:e.cur]), " \t")
	if strings.HasSuffix(before, "{") || strings.HasSuffix(before, "(") || strings.HasSuffix(before, "[") {
		ind = append(ind, ' ', ' ')
	}
	e.insert(ind)
}

// dedent takes two spaces of indentation away before a closing brace, bracket, or parenthesis
// typed first on its line.
func (e *editor) dedent() {
	start := e.lineStart(e.cur)
	if e.cur-start < 2 || strings.TrimLeft(string(e.buf[start:e.cur]), " ") != "" {
		return
	}
	e.buf = append(e.buf[:e.cur-2], e.buf[e.cur:]...)
	e.cur -= 2
}

func (e *editor) insert(rs []rune) {
	e.buf = append(e.buf[:e.cur], append(append([]rune(nil), rs...), e.buf[e.cur:]...)...)
	e.cur += len(rs)
}

func (e *editor) deleteAt() {
	if e.cur < len(e.buf) {
		to := e.next(e.cur)
		e.buf = append(e.buf[:e.cur], e.buf[to:]...)
	}
}

// prev is the position before i, stepping over the marks that take no column, so that the
// cursor never lands inside a grapheme cluster.
func (e *editor) prev(i int) int {
	i--
	for i > 0 && runeWidth(e.buf[i]) == 0 {
		i--
	}
	return i
}

// next is the position after i, stepping over the marks that follow the rune at i.
func (e *editor) next(i int) int {
	i++
	for i < len(e.buf) && runeWidth(e.buf[i]) == 0 {
		i++
	}
	return i
}

func (e *editor) wordStart(i int) int {
	for i > 0 && !isWordRune(e.buf[i-1]) {
		i--
	}
	for i > 0 && isWordRune(e.buf[i-1]) {
		i--
	}
	return i
}

func (e *editor) wordEnd(i int) int {
	for i < len(e.buf) && !isWordRune(e.buf[i]) {
		i++
	}
	for i < len(e.buf) && isWordRune(e.buf[i]) {
		i++
	}
	return i
}

// browse shows the previous or the next line of the history, keeping the line typed as the
// draft after the last entry.
func (e *editor) browse(dir int) {
	n := len(e.hist.lines)
	idx := e.histIdx + dir
	if idx < 0 || idx > n {
		return
	}
	if e.histIdx == n {
		e.draft = append([]rune(nil), e.buf...)
	}
	e.histIdx = idx
	if idx == n {
		e.buf = append(e.buf[:0], e.draft...)
	} else {
		e.buf = append(e.buf[:0], []rune(e.hist.lines[idx])...)
	}
	e.cur = len(e.buf)
	e.comp = completion{}
}

// finish draws the line as entered, without the ghost and the menu, and moves to the next
// line.
func (e *editor) finish() {
	e.comp = completion{}
	e.draw(true)
	e.prompting = false
	io.WriteString(e.out, "\r\n")
}

// draw redraws the prompt, the line, the ghost, and the menu from the first row the last
// drawing took, and puts the cursor back. With final set it draws the line alone and leaves
// the cursor after it, for what ends the line.
func (e *editor) draw(final bool) {
	cols, rows := e.size()
	if cols <= 0 {
		cols = 80
	}
	var sb strings.Builder
	if e.cursorRow > 0 {
		fmt.Fprintf(&sb, "\x1b[%dA", e.cursorRow)
	}
	sb.WriteString("\r\x1b[J")
	text := string(e.buf)
	cls := classes(e.pending + text)
	ghost := ""
	if !final {
		ghost = e.ghost()
	}
	// each line of the input after its prompt, the first after the prompt and the others
	// after the continuation prompt; row is where the line starts
	lines := strings.Split(text, "\n")
	row, endRow, off, at := 0, 0, 0, 0
	cursorRow, cursorCol := 0, 0
	for i, ln := range lines {
		prompt := e.prompt
		if i > 0 {
			sb.WriteString("\r\n")
			prompt = contPrompt
		}
		sb.WriteString(prompt)
		e.th.paint(&sb, ln, cls, len(e.pending)+off)
		rs := []rune(ln)
		if e.cur >= at && e.cur <= at+len(rs) {
			w := stringWidth(prompt) + runesWidth(rs[:e.cur-at])
			cursorRow, cursorCol = row+w/cols, w%cols
		}
		w := stringWidth(prompt) + runesWidth(rs)
		if i < len(lines)-1 {
			// a row filled to its end leaves the cursor waiting at its last column, from
			// which the line break goes to the next row
			row += max(w-1, 0)/cols + 1
		} else {
			if ghost != "" {
				sb.WriteString(e.th.dim(ghost, stringWidth(ghost)))
				w += stringWidth(ghost)
			}
			endRow = row + max(w-1, 0)/cols
			if !final && w > 0 && w%cols == 0 {
				// the cursor goes on in the next row, where the line would go on
				sb.WriteString("\r\n")
				endRow++
			}
		}
		off += len(ln) + 1
		at += len(rs) + 1
	}
	if final {
		e.cursorRow = 0
		io.WriteString(e.out, sb.String())
		return
	}
	menu := e.menuLines(cols, rows)
	for _, l := range menu {
		sb.WriteString("\r\n")
		sb.WriteString(l)
	}
	sb.WriteString("\r")
	if up := endRow + len(menu) - cursorRow; up > 0 {
		fmt.Fprintf(&sb, "\x1b[%dA", up)
	}
	if cursorCol > 0 {
		fmt.Fprintf(&sb, "\x1b[%dC", cursorCol)
	}
	e.cursorRow = cursorRow
	io.WriteString(e.out, sb.String())
}

// contPrompt is the prompt of the lines of an input after its first.
const contPrompt = ". "
