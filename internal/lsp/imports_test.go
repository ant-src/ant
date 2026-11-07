package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixes and completions that import a name from another module of the workspace.
func TestImports(t *testing.T) {
	files := map[string]string{
		"lib/math.ant":     "export func sum(a: int, b: int): int { a + b }\n\nexport type Pair = { a: int, b: int }\n\nfunc sumPrivate(): int { 1 }\n",
		"lib/other.ant":    "export func sum(a: int): int { a }\n",
		"lib/cycle.ant":    "import \"../sum_test.ant\" { helper }\n\nexport func summit(): int { 1 }\n",
		"script.ant":       "export func sumScript(): int { 1 }\nprintln(1)\n",
		"helpers_test.ant": "export func sumHelper(): int { 1 }\n",
	}
	c := start(t, files)
	fixes := func(uri, src string) []CodeAction {
		t.Helper()
		diags := c.diagnostics(uri)
		p := CodeActionParams{TextDocument: TextDocumentIdentifier{URI: uri}, Range: Range{End: at(src, src[len(src)-1:], strings.Count(src, src[len(src)-1:]), 1)}}
		p.Context.Diagnostics = diags
		var acts []CodeAction
		c.call("textDocument/codeAction", p, &acts)
		return acts
	}
	titles := func(acts []CodeAction) string {
		var out []string
		for _, a := range acts {
			out = append(out, a.Title)
		}
		return strings.Join(out, "; ")
	}

	// no import yet: one fix per module that exports the name, none preferred
	src := "//! Sums.\n\nprintln(sum(1, 2))\n"
	uri := c.open("main.ant", src)
	acts := fixes(uri, src)
	if got := titles(acts); got != `Import sum from "./lib/math.ant"; Import sum from "./lib/other.ant"` {
		t.Fatalf("fixes: %s", got)
	}
	if acts[0].IsPreferred {
		t.Errorf("one of two candidates is preferred")
	}
	if got := applyAll(src, acts[0].Edit.Changes[uri]); got != "//! Sums.\n\nimport \"./lib/math.ant\" { sum }\n\nprintln(sum(1, 2))\n" {
		t.Errorf("after the fix:\n%s", got)
	}

	// a list to add to, and a type
	src = "import \"./lib/math.ant\" { sum }\n\nvar p: Pair = Pair{ a: sum(1, 2), b: 0 }\nprintln(p)\n"
	c.change(uri, 2, src)
	acts = fixes(uri, src)
	if got := titles(acts); got != `Import Pair from "./lib/math.ant"` {
		t.Errorf("one fix for Pair, as a type and as a tag: %s", got)
	}
	if len(acts) == 0 || !acts[0].IsPreferred || !strings.HasPrefix(applyAll(src, acts[0].Edit.Changes[uri]), "import \"./lib/math.ant\" { sum, Pair }\n") {
		t.Errorf("fixes for Pair: %s %+v", titles(acts), acts)
	}

	// a namespace import qualifies the use
	src = "import \"./lib/math.ant\" M\n\nprintln(sum(1, 2))\n"
	c.change(uri, 3, src)
	acts = fixes(uri, src)
	if len(acts) == 0 || acts[0].Title != "Use M.sum" || applyAll(src, acts[0].Edit.Changes[uri]) != "import \"./lib/math.ant\" M\n\nprintln(M.sum(1, 2))\n" {
		t.Errorf("fixes with a namespace: %s", titles(acts))
	}

	// a private name is no module's to import but a test file's
	src = "println(sumPrivate())\n"
	c.change(uri, 4, src)
	if acts := fixes(uri, src); len(acts) != 0 {
		t.Errorf("fixes for a private name: %s", titles(acts))
	}

	// completion: names by the word typed, past what is in scope, each with its import; in a test
	// file private ones and a test file's too; no script's, nothing that would make a cycle
	test := "test \"sums\" {\n  assert(su\n}\n"
	testURI := c.open("sum_test.ant", test)
	var list CompletionList
	c.call("textDocument/completion", CompletionParams{TextDocumentPositionParams: c.pos(testURI, at(test, "su\n", 1, 2))}, &list)
	if !list.IsIncomplete {
		t.Errorf("the list depends on the word")
	}
	var labels []string
	for _, item := range list.Items {
		if item.LabelDetails != nil && item.LabelDetails.Description != "" {
			labels = append(labels, item.Label+" "+item.LabelDetails.Description)
		}
	}
	if got := strings.Join(labels, ", "); got != "sum ./lib/math.ant, sum ./lib/other.ant, sumHelper ./helpers_test.ant, sumPrivate ./lib/math.ant" {
		t.Errorf("importable items: %s", got)
	}
	for _, item := range list.Items {
		if item.Label == "sum" && item.LabelDetails != nil && item.LabelDetails.Description == "./lib/math.ant" {
			if got := applyAll(test, item.AdditionalTextEdits); got != "import \"./lib/math.ant\" { sum }\n\n"+test {
				t.Errorf("the item's import:\n%s", got)
			}
		}
	}
	// where a type is written, only types
	src = "var p: Pa\n"
	c.change(uri, 5, src)
	list = CompletionList{}
	c.call("textDocument/completion", CompletionParams{TextDocumentPositionParams: c.pos(uri, at(src, "Pa\n", 1, 2))}, &list)
	found := false
	for _, item := range list.Items {
		if item.Label == "Pair" && len(item.AdditionalTextEdits) == 1 {
			found = true
		}
	}
	if !found {
		t.Errorf("no Pair to import where a type is written: %+v", list.Items)
	}

	// a test file imports a private name, and the checker takes it
	test = "test \"private\" {\n  assert(sumPrivate() == 1)\n}\n"
	c.change(testURI, 2, test)
	acts = fixes(testURI, test)
	if len(acts) != 1 || acts[0].Title != `Import sumPrivate from "./lib/math.ant"` {
		t.Fatalf("fixes in a test file: %s", titles(acts))
	}
	test = applyAll(test, acts[0].Edit.Changes[testURI])
	c.change(testURI, 3, test)
	if diags := c.diagnostics(testURI); len(diags) != 0 {
		t.Errorf("after the import: %+v", diags)
	}
}

// A walk of the workspace folders stands while the client watches the files and reports no
// change; without a watching client, each walk sees the files as they are.
func TestWalkWhileWatching(t *testing.T) {
	dir := t.TempDir()
	write := func(name string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("export func f(): int { 1 }\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	names := func(files []*workspaceFile) string {
		var out []string
		for _, f := range files {
			out = append(out, filepath.Base(f.doc.path))
		}
		return strings.Join(out, " ")
	}
	write("a.ant")
	s := &Server{docs: map[string]*document{}, roots: []string{dir}}
	if got := names(s.walk()); got != "a.ant" {
		t.Fatalf("first walk: %s", got)
	}
	write("b.ant")
	if got := names(s.walk()); got != "a.ant b.ant" {
		t.Errorf("without watching, a walk sees a new file: %s", got)
	}
	s.watching = true
	s.walk()
	write("c.ant")
	if got := names(s.walk()); got != "a.ant b.ant" {
		t.Errorf("while watching, the last walk stands: %s", got)
	}
	s.diskGen++ // workspace/didChangeWatchedFiles
	if got := names(s.walk()); got != "a.ant b.ant c.ant" {
		t.Errorf("after a change, a walk sees it: %s", got)
	}
}

// TestImportItems: the list is incomplete before a word is typed, so that the importable names
// come as it grows, and a function's item inserts a call as other function items do, qualified
// through a namespace import too, with its signature and doc.
func TestImportItems(t *testing.T) {
	var snippets ClientCapabilities
	snippets.TextDocument.Completion.CompletionItem.SnippetSupport = true
	files := map[string]string{
		"lib/math.ant": "// sum adds two numbers. It is exact.\nexport func sum(a: int, b: int): int { a + b }\n\nexport func one(): int { 1 }\n",
	}
	c := startWith(t, files, snippets)
	complete := func(uri, src, needle string) CompletionList {
		t.Helper()
		var list CompletionList
		c.call("textDocument/completion", CompletionParams{TextDocumentPositionParams: c.pos(uri, at(src, needle, 1, len(needle)))}, &list)
		return list
	}
	find := func(list CompletionList, label string) *CompletionItem {
		for i := range list.Items {
			if it := &list.Items[i]; it.Label == label && it.LabelDetails != nil && it.LabelDetails.Description != "" {
				return it
			}
		}
		return nil
	}
	src := "var v = \n"
	uri := c.open("main.ant", src)
	if list := complete(uri, src, "= "); !list.IsIncomplete {
		t.Errorf("with no word typed, the list is complete")
	}
	src = "var v = su\n"
	c.change(uri, 2, src)
	list := complete(uri, src, "su")
	it := find(list, "sum")
	if it == nil {
		t.Fatalf("no sum to import: %+v", list.Items)
	}
	if it.InsertText != "sum($0)" || it.InsertTextFormat != 2 || it.LabelDetails.Detail != "(a: int, b: int): int" || it.Documentation == nil || !strings.Contains(it.Documentation.Value, "sum adds two numbers.") {
		t.Errorf("sum's item: %+v %+v", it, it.LabelDetails)
	}
	src = "var v = on\n"
	c.change(uri, 3, src)
	if it := find(complete(uri, src, "on"), "one"); it == nil || it.InsertText != "one()" {
		t.Errorf("one's item: %+v", it)
	}
	// through a namespace import, the edit writes the qualified call
	src = "import \"./lib/math.ant\" M\n\nvar v = su\n"
	c.change(uri, 4, src)
	list = complete(uri, src, "= su")
	it = find(list, "sum")
	if it == nil || it.TextEdit == nil || it.TextEdit.NewText != "M.sum($0)" || it.InsertTextFormat != 2 {
		t.Errorf("sum's qualified item: %+v", it)
	}
}
