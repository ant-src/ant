package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// corpusDir holds the programs of ant test: test files with the companions and helpers they
// use, each test file ending in a trailer, as the programs under tests/run do (tests/README.md):
//
//	// args: -v -run x     the flags given to ant test before the file
//	// out: <line>         one expected line of standard output, in order
//	// exit: <n>           the expected exit status; default 0
//	// err: <text>         standard error must contain this text; may repeat
//
// Times are compared as (0.00s), and JSON elapsed times as 0. A file whose trailer has no
// "// expect" line is a helper another test file imports.
const corpusDir = "../../tests/test"

var (
	elapsedText = regexp.MustCompile(`\(\d+\.\d\ds\)`)
	elapsedJSON = regexp.MustCompile(`"elapsed":[0-9.e-]+`)
)

func TestTestCorpus(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(corpusDir, "*_test.ant"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no test files under %s: %v", corpusDir, err)
	}
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		i := strings.LastIndex(string(src), "\n// expect\n")
		if i < 0 {
			continue
		}
		t.Run(strings.TrimSuffix(filepath.Base(path), ".ant"), func(t *testing.T) {
			var args, out, errs []string
			exit := 0
			for _, line := range strings.Split(strings.TrimRight(string(src[i+len("\n// expect\n"):]), "\n"), "\n") {
				key, rest, _ := strings.Cut(line, ":")
				rest = strings.TrimPrefix(rest, " ")
				switch key {
				case "// args":
					args = strings.Fields(rest)
				case "// out":
					out = append(out, rest)
				case "// exit":
					exit, err = strconv.Atoi(rest)
					if err != nil {
						t.Fatalf("bad exit line %q", line)
					}
				case "// err":
					errs = append(errs, rest)
				default:
					t.Fatalf("unexpected trailer line %q", line)
				}
			}
			var stdout, stderr bytes.Buffer
			status := run(append(append([]string{"test"}, args...), path), strings.NewReader(""), &stdout, &stderr)
			prefix := corpusDir + "/"
			got := strings.ReplaceAll(stdout.String(), prefix, "")
			got = elapsedText.ReplaceAllString(got, "(0.00s)")
			got = elapsedJSON.ReplaceAllString(got, `"elapsed":0`)
			want := strings.Join(out, "\n")
			if len(out) > 0 {
				want += "\n"
			}
			if got != want {
				t.Errorf("stdout:\n%s\nwant:\n%s", got, want)
			}
			if status != exit {
				t.Errorf("exit status %d, want %d\nstderr:\n%s", status, exit, stderr.String())
			}
			for _, text := range errs {
				if !strings.Contains(stderr.String(), text) {
					t.Errorf("stderr lacks %q:\n%s", text, stderr.String())
				}
			}
			// Every test file type-checks, with its companion.
			var cout, cerr bytes.Buffer
			if status := run([]string{"check", path}, strings.NewReader(""), &cout, &cerr); status != 0 {
				t.Errorf("ant check: status %d\n%s", status, cerr.String())
			}
		})
	}
}
