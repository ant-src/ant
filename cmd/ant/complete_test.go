package main

import (
	"slices"
	"strings"
	"testing"
)

func TestComplete(t *testing.T) {
	t.Chdir(tree(t, map[string]string{
		"main.ant":         "",
		"notes.txt":        "",
		".hidden.ant":      "",
		"lib/util.ant":     "",
		"my dir/inner.ant": "",
		"tab\tname.ant":    "",
	}))
	cases := []struct {
		line string
		want []string
	}{
		{"ant ", []string{"run", "check", "fmt", "doc", "lsp", "dap", "help", "lib", "main.ant", "my dir"}},
		{"ant r", []string{"run"}},
		{"ant m", []string{"main.ant", "my dir"}},
		{"ant -", []string{"-h", "--help", "--comp"}},
		{"FOO=1 ant r", []string{"run"}},
		{"sudo ./bin/ant run l", []string{"lib"}},
		{"ant --comp ", []string{"bash", "zsh", "fish"}},
		{"ant --comp z", []string{"zsh"}},
		{"ant --comp zsh ", nil},
		{"ant run ", []string{"lib", "main.ant", "my dir"}},
		{"ant run l", []string{"lib"}},
		{"ant run lib/", []string{"lib/util.ant"}},
		{"ant run .", []string{".hidden.ant"}},
		{"ant run my\\ d", []string{"my dir"}},
		{"ant run 'my dir/", []string{"my dir/inner.ant"}},
		{"ant run main.ant ", []string{"lib", "main.ant", "my dir", "notes.txt"}},
		{"ant run main.ant -", nil},
		{"ant main.ant n", []string{"notes.txt"}},
		{"ant check ", []string{"lib", "main.ant", "my dir"}},
		{"ant check main.ant ", nil},
		{"ant fmt -", []string{"--check"}},
		{"ant fmt --check ", []string{"lib", "main.ant", "my dir"}},
		{"ant fmt main.ant ", []string{"lib", "main.ant", "my dir"}},
		{"ant doc ", []string{"lib", "main.ant", "my dir"}},
		{"ant help ", nil},
		{"ant frobnicate ", nil},
	}
	for _, c := range cases {
		var got []string
		for _, cand := range complete(c.line) {
			got = append(got, cand.value)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("%q: got %q, want %q", c.line, got, c.want)
		}
	}
}

func TestCompleteHome(t *testing.T) {
	t.Setenv("HOME", tree(t, map[string]string{"proj/main.ant": ""}))
	cases := []struct {
		line string
		want []string
	}{
		{"ant run ~", []string{"~/proj"}},
		{"ant run ~/p", []string{"~/proj"}},
		{"ant run ~/proj/", []string{"~/proj/main.ant"}},
		{"ant run ~root/", nil},
	}
	for _, c := range cases {
		var got []string
		for _, cand := range complete(c.line) {
			got = append(got, cand.value)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("%q: got %q, want %q", c.line, got, c.want)
		}
	}
}

func TestCompletionReply(t *testing.T) {
	t.Chdir(tree(t, map[string]string{"main.ant": "", "lib/util.ant": ""}))
	cases := []struct {
		env, val string
		args     []string
		want     string
	}{
		{"COMP_ZSH", "ant run ", nil, "lib/\tlib\tdir\nmain.ant\tmain.ant\tfile\n"},
		{"COMP_ZSH", "ant c", nil, "check\tcheck\tword\n"},
		{"COMP_FISH", "ant run ", nil, "lib/\nmain.ant\n"},
		{"COMP_LINE", "ant run m", []string{"ant", "m", "run"}, "main.ant\n"},
	}
	for _, c := range cases {
		t.Run(c.env, func(t *testing.T) {
			t.Setenv(c.env, c.val)
			var out strings.Builder
			if !completion(c.args, &out) {
				t.Fatal("did not complete")
			}
			if out.String() != c.want {
				t.Errorf("got %q, want %q", out.String(), c.want)
			}
		})
	}
}

func TestCompCommand(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		var out, errOut strings.Builder
		if status := run([]string{"--comp", shell}, strings.NewReader(""), &out, &errOut); status != 0 {
			t.Errorf("%s: status %d: %s", shell, status, errOut.String())
		}
		if !strings.Contains(out.String(), "ant") {
			t.Errorf("%s: script %q", shell, out.String())
		}
	}
	var out, errOut strings.Builder
	if status := run([]string{"--comp", "tcsh"}, strings.NewReader(""), &out, &errOut); status != 2 {
		t.Errorf("tcsh: status %d, want 2", status)
	}
}
