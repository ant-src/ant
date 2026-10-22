package ant

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ant-src/ant/ast"
)

// TestInlineCachesHoldNoRunState runs a program that reads fields and builds records, then checks
// that every inline cache the run filled is keyed by a tag's field list. A key holding the tag
// would keep the tag's functions, and through them the finished run's module state, alive for
// as long as the Program is.
func TestInlineCachesHoldNoRunState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.ant")
	src := `type Point = { x: int, y: int }
func Point.sum(p: Point): int { p.x + p.y }
var big: Array[int] = Array.range(0, 1000)
func main(): void {
  var p = Point{ x: 1, y: 2 }
  p.x = 3
  print(Point.sum(p) + Array.length(big))
}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	prog, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := prog.Run(context.Background(), WithStdout(io.Discard)); err != nil {
		t.Fatal(err)
	}
	filled := 0
	visit(reflect.ValueOf(prog.prog.Root.File), map[uintptr]bool{}, func(c *ast.Cache) {
		e := c.Load()
		if e == nil {
			return
		}
		filled++
		if _, ok := e.Key.(*string); !ok {
			t.Errorf("a cache is keyed by %T, want *string", e.Key)
		}
	})
	if filled < 3 {
		t.Errorf("%d caches filled, want at least 3", filled)
	}
}

// visit calls f on every *ast.Cache reachable from v.
func visit(v reflect.Value, seen map[uintptr]bool, f func(*ast.Cache)) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() || seen[v.Pointer()] {
			return
		}
		seen[v.Pointer()] = true
		if c, ok := v.Interface().(*ast.Cache); ok {
			f(c)
			return
		}
		visit(v.Elem(), seen, f)
	case reflect.Interface:
		if !v.IsNil() {
			visit(v.Elem(), seen, f)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				visit(v.Field(i), seen, f)
			}
		}
	case reflect.Slice:
		for i := range v.Len() {
			visit(v.Index(i), seen, f)
		}
	}
}
