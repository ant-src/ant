package lsp

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ant-src/ant/ast"
)

// The checker's messages for a name that nothing declares: a value, or where a type is written.
var (
	notDefined  = regexp.MustCompile(`^(\S+) is not defined$`)
	unknownType = regexp.MustCompile(`^unknown type (\S+)$`)
)

// export is a top-level declaration a module exports, which another module may import.
type export struct {
	name string
	kind int // the completion item kind
	file *workspaceFile
}

// importable lists the exports of the workspace's modules that the module at path may import
// (spec/lsp.md section 4, Imports): not its own, nor a script's or a declaration file's, nor a
// test file's unless the module is a test file too, nor those of a module that imports it,
// directly or through others, since importing that would make a cycle. Only the names that
// keep reports are listed.
func (s *Server) importable(path string, keep func(name string) bool) []export {
	files := s.workspaceFiles()
	test := strings.HasSuffix(path, "_test.ant")
	// The modules each module imports, by path, to find those that reach path.
	importers := map[string][]string{}
	for _, f := range files {
		dir := filepath.Dir(f.doc.path)
		for _, imp := range f.file.Imports {
			target := filepath.Clean(filepath.Join(dir, filepath.FromSlash(imp.Path)))
			importers[target] = append(importers[target], f.doc.path)
		}
	}
	reaches := map[string]bool{path: true}
	queue := []string{path}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, q := range importers[p] {
			if !reaches[q] {
				reaches[q] = true
				queue = append(queue, q)
			}
		}
	}
	var out []export
	for _, f := range files {
		if reaches[f.doc.path] || f.file.Script || f.file.Declarations || strings.HasSuffix(f.doc.path, "_test.ant") && !test {
			continue
		}
		for _, decl := range f.file.Decls {
			var name string
			kind := 0
			switch x := decl.(type) {
			case *ast.FuncDecl:
				if x.Export && x.Owner == "" {
					name, kind = x.Name, ItemFunction
				}
			case *ast.TypeDecl:
				if x.Export && x.Owner == "" {
					name, kind = x.Name, ItemStruct
				}
			case *ast.VarDecl:
				if nt, ok := x.Target.(*ast.NameTarget); ok && x.Export && x.Owner == "" {
					name, kind = nt.Name, ItemVariable
				}
			}
			if name != "" && keep(name) {
				out = append(out, export{name: name, kind: kind, file: f})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].name != out[j].name {
			return out[i].name < out[j].name
		}
		return out[i].file.doc.path < out[j].file.doc.path
	})
	return out
}

// importPath is the path an import in the module at from writes for the module at to: relative,
// with forward slashes, and starting with ./ or ../ (section 10).
func importPath(from, to string) string {
	rel, err := filepath.Rel(filepath.Dir(from), to)
	if err != nil {
		return filepath.ToSlash(to)
	}
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, "../") {
		rel = "./" + rel
	}
	return rel
}

// importUse is how the module of d comes to use the export e at the name written over use: the
// title of the change, its edits, and whether the use is qualified instead of imported. An
// import of e's module that binds it as a namespace makes the use Namespace.name; one that
// destructures it gets the name added to its list; otherwise a new import follows the last one,
// or the module's doc comment when there is none.
func (s *Server) importUse(d *document, f *ast.File, e export, use Range) (string, []TextEdit, bool) {
	target := e.file.doc.path
	dir := filepath.Dir(d.path)
	var list *ast.Import
	for _, imp := range f.Imports {
		if filepath.Clean(filepath.Join(dir, filepath.FromSlash(imp.Path))) != target {
			continue
		}
		if imp.Namespace != "" {
			return "Use " + imp.Namespace + "." + e.name, []TextEdit{{Range: use, NewText: imp.Namespace + "." + e.name}}, true
		}
		if list == nil && len(imp.Names) > 0 {
			list = imp
		}
	}
	path := importPath(d.path, target)
	title := "Import " + e.name + " from \"" + path + "\""
	if list != nil {
		at := d.position(d.offsetOf(list.Names[len(list.Names)-1].End), s.utf16)
		return title, []TextEdit{{Range: Range{Start: at, End: at}, NewText: ", " + e.name}}, false
	}
	line := "import \"" + path + "\" { " + e.name + " }"
	if n := len(f.Imports); n > 0 {
		at := d.position(d.offsetOf(f.Imports[n-1].End), s.utf16)
		return title, []TextEdit{{Range: Range{Start: at, End: at}, NewText: "\n" + line}}, false
	}
	// Past a #! line and the //! lines, and the blank lines after them.
	off := 0
	for off < len(d.text) {
		end := strings.IndexByte(d.text[off:], '\n')
		if end < 0 {
			end = len(d.text) - off
		}
		l := strings.TrimSpace(d.text[off : off+end])
		if !(off == 0 && strings.HasPrefix(l, "#!")) && !strings.HasPrefix(l, "//!") && (l != "" || off == 0) {
			break
		}
		off += end + 1
	}
	if off > len(d.text) {
		off = len(d.text)
	}
	text := line + "\n"
	if off < len(d.text) {
		text += "\n"
	}
	if off > 0 && !strings.HasSuffix(d.text[:off], "\n\n") {
		text = "\n" + text
	}
	at := d.position(off, s.utf16)
	return title, []TextEdit{{Range: Range{Start: at, End: at}, NewText: text}}, false
}

// importActions are the fixes for a name that is not defined: one per module of the workspace
// that exports it, only a type where a type is written, preferred when there is only one.
func (s *Server) importActions(a *analysis, d *document, name string, types bool, dg Diagnostic) []CodeAction {
	if d.path == "" {
		return nil // a document not on disk has no path to import relative to
	}
	var exports []export
	for _, e := range s.importable(d.path, func(n string) bool { return n == name }) {
		if !types || e.kind == ItemStruct {
			exports = append(exports, e)
		}
	}
	var out []CodeAction
	for _, e := range exports {
		title, edits, _ := s.importUse(d, a.module().File, e, dg.Range)
		out = append(out, CodeAction{Title: title, Kind: "quickfix", Diagnostics: []Diagnostic{dg}, IsPreferred: len(exports) == 1,
			Edit: &WorkspaceEdit{Changes: map[string][]TextEdit{d.uri: edits}}})
	}
	return out
}

// importNamesOutOfScope offers the exports of the modules the module may import whose names start with
// the word under the cursor, past the names in scope, each with the import it needs; only types
// where a type is written. With no word typed it offers none, and the list is incomplete so
// that the editor asks again as the word grows.
func (c *completer) importNamesOutOfScope(types bool) {
	word := strings.ToLower(c.d.text[c.wordStart:c.off])
	if word == "" || c.d.path == "" {
		return
	}
	c.incomplete = true
	inScope := map[string]bool{}
	for _, item := range c.items {
		inScope[item.Label] = true
	}
	exports := c.s.importable(c.d.path, func(n string) bool {
		return !inScope[n] && strings.HasPrefix(strings.ToLower(n), word)
	})
	start, end := c.d.position(c.wordStart, c.s.utf16), c.d.position(c.wordEnd, c.s.utf16)
	for _, e := range exports {
		if types && e.kind != ItemStruct {
			continue
		}
		path := importPath(c.d.path, e.file.doc.path)
		item := CompletionItem{Label: e.name, Kind: e.kind, LabelDetails: &CompletionLabelDetails{Description: path},
			Detail: "import from \"" + path + "\"", SortText: "~" + e.name}
		_, edits, qualified := c.s.importUse(c.d, c.a.module().File, e, Range{Start: start, End: end})
		if qualified {
			item.TextEdit = &edits[0]
			item.FilterText = e.name
		} else {
			item.AdditionalTextEdits = edits
		}
		c.add(item)
	}
}
