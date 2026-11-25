// Package doc prints declarations with their doc comments, as ant doc does (spec/doc.md).
package doc

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"ant-lang.org/ast"
	"ant-lang.org/internal/highlight"
	"ant-lang.org/internal/near"
)

// Printer renders documentation, preserving source line breaks and spacing. Color highlights
// signatures and examples and emphasizes namespace headings; the zero value prints no escapes.
type Printer struct {
	Color bool
}

// Print prints what name names among decls: a declaration with its doc, a namespace's section
// label from namespaceDocs, and the signatures a namespace or a type holds. It reports whether name
// named anything. A name is matched without regard to case, and Owner.name names a member.
func (p *Printer) Print(w io.Writer, decls []ast.Decl, name string, namespaceDocs map[string]string) bool {
	owner, member, dotted := strings.Cut(name, ".")
	decls = withInterfaceFuncs(decls)
	found := false
	for _, d := range decls {
		o, n := declName(d)
		if dotted && strings.EqualFold(o, owner) && strings.EqualFold(n, member) ||
			!dotted && o == "" && strings.EqualFold(n, name) {
			p.PrintDoc(w, Sig(d), docOf(d))
			found = true
		}
	}
	if !dotted && !found {
		for ns, doc := range namespaceDocs {
			if !strings.EqualFold(ns, name) || doc == "" {
				continue
			}
			if p.Color {
				ns = "\x1b[1m" + ns + highlight.Reset
			}
			fmt.Fprintln(w, ns)
			fmt.Fprintln(w)
			p.PrintIndented(w, doc)
			found = true
			break
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
				p.PrintSignature(w, Sig(d))
				found = true
			}
		}
	}
	return found
}

// Suggest is what to look at instead when name names nothing among decls (spec/doc.md section 1):
// the declarations of other owners with its member name, else the names nearest to it, else, for
// Owner.name with a known owner, the owner's listing; "" when there is nothing to suggest.
func Suggest(decls []ast.Decl, name string, namespaceDocs map[string]string) string {
	var names []string
	owners := map[string]string{}
	for ns := range namespaceDocs {
		names = append(names, ns)
		owners[strings.ToLower(ns)] = ns
	}
	for _, d := range withInterfaceFuncs(decls) {
		o, n := declName(d)
		switch {
		case n == "":
		case o == "":
			names = append(names, n)
		default:
			names = append(names, o+"."+n)
			owners[strings.ToLower(o)] = o
		}
	}
	sort.Strings(names)
	owner, member, dotted := strings.Cut(name, ".")
	if !dotted {
		member = name
	}
	// the same member of another owner first, as String.join for Array.join, nearest first
	var same []string
	for _, n := range names {
		if _, m, ok := strings.Cut(n, "."); ok && strings.EqualFold(m, member) {
			same = append(same, n)
		}
	}
	lower := strings.ToLower(name)
	sort.SliceStable(same, func(i, j int) bool {
		return near.Distance(lower, strings.ToLower(same[i])) < near.Distance(lower, strings.ToLower(same[j]))
	})
	if len(same) > 3 {
		same = same[:3]
	}
	if len(same) == 0 {
		same = near.Nearest(name, names)
	}
	if len(same) > 0 {
		return "did you mean " + strings.Join(same, " or ") + "?"
	}
	if o, ok := owners[strings.ToLower(owner)]; dotted && ok {
		return "ant doc " + o + " lists its members"
	}
	return ""
}

// withInterfaceFuncs adds, after each interface, its functions as declarations the interface owns
// (spec/doc.md section 1), so that Shape.area finds one and Shape lists them.
func withInterfaceFuncs(decls []ast.Decl) []ast.Decl {
	var out []ast.Decl
	for _, d := range decls {
		out = append(out, d)
		if id, ok := d.(*ast.InterfaceDecl); ok {
			for _, f := range id.Funcs {
				out = append(out, &ast.FuncDecl{Pos: f.Pos, End: f.End, Export: id.Export, Owner: id.Name, OwnerPos: id.NamePos,
					Name: f.Name, NamePos: f.NamePos, Params: f.Params, Return: f.Return, Doc: f.Doc, Sig: f.Sig})
			}
		}
	}
	return out
}

// PrintDoc prints a declaration's signature and, indented, its doc.
func (p *Printer) PrintDoc(w io.Writer, sig, doc string) {
	p.PrintSignature(w, sig)
	if doc != "" {
		fmt.Fprintln(w)
		p.PrintIndented(w, doc)
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
	case *ast.InterfaceDecl:
		return "", d.Name
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
	case *ast.InterfaceDecl:
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
	case *ast.InterfaceDecl:
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
	case *ast.InterfaceDecl:
		return d.Export
	case *ast.VarDecl:
		return d.Export
	}
	return false
}
