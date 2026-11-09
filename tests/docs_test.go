package tests

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/format"
	"github.com/ant-src/ant/internal/lexer"
	"github.com/ant-src/ant/parser"
)

// TestDocs holds every page under ../docs to the contract ../docs/README.md states for writers:
// a status line, every ant block formatted and passing, every relative link naming a file, and on
// a reference page only the sections of the page format, in order. ANT_DOCS_ONLY limits it to
// the pages whose path under docs/ starts with its value, such as "lib/Array/map.md" or "tour/".
// TestDocsReference, in docs_reference_test.go, keeps the reference in step with the prelude.

const docsRoot = "../docs"

// docTimeout bounds one run of a block, and docCasesTimeout one block of cases. They are generous
// because every page runs at once, under the race detector in CI.
const (
	docTimeout      = 30 * time.Second
	docCasesTimeout = 30 * time.Second
)

// docSections are the second-level headings a function, type, keyword, operator, literal, or
// built-in type page may have, in the order they must appear (README.md, Page format).
var docSections = []string{
	"Details", "Basic examples", "Scope", "Named arguments", "Applications",
	"Properties and relations", "Possible issues", "Neat examples", "See also",
}

var statusLine = regexp.MustCompile(`^<!-- status: (skeleton|draft|reviewed) -->$`)

// docBlock is a fenced ant block of a page.
type docBlock struct {
	line       int    // the line of the opening fence; the code starts on the next
	name       string // the file name of a block written ```ant [path]
	signature  bool   // a ```ant signature block of the generated part
	src        string
	output     *docBlock // the text block right after it: its expected standard output
	directives []docDirective
}

// docDirective is a <!-- test: ... --> comment, which prepares or changes the test of the next ant
// block.
type docDirective struct {
	line int
	name string
	arg  string
	body string
}

// docLine is a line of a page outside every fence and comment: prose, headings, and links.
type docLine struct {
	n    int
	text string
}

// docPageParts is what parsePage finds in a page.
type docPageParts struct {
	blocks []*docBlock
	prose  []docLine
}

var (
	fenceOpen   = regexp.MustCompile("^( {0,3})(`{3,}|~{3,})\\s*([\\w-]*)\\s*(.*)$")
	heading2    = regexp.MustCompile(`^ {0,3}## +(.+?)\s*$`)
	mdLink      = regexp.MustCompile(`\]\(\s*<?([^)\s>]+)>?(?:\s+"[^"]*")?\s*\)`)
	codeSpan    = regexp.MustCompile("`[^`]*`")
	testComment = regexp.MustCompile(`(?i)^tests?\s*:`)
)

// docDirectives lists every directive: whether it takes an argument on its first line, and
// whether it takes a body on the lines after.
var docDirectives = map[string]struct{ arg, body bool }{
	"file":       {true, true},
	"stdin":      {false, true},
	"args":       {true, false},
	"before":     {false, true},
	"cmd":        {true, false},
	"exit":       {true, false},
	"check-only": {true, false},
}

// parsePage splits a page into its ant blocks and its prose, and returns the problems with its
// structure.
func parsePage(rel, src string) (docPageParts, []string) {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var parts docPageParts
	var problems []string
	var pending []docDirective
	var last *docBlock // the last ant block, while only blank lines follow it
	inGenerated := false
	bad := func(line int, format string, args ...any) {
		problems = append(problems, fmt.Sprintf("%s:%d: %s", rel, line, fmt.Sprintf(format, args...)))
	}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		switch line {
		case genStart:
			inGenerated = true
		case genEnd:
			inGenerated = false
		}
		if m := fenceOpen.FindStringSubmatch(line); m != nil {
			indent, fence, lang, info := len(m[1]), m[2], m[3], strings.TrimSpace(m[4])
			start := i + 1
			var body []string
			closed := false
			for i++; i < len(lines); i++ {
				l := strings.TrimSpace(lines[i])
				if strings.HasPrefix(l, fence) && strings.Trim(l, fence[:1]) == "" {
					closed = true
					break
				}
				l = lines[i]
				for k := 0; k < indent && strings.HasPrefix(l, " "); k++ {
					l = l[1:]
				}
				body = append(body, l)
			}
			if !closed {
				bad(start, "the block is not closed")
				break
			}
			b := &docBlock{line: start, src: strings.Join(body, "\n") + "\n"}
			switch {
			case lang == "text" && last != nil:
				last.output = b
				last = nil
				continue
			case lang != "ant":
				if len(pending) > 0 {
					bad(pending[0].line, "test: %s comes before a %q block, not an ant block", pending[0].name, lang)
					pending = nil
				}
				last = nil
				continue
			}
			switch {
			case info == "signature":
				b.signature = true
				if !inGenerated {
					bad(start, "a signature block belongs to the generated part only")
				}
			case inGenerated:
				bad(start, "the generated part holds signature blocks only")
			case info == "":
			case strings.HasPrefix(info, "[") && strings.HasSuffix(info, "]"):
				b.name = info[1 : len(info)-1]
			default:
				bad(start, "unknown ant block info %q; use [file name] or nothing", info)
			}
			b.directives, pending = pending, nil
			parts.blocks = append(parts.blocks, b)
			last = b
			continue
		}
		if !strings.HasPrefix(strings.TrimLeft(line, " "), "<!--") || len(line)-len(strings.TrimLeft(line, " ")) > 3 {
			if strings.TrimSpace(line) != "" {
				last = nil
			}
			parts.prose = append(parts.prose, docLine{i + 1, line})
			continue
		}
		start := i + 1
		text := strings.TrimPrefix(strings.TrimLeft(line, " "), "<!--")
		for !strings.Contains(text, "-->") && i+1 < len(lines) {
			i++
			text += "\n" + lines[i]
		}
		if !strings.Contains(text, "-->") {
			bad(start, "the comment is not closed")
			break
		}
		text, _, _ = strings.Cut(text, "-->")
		text = strings.TrimSpace(text)
		if !testComment.MatchString(text) {
			continue
		}
		rest, ok := strings.CutPrefix(text, "test:")
		if !ok {
			bad(start, "a directive starts with test: exactly")
			continue
		}
		head, body, _ := strings.Cut(rest, "\n")
		name, arg, _ := strings.Cut(strings.TrimSpace(head), " ")
		arg = strings.TrimSpace(arg)
		spec, known := docDirectives[name]
		switch {
		case name == "error" || name == "panics":
			bad(start, "test: %s is gone: mark the failing line with a // %s: comment, so the reader sees it", name, strings.TrimSuffix(name, "s"))
		case !known:
			bad(start, "unknown directive test: %s", name)
		case spec.arg && arg == "":
			bad(start, "test: %s needs an argument", name)
		case !spec.arg && arg != "":
			bad(start, "test: %s takes no argument; its body goes on the lines after", name)
		case !spec.body && strings.TrimSpace(body) != "":
			bad(start, "test: %s takes no body", name)
		}
		if body != "" {
			body += "\n"
		}
		pending = append(pending, docDirective{line: start, name: name, arg: arg, body: body})
	}
	if len(pending) > 0 {
		bad(pending[0].line, "test: %s comes before no ant block", pending[0].name)
	}
	return parts, problems
}

// caseInput is one input of an ant block, read as the REPL reads it: a line, or as many lines as
// it takes to close what one opens.
type caseInput struct {
	first, last int    // its lines in the block, from 1
	src         string // the text, with a case's comment turned into the table's "// =>"
	isCase      bool
}

// caseInputs splits an ant block into its inputs and reports whether any is a case: an
// expression, alone in its input, whose last line ends in a // comment giving its value (README.md,
// Cases). A comment after a declaration, an assignment, a loop, or any other statement that is not
// a value is an ordinary comment. The comment is found by the lexer, so a // inside a string is
// not one.
func caseInputs(src string) ([]caseInput, bool) {
	var inputs []caseInput
	any := false
	var buf []string
	first := 0
	lines := strings.Split(strings.TrimSuffix(src, "\n"), "\n")
	for i, line := range lines {
		if len(buf) == 0 {
			if strings.TrimSpace(line) == "" {
				continue
			}
			first = i + 1
		}
		buf = append(buf, line)
		if ant.Incomplete(strings.Join(buf, "\n")) && i < len(lines)-1 {
			continue
		}
		in := caseInput{first: first, last: i + 1, src: strings.Join(buf, "\n")}
		buf = nil
		if expr, value, ok := caseSplit(in.src); ok {
			in.isCase = true
			in.src = expr + " // => " + value
			any = true
		}
		inputs = append(inputs, in)
	}
	return inputs, any
}

// docMarker is a comment that marks the line of a program that fails: "// error: <message>" for a
// line ant check rejects, "// panic: <message>" for a line that panics. At the end of a line it
// marks that line; on a line of its own, the code line above it.
type docMarker struct {
	kind, text string
	line       int // in the block, from 1
}

var markerLine = regexp.MustCompile(`^//\s*(error|panic):\s*(.*)$`)

// blockMarkers finds the markers of a block. The comments come from the lexer, so a // inside a
// string is not one; a block the lexer rejects, which can only show a compile error, is read line
// by line.
func blockMarkers(src string) ([]docMarker, []string) {
	lines := strings.Split(strings.TrimSuffix(src, "\n"), "\n")
	type comment struct {
		text string
		line int
		own  bool
	}
	var comments []comment
	if _, cs, err := lexer.LexLayout(src, false); err == nil {
		for _, c := range cs {
			start := strings.LastIndex(src[:c.Off], "\n") + 1
			comments = append(comments, comment{c.Text, strings.Count(src[:c.Off], "\n") + 1, strings.TrimSpace(src[start:c.Off]) == ""})
		}
	} else {
		lax := regexp.MustCompile(`//\s*(?:error|panic):`)
		for i, l := range lines {
			if m := lax.FindStringIndex(l); m != nil {
				comments = append(comments, comment{l[m[0]:], i + 1, strings.TrimSpace(l[:m[0]]) == ""})
			}
		}
	}
	var markers []docMarker
	var problems []string
	for _, c := range comments {
		m := markerLine.FindStringSubmatch(c.text)
		if m == nil {
			continue
		}
		mk := docMarker{kind: m[1], text: strings.TrimSpace(m[2]), line: c.line}
		if mk.text == "" {
			problems = append(problems, fmt.Sprintf("line %d: // %s: needs the message", c.line, mk.kind))
			continue
		}
		if c.own {
			mk.line = 0
			for i := c.line - 2; i >= 0; i-- {
				if t := strings.TrimSpace(lines[i]); t != "" && !strings.HasPrefix(t, "//") {
					mk.line = i + 1
					break
				}
			}
			if mk.line == 0 {
				problems = append(problems, fmt.Sprintf("line %d: // %s: marks the code line above it, and there is none", c.line, mk.kind))
				continue
			}
		}
		markers = append(markers, mk)
	}
	return markers, problems
}

// caseSplit splits an input into an expression and the value its trailing comment gives, if the
// input is a case.
func caseSplit(src string) (string, string, bool) {
	_, comments, err := lexer.LexLayout(src, false)
	if err != nil || len(comments) == 0 {
		return "", "", false
	}
	c := comments[len(comments)-1]
	if !strings.HasPrefix(c.Text, "//") || strings.TrimRight(src[c.End:], " \t") != "" {
		return "", "", false
	}
	lineStart := strings.LastIndex(src[:c.Off], "\n") + 1
	if strings.TrimSpace(src[lineStart:c.Off]) == "" {
		return "", "", false // a comment on a line of its own
	}
	expr := strings.TrimRight(src[:c.Off], " \t")
	value := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
	if strings.HasPrefix(value, "error:") {
		return "", "", false // a line ant check rejects, not a value
	}
	f, err := parser.Parse("case.ant", expr, parser.Session)
	if err != nil || len(f.Decls) > 0 || len(f.Stmts) != 1 || len(f.Imports) > 0 {
		return "", "", false
	}
	switch f.Stmts[0].(type) {
	case *ast.Assign, *ast.For, *ast.While, *ast.Go, *ast.Defer, *ast.Return, *ast.Break,
		*ast.Continue, *ast.Test, *ast.Select:
		return "", "", false
	}
	if _, ok := f.Stmts[0].(ast.Expr); !ok {
		return "", "", false
	}
	return expr, value, true
}

func TestDocs(t *testing.T) {
	only := os.Getenv("ANT_DOCS_ONLY")
	files := map[string]bool{} // every file under docs/, by its exact path
	var pages []string
	err := filepath.WalkDir(docsRoot, func(file string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(docsRoot, file)
		rel = filepath.ToSlash(rel)
		files[rel] = true
		if strings.HasSuffix(rel, ".md") && strings.HasPrefix(rel, only) {
			pages = append(pages, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) == 0 {
		t.Fatalf("no pages under %s match %q", docsRoot, only)
	}
	bin := buildAnt(t)
	for _, rel := range pages {
		t.Run(rel, func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(filepath.Join(docsRoot, rel))
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range pageProblems(t, bin, rel, string(data), files) {
				t.Error(p)
			}
		})
	}
}

// buildAnt builds the ant command into a temporary directory, so a page runs its examples as a
// reader does.
func buildAnt(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "ant")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	out, err := exec.Command("go", "build", "-o", bin, "github.com/ant-src/ant/cmd/ant").CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

func pageProblems(t *testing.T, bin, rel, src string, files map[string]bool) []string {
	status := ""
	first, _, _ := strings.Cut(src, "\n")
	if m := statusLine.FindStringSubmatch(strings.TrimRight(first, "\r")); m != nil {
		status = m[1]
	} else if rel != "README.md" {
		return []string{rel + ":1: the first line must be <!-- status: skeleton|draft|reviewed -->"}
	}
	parts, problems := parsePage(rel, src)
	problems = append(problems, linkProblems(rel, parts.prose, files)...)
	problems = append(problems, headingProblems(rel, parts, status)...)

	dir := t.TempDir()
	// Files first: every named block and every file directive of the page, so any block can use
	// any of them.
	for _, b := range parts.blocks {
		for _, d := range b.directives {
			if d.name == "file" {
				problems = append(problems, writeDocFile(rel, d.line, dir, d.arg, d.body)...)
			}
		}
		if b.name != "" {
			problems = append(problems, writeDocFile(rel, b.line, dir, b.name, b.src)...)
		}
	}
	for _, b := range parts.blocks {
		if !b.signature {
			problems = append(problems, blockProblems(t, bin, rel, dir, b)...)
		}
	}
	return problems
}

func writeDocFile(rel string, line int, dir, name, body string) []string {
	if !filepath.IsLocal(name) {
		return []string{fmt.Sprintf("%s:%d: %s must be a relative path inside the page's directory", rel, line, name)}
	}
	file := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return []string{fmt.Sprintf("%s:%d: %v", rel, line, err)}
	}
	if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
		return []string{fmt.Sprintf("%s:%d: %v", rel, line, err)}
	}
	return nil
}

// pageLines rewrites the positions a tool reports in file, "file:12:", to the lines of the page:
// the block's code starts on line start, after hidden lines of test: before.
func pageLines(msg, file, rel string, start, hidden int) string {
	re := regexp.MustCompile(`(?:[^\s:]*[/\\])?` + regexp.QuoteMeta(filepath.Base(file)) + `:(\d+)`)
	return re.ReplaceAllStringFunc(msg, func(m string) string {
		n, _ := strconv.Atoi(re.FindStringSubmatch(m)[1])
		if n <= hidden {
			return fmt.Sprintf("%s (test: before, line %d)", rel, n)
		}
		return fmt.Sprintf("%s:%d", rel, start+n-hidden-1)
	})
}

func blockProblems(t *testing.T, bin, rel, dir string, b *docBlock) []string {
	var problems []string
	bad := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf("%s:%d: %s", rel, b.line, fmt.Sprintf(format, args...)))
	}
	ds := map[string]docDirective{}
	for _, d := range b.directives {
		if _, dup := ds[d.name]; dup && d.name != "file" {
			problems = append(problems, fmt.Sprintf("%s:%d: a second test: %s for the same block", rel, d.line, d.name))
		}
		ds[d.name] = d
	}
	has := func(name string) bool { _, ok := ds[name]; return ok }
	markers, mp := blockMarkers(b.src)
	for _, p := range mp {
		bad("%s", p)
	}
	var errs, panics []docMarker
	for _, m := range markers {
		if m.kind == "error" {
			errs = append(errs, m)
		} else {
			panics = append(panics, m)
		}
	}
	fails := len(errs) > 0
	before := ds["before"].body
	hidden := strings.Count(before, "\n")
	code := b.line + 1

	ext := filepath.Ext(b.name)
	isAnt := b.name == "" || ext == ".ant"
	if isAnt {
		name := "main.ant"
		if b.name != "" {
			name = b.name
		}
		out, err := format.Source(name, []byte(b.src))
		switch {
		case err != nil && fails:
			// Code that does not parse cannot be formatted; it shows a syntax error.
		case err != nil:
			bad("%v", err)
			return problems
		case string(out) != b.src:
			bad("the block is not formatted as ant fmt formats it:\n%s", out)
		}
	}

	if inputs, isCases := caseInputs(b.src); isCases && b.name == "" && !fails {
		for name := range ds {
			if name != "before" && name != "file" {
				bad("a block of cases takes only test: before, not test: %s", name)
			}
		}
		if b.output != nil {
			bad("a block of cases has no output block; its comments give the values")
		}
		var src strings.Builder
		src.WriteString(before)
		n := 0
		for _, in := range inputs {
			for ; n < in.first-1; n++ {
				src.WriteString("\n")
			}
			src.WriteString(in.src + "\n")
			n = in.last
		}
		fake := filepath.Join(dir, "cases.ant")
		for _, p := range tableProblems(t.Context(), fake, src.String(), docCasesTimeout) {
			// The table speaks of its own // => marker; a page writes the value after //.
			p = strings.ReplaceAll(p, "give it a // => or bind it with var", "end its last line with // and its value, or bind it with var")
			problems = append(problems, pageLines(p, fake, rel, code, hidden))
		}
		return problems
	}

	// A program. It runs when it is the page's script or names a command; a named block is a
	// module, a test file, or a data file the others use.
	if fails {
		if _, isCases := caseInputs(b.src); isCases {
			bad("a block with // error: has no cases: it does not run, so their values would not be checked")
		}
		if !isAnt || has("check-only") || has("cmd") {
			bad("// error: marks a line of an ant program without test: check-only or test: cmd")
		}
		if len(panics) > 0 {
			bad("a block with // error: does not run, so it has no // panic:")
		}
	}
	running := !fails && !has("check-only") && (b.name == "" || has("cmd"))
	if len(panics) > 0 && !running {
		bad("// panic: marks a line of a block that runs")
	}
	if len(panics) > 1 {
		bad("a program panics once: one // panic: per block")
	}
	for _, d := range b.directives {
		ok := true
		switch d.name {
		case "before", "args":
			ok = b.name == "" && !has("cmd")
		case "stdin", "exit":
			ok = running
		case "cmd":
			ok = !fails && !has("check-only")
		case "check-only":
			ok = isAnt
		}
		if !ok {
			problems = append(problems, fmt.Sprintf("%s:%d: test: %s does not apply to this block", rel, d.line, d.name))
		}
	}
	if has("exit") && len(panics) > 0 {
		bad("test: exit and // panic: exclude each other")
	}
	if b.output != nil && !running {
		bad("a block that does not run has no output block")
	}
	want := 0
	if e, ok := ds["exit"]; ok {
		n, err := strconv.Atoi(e.arg)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s:%d: test: exit needs a number, not %q", rel, e.line, e.arg))
		}
		want = n
	}
	if len(problems) > 0 {
		return problems
	}

	main := b.name
	if main == "" {
		main = "main.ant"
		if err := os.WriteFile(filepath.Join(dir, main), []byte(before+b.src), 0o644); err != nil {
			return append(problems, err.Error())
		}
	}
	if isAnt {
		out, status := runDocAnt(bin, dir, "", "check", main)
		out = pageLines(out, main, rel, code, hidden)
		switch {
		case fails && status == 0:
			bad("ant check passes, but the block marks a line with // error:")
		case fails:
			problems = append(problems, errorMarkerProblems(rel, out, code, errs)...)
		case status != 0:
			bad("ant check:\n%s", out)
		}
	}
	if !running {
		return problems
	}
	args := append([]string{"run", main}, strings.Fields(ds["args"].arg)...)
	if c, ok := ds["cmd"]; ok {
		f := strings.Fields(c.arg)
		if len(f) == 0 || f[0] != "ant" {
			bad("test: cmd runs an ant command, such as test: cmd ant test .")
			return problems
		}
		args = f[1:]
	}
	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), docTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(ds["stdin"].body)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	status := 0
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			bad("ant %s: %v", strings.Join(args, " "), err)
			return problems
		}
		status = exit.ExitCode()
	}
	errText := pageLines(stderr.String(), main, rel, code, hidden)
	if len(panics) > 0 {
		p := panics[0]
		line := code + p.line - 1
		switch {
		case status == 0 || !strings.Contains(stderr.String(), "panic: "+p.text):
			bad("the program does not panic with %q; it exits with %d:\n%s", p.text, status, errText)
		case !strings.Contains(errText, fmt.Sprintf("%s:%d:", rel, line)):
			bad("the program panics, but not at line %d, which // panic: marks:\n%s", line, errText)
		}
	} else if status != want {
		bad("ant %s exits with %d, not %d:\n%s%s", strings.Join(args, " "), status, want, stdout.String(), errText)
	}
	// One trailing line break does not count, so output that does not end a line can be shown.
	if b.output != nil && strings.TrimSuffix(stdout.String(), "\n") != strings.TrimSuffix(b.output.src, "\n") {
		bad("the output differs from the text block at line %d:\n--- got\n%s--- want\n%s", b.output.line, stdout.String(), b.output.src)
	}
	return problems
}

func runDocAnt(bin, dir, stdin string, args ...string) (string, int) {
	ctx, cancel := context.WithTimeout(context.Background(), docTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(out), exit.ExitCode()
	}
	if err != nil {
		return err.Error(), -1
	}
	return string(out), 0
}

// linkProblems checks that every relative link of a page's prose names a file under docs/, with
// the case it is written in, whatever the file system ignores.
func linkProblems(rel string, prose []docLine, files map[string]bool) []string {
	var problems []string
	for _, l := range prose {
		for _, m := range mdLink.FindAllStringSubmatch(codeSpan.ReplaceAllString(l.text, ""), -1) {
			target := m[1]
			// The docs are for people using Ant; the spec is for people implementing it.
			if strings.Contains(target, "github.com/ant-src/spec") {
				problems = append(problems, fmt.Sprintf("%s:%d: the link %s names the spec; link a docs page instead, or none", rel, l.n, m[1]))
				continue
			}
			// A link from the site's root, such as /vscode, names a page of the site, not of docs/.
			if strings.Contains(target, "://") || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "/") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			target, _, _ = strings.Cut(target, "#")
			path := filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(rel), target)))
			if !files[path] {
				problems = append(problems, fmt.Sprintf("%s:%d: the link %s names no file under docs/", rel, l.n, m[1]))
			}
		}
	}
	return problems
}

// headingProblems checks the sections of a reference page: only the headings of docSections, in
// their order, and once the page is written, Basic examples with an example in it.
func headingProblems(rel string, parts docPageParts, status string) []string {
	leaf := strings.HasPrefix(rel, "lib/") && !strings.HasSuffix(rel, "/index.md") && !strings.HasPrefix(rel, "lib/guides/") ||
		strings.HasPrefix(rel, "language/") && strings.Count(rel, "/") == 2
	if !leaf {
		return nil
	}
	var problems []string
	at := -1
	basic, basicEnd := 0, 0 // the lines of Basic examples and of the heading after it
	for _, l := range parts.prose {
		m := heading2.FindStringSubmatch(l.text)
		if m == nil {
			continue
		}
		title := strings.TrimRight(m[1], "# ")
		if basic > 0 && basicEnd == 0 {
			basicEnd = l.n
		}
		k := -1
		for j, s := range docSections {
			if s == title {
				k = j
			}
		}
		switch {
		case k < 0:
			problems = append(problems, fmt.Sprintf("%s:%d: ## %s is not a section of the page format", rel, l.n, title))
		case k <= at:
			problems = append(problems, fmt.Sprintf("%s:%d: ## %s is out of order", rel, l.n, title))
		default:
			at = k
		}
		if title == "Basic examples" {
			basic = l.n
		}
	}
	if status == "skeleton" {
		return problems
	}
	examples := 0
	for _, b := range parts.blocks {
		if basic > 0 && b.line > basic && (basicEnd == 0 || b.line < basicEnd) {
			examples++
		}
	}
	if examples == 0 {
		problems = append(problems, fmt.Sprintf("%s:1: a %s page needs ## Basic examples with an example", rel, status))
	}
	return problems
}

// errorMarkerProblems holds the errors ant check reports, out, to the lines // error: marks: each
// marked line has an error containing its text, and each error is on a marked line.
func errorMarkerProblems(rel, out string, code int, errs []docMarker) []string {
	type diag struct {
		line int // 0 on a hidden line of test: before
		msg  string
	}
	var diags []diag
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(rel) + `(?::(\d+)| \(test: before, line \d+\)):\d+: (.*)$`)
	for _, l := range strings.Split(out, "\n") {
		if m := re.FindStringSubmatch(l); m != nil {
			n, _ := strconv.Atoi(m[1])
			diags = append(diags, diag{n, m[2]})
		}
	}
	var problems []string
	marked := map[int]bool{}
	for _, e := range errs {
		line := code + e.line - 1
		marked[line] = true
		found := false
		for _, d := range diags {
			found = found || d.line == line && strings.Contains(d.msg, e.text)
		}
		if !found {
			problems = append(problems, fmt.Sprintf("%s:%d: ant check does not report %q at this line:\n%s", rel, line, e.text, out))
		}
	}
	for _, d := range diags {
		if !marked[d.line] {
			problems = append(problems, fmt.Sprintf("%s:%d: ant check reports %q, but no // error: marks this line", rel, d.line, d.msg))
		}
	}
	return problems
}
