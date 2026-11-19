package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/internal/lsp"
)

// codeSubcommands are the subcommands of ant code, with the flags each takes beyond -json and
// -root (private/tools/code.md section 1).
var codeSubcommands = map[string][]string{
	"def":     nil,
	"refs":    {"-decl"},
	"type":    nil,
	"symbols": {"-match", "-all"},
	"rename":  {"-diff", "-force"},
}

var codeOperands = map[string]string{
	"def":     "<loc>",
	"refs":    "<loc>",
	"type":    "<loc>",
	"symbols": "[path...]",
	"rename":  "<loc> <name>",
}

// codeStatus is the exit status of each error code (code.md section 3).
var codeStatus = map[string]int{
	"usage": 2, "load": 2, "not-found": 1, "ambiguous": 1, "refused": 1, "check": 1,
	"incomplete": 2, "changed": 1, "write": 2,
}

// codeCommand is ant code (private/tools/code.md): the questions an editor asks the language
// server, answered from a terminal by the server's own analysis.
func codeCommand(args []string, stdout, stderr io.Writer) int {
	c := &coder{stdout: stdout, stderr: stderr}
	for _, a := range args {
		if a == "-json" || a == "--json" {
			c.json = true // so that a usage error is JSON too
		}
	}
	if len(args) == 0 || codeOperands[args[0]] == "" {
		if !c.json {
			fmt.Fprintln(stderr, commandUsage("code"))
			return 2
		}
		return c.fail(&lsp.CodeError{Code: "usage", Message: commandUsage("code")})
	}
	sub := args[0]
	c.sub = sub
	fl := flag.NewFlagSet("ant code "+sub, flag.ContinueOnError)
	fl.SetOutput(io.Discard)
	fl.BoolVar(&c.json, "json", false, "print JSON instead of text")
	root := fl.String("root", ".", "the workspace")
	decl := fl.Bool("decl", false, "include the declaration itself")
	match := fl.String("match", "", "only declarations whose name contains text")
	all := fl.Bool("all", false, "include fields and the declarations in functions")
	diff := fl.Bool("diff", false, "print a unified diff and write nothing")
	force := fl.Bool("force", false, "rename even when the search is incomplete or a program has errors")
	err := fl.Parse(args[1:])
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(stdout, c.usage())
		return 0
	}
	if err == nil {
		// a flag of another subcommand
		fl.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "json", "root":
			default:
				if !contains(codeSubcommands[sub], "-"+f.Name) {
					err = fmt.Errorf("flag provided but not defined: -%s", f.Name)
				}
			}
		})
	}
	operands := fl.Args()
	if err == nil {
		switch sub {
		case "symbols":
		case "rename":
			if len(operands) != 2 {
				err = errors.New("rename takes a location and a new name")
			}
		default:
			if len(operands) != 1 {
				err = fmt.Errorf("%s takes one location", sub)
			}
		}
	}
	if err != nil {
		return c.fail(&lsp.CodeError{Code: "usage", Message: err.Error() + "\n" + c.usage()})
	}
	an, err := lsp.NewCode(*root)
	if err != nil {
		return c.fail(&lsp.CodeError{Code: "usage", Message: "-root: " + err.Error()})
	}
	switch sub {
	case "def":
		return c.def(an, operands[0])
	case "refs":
		return c.refs(an, operands[0], *decl)
	case "type":
		return c.typ(an, operands[0])
	case "symbols":
		return c.symbols(an, operands, *match, *all)
	default:
		return c.rename(an, operands[0], operands[1], *diff, *force)
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// coder prints the answers of one ant code request.
type coder struct {
	sub            string
	json           bool
	stdout, stderr io.Writer
}

func (c *coder) usage() string {
	flags := "[-json] [-root dir]"
	for _, f := range codeSubcommands[c.sub] {
		switch f {
		case "-match":
			flags += " [-match text]"
		default:
			flags += " [" + f + "]"
		}
	}
	return "usage: ant code " + c.sub + " " + flags + " " + codeOperands[c.sub]
}

// rel is how a path is printed: relative to the current directory when the file lies under it.
func rel(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	wd, err := os.Getwd()
	if err != nil {
		return path
	}
	r, err := filepath.Rel(wd, path)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return path
	}
	return r
}

func locText(l lsp.CodeLoc) string {
	return fmt.Sprintf("%s:%d:%d", rel(l.Path), l.Start.Line, l.Start.Col)
}

type jsonLoc struct {
	Path    string  `json:"path"`
	Line    int     `json:"line"`
	Column  int     `json:"column"`
	Start   jsonPos `json:"start"`
	End     jsonPos `json:"end"`
	Prelude bool    `json:"prelude,omitempty"`
}

func toJSONLoc(l lsp.CodeLoc) jsonLoc {
	return jsonLoc{Path: rel(l.Path), Line: l.Start.Line, Column: l.Start.Col,
		Start: jsonPos{l.Start.Line, l.Start.Col}, End: jsonPos{l.End.Line, l.End.Col}, Prelude: l.Prelude}
}

func relAll(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = rel(p)
	}
	return out
}

// emit writes v as one JSON object and a newline.
func (c *coder) emit(v any) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.Encode(v)
	c.stdout.Write(b.Bytes())
}

// codeFailure is a failed request as JSON prints it.
type codeFailure struct {
	Code       string           `json:"code"`
	Message    string           `json:"message"`
	Candidates []jsonLoc        `json:"candidates,omitempty"`
	Errors     []jsonDiagnostic `json:"errors,omitempty"`
	Skipped    []string         `json:"skipped,omitempty"`
	Paths      []string         `json:"paths,omitempty"`
	Written    *[]string        `json:"written,omitempty"`
	Unwritten  []string         `json:"unwritten,omitempty"`
}

// fail reports a request that could not be answered and returns its status.
func (c *coder) fail(e *lsp.CodeError) int {
	return c.failWith(codeFailure{Code: e.Code, Message: e.Message}, e)
}

func (c *coder) failWith(f codeFailure, e *lsp.CodeError) int {
	if e != nil {
		for _, l := range e.Candidates {
			f.Candidates = append(f.Candidates, toJSONLoc(l))
		}
		if len(e.Errors) > 0 {
			f.Errors = jsonDiagnostics(antDiags(e.Errors))
		}
		f.Skipped = relAll(e.Skipped)
	}
	if len(f.Skipped) == 0 {
		f.Skipped = nil
	}
	if c.json {
		c.emit(map[string]any{"error": f})
		return codeStatus[f.Code]
	}
	fmt.Fprintf(c.stderr, "ant code: %s\n", f.Message)
	if e != nil {
		for _, l := range e.Candidates {
			fmt.Fprintf(c.stderr, "%s: %s\n", locText(l), strings.TrimSpace(l.Line))
		}
		if len(e.Errors) > 0 {
			printDiagnostics(c.stderr, antDiags(e.Errors))
		}
	}
	for _, p := range f.Skipped {
		fmt.Fprintf(c.stderr, "ant code: cannot read or parse %s\n", p)
	}
	for _, p := range f.Paths {
		fmt.Fprintf(c.stderr, "ant code: %s changed since it was analyzed\n", p)
	}
	if f.Written != nil {
		for _, p := range *f.Written {
			fmt.Fprintf(c.stderr, "ant code: wrote %s\n", p)
		}
		for _, p := range f.Unwritten {
			fmt.Fprintf(c.stderr, "ant code: did not write %s\n", p)
		}
	}
	return codeStatus[f.Code]
}

// antDiags turns the errors of a program into ant check's, with paths as printed.
func antDiags(ds []lsp.CodeDiag) []ant.Diagnostic {
	out := make([]ant.Diagnostic, len(ds))
	for i, d := range ds {
		out[i] = ant.Diagnostic{Pos: ant.Position{Path: rel(d.Path), Line: d.Pos.Line, Col: d.Pos.Col},
			End: ant.Position{Path: rel(d.Path), Line: d.End.Line, Col: d.End.Col}, Message: d.Message}
	}
	return out
}

func (c *coder) def(an *lsp.Code, loc string) int {
	l, sig, err := an.Def(loc)
	if err != nil {
		return c.fail(err)
	}
	if c.json {
		c.emit(struct {
			Location  jsonLoc `json:"location"`
			Signature string  `json:"signature"`
		}{toJSONLoc(l), sig})
		return 0
	}
	fmt.Fprintf(c.stdout, "%s: %s\n", locText(l), sig)
	return 0
}

// incomplete names the files a search skipped, on standard error, and returns the status.
func (c *coder) incomplete(skipped []string) int {
	for _, p := range skipped {
		fmt.Fprintf(c.stderr, "ant code: cannot read or parse %s, which may use the name\n", rel(p))
	}
	if len(skipped) > 0 {
		return 2
	}
	return 0
}

func (c *coder) refs(an *lsp.Code, loc string, decl bool) int {
	res, err := an.Refs(loc, decl)
	if err != nil {
		return c.fail(err)
	}
	if c.json {
		type jsonRef struct {
			Location jsonLoc `json:"location"`
			Alias    *string `json:"alias"`
		}
		out := struct {
			Declaration *jsonLoc  `json:"declaration"`
			References  []jsonRef `json:"references"`
			Complete    bool      `json:"complete"`
			Skipped     []string  `json:"skipped"`
		}{References: []jsonRef{}, Complete: len(res.Skipped) == 0, Skipped: relAll(res.Skipped)}
		if res.Decl != nil {
			l := toJSONLoc(*res.Decl)
			out.Declaration = &l
		}
		for _, r := range res.Refs {
			jr := jsonRef{Location: toJSONLoc(r.Loc)}
			if r.Alias != "" {
				alias := r.Alias
				jr.Alias = &alias
			}
			out.References = append(out.References, jr)
		}
		c.emit(out)
		return c.incomplete(res.Skipped)
	}
	files := map[string]bool{}
	for _, r := range res.Refs {
		line := strings.TrimSpace(r.Loc.Line)
		if r.Alias != "" {
			line += " (as " + r.Alias + ")"
		}
		fmt.Fprintf(c.stdout, "%s: %s\n", locText(r.Loc), line)
		files[r.Loc.Path] = true
	}
	fmt.Fprintf(c.stderr, "%s in %s\n", plural(len(res.Refs), "reference"), plural(len(files), "file"))
	return c.incomplete(res.Skipped)
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}

func (c *coder) typ(an *lsp.Code, loc string) int {
	res, err := an.Type(loc)
	if err != nil {
		return c.fail(err)
	}
	if c.json {
		c.emit(struct {
			Kind     string  `json:"kind"`
			Text     string  `json:"text"`
			Doc      string  `json:"doc"`
			Location jsonLoc `json:"location"`
		}{res.Kind, res.Text, res.Doc, toJSONLoc(res.Loc)})
		return 0
	}
	fmt.Fprintln(c.stdout, res.Text)
	if res.Doc != "" {
		for _, l := range strings.Split(strings.TrimRight(res.Doc, "\n"), "\n") {
			if l == "" {
				fmt.Fprintln(c.stdout)
			} else {
				fmt.Fprintln(c.stdout, "    "+l)
			}
		}
	}
	return 0
}

func (c *coder) symbols(an *lsp.Code, paths []string, match string, all bool) int {
	syms, broken, err := an.Symbols(paths, match, all)
	if err != nil {
		return c.fail(err)
	}
	if c.json {
		type jsonSymbol struct {
			Location  jsonLoc `json:"location"`
			Kind      string  `json:"kind"`
			Name      string  `json:"name"`
			Signature string  `json:"signature"`
			Exported  bool    `json:"exported"`
			Container *string `json:"container"`
		}
		out := struct {
			Symbols []jsonSymbol `json:"symbols"`
		}{[]jsonSymbol{}}
		for _, s := range syms {
			js := jsonSymbol{Location: toJSONLoc(s.Loc), Kind: s.Kind, Name: s.Name, Signature: s.Signature, Exported: s.Exported}
			if s.Container != "" {
				container := s.Container
				js.Container = &container
			}
			out.Symbols = append(out.Symbols, js)
		}
		c.emit(out)
	} else {
		for _, s := range syms {
			fmt.Fprintf(c.stdout, "%s: %s\n", locText(s.Loc), s.Signature)
		}
	}
	for _, p := range broken {
		fmt.Fprintf(c.stderr, "ant code: %s has syntax errors; listed what could be read\n", rel(p))
	}
	switch {
	case len(broken) > 0:
		return 2
	case len(syms) == 0:
		return 1
	}
	return 0
}

func (c *coder) rename(an *lsp.Code, loc, name string, diff, force bool) int {
	res, err := an.Rename(loc, name, force)
	if err != nil {
		return c.fail(err)
	}
	if len(res.Errors) > 0 {
		printDiagnostics(c.stderr, antDiags(res.Errors))
		fmt.Fprintln(c.stderr, "ant code: warning: renaming in programs with errors, where a use may be missed")
	}
	if len(res.Skipped) > 0 {
		for _, p := range res.Skipped {
			fmt.Fprintf(c.stderr, "ant code: warning: cannot read or parse %s, which may use the name\n", rel(p))
		}
	}
	written := false
	if !diff && len(res.Files) > 0 {
		if status := c.write(res.Files); status != 0 {
			return status
		}
		written = true
	}
	if c.json {
		type jsonEdit struct {
			Location jsonLoc `json:"location"`
			NewText  string  `json:"newText"`
		}
		type jsonFile struct {
			Path  string     `json:"path"`
			Edits []jsonEdit `json:"edits"`
		}
		out := struct {
			Files    []jsonFile `json:"files"`
			Written  bool       `json:"written"`
			Complete bool       `json:"complete"`
			Skipped  []string   `json:"skipped"`
		}{Files: []jsonFile{}, Written: written, Complete: len(res.Skipped) == 0, Skipped: relAll(res.Skipped)}
		for _, f := range res.Files {
			jf := jsonFile{Path: rel(f.Path)}
			for _, e := range f.Edits {
				jf.Edits = append(jf.Edits, jsonEdit{toJSONLoc(e.Loc), e.NewText})
			}
			out.Files = append(out.Files, jf)
		}
		c.emit(out)
		return 0
	}
	for _, f := range res.Files {
		if diff {
			fmt.Fprint(c.stdout, unifiedDiff(rel(f.Path), f.Old, f.New))
		} else {
			fmt.Fprintf(c.stdout, "%s: %s\n", rel(f.Path), plural(len(f.Edits), "edit"))
		}
	}
	return 0
}

// write replaces each file with its new text, after checking that none changed since it was
// read (code.md section 9).
func (c *coder) write(files []lsp.CodeFile) int {
	var changed []string
	for _, f := range files {
		now, err := os.ReadFile(f.Path)
		if err != nil || string(now) != f.Old {
			changed = append(changed, rel(f.Path))
		}
	}
	if len(changed) > 0 {
		return c.failWith(codeFailure{Code: "changed", Message: "files changed since they were analyzed; nothing was written", Paths: changed}, nil)
	}
	written := []string{}
	for i, f := range files {
		if err := replace(f.Path, []byte(f.New)); err != nil {
			unwritten := relAll(filesPaths(files[i:]))
			return c.failWith(codeFailure{Code: "write", Message: err.Error(), Written: &written, Unwritten: unwritten}, nil)
		}
		written = append(written, rel(f.Path))
	}
	return 0
}

func filesPaths(files []lsp.CodeFile) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Path
	}
	return out
}

// unifiedDiff renders the change from old to new as a unified diff with three lines of context.
// A rename changes names within lines, so each line of old has its line in new.
func unifiedDiff(path, old, new string) string {
	a, b := splitLines(old), splitLines(new)
	var changed []int
	for i := range a {
		if a[i] != b[i] {
			changed = append(changed, i)
		}
	}
	var out strings.Builder
	fmt.Fprintf(&out, "--- a/%s\n+++ b/%s\n", filepath.ToSlash(path), filepath.ToSlash(path))
	line := func(prefix, l string) {
		out.WriteString(prefix)
		out.WriteString(l)
		if !strings.HasSuffix(l, "\n") {
			out.WriteString("\n\\ No newline at end of file\n")
		}
	}
	for k := 0; k < len(changed); {
		// a hunk holds the changes no more than six lines apart, with three lines around them
		j := k + 1
		for j < len(changed) && changed[j]-changed[j-1]-1 <= 6 {
			j++
		}
		start, end := max(changed[k]-3, 0), min(changed[j-1]+4, len(a))
		fmt.Fprintf(&out, "@@ -%d,%d +%d,%d @@\n", start+1, end-start, start+1, end-start)
		for i := start; i < end; {
			if a[i] == b[i] {
				line(" ", a[i])
				i++
				continue
			}
			run := i
			for run < end && a[run] != b[run] {
				run++
			}
			for x := i; x < run; x++ {
				line("-", a[x])
			}
			for x := i; x < run; x++ {
				line("+", b[x])
			}
			i = run
		}
		k = j
	}
	return out.String()
}

// splitLines splits text after each \n, keeping it.
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	lines := strings.SplitAfter(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
