package ant_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)

func seedPrograms(f *testing.F) {
	for _, pattern := range []string{"tests/run/*.ant", "tests/check/*.ant", "tests/unchecked/*.ant"} {
		paths, _ := filepath.Glob(pattern)
		for _, path := range paths {
			src, err := os.ReadFile(path)
			if err == nil {
				f.Add(string(src))
			}
		}
	}
}

// load loads src as a program in an empty file system, so an import can reach nothing.
func load(src string) (*ant.Program, error) {
	return ant.Load("main.ant", ant.WithSource("main.ant", src), ant.WithFS(fstest.MapFS{}))
}

// FuzzCheck checks that the checker never panics on any program that loads.
func FuzzCheck(f *testing.F) {
	seedPrograms(f)
	f.Fuzz(func(t *testing.T, src string) {
		prog, err := load(src)
		if err != nil {
			if _, ok := err.(*ant.LoadError); !ok {
				t.Fatalf("error of type %T: %v", err, err)
			}
			return
		}
		prog.Check()
	})
}

// unsafeForFuzz matches programs FuzzRun skips: goroutines and blocking host calls can outlive
// the deadline, and large counts or literals can exhaust memory before it.
var unsafeForFuzz = regexp.MustCompile(`\bgo\b|Mutex|RwMutex|WaitGroup|Once|Cond|Atomic|Time\.|File\.|Http\.|Os\.|repeat|filled|\d{5,}`)

// FuzzRun checks that running any program that loads, checked or not, ends in a normal return,
// an Err from the script, an Ant panic, an exit, or the deadline, never a Go panic.
func FuzzRun(f *testing.F) {
	seedPrograms(f)
	f.Fuzz(func(t *testing.T, src string) {
		if unsafeForFuzz.MatchString(src) {
			return
		}
		prog, err := load(src)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		err = prog.Run(ctx, ant.WithStdout(io.Discard), ant.WithStderr(io.Discard))
		switch err {
		case nil, context.DeadlineExceeded:
			return
		}
		switch err.(type) {
		case *ant.ResultError, *ant.PanicError, *ant.ExitError:
		default:
			t.Fatalf("error of type %T: %v", err, err)
		}
	})
}

// FuzzStrings runs the String functions over arbitrary input through a session: they must
// respect code point boundaries and never panic in Go.
func FuzzStrings(f *testing.F) {
	for _, s := range []string{"", "héllo", "a\xffb", "日本語", "\x00", "ab"} {
		f.Add(s, 1, 2)
	}
	var input string
	ns := ant.Namespace{
		Name:       "Fuzz",
		Signatures: "func Fuzz.input(): string\n",
		Funcs: map[string]ant.HostFunc{
			"input": func(rt ant.Runtime, args []value.Value) (value.Value, error) { return input, nil },
		},
	}
	session, err := ant.NewSession(context.Background(), ant.WithNamespace(ns), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard))
	if err != nil {
		f.Fatal(err)
	}
	defer session.Close()
	exprs := []string{
		"String.length(Fuzz.input())",
		"String.slice(Fuzz.input(), I, J)",
		"String.split(Fuzz.input(), String.slice(Fuzz.input(), 0, 1))",
		"String.toUpper(Fuzz.input()) + String.toLower(Fuzz.input())",
		"String.indexOf(Fuzz.input(), String.slice(Fuzz.input(), I, J))",
		"String.replace(Fuzz.input(), String.slice(Fuzz.input(), 0, I), \"x\")",
		"String.trim(Fuzz.input())",
		"String.codePoints(Fuzz.input())",
		"String.fromBytes(String.toBytes(Fuzz.input())) == Fuzz.input()",
		"String.from(Fuzz.input())",
		"Json.stringify(Fuzz.input())",
		"Json.parse(Fuzz.input())",
	}
	f.Fuzz(func(t *testing.T, s string, i, j int) {
		input = s
		for _, e := range exprs {
			e = strings.NewReplacer("I", strconv.Itoa(i%64), "J", strconv.Itoa(j%64)).Replace(e)
			session.Eval(e) // an Ant panic comes back as an error, which is fine
		}
	})
}
