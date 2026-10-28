package lsp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/internal/token"
)

// reference is one mention of an object: the module and the span of the name.
type reference struct {
	path     string
	from, to ast.Pos
	decl     bool // the declaration itself
}

// referencesOf finds every mention of o in the program: the uses, and the declaration.
func (a *analysis) referencesOf(o *check.Object) []reference {
	var refs []reference
	in := a.info
	for path, m := range a.modules {
		ast.Inspect(m.File, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				if in.Objects[x] == o && !a.throughAlias(path, x) {
					refs = append(refs, reference{path, x.Pos, x.End, false})
				}
			case *ast.Field:
				if in.Members[x] == o && x.Name != "" {
					refs = append(refs, reference{path, x.NamePos, x.End, false})
				}
			case *ast.TypeName:
				for i, obj := range in.TypeNames[x] {
					if obj == o && i < len(x.PathPos) {
						refs = append(refs, reference{path, x.PathPos[i], after(x.PathPos[i], x.Path[i]), false})
					}
				}
			case *ast.RecordEntry:
				if in.Entries[x] == o && x.Name != "" {
					refs = append(refs, reference{path, x.Pos, after(x.Pos, x.Name), false})
				}
			case *ast.ImportName:
				if in.Decls[x] == o {
					refs = append(refs, reference{path, x.Pos, after(x.Pos, x.Name), false})
				}
			case *ast.Call:
				if ci := in.Calls[x]; ci != nil && ci.Callee != nil {
					if fd, ok := ci.Callee.Decl.(*ast.FuncDecl); ok {
						for i := range x.Args {
							arg := &x.Args[i]
							if arg.Name == "" {
								continue
							}
							for j := range fd.Params {
								if fd.Params[j].Name == arg.Name && in.Decls[&fd.Params[j]] == o {
									refs = append(refs, reference{path, arg.Pos, after(arg.Pos, arg.Name), false})
								}
							}
						}
					}
				}
			case *ast.FuncDecl:
				if in.Decls[x] == o {
					refs = append(refs, reference{path, x.NamePos, after(x.NamePos, x.Name), true})
				}
				if x.Owner != "" && a.named(path, x.Owner) == o {
					refs = append(refs, reference{path, x.OwnerPos, after(x.OwnerPos, x.Owner), false})
				}
			case *ast.TypeDecl:
				if in.Decls[x] == o {
					refs = append(refs, reference{path, x.NamePos, after(x.NamePos, x.Name), true})
				}
			case *ast.NameTarget:
				if in.Decls[x] == o {
					refs = append(refs, reference{path, x.Pos, x.End, true})
				}
			case *ast.Param:
				if in.Decls[x] == o {
					refs = append(refs, reference{path, x.Pos, after(x.Pos, x.Name), true})
				}
			case *ast.TypeParam:
				if in.Decls[x] == o {
					refs = append(refs, reference{path, x.Pos, after(x.Pos, x.Name), true})
				}
			case *ast.FieldType:
				if in.Decls[x] == o {
					refs = append(refs, reference{path, x.Pos, after(x.Pos, x.Name), true})
				}
			case *ast.Import:
				if x.Namespace != "" && in.Decls[x] == o {
					refs = append(refs, reference{path, x.NamespacePos, after(x.NamespacePos, x.Namespace), true})
				}
			}
			return true
		})
	}
	return refs
}

// throughAlias reports whether an identifier reaches its object through an import's alias, which
// has a name of its own (section 10): renaming the export leaves it as it is.
func (a *analysis) throughAlias(path string, id *ast.Ident) bool {
	m := a.modules[path]
	if m == nil || m.Resolve == nil {
		return false
	}
	in, ok := m.Resolve.Defs[id].(*ast.ImportName)
	return ok && in.Alias != in.Name
}

// references answers textDocument/references.
func (s *Server) references(p *ReferenceParams) []Location {
	d, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return nil
	}
	a := s.analyze(d)
	t := a.at(a.module().File.Path, d.posOf(d.offset(p.Position, s.utf16)))
	if t == nil || t.obj == nil {
		return nil
	}
	out := []Location{}
	for _, r := range a.referencesOf(t.obj) {
		if r.decl && !p.Context.IncludeDeclaration {
			continue
		}
		if loc, ok := a.location(r.path, r.from, r.to); ok {
			out = append(out, loc)
		}
	}
	return out
}

// highlight answers textDocument/documentHighlight: the mentions in this document.
func (s *Server) highlight(p *TextDocumentPositionParams) []DocumentHighlight {
	d, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return nil
	}
	a := s.analyze(d)
	path := a.module().File.Path
	t := a.at(path, d.posOf(d.offset(p.Position, s.utf16)))
	if t == nil || t.obj == nil {
		return nil
	}
	out := []DocumentHighlight{}
	for _, r := range a.referencesOf(t.obj) {
		if r.path == path {
			out = append(out, DocumentHighlight{Range: s.rng(d, r.from, r.to), Kind: 1})
		}
	}
	return out
}

// renamable says whether an object can be renamed, or why not (spec/lsp.md section 4).
func renamable(o *check.Object) *rpcError {
	switch {
	case o == nil:
		return &rpcError{Code: codeInvalidParams, Message: "nothing to rename here"}
	case o.Path == "prelude.ant" || o.Decl == nil && o.Kind != check.Module:
		return &rpcError{Code: codeInvalidParams, Message: "a name of the prelude cannot be renamed"}
	case o.Kind == check.Field:
		return &rpcError{Code: codeInvalidParams, Message: "a record field is not renamed"}
	case o.Kind == check.Namespace:
		return &rpcError{Code: codeInvalidParams, Message: "a namespace cannot be renamed"}
	}
	return nil
}

// validName says whether name can be given to a declaration: an identifier that is neither a
// keyword, a reserved word, nor a prelude name (section 10).
func validName(name string) *rpcError {
	if name == "" || name == "_" {
		return &rpcError{Code: codeInvalidParams, Message: "the new name must be an identifier"}
	}
	for i := 0; i < len(name); i++ {
		if !isIdentByte(name[i]) || i == 0 && name[0] >= '0' && name[0] <= '9' {
			return &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%q is not an identifier", name)}
		}
	}
	switch {
	case token.IsKeyword(name):
		return &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%s is a keyword", name)}
	case token.IsReserved(name):
		return &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%s is a reserved word", name)}
	case prelude.IsName(name):
		return &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%s is a prelude name and cannot be redeclared", name)}
	}
	return nil
}

// prepareRename answers textDocument/prepareRename: the range of the name, or why it cannot be
// renamed.
func (s *Server) prepareRename(p *TextDocumentPositionParams) (any, *rpcError) {
	d, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return nil, nil
	}
	a := s.analyze(d)
	t := a.at(a.module().File.Path, d.posOf(d.offset(p.Position, s.utf16)))
	if t == nil || t.obj == nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: "nothing to rename here"}
	}
	if err := renamable(t.obj); err != nil {
		return nil, err
	}
	if alias := a.aliasAt(t); alias != nil {
		return s.rng(d, t.from, t.to), nil
	}
	return s.rng(d, t.from, t.to), nil
}

// aliasAt reports the import name whose alias the target is or refers to, when the alias
// differs from the export's name: renaming it renames the alias, not the export.
func (a *analysis) aliasAt(t *target) *ast.ImportName {
	switch n := t.node.(type) {
	case *ast.ImportName:
		if n.Alias != n.Name {
			return n
		}
	case *ast.Ident:
		m := a.modules[t.path]
		if m == nil || m.Resolve == nil {
			return nil
		}
		if in, ok := m.Resolve.Defs[n].(*ast.ImportName); ok && in.Alias != in.Name {
			return in
		}
	}
	return nil
}

// rename answers textDocument/rename.
func (s *Server) rename(p *RenameParams) (any, *rpcError) {
	d, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return nil, nil
	}
	if err := validName(p.NewName); err != nil {
		return nil, err
	}
	a := s.analyze(d)
	path := a.module().File.Path
	t := a.at(path, d.posOf(d.offset(p.Position, s.utf16)))
	if t == nil || t.obj == nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: "nothing to rename here"}
	}
	if err := renamable(t.obj); err != nil {
		return nil, err
	}
	edit := WorkspaceEdit{Changes: map[string][]TextEdit{}}
	add := func(rpath string, from, to ast.Pos) {
		if rd := a.doc(rpath); rd != nil {
			edit.Changes[rd.uri] = append(edit.Changes[rd.uri], TextEdit{Range: s.rng(rd, from, to), NewText: p.NewName})
		}
	}
	if alias := a.aliasAt(t); alias != nil {
		// the alias and its uses in this module only
		add(path, alias.AliasPos, after(alias.AliasPos, alias.Alias))
		for _, id := range a.modules[path].Resolve.Uses[alias] {
			add(path, id.Pos, id.End)
		}
		return edit, nil
	}
	for _, r := range a.referencesOf(t.obj) {
		add(r.path, r.from, r.to)
	}
	if t.obj.Exported && t.obj.Kind != check.Module {
		// Every module of the workspace that imports the declaring module (spec/lsp.md
		// section 4): each is the root of a program of its own, in which the object is found
		// again by where it is declared.
		declPath, from, _, ok := a.declRange(t.obj)
		if ok {
			for _, other := range s.importers(declPath, a.root.uri) {
				oa := s.analyze(other)
				if same := oa.objectDeclaredAt(declPath, from); same != nil {
					for _, r := range oa.referencesOf(same) {
						if r.path != declPath {
							add2(&edit, s, oa, r, p.NewName)
						}
					}
				}
			}
		}
	}
	return edit, nil
}

// add2 adds a rename edit found in another program, once per document.
func add2(edit *WorkspaceEdit, s *Server, a *analysis, r reference, name string) {
	rd := a.doc(r.path)
	if rd == nil {
		return
	}
	te := TextEdit{Range: s.rng(rd, r.from, r.to), NewText: name}
	for _, have := range edit.Changes[rd.uri] {
		if have == te {
			return
		}
	}
	edit.Changes[rd.uri] = append(edit.Changes[rd.uri], te)
}

// objectDeclaredAt finds the object declared at a position of a module in this analysis.
func (a *analysis) objectDeclaredAt(path string, pos ast.Pos) *check.Object {
	mi, ok := a.info.Modules[path]
	if !ok {
		return nil
	}
	for _, o := range mi.Names {
		if p, from, _, ok := a.declRange(o); ok && p == path && from == pos {
			return o
		}
	}
	return nil
}

// importers lists the documents, open or in the workspace, that import the module at path,
// other than the one at skip: each becomes the root of a program the rename reaches into.
func (s *Server) importers(path, skip string) []*document {
	var out []*document
	seen := map[string]bool{}
	consider := func(d *document) {
		if d.uri == skip || seen[d.path] || d.path == "" {
			return
		}
		seen[d.path] = true
		if !strings.Contains(d.text, "import") {
			return
		}
		a := s.analyze(d)
		if _, ok := a.modules[path]; ok && a.module().File.Path != path {
			out = append(out, d)
		}
	}
	for _, d := range s.docs {
		consider(d)
	}
	for _, root := range s.roots {
		filepath.WalkDir(root, func(p string, e os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if e.IsDir() {
				if p != root && (strings.HasPrefix(e.Name(), ".") || e.Name() == "node_modules") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(p, ".ant") {
				return nil
			}
			if _, open := s.docs[fileURI(p)]; open {
				return nil
			}
			src, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			consider(newDocument(fileURI(p), 0, string(src)))
			return nil
		})
	}
	return out
}
