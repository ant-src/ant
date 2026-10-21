// Package resolve annotates a parsed module: every identifier gets a reference, every binding a
// slot, and every scope its size. It enforces the scope rules the parser leaves to it: no
// redeclaration in one scope, no declaration shadowing an import or a prelude name, and a
// function on a type belongs to a type of its own module.
package resolve

import (
	"fmt"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/prelude"
)

// Error is a scope error. It is a syntax error in the sense of section 10.
type Error struct {
	Pos ast.Pos
	Msg string
}

func (e *Error) Error() string { return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Col, e.Msg) }

// Options adjusts the rules for a session.
type Options struct {
	// Session allows a top-level name to be declared again, replacing the earlier binding.
	Session bool
	// Fields carries the module fields of earlier inputs in a session, so a later input can
	// refer to them and redeclare them. It is updated in place.
	Fields *ast.ModuleScope
	// Extra names the prelude binds in this program, such as host namespaces.
	Extra []string
}

type binding struct {
	name           string
	pos            ast.Pos
	scope          *scope
	captured       bool
	reassigned     bool
	assignedNested bool
	module         bool
	index          int // module field index, or slot index after finalize
	shared         bool
	imp            int    // RefImport: the import's index; -1 otherwise
	impName        string // RefImport: the export's name, "" for a namespace
}

type scope struct {
	parent   *scope
	names    map[string]*binding
	order    []*binding
	boundary bool // a function, lambda, go block, or defer block
	fn       bool // always allocates an environment
	module   bool
	block    *ast.Block
	loop     *ast.For
	fnScope  *ast.FuncScope
	xBinding *binding
	alloc    bool
}

type use struct {
	id   *ast.Ident
	b    *binding
	from *scope
}

type resolver struct {
	errs      []*Error
	uses      []use
	targetsOf []targetBinding
	paramsOf  []paramBinding
	scopes    []*scope
	module    *scope
	fields    *ast.ModuleScope
	session   bool
	extra     map[string]bool
}

func (r *resolver) isPrelude(name string) bool { return prelude.IsName(name) || r.extra[name] }

// Resolve annotates f and returns the scope errors it finds.
func Resolve(f *ast.File, opts Options) []*Error {
	r := &resolver{session: opts.Session, extra: map[string]bool{}}
	for _, n := range opts.Extra {
		r.extra[n] = true
	}
	r.fields = opts.Fields
	if r.fields == nil {
		r.fields = &ast.ModuleScope{Index: map[string]int{}}
	}
	f.Scope = r.fields
	r.module = r.push(&scope{module: true})
	if opts.Session {
		for name, idx := range r.fields.Index {
			r.module.names[name] = &binding{name: name, scope: r.module, module: true, index: idx, imp: -1}
		}
	}
	r.declareModule(f)
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.VarDecl:
			r.expr(d.Init, r.module)
		case *ast.FuncDecl:
			r.funcDecl(d)
		}
	}
	for _, s := range f.Stmts {
		if _, ok := s.(*ast.VarDecl); ok {
			continue // its initializer was resolved with the declarations
		}
		r.stmt(s, r.module)
	}
	r.finalize()
	return r.errs
}

func (r *resolver) errorf(pos ast.Pos, format string, args ...any) {
	r.errs = append(r.errs, &Error{Pos: pos, Msg: fmt.Sprintf(format, args...)})
}

func (r *resolver) push(s *scope) *scope {
	s.names = map[string]*binding{}
	r.scopes = append(r.scopes, s)
	return s
}

// declare binds name in s, enforcing the shadowing rules.
func (r *resolver) declare(s *scope, name string, pos ast.Pos) *binding {
	b := &binding{name: name, pos: pos, scope: s, imp: -1}
	if name == "_" {
		return b
	}
	if r.isPrelude(name) {
		r.errorf(pos, "%s is a prelude name and cannot be redeclared", name)
		return b
	}
	if old, ok := s.names[name]; ok {
		switch {
		case s.module && r.session && old.imp >= 0:
			// a session declaration replaces an earlier import
			s.names[name] = b
			s.order = append(s.order, b)
			b.module = true
			if idx, ok := r.fields.Index[name]; ok {
				b.index = idx
			} else {
				b.index = len(r.fields.Fields)
				r.fields.Fields = append(r.fields.Fields, name)
				r.fields.Index[name] = b.index
			}
			return b
		case s.module && r.session:
			// A session replaces the binding; the field keeps its index.
			b.module = true
			b.index = old.index
			s.names[name] = b
			return b
		case old.imp >= 0:
			r.errorf(pos, "%s is already bound by an import", name)
			return b
		default:
			r.errorf(pos, "%s is already declared in this scope", name)
			return b
		}
	}
	s.names[name] = b
	s.order = append(s.order, b)
	if s.module {
		b.module = true
		if idx, ok := r.fields.Index[name]; ok {
			b.index = idx
		} else {
			b.index = len(r.fields.Fields)
			r.fields.Fields = append(r.fields.Fields, name)
			r.fields.Index[name] = b.index
		}
	}
	return b
}

func (r *resolver) declareImport(s *scope, name string, pos ast.Pos, imp int, impName string) {
	if r.isPrelude(name) {
		r.errorf(pos, "%s is a prelude name and cannot be redeclared", name)
		return
	}
	if _, ok := s.names[name]; ok && !r.session {
		r.errorf(pos, "%s is already bound by an import", name)
		return
	}
	s.names[name] = &binding{name: name, pos: pos, scope: s, imp: imp, impName: impName}
}

func (r *resolver) declareModule(f *ast.File) {
	for i, imp := range f.Imports {
		if imp.Namespace != "" {
			r.declareImport(r.module, imp.Namespace, imp.Pos, i, "")
			continue
		}
		for _, n := range imp.Names {
			r.declareImport(r.module, n.Alias, n.Pos, i, n.Name)
		}
	}
	types := map[string]*ast.TypeDecl{}
	typeFuncs := map[string]map[string]bool{}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.VarDecl:
			r.target(d.Target, r.module)
		case *ast.FuncDecl:
			if d.Owner == "" {
				r.declare(r.module, d.Name, d.Pos)
			}
		case *ast.TypeDecl:
			r.declare(r.module, d.Name, d.Pos)
			types[d.Name] = d
		}
	}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Owner == "" {
			continue
		}
		td, ok := types[fd.Owner]
		if !ok {
			r.errorf(fd.Pos, "func %s.%s: %s is not a type declared in this module", fd.Owner, fd.Name, fd.Owner)
			continue
		}
		if typeFuncs[fd.Owner] == nil {
			typeFuncs[fd.Owner] = map[string]bool{}
		}
		if typeFuncs[fd.Owner][fd.Name] {
			r.errorf(fd.Pos, "%s.%s is already declared", fd.Owner, fd.Name)
		}
		typeFuncs[fd.Owner][fd.Name] = true
		if rt, ok := td.Body.(*ast.RecordType); ok {
			for _, fld := range rt.Fields {
				if fld.Name == fd.Name {
					r.errorf(fd.Pos, "%s.%s shares its name with a field of %s", fd.Owner, fd.Name, fd.Owner)
				}
			}
		}
	}
}

// ---- declarations

func (r *resolver) funcDecl(d *ast.FuncDecl) {
	s := r.push(&scope{parent: r.module, boundary: true, fn: true})
	d.Scope = &ast.FuncScope{}
	s.fnScope = d.Scope
	r.params(d.Params, s)
	if d.Body != nil {
		r.stmts(d.Body.Stmts, s)
	}
}

func (r *resolver) params(ps []ast.Param, s *scope) {
	for i := range ps {
		b := r.declare(s, ps[i].Name, ps[i].Pos)
		r.paramsOf = append(r.paramsOf, paramBinding{&ps[i], b})
	}
	for _, p := range ps {
		if p.Default != nil {
			r.expr(p.Default, s)
		}
	}
}

// target declares the names a var or for binds.
func (r *resolver) target(t ast.Target, s *scope) {
	switch t := t.(type) {
	case *ast.NameTarget:
		b := r.declare(s, t.Name, t.Pos)
		t.Slot = ast.Slot{} // filled in finalize
		r.bindTarget(t, b)
	case *ast.RecordTarget:
		for _, f := range t.Fields {
			r.target(f.Target, s)
		}
	case *ast.ArrayTarget:
		for _, e := range t.Elems {
			r.target(e, s)
		}
		if t.Rest != nil {
			r.target(t.Rest, s)
		}
	}
}

// targets records which NameTarget each binding fills, so finalize can write its slot.
func (r *resolver) bindTarget(t *ast.NameTarget, b *binding) {
	r.targetsOf = append(r.targetsOf, targetBinding{t, b})
}

type targetBinding struct {
	t *ast.NameTarget
	b *binding
}

type paramBinding struct {
	p *ast.Param
	b *binding
}

// ---- statements and expressions

func (r *resolver) stmts(ss []ast.Stmt, s *scope) {
	for _, st := range ss {
		r.stmt(st, s)
	}
}

func (r *resolver) stmt(st ast.Stmt, s *scope) {
	switch st := st.(type) {
	case *ast.VarDecl:
		if s.module {
			// Already declared with the module; a session statement list reaches here too.
			r.expr(st.Init, s)
			return
		}
		r.target(st.Target, s)
		r.expr(st.Init, s)
	case ast.Expr:
		r.expr(st, s)
	}
}

// block resolves a block's statements in a new scope, unless it declares nothing.
func (r *resolver) block(b *ast.Block, parent *scope) {
	s := r.push(&scope{parent: parent, block: b})
	r.stmts(b.Stmts, s)
}

func (r *resolver) body(kind string, b *ast.Block, parent *scope, fs *ast.FuncScope) *scope {
	s := r.push(&scope{parent: parent, boundary: true, fn: true, fnScope: fs})
	if kind == "lambda" {
		s.xBinding = r.declare(s, "x", b.Pos)
	}
	r.stmts(b.Stmts, s)
	return s
}

func (r *resolver) expr(e ast.Expr, s *scope) {
	switch e := e.(type) {
	case nil:
	case *ast.Ident:
		r.use(e, s)
	case *ast.IntLit, *ast.FloatLit, *ast.StringLit, *ast.BoolLit, *ast.NullLit, *ast.VoidLit,
		*ast.Break, *ast.Continue, *ast.Placeholder:
	case *ast.Template:
		for _, p := range e.Parts {
			if p.Expr != nil {
				r.expr(p.Expr, s)
			}
		}
	case *ast.RecordLit:
		r.expr(e.Head, s)
		for _, en := range e.Entries {
			if en.Spread != nil {
				r.expr(en.Spread, s)
			} else {
				r.expr(en.Value, s)
			}
		}
	case *ast.MapLit:
		for _, en := range e.Entries {
			r.expr(en.Key, s)
			r.expr(en.Value, s)
		}
	case *ast.ArrayLit:
		for _, el := range e.Elems {
			r.expr(el.Value, s)
		}
	case *ast.Lambda:
		e.Scope = &ast.FuncScope{}
		r.body("lambda", e.Body, s, e.Scope)
	case *ast.FuncLit:
		e.Scope = &ast.FuncScope{}
		fs := r.push(&scope{parent: s, boundary: true, fn: true, fnScope: e.Scope})
		r.params(e.Params, fs)
		r.stmts(e.Body.Stmts, fs)
	case *ast.Field:
		r.expr(e.X, s)
	case *ast.Index:
		r.expr(e.X, s)
		r.expr(e.Index, s)
	case *ast.Call:
		if e.Pipe {
			r.expr(e.Args[0].Value, s)
			r.expr(e.Fn, s)
			for _, a := range e.Args[1:] {
				r.expr(a.Value, s)
			}
			return
		}
		r.expr(e.Fn, s)
		for _, a := range e.Args {
			r.expr(a.Value, s)
		}
	case *ast.Unary:
		r.expr(e.X, s)
	case *ast.Binary:
		r.expr(e.X, s)
		r.expr(e.Y, s)
	case *ast.Is:
		r.expr(e.X, s)
		if e.Path != nil {
			r.expr(e.Path, s)
		}
	case *ast.Assign:
		r.assignTarget(e.Target, s)
		r.expr(e.Value, s)
	case *ast.If:
		r.expr(e.Cond, s)
		r.block(e.Then, s)
		switch el := e.Else.(type) {
		case *ast.If:
			r.expr(el, s)
		case *ast.Block:
			r.block(el, s)
		}
	case *ast.While:
		r.expr(e.Cond, s)
		r.block(e.Body, s)
	case *ast.For:
		r.expr(e.Iter, s)
		it := r.push(&scope{parent: s, loop: e})
		r.target(e.Target, it)
		r.stmts(e.Body.Stmts, it)
	case *ast.Go:
		e.Scope = &ast.FuncScope{}
		r.body("go", e.Body, s, e.Scope)
	case *ast.Defer:
		e.Scope = &ast.FuncScope{}
		r.body("defer", e.Body, s, e.Scope)
	case *ast.Return:
		r.expr(e.Value, s)
	case *ast.Block:
		r.block(e, s)
	default:
		panic(fmt.Sprintf("resolve: unexpected %T", e))
	}
}

func (r *resolver) assignTarget(t ast.Node, s *scope) {
	switch t := t.(type) {
	case *ast.Ident:
		r.use(t, s)
		r.markAssigned(t, s)
	case *ast.Field:
		r.expr(t.X, s)
	case *ast.Index:
		r.expr(t.X, s)
		r.expr(t.Index, s)
	case *ast.AssignArray:
		for _, e := range t.Elems {
			r.assignTarget(e, s)
		}
		if t.Rest != nil {
			r.assignTarget(t.Rest, s)
		}
	}
}

// markAssigned records that a local is assigned, and whether from a nested body.
func (r *resolver) markAssigned(id *ast.Ident, s *scope) {
	crossed := false
	for sc := s; sc != nil; sc = sc.parent {
		if b, ok := sc.names[id.Name]; ok {
			if b.module || b.imp >= 0 {
				return
			}
			b.reassigned = true
			if crossed {
				b.assignedNested = true
			}
			return
		}
		if sc.boundary {
			crossed = true
		}
	}
}

// use resolves an identifier read or write.
func (r *resolver) use(id *ast.Ident, s *scope) {
	if id.Name == "_" {
		id.Ref = ast.Ref{Kind: ast.RefDiscard, Name: "_"}
		return
	}
	crossed := false
	for sc := s; sc != nil; sc = sc.parent {
		if b, ok := sc.names[id.Name]; ok {
			switch {
			case b.imp >= 0:
				id.Ref = ast.Ref{Kind: ast.RefImport, Index: b.imp, Name: b.impName}
			case b.module:
				id.Ref = ast.Ref{Kind: ast.RefModule, Index: b.index, Name: id.Name}
			default:
				if crossed {
					b.captured = true
				}
				r.uses = append(r.uses, use{id: id, b: b, from: s})
			}
			return
		}
		if sc.boundary {
			crossed = true
		}
	}
	if r.isPrelude(id.Name) {
		id.Ref = ast.Ref{Kind: ast.RefPrelude, Name: id.Name}
		return
	}
	id.Ref = ast.Ref{Kind: ast.RefUnbound, Name: id.Name}
}

// ---- finalize

func (r *resolver) finalize() {
	// Slot numbering and environment allocation per scope.
	for _, s := range r.scopes {
		if s.module {
			continue
		}
		plain, shared := 0, 0
		for _, b := range s.order {
			if b.captured {
				b.shared = true
				b.index = shared
				shared++
			} else {
				b.index = plain
				plain++
			}
		}
		s.alloc = s.fn || len(s.order) > 0
		bs := ast.BlockScope{Plain: plain, Shared: shared}
		switch {
		case s.fnScope != nil:
			s.fnScope.BlockScope = bs
			if s.xBinding != nil {
				s.fnScope.X = ast.Slot{Index: s.xBinding.index, Shared: s.xBinding.shared, Reassigned: s.xBinding.reassigned, AssignedNested: s.xBinding.assignedNested}
			}
		case s.loop != nil:
			s.loop.Scope = &ast.BlockScope{Plain: plain, Shared: shared}
			if !s.alloc {
				s.loop.Scope = nil
			}
		case s.block != nil:
			if s.alloc {
				s.block.Scope = &bs
			}
		}
	}
	for _, pb := range r.paramsOf {
		pb.p.Slot = ast.Slot{Index: pb.b.index, Shared: pb.b.shared, Reassigned: pb.b.reassigned, AssignedNested: pb.b.assignedNested}
	}
	for _, tb := range r.targetsOf {
		if tb.t.Name == "_" {
			continue
		}
		if tb.b.module {
			tb.t.Slot = ast.Slot{Index: tb.b.index, Module: true}
		} else {
			tb.t.Slot = ast.Slot{Index: tb.b.index, Shared: tb.b.shared, Reassigned: tb.b.reassigned, AssignedNested: tb.b.assignedNested}
		}
	}
	for _, u := range r.uses {
		depth := 0
		for sc := u.from; sc != u.b.scope; sc = sc.parent {
			if sc.alloc {
				depth++
			}
		}
		u.id.Ref = ast.Ref{Kind: ast.RefLocal, Depth: depth, Index: u.b.index, Shared: u.b.shared, Name: u.id.Name}
	}
}
