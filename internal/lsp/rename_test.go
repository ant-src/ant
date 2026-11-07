package lsp

import (
	"strings"
	"testing"
)

func rename(c *client, uri string, p Position, name string) (WorkspaceEdit, *rpcError) {
	var edit WorkspaceEdit
	err := c.call("textDocument/rename", RenameParams{TextDocumentPositionParams: c.pos(uri, p), NewName: name}, &edit)
	return edit, err
}

func TestRenameNamespaceImport(t *testing.T) {
	main := "import \"./lib/geometry.ant\" Geo\n\nfunc f(): float64 {\n  var a: Geo.Point = Geo.origin\n  a.x\n}\n"
	c := start(t, map[string]string{"lib/geometry.ant": geometry, "main.ant": main})
	uri := c.open("main.ant", main)
	// the namespace is declared by the import, in main
	var locs []Location
	c.call("textDocument/definition", c.pos(uri, at(main, "Geo.origin", 1, 0)), &locs)
	if len(locs) != 1 || locs[0].URI != uri || locs[0].Range.Start != at(main, "Geo\n", 1, 0) {
		t.Errorf("definition of Geo: %+v", locs)
	}
	if _, err := rename(c, uri, at(main, "Geo.origin", 1, 0), "f"); err == nil || !strings.Contains(err.Message, "already declared") {
		t.Errorf("renaming Geo to f, a name of main: %v", err)
	}
	// distance is a name of the library only
	edit, err := rename(c, uri, at(main, "Geo.origin", 1, 0), "distance")
	if err != nil || len(edit.Changes[uri]) != 3 {
		t.Errorf("renaming Geo to distance: %v %+v", err, edit.Changes)
	}
	local := strings.Replace(main, "  var a:", "  var G = 1.0\n  var a:", 1)
	c.change(uri, 2, local)
	if _, err := rename(c, uri, at(local, "Geo.origin", 1, 0), "G"); err == nil || !strings.Contains(err.Message, "already declared") {
		t.Errorf("renaming Geo to the local G: %v", err)
	}
}

func TestRenameLambdaX(t *testing.T) {
	c := start(t, workspace())
	uri := c.open("main.ant", mainSrc)
	if _, err := rename(c, uri, at(mainSrc, "x * times", 1, 0), "y"); err == nil || !strings.Contains(err.Message, "lambda") {
		t.Errorf("renaming a lambda's x: %v", err)
	}
	if err := c.call("textDocument/prepareRename", c.pos(uri, at(mainSrc, "x * times", 1, 0)), nil); err == nil {
		t.Error("prepareRename on a lambda's x was allowed")
	}
}

func TestAliasReferences(t *testing.T) {
	main := "import \"./lib/geometry.ant\" { Point: P, distance }\n\nfunc f(a: P): P { P{ x: a.x, y: 0.0 } }\n\nvar q: Array[P] = []\n"
	want := "import \"./lib/geometry.ant\" { Point: Pt, distance }\n\nfunc f(a: Pt): Pt { Pt{ x: a.x, y: 0.0 } }\n\nvar q: Array[Pt] = []\n"
	c := start(t, map[string]string{"lib/geometry.ant": geometry, "main.ant": main})
	uri := c.open("main.ant", main)
	libURI := c.uri("lib/geometry.ant")
	// renaming the alias, from the import or from a type name written through it
	for _, p := range []Position{at(main, "P,", 1, 0), at(main, "a: P", 1, 3), at(main, "): P", 1, 3), at(main, "[P]", 1, 1)} {
		edit, err := rename(c, uri, p, "Pt")
		if err != nil {
			t.Fatalf("renaming P at %v: %v", p, err)
		}
		if got := applyAll(main, edit.Changes[uri]); got != want {
			t.Errorf("renaming P at %v:\n%s", p, got)
		}
		if _, touched := edit.Changes[libURI]; touched {
			t.Errorf("renaming the alias P at %v touched the library", p)
		}
	}
	// references and highlights of the alias
	rp := ReferenceParams{TextDocumentPositionParams: c.pos(uri, at(main, "a: P", 1, 3))}
	rp.Context.IncludeDeclaration = true
	var locs []Location
	c.call("textDocument/references", rp, &locs)
	if len(locs) != 5 {
		t.Errorf("references of P: %+v", locs)
	}
	var hl []DocumentHighlight
	c.call("textDocument/documentHighlight", c.pos(uri, at(main, "P{", 1, 0)), &hl)
	if len(hl) != 5 {
		t.Errorf("highlights of P: %+v", hl)
	}
}

func TestAliasUsesReferenced(t *testing.T) {
	c := start(t, workspace())
	uri := c.open("main.ant", mainSrc)
	rp := ReferenceParams{TextDocumentPositionParams: c.pos(uri, at(mainSrc, "dist(p", 1, 0))}
	rp.Context.IncludeDeclaration = true
	var locs []Location
	c.call("textDocument/references", rp, &locs)
	if len(locs) != 2 || locs[0].URI != uri || locs[1].Range.Start != at(mainSrc, "dist(p", 1, 0) {
		t.Errorf("references of dist: %+v", locs)
	}
	var hl []DocumentHighlight
	c.call("textDocument/documentHighlight", c.pos(uri, at(mainSrc, "dist(p", 1, 0)), &hl)
	if len(hl) != 2 {
		t.Errorf("highlights of dist: %+v", hl)
	}
	// renaming the export from the import's name renames the export, not the alias
	edit, err := rename(c, uri, at(mainSrc, "distance: dist", 1, 0), "far")
	if err != nil || applyAll(mainSrc, edit.Changes[uri]) != strings.Replace(mainSrc, "distance: dist", "far: dist", 1) {
		t.Errorf("renaming distance from the import: %v %+v", err, edit.Changes)
	}
}
