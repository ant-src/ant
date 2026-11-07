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
