package main

import (
	"context"
	"fmt"
	"io"

	"github.com/ant-src/ant/internal/lsp"
)

// lspCommand is ant lsp (spec/lsp.md section 1): it serves the Language Server Protocol on
// standard input and standard output until the client sends exit, logging to standard error.
func lspCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: ant lsp")
		return 2
	}
	return lsp.Serve(context.Background(), stdin, stdout, stderr)
}
