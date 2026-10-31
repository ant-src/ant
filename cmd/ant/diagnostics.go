package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/diag"
)

// printDiagnostics writes each diagnostic with the code it is about, then how many there were;
// colored when w is a terminal, unless NO_COLOR is set or the terminal is dumb.
func printDiagnostics(w io.Writer, diags []ant.Diagnostic) {
	p := &diag.Printer{Color: colors(w), Read: func(path string) (string, error) {
		b, err := os.ReadFile(path)
		return string(b), err
	}}
	for _, d := range diags {
		p.Print(w, d.Pos.Path, ast.Pos{Line: d.Pos.Line, Col: d.Pos.Col}, d.Message)
	}
	p.Summary(w, len(diags))
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
