package format

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ant-src/ant/parser"
)

// FuzzFormat checks the guarantees of fmt.md section 2 on any input that parses as a module.
func FuzzFormat(f *testing.F) {
	for _, pattern := range []string{"testdata/*.ant", "../../examples/*.ant", "../../tests/run/*.ant", "../../tests/check/*.ant"} {
		paths, _ := filepath.Glob(pattern)
		for _, path := range paths {
			if src, err := os.ReadFile(path); err == nil {
				f.Add(string(src))
			}
		}
	}
	f.Add("func f() { a; b // c\n}")
	f.Add("func f() { g(a // c\n, b) }")
	f.Add("func f() { var [a, ...b] = [\n1]\n}")
	f.Add("func f() { x\n-1\n(y\n-1) }")
	f.Add("func f() { `${ {x}(1) } ${\n a }` }")
	f.Add("{};/0/")
	f.Fuzz(func(t *testing.T, src string) {
		if _, err := parser.Parse("fuzz.ant", src, parser.Program); err != nil {
			return
		}
		checkProperties(t, "fuzz.ant", src)
	})
}
