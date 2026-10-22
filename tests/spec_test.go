package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/prelude"
)

// The tests in this file keep the corpus traceable to the specification.
//
// Every test program starts with a tag line naming the heading or headings of the specification
// it proves, in the form "// spec: 4 Unions; 10; stdlib Json": a chapter number followed by a
// subsection heading, a chapter number alone for the chapter's own text, or "stdlib" and a
// namespace of stdlib.md. A program that triggers a runtime panic also carries
// "// panics: <key>" naming a row of the chapter 12 table (see panicRows) or a ◆ function of
// chapter 11 or of stdlib.md, such as "Array.slice" or "Time.plainDate".
//
// Programs under a directory named lib are modules the tests import, not tests, and carry no
// tags.

// specDir finds the specification checkout: the directory of the file ANT_SPEC names, or the
// sibling checkout of github.com/ant-src/spec.
func specDir(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("ANT_SPEC"); p != "" {
		return filepath.Dir(p)
	}
	p := filepath.Join("..", "..", "spec")
	if _, err := os.Stat(filepath.Join(p, "ant-lang.md")); err != nil {
		t.Skip("the specification is not checked out next to this repository; set ANT_SPEC")
	}
	return p
}

func readSpec(t *testing.T, name string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(specDir(t), name))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(string(data), "\n")
}

var (
	chapterLine = regexp.MustCompile(`^## (\d+)\. (.+)$`)
	sectionLine = regexp.MustCompile(`^### (.+)$`)
	stdlibLine  = regexp.MustCompile(`^## (\w+)$`)
)

// specHeadings returns the tag keys of every heading of the specification in order: "N" for a
// chapter, "N Title" for a subsection with backticks removed, "stdlib Name" for a namespace.
func specHeadings(t *testing.T) []string {
	t.Helper()
	var keys []string
	chapter := ""
	for _, line := range readSpec(t, "ant-lang.md") {
		if m := chapterLine.FindStringSubmatch(line); m != nil {
			chapter = m[1]
			keys = append(keys, chapter)
			continue
		}
		if m := sectionLine.FindStringSubmatch(line); m != nil && chapter != "" {
			keys = append(keys, chapter+" "+strings.ReplaceAll(m[1], "`", ""))
		}
	}
	for _, line := range readSpec(t, "stdlib.md") {
		if m := stdlibLine.FindStringSubmatch(line); m != nil {
			keys = append(keys, "stdlib "+m[1])
		}
	}
	return keys
}

// program is a test program with its tags.
type program struct {
	path   string
	src    string
	spec   []string
	panics []string
}

// corpus returns every test program under tests/ and examples/, with its tags, and the
// modules under lib directories with none.
func corpus(t *testing.T) (programs, modules []program) {
	t.Helper()
	var paths []string
	for _, root := range []string{".", filepath.Join("..", "examples")} {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".ant") {
				paths = append(paths, path)
			}
			return nil
		})
	}
	sort.Strings(paths)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		p := program{path: path, src: string(data)}
		isModule := false
		for _, elem := range strings.Split(filepath.ToSlash(path), "/") {
			if elem == "lib" {
				isModule = true
			}
		}
		for _, line := range strings.Split(p.src, "\n") {
			if !strings.HasPrefix(line, "//") {
				break
			}
			if rest, ok := strings.CutPrefix(line, "// spec:"); ok {
				p.spec = append(p.spec, splitTags(rest)...)
			}
			if rest, ok := strings.CutPrefix(line, "// panics:"); ok {
				p.panics = append(p.panics, splitTags(rest)...)
			}
		}
		if isModule {
			if p.spec != nil || p.panics != nil {
				t.Errorf("%s: a module under lib carries tags", path)
			}
			modules = append(modules, p)
			continue
		}
		if len(p.spec) == 0 {
			t.Errorf("%s: no \"// spec:\" tag", path)
		}
		programs = append(programs, p)
	}
	if len(programs) == 0 {
		t.Fatal("no test programs found")
	}
	return programs, modules
}

func splitTags(s string) []string {
	var tags []string
	for _, tag := range strings.Split(s, ";") {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// checkPending reports a required key that is neither covered nor pending, and a pending key
// that is covered, since the pending list must shrink as the corpus grows.
func checkPending(t *testing.T, what string, required []string, covered map[string]bool, pending []string) {
	t.Helper()
	isPending := make(map[string]bool, len(pending))
	for _, k := range pending {
		isPending[k] = true
	}
	for _, k := range required {
		switch {
		case !covered[k] && !isPending[k]:
			t.Errorf("no test program names %s %q", what, k)
		case covered[k] && isPending[k]:
			t.Errorf("%s %q is covered now; remove it from the pending list", what, k)
		}
	}
	for _, k := range pending {
		known := false
		for _, r := range required {
			if r == k {
				known = true
			}
		}
		if !known {
			t.Errorf("pending %s %q is not in the specification", what, k)
		}
	}
}

// pendingSpec lists the headings no test program names yet.
var pendingSpec = []string{}

// TestSpecCoverage checks that every tag names a heading of the specification and that every
// heading has a tag. Under -v it prints the matrix.
func TestSpecCoverage(t *testing.T) {
	headings := specHeadings(t)
	known := make(map[string]bool, len(headings))
	for _, h := range headings {
		known[h] = true
	}
	programs, _ := corpus(t)
	count := make(map[string]int)
	for _, p := range programs {
		for _, tag := range p.spec {
			if !known[tag] {
				t.Errorf("%s: tag %q names no heading of the specification", p.path, tag)
				continue
			}
			count[tag]++
			if chapter, _, ok := strings.Cut(tag, " "); ok && chapter != "stdlib" {
				count[chapter]++
			}
		}
	}
	var required []string
	for _, h := range headings {
		if h != "1" {
			required = append(required, h)
		}
	}
	covered := make(map[string]bool)
	for k, n := range count {
		covered[k] = n > 0
	}
	checkPending(t, "heading", required, covered, pendingSpec)
	if testing.Verbose() {
		for _, h := range headings {
			t.Logf("%4d  %s", count[h], h)
		}
	}
}

// pendingLibrary lists the prelude functions and constants no test program uses yet.
var pendingLibrary = []string{}

// TestLibraryCoverage checks that every function and constant the prelude declares is used by
// some test program.
func TestLibraryCoverage(t *testing.T) {
	decls, err := prelude.Decls()
	if err != nil {
		t.Fatal(err)
	}
	programs, modules := corpus(t)
	var all strings.Builder
	for _, p := range append(programs, modules...) {
		all.WriteString(p.src)
		all.WriteString("\n")
	}
	src := all.String()
	var required []string
	covered := make(map[string]bool)
	for _, d := range decls.Decls {
		var name string
		var pattern string
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Owner == "" {
				name = d.Name
				pattern = `\b` + regexp.QuoteMeta(d.Name) + `\s*\(`
			} else {
				name = d.Owner + "." + d.Name
				pattern = `\b` + regexp.QuoteMeta(name) + `\b`
			}
		case *ast.VarDecl:
			target, ok := d.Target.(*ast.NameTarget)
			if !ok || d.Owner == "" {
				continue
			}
			name = d.Owner + "." + target.Name
			pattern = `\b` + regexp.QuoteMeta(name) + `\b`
		default:
			continue
		}
		required = append(required, name)
		covered[name] = regexp.MustCompile(pattern).MatchString(src)
	}
	checkPending(t, "library name", required, covered, pendingLibrary)
}

// panicRows maps each row of the chapter 12 panics table, in order, to the keys a program
// names in its "// panics:" tag. A row with a kind part and a ◆ value part has one key for
// each.
var panicRows = [][]string{
	{"arith-kind"},                           // + - * / %, comparisons, unary -: kinds or tags differ
	{"bool-kind"},                            // ! && || if while: not a bool
	{"int-overflow"},                         // ◆ integer + - *, compound assignment, Array.sum
	{"div-zero"},                             // ◆ integer / %
	{"shift"},                                // ◆ Bits.shiftLeft, Bits.shiftRight
	{"conversion"},                           // ◆ numeric conversion
	{"string-bool"},                          // string, bool on the wrong kind
	{"field"},                                // a.b, a.b = v
	{"index", "index-range"},                 // a[i], a[i] = v: not a container; ◆ out of range
	{"call"},                                 // call: arity, names, non-function, constructor
	{"record-literal"},                       // record literal
	{"array-literal"},                        // array literal
	{"is"},                                   // is: right side is not a tag or constructor
	{"for", "for-yield"},                     // for: not iterable; ◆ yield after false
	{"destructuring", "destructuring-short"}, // destructuring: wrong kind; ◆ too few elements
	{"variable"},                             // no scope binds the name
	{"underscore"},                           // _ read
	{"library-kind", "library-value"},        // library function: wrong kind; ◆ documented case
	{"var-init"},                             // ◆ var read before its initializer has run
	{"panic"},                                // ◆ panic
	{"depth"},                                // ◆ call depth
	{"try"},                                  // try on a non-Result
	{"select-closed"},                        // ◆ select: a send clause chosen on a closed channel
	{"never"},                                // ??, ?., ==, != never panic
}

// pendingPanics lists the panic keys no test program triggers yet.
var pendingPanics = []string{}

// TestBoundedSleeps checks the corpus's rule that every wait has a bound: a Time.sleep with a
// literal count above a second and no context would hold the suite, so none is allowed under
// tests/run. A sleep with a context is bounded by it, whatever its count.
func TestBoundedSleeps(t *testing.T) {
	long := regexp.MustCompile(`Time\.sleep\((\d+)\)`)
	programs, _ := corpus(t)
	for _, p := range programs {
		for _, m := range long.FindAllStringSubmatch(p.src, -1) {
			if len(m[1]) > 4 || len(m[1]) == 4 && m[1] > "1000" {
				t.Errorf("%s: Time.sleep(%s) without a context would hold the suite; bound it with a context", p.path, m[1])
			}
		}
	}
}

// panicFunctions returns the ◆ functions of a specification file's code blocks, in order: the
// signatures whose comment carries ◆ and whose name has an owner. In ant-lang.md only chapter 11
// is read, so that the conversions (int8 and "the other nine") stay the table's conversion row; in
// stdlib.md every block is. A signature wrapped over several lines continues on lines that start
// with white space, as Time.with does, and its ◆ may sit on any of them, so the lines of a
// signature are joined before its name is read.
func panicFunctions(t *testing.T, file string) []string {
	t.Helper()
	var names []string
	known := make(map[string]bool)
	signature := ""
	flush := func() {
		if strings.Contains(signature, "◆") {
			name := signature
			if i := strings.IndexAny(name, "[("); i >= 0 {
				name = name[:i]
			}
			if strings.Contains(name, ".") && !strings.ContainsAny(name, " \t") && !known[name] {
				known[name] = true
				names = append(names, name)
			}
		}
		signature = ""
	}
	inChapter := file != "ant-lang.md"
	inBlock := false
	for _, line := range readSpec(t, file) {
		if file == "ant-lang.md" {
			if strings.HasPrefix(line, "## 11.") {
				inChapter = true
			} else if strings.HasPrefix(line, "## 12.") {
				inChapter = false
			}
		}
		if !inChapter {
			continue
		}
		if strings.HasPrefix(line, "```") {
			flush()
			inBlock = !inBlock
			continue
		}
		if !inBlock {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			signature += " " + strings.TrimSpace(line)
			continue
		}
		flush()
		signature = line
	}
	flush()
	return names
}

// TestPanicCoverage checks that every row of the chapter 12 panics table and every ◆ function
// of chapter 11 and of stdlib.md has a program that triggers it.
func TestPanicCoverage(t *testing.T) {
	lines := readSpec(t, "ant-lang.md")
	var required []string
	known := make(map[string]bool)
	for _, file := range []string{"ant-lang.md", "stdlib.md"} {
		names := panicFunctions(t, file)
		if len(names) == 0 {
			t.Fatalf("no ◆ functions found in %s", file)
		}
		for _, name := range names {
			if !known[name] {
				known[name] = true
				required = append(required, name)
			}
		}
	}
	// Rows of the panics table.
	rows := 0
	inTable := false
	for _, line := range lines {
		if strings.HasPrefix(line, "### Runtime panics") {
			inTable = true
			continue
		}
		if !inTable {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			break
		}
		if strings.HasPrefix(line, "| ") && !strings.HasPrefix(line, "| Operation") && !strings.HasPrefix(line, "| ---") {
			rows++
		}
	}
	if rows != len(panicRows) {
		t.Fatalf("the panics table has %d rows, panicRows has %d; update panicRows", rows, len(panicRows))
	}
	for _, keys := range panicRows {
		for _, k := range keys {
			known[k] = true
			required = append(required, k)
		}
	}
	programs, _ := corpus(t)
	covered := make(map[string]bool)
	for _, p := range programs {
		for _, k := range p.panics {
			if !known[k] {
				t.Errorf("%s: panic key %q is neither a table row nor a ◆ function", p.path, k)
				continue
			}
			covered[k] = true
		}
	}
	checkPending(t, "panic", required, covered, pendingPanics)
}

// TestUncheckedRejected checks that ant check rejects every program under unchecked/: they are
// the kind panics the checker rules out, so a clean check would be a soundness bug.
func TestUncheckedRejected(t *testing.T) {
	paths, _ := filepath.Glob("unchecked/*.ant")
	if len(paths) == 0 {
		t.Fatal("no programs under unchecked/")
	}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".ant")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			abs, _ := filepath.Abs(path)
			prog, err := ant.Load(abs)
			if err != nil {
				return // a load error is a rejection too
			}
			if len(prog.Check()) == 0 {
				t.Errorf("%s checks clean but panics at run time", path)
			}
		})
	}
}
