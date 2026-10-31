package repl

import (
	"strings"

	"github.com/ant-src/ant/internal/highlight"
)

// A class is what a byte of an input is, for its color: a class of the highlight package, or the
// ghost of a completion.
type class = highlight.Class

const (
	clsPlain    = highlight.Plain
	clsKeyword  = highlight.Keyword
	clsConstant = highlight.Constant
	clsString   = highlight.String
	clsNumber   = highlight.Number
	clsRegex    = highlight.Regex
	clsComment  = highlight.Comment
	clsType     = highlight.Type
	clsFunction = highlight.Function
	clsMarkup   = highlight.Markup
	clsBad      = highlight.Bad
	clsGhost    = highlight.Bad + 1 // the completion suggested after the cursor
)

// classes returns the class of every byte of src, which may be unfinished.
func classes(src string) []class { return highlight.Classes(src) }

// A theme is the escape sequence that starts each class, or "" for none, and the one that
// marks the reply the menu selected.
type theme struct {
	class    [clsGhost + 1]string
	selected string
}

// ansiTheme colors with the terminal's own sixteen colors, so that it suits a light or a dark
// background alike.
var ansiTheme = theme{
	class:    withGhost(highlight.ANSI, "\x1b[90m"), // bright black
	selected: "\x1b[7m",                             // inverse
}

// plainTheme colors nothing, for NO_COLOR; the ghost is still told from the text, and the
// selected reply from the others, by the attributes that are not colors.
var plainTheme = theme{
	class:    [clsGhost + 1]string{clsGhost: "\x1b[2m"}, // faint
	selected: "\x1b[7m",
}

func withGhost(colors [highlight.Bad + 1]string, ghost string) [clsGhost + 1]string {
	var out [clsGhost + 1]string
	copy(out[:], colors[:])
	out[clsGhost] = ghost
	return out
}

const reset = highlight.Reset

// paint writes s colored by the class of each of its bytes, which cls gives from byte base of
// the source it was computed for, resetting the color at the end of every run.
func (th *theme) paint(sb *strings.Builder, s string, cls []class, base int) {
	highlight.Paint(sb, s, cls, base, th.class[:])
}
