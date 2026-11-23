package lsp

import (
	"fmt"
	"path/filepath"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/internal/resolve"
	"github.com/ant-src/ant/internal/token"
	"github.com/ant-src/ant/parser"
)

// reference is one mention of an object: the module and the span of the name. shorthand is set
// when the name is a record entry or destructuring written short, { x } for { x: x }, whose
// field must keep its name when the variable is renamed.
type reference struct {
	path      string
	from, to  ast.Pos
	decl      bool // the declaration itself
	shorthand bool
	// viaAlias is set for a use through an import's alias, which has a name of its own (section
	// 10): renaming the export leaves it as it is.
	viaAlias bool
	// qualified is set for a name looked up in a type, a module, or a function's parameters
	// rather than in the scopes around it: a member, a named argument, the export an alias
	// imports.
	qualified bool
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
				if in.Objects[x] == o {
					refs = append(refs, reference{path: path, from: x.Pos, to: x.End, shorthand: shortIdents[x], viaAlias: a.throughAlias(path, x)})
				}
			case *ast.Field:
				if in.Members[x] == o && x.Name != "" {
					refs = append(refs, reference{path: path, from: x.NamePos, to: x.End, qualified: true})
				}
			case *ast.TypeName:
				for i, obj := range in.TypeNames[x] {
					if obj == o && i < len(x.PathPos) {
						refs = append(refs, reference{path: path, from: x.PathPos[i], to: after(x.PathPos[i], x.Path[i]), viaAlias: i == 0 && a.aliased(path, x.Path[0]), qualified: i > 0})
					}
				}
			case *ast.RecordEntry:
				if in.Entries[x] == o && x.Name != "" {
					refs = append(refs, reference{path: path, from: x.Pos, to: after(x.Pos, x.Name), qualified: true})
				}
			case *ast.ImportName:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.Pos, to: after(x.Pos, x.Name), qualified: x.Alias != x.Name})
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
									refs = append(refs, reference{path: path, from: arg.Pos, to: after(arg.Pos, arg.Name), qualified: true})
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
			case *ast.InterfaceDecl:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.NamePos, to: after(x.NamePos, x.Name), decl: true})
				}
			case *ast.InterfaceFunc:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.NamePos, to: after(x.NamePos, x.Name), decl: true})
				}
			case *ast.NameTarget:
				if in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: x.Pos, to: x.End, decl: true, shorthand: shortTargets[x]})
				}
			case *ast.VarDecl:
				// a constant of a declaration file, let Owner.NAME: T, is declared by the whole
				// declaration and named by its target
				if nt, ok := x.Target.(*ast.NameTarget); ok && x.Owner != "" && in.Decls[x] == o {
					refs = append(refs, reference{path: path, from: nt.Pos, to: nt.End, decl: true})
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

// aliasRefs finds the mentions of an import's alias in the module at path that declares it: the
// alias, and the identifiers and type names written through it.
func (a *analysis) aliasRefs(path string, alias *ast.ImportName) []reference {
	m := a.modules[path]
	refs := []reference{{path: path, from: alias.AliasPos, to: after(alias.AliasPos, alias.Alias), decl: true}}
	if m == nil || m.Resolve == nil {
		return refs
	}
	obj := a.info.Decls[alias]
	short := map[*ast.Ident]bool{}
	ast.Inspect(m.File, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.RecordLit:
			for i := range x.Entries {
				e := &x.Entries[i]
				if id, ok := e.Value.(*ast.Ident); ok && e.Name != "" && id.Pos == e.Pos {
					short[id] = true
				}
			}
		case *ast.Ident:
			if m.Resolve.Defs[x] == ast.Node(alias) {
				refs = append(refs, reference{path: path, from: x.Pos, to: x.End, shorthand: short[x]})
			}
		case *ast.TypeName:
			if objs := a.info.TypeNames[x]; len(objs) > 0 && obj != nil && objs[0] == obj && len(x.PathPos) > 0 && x.Path[0] == alias.Alias {
				refs = append(refs, reference{path: path, from: x.PathPos[0], to: after(x.PathPos[0], x.Path[0])})
			}
		}
		return true
	})
	return refs
}

// refsAt finds the mentions of what the target names: of an alias when it is on one, else of
// its object.
func (a *analysis) refsAt(t *target) []reference {
	if alias := a.aliasAt(t); alias != nil {
		return a.aliasRefs(t.path, alias)
	}
	return a.referencesOf(t.obj)
}

// references answers textDocument/references: the mentions in the declaration's program and,
// for one that reaches importers, in theirs (spec/lsp.md section 4).
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
	for _, m := range s.mentionsAt(a, t) {
		if m.decl && !p.Context.IncludeDeclaration {
			continue
		}
		if loc, ok := m.a.location(m.path, m.from, m.to); ok {
			out = append(out, loc)
		}
	}
	return out
}

// program is a program a declaration's uses are searched in, with the declaration's object in it.
type program struct {
	a *analysis
	o *check.Object
}

// mention is a reference found in the program of an analysis.
type mention struct {
	a *analysis
	reference
}

// programs lists the programs the uses of o are searched in: a's own, and, when o reaches
// importers, the program rooted at each module of the workspace that imports o's module (spec/
// lsp.md section 4), each with the object it finds declared where o is. A declaration of a host's
// declaration file reaches every program it is read for (hosted).
func (s *Server) programs(a *analysis, o *check.Object) []program {
	out := []program{{a, o}}
	if !a.reachesImporters(o) {
		return out
	}
	declPath, from, _, ok := a.declRange(o)
	if !ok {
		return out
	}
	for _, other := range append(s.importers(declPath, a.root.uri), s.hosted(declPath, a.root.uri)...) {
		oa := s.analyze(other)
		if same := oa.objectDeclaredAt(declPath, from); same != nil {
			out = append(out, program{oa, same})
		}
	}
	return out
}

// mentionsAt finds the mentions of what the target names, each once: of an alias in its module
// when the target is on one, else of its object in every program it reaches.
func (s *Server) mentionsAt(a *analysis, t *target) []mention {
	var out []mention
	if alias := a.aliasAt(t); alias != nil {
		for _, r := range a.aliasRefs(t.path, alias) {
			out = append(out, mention{a, r})
		}
		return out
	}
	seen := map[reference]bool{}
	for _, pr := range s.programs(a, t.obj) {
		for _, r := range pr.a.referencesOf(pr.o) {
			key := reference{path: r.path, from: r.from}
			if !seen[key] {
				seen[key] = true
				out = append(out, mention{pr.a, r})
			}
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
	for _, r := range a.refsAt(t) {
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
	case o.Path == prelude.Path || o.Decl == nil && o.Kind != check.Module:
		return &rpcError{Code: codeInvalidParams, Message: "a name of the prelude cannot be renamed"}
	case o.Kind == check.Field:
		return &rpcError{Code: codeInvalidParams, Message: "a record field is not renamed"}
	case o.Kind == check.Namespace:
		return &rpcError{Code: codeInvalidParams, Message: "a namespace cannot be renamed"}
	case isLambda(o.Decl):
		return &rpcError{Code: codeInvalidParams, Message: "a lambda's x cannot be renamed"}
	}
	return nil
}

func isLambda(n ast.Node) bool {
	_, ok := n.(*ast.Lambda)
	return ok
}

// validName says whether name can be given to a declaration: an identifier that is neither a
// keyword, a reserved word, nor a prelude name (section 10).
func validName(name string) *rpcError {
	if name == "" || name == "_" {
		return &rpcError{Code: codeInvalidParams, Message: "the new name must be an identifier"}
	}
	for i := 0; i < len(name); i++ {
		if !token.IsIdentByte(name[i]) || i == 0 && name[0] >= '0' && name[0] <= '9' {
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
		if n.Alias != n.Name && t.from == n.AliasPos {
			return n
		}
	case *ast.TypeName:
		// a type name whose first part is the alias
		m := a.modules[t.path]
		objs := a.info.TypeNames[n]
		if m == nil || m.Resolve == nil || len(objs) == 0 || len(n.PathPos) == 0 || t.from != n.PathPos[0] {
			return nil
		}
		if sc := m.Resolve.Scopes[m.File]; sc != nil {
			for _, b := range sc.Names {
				if in, ok := b.Node.(*ast.ImportName); ok && b.Name == n.Path[0] && in.Alias != in.Name && a.info.Decls[in] == objs[0] {
					return in
				}
			}
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
	edits, err := s.renameAt(a, a.at(a.module().File.Path, d.posOf(d.offset(p.Position, s.utf16))), p.NewName)
	if err != nil {
		return nil, err
	}
	edit := WorkspaceEdit{Changes: map[string][]TextEdit{}}
	for _, e := range edits {
		rd := e.a.doc(e.path)
		edit.Changes[rd.uri] = append(edit.Changes[rd.uri], TextEdit{Range: s.rng(rd, e.from, e.to), NewText: e.text})
	}
	return edit, nil
}

// renameEdit is one name a rename changes, with the analysis whose program it was found in.
type renameEdit struct {
	mention
	text string
}

// renameAt finds the edits that rename what the target names to name, each once, or why it
// cannot be renamed. The name is valid (validName).
func (s *Server) renameAt(a *analysis, t *target, name string) ([]renameEdit, *rpcError) {
	if t == nil || t.obj == nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: "nothing to rename here"}
	}
	if err := renamable(t.obj); err != nil {
		return nil, err
	}
	var edits []renameEdit
	seen := map[reference]bool{}
	add := func(m mention, field string) {
		key := reference{path: m.path, from: m.from}
		if !seen[key] && m.a.doc(m.path) != nil {
			seen[key] = true
			edits = append(edits, renameEdit{m, m.newText(name, field)})
		}
	}
	if alias := a.aliasAt(t); alias != nil {
		// the alias and its uses in this module only, which must not clash with the names bound
		// around them
		refs := a.aliasRefs(t.path, alias)
		for _, r := range refs {
			if a.boundAround(t.path, r.from, name, func(b resolve.Binding) bool { return b.Node == ast.Node(alias) }) {
				return nil, &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%s is already declared where %s is used", name, alias.Alias)}
			}
		}
		for _, r := range refs {
			add(mention{a, r}, alias.Alias)
		}
		return edits, nil
	}
	// Every module of the workspace that imports the declaring module, when the declaration
	// reaches importers (spec/lsp.md section 4), is the root of a program of its own, in which
	// the object is found again by where it is declared. A name the importer binds must not
	// clash there.
	for _, pr := range s.programs(a, t.obj) {
		if err := pr.a.collides(pr.o, name); err != nil {
			return nil, err
		}
		for _, r := range pr.a.referencesOf(pr.o) {
			if !r.viaAlias {
				add(mention{pr.a, r}, t.obj.Name)
			}
		}
	}
	return edits, nil
}

// reachesImporters reports whether renaming o changes other modules: an export, or a
// parameter of an exported function, which named arguments elsewhere spell; since a test
// file may import any top-level declaration of a module (section 10, Tests), any top-level
// declaration and any parameter of a top-level function; and a field of a declared record type,
// which any module that gets a value of the type can read.
func (a *analysis) reachesImporters(o *check.Object) bool {
	switch o.Kind {
	case check.Module:
		return false
	case check.Param:
		return a.enclosingFunc(o) != nil
	case check.Field:
		return o.Decl != nil
	}
	return o.Exported || a.topLevel(o)
}

// topLevel reports whether o is declared at the top level of its module.
func (a *analysis) topLevel(o *check.Object) bool {
	f := a.file(a.pathOf(o))
	if f == nil || o.Decl == nil {
		return false
	}
	for _, d := range f.Decls {
		if ast.Node(d) == o.Decl {
			return true
		}
		if vd, ok := d.(*ast.VarDecl); ok {
			found := false
			ast.Inspect(vd.Target, func(n ast.Node) bool {
				found = found || n == o.Decl
				return !found
			})
			if found {
				return true
			}
		}
	}
	return false
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

// objectDeclaredAt finds the object whose name is declared at a position of a module in this
// analysis: a top-level declaration, a parameter, or a field of a record type.
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
		case *ast.InterfaceDecl:
			if x.NamePos == pos {
				found = a.info.Decls[x]
			}
		case *ast.InterfaceFunc:
			if x.NamePos == pos {
				found = a.info.Decls[x]
			}
		case *ast.NameTarget:
			if x.Pos == pos {
				found = a.info.Decls[x]
			}
		case *ast.VarDecl:
			if nt, ok := x.Target.(*ast.NameTarget); ok && x.Owner != "" && nt.Pos == pos {
				found = a.info.Decls[x]
				return false
			}
		case *ast.Param:
			if x.Pos == pos {
				found = a.info.Decls[x]
			}
		case *ast.FieldType:
			if x.Pos == pos {
				found = a.info.Decls[x]
			}
		}
		return true
	})
	return found
}

// importers lists the documents, open or in the workspace, whose imports lead to the module at
// path, directly or through other modules, other than the one at skip: each becomes the root of a
// program the search reaches into, since a module that gets a value of a type from another can
// call the functions on it. Which files import which is read from their parsed imports, so only
// those that lead to path are analyzed.
func (s *Server) importers(path, skip string) []*document {
	canonical := map[string]string{}
	key := func(p string) string {
		if k, ok := canonical[p]; ok {
			return k
		}
		k := filepath.Clean(p)
		if real, err := filepath.EvalSymlinks(k); err == nil {
			k = real
		}
		canonical[p] = k
		return k
	}
	// the files that import each module, by the module's key
	importedBy := map[string][]*workspaceFile{}
	for _, f := range s.workspaceFiles() {
		for _, imp := range f.file.Imports {
			target := filepath.FromSlash(imp.Path)
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(f.doc.path), target)
			}
			k := key(target)
			importedBy[k] = append(importedBy[k], f)
		}
	}
	var out []*document
	reached := map[string]bool{key(path): true}
	queue := []string{key(path)}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, f := range importedBy[p] {
			k := key(f.doc.path)
			if reached[k] {
				continue
			}
			reached[k] = true
			queue = append(queue, k)
			if f.doc.uri != skip {
				out = append(out, f.doc)
			}
		}
	}
	return out
}

// hosted lists the documents of the workspace that a host's declaration file at path is read
// for (section 10, Declaration files), other than the one at skip: the modules in its directory,
// each the root of a program whose prelude its declarations join with no import. It lists none
// for a path that is not a declaration file.
func (s *Server) hosted(path, skip string) []*document {
	if parser.ModeOf(path) != parser.Declarations || filepath.Base(path) == prelude.Path {
		return nil
	}
	var out []*document
	for _, f := range s.workspaceFiles() {
		p := f.doc.path
		if f.doc.uri != skip && parser.ModeOf(p) != parser.Declarations && sameFile(filepath.Dir(p), filepath.Dir(path)) {
			out = append(out, f.doc)
		}
	}
	return out
}

// sameBinding reports whether b binds o under its own name: its declaration, an import that
// does not rename it, or the import that binds a module's namespace.
func (a *analysis) sameBinding(o *check.Object, b resolve.Binding) bool {
	return o.Decl != nil && b.Node == o.Decl || b.Name == o.Name && a.info.Decls[b.Node] == o
}

// collides reports a name already bound where the renamed binding is declared or used, which
// the rename would shadow or clash with (section 10): among the top-level names of its module,
// and in the scopes around its declaration and around each use, in every module of the program,
// an importer's included.
func (a *analysis) collides(o *check.Object, name string) *rpcError {
	declPath, from, _, ok := a.declRange(o)
	if !ok {
		return nil
	}
	if declPath == "" {
		declPath = a.pathOf(o)
	}
	clash := &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%s is already declared where %s is used", name, o.Name)}
	if mi, ok := a.info.Modules[declPath]; ok {
		for _, other := range mi.Names {
			if other.Name == name {
				return clash
			}
		}
	}
	self := func(b resolve.Binding) bool { return a.sameBinding(o, b) }
	if a.boundAround(declPath, from, name, self) {
		return clash
	}
	for _, r := range a.referencesOf(o) {
		if r.viaAlias || r.qualified || !a.boundAround(r.path, r.from, name, self) {
			continue
		}
		if r.path != declPath {
			return &rpcError{Code: codeInvalidParams, Message: fmt.Sprintf("%s is already declared in %s, which imports %s", name, filepath.Base(r.path), o.Name)}
		}
		return clash
	}
	return nil
}

// boundAround reports whether name is bound, other than by a binding self accepts, in a scope
// around pos in the module at path.
func (a *analysis) boundAround(path string, pos ast.Pos, name string, self func(resolve.Binding) bool) bool {
	m := a.modules[path]
	if m == nil || m.Resolve == nil {
		return false
	}
	for _, n := range ast.Path(m.File, pos) {
		sc, ok := m.Resolve.Scopes[n]
		if !ok {
			continue
		}
		for _, b := range sc.Names {
			if b.Name == name && !self(b) {
				return true
			}
		}
	}
	return false
}
