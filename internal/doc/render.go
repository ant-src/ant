package doc

import (
	"fmt"
	"io"
	"strings"

	"ant-lang.org/internal/highlight"
)

// PrintSignature prints a signature as written, highlighted when Color is set.
func (p *Printer) PrintSignature(w io.Writer, sig string) { p.code(w, sig, "") }

// PrintIndented prints a doc indented by four spaces, preserving source line breaks and spacing.
// Indented code is highlighted when Color is set.
func (p *Printer) PrintIndented(w io.Writer, doc string) {
	lines := strings.Split(doc, "\n")
	for i := 0; i < len(lines); {
		switch {
		case indented(lines[i]):
			// Lex a whole example together: strings and comments may span lines. Blank
			// lines within it keep their place, without losing the lexer's context.
			j := i + 1
			for j < len(lines) && (indented(lines[j]) || lines[j] == "") {
				j++
			}
			p.code(w, strings.Join(lines[i:j], "\n"), "    ")
			i = j
		case lines[i] == "":
			fmt.Fprintln(w)
			i++
		default:
			fmt.Fprintln(w, "    "+lines[i])
			i++
		}
	}
}

func indented(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}

// code adds indent to non-empty lines without changing their contents. Classes are computed
// before adding indentation, and painting each line separately resets color at every newline.
func (p *Printer) code(w io.Writer, src, indent string) {
	var classes []highlight.Class
	if p.Color {
		classes = highlight.Classes(src)
	}
	off := 0
	for line := range strings.SplitSeq(src, "\n") {
		var out strings.Builder
		if line != "" {
			out.WriteString(indent)
		}
		if p.Color {
			highlight.Paint(&out, line, classes, off, highlight.ANSI[:])
		} else {
			out.WriteString(line)
		}
		fmt.Fprintln(w, out.String())
		off += len(line) + 1
	}
}
