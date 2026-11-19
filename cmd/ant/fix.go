package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/internal/lsp"
)

// fixRounds bounds how many times fix checks again after applying fixes: fixes that keep
// producing fixes are a bug in the checker, not something to apply.
const fixRounds = 10

// fix is ant code fix (code.md, fix): it applies every auto fix of the diagnostics of the
// modules under paths, checks again, and repeats until none is left, then writes each file it
// changed once, or prints the change as a diff. A file with a syntax error is left as it is,
// since what recovery dropped may be what a fix would touch.
func (c *coder) fix(root string, paths []string, only map[string]bool, diff bool) int {
	if len(paths) == 0 {
		paths = []string{root}
	}
	files, err := checkSources.files(paths)
	if err != nil {
		return c.fail(&lsp.CodeError{Code: "usage", Message: err.Error()})
	}
	mine := map[string]bool{}
	for _, f := range files {
		mine[f] = true
	}
	old := map[string]string{} // each file as it was read
	now := map[string]string{} // each file as the fixes left it
	read := func(path string) (string, bool) {
		if src, ok := now[path]; ok {
			return src, true
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "", false
		}
		old[path], now[path] = string(b), string(b)
		return string(b), true
	}
	applied := map[string]int{}
	var diags []ant.Diagnostic
	for round := 0; ; round++ {
		var opts []ant.Option
		for path, src := range now {
			opts = append(opts, ant.WithSource(path, src))
		}
		diags, _ = ant.Check(files, opts...)
		edits, fixes := autoEdits(diags, mine, only)
		if len(edits) == 0 {
			break
		}
		if round == fixRounds {
			printDiagnostics(c.stderr, diags)
			return c.fail(&lsp.CodeError{Code: "unsettled", Message: fmt.Sprintf("fixes still apply after %d rounds; nothing was written", fixRounds)})
		}
		for path, es := range edits {
			src, ok := read(path)
			if !ok {
				return c.fail(&lsp.CodeError{Code: "load", Message: "cannot read " + rel(path)})
			}
			now[path] = applyFileEdits(src, es)
			applied[path] += fixes[path]
		}
	}
	var changed []lsp.CodeFile
	for path, src := range now {
		if src != old[path] {
			changed = append(changed, lsp.CodeFile{Path: path, Old: old[path], New: src})
		}
	}
	slices.SortFunc(changed, func(a, b lsp.CodeFile) int { return strings.Compare(a.Path, b.Path) })
	written := false
	if !diff && len(changed) > 0 {
		if status := c.write(changed); status != 0 {
			return status
		}
		written = true
	}
	var left []ant.Diagnostic
	for _, d := range diags {
		if len(only) == 0 || only[d.Code] {
			left = append(left, d)
		}
	}
	if c.json {
		type jsonFile struct {
			Path  string `json:"path"`
			Fixes int    `json:"fixes"`
		}
		out := struct {
			Files       []jsonFile       `json:"files"`
			Written     bool             `json:"written"`
			Diagnostics []jsonDiagnostic `json:"diagnostics"`
		}{Files: []jsonFile{}, Written: written, Diagnostics: jsonDiagnostics(left)}
		for _, f := range changed {
			out.Files = append(out.Files, jsonFile{rel(f.Path), applied[f.Path]})
		}
		c.emit(out)
	} else {
		for _, f := range changed {
			if diff {
				fmt.Fprint(c.stdout, fileDiff(rel(f.Path), f.Old, f.New))
			} else {
				word := "fixes"
				if applied[f.Path] == 1 {
					word = "fix"
				}
				fmt.Fprintf(c.stdout, "%s: %d %s\n", rel(f.Path), applied[f.Path], word)
			}
		}
		if len(left) > 0 {
			printDiagnostics(c.stderr, left)
		}
	}
	if len(left) > 0 {
		return 1
	}
	return 0
}

// fileEdit is one edit of a fix, by byte position in its file.
type fileEdit struct {
	start, end ant.Position
	text       string
}

// autoEdits gathers, by file, the edits of the fixes fix applies, and how many fixes they are:
// those of a diagnostic whose only fix is auto, in a file of mine with no syntax error, with a
// code in only when only is set. A fix whose edits overlap one taken before it waits for the
// next round.
func autoEdits(diags []ant.Diagnostic, mine, only map[string]bool) (map[string][]fileEdit, map[string]int) {
	broken := map[string]bool{}
	for _, d := range diags {
		if d.Code == "syntax" {
			broken[d.Pos.Path] = true
		}
	}
	out := map[string][]fileEdit{}
	count := map[string]int{}
	for _, d := range diags {
		if len(d.Fixes) != 1 || !d.Fixes[0].Auto || len(only) > 0 && !only[d.Code] {
			continue
		}
		var es []fileEdit
		ok := true
		for _, e := range d.Fixes[0].Edits {
			path := e.Start.Path
			if !mine[path] || broken[path] {
				ok = false
				break
			}
			es = append(es, fileEdit{start: ant.Position{Path: path, Line: e.Start.Line, Col: e.Start.Col},
				end: ant.Position{Path: path, Line: e.End.Line, Col: e.End.Col}, text: e.Text})
		}
		if !ok || slices.ContainsFunc(es, func(e fileEdit) bool { return overlaps(out[e.start.Path], e) }) {
			continue
		}
		for _, e := range es {
			out[e.start.Path] = append(out[e.start.Path], e)
		}
		count[es[0].start.Path]++
	}
	return out, count
}

// overlaps reports whether e touches the text of any of es.
func overlaps(es []fileEdit, e fileEdit) bool {
	for _, x := range es {
		if before(e.start, x.end) && before(x.start, e.end) || e.start == x.start {
			return true
		}
	}
	return false
}

func before(a, b ant.Position) bool {
	return a.Line < b.Line || a.Line == b.Line && a.Col < b.Col
}

// applyFileEdits applies non-overlapping edits to src, the last first, so that each position
// still counts in src as it was.
func applyFileEdits(src string, es []fileEdit) string {
	slices.SortFunc(es, func(a, b fileEdit) int {
		if before(a.start, b.start) {
			return 1
		}
		if before(b.start, a.start) {
			return -1
		}
		return 0
	})
	lines := strings.SplitAfter(src, "\n")
	offset := func(p ant.Position) int {
		n := 0
		for i := 0; i < p.Line-1 && i < len(lines); i++ {
			n += len(lines[i])
		}
		return min(n+p.Col-1, len(src))
	}
	for _, e := range es {
		src = src[:offset(e.start)] + e.text + src[offset(e.end):]
	}
	return src
}

// fileDiff is unifiedDiff, which pairs the lines of old and new one to one, or, when a fix
// changed how many lines there are, one hunk that replaces the whole file.
func fileDiff(path, old, new string) string {
	a, b := splitLines(old), splitLines(new)
	if len(a) == len(b) {
		return unifiedDiff(path, old, new)
	}
	var out strings.Builder
	fmt.Fprintf(&out, "--- a/%s\n+++ b/%s\n@@ -1,%d +1,%d @@\n", filepath.ToSlash(path), filepath.ToSlash(path), len(a), len(b))
	for _, l := range a {
		out.WriteString("-" + l + "\n")
	}
	for _, l := range b {
		out.WriteString("+" + l + "\n")
	}
	return out.String()
}
