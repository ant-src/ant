package repl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
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
	hist     *history
	raw      func() (restore func(), err error) // puts the terminal in raw mode; nil for a terminal already raw

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
	e.prompt, e.pending = prompt, pending
	e.buf, e.cur = e.buf[:0], 0
	e.histIdx = len(e.hist.lines)
	e.draft = nil
	e.comp = completion{}
	e.cursorRow = 0
	e.updateCompletion()
	e.draw(false)
	for {
		k, err := e.readKey()
		if err != nil {
			e.finish()
			return "", err
		}
		if k.kind != keyTab && k.kind != keyShiftTab {
			e.closeMenu()
		}
		switch k.kind {
		case keyRune:
			e.insert([]rune{k.r})
		case keyEnter:
			e.finish()
			line := string(e.buf)
			e.hist.add(line)
			return line, nil
		case keyTab:
			if e.moreKeys() {
				e.insert([]rune("  ")) // a tab in pasted text indents
			} else {
				e.completeKey(1)
			}
		case keyShiftTab:
			if !e.moreKeys() {
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
				return "", io.EOF
			}
			e.deleteAt()
		case keyInterrupt:
			e.comp = completion{}
			e.draw(true)
			io.WriteString(e.out, "^C\r\n")
			return string(e.buf), errInterrupt
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
			e.cur = 0
		case keyEnd:
			if e.cur == len(e.buf) {
				e.acceptGhost()
			}
			e.cur = len(e.buf)
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
			e.buf = e.buf[:e.cur]
		case keyKillToStart:
			e.buf = append([]rune(nil), e.buf[e.cur:]...)
			e.cur = 0
		case keyUp:
			e.browse(-1)
		case keyDown:
			e.browse(1)
		case keyClear:
			io.WriteString(e.out, "\x1b[H\x1b[2J")
			e.cursorRow = 0
		case keyEscape, keyNone:
		}
		if e.moreKeys() {
			continue // pasted text is drawn once, at its end
		}
		e.updateCompletion()
		e.draw(false)
	}
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
	line := string(e.buf)
	cls := classes(e.pending + line)
	sb.WriteString(e.prompt)
	e.th.paint(&sb, line, cls, len(e.pending))
	total := stringWidth(e.prompt) + runesWidth(e.buf)
	if !final {
		if g := e.ghost(); g != "" {
			sb.WriteString(e.th.dim(g, stringWidth(g)))
			total += stringWidth(g)
		}
	}
	// the row the last cell is on; after a row filled to its end the cursor waits at its last
	// column, so a line break takes it to the next row, where the line goes on
	if final {
		e.cursorRow = 0
		io.WriteString(e.out, sb.String())
		return
	}
	endRow := 0
	if total > 0 {
		endRow = (total - 1) / cols
		if total%cols == 0 {
			sb.WriteString("\r\n")
			endRow++
		}
	}
	menu := e.menuLines(cols, rows)
	for _, l := range menu {
		sb.WriteString("\r\n")
		sb.WriteString(l)
	}
	cursor := stringWidth(e.prompt) + runesWidth(e.buf[:e.cur])
	cursorRow, cursorCol := cursor/cols, cursor%cols
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
