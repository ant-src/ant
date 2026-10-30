package repl

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/internal/token"
)

// completion is the state of completing the line being edited: the replies for the word at
// the cursor, the menu Tab cycles through, and what the ghost after the cursor shows.
type completion struct {
	query    string           // the line the replies are for
	at       int              // the cursor they are for, in runes
	start    int              // where the completed word starts in the line, in runes
	tail     []rune           // what follows the word, which a reply leaves in place
	word     string           // the word being completed
	member   bool             // the word follows a dot or a pipe, so an empty word has a ghost
	replies  []ant.Completion // what can be written, sorted
	menu     bool             // Tab cycles through the replies
	selected int              // the reply the menu inserted
}

// completeMaxRows is the most rows the menu takes.
const completeMaxRows = 10

// updateCompletion computes the replies as the line is typed, which the ghost and the list
// show: only at the end of the line, so that nothing is drawn over the text after the cursor.
// Tab completes anywhere (computeCompletion).
func (e *editor) updateCompletion() {
	if e.comp.menu {
		return // the menu keeps its replies, wherever the cursor is
	}
	if e.cur != len(e.buf) {
		e.comp = completion{}
		return
	}
	e.computeCompletion()
}

// computeCompletion computes the replies for the word at the cursor: the identifier before it
// is what a reply must begin with, and the rest of the identifier after it is replaced too.
func (e *editor) computeCompletion() {
	c := &e.comp
	line := string(e.buf)
	if e.complete == nil {
		*c = completion{}
		return
	}
	if c.menu || c.query == line && c.at == e.cur && c.replies != nil {
		return // the menu keeps its replies; the same line has the same replies
	}
	*c = completion{query: line, at: e.cur}
	off := len(string(e.buf[:e.cur]))
	src := e.pending + line
	start := off
	for start > 0 && token.IsIdentByte(line[start-1]) {
		start--
	}
	end := off
	for end < len(line) && token.IsIdentByte(line[end]) {
		end++
	}
	c.word = line[start:off]
	c.start = len([]rune(line[:start]))
	c.tail = []rune(line[end:])
	before := strings.TrimRight(line[:start], " \t")
	c.member = strings.HasSuffix(before, ".") || strings.HasSuffix(before, "|")
	var items []ant.Completion
	if e.pending == "" && isCommand(line) {
		var csrc string
		var coff int
		var ok bool
		items, csrc, coff, ok = commandCompletions(strings.TrimLeft(line, " "), off-(len(line)-len(strings.TrimLeft(line, " "))))
		if ok {
			items = e.complete(csrc, coff)
		}
	} else {
		items = e.complete(src, len(e.pending)+off)
	}
	seen := map[string]bool{}
	for _, it := range items {
		if i := strings.IndexByte(it.Insert, '.'); i >= 0 && it.Kind != "file" && it.Kind != "folder" {
			// a qualified name, as the functions offered after a pipe are: Array.map,
			// Array.filter; completed a word at a time, so the namespace comes first, and
			// the dot is typed
			ns := it.Insert[:i]
			it = ant.Completion{Label: ns, Insert: ns, Kind: "namespace"}
		}
		if !strings.HasPrefix(it.Insert, c.word) || seen[it.Insert] {
			continue
		}
		seen[it.Insert] = true
		c.replies = append(c.replies, it)
	}
	sort.SliceStable(c.replies, func(i, j int) bool {
		a, b := c.replies[i], c.replies[j]
		if a.Preferred != b.Preferred {
			return a.Preferred
		}
		if a.Preferred {
			return false // the parameters and the fields keep their order
		}
		if ra, rb := kindRank(a.Kind), kindRank(b.Kind); ra != rb {
			return ra < rb
		}
		return a.Label < b.Label
	})
	if c.replies == nil {
		c.replies = []ant.Completion{} // computed, and none
	}
}

// kindRank orders the replies: what the session declared and the names in scope first, then
// the prelude's namespaces and types, then the keywords.
func kindRank(kind string) int {
	switch kind {
	case "variable", "constant", "field":
		return 0
	case "function":
		return 1
	case "namespace", "module", "type":
		return 2
	case "file", "folder":
		return 3
	case "keyword":
		return 4
	}
	return 5
}

// ghost is the rest of the first reply extending the word, shown dimmed after the cursor: for
// an empty word only after a dot or a pipe, or when the position asks for the reply, as a
// call asks for a parameter.
func (e *editor) ghost() string {
	c := &e.comp
	if c.menu || e.cur != len(e.buf) || c.query != string(e.buf) || c.at != e.cur {
		return ""
	}
	for _, r := range c.replies {
		if len(r.Insert) <= len(c.word) {
			continue
		}
		if c.word == "" && !c.member && !r.Preferred {
			return ""
		}
		return r.Insert[len(c.word):]
	}
	return ""
}

// acceptGhost writes the ghost into the line. It reports whether there was one.
func (e *editor) acceptGhost() bool {
	e.updateCompletion()
	g := e.ghost()
	if g == "" {
		return false
	}
	e.insert([]rune(g))
	e.updateCompletion()
	return true
}

// completeKey handles Tab and Shift-Tab: it writes the only reply, or the common prefix of the
// replies, and otherwise opens the menu and cycles through them.
func (e *editor) completeKey(dir int) {
	c := &e.comp
	if !c.menu {
		e.computeCompletion()
	}
	n := len(c.replies)
	if n == 0 {
		return
	}
	if c.menu {
		c.selected = (c.selected + dir + n) % n
		e.replaceWord(c.replies[c.selected].Insert)
		return
	}
	if n == 1 {
		e.replaceWord(c.replies[0].Insert)
		*c = completion{}
		e.updateCompletion()
		return
	}
	if prefix := commonPrefix(c.replies); len(prefix) > len(c.word) {
		e.replaceWord(prefix)
		*c = completion{}
		e.updateCompletion()
		return
	}
	c.menu = true
	c.selected = 0
	if dir < 0 {
		c.selected = n - 1
	}
	e.replaceWord(c.replies[c.selected].Insert)
}

// replaceWord puts text in place of the word being completed, leaving what follows it, and
// the cursor after text. The query and the replies stay those of the word, so that the menu
// can go on cycling.
func (e *editor) replaceWord(text string) {
	c := &e.comp
	e.buf = append(append(e.buf[:c.start], []rune(text)...), c.tail...)
	e.cur = c.start + len([]rune(text))
}

// closeMenu ends menu cycling, keeping the reply written; the replies are then computed for
// the new line.
func (e *editor) closeMenu() {
	if e.comp.menu {
		e.comp = completion{}
	}
}

// commonPrefix is the longest prefix every reply shares, cut to whole runes.
func commonPrefix(replies []ant.Completion) string {
	prefix := replies[0].Insert
	for _, r := range replies[1:] {
		i := 0
		for i < len(prefix) && i < len(r.Insert) && prefix[i] == r.Insert[i] {
			i++
		}
		prefix = prefix[:i]
	}
	for !utf8.ValidString(prefix) {
		prefix = prefix[:len(prefix)-1]
	}
	return prefix
}

// showMenu reports whether the replies are listed under the line: when the menu cycles, or
// when there is more than one reply for a word being typed.
func (e *editor) showMenu() bool {
	c := &e.comp
	if c.menu {
		return true
	}
	return c.word != "" && len(c.replies) > 1
}

// menuLines renders the replies in columns under the line, at most completeMaxRows rows and a
// third of the screen, scrolled to the selected reply, with the reply's detail on a last line.
func (e *editor) menuLines(width, height int) []string {
	c := &e.comp
	if !e.showMenu() || width < 4 {
		return nil
	}
	maxWidth := 1
	for _, r := range c.replies {
		maxWidth = max(maxWidth, stringWidth(r.Label))
	}
	colWidth := min(maxWidth, width-2) + 2
	cols := max(1, (width-1+2)/colWidth)
	rows := (len(c.replies) + cols - 1) / cols
	maxRows := completeMaxRows
	if height > 0 {
		maxRows = max(1, min(completeMaxRows, (height-2)/3))
	}
	shown := min(rows, maxRows)
	first := 0
	if rows > shown && c.menu {
		first = max(0, min(c.selected/cols-shown/2, rows-shown))
	}
	var lines []string
	for row := first; row < first+shown; row++ {
		var line strings.Builder
		w := 0
		for col := 0; col < cols; col++ {
			i := row*cols + col
			if i >= len(c.replies) {
				break
			}
			text := truncate(c.replies[i].Label, colWidth-2, "…")
			if c.menu && i == c.selected {
				line.WriteString(e.th.selected + text + reset)
			} else {
				line.WriteString(text)
			}
			w += stringWidth(text)
			if col < cols-1 && i+1 < len(c.replies) {
				pad := colWidth - stringWidth(text)
				line.WriteString(strings.Repeat(" ", pad))
				w += pad
			}
		}
		lines = append(lines, line.String())
	}
	if rows > shown {
		lines = append(lines, e.th.dim(fmt.Sprintf("rows %d to %d of %d", first+1, first+shown, rows), width-1))
	}
	if c.menu {
		r := c.replies[c.selected]
		detail := r.Label
		if r.Detail != "" {
			sep := " "
			if strings.HasPrefix(r.Detail, "(") || strings.HasPrefix(r.Detail, "[") {
				sep = ""
			}
			detail += sep + r.Detail
		}
		lines = append(lines, e.th.dim(truncate(detail, width-1, "…"), width-1))
	}
	return lines
}

// dim renders s in the ghost's color, cut to width.
func (th *theme) dim(s string, width int) string {
	s = truncate(s, width, "…")
	if th.class[clsGhost] == "" {
		return s
	}
	return th.class[clsGhost] + s + reset
}
