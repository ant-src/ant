// Package doc prints declarations with their doc comments, as ant doc does (spec/doc.md).
package doc

import (
	"fmt"
	"io"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/prelude"
)

// Print prints what name names among decls: a declaration with its doc, a namespace's doc
// when inPrelude, and the signatures a namespace or a type holds. It reports whether name
// named anything. A name is matched without regard to case, and Owner.name names a member.
func Print(w io.Writer, decls []ast.Decl, name string, inPrelude bool) bool {
	owner, member, dotted := strings.Cut(name, ".")
	found := false
	for _, d := range decls {
		o, n := declName(d)
		if dotted && strings.EqualFold(o, owner) && strings.EqualFold(n, member) ||
			!dotted && o == "" && strings.EqualFold(n, name) {
			PrintDoc(w, Sig(d), docOf(d))
			found = true
		}
	}
	if !dotted && !found && inPrelude {
		if ns, doc := prelude.NamespaceDoc(name); doc != "" {
			PrintDoc(w, ns, doc)
			found = true
		}
	}
	if !dotted {
		own := found
		for _, d := range decls {
			if o, _ := declName(d); strings.EqualFold(o, name) {
				if own {
					fmt.Fprintln(w)
					own = false
				}
				fmt.Fprintln(w, Sig(d))
				found = true
			}
		}
	}
	return found
}

// PrintDoc prints a declaration's signature, or a namespace's name, and, indented, its doc.
func PrintDoc(w io.Writer, sig, doc string) {
	fmt.Fprintln(w, sig)
	if doc != "" {
		fmt.Fprintln(w)
		PrintIndented(w, doc)
	}
}

// PrintIndented prints a doc with every non-empty line indented by four spaces.
func PrintIndented(w io.Writer, doc string) {
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

// Sig is a declaration's signature as written.
func Sig(d ast.Decl) string {
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

// Exported reports whether a declaration is exported.
func Exported(d ast.Decl) bool {
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
