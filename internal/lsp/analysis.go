package lsp

import (
	"path/filepath"
	"sort"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/parser"
)

// preludeURI is the document the prelude's declarations are shown in (spec/lsp.md section 4).
const preludeURI = "ant:/" + prelude.Path

// analysis is a program rooted at one document (spec/lsp.md section 2): its modules, the
// checker's findings, and every diagnostic.
type analysis struct {
	root    *document
	prog    *loader.Program
	info    *check.Info
	diags   []loader.Diagnostic
	modules map[string]*loader.Module // by the path the module's file carries
	docs    map[string]*document      // the text of each module, open or read from disk
	server  *Server
	names   map[string]bool // the names docs may link to, made on first use
}

// analyze loads and checks the program rooted at doc, with every open document standing in for
// its file. Results are cached until a document changes.
func (s *Server) analyze(doc *document) *analysis {
	if s.cacheGen != s.gen {
		s.cache = map[string]*analysis{}
		s.cacheGen = s.gen
	}
	if a, ok := s.cache[doc.uri]; ok {
		return a
	}
	open := s.byPath()
	sources := map[string]string{}
	for path, d := range open {
		sources[path] = d.text
	}
	a := &analysis{root: doc, modules: map[string]*loader.Module{}, docs: map[string]*document{}, server: s}
	var prog *loader.Program
	if doc == s.prelude {
		// The prelude is a program of its own: a declaration file the checker has declared
		// already, parsed once for every program.
		f, _ := prelude.Decls()
		m := &loader.Module{Path: prelude.Path, File: f, Src: doc.text, Layout: prelude.Layout()}
		prog = &loader.Program{Modules: []*loader.Module{m}, Root: m}
		open[prelude.Path] = doc
	} else {
		prog, a.diags = loader.Load(doc.path, loader.Config{Mode: parser.Program, Recover: true, Sources: sources})
		if prog == nil {
			// The root could not be read, which cannot happen for an open document; analyze its
			// text alone.
			prog, _ = loader.Load(doc.path, loader.Config{Mode: parser.Program, Recover: true, Sources: map[string]string{doc.path: doc.text}})
		}
	}
	a.prog = prog
	for _, m := range prog.Modules {
		a.modules[m.File.Path] = m
		if d, ok := open[m.File.Path]; ok {
			a.docs[m.File.Path] = d
		} else {
			a.docs[m.File.Path] = newDocument(fileURI(m.File.Path), 0, m.Src)
		}
	}
	a.info = check.NewInfo()
	a.diags = append(a.diags, check.Program(prog, check.Config{Info: a.info})...)
	sort.SliceStable(a.diags, func(i, j int) bool {
		x, y := a.diags[i], a.diags[j]
		if x.Path != y.Path {
			return x.Path < y.Path
		}
		return x.Pos.Before(y.Pos)
	})
	s.cache[doc.uri] = a
	return a
}

// module is the root's module.
func (a *analysis) module() *loader.Module { return a.prog.Root }

// knows reports whether name, as a doc comment brackets it, is a declaration the program can
// see: a prelude name or Namespace.member, or a top-level name of the root module or a member
// of one, so that doc.md's links become code and other bracketed text stays as written.
func (a *analysis) knows(name string) bool {
	if a.names == nil {
		a.names = map[string]bool{}
		note := func(prefix string, objs []*check.Object) {
			for _, o := range objs {
				a.names[prefix+o.Name] = true
				if o.Kind == check.Namespace || o.Kind == check.TypeKind || o.Kind == check.Module {
					for _, m := range o.Members() {
						a.names[prefix+o.Name+"."+m.Name] = true
					}
				}
			}
		}
		note("", a.info.Prelude)
		if mi, ok := a.info.Modules[a.module().File.Path]; ok {
			note("", mi.Names)
		}
	}
	return a.names[name]
}

// doc returns the document of a module by path, the prelude's for prelude.Path, or nil.
func (a *analysis) doc(path string) *document {
	if path == prelude.Path {
		return a.server.preludeDoc()
	}
	if d, ok := a.docs[path]; ok {
		return d
	}
	// a path spelled differently from the module's own
	for p, d := range a.docs {
		if sameFile(p, path) {
			return d
		}
	}
	return nil
}

// file returns the parsed module at path, the prelude's for prelude.Path, or nil.
func (a *analysis) file(path string) *ast.File {
	if path == prelude.Path {
		f, _ := prelude.Decls()
		return f
	}
	if m, ok := a.modules[path]; ok {
		return m.File
	}
	return nil
}

func sameFile(a, b string) bool {
	if a == b {
		return true
	}
	x, errx := filepath.EvalSymlinks(a)
	y, erry := filepath.EvalSymlinks(b)
	return errx == nil && erry == nil && x == y
}

// preludeDoc is the prelude as a document, made once: the server's own, which it analyzes as a
// declaration file, whatever copy the client opened under its URI.
func (s *Server) preludeDoc() *document {
	if s.prelude == nil {
		s.prelude = newDocument(preludeURI, 0, prelude.Source())
		s.prelude.path = prelude.Path
	}
	return s.prelude
}

// rangeOf converts a node's span in the module at path to a protocol range.
func (a *analysis) rangeOf(path string, n ast.Node) (Range, bool) {
	d := a.doc(path)
	if d == nil {
		return Range{}, false
	}
	return a.server.rng(d, ast.Start(n), n.EndPosition()), true
}

// rng converts two parser positions in d to a protocol range.
func (s *Server) rng(d *document, from, to ast.Pos) Range {
	return Range{Start: d.position(d.offsetOf(from), s.utf16), End: d.position(d.offsetOf(to), s.utf16)}
}

// nameRange is the range of a name of n characters at p.
func (s *Server) nameRange(d *document, p ast.Pos, name string) Range {
	return s.rng(d, p, ast.Pos{Line: p.Line, Col: p.Col + len(name)})
}

// location is a node's location in the module at path.
func (a *analysis) location(path string, from, to ast.Pos) (Location, bool) {
	d := a.doc(path)
	if d == nil {
		return Location{}, false
	}
	return Location{URI: d.uri, Range: a.server.rng(d, from, to)}, true
}
