package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
	fl.Usage = func() {
		fmt.Fprintln(stderr, "usage: ant check [-json] [path ...]")
		fl.PrintDefaults()
	}
	if err := fl.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	files, err := checkFiles(fl.Args())
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

// checkFiles lists the files the paths name: a file itself, whatever its name, or every .ant
// file under a directory. With no path, it is ".".
func checkFiles(paths []string) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			files = append(files, p)
			continue
		}
		found, err := antFiles(p, func(name string) bool { return strings.HasSuffix(name, ".ant") })
		if err != nil {
			return nil, err
		}
		files = append(files, found...)
	}
	return files, nil
}

// antFiles lists the files under dir whose names keep accepts, in lexical order, leaving out
// the directories whose names begin with a dot.
func antFiles(dir string, keep func(name string) bool) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != dir && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if keep(d.Name()) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

type jsonPos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type jsonDiagnostic struct {
	Path    string   `json:"path,omitempty"`
	Line    int      `json:"line,omitempty"`
	Column  int      `json:"column,omitempty"`
	Message string   `json:"message"`
	Start   *jsonPos `json:"start,omitempty"`
	End     *jsonPos `json:"end,omitempty"`
}

// writeJSONDiagnostics writes each diagnostic as a JSON object on a line of its own: its position
// as the text output prints it, its message, and the span the text output underlines.
func writeJSONDiagnostics(w io.Writer, diags []ant.Diagnostic) {
	p := &diag.Printer{Read: func(path string) (string, error) {
		b, err := os.ReadFile(path)
		return string(b), err
	}}
	for _, d := range diags {
		jd := jsonDiagnostic{Path: d.Pos.Path, Line: d.Pos.Line, Column: d.Pos.Col, Message: d.Message}
		from, to := p.Range(d.Pos.Path, ast.Pos{Line: d.Pos.Line, Col: d.Pos.Col}, ast.Pos{Line: d.End.Line, Col: d.End.Col})
		if from.Line > 0 {
			jd.Start = &jsonPos{from.Line, from.Col}
			jd.End = &jsonPos{to.Line, to.Col}
		}
		b, _ := json.Marshal(jd)
		w.Write(append(b, '\n'))
	}
}
