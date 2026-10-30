package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/doc"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/parser"
)

// docCommand is ant doc (spec/doc.md section 1): it prints a declaration of the prelude or of a
// module with its doc comment, a module's doc, or the signatures a namespace, a type, or a module
// holds.
func docCommand(args []string, stdout, stderr io.Writer) int {
	var decls []ast.Decl
	var name string
	inPrelude := false
	switch {
	case len(args) == 1 && !strings.HasSuffix(args[0], ".ant"):
		f, err := prelude.Decls()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		decls, name, inPrelude = f.Decls, args[0], true
	case (len(args) == 1 || len(args) == 2) && strings.HasSuffix(args[0], ".ant"):
		src, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		f, err := parser.Parse(args[0], string(src), parser.ModeOf(args[0]))
		if err != nil {
			fmt.Fprintf(stderr, "%s:%v\n", args[0], err)
			return 2
		}
		// every declaration of a declaration file is visible (section 10, Declaration files)
		for _, d := range f.Decls {
			if f.Declarations || doc.Exported(d) {
				decls = append(decls, d)
			}
		}
		if len(args) == 1 {
			if f.Doc == "" && len(decls) == 0 {
				if f.Declarations {
					fmt.Fprintf(stderr, "ant doc: %s has no doc and declares nothing\n", args[0])
				} else {
					fmt.Fprintf(stderr, "ant doc: %s has no doc and exports nothing\n", args[0])
				}
				return 1
			}
			if f.Doc != "" {
				doc.PrintIndented(stdout, f.Doc)
				if len(decls) > 0 {
					fmt.Fprintln(stdout)
				}
			}
			for _, d := range decls {
				fmt.Fprintln(stdout, doc.Sig(d))
			}
			return 0
		}
		name = args[1]
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if !doc.Print(stdout, decls, name, inPrelude) {
		fmt.Fprintf(stderr, "ant doc: no declaration %s\n", name)
		return 1
	}
	return 0
}
