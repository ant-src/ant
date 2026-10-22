package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedCorpus adds the test programs as seeds.
func seedCorpus(f *testing.F) {
	for _, pattern := range []string{"../tests/run/*.ant", "../tests/check/*.ant", "../tests/unchecked/*.ant"} {
		paths, _ := filepath.Glob(pattern)
		for _, path := range paths {
			src, err := os.ReadFile(path)
			if err == nil {
				f.Add(string(src))
			}
		}
	}
	f.Add("var x = ")
	f.Add("func f() {")
	f.Add(strings.Repeat("(", 50) + "1" + strings.Repeat(")", 50))
	f.Add("x |> f(_) |> g(1, _)")
	f.Add("func f() { select { var v = Chan.receive(c) { v }\n Chan.send(d, 1) { }\n else { 0 } } }")
	f.Add("func f() { while true { select { Context.wait(ctx) { return }\n Time.sleep(10) { break } } } }")
}

// FuzzParse checks that the parser never panics in any mode and that every error is a *Error.
func FuzzParse(f *testing.F) {
	seedCorpus(f)
	f.Fuzz(func(t *testing.T, src string) {
		for _, mode := range []Mode{Program, Session, Spec, Declarations} {
			_, err := Parse("fuzz.ant", src, mode)
			if err != nil {
				if _, ok := err.(*Error); !ok {
					t.Fatalf("mode %v: error of type %T: %v", mode, err, err)
				}
			}
		}
		if _, err := ParseExpr(src); err != nil {
			if _, ok := err.(*Error); !ok {
				t.Fatalf("ParseExpr: error of type %T: %v", err, err)
			}
		}
	})
}
