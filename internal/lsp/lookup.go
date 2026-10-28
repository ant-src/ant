package lsp

import (
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
)

// target is what the cursor is on: a name with the object it refers to, an import's path, or a
// markup element or attribute.
type target struct {
	obj      *check.Object
	node     ast.Node
	from, to ast.Pos // the name's span
	path     string  // the module the name is in
	// imp is set when the cursor is on an import's path, element and attr for markup names.
	imp     *ast.Import
	element *ast.HtmlElement
	attr    *ast.HtmlAttr
}

// at finds what is under pos in the module at path, or nil.
func (a *analysis) at(path string, pos ast.Pos) *target {
	f := a.file(path)
	if f == nil {
		return nil
	}
	in := a.info
	nodes := ast.Path(f, pos)
	for i := len(nodes) - 1; i >= 0; i-- {
		switch n := nodes[i].(type) {
		case *ast.Ident:
			return &target{obj: in.Objects[n], node: n, from: n.Pos, to: n.End, path: path}
		case *ast.Field:
			if !pos.Before(n.NamePos) || pos == n.Pos {
				return &target{obj: in.Members[n], node: n, from: n.NamePos, to: n.End, path: path}
			}
		case *ast.TypeName:
			for j, p := range n.PathPos {
				if within(pos, p, n.Path[j]) {
					t := &target{node: n, from: p, to: after(p, n.Path[j]), path: path}
					if objs := in.TypeNames[n]; j < len(objs) {
						t.obj = objs[j]
					}
					return t
				}
			}
		case *ast.NameTarget:
			return &target{obj: in.Decls[n], node: n, from: n.Pos, to: n.End, path: path}
		case *ast.Param:
			if within(pos, n.Pos, n.Name) {
				return &target{obj: in.Decls[n], node: n, from: n.Pos, to: after(n.Pos, n.Name), path: path}
			}
		case *ast.TypeParam:
			if within(pos, n.Pos, n.Name) {
				return &target{obj: in.Decls[n], node: n, from: n.Pos, to: after(n.Pos, n.Name), path: path}
			}
		case *ast.FuncDecl:
			if within(pos, n.NamePos, n.Name) {
				return &target{obj: in.Decls[n], node: n, from: n.NamePos, to: after(n.NamePos, n.Name), path: path}
			}
			if n.Owner != "" && within(pos, n.OwnerPos, n.Owner) {
				return &target{obj: a.named(path, n.Owner), node: n, from: n.OwnerPos, to: after(n.OwnerPos, n.Owner), path: path}
			}
		case *ast.TypeDecl:
			if within(pos, n.NamePos, n.Name) {
				return &target{obj: in.Decls[n], node: n, from: n.NamePos, to: after(n.NamePos, n.Name), path: path}
			}
		case *ast.FieldType:
			if within(pos, n.Pos, n.Name) {
				return &target{obj: in.Decls[n], node: n, from: n.Pos, to: after(n.Pos, n.Name), path: path}
			}
		case *ast.ImportName:
			if n.Alias != n.Name && within(pos, n.AliasPos, n.Alias) {
				return &target{obj: in.Decls[n], node: n, from: n.AliasPos, to: after(n.AliasPos, n.Alias), path: path}
			}
			return &target{obj: in.Decls[n], node: n, from: n.Pos, to: after(n.Pos, n.Name), path: path}
		case *ast.Import:
			if !pos.Before(n.PathPos) && pos.Before(n.PathEnd) {
				return &target{imp: n, node: n, from: n.PathPos, to: n.PathEnd, path: path}
			}
			if n.Namespace != "" && within(pos, n.NamespacePos, n.Namespace) {
				return &target{obj: in.Decls[n], node: n, from: n.NamespacePos, to: after(n.NamespacePos, n.Namespace), path: path}
			}
		case *ast.RecordEntry:
			if n.Name != "" && within(pos, n.Pos, n.Name) {
				return &target{obj: in.Entries[n], node: n, from: n.Pos, to: after(n.Pos, n.Name), path: path}
			}
		case *ast.HtmlElement:
			if within(pos, n.NamePos, n.Name) {
				return &target{element: n, node: n, from: n.NamePos, to: after(n.NamePos, n.Name), path: path}
			}
			if n.Close != (ast.Pos{}) {
				namePos := ast.Pos{Line: n.Close.Line, Col: n.Close.Col + 2}
				if within(pos, namePos, n.Name) {
					return &target{element: n, node: n, from: namePos, to: after(namePos, n.Name), path: path}
				}
			}
		case *ast.HtmlAttr:
			if !n.Spread && within(pos, n.Pos, n.Name) {
				return &target{attr: n, node: n, from: n.Pos, to: after(n.Pos, n.Name), path: path}
			}
		}
	}
	return nil
}

// within reports whether pos lies in the name written at p.
func within(pos, p ast.Pos, name string) bool {
	return pos.Line == p.Line && pos.Col >= p.Col && pos.Col <= p.Col+len(name)
}

// after is the position just after the name written at p.
func after(p ast.Pos, name string) ast.Pos { return ast.Pos{Line: p.Line, Col: p.Col + len(name)} }

// named finds a top-level object of the module at path by name.
func (a *analysis) named(path, name string) *check.Object {
	m, ok := a.info.Modules[path]
	if !ok {
		return nil
	}
	for _, o := range m.Names {
		if o.Name == name {
			return o
		}
	}
	return nil
}

// declRange is where an object is declared: the module path and the span of its name, or false
// for an object with no declaration.
func (a *analysis) declRange(o *check.Object) (string, ast.Pos, ast.Pos, bool) {
	if o == nil {
		return "", ast.Pos{}, ast.Pos{}, false
	}
	switch d := o.Decl.(type) {
	case *ast.FuncDecl:
		return o.Path, d.NamePos, after(d.NamePos, d.Name), true
	case *ast.TypeDecl:
		return o.Path, d.NamePos, after(d.NamePos, d.Name), true
	case *ast.NameTarget:
		return o.Path, d.Pos, d.End, true
	case *ast.Param:
		return o.Path, d.Pos, after(d.Pos, d.Name), true
	case *ast.TypeParam:
		return o.Path, d.Pos, after(d.Pos, d.Name), true
	case *ast.FieldType:
		return o.Path, d.Pos, after(d.Pos, d.Name), true
	case *ast.Lambda:
		return o.Path, d.Pos, ast.Pos{Line: d.Pos.Line, Col: d.Pos.Col + 1}, true
	case *ast.VarDecl:
		if nt, ok := d.Target.(*ast.NameTarget); ok {
			return o.Path, nt.Pos, nt.End, true
		}
		return o.Path, d.Pos, d.End, true
	case *ast.Import:
		return a.importer(d), d.NamespacePos, after(d.NamespacePos, d.Namespace), true
	}
	if o.Kind == check.Module && o.Path != "" {
		return o.Path, ast.Pos{Line: 1, Col: 1}, ast.Pos{Line: 1, Col: 1}, true
	}
	return "", ast.Pos{}, ast.Pos{}, false
}

// importer finds the module that holds an import.
func (a *analysis) importer(imp *ast.Import) string {
	for path, m := range a.modules {
		for _, i := range m.File.Imports {
			if i == imp {
				return path
			}
		}
	}
	return ""
}

// pathOf is the module path of a local's declaration, which the object does not carry: the
// module whose tree holds the node.
func (a *analysis) pathOf(o *check.Object) string {
	if o.Path != "" {
		return o.Path
	}
	for path, m := range a.modules {
		found := false
		ast.Inspect(m.File, func(n ast.Node) bool {
			if n == o.Decl {
				found = true
			}
			return !found
		})
		if found {
			return path
		}
	}
	return ""
}
