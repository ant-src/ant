package lsp

import (
	"fmt"
	"path/filepath"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/internal/token"
)

// reference is one mention of an object: the module and the span of the name. shorthand is set
// when the name is a record entry or destructuring written short, { x } for { x: x }, whose
// field must keep its name when the variable is renamed.
type reference struct {
	path      string
	from, to  ast.Pos
	decl      bool // the declaration itself
	shorthand bool
}

// newText is what a reference becomes when its name is renamed.
func (r reference) newText(name, field string) string {
	if r.shorthand {
		return field + ": " + name
	}
	return name
}

// referencesOf finds every mention of o in the program: the uses, and the declaration.
func (a *analysis) referencesOf(o *check.Object) []reference {
	var refs []reference
	in := a.info
	for path, m := range a.modules {
		shortIdents := map[*ast.Ident]bool{}
		shortTargets := map[*ast.NameTarget]bool{}
		ast.Inspect(m.File, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.RecordLit:
				for i := range x.Entries {
					e := &x.Entries[i]
					if id, ok := e.Value.(*ast.Ident); ok && e.Name != "" && id.Pos == e.Pos {
						shortIdents[id] = true
					}
				}
			case *ast.RecordTarget:
				for i := range x.Fields {
					f := &x.Fields[i]
					if nt, ok := f.Target.(*ast.NameTarget); ok && nt.Pos == f.Pos {
						shortTargets[nt] = true
					}
				}
			case *ast.Ident:
				if in.Objects[x] == o && !a.throughAlias(path, x) {
					refs = append(refs, reference{path, x.Pos, x.End, false, shortIdents[x]})
				}
			case *ast.Field:
				if in.Members[x] == o && x.Name != "" {
					refs = append(refs, reference{path: path, from: x.NamePos, to: x.End})
				}
			case *ast.TypeName:
				for i, obj := range in.TypeNames[x] {
					if obj == o && i < len(x.PathPos) && !(i == 0 && a.aliased(path, x.Path[0])) {
						refs = append(refs, reference{path: path, from: x.PathPos[i], to: after(x.PathPos[i], x.Path[i])})
					}
				}
			case *ast.RecordEntry:
				if in.Entries[x] == o && x.Name != "" {
					refs = append(refs, reference{path: path, from: x.Pos, to: after(x.Pos, x.Name)})
				}
			case *ast.ImportName:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.Pos, to: after(x.Pos, x.Name)})
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
									refs = append(refs, reference{path: path, from: arg.Pos, to: after(arg.Pos, arg.Name)})
								}
							}
						}
					}
				}
			case *ast.FuncDecl:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.NamePos, to: after(x.NamePos, x.Name), decl: true})
				}
				if x.Owner != "" && a.named(path, x.Owner) == o {
					refs = append(refs, reference{path: path, from: x.OwnerPos, to: after(x.OwnerPos, x.Owner)})
				}
			case *ast.TypeDecl:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.NamePos, to: after(x.NamePos, x.Name), decl: true})
				}
			case *ast.NameTarget:
				if in.Decls[x] == o {
					refs = append(refs, reference{path, x.Pos, x.End, true, shortTargets[x]})
				}
			case *ast.Param:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.Pos, to: after(x.Pos, x.Name), decl: true})
				}
			case *ast.TypeParam:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.Pos, to: after(x.Pos, x.Name), decl: true})
				}
			case *ast.FieldType:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.Pos, to: after(x.Pos, x.Name), decl: true})
				}
			case *ast.Import:
				if x.Namespace != "" && in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.NamespacePos, to: after(x.NamespacePos, x.Namespace), decl: true})
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

// aliased reports whether name is bound at the top level of the module at path by an import
// that renames what it imports, as a type name written through it is.
func (a *analysis) aliased(path, name string) bool {
	m := a.modules[path]
	if m == nil || m.Resolve == nil {
		return false
	}
	sc := m.Resolve.Scopes[m.File]
	if sc == nil {
		return false
	}
	for _, b := range sc.Names {
		if b.Name == name {
			in, ok := b.Node.(*ast.ImportName)
			return ok && in.Alias != in.Name
		}
	}
	return false
}

// references answers textDocument/references.
func (s *Server) references(p *ReferenceParams) []Location {
	d := s.doc(p.TextDocument.URI)
	if d == nil {
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
	d := s.doc(p.TextDocument.URI)
	if d == nil {
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
	d := s.doc(p.TextDocument.URI)
	if d == nil {
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
	d := s.doc(p.TextDocument.URI)
	if d == nil {
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
		// the alias and its uses in this module only, which must not clash with its names
		if sc := a.modules[path].Resolve.Scopes[a.module().File]; sc != nil {
			for _, b := range sc.Names {
				if b.Name == p.NewName && b.Node != ast.Node(alias) {
					return nil, &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%s is already declared in this module", p.NewName)}
				}
			}
		}
		add(path, alias.AliasPos, after(alias.AliasPos, alias.Alias))
		for _, id := range a.modules[path].Resolve.Uses[alias] {
			add(path, id.Pos, id.End)
		}
		return edit, nil
	}
	if err := a.collides(t.obj, p.NewName); err != nil {
		return nil, err
	}
	if err := a.importerClash(t.obj, p.NewName); err != nil {
		return nil, err // the module the rename starts from may import the name itself
	}
	for _, r := range a.referencesOf(t.obj) {
		add2(&edit, s, a, r, p.NewName, t.obj.Name)
	}
	if a.reachesImporters(t.obj) {
		// Every module of the workspace that imports the declaring module (spec/lsp.md
		// section 4): each is the root of a program of its own, in which the object is found
		// again by where it is declared. A name the importer binds must not clash there.
		declPath, from, _, ok := a.declRange(t.obj)
		if ok {
			for _, other := range s.importers(declPath, a.root.uri) {
				oa := s.analyze(other)
				same := oa.objectDeclaredAt(declPath, from)
				if same == nil {
					continue
				}
				if err := oa.importerClash(same, p.NewName); err != nil {
					return nil, err
				}
				for _, r := range oa.referencesOf(same) {
					if r.path != declPath {
						add2(&edit, s, oa, r, p.NewName, t.obj.Name)
					}
				}
			}
		}
	}
	return edit, nil
}

// reachesImporters reports whether renaming o changes other modules: an export, or a
// parameter of an exported function, which named arguments elsewhere spell.
func (a *analysis) reachesImporters(o *check.Object) bool {
	switch o.Kind {
	case check.Module:
		return false
	case check.Param:
		if fd := a.enclosingFunc(o); fd != nil {
			return fd.Export
		}
		return false
	}
	return o.Exported
}

// enclosingFunc is the declared function whose parameter o is, or nil.
func (a *analysis) enclosingFunc(o *check.Object) *ast.FuncDecl {
	param, ok := o.Decl.(*ast.Param)
	if !ok {
		return nil
	}
	f := a.file(a.pathOf(o))
	if f == nil {
		return nil
	}
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok {
			for i := range fd.Params {
				if &fd.Params[i] == param {
					return fd
				}
			}
		}
	}
	return nil
}

// importerClash reports a name the root module of this analysis, which imports o, already
// binds, when it imports o under its own name.
func (a *analysis) importerClash(o *check.Object, name string) *rpcError {
	m := a.module()
	if m.Resolve == nil {
		return nil
	}
	sc := m.Resolve.Scopes[m.File]
	if sc == nil {
		return nil
	}
	imported := false
	for _, b := range sc.Names {
		if in, ok := b.Node.(*ast.ImportName); ok && a.info.Decls[in] == o && in.Alias == in.Name {
			imported = true
		}
	}
	if !imported {
		return nil
	}
	for _, b := range sc.Names {
		if b.Name == name {
			if in, ok := b.Node.(*ast.ImportName); ok && a.info.Decls[in] == o {
				continue
			}
			return &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%s is already declared in %s, which imports %s", name, filepath.Base(m.File.Path), o.Name)}
		}
	}
	return nil
}

// add2 adds a rename edit, once per document and range.
func add2(edit *WorkspaceEdit, s *Server, a *analysis, r reference, name, field string) {
	rd := a.doc(r.path)
	if rd == nil {
		return
	}
	te := TextEdit{Range: s.rng(rd, r.from, r.to), NewText: r.newText(name, field)}
	for _, have := range edit.Changes[rd.uri] {
		if have == te {
			return
		}
	}
	edit.Changes[rd.uri] = append(edit.Changes[rd.uri], te)
}

// objectDeclaredAt finds the object whose name is declared at a position of a module in this
// analysis: a top-level declaration, or a parameter.
func (a *analysis) objectDeclaredAt(path string, pos ast.Pos) *check.Object {
	f := a.file(path)
	if f == nil {
		return nil
	}
	var found *check.Object
	ast.Inspect(f, func(n ast.Node) bool {
		if found != nil {
			return false
		}
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.NamePos == pos {
				found = a.info.Decls[x]
			}
		case *ast.TypeDecl:
			if x.NamePos == pos {
				found = a.info.Decls[x]
			}
		case *ast.NameTarget:
			if x.Pos == pos {
				found = a.info.Decls[x]
			}
		case *ast.Param:
			if x.Pos == pos {
				found = a.info.Decls[x]
			}
		}
		return true
	})
	return found
}

// importers lists the documents, open or in the workspace, whose imports lead to the module at
// path, other than the one at skip: each becomes the root of a program the rename reaches into.
// Which files import it is read from their parsed imports, so only those are analyzed.
func (s *Server) importers(path, skip string) []*document {
	var out []*document
	for _, f := range s.workspaceFiles() {
		if f.doc.uri == skip || f.doc.path == path {
			continue
		}
		imports := false
		for _, imp := range f.file.Imports {
			if sameFile(filepath.Join(filepath.Dir(f.doc.path), filepath.FromSlash(imp.Path)), path) {
				imports = true
			}
		}
		if imports {
			out = append(out, f.doc)
		}
	}
	return out
}

// collides reports a name already bound where the renamed binding is declared or used, which
// the rename would shadow or clash with (section 10): in the scopes around its declaration and
// around each use in its module, and among the module's top-level names.
func (a *analysis) collides(o *check.Object, name string) *rpcError {
	declPath, from, _, ok := a.declRange(o)
	if !ok {
		return nil
	}
	if declPath == "" {
		declPath = a.pathOf(o)
	}
	m := a.modules[declPath]
	if m == nil {
		return nil
	}
	clash := &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%s is already declared where %s is used", name, o.Name)}
	if mi, ok := a.info.Modules[declPath]; ok {
		for _, other := range mi.Names {
			if other.Name == name {
				return clash
			}
		}
	}
	if m.Resolve == nil {
		return nil
	}
	// the scopes around the declaration and around each use
	var positions []ast.Pos
	positions = append(positions, from)
	for _, r := range a.referencesOf(o) {
		if r.path == declPath {
			positions = append(positions, r.from)
		}
	}
	for _, pos := range positions {
		for _, n := range ast.Path(m.File, pos) {
			sc, ok := m.Resolve.Scopes[n]
			if !ok {
				continue
			}
			for _, b := range sc.Names {
				if b.Name == name && b.Node != o.Decl {
					return clash
				}
			}
		}
	}
	return nil
}
