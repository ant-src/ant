package interp

import (
	"slices"
	"sync"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/value"
)

// module is a module's run-time state: its record and resolved imports.
type module struct {
	m       *loader.Module
	path    string // m's path, which a session's later inputs keep
	rec     *value.Record
	imports []importRef
	root    *Env
	tags    map[string]*value.Tag // the tag of each type declaration, for functions on types
	// grow is a session module's lock: an input that declares new names replaces rec and
	// imports under it while goroutines of earlier inputs use them under its read side. Nil
	// for a program's modules, which never change.
	grow *sync.RWMutex
}

// get reads the module's variable at i.
func (m *module) get(i int) value.Value {
	if m.grow == nil {
		return m.rec.Get(i)
	}
	m.grow.RLock()
	v := m.rec.Get(i)
	m.grow.RUnlock()
	return v
}

// set stores the module's variable at i.
func (m *module) set(i int, v value.Value) {
	if m.grow == nil {
		m.rec.Set(i, v)
		return
	}
	m.grow.RLock()
	m.rec.Set(i, v)
	m.grow.RUnlock()
}

// importAt returns the module's import at i.
func (m *module) importAt(i int) importRef {
	if m.grow == nil {
		return m.imports[i]
	}
	m.grow.RLock()
	ir := m.imports[i]
	m.grow.RUnlock()
	return ir
}

type importRef struct {
	rec       *value.Record
	namespace bool // the import binds the whole module to one name, rather than destructuring it
}

// setupModules creates every module record with its tags and functions, so that all names
// exist before any initializer runs (section 10).
func (in *Instance) setupModules() {
	in.modules = make([]*module, len(in.prog.Modules))
	for _, lm := range in.prog.Modules {
		in.modules[lm.Index] = newModule(lm)
	}
	for _, m := range in.modules {
		in.linkModule(m, func(i int) *module { return in.modules[i] })
	}
}

func newModule(lm *loader.Module) *module {
	f := lm.File
	tag := value.NewTag(f.Path, value.RecordTag, f.Scope.Fields)
	m := &module{m: lm, path: f.Path, rec: value.NewRecord(tag, nil), tags: map[string]*value.Tag{}}
	m.root = &Env{mod: m}
	return m
}

// linkModule resolves a module's imports and binds its types and functions.
func (in *Instance) linkModule(m *module, lookup func(int) *module) {
	f := m.m.File
	m.imports = nil
	for _, imp := range f.Imports {
		target := lookup(imp.Module)
		if imp.Namespace != "" {
			m.imports = append(m.imports, importRef{rec: target.rec, namespace: true})
			continue
		}
		m.imports = append(m.imports, importRef{rec: target.rec})
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.TypeDecl:
			v := TypeValue(d)
			setDefaults(v, d, m.root)
			m.rec.SetField(d.Name, v)
			switch x := v.(type) {
			case *value.Tag:
				m.tags[d.Name] = x
			case *value.Constructor:
				m.tags[d.Name] = x.Tag
			}
		}
	}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		fn := &Func{Name: fd.Name, Params: fd.Params, Body: fd.Body, Scope: fd.Scope, Env: m.root}
		if fd.Owner == "" {
			m.rec.SetField(fd.Name, fn)
			continue
		}
		fn.Name = fd.Owner + "." + fd.Name
		if tag := m.tags[fd.Owner]; tag != nil { // nil only for a session type never bound
			tag.SetFunc(fd.Name, fn)
		}
	}
}

// bindLocalTypes binds the types a block declares in env, its environment for this run: one
// value per declaration for the whole run, however often the block runs, so every value it tags
// has one type (section 4). The first run makes them, and gives their field defaults an
// environment that keeps only the types visible in the block, since a default may name the
// block's types but none of its variables (section 4, Field defaults). The defaults are set
// before any goroutine can see the types.
func (in *Instance) bindLocalTypes(b *ast.Block, env *Env) {
	v, ok := in.localTypes.Load(b)
	if !ok {
		fresh := make([]value.Value, len(b.Types))
		for i, d := range b.Types {
			fresh[i] = TypeValue(d)
			env.set(d.Slot, fresh[i])
		}
		var types *Env
		for i, d := range b.Types {
			if hasDefaults(d) {
				if types == nil {
					types = typesOnly(env)
				}
				setDefaults(fresh[i], d, types)
			}
		}
		v, _ = in.localTypes.LoadOrStore(b, fresh)
	}
	for i, d := range b.Types {
		env.set(d.Slot, v.([]value.Value)[i])
	}
}

// typesOnly copies env and its parents up to the module's root, keeping only the types bound in
// them, so a field default holds no variable of the run that first bound its type.
func typesOnly(e *Env) *Env {
	if e == nil || e.parent == nil {
		return e // the module's root, which lives as long as the module
	}
	c := &Env{parent: typesOnly(e.parent), mod: e.mod, plain: make([]value.Value, len(e.plain)), shared: make([]value.Slot, len(e.shared))}
	for i, v := range e.plain {
		if !isType(v) {
			v = value.Uninitialized
		}
		c.plain[i] = v
	}
	for i := range e.shared {
		if v := e.shared[i].Load(); isType(v) {
			c.shared[i].Store(v)
		}
	}
	return c
}

func isType(v value.Value) bool {
	switch v.(type) {
	case *value.Tag, *value.Constructor:
		return true
	}
	return false
}

// fieldDefaults are the default expressions of a record tag's fields, nil for a field without
// one, and the environment of the declaration they are evaluated in.
type fieldDefaults struct {
	exprs []ast.Expr
	env   *Env
}

// NativeDefaults are the field defaults of a record tag the library builds, as the prelude
// declares them: per field, nil or a function that makes a fresh value.
type NativeDefaults []func() value.Value

func hasDefaults(d *ast.TypeDecl) bool {
	rt, ok := d.Body.(*ast.RecordType)
	return ok && slices.ContainsFunc(rt.Fields, func(f ast.FieldType) bool { return f.Default != nil })
}

// setDefaults gives the record tag v declares its field defaults, evaluated in env, if it has
// any.
func setDefaults(v value.Value, d *ast.TypeDecl, env *Env) {
	tag, ok := v.(*value.Tag)
	if !ok || !hasDefaults(d) {
		return
	}
	rt := d.Body.(*ast.RecordType)
	fd := &fieldDefaults{exprs: make([]ast.Expr, len(rt.Fields)), env: env}
	for i, f := range rt.Fields {
		fd.exprs[i] = f.Default
	}
	tag.Defaults = fd
}

// TypeValue builds what a type declaration binds: a record tag, a constructor, or an alias tag.
func TypeValue(d *ast.TypeDecl) value.Value {
	switch body := d.Body.(type) {
	case *ast.RecordType:
		fields := make([]string, len(body.Fields))
		for i, f := range body.Fields {
			fields[i] = f.Name
		}
		return value.NewTag(d.Name, value.RecordTag, fields)
	case *ast.TypeName:
		if len(body.Path) == 1 && len(body.Args) == 0 {
			if k, ok := value.KindByName(body.Path[0]); ok && k.IsPrimitive() {
				return &value.Constructor{Tag: value.NewPrimitiveTag(d.Name, k)}
			}
		}
	}
	return value.NewTag(d.Name, value.AliasTag, nil)
}

// initModules runs every module's top-level initializers in dependency order.
func (in *Instance) initModules(g *Goroutine) {
	for _, m := range in.modules {
		in.initModule(g, m)
	}
}

func (in *Instance) initModule(g *Goroutine, m *module) {
	g.push("module "+m.m.File.Path, m.m.File.Path, ast.Pos{})
	if in.dbg != nil {
		m.root.frame = g.stack[len(g.stack)-1].id
	}
	if m.m.File.Script {
		in.runScript(g, m)
		g.stack = g.stack[:len(g.stack)-1]
		return
	}
	for _, d := range m.m.File.Decls {
		vd, ok := d.(*ast.VarDecl)
		if !ok {
			continue
		}
		if in.dbg != nil {
			g.stopPoint(m.root, ast.Start(vd))
		}
		v := g.evalValue(vd.Init, m.root)
		g.bind(vd.Target, v, m.root)
	}
	g.stack = g.stack[:len(g.stack)-1]
}

// runScript runs a script's body: its top-level var initializers and statements in source
// order, in the module scope, as the body of a function only the runtime calls. Its deferred
// blocks run when it ends, and an Err it returns ends the program (section 10, Scripts); any
// other result, and the value of its last expression, are discarded.
func (in *Instance) runScript(g *Goroutine, m *module) {
	body := &ast.Block{Stmts: m.m.File.Stmts, HasDefer: true}
	v, c := g.evalBlock(body, m.root)
	if c != ctlReturn {
		return
	}
	if rec, ok := v.(*value.Record); ok && rec.Tag == value.ErrTag {
		panic(&resultSignal{err: rec.Get(0)})
	}
}

// importValue reads an imported name: the module record for a namespace, or a field.
func (g *Goroutine) importValue(env *Env, ref ast.Ref, pos ast.Pos) value.Value {
	ir := env.mod.importAt(ref.Index)
	if ir.namespace {
		return ir.rec
	}
	v, ok := ir.rec.Field(ref.Name)
	if !ok {
		g.panicf(pos, "%s has no export %s", ir.rec.Tag.Name, ref.Name)
	}
	if v == value.Uninitialized {
		g.panicf(pos, "%s is read before its initializer has run", ref.Name)
	}
	return v
}
