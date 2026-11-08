package tests

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/format"
	"github.com/ant-src/ant/internal/prelude"
)

// TestDocsReference keeps the standard library reference under ../docs/lib in step with
// internal/prelude/prelude.d.ant: one page per declaration, no page for anything else, and on
// each page a generated part, the title, signature, and doc comment, that matches the prelude.
// With ANT_DOCS_UPDATE=1 it writes instead: a skeleton for every missing page, and the generated
// part of every existing one. Everything after the generated part is the writers' and is never
// touched.

const (
	genStart = "<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->"
	genEnd   = "<!-- end generated -->"
)

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

// docPage is a page of the reference: its path under docs/, its title, the namespace doc for a
// namespace page, and the declarations it documents.
type docPage struct {
	path  string
	title string
	nsDoc string
	decls []ast.Decl
}

// referencePages maps every page the prelude needs to what it documents, and every documented
// name to its page. A function is lib/<Namespace>/<name>.md, or lib/prelude/<name>.md without a
// namespace; a type is lib/<Namespace>/types/<Name>.md, so that Time.Zone and Time.zone are two
// files on a file system that ignores case; a namespace, and a type named like one, is
// lib/<Namespace>/index.md.
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

func sigDoc(d ast.Decl) (string, string) {
	switch d := d.(type) {
	case *ast.FuncDecl:
		return d.Sig, d.Doc
	case *ast.TypeDecl:
		return d.Sig, d.Doc
	}
	return "", ""
}

// docLink is a link of a doc comment, a name in brackets (doc.md, section 2). A bracket right
// after a name or a bracket, as in Array[int] or Map[K][V], holds type arguments, not a link.
var docLink = regexp.MustCompile(`(^|[^\w\]])\[([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)?)\]`)

// docProse renders a paragraph of a doc comment as Markdown: a link to a declaration with a page
// becomes a link to it, in code, and the characters Markdown or the site would read, as in a * b
// or <p>, are escaped. A bracketed text that names nothing stays as written.
func docProse(text, from string, names map[string]string) string {
	var b strings.Builder
	last := 0
	for _, m := range docLink.FindAllStringSubmatchIndex(text, -1) {
		name := text[m[4]:m[5]]
		target, ok := names[name]
		if !ok {
			continue
		}
		b.WriteString(escapeDocMarkdown(text[last:m[3]]))
		if target == from {
			b.WriteString("`" + name + "`")
		} else {
			rel, _ := filepath.Rel(filepath.Dir(from), target)
			b.WriteString("[`" + name + "`](" + filepath.ToSlash(rel) + ")")
		}
		last = m[1]
	}
	b.WriteString(escapeDocMarkdown(text[last:]))
	return b.String()
}

// escapeDocMarkdown escapes the characters that would format plain text, as the language server
// does for a hover.
func escapeDocMarkdown(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '`', '*', '_', '[', ']', '<', '>', '#', '|', '~', '{', '}':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// splitDoc separates a doc comment's paragraphs from its examples. A code line is indented by at
// least one space beyond the comment's own (doc.md, section 2); a run of them is one example,
// without the indent the run shares. A code line that is only "// value" gives the value of the
// line before it.
func splitDoc(doc string) ([]string, [][]string) {
	var paras []string
	var examples [][]string
	var para, code []string
	flushPara := func() {
		if len(para) > 0 {
			paras = append(paras, strings.Join(para, " "))
			para = nil
		}
	}
	flushCode := func() {
		for len(code) > 0 && code[len(code)-1] == "" {
			code = code[:len(code)-1]
		}
		if len(code) == 0 {
			return
		}
		indent := -1
		for _, l := range code {
			n := len(l) - len(strings.TrimLeft(l, " "))
			if l != "" && (indent < 0 || n < indent) {
				indent = n
			}
		}
		var ex []string
		for _, l := range code {
			l = l[min(indent, len(l)):]
			if v, ok := strings.CutPrefix(strings.TrimSpace(l), "// "); ok && strings.HasPrefix(l, " ") && len(ex) > 0 {
				ex[len(ex)-1] += " // " + v
				continue
			}
			ex = append(ex, l)
		}
		examples = append(examples, ex)
		code = nil
	}
	for _, line := range strings.Split(doc, "\n") {
		switch {
		case strings.HasPrefix(line, " "):
			flushPara()
			code = append(code, line)
		case strings.TrimSpace(line) == "":
			// A blank line ends a paragraph, but not an example: the next code line continues it.
			flushPara()
			if len(code) > 0 {
				code = append(code, "")
			}
		default:
			flushCode()
			para = append(para, line)
		}
	}
	flushPara()
	flushCode()
	return paras, examples
}

// generated is the part of a reference page that comes from the prelude: the title, each
// declaration's signature, and its doc without the examples, which the page shows as cases.
func generated(p *docPage, names map[string]string) string {
	var b strings.Builder
	b.WriteString(genStart + "\n\n# " + p.title + "\n")
	prose := func(doc string) {
		paras, _ := splitDoc(doc)
		for _, para := range paras {
			b.WriteString("\n" + docProse(para, p.path, names) + "\n")
		}
	}
	prose(p.nsDoc)
	for _, d := range p.decls {
		sig, doc := sigDoc(d)
		b.WriteString("\n```ant signature\n" + sig + "\n```\n")
		prose(doc)
	}
	b.WriteString("\n" + genEnd + "\n")
	return b.String()
}

// skeleton is a new reference page: the status line, the generated part, and the doc's examples
// with cases as Basic examples.
func skeleton(p *docPage, names map[string]string) string {
	var b strings.Builder
	b.WriteString("<!-- status: skeleton -->\n" + generated(p, names))
	var blocks []string
	for _, d := range p.decls {
		_, doc := sigDoc(d)
		_, examples := splitDoc(doc)
		for _, ex := range examples {
			src := strings.Join(ex, "\n") + "\n"
			if _, isCases := caseInputs(src); !isCases {
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

func TestDocsReference(t *testing.T) {
	pages, names := referencePages(t)
	update := os.Getenv("ANT_DOCS_UPDATE") == "1"
	paths := make([]string, 0, len(pages))
	for path := range pages {
		paths = append(paths, path)
	}
	sort.Strings(paths)
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
		want := generated(p, names)
		// oxfmt rewraps the prose of every Markdown file, so only the words count.
		if strings.Join(strings.Fields(src[i:end]), " ") == strings.Join(strings.Fields(want), " ") {
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
	err := filepath.WalkDir(filepath.Join(docsRoot, "lib"), func(file string, d os.DirEntry, err error) error {
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
	if err != nil {
		t.Fatal(err)
	}
}

// TestDocsGenerator checks the parts of the generator that the prelude does not exercise today.
func TestDocsGenerator(t *testing.T) {
	names := map[string]string{"int": "lib/prelude/int.md", "Array.filter": "lib/Array/filter.md", "Array.map": "lib/Array/map.md"}
	for _, c := range []struct{ in, want string }{
		{"see [Array.filter]", "see [`Array.filter`](filter.md)"},
		{"[Array.filter] first", "[`Array.filter`](filter.md) first"},
		{"an Array[int] value", `an Array\[int\] value`},
		{"Map[K][int]", `Map\[K\]\[int\]`},
		{"itself, [Array.map]", "itself, `Array.map`"},
		{"[nothing] here", `\[nothing\] here`},
		{"a * b and <p>{x}</p>", `a \* b and \<p\>\{x\}\</p\>`},
	} {
		if got := docProse(c.in, "lib/Array/map.md", names); got != c.want {
			t.Errorf("docProse(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	paras, examples := splitDoc("First line\nsame paragraph.\n\n  var xs = [1]\n\n  f(xs)\n    // 2\n  g()   // 3\n\nAfter.\n\n h()")
	if strings.Join(paras, "|") != "First line same paragraph.|After." {
		t.Errorf("paragraphs: %q", paras)
	}
	if len(examples) != 2 || strings.Join(examples[0], "|") != "var xs = [1]||f(xs) // 2|g()   // 3" || strings.Join(examples[1], "|") != "h()" {
		t.Errorf("examples: %q", examples)
	}
}
