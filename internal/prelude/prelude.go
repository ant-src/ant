package prelude

import (
	_ "embed"
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
