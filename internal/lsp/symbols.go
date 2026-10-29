package lsp

import (
	"sort"
	"strings"

	"github.com/ant-src/ant/ast"
)

// documentSymbols answers textDocument/documentSymbol: the module's imports, types with their
// fields and functions, functions, and top-level variables (spec/lsp.md section 4). In a
// declaration file, the types and constants an owner qualifies go under it too, and an owner
// with no type declaration of its own is a namespace, listed where its first member is.
func (s *Server) documentSymbols(p *DocumentSymbolParams) []DocumentSymbol {
	d := s.doc(p.TextDocument.URI)
	if d == nil {
		return nil
	}
	a := s.analyze(d)
	f := a.module().File
	out := []DocumentSymbol{}
	for _, imp := range f.Imports {
		name := imp.Path
		if imp.Namespace != "" {
			name = imp.Namespace
		}
		out = append(out, DocumentSymbol{Name: name, Detail: imp.Path, Kind: SymbolModule,
			Range: s.rng(d, imp.Pos, imp.End), SelectionRange: s.rng(d, imp.PathPos, imp.PathEnd)})
	}
	// the declarations on each owner go under it
	onOwner := map[string][]DocumentSymbol{}
	isType := map[string]bool{}
	for _, decl := range f.Decls {
		switch x := decl.(type) {
		case *ast.FuncDecl:
			if x.Owner != "" {
				onOwner[x.Owner] = append(onOwner[x.Owner], s.funcSymbol(a, d, x))
			}
		case *ast.TypeDecl:
			if x.Owner != "" {
				onOwner[x.Owner] = append(onOwner[x.Owner], s.typeSymbol(a, d, x, nil))
			} else {
				isType[x.Name] = true
			}
		case *ast.VarDecl:
			if x.Owner != "" {
				onOwner[x.Owner] = append(onOwner[x.Owner], s.varSymbols(a, d, x)...)
			}
		}
	}
	listed := map[string]bool{}
	namespace := func(owner string) {
		if !isType[owner] && !listed[owner] {
			listed[owner] = true
			out = append(out, namespaceSymbol(owner, onOwner[owner]))
		}
	}
	for _, decl := range f.Decls {
		switch x := decl.(type) {
		case *ast.FuncDecl:
			if x.Owner == "" {
				out = append(out, s.funcSymbol(a, d, x))
			} else {
				namespace(x.Owner)
			}
		case *ast.TypeDecl:
			if x.Owner == "" {
				out = append(out, s.typeSymbol(a, d, x, onOwner[x.Name]))
			} else {
				namespace(x.Owner)
			}
		case *ast.VarDecl:
			if x.Owner == "" {
				out = append(out, s.varSymbols(a, d, x)...)
			} else {
				namespace(x.Owner)
			}
		}
	}
	return out
}

// typeSymbol is a type's symbol: a struct with its fields as children when it is a record, with
// the symbols declared on it after them.
func (s *Server) typeSymbol(a *analysis, d *document, x *ast.TypeDecl, on []DocumentSymbol) DocumentSymbol {
	sym := DocumentSymbol{Name: x.Name, Kind: SymbolClass, Range: s.rng(d, x.Pos, x.End),
		SelectionRange: s.nameRange(d, x.NamePos, x.Name)}
	if rt, ok := x.Body.(*ast.RecordType); ok {
		sym.Kind = SymbolStruct
		for i := range rt.Fields {
			fld := &rt.Fields[i]
			sym.Children = append(sym.Children, DocumentSymbol{Name: fld.Name, Detail: a.typeText(d, fld.Type), Kind: SymbolField,
				Range: s.rng(d, fld.Pos, fld.End), SelectionRange: s.nameRange(d, fld.Pos, fld.Name)})
		}
	} else if o := a.info.Decls[x]; o != nil && o.TypeDecl() != nil {
		sym.Detail = typeText(x.Body, d)
	}
	sym.Children = append(sym.Children, on...)
	return sym
}

// varSymbols are the symbols of the names a var binds, with their types.
func (s *Server) varSymbols(a *analysis, d *document, x *ast.VarDecl) []DocumentSymbol {
	var out []DocumentSymbol
	for _, nt := range targetNames(x.Target) {
		sym := DocumentSymbol{Name: nt.Name, Kind: SymbolVariable, Range: s.rng(d, x.Pos, x.End), SelectionRange: s.rng(d, nt.Pos, nt.End)}
		o := a.info.Decls[nt]
		if o == nil {
			o = a.info.Decls[x] // a declaration file's constant is declared by its var
		}
		if o != nil {
			sym.Detail = typeString(o.Type())
		}
		out = append(out, sym)
	}
	return out
}

// namespaceSymbol groups the symbols declared on an owner that is not a type, spanning them.
func namespaceSymbol(name string, members []DocumentSymbol) DocumentSymbol {
	sym := DocumentSymbol{Name: name, Kind: SymbolNamespace, Children: members}
	for i, m := range members {
		if i == 0 || before(m.Range.Start, sym.Range.Start) {
			sym.Range.Start = m.Range.Start
		}
		if i == 0 || before(sym.Range.End, m.Range.End) {
			sym.Range.End = m.Range.End
		}
	}
	sym.SelectionRange = sym.Range
	if len(members) > 0 {
		sym.SelectionRange = members[0].SelectionRange
	}
	return sym
}

// before reports whether a comes before b.
func before(a, b Position) bool {
	return a.Line < b.Line || a.Line == b.Line && a.Character < b.Character
}

func (s *Server) funcSymbol(a *analysis, d *document, fd *ast.FuncDecl) DocumentSymbol {
	sym := DocumentSymbol{Name: fd.Name, Kind: SymbolFunction, Range: s.rng(d, fd.Pos, fd.End), SelectionRange: s.nameRange(d, fd.NamePos, fd.Name)}
	if fd.Owner != "" {
		sym.Kind = SymbolMethod
	}
	if o := a.info.Decls[fd]; o != nil {
		sym.Detail = strings.TrimPrefix(a.signatureLabel(o), fd.Owner+"."+fd.Name)
		sym.Detail = strings.TrimPrefix(sym.Detail, fd.Name)
	}
	return sym
}

// targetNames lists the names a target binds.
func targetNames(t ast.Target) []*ast.NameTarget {
	switch t := t.(type) {
	case *ast.NameTarget:
		if t.Name == "_" {
			return nil
		}
		return []*ast.NameTarget{t}
	case *ast.RecordTarget:
		var out []*ast.NameTarget
		for _, f := range t.Fields {
			out = append(out, targetNames(f.Target)...)
		}
		return out
	case *ast.ArrayTarget:
		var out []*ast.NameTarget
		for _, e := range t.Elems {
			out = append(out, targetNames(e)...)
		}
		if t.Rest != nil {
			out = append(out, targetNames(t.Rest)...)
		}
		return out
	}
	return nil
}

// typeText is a type expression as written.
func typeText(t ast.Type, d *document) string {
	if t == nil || d == nil {
		return ""
	}
	return d.text[d.offsetOf(ast.Start(t)):d.offsetOf(t.EndPosition())]
}

func (a *analysis) typeText(d *document, t ast.Type) string { return typeText(t, d) }

// workspaceSymbols answers workspace/symbol: the top-level declarations of every .ant file under
// the workspace folders whose name contains the query, matched without regard to case.
func (s *Server) workspaceSymbols(p *WorkspaceSymbolParams) []SymbolInformation {
	query := strings.ToLower(p.Query)
	out := []SymbolInformation{}
	visit := func(f *ast.File, d *document) {
		for _, decl := range f.Decls {
			name, kind, pos, end := declInfo(decl)
			if name == "" || !strings.Contains(strings.ToLower(name), query) {
				continue
			}
			container := ""
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Owner != "" {
				container = fd.Owner
			}
			out = append(out, SymbolInformation{Name: name, Kind: kind, ContainerName: container,
				Location: Location{URI: d.uri, Range: s.rng(d, pos, end)}})
		}
	}
	for _, f := range s.workspaceFiles() {
		visit(f.file, f.doc)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// declInfo names a declaration for the symbol listings.
func declInfo(decl ast.Decl) (name string, kind int, pos, end ast.Pos) {
	switch x := decl.(type) {
	case *ast.FuncDecl:
		kind = SymbolFunction
		if x.Owner != "" {
			kind = SymbolMethod
		}
		return x.Name, kind, x.NamePos, after(x.NamePos, x.Name)
	case *ast.TypeDecl:
		kind = SymbolClass
		if _, ok := x.Body.(*ast.RecordType); ok {
			kind = SymbolStruct
		}
		return x.Name, kind, x.NamePos, after(x.NamePos, x.Name)
	case *ast.VarDecl:
		if nt, ok := x.Target.(*ast.NameTarget); ok {
			return nt.Name, SymbolVariable, nt.Pos, nt.End
		}
	}
	return "", 0, ast.Pos{}, ast.Pos{}
}
