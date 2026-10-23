// Package format lays out Ant source as spec/fmt.md describes: it reprints the tokens, comments,
// and line breaks of a module, changing only whitespace, statement semicolons, and trailing
// commas.
package format

import (
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/lexer"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/parser"
)

// Source formats one module. A source that does not parse returns the parser's error.
func Source(path string, src []byte) ([]byte, error) {
	s := string(src)
	_, l, err := parser.ParseLayout(path, s)
	if err != nil {
		return nil, err
	}
	p := &printer{src: s, roles: l.Roles, prefix: l.Prefix}
	p.comments = l.Comments
	file := p.build(l.Tokens, 0, len(s))
	p.file(file)
	out := strings.TrimRight(align(p.out.String(), p.trails), " \t\n")
	if out == "" {
		return nil, nil
	}
	return []byte(out + "\n"), nil
}

// elem is one thing in a sequence: a token, a comment, or a bracketed group.
type elem struct {
	tok *token.Token
	com *token.Comment
	grp *group
}

type group struct {
	open, close token.Token
	role        parser.Role
	elems       []elem // Newline tokens only inside a block
}

func (e elem) off() int {
	switch {
	case e.tok != nil:
		return e.tok.Off
	case e.com != nil:
		return e.com.Off
	}
	return e.grp.open.Off
}

func (e elem) end() int {
	switch {
	case e.tok != nil:
		return e.tok.End
	case e.com != nil:
		return e.com.End
	}
	return e.grp.close.End
}

// first and last are the tokens at the edges of a token or a group, for spacing.
func (e elem) first() token.Token {
	if e.grp != nil {
		return e.grp.open
	}
	return *e.tok
}

func (e elem) last() token.Token {
	if e.grp != nil {
		return e.grp.close
	}
	return *e.tok
}

func (e elem) is(text string) bool { return e.tok != nil && e.tok.Is(text) }

func (e elem) isNewline() bool { return e.tok != nil && e.tok.Kind == token.Newline }

type printer struct {
	src      string
	roles    map[int]parser.Role
	prefix   map[int]bool // the offsets of the - and ! the parser read as prefix
	comments []token.Comment
	out      strings.Builder
	indent   int   // the indentation level of the line being written
	pending  bool  // the line's indentation is not written yet
	lastEnd  int   // the source offset after the last thing printed
	trails   [][2]int // the output spans of the one-line comments written after code on a line
}

// ---- building the tree

// build merges the tokens and the comments between from and to into one sequence, nested by
// brackets. The top level behaves as a block.
func (p *printer) build(toks []token.Token, from, to int) *group {
	lo := sort.Search(len(p.comments), func(i int) bool { return p.comments[i].Off >= from })
	hi := sort.Search(len(p.comments), func(i int) bool { return p.comments[i].Off >= to })
	coms := p.comments[lo:hi]
	root := &group{role: parser.Role{Block: true}}
	stack := []*group{root}
	add := func(e elem) {
		g := stack[len(stack)-1]
		if e.isNewline() && !g.role.Block {
			return
		}
		g.elems = append(g.elems, e)
	}
	ci := 0
	for i := range toks {
		t := &toks[i]
		if t.Kind == token.EOF {
			break
		}
		for ci < len(coms) && coms[ci].Off < t.Off {
			add(elem{com: &coms[ci]})
			ci++
		}
		for ci < len(coms) && coms[ci].Off < t.End {
			ci++ // inside a template: printed with its interpolation
		}
		switch {
		case t.Is("(") || t.Is("[") || t.Is("{"):
			stack = append(stack, &group{open: *t, role: p.roles[t.Off]})
		case (t.Is(")") || t.Is("]") || t.Is("}")) && len(stack) > 1:
			g := stack[len(stack)-1]
			g.close = *t
			stack = stack[:len(stack)-1]
			add(elem{grp: g})
		default:
			add(elem{tok: t})
		}
	}
	for ; ci < len(coms); ci++ {
		add(elem{com: &coms[ci]})
	}
	return root
}

// ---- statements and items

// unit is a statement of a block or an item of a list: the comments on their own lines before
// it, its elements, and, for an item, the comments after its last token.
type unit struct {
	lead  []elem
	body  []elem
	trail []elem
}

func (u unit) empty() bool { return len(u.lead) == 0 && len(u.body) == 0 && len(u.trail) == 0 }

// start is the source offset of the unit's first comment or element.
func (u unit) start() int {
	if len(u.lead) > 0 {
		return u.lead[0].off()
	}
	if len(u.body) > 0 {
		return u.body[0].off()
	}
	return u.trail[0].off()
}

// ownLine reports whether only spaces and tabs precede the comment on its line.
func (p *printer) ownLine(c *token.Comment) bool {
	i := strings.LastIndexByte(p.src[:c.Off], '\n')
	return strings.Trim(p.src[i+1:c.Off], " \t\r") == ""
}

// split cuts a group's elements into units at the separators: newlines and semicolons in a
// block, commas in a list. A comment that starts a unit on the line where the unit before it
// ended belongs to that unit, or, at the start, after the opening bracket; in a block, only when
// no code follows it on the line, since a; /* c */ b leads b with it.
func (p *printer) split(g *group) (afterOpen []elem, units []unit) {
	var cur unit
	prevEnd := g.open.End // the end of the last element that is not a separator
	sep := func(e elem) bool {
		if g.role.Block {
			return e.isNewline() || e.is(";")
		}
		return e.is(",")
	}
	flush := func() {
		if !cur.empty() {
			units = append(units, cur)
		}
		cur = unit{}
	}
	// codeFollows reports whether the element after i, past any comments, is code on its line.
	codeFollows := func(i int) bool {
		for _, e := range g.elems[i+1:] {
			if e.com == nil {
				return !sep(e) && p.newlines(g.elems[i].end(), e.off()) == 0
			}
		}
		return false
	}
	for i, e := range g.elems {
		if sep(e) {
			flush()
			continue
		}
		sameLine := p.newlines(prevEnd, e.off()) == 0
		prevEnd = e.end()
		switch {
		case e.com != nil && len(cur.body) == 0:
			switch {
			case len(cur.lead) > 0 || p.ownLine(e.com) || !sameLine:
				cur.lead = append(cur.lead, e)
			case len(units) > 0 && g.role.Block && codeFollows(i):
				cur.lead = append(cur.lead, e)
			case len(units) > 0 && g.role.Block:
				units[len(units)-1].body = append(units[len(units)-1].body, e)
			case len(units) > 0:
				units[len(units)-1].trail = append(units[len(units)-1].trail, e)
			default:
				afterOpen = append(afterOpen, e)
			}
		case e.com != nil && !g.role.Block:
			cur.trail = append(cur.trail, e)
		default:
			if e.com == nil {
				cur.body = append(cur.body, cur.trail...)
				cur.trail = nil
			}
			cur.body = append(cur.body, e)
		}
	}
	flush()
	return afterOpen, units
}

// newlines counts the line breaks in the source between two offsets, which hold only
// whitespace and separators.
func (p *printer) newlines(from, to int) int {
	if from >= to {
		return 0
	}
	return strings.Count(p.src[from:to], "\n")
}

// broken reports whether the author broke the group across lines: for a block, any line break
// between its elements; for a list, one between its brackets and its items or between two items.
func (p *printer) broken(g *group, afterOpen []elem, units []unit) bool {
	var es []elem
	for _, e := range g.elems {
		if !e.isNewline() {
			es = append(es, e)
		}
	}
	if g.role.Block {
		prev := g.open.End
		for _, e := range es {
			if p.newlines(prev, e.off()) > 0 {
				return true
			}
			prev = e.end()
		}
		return p.newlines(prev, g.close.Off) > 0
	}
	prev := g.open.End
	for _, e := range afterOpen {
		prev = e.end()
	}
	for _, u := range units {
		if len(u.lead) > 0 || len(u.body) == 0 {
			return true
		}
		if p.newlines(prev, u.body[0].off()) > 0 {
			return true
		}
		prev = u.body[len(u.body)-1].end()
		for _, c := range u.trail {
			if p.newlines(prev, c.off()) > 0 {
				return true
			}
			prev = c.end()
		}
	}
	return p.newlines(prev, g.close.Off) > 0
}

// ---- output

func (p *printer) write(s string) {
	if p.pending {
		p.out.WriteString(strings.Repeat("  ", p.indent))
		p.pending = false
	}
	p.out.WriteString(s)
}

// newline ends the line and starts one at the given level, after one blank line when blank.
func (p *printer) newline(indent int, blank bool) {
	p.trim()
	p.out.WriteByte('\n')
	if blank {
		p.out.WriteByte('\n')
	}
	p.indent = indent
	p.pending = true
}

func (p *printer) trim() {
	s := p.out.String()
	if t := strings.TrimRight(s, " \t"); len(t) != len(s) {
		p.out.Reset()
		p.out.WriteString(t)
	}
}

// blankBefore reports whether the author left a blank line before the offset.
func (p *printer) blankBefore(off int) bool { return p.newlines(p.lastEnd, off) >= 2 }

func (p *printer) token(t token.Token) {
	switch {
	case t.Kind == token.Template:
		p.template(t)
	case t.Off == t.End:
		p.write(t.Text) // a separator the printer added
	default:
		p.write(p.src[t.Off:t.End])
	}
	p.lastEnd = t.End
}

// comment writes a comment's text with the spaces and tabs at the end of each line removed.
func (p *printer) comment(c *token.Comment) {
	lines := strings.Split(c.Text, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}
	out := p.out.String()
	afterCode := !p.pending && strings.TrimSpace(out[strings.LastIndexByte(out, '\n')+1:]) != ""
	start := p.out.Len()
	p.write(strings.Join(lines, "\n"))
	if afterCode && len(lines) == 1 {
		p.trails = append(p.trails, [2]int{start, p.out.Len()})
	}
	p.lastEnd = c.End
}

// trailing writes a comment that ends a line of code, one space after the code; align moves it
// to the column of its run afterwards.
func (p *printer) trailing(c *token.Comment) {
	p.write(" ")
	p.comment(c)
}

// align lines up the trailing comments of each run of consecutive lines that have the same
// indentation and each end in one: every comment of the run starts one space after the longest
// code in it, counted in code points (fmt.md 4.8). A run of one keeps its single space. Which
// comments end their lines is decided on the output, not the source, so that formatting the
// output aligns the same comments again.
func align(out string, trails [][2]int) string {
	if len(trails) == 0 {
		return out
	}
	lines := strings.SplitAfter(out, "\n")
	at := map[int]int{} // line index to the comment's byte offset in that line
	line, start := 0, 0
	for _, span := range trails {
		for start+len(lines[line]) <= span[0] {
			start += len(lines[line])
			line++
		}
		if strings.TrimSpace(lines[line][span[1]-start:]) == "" {
			at[line] = span[0] - start
		}
	}
	indent := func(s string) int { return len(s) - len(strings.TrimLeft(s, " ")) }
	width := func(i int) int { return utf8.RuneCountInString(strings.TrimRight(lines[i][:at[i]], " ")) }
	for i := 0; i < len(lines); {
		if _, ok := at[i]; !ok {
			i++
			continue
		}
		j, col := i, 0
		for ; j < len(lines); j++ {
			if _, ok := at[j]; !ok || indent(lines[j]) != indent(lines[i]) {
				break
			}
			col = max(col, width(j)+1)
		}
		for k := i; k < j; k++ {
			code := strings.TrimRight(lines[k][:at[k]], " ")
			lines[k] = code + strings.Repeat(" ", col-width(k)) + lines[k][at[k]:]
		}
		i = j
	}
	return strings.Join(lines, "")
}

// endsLine reports whether nothing but whitespace follows the comment on its source line.
func (p *printer) endsLine(c *token.Comment) bool {
	if strings.HasPrefix(c.Text, "//") {
		return true
	}
	rest := p.src[c.End:]
	if i := strings.IndexByte(rest, '\n'); i >= 0 {
		rest = rest[:i]
	}
	return strings.Trim(rest, " \t\r") == ""
}

// ---- layout

func (p *printer) file(root *group) {
	afterOpen, units := p.split(root)
	if len(afterOpen) > 0 {
		// A file has no opening bracket: a comment after a leading ; leads the first statement.
		units = append([]unit{{lead: afterOpen}}, units...)
	}
	for i, u := range units {
		if i > 0 {
			p.newline(0, p.blankBefore(u.start()))
		}
		p.unit(u, 0)
	}
}

// unit writes a statement or an item: its leading comments on their own lines, then its body,
// whose continuation lines are one level deeper than its first.
func (p *printer) unit(u unit, indent int) {
	for i, c := range u.lead {
		if i > 0 {
			p.breakBefore(c, indent)
		}
		p.comment(c.com)
	}
	if len(u.lead) > 0 && len(u.body) > 0 {
		p.breakBefore(u.body[0], indent)
	}
	p.seq(u.body, indent)
}

// breakBefore starts a new line before e when the author wrote one there, and otherwise
// leaves one space, as after a block comment that shares its line with code.
func (p *printer) breakBefore(e elem, indent int) {
	if p.newlines(p.lastEnd, e.off()) > 0 {
		p.newline(indent, p.blankBefore(e.off()))
	} else {
		p.write(" ")
	}
}

// seq writes a statement or an item's elements. A line break the author wrote between two of them
// is kept, and the new line is one level deeper than base, except that an else takes the level
// of the line that holds its if.
func (p *printer) seq(es []elem, base int) {
	ifIndent := base
	var prev *elem
	for i := range es {
		e := es[i]
		if prev != nil {
			switch {
			case p.newlines(prev.end(), e.off()) > 0:
				indent := base + 1
				if beforeElse(es[i:]) {
					indent = ifIndent
				}
				p.newline(indent, p.blankBefore(e.off()))
			case e.com != nil && p.endsLine(e.com):
				p.trailing(e.com)
				prev = &es[i]
				continue
			case e.com != nil || prev.com != nil:
				p.write(" ")
			default:
				p.write(strings.Repeat(" ", p.space(prev.last(), e.first())))
			}
		}
		switch {
		case e.com != nil:
			p.comment(e.com)
		case e.grp != nil:
			p.group(e.grp)
		default:
			if e.is("if") {
				ifIndent = p.indent
			}
			p.token(*e.tok)
		}
		prev = &es[i]
	}
}

// beforeElse reports whether es starts with an else, after any comments, which take its level.
func beforeElse(es []elem) bool {
	for len(es) > 0 && es[0].com != nil {
		es = es[1:]
	}
	return len(es) > 0 && es[0].is("else")
}

func (p *printer) group(g *group) {
	p.token(g.open)
	afterOpen, units := p.split(g)
	line := p.indent
	if !p.broken(g, afterOpen, units) {
		p.flat(g, afterOpen, units)
	} else {
		for i, c := range afterOpen {
			if !p.endsLine(c.com) && len(units) > 0 {
				// It shares its line with the first unit, which it leads.
				units[0].lead = append(slices.Clone(afterOpen[i:]), units[0].lead...)
				break
			}
			p.trailing(c.com)
		}
		last := -1
		for i, u := range units {
			if len(u.body) > 0 {
				last = i
			}
		}
		if g.role.Block {
			keepSemicolons(units)
		}
		for i, u := range units {
			p.newline(line+1, i > 0 && p.blankBefore(u.start()))
			p.unit(u, line+1)
			if !g.role.Block && len(u.body) > 0 && (i < last || g.role.Comma) {
				p.write(",")
			}
			for _, c := range u.trail {
				if p.newlines(p.lastEnd, c.off()) > 0 {
					p.newline(line+1, p.blankBefore(c.off()))
					p.comment(c.com)
				} else {
					p.trailing(c.com)
				}
			}
		}
		p.newline(line, false)
	}
	p.token(g.close)
}

// keepSemicolons ends each statement that comes before one starting with a regular expression
// literal with a ;, before its trailing comments. Without it the / would follow an operand
// across the line break and read as division (section 2). Units that hold only comments are
// skipped, since the ; must be the last token before the /.
func keepSemicolons(units []unit) {
	for i := 1; i < len(units); i++ {
		if b := units[i].body; len(b) == 0 || b[0].tok == nil || b[0].tok.Kind != token.Regex {
			continue
		}
		j, n := i-1, 0
		for ; j >= 0; j-- {
			for n = len(units[j].body); n > 0 && units[j].body[n-1].com != nil; n-- {
			}
			if n > 0 {
				break
			}
		}
		if j < 0 {
			continue
		}
		prev := &units[j]
		end := prev.body[n-1].end()
		semi := elem{tok: &token.Token{Kind: token.Op, Text: ";", Off: end, End: end}}
		prev.body = append(prev.body[:n:n], append([]elem{semi}, prev.body[n:]...)...)
	}
}

// flat writes a group on the line it opens on: items separated by ", ", statements by "; ", and
// one space inside a brace that holds anything.
func (p *printer) flat(g *group, afterOpen []elem, units []unit) {
	sep := ","
	if g.role.Block {
		sep = ";"
	}
	inner := g.open.Is("{") && (len(afterOpen) > 0 || len(units) > 0)
	if inner {
		p.write(" ")
	}
	for _, c := range afterOpen {
		p.comment(c.com)
		p.write(" ")
	}
	for i, u := range units {
		if i > 0 {
			p.write(" ")
		}
		for _, c := range u.lead {
			p.comment(c.com)
			p.write(" ")
		}
		p.seq(u.body, p.indent)
		if i < len(units)-1 {
			p.write(sep)
		}
		for _, c := range u.trail {
			p.write(" ")
			p.comment(c.com)
		}
	}
	if inner {
		p.trim()
		p.write(" ")
	}
}

// space is the spacing between two tokens on one line, by the table of fmt.md 4.7.
func (p *printer) space(prev, cur token.Token) int {
	switch {
	case prev.Is("(") || prev.Is("["):
		return 0
	case prev.Is("{"):
		if cur.Is("}") {
			return 0
		}
		return 1
	case cur.Is(")") || cur.Is("]"):
		return 0
	case cur.Is("}"):
		return 1
	case cur.Is(".") && prev.Kind == token.Int:
		return 1 // 0.x would read as a float
	case cur.Is(",") || cur.Is(";") || cur.Is(":") || cur.Is(".") || cur.Is("?."):
		return 0
	case prev.Is(",") || prev.Is(";") || prev.Is(":"):
		return 1
	case prev.Is(".") || prev.Is("?.") || prev.Is("..."):
		return 0
	case (cur.Is("(") || cur.Is("[")) && (lexer.Operand(prev) || prev.Is("func")):
		return 0
	case cur.Is("{") && p.roles[cur.Off].Typed:
		return 0
	case p.prefix[prev.Off] && prev.Kind == token.Op:
		return 0
	}
	return 1
}

// ---- templates

// template writes a template string: its text as written, and each interpolation's code with
// the spacing of 4.7, keeping every line break and the indentation after it as written.
func (p *printer) template(t token.Token) {
	at := t.Off
	for _, part := range t.Parts {
		if part.Tokens == nil {
			continue
		}
		p.write(p.src[at:part.Off])
		p.write("${")
		g := p.build(part.Tokens, part.Off, part.End)
		end := p.inline(g.elems, part.Off+2)
		if close := part.End - 1; p.newlines(end, close) > 0 {
			p.write(breaks(p.src[end:close]))
		}
		p.write("}")
		at = part.End
	}
	p.write(p.src[at:t.End])
}

// breaks is the whitespace between two tokens of an interpolation that holds a line break, as
// written, except that no line ends in spaces, tabs, or a \r (4.1): only the text of a template
// is kept byte for byte.
func breaks(gap string) string {
	lines := strings.Split(gap, "\n")
	for i := range lines[:len(lines)-1] {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}
	return strings.Join(lines, "\n")
}

// inline writes the elements of an interpolation, token by token, and returns the source offset
// after the last one.
func (p *printer) inline(es []elem, from int) int {
	var prev *token.Token
	prevEnd := from
	var walk func(es []elem)
	gap := func(off int, cur *token.Token) {
		switch {
		case p.newlines(prevEnd, off) > 0:
			p.write(breaks(p.src[prevEnd:off]))
		case prevEnd == from:
		case prev == nil || cur == nil:
			p.write(" ")
		default:
			p.write(strings.Repeat(" ", p.space(*prev, *cur)))
		}
	}
	walk = func(es []elem) {
		for _, e := range es {
			switch {
			case e.isNewline():
				continue
			case e.com != nil:
				gap(e.com.Off, nil)
				p.comment(e.com)
				prev, prevEnd = nil, e.com.End
			case e.grp != nil:
				open, close := e.grp.open, e.grp.close
				gap(open.Off, &open)
				p.token(open)
				prev, prevEnd = &open, open.End
				walk(e.grp.elems)
				gap(close.Off, &close)
				p.token(close)
				prev, prevEnd = &close, close.End
			default:
				t := *e.tok
				gap(t.Off, &t)
				p.token(t)
				prev, prevEnd = &t, t.End
			}
		}
	}
	walk(es)
	return prevEnd
}
