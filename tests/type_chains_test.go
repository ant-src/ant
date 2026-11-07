package tests

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ant-src/ant"
)

// TestLongTypeChains checks and runs programs with long chains of type declarations, each naming
// the one before (section 4, Type declarations). The checks that a declaration does not reach
// itself through unions alone, or with other type arguments, once followed every declaration a
// chain reaches from each of its links, which took time quadratic or cubic in its length.
func TestLongTypeChains(t *testing.T) {
	chain := func(b *strings.Builder, name string, n int, generic bool) {
		p, arg := "", ""
		if generic {
			p, arg = "[T]", "[T]"
		}
		fmt.Fprintf(b, "type %s0%s = %s | null\n", name, p, map[bool]string{false: "int", true: "T"}[generic])
		for i := 1; i <= n; i++ {
			fmt.Fprintf(b, "type %s%d%s = Array[%s%d%s]\n", name, i, p, name, i-1, arg)
		}
	}
	cases := []struct {
		name  string
		src   func() string
		check bool
	}{
		{"two chains of aliases", func() string {
			var b strings.Builder
			chain(&b, "D", 8000, false)
			chain(&b, "E", 8000, false)
			b.WriteString("println(Reflect.erase(type(D8000)) == Reflect.erase(type(E8000)))\n")
			return b.String()
		}, true},
		{"a chain of generic aliases, checked", func() string {
			var b strings.Builder
			chain(&b, "D", 1000, true)
			b.WriteString("println(type(D1000[int]) == type(D1000[int]))\n")
			return b.String()
		}, true},
		{"a chain of generic aliases, run", func() string {
			var b strings.Builder
			chain(&b, "D", 8000, true)
			b.WriteString("println(type(D8000[int]) == type(D8000[int]))\n")
			return b.String()
		}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := c.src()
			start := time.Now()
			if c.check {
				if diags := checkProgram(t, src); diags != "" {
					t.Fatalf("the checker rejects the program:\n%s", diags)
				}
			}
			prog, err := ant.Load("chain.ant", ant.WithSource("chain.ant", src), ant.WithFS(fstest.MapFS{}))
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := prog.Run(context.Background(), ant.WithStdout(&out)); err != nil {
				t.Fatal(err)
			}
			if out.String() != "true\n" {
				t.Errorf("the program printed %q, want true", out.String())
			}
			if d := time.Since(start); d > 5*time.Second {
				t.Errorf("took %v", d)
			}
		})
	}
}
