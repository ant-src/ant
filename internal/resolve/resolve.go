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
	// Info, when set, is filled with what the resolver learned about the module's names, for an
	// editor.
	Info *Info
}

// Info is what the resolver learned about a module's names: where each identifier is declared
// and the names each scope holds.
type Info struct {
	// Defs maps each identifier to the node that declares it: a *ast.NameTarget, a *ast.Param,
	// a *ast.FuncDecl, a *ast.TypeDecl, a *ast.Import that binds a namespace, a *ast.ImportName,
	// or the *ast.Lambda whose x it is. A prelude name or an unbound one has no entry.
	Defs map[*ast.Ident]ast.Node
	// Uses maps each declaring node to the identifiers that refer to it, in source order.
	Uses map[ast.Node][]*ast.Ident
	// Scopes maps the node that opens a scope to it: the *ast.File for the module's, a
	// function's or lambda's body *ast.Block for its parameters and locals, a *ast.Block, the
	// body of a for loop for its variable, and the body of a select clause for its binding.
	Scopes map[ast.Node]*Scope
}

// Scope is a lexical scope: the names declared in it, in order, and the scope around it.
type Scope struct {
	Parent *Scope
	Names  []Binding
}

// Binding is a name a scope declares and the node that declares it.
type Binding struct {
	Name string
	Pos  ast.Pos
	Node ast.Node
	// Type is set for a type declared in a block, Import for an import's binding.
	Type, Import bool
}

func (in *Info) def(id *ast.Ident, n ast.Node) {
	if in == nil || n == nil {
		return
	}
	in.Defs[id] = n
	in.Uses[n] = append(in.Uses[n], id)
}

// NewInfo returns an empty Info.
func NewInfo() *Info {
	return &Info{Defs: map[*ast.Ident]ast.Node{}, Uses: map[ast.Node][]*ast.Ident{}, Scopes: map[ast.Node]*Scope{}}
}

type binding struct {
	name           string
	pos            ast.Pos
	scope          *scope
	captured       bool
	reassigned     bool
	assignedNested bool
	discard        bool // _, which binds nothing
	isType         bool // a type declared in a block
	module         bool
	index          int // module field index, or slot index after finalize
	shared         bool
	imp            int      // RefImport: the import's index; -1 otherwise
	impName        string   // RefImport: the export's name, "" for a namespace
	node           ast.Node // the declaring node, for Info
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
	clause   *ast.SelectClause
	fnScope  *ast.FuncScope
	xBinding *binding
	alloc    bool
	node     ast.Node // the node that opens the scope, for Info
	info     *Scope
}

// loopFacts collects the locals declared outside a loop that its body assigns.
type loopFacts struct {
	outer    *scope
	assigned *[]string
	seen     map[*binding]bool
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
	typesOf   []typeBinding
	tparams   []ast.TypeParam // the type parameters of the function being resolved
	paramsOf  []paramBinding
	scopes    []*scope
	loops     []*loopFacts // the loops whose bodies are being resolved, innermost last
	module    *scope
	fields    *ast.ModuleScope
	session   bool
	// script is set for a script, whose top-level vars are locals of its body for narrowing, so
	// their assignments are tracked as a local's are (section 7).
	script bool
	// defaultOf is the block scope of the local type whose field defaults are being resolved,
	// or nil: a default there may name the block's types but none of its variables.
	defaultOf *scope
	extra     map[string]bool
	info      *Info
}

func (r *resolver) isPrelude(name string) bool { return prelude.IsName(name) || r.extra[name] }

// Resolve annotates f and returns the scope errors it finds.
func Resolve(f *ast.File, opts Options) []*Error {
	r := &resolver{session: opts.Session, script: f.Script, extra: map[string]bool{}, info: opts.Info}
	for _, n := range opts.Extra {
		r.extra[n] = true
	}
	r.fields = opts.Fields
	if r.fields == nil {
		r.fields = &ast.ModuleScope{Index: map[string]int{}}
	}
	f.Scope = r.fields
	r.module = r.push(&scope{module: true, node: f})
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
		case *ast.TypeDecl:
			r.defaults(d, r.module)
		}
	}
	for _, s := range f.Stmts {
		if _, ok := s.(*ast.VarDecl); ok {
			continue // its initializer was resolved with the declarations
		}
		r.stmt(s, r.module)
	}
	r.finalize()
	if r.session && len(r.errs) == 0 {
		r.keepTypes(f)
	}
	return r.errs
}

// keepTypes records in a session's fields which names are now types, so a later input can
// declare functions on them: a name this input declares is a type only if declared as one.
func (r *resolver) keepTypes(f *ast.File) {
	if r.fields.Types == nil {
		r.fields.Types = map[string]*ast.TypeDecl{}
	}
	for name, b := range r.module.names {
		if b.pos != (ast.Pos{}) {
			delete(r.fields.Types, name)
		}
	}
	for _, d := range f.Decls {
		if td, ok := d.(*ast.TypeDecl); ok {
			r.fields.Types[td.Name] = td
		}
	}
}

func (r *resolver) errorf(pos ast.Pos, format string, args ...any) {
	r.errs = append(r.errs, &Error{Pos: pos, Msg: fmt.Sprintf(format, args...)})
}

func (r *resolver) push(s *scope) *scope {
	s.names = map[string]*binding{}
	r.scopes = append(r.scopes, s)
	if r.info != nil && s.node != nil {
		s.info = &Scope{}
		if s.parent != nil {
			s.info.Parent = s.parent.info
		}
		r.info.Scopes[s.node] = s.info
	}
	return s
}

// noteBinding records a binding in its scope's Info.
func (r *resolver) noteBinding(s *scope, b *binding) {
	if s.info == nil || b.node == nil {
		return
	}
	s.info.Names = append(s.info.Names, Binding{Name: b.name, Pos: b.pos, Node: b.node, Type: b.isType, Import: b.imp >= 0})
}

// declare binds name in s, enforcing the shadowing rules. node is what declares it.
func (r *resolver) declare(s *scope, name string, pos ast.Pos, node ast.Node) *binding {
	b := &binding{name: name, pos: pos, scope: s, imp: -1, node: node}
	if _, isType := node.(*ast.TypeDecl); isType {
		b.isType = true
	}
	if name == "_" {
		b.discard = true
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
			r.noteBinding(s, b)
			return b
		case s.module && r.session:
			// A session replaces the binding; the field keeps its index.
			b.module = true
			b.index = old.index
			s.names[name] = b
			r.noteBinding(s, b)
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
	r.noteBinding(s, b)
	return b
}

func (r *resolver) declareImport(s *scope, name string, pos ast.Pos, imp int, impName string, node ast.Node) {
	if r.isPrelude(name) {
		r.errorf(pos, "%s is a prelude name and cannot be redeclared", name)
		return
	}
	if _, ok := s.names[name]; ok && !r.session {
		r.errorf(pos, "%s is already bound by an import", name)
		return
	}
	b := &binding{name: name, pos: pos, scope: s, imp: imp, impName: impName, node: node}
	s.names[name] = b
	r.noteBinding(s, b)
}

func (r *resolver) declareModule(f *ast.File) {
	for i, imp := range f.Imports {
		if imp.Namespace != "" {
			r.declareImport(r.module, imp.Namespace, imp.Pos, i, "", imp)
			continue
		}
		for j := range imp.Names {
			n := &imp.Names[j]
			r.declareImport(r.module, n.Alias, n.Pos, i, n.Name, n)
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
				r.declare(r.module, d.Name, d.Pos, d)
			}
			r.typeParams(d.TypeParams)
		case *ast.TypeDecl:
			r.declare(r.module, d.Name, d.Pos, d)
			types[d.Name] = d
			r.typeParams(d.TypeParams)
		}
	}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Owner == "" {
			continue
		}
		td, ok := types[fd.Owner]
		if b := r.module.names[fd.Owner]; !ok && r.session && b != nil && b.pos == (ast.Pos{}) {
			td, ok = r.fields.Types[fd.Owner] // a type of an earlier input, not redeclared here
		}
		if !ok {
			r.errorf(fd.Pos, "func %s.%s: %s is not a type declared at the top level of this module", fd.Owner, fd.Name, fd.Owner)
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

// typeParams enforces on type parameters the rule that a prelude name is never shadowed, so
// that int means one thing in every scope (section 10), and that a list names each once.
func (r *resolver) typeParams(tps []ast.TypeParam) {
	seen := map[string]bool{}
	for _, tp := range tps {
		switch {
		case r.isPrelude(tp.Name):
			r.errorf(tp.Pos, "%s is a prelude name and cannot be redeclared", tp.Name)
		case seen[tp.Name]:
			r.errorf(tp.Pos, "%s is already declared in this scope", tp.Name)
		}
		seen[tp.Name] = true
	}
}

// ---- declarations

func (r *resolver) funcDecl(d *ast.FuncDecl) {
	s := r.push(&scope{parent: r.module, boundary: true, fn: true, node: d.Body})
	r.tparams = d.TypeParams
	defer func() { r.tparams = nil }()
	d.Scope = &ast.FuncScope{}
	s.fnScope = d.Scope
	r.params(d.Params, s)
	if d.Body != nil {
		r.stmts(d.Body, s)
	}
}

func (r *resolver) params(ps []ast.Param, s *scope) {
	for i := range ps {
		b := r.declare(s, ps[i].Name, ps[i].Pos, &ps[i])
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
		b := r.declare(s, t.Name, t.Pos, t)
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

type typeBinding struct {
	d *ast.TypeDecl
	b *binding
}

type paramBinding struct {
	p *ast.Param
	b *binding
}

// defaults resolves the field defaults of a record type declared in s. They are evaluated at
// literals anywhere in the run but in the scope of the declaration (section 4, Records), and a
// local type has one tag for the whole run, so its defaults may not close over the variables
// of one run of its block.
func (r *resolver) defaults(d *ast.TypeDecl, s *scope) {
	rt, ok := d.Body.(*ast.RecordType)
	if !ok {
		return
	}
	savedOf, savedLoops := r.defaultOf, r.loops
	if !s.module {
		r.defaultOf = s
	}
	r.loops = nil
	for _, f := range rt.Fields {
		if f.Default != nil {
			r.expr(f.Default, s)
		}
	}
	r.defaultOf, r.loops = savedOf, savedLoops
}

// ---- statements and expressions

// stmts resolves a block's statements in s. Its types are declared first, so they are visible
// throughout the block (section 4, Type declarations).
func (r *resolver) stmts(b *ast.Block, s *scope) {
	for _, td := range b.Types {
		for _, tp := range r.tparams {
			if tp.Name == td.Name {
				r.errorf(td.Pos, "%s is a type parameter of the enclosing function and cannot be redeclared", td.Name)
			}
		}
		b := r.declare(s, td.Name, td.Pos, td)
		r.typesOf = append(r.typesOf, typeBinding{td, b})
		r.typeParams(td.TypeParams)
	}
	for _, st := range b.Stmts {
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
	case *ast.TypeDecl:
		// A local type's defaults are resolved where it is declared, so a variable declared
		// above it in its block is in scope and refused, not taken for a top-level name.
		r.defaults(st, s)
	case ast.Expr:
		r.expr(st, s)
	}
}

// block resolves a block's statements in a new scope, unless it declares nothing.
func (r *resolver) block(b *ast.Block, parent *scope) {
	s := r.push(&scope{parent: parent, block: b, node: b})
	r.stmts(b, s)
}

// body resolves the body of a lambda, whose node binds x, or of a go or defer block.
func (r *resolver) body(lambda *ast.Lambda, b *ast.Block, parent *scope, fs *ast.FuncScope) {
	s := r.push(&scope{parent: parent, boundary: true, fn: true, fnScope: fs, node: b})
	if lambda != nil {
		s.xBinding = r.declare(s, "x", b.Pos, lambda)
	}
	r.stmts(b, s)
}

func (r *resolver) expr(e ast.Expr, s *scope) {
	switch e := e.(type) {
	case nil:
	case *ast.Ident:
		r.use(e, s)
		if e.Ref.Kind == ast.RefPrelude && e.Name == "assert" {
			r.errorf(e.Pos, "assert is not a value; it can only be called, as assert(cond) or assert(cond, message)")
		}
	case *ast.IntLit, *ast.FloatLit, *ast.StringLit, *ast.RegexLit, *ast.BoolLit, *ast.NullLit, *ast.VoidLit,
		*ast.Break, *ast.Continue, *ast.Placeholder, *ast.BadExpr:
	case *ast.Template:
		for _, p := range e.Parts {
			if p.Expr != nil {
				r.expr(p.Expr, s)
			}
		}
	case *ast.Html:
		r.markup(e.Content, s)
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
		r.body(e, e.Body, s, e.Scope)
	case *ast.FuncLit:
		e.Scope = &ast.FuncScope{}
		fs := r.push(&scope{parent: s, boundary: true, fn: true, fnScope: e.Scope, node: e.Body})
		r.params(e.Params, fs)
		r.stmts(e.Body, fs)
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
		if id, ok := e.Fn.(*ast.Ident); ok && e.Assert != nil && !e.Optional {
			r.use(id, s) // the one place the name assert may stand (section 8, assert)
		} else {
			r.expr(e.Fn, s)
		}
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
		r.loopBody(s, &e.Assigned, func() { r.block(e.Body, s) })
	case *ast.For:
		r.expr(e.Iter, s)
		it := r.push(&scope{parent: s, loop: e, node: e.Body})
		r.target(e.Target, it)
		r.loopBody(s, &e.Assigned, func() { r.stmts(e.Body, it) })
	case *ast.Select:
		// The heads' operands belong to the enclosing scope; each body is a scope of its own
		// with the clause's binding, as a loop's iteration is.
		for _, c := range e.Clauses {
			if c.Head != nil {
				r.expr(c.Head, s)
			}
			cs := r.push(&scope{parent: s, clause: c, node: c.Body})
			if c.Target != nil {
				r.target(c.Target, cs)
			}
			r.stmts(c.Body, cs)
		}
	case *ast.Go:
		e.Scope = &ast.FuncScope{}
		r.body(nil, e.Body, s, e.Scope)
	case *ast.Defer:
		e.Scope = &ast.FuncScope{}
		r.body(nil, e.Body, s, e.Scope)
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
	case *ast.BadExpr:
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

// loopBody resolves the body of a loop in the scope outer, recording in assigned the locals
// declared outside the loop that the body assigns, at any depth: the checker drops their
// narrowing for the whole loop, since an assignment may run before the next iteration (section 7).
func (r *resolver) loopBody(outer *scope, assigned *[]string, resolve func()) {
	*assigned = nil
	r.loops = append(r.loops, &loopFacts{outer: outer, assigned: assigned, seen: map[*binding]bool{}})
	resolve()
	r.loops = r.loops[:len(r.loops)-1]
}

// markAssigned records that a local is assigned, whether from a nested body, and in which
// enclosing loops.
func (r *resolver) markAssigned(id *ast.Ident, s *scope) {
	crossed := false
	for sc := s; sc != nil; sc = sc.parent {
		if b, ok := sc.names[id.Name]; ok {
			if b.imp >= 0 || b.module && !r.script {
				return
			}
			b.reassigned = true
			if crossed {
				b.assignedNested = true
			}
			for _, l := range r.loops {
				if !l.seen[b] && encloses(b.scope, l.outer) {
					l.seen[b] = true
					*l.assigned = append(*l.assigned, b.name)
				}
			}
			return
		}
		if sc.boundary {
			crossed = true
		}
	}
}

// encloses reports whether outer is s or one of its parents.
func encloses(outer, s *scope) bool {
	for ; s != nil; s = s.parent {
		if s == outer {
			return true
		}
	}
	return false
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
			if r.defaultOf != nil && !b.module && b.imp < 0 && !b.isType && encloses(b.scope, r.defaultOf) {
				r.errorf(id.Pos, "a field default of a local type cannot use %s, a variable or parameter of an enclosing block", id.Name)
			}
			r.info.def(id, b.node)
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

// slotOf is where a parameter or target binding lives, once finalize has numbered the slots.
func slotOf(b *binding) ast.Slot {
	switch {
	case b.discard:
		return ast.Slot{Discard: true}
	case b.module:
		return ast.Slot{Index: b.index, Module: true, Reassigned: b.reassigned, AssignedNested: b.assignedNested}
	}
	return ast.Slot{Index: b.index, Shared: b.shared, Reassigned: b.reassigned, AssignedNested: b.assignedNested}
}

func (r *resolver) finalize() {
	// Slot numbering and environment allocation per scope.
	for _, s := range r.scopes {
		if s.module {
			continue
		}
		plain, shared := 0, 0
		var names []ast.SlotName
		for _, b := range s.order {
			if b.captured {
				b.shared = true
				b.index = shared
				shared++
			} else {
				b.index = plain
				plain++
			}
			names = append(names, ast.SlotName{Name: b.name, Index: b.index, Shared: b.shared, Type: b.isType})
		}
		s.alloc = s.fn || len(s.order) > 0
		bs := ast.BlockScope{Plain: plain, Shared: shared, Names: names}
		switch {
		case s.fnScope != nil:
			s.fnScope.BlockScope = bs
			if s.xBinding != nil {
				s.fnScope.X = ast.Slot{Index: s.xBinding.index, Shared: s.xBinding.shared, Reassigned: s.xBinding.reassigned, AssignedNested: s.xBinding.assignedNested}
			}
		case s.loop != nil:
			s.loop.Scope = &ast.BlockScope{Plain: plain, Shared: shared, Names: names}
			if !s.alloc {
				s.loop.Scope = nil
			}
		case s.clause != nil:
			if s.alloc {
				s.clause.Scope = &bs
			}
		case s.block != nil:
			if s.alloc {
				s.block.Scope = &bs
			}
		}
	}
	for _, pb := range r.paramsOf {
		pb.p.Slot = slotOf(pb.b)
	}
	for _, tb := range r.targetsOf {
		tb.t.Slot = slotOf(tb.b)
	}
	for _, tb := range r.typesOf {
		tb.d.Slot = slotOf(tb.b)
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

// markup resolves the expressions of an HTML expression's content, attributes first.
func (r *resolver) markup(content []ast.HtmlNode, s *scope) {
	for _, n := range content {
		switch {
		case n.Element != nil:
			for _, a := range n.Element.Attrs {
				if a.Expr != nil {
					r.expr(a.Expr, s)
				}
			}
			r.markup(n.Element.Content, s)
		case n.Expr != nil:
			r.expr(n.Expr, s)
		}
	}
}
