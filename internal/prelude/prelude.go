package prelude

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/parser"
)

//go:embed prelude.d.ant
var source string

// Path is the prelude's file name, a declaration file (section 10, Declaration files): the path
// its declarations carry and the document an editor opens for one, ant:/prelude.d.ant.
const Path = "prelude.d.ant"

var (
	once sync.Once
	file *ast.File
	err  error
)

// Source is the text of prelude.d.ant, which an editor shows for a prelude declaration.
func Source() string { return source }

// Layout returns the prelude's layout, for an editor that shows its document, made once.
var Layout = sync.OnceValue(func() *parser.Layout {
	_, l, _ := parser.ParseLayout(Path, source)
	return l
})

// Decls returns the prelude's declarations, parsed once.
func Decls() (*ast.File, error) {
	once.Do(func() { file, err = parser.Parse(Path, source, parser.Declarations) })
	return file, err
}

// NamespaceDoc returns the namespace name, matched without regard to case, as the prelude spells
// it, with its doc, or "", "" when it has none. A namespace's doc is the paragraph under its section label in the prelude, a run
// of comments of its own between blank lines: "// Array", an empty "//", and the text.
func NamespaceDoc(name string) (string, string) {
	f, _ := Decls()
	for ns, doc := range f.NamespaceDocs {
		if strings.EqualFold(ns, name) {
			return ns, doc
		}
	}
	return "", ""
}
