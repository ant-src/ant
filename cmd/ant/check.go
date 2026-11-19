package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/diag"
)

// checkCommand checks the files the paths name, and every .ant file under the directories,
// together: each module once, however many of them import it.
func checkCommand(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("ant check", flag.ContinueOnError)
	fl.SetOutput(stderr)
	jsonOut := fl.Bool("json", false, "print each error as a JSON line on standard output")
	explainCode := fl.String("explain", "", "print what a code means, with an example, and exit")
	fl.Usage = func() {
		fmt.Fprintln(stderr, "usage: ant check [-json] [path ...]\n       ant check -explain code")
		fl.PrintDefaults()
	}
	if err := fl.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *explainCode != "" {
		if fl.NArg() > 0 || *jsonOut {
			fmt.Fprintln(stderr, "ant check: -explain takes a code and nothing else")
			return 2
		}
		if !explain(stdout, *explainCode) {
			fmt.Fprintf(stderr, "ant check: unknown code %q; the codes are listed under Codes in the documentation of ant check\n", *explainCode)
			return 2
		}
		return 0
	}
	files, err := checkSources.files(fl.Args())
	if err != nil {
		fmt.Fprintf(stderr, "ant check: %s\n", err)
		return 2
	}
	diags, loaded := ant.Check(files)
	if *jsonOut {
		writeJSONDiagnostics(stdout, diags)
	} else if len(diags) > 0 {
		printDiagnostics(stderr, diags)
	}
	switch {
	case !loaded:
		return 2
	case len(diags) > 0:
		return 1
	}
	return 0
}

// checkSources are the files ant check takes: any file named directly, and the .ant files under
// a directory, of which there must be one.
var checkSources = sources{keep: func(name string) bool { return strings.HasSuffix(name, ".ant") }, what: ".ant files"}

type jsonPos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type jsonDiagnostic struct {
	Path     string        `json:"path"`
	Line     int           `json:"line"`
	Column   int           `json:"column"`
	Severity string        `json:"severity,omitempty"`
	Code     string        `json:"code,omitempty"`
	Message  string        `json:"message"`
	Start    *jsonPos      `json:"start,omitempty"`
	End      *jsonPos      `json:"end,omitempty"`
	Help     string        `json:"help,omitempty"`
	Fixes    []jsonFix     `json:"fixes,omitempty"`
	Related  []jsonRelated `json:"related,omitempty"`
}

type jsonFix struct {
	Title string     `json:"title"`
	Auto  bool       `json:"auto"`
	Edits []jsonEdit `json:"edits"`
}

type jsonEdit struct {
	Path  string  `json:"path"`
	Start jsonPos `json:"start"`
	End   jsonPos `json:"end"`
	Text  string  `json:"text"`
}

type jsonRelated struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

// writeJSONDiagnostics writes each diagnostic as a JSON object on a line of its own: its position
// as the text output prints it, line and column 0 for an error about a whole file, its message,
// and the span the text output underlines.
func writeJSONDiagnostics(w io.Writer, diags []ant.Diagnostic) {
	for _, jd := range jsonDiagnostics(diags) {
		b, _ := json.Marshal(jd)
		w.Write(append(b, '\n'))
	}
}

// writeJSONCheckErrors writes the diagnostics as ant test -json events: each as ant check -json
// writes it, with the action "error".
func writeJSONCheckErrors(w io.Writer, diags []ant.Diagnostic) {
	for _, jd := range jsonDiagnostics(diags) {
		b, _ := json.Marshal(struct {
			Action string `json:"action"`
			jsonDiagnostic
		}{"error", jd})
		w.Write(append(b, '\n'))
	}
}

func jsonDiagnostics(diags []ant.Diagnostic) []jsonDiagnostic {
	p := &diag.Printer{Read: func(path string) (string, error) {
		b, err := os.ReadFile(path)
		return string(b), err
	}}
	out := make([]jsonDiagnostic, 0, len(diags))
	for _, d := range diags {
		jd := jsonDiagnostic{Path: d.Pos.Path, Line: d.Pos.Line, Column: d.Pos.Col, Severity: d.Severity, Code: d.Code, Message: d.Message}
		from, to := p.Range(d.Pos.Path, ast.Pos{Line: d.Pos.Line, Col: d.Pos.Col}, ast.Pos{Line: d.End.Line, Col: d.End.Col})
		if from.Line > 0 {
			jd.Start = &jsonPos{from.Line, from.Col}
			jd.End = &jsonPos{to.Line, to.Col}
		}
		jd.Help = d.Help
		for _, f := range d.Fixes {
			jf := jsonFix{Title: f.Title, Auto: f.Auto, Edits: []jsonEdit{}}
			for _, e := range f.Edits {
				jf.Edits = append(jf.Edits, jsonEdit{Path: e.Start.Path, Start: jsonPos{e.Start.Line, e.Start.Col}, End: jsonPos{e.End.Line, e.End.Col}, Text: e.Text})
			}
			jd.Fixes = append(jd.Fixes, jf)
		}
		for _, r := range d.Related {
			jd.Related = append(jd.Related, jsonRelated{Path: r.Pos.Path, Line: r.Pos.Line, Column: r.Pos.Col, Message: r.Message})
		}
		out = append(out, jd)
	}
	return out
}
