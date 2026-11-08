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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/format"
	"github.com/ant-src/ant/internal/prelude"
)

// The tests in this file hold the documentation under ../docs to its contract, which
// ../docs/README.md states for writers: every page starts with a status line, every ant block
// is formatted and passes, every relative link resolves, and the standard library reference has
// one page per prelude declaration whose generated part matches the prelude.
//
// TestDocs runs the pages. ANT_DOCS_ONLY limits it to the pages whose path under docs/ starts
// with its value, such as "lib/Array/map.md" or "tour/".
//
// TestDocsReference checks the reference against internal/prelude/prelude.d.ant. With
// ANT_DOCS_UPDATE=1 it writes instead: a skeleton for every missing page, and the generated part
// of every existing one.

const docsRoot = "../docs"

// docSections are the second-level headings a function, type, keyword, or operator page may
// have, in the order they must appear (README.md, Page format).
var docSections = []string{
	"Details", "Basic examples", "Scope", "Named arguments", "Applications",
	"Properties and relations", "Possible issues", "Neat examples", "See also",
}

// foldedUnions are the prelude unions whose members share the union's page: the members are
// cases of one idea, mostly empty records, and read best side by side.
var foldedUnions = []string{
	"Result", "Html.Node", "Reflect.Shape", "Time.Unit", "Time.Disambiguation", "Time.RoundingMode",
}

// typeHome places the prelude types that have no namespace of their own.
var typeHome = map[string]string{
	"Entry": "Map", "Indexed": "Array", "Seq": "Iter",
	"Integer": "Math", "Number": "Math", "Ordered": "Math",
}

const (
	genStart = "<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->"
	genEnd   = "<!-- end generated -->"
)

var statusLine = regexp.MustCompile(`^<!-- status: (skeleton|draft|reviewed) -->$`)

// docPage is a page of the reference: its path under docs/, its title, the namespace doc for a
// namespace page, and the declarations it documents.
type docPage struct {
	path  string
	title string
	nsDoc string
	decls []ast.Decl
}

// referencePages maps every page the prelude needs to what it documents, and every documented
// name to its page.
func referencePages(t *testing.T) (map[string]*docPage, map[string]string) {
	t.Helper()
	file, err := prelude.Decls()
	if err != nil {
		t.Fatal(err)
	}
	namespaces := map[string]bool{}
	types := map[string]*ast.TypeDecl{}
	for _, d := range file.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Owner != "" {
				namespaces[d.Owner] = true
			}
		case *ast.TypeDecl:
			if d.Owner != "" {
				namespaces[d.Owner] = true
			}
			types[qualified(d.Owner, d.Name)] = d
		}
	}
	for name := range types {
		if _, doc := prelude.NamespaceDoc(name); doc != "" {
			namespaces[name] = true
		}
	}
	fold := map[string]string{}
	for _, u := range foldedUnions {
		d := types[u]
		if d == nil {
			t.Fatalf("folded union %s is not in the prelude", u)
		}
		_, body, _ := strings.Cut(d.Sig, "=")
		for _, m := range strings.Split(body, "|") {
			m, _, _ = strings.Cut(strings.TrimSpace(m), "[")
			if types[m] != nil {
				fold[m] = u
			}
		}
	}
	pages := map[string]*docPage{}
	names := map[string]string{}
	page := func(path, title string) *docPage {
		if p := pages[path]; p != nil {
			return p
		}
		p := &docPage{path: path, title: title}
		pages[path] = p
		return p
	}
	nsPage := func(ns string) *docPage {
		p := page("lib/"+ns+"/index.md", ns)
		_, p.nsDoc = prelude.NamespaceDoc(ns)
		names[ns] = p.path
		return p
	}
	page("lib/prelude/index.md", "Prelude")
	for ns := range namespaces {
		nsPage(ns)
	}
	var typePage func(name string) *docPage
	typePage = func(name string) *docPage {
		if u, ok := fold[name]; ok && u != name {
			return typePage(u)
		}
		owner, base, dotted := strings.Cut(name, ".")
		switch {
		case !dotted && namespaces[name]:
			return nsPage(name)
		case !dotted:
			home, ok := typeHome[name]
			if !ok {
				t.Fatalf("prelude type %s has no namespace; add it to typeHome", name)
			}
			return page("lib/"+home+"/types/"+name+".md", name)
		default:
			return page("lib/"+owner+"/types/"+base+".md", name)
		}
	}
	for _, d := range file.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			name := qualified(d.Owner, d.Name)
			dir := d.Owner
			if dir == "" {
				dir = "prelude"
			}
			p := page("lib/"+dir+"/"+d.Name+".md", name)
			p.decls = append(p.decls, d)
			names[name] = p.path
		case *ast.TypeDecl:
			name := qualified(d.Owner, d.Name)
			p := typePage(name)
			p.decls = append(p.decls, d)
			names[name] = p.path
		}
	}
	// A folded union's page shows the union before its members.
	for _, u := range foldedUnions {
		p := pages[names[u]]
		sort.SliceStable(p.decls, func(i, j int) bool {
			d, ok := p.decls[i].(*ast.TypeDecl)
			return ok && qualified(d.Owner, d.Name) == u
		})
	}
	return pages, names
}

func qualified(owner, name string) string {
	if owner == "" {
		return name
	}
	return owner + "." + name
}

var docLink = regexp.MustCompile(`\[([A-Za-z_][\w.]*)\]`)

// generated is the part of a reference page that comes from the prelude: the title, each
// declaration's signature, and its doc without the examples, which the page shows as cases.
func generated(p *docPage, names map[string]string) string {
	var b strings.Builder
	b.WriteString(genStart + "\n\n# " + p.title + "\n")
	prose := func(doc string) {
		text, _ := splitDoc(doc)
		if text == "" {
			return
		}
		text = docLink.ReplaceAllStringFunc(text, func(m string) string {
			name := m[1 : len(m)-1]
			target, ok := names[name]
			if !ok {
				return m
			}
			if target == p.path {
				return name
			}
			rel, _ := filepath.Rel(filepath.Dir(p.path), target)
			return "[" + name + "](" + filepath.ToSlash(rel) + ")"
		})
		b.WriteString("\n" + text + "\n")
	}
	prose(p.nsDoc)
	for _, d := range p.decls {
		sig, doc := "", ""
		switch d := d.(type) {
		case *ast.FuncDecl:
			sig, doc = d.Sig, d.Doc
		case *ast.TypeDecl:
			sig, doc = d.Sig, d.Doc
		}
		b.WriteString("\n```ant signature\n" + sig + "\n```\n")
		prose(doc)
	}
	b.WriteString("\n" + genEnd + "\n")
	return b.String()
}

// splitDoc separates a doc comment's prose from its examples, the runs of lines indented by four
// spaces (doc.md, section 2). A line that is only "// value" gives the value of the line before.
func splitDoc(doc string) (string, [][]string) {
	var prose []string
	var examples [][]string
	var cur []string
	flush := func() {
		if cur != nil {
			examples = append(examples, cur)
			cur = nil
		}
	}
	for _, line := range strings.Split(doc, "\n") {
		code, ok := strings.CutPrefix(line, "    ")
		if !ok {
			if line == "" && cur != nil {
				continue
			}
			flush()
			prose = append(prose, line)
			continue
		}
		if v, ok := strings.CutPrefix(strings.TrimSpace(code), "// "); ok && len(cur) > 0 && strings.HasPrefix(code, " ") {
			cur[len(cur)-1] += " // " + v
			continue
		}
		cur = append(cur, code)
	}
	flush()
	text := strings.TrimSpace(regexp.MustCompile(`\n{3,}`).ReplaceAllString(strings.Join(prose, "\n"), "\n\n"))
	return text, examples
}

// skeleton is a new reference page: the status line, the generated part, and the doc's examples
// with cases as Basic examples.
func skeleton(p *docPage, names map[string]string) string {
	var b strings.Builder
	b.WriteString("<!-- status: skeleton -->\n" + generated(p, names))
	var blocks []string
	for _, d := range p.decls {
		doc := ""
		switch d := d.(type) {
		case *ast.FuncDecl:
			doc = d.Doc
		case *ast.TypeDecl:
			doc = d.Doc
		}
		_, examples := splitDoc(doc)
		for _, ex := range examples {
			src := strings.Join(ex, "\n") + "\n"
			if !isCaseBlock(src) {
				continue
			}
			if out, err := format.Source("example.ant", []byte(src)); err == nil {
				blocks = append(blocks, "```ant\n"+string(out)+"```\n")
			}
		}
	}
	if len(blocks) > 0 && !strings.HasSuffix(p.path, "/index.md") {
		b.WriteString("\n## Basic examples\n\n" + strings.Join(blocks, "\n"))
	}
	return b.String()
}

// TestDocsReference checks that the reference has a page for every prelude declaration, no page
// for anything else, and a generated part that matches the prelude.
func TestDocsReference(t *testing.T) {
	pages, names := referencePages(t)
	update := os.Getenv("ANT_DOCS_UPDATE") == "1"
	paths := make([]string, 0, len(pages))
	for path := range pages {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	// Types live under types/, so that Time.Zone and Time.zone are two files on a file system that
	// ignores case.
	folded := map[string]string{}
	for _, path := range paths {
		if other, ok := folded[strings.ToLower(path)]; ok {
			t.Errorf("%s and %s differ only in case", other, path)
		}
		folded[strings.ToLower(path)] = path
	}
	for _, path := range paths {
		p := pages[path]
		file := filepath.Join(docsRoot, path)
		data, err := os.ReadFile(file)
		if errors.Is(err, os.ErrNotExist) {
			if !update {
				t.Errorf("%s: no page for %s; run ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference", path, p.title)
				continue
			}
			if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file, []byte(skeleton(p, names)), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		src := string(data)
		i := strings.Index(src, genStart)
		j := strings.Index(src, genEnd)
		if i < 0 || j < i {
			t.Errorf("%s: no generated part", path)
			continue
		}
		end := j + len(genEnd)
		if end < len(src) && src[end] == '\n' {
			end++
		}
		have := src[i:end]
		want := generated(p, names)
		// oxfmt rewraps the prose of every Markdown file (CLAUDE.md), so only the words count.
		if strings.Join(strings.Fields(have), " ") == strings.Join(strings.Fields(want), " ") {
			continue
		}
		if !update {
			t.Errorf("%s: the generated part does not match the prelude; run ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference", path)
			continue
		}
		if err := os.WriteFile(file, []byte(src[:i]+want+src[end:]), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Every page under lib/ documents a declaration, except the overview and the topic guides.
	filepath.WalkDir(filepath.Join(docsRoot, "lib"), func(file string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(file, ".md") {
			return err
		}
		rel, _ := filepath.Rel(docsRoot, file)
		rel = filepath.ToSlash(rel)
		if rel == "lib/index.md" || strings.HasPrefix(rel, "lib/guides/") || pages[rel] != nil {
			return nil
		}
		t.Errorf("%s: documents no prelude declaration", rel)
		return nil
	})
}

// docBlock is a fenced block of a page.
type docBlock struct {
	line       int    // the line of the opening fence
	lang       string // ant, text, sh, ...
	name       string // the file name of an ant block written as [path]
	signature  bool   // an ant block of the generated part, which shows a signature
	src        string
	output     *docBlock // the text block right after an ant block: its expected output
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

var (
	fenceOpen = regexp.MustCompile("^(`{3,}|~{3,})\\s*([\\w-]*)\\s*(.*)$")
	heading2  = regexp.MustCompile(`^## (.+)$`)
	mdLink    = regexp.MustCompile(`\]\(([^)\s]+)\)`)
	codeSpan  = regexp.MustCompile("`[^`]*`")
)

// docDirectives lists every directive and whether its argument is required.
var docDirectives = map[string]bool{
	"file": true, "stdin": false, "args": true, "before": false, "cmd": true,
	"exit": true, "panics": true, "error": true, "check-only": true,
}

// parsePage splits a page into its blocks and returns the problems with its structure.
func parsePage(path, src string) ([]*docBlock, []string) {
	lines := strings.Split(src, "\n")
	var blocks []*docBlock
	var problems []string
	var pending []docDirective
	var last *docBlock // the last ant block, while only blank lines follow it
	bad := func(line int, format string, args ...any) {
		problems = append(problems, fmt.Sprintf("%s:%d: %s", path, line, fmt.Sprintf(format, args...)))
	}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if m := fenceOpen.FindStringSubmatch(line); m != nil {
			fence := m[1]
			b := &docBlock{line: i + 1, lang: m[2]}
			info := strings.TrimSpace(m[3])
			var body []string
			closed := false
			for i++; i < len(lines); i++ {
				l := strings.TrimRight(lines[i], " ")
				if strings.HasPrefix(l, fence) && strings.Trim(l, fence[:1]) == "" {
					closed = true
					break
				}
				body = append(body, lines[i])
			}
			if !closed {
				bad(b.line, "the block is not closed")
				break
			}
			b.src = strings.Join(body, "\n") + "\n"
			switch {
			case b.lang == "text" && last != nil:
				last.output = b
				last = nil
				continue
			case b.lang != "ant":
				if len(pending) > 0 {
					bad(pending[0].line, "test: %s comes before a %q block, not an ant block", pending[0].name, b.lang)
					pending = nil
				}
				last = nil
				continue
			}
			switch {
			case info == "":
			case info == "signature":
				b.signature = true
			case strings.HasPrefix(info, "[") && strings.HasSuffix(info, "]"):
				b.name = info[1 : len(info)-1]
			default:
				bad(b.line, "unknown ant block info %q; use [file name] or nothing", info)
			}
			b.directives, pending = pending, nil
			blocks = append(blocks, b)
			last = b
			continue
		}
		if strings.TrimSpace(line) != "" {
			last = nil
		}
		if !strings.HasPrefix(line, "<!--") {
			continue
		}
		start := i
		text := strings.TrimPrefix(line, "<!--")
		for !strings.Contains(text, "-->") && i+1 < len(lines) {
			i++
			text += "\n" + lines[i]
		}
		text, _, _ = strings.Cut(text, "-->")
		text = strings.TrimSpace(text)
		rest, ok := strings.CutPrefix(text, "test:")
		if !ok {
			continue
		}
		head, body, _ := strings.Cut(rest, "\n")
		name, arg, _ := strings.Cut(strings.TrimSpace(head), " ")
		needsArg, known := docDirectives[name]
		switch {
		case !known:
			bad(start+1, "unknown directive test: %s", name)
		case needsArg && strings.TrimSpace(arg) == "":
			bad(start+1, "test: %s needs an argument", name)
		}
		if body != "" {
			body += "\n"
		}
		pending = append(pending, docDirective{line: start + 1, name: name, arg: strings.TrimSpace(arg), body: body})
	}
	if len(pending) > 0 {
		bad(pending[0].line, "test: %s comes before no ant block", pending[0].name)
	}
	return blocks, problems
}

// isCaseBlock reports whether an ant block is a block of cases: an unindented expression followed
// by a comment, the comment giving its value (README.md, Examples). A comment after a
// declaration, an assignment, or an indented line is an ordinary comment.
func isCaseBlock(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		if _, _, ok := caseSplit(line); ok {
			return true
		}
	}
	return false
}

// caseSplit splits a case line into its expression and its value.
func caseSplit(line string) (string, string, bool) {
	i := strings.LastIndex(line, " // ")
	if i < 0 || line == "" || line[0] == ' ' || line[0] == '\t' {
		return "", "", false
	}
	expr := strings.TrimRight(line[:i], " ")
	if notCase.MatchString(expr) {
		return "", "", false
	}
	return expr, line[i+4:], true
}

// notCase matches the start of a line that is not an expression: a declaration, a test, or an
// assignment.
var notCase = regexp.MustCompile(`^(var|let|import|export|func|type|interface|test)\b|^[\w.\[\]"]+\s*[-+*/%]?=[^=]|^[}\])]`)

func TestDocs(t *testing.T) {
	only := os.Getenv("ANT_DOCS_ONLY")
	var pages []string
	err := filepath.WalkDir(docsRoot, func(file string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(file, ".md") {
			return err
		}
		rel, _ := filepath.Rel(docsRoot, file)
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, only) {
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
			for _, p := range pageProblems(t, bin, rel) {
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
	out, err := exec.Command("go", "build", "-o", bin, "github.com/ant-src/ant/cmd/ant").CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

func pageProblems(t *testing.T, bin, rel string) []string {
	data, err := os.ReadFile(filepath.Join(docsRoot, rel))
	if err != nil {
		return []string{err.Error()}
	}
	src := string(data)
	status := ""
	first, _, _ := strings.Cut(src, "\n")
	if m := statusLine.FindStringSubmatch(first); m != nil {
		status = m[1]
	}
	blocks, problems := parsePage(rel, src)
	bad := func(line int, format string, args ...any) {
		problems = append(problems, fmt.Sprintf("%s:%d: %s", rel, line, fmt.Sprintf(format, args...)))
	}
	if status == "" && rel != "README.md" {
		bad(1, "the first line must be <!-- status: skeleton|draft|reviewed -->")
	}
	problems = append(problems, linkProblems(rel, src)...)
	problems = append(problems, headingProblems(rel, src, status)...)

	dir := t.TempDir()
	// Files first: every named block and every file directive of the page, so any block can use
	// any of them.
	for _, b := range blocks {
		for _, d := range b.directives {
			if d.name == "file" {
				problems = append(problems, writeDocFile(dir, d.arg, d.body)...)
			}
		}
		if b.name != "" {
			problems = append(problems, writeDocFile(dir, b.name, b.src)...)
		}
	}
	for _, b := range blocks {
		if b.signature {
			continue
		}
		problems = append(problems, blockProblems(t, bin, rel, dir, b)...)
	}
	return problems
}

func writeDocFile(dir, name, body string) []string {
	if filepath.IsAbs(name) || strings.Contains(name, "..") {
		return []string{fmt.Sprintf("file %s must be a relative path inside the page's directory", name)}
	}
	file := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return []string{err.Error()}
	}
	if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
		return []string{err.Error()}
	}
	return nil
}

func blockProblems(t *testing.T, bin, rel, dir string, b *docBlock) []string {
	var problems []string
	bad := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf("%s:%d: %s", rel, b.line, fmt.Sprintf(format, args...)))
	}
	ds := map[string]docDirective{}
	for _, d := range b.directives {
		ds[d.name] = d
	}
	_, wantError := ds["error"]
	if !wantError {
		name := "main.ant"
		if b.name != "" {
			name = b.name
		}
		if strings.HasSuffix(name, ".ant") {
			out, err := format.Source(name, []byte(b.src))
			if err != nil {
				bad("%v", err)
				return problems
			}
			if string(out) != b.src {
				bad("the block is not formatted as ant fmt formats it:\n%s", out)
			}
		}
	}
	before := ds["before"].body

	if isCaseBlock(b.src) {
		for name := range ds {
			if name != "before" {
				bad("a block of cases takes only test: before, not test: %s", name)
			}
		}
		if b.name != "" || b.output != nil {
			bad("a block of cases has no file name and no output block")
		}
		// The block's lines stay at their page lines, so a problem names the line of the page.
		pad := b.line - strings.Count(before, "\n")
		if pad < 0 {
			pad = 0
		}
		var src strings.Builder
		src.WriteString(before + strings.Repeat("\n", pad))
		for _, line := range strings.Split(strings.TrimSuffix(b.src, "\n"), "\n") {
			if expr, value, ok := caseSplit(line); ok {
				line = expr + " // => " + value
			}
			src.WriteString(line + "\n")
		}
		fake := filepath.Join(dir, "cases.ant")
		for _, p := range tableProblems(t.Context(), fake, src.String(), tableTimeout) {
			problems = append(problems, strings.Replace(p, fake, rel, 1))
		}
		return problems
	}

	if b.output != nil && (ds["check-only"].name != "" || wantError) {
		bad("a block that does not run has no output block")
	}
	main := b.name
	if main == "" {
		main = "main.ant"
		if err := os.WriteFile(filepath.Join(dir, main), []byte(before+b.src), 0o644); err != nil {
			return append(problems, err.Error())
		}
	} else if before != "" {
		bad("a named block takes no test: before")
	}
	isTest := strings.HasSuffix(main, "_test.ant")
	if strings.HasSuffix(main, ".ant") && !isTest {
		out, code := runDocAnt(bin, dir, "", "check", main)
		switch {
		case wantError && code == 0:
			bad("ant check passes, but test: error expects %q", ds["error"].arg)
		case wantError && !strings.Contains(out, ds["error"].arg):
			bad("ant check does not report %q:\n%s", ds["error"].arg, out)
		case !wantError && code != 0:
			bad("ant check:\n%s", out)
		}
	}
	if wantError || ds["check-only"].name != "" {
		return problems
	}
	var args []string
	if c, ok := ds["cmd"]; ok {
		f := strings.Fields(c.arg)
		if len(f) == 0 || f[0] != "ant" {
			bad("test: cmd runs an ant command, such as test: cmd ant test .")
			return problems
		}
		args = f[1:]
	} else if b.name != "" {
		return problems // a module, which the other blocks import
	} else {
		args = append([]string{"run", main}, strings.Fields(ds["args"].arg)...)
	}
	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(ds["stdin"].body)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			bad("ant %s: %v", strings.Join(args, " "), err)
			return problems
		}
		code = exit.ExitCode()
	}
	want := 0
	if e, ok := ds["exit"]; ok {
		fmt.Sscan(e.arg, &want)
	}
	if p, ok := ds["panics"]; ok {
		if !strings.Contains(stderr.String(), "panic: "+p.arg) {
			bad("the program does not panic with %q:\n%s", p.arg, stderr.String())
		}
	} else if code != want {
		bad("ant %s exits with %d, not %d:\n%s%s", strings.Join(args, " "), code, want, stdout.String(), stderr.String())
	}
	if b.output != nil && stdout.String() != b.output.src {
		bad("the output differs from the text block at line %d:\n--- got\n%s--- want\n%s", b.output.line, stdout.String(), b.output.src)
	}
	return problems
}

func runDocAnt(bin, dir, stdin string, args ...string) (string, int) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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

// linkProblems checks that every relative link of a page names a file under docs/.
func linkProblems(rel, src string) []string {
	var problems []string
	inFence := ""
	for i, line := range strings.Split(src, "\n") {
		if m := fenceOpen.FindStringSubmatch(line); m != nil && inFence == "" {
			inFence = m[1]
			continue
		}
		if inFence != "" {
			if strings.HasPrefix(line, inFence) && strings.Trim(strings.TrimSpace(line), inFence[:1]) == "" {
				inFence = ""
			}
			continue
		}
		for _, m := range mdLink.FindAllStringSubmatch(codeSpan.ReplaceAllString(line, ""), -1) {
			target := m[1]
			// A link from the site's root, such as /vscode, names a page of the site, not of docs/.
			if strings.Contains(target, "://") || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "/") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			target, _, _ = strings.Cut(target, "#")
			file := filepath.Join(docsRoot, filepath.Dir(rel), filepath.FromSlash(target))
			if _, err := os.Stat(file); err != nil {
				problems = append(problems, fmt.Sprintf("%s:%d: the link %s names no page", rel, i+1, m[1]))
			}
		}
	}
	return problems
}

// headingProblems checks the sections of a function, type, keyword, or operator page: only the
// headings of docSections, in their order, and Basic examples once the page is written.
func headingProblems(rel, src, status string) []string {
	leaf := strings.HasPrefix(rel, "lib/") && !strings.HasSuffix(rel, "/index.md") && !strings.HasPrefix(rel, "lib/guides/") ||
		strings.HasPrefix(rel, "language/") && strings.Count(rel, "/") == 2
	if !leaf {
		return nil
	}
	var problems []string
	at := -1
	basic := false
	inFence := ""
	for i, line := range strings.Split(src, "\n") {
		if m := fenceOpen.FindStringSubmatch(line); m != nil && inFence == "" {
			inFence = m[1]
			continue
		}
		if inFence != "" {
			if strings.HasPrefix(line, inFence) && strings.Trim(strings.TrimSpace(line), inFence[:1]) == "" {
				inFence = ""
			}
			continue
		}
		m := heading2.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		k := -1
		for j, s := range docSections {
			if s == m[1] {
				k = j
			}
		}
		switch {
		case k < 0:
			problems = append(problems, fmt.Sprintf("%s:%d: ## %s is not a section of the page format", rel, i+1, m[1]))
		case k <= at:
			problems = append(problems, fmt.Sprintf("%s:%d: ## %s is out of order", rel, i+1, m[1]))
		default:
			at = k
		}
		if m[1] == "Basic examples" {
			basic = true
		}
	}
	if !basic && status != "skeleton" && status != "" {
		problems = append(problems, fmt.Sprintf("%s:1: a %s page needs ## Basic examples", rel, status))
	}
	return problems
}
