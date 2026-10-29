package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ant-src/ant/ast"
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
			if f.Declarations || exported(d) {
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
				printIndented(stdout, f.Doc)
				if len(decls) > 0 {
					fmt.Fprintln(stdout)
				}
			}
			for _, d := range decls {
				fmt.Fprintln(stdout, sig(d))
			}
			return 0
		}
		name = args[1]
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
	owner, member, dotted := strings.Cut(name, ".")
	found := false
	for _, d := range decls {
		o, n := declName(d)
		if dotted && strings.EqualFold(o, owner) && strings.EqualFold(n, member) ||
			!dotted && o == "" && strings.EqualFold(n, name) {
			printDoc(stdout, sig(d), docOf(d))
			found = true
		}
	}
	if !dotted && !found && inPrelude {
		if ns, doc := prelude.NamespaceDoc(name); doc != "" {
			printDoc(stdout, ns, doc)
			found = true
		}
	}
	if !dotted {
		own := found
		for _, d := range decls {
			if o, _ := declName(d); strings.EqualFold(o, name) {
				if own {
					fmt.Fprintln(stdout)
					own = false
				}
				fmt.Fprintln(stdout, sig(d))
				found = true
			}
		}
	}
	if !found {
		fmt.Fprintf(stderr, "ant doc: no declaration %s\n", name)
		return 1
	}
	return 0
}

// printDoc prints a declaration's signature, or a namespace's name, and, indented, its doc.
func printDoc(w io.Writer, sig, doc string) {
	fmt.Fprintln(w, sig)
	if doc != "" {
		fmt.Fprintln(w)
		printIndented(w, doc)
	}
}

// printIndented prints a doc with every non-empty line indented by four spaces.
func printIndented(w io.Writer, doc string) {
	for line := range strings.SplitSeq(doc, "\n") {
		if line != "" {
			line = "    " + line
		}
		fmt.Fprintln(w, line)
	}
}

// declName is the owner and name a declaration is looked up by; a var that destructures has
// no name.
func declName(d ast.Decl) (owner, name string) {
	switch d := d.(type) {
	case *ast.FuncDecl:
		return d.Owner, d.Name
	case *ast.TypeDecl:
		return d.Owner, d.Name
	case *ast.VarDecl:
		if t, ok := d.Target.(*ast.NameTarget); ok {
			return d.Owner, t.Name
		}
	}
	return "", ""
}

func sig(d ast.Decl) string {
	switch d := d.(type) {
	case *ast.FuncDecl:
		return d.Sig
	case *ast.TypeDecl:
		return d.Sig
	case *ast.VarDecl:
		return d.Sig
	}
	return ""
}

func docOf(d ast.Decl) string {
	switch d := d.(type) {
	case *ast.FuncDecl:
		return d.Doc
	case *ast.TypeDecl:
		return d.Doc
	case *ast.VarDecl:
		return d.Doc
	}
	return ""
}

func exported(d ast.Decl) bool {
	switch d := d.(type) {
	case *ast.FuncDecl:
		return d.Export
	case *ast.TypeDecl:
		return d.Export
	case *ast.VarDecl:
		return d.Export
	}
	return false
}
