package ant_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ant-src/ant"
)

// A tag the library made is shared by every program in the process, so what reflection resolves
// from it must not depend on which program reached it first: each program resolves the prelude's
// declarations for itself.
func TestReflectionAcrossPrograms(t *testing.T) {
	src := `
var s = Reflect.shape(Reflect.erase(type(Http.Reply)))
if s is Reflect.Record {
  for f in s.fields {
    if f.name == "body" { println(Reflect.assignable(Reflect.erase(type(Bytes)), f.type)) }
  }
}
`
	for i := range 2 {
		var out strings.Builder
		if err := loadSrc(t, src).Run(context.Background(), ant.WithStdout(&out)); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
		if out.String() != "true\n" {
			t.Errorf("run %d: stdout %q", i, out.String())
		}
	}
}
