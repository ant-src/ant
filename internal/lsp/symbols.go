package lsp

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/parser"
)

// documentSymbols answers textDocument/documentSymbol: the module's imports, types with their
// fields and functions, functions, and top-level variables (spec/lsp.md section 4).
func (s *Server) documentSymbols(p *DocumentSymbolParams) []DocumentSymbol {
	d, ok := s.docs[p.TextDocument.URI]
	if !ok {
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
	// the functions declared on each type go under it
	onType := map[string][]DocumentSymbol{}
	for _, decl := range f.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok && fd.Owner != "" {
			onType[fd.Owner] = append(onType[fd.Owner], s.funcSymbol(a, d, fd))
		}
	}
	for _, decl := range f.Decls {
		switch x := decl.(type) {
		case *ast.FuncDecl:
			if x.Owner == "" {
				out = append(out, s.funcSymbol(a, d, x))
			}
		case *ast.TypeDecl:
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
			sym.Children = append(sym.Children, onType[x.Name]...)
			out = append(out, sym)
		case *ast.VarDecl:
			for _, nt := range targetNames(x.Target) {
				sym := DocumentSymbol{Name: nt.Name, Kind: SymbolVariable, Range: s.rng(d, x.Pos, x.End), SelectionRange: s.rng(d, nt.Pos, nt.End)}
				if o := a.info.Decls[nt]; o != nil {
					sym.Detail = typeString(o.Type())
				}
				out = append(out, sym)
			}
		}
	}
	return out
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
	seen := map[string]bool{}
	visit := func(path string, f *ast.File, d *document) {
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
	for _, d := range s.docs {
		if d.path == "" {
			continue
		}
		seen[d.path] = true
		a := s.analyze(d)
		visit(d.path, a.module().File, d)
	}
	for _, root := range s.roots {
		filepath.WalkDir(root, func(path string, e os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if e.IsDir() {
				if path != root && (strings.HasPrefix(e.Name(), ".") || e.Name() == "node_modules") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".ant") || seen[path] {
				return nil
			}
			seen[path] = true
			src, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			f, _, _ := parser.ParseRecover(path, string(src))
			visit(path, f, newDocument(fileURI(path), 0, string(src)))
			return nil
		})
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

var _ = check.Var
