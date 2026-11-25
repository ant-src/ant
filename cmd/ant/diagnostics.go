package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"ant-lang.org"
	"ant-lang.org/ast"
	"ant-lang.org/internal/diag"
)

// printDiagnostics writes each diagnostic, its code in brackets after the message, with the
// source it is about, then how many there were; colored when w is a terminal, unless NO_COLOR is
// set or the terminal is dumb.
func printDiagnostics(w io.Writer, diags []ant.Diagnostic) {
	p := &diag.Printer{Color: colors(w), Read: func(path string) (string, error) {
		b, err := os.ReadFile(path)
		return string(b), err
	}}
	for _, d := range diags {
		p.PrintSpan(w, d.Pos.Path, ast.Pos{Line: d.Pos.Line, Col: d.Pos.Col}, ast.Pos{Line: d.End.Line, Col: d.End.Col}, withCode(d), notes(d)...)
	}
	errs := len(ant.Errors(diags))
	p.Summary(w, errs, len(diags)-errs)
}

// printLoadError writes why a program could not be loaded: its diagnostics with their code.
func printLoadError(w io.Writer, err error) {
	var le *ant.LoadError
	if errors.As(err, &le) && len(le.Diagnostics) > 0 {
		printDiagnostics(w, le.Diagnostics)
		return
	}
	fmt.Fprintln(w, err)
}

func colors(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor || os.Getenv("TERM") == "dumb" {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// withCode is the message of d followed by its code in brackets, as spec/check.md section 2 has
// ant check print it; a diagnostic with no code, which is not about the program, prints bare.
func withCode(d ant.Diagnostic) string {
	msg := d.Message
	if d.Severity == "warning" {
		msg = "warning: " + msg
	}
	if d.Code == "" {
		return msg
	}
	return msg + " [" + d.Code + "]"
}

// notes are the lines printed under a diagnostic's source: its help, the title of each fix, and
// each related place (spec/check.md section 2).
func notes(d ant.Diagnostic) []string {
	var out []string
	if d.Help != "" {
		out = append(out, "help: "+d.Help)
	}
	for _, f := range d.Fixes {
		out = append(out, "fix: "+f.Title)
	}
	for _, r := range d.Related {
		out = append(out, fmt.Sprintf("note: %s: %s", r.Pos, r.Message))
	}
	return out
}
