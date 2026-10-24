package prelude

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/parser"
)

//go:embed prelude.ant
var source string

var (
	once sync.Once
	file *ast.File
	err  error
)

// Decls returns the prelude's declarations, parsed once.
func Decls() (*ast.File, error) {
	once.Do(func() { file, err = parser.Parse("prelude.ant", source, parser.Declarations) })
	return file, err
}

var namespaceDocs = sync.OnceValue(func() map[string][2]string {
	docs := map[string][2]string{}
	blocks := strings.Split(source, "\n\n")
	for _, b := range blocks {
		lines := strings.Split(strings.TrimSpace(b), "\n")
		name, ok := strings.CutPrefix(lines[0], "// ")
		if !ok || !IsName(name) || len(lines) < 3 || lines[1] != "//" {
			continue
		}
		var text []string
		for _, l := range lines[2:] {
			l, ok := strings.CutPrefix(l, "//")
			if !ok {
				break // a declaration: this block is not a label
			}
			text = append(text, strings.TrimPrefix(l, " "))
		}
		if len(text) == len(lines)-2 {
			docs[strings.ToLower(name)] = [2]string{name, strings.Join(text, "\n")}
		}
	}
	return docs
})

// NamespaceDoc returns the namespace name, matched without regard to case, as the prelude spells
// it, with its doc, or "", "" when it has none. A namespace's doc is the paragraph under its section label in the prelude, a run
// of comments of its own between blank lines: "// Array", an empty "//", and the text.
func NamespaceDoc(name string) (string, string) {
	d := namespaceDocs()[strings.ToLower(name)]
	return d[0], d[1]
}
