package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ant-src/ant/docs"
	"github.com/ant-src/ant/internal/codes"
)

// TestExplainEveryCode checks that the page of ant check has an entry for every code, and only
// for codes, so that ant check -explain answers for each one.
func TestExplainEveryCode(t *testing.T) {
	for _, c := range codes.All {
		status, out, errOut := checkRun("-explain", string(c))
		if status != 0 || !strings.HasPrefix(out, string(c)+"\n\n") || errOut != "" {
			t.Errorf("-explain %s: status %d, output %q, errors %q", c, status, out, errOut)
		}
	}
	inCodes := false
	entry := regexp.MustCompile("^### `([^`]*)`$")
	for _, l := range strings.Split(docs.Check, "\n") {
		if strings.HasPrefix(l, "## ") {
			inCodes = l == "## Codes"
		}
		if m := entry.FindStringSubmatch(l); inCodes && m != nil && !codes.Known(codes.Code(m[1])) {
			t.Errorf("## Codes has an entry for %s, which is not a code", m[1])
		}
	}
}

// TestExplainPrintsExamples checks the layout of an entry: the code, its rule, and its examples
// indented by four spaces.
func TestExplainPrintsExamples(t *testing.T) {
	_, out, _ := checkRun("-explain", "assign")
	want := "assign\n\nAn assignment to a `let`, an import, or something else that cannot be assigned.\n\n" +
		"    let count = 0\n    count += 1 // error: cannot assign to count, a let; declare it with var to reassign it\n\n" +
		"Fixed:\n\n    var count = 0\n    count += 1\n"
	if out != want {
		t.Errorf("-explain assign printed\n%s\nwant\n%s", out, want)
	}
}

// TestExplainUsage checks that an unknown code, and -explain with a path, are usage errors.
func TestExplainUsage(t *testing.T) {
	if status, out, errOut := checkRun("-explain", "no-such-code"); status != 2 || out != "" || !strings.Contains(errOut, `unknown code "no-such-code"`) {
		t.Errorf("unknown code: status %d, output %q, errors %q", status, out, errOut)
	}
	if status, _, errOut := checkRun("-explain", "syntax", "main.ant"); status != 2 || !strings.Contains(errOut, "-explain takes a code") {
		t.Errorf("-explain with a path: status %d, errors %q", status, errOut)
	}
}
