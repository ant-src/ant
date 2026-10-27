// Package check is the type checker (sections 4 through 8 and 10 of the specification). It
// walks every module of a loaded program, including code nothing calls, and reports every
// rule the specification calls a compile error.
package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// Diagnostic is one checker error, of the loader's shape.
type Diagnostic = loader.Diagnostic

// HostNamespace is a host namespace's parsed signatures.
type HostNamespace struct {
	Name  string
	Decls *ast.File
}

// Config adjusts the prelude a program sees.
type Config struct {
	Namespaces []HostNamespace
	Removed    []string
}

// Program checks every module of prog and returns the errors sorted by position.
func Program(prog *loader.Program, cfg Config) []Diagnostic {
	c := &Checker{prog: prog, seen: map[string]bool{}, flow: flow{}, shorted: map[ast.Expr]types.Type{}}
	c.loadPrelude(cfg)
	c.mods = make([]*moduleInfo, len(prog.Modules))
	for _, m := range prog.Modules {
		c.mods[m.Index] = &moduleInfo{m: m, path: m.File.Path, index: m.Index, names: map[string]*entity{}, exports: map[string]*entity{}}
	}
	for _, m := range prog.Modules {
		c.declareModule(c.mods[m.Index])
	}
	for _, m := range prog.Modules {
		c.resolveTypes(c.mods[m.Index])
	}
	for _, m := range prog.Modules {
		c.resolveSignatures(c.mods[m.Index])
	}
	for _, m := range prog.Modules {
		c.checkModule(c.mods[m.Index])
	}
	sort.SliceStable(c.diags, func(i, j int) bool {
		a, b := c.diags[i], c.diags[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Pos.Line != b.Pos.Line {
			return a.Pos.Line < b.Pos.Line
		}
		return a.Pos.Col < b.Pos.Col
	})
	return c.diags
}

// Checker holds the state of one check.
type Checker struct {
	prog    *loader.Program
	diags   []Diagnostic
	seen    map[string]bool
	mods    []*moduleInfo
	prelude map[string]*entity
	// bytes is the declaration of Bytes, built in like string: len, for, indexing and is take
	// it whether or not the host removed the namespace of its functions.
	bytes *types.Decl
	// html is the declaration of Html, the type of an HTML expression, which stays when the host
	// removes the namespace.
	html *types.Decl
	// unresolved holds the type declarations whose bodies are not resolved yet, so that a type
	// argument list naming a later declaration resolves it first and can be checked at the use.
	unresolved map[*types.Decl]*typeInfo
	// localTypes holds the types declared in blocks, made once per declaration however often
	// the block is checked.
	localTypes map[*ast.TypeDecl]*typeInfo

	// the context of the function or initializer being checked
	mod  *moduleInfo
	fn   *fnCtx
	flow flow
	// rigid holds the type parameters of the function being checked and of those around it:
	// each stands for a type the caller chooses, so a call inside the body cannot bind it.
	rigid map[*types.ParamDecl]bool
	// shorted holds, for each postfix link of a chain that a ?. may skip, the link's type when
	// nothing was null; the links after it in the chain read that type (see chainRecv).
	shorted map[ast.Expr]types.Type
}

func (c *Checker) errorf(pos ast.Pos, format string, args ...any) {
	path := ""
	if c.mod != nil {
		path = c.mod.path
	}
	msg := fmt.Sprintf(format, args...)
	key := fmt.Sprintf("%s:%d:%d:%s", path, pos.Line, pos.Col, msg)
	if c.seen[key] {
		return
	}
	c.seen[key] = true
	c.diags = append(c.diags, Diagnostic{Path: path, Pos: pos, Msg: msg})
}

// ---- entities

type entityKind int

const (
	entType      entityKind = iota // a declared type, with the functions declared on it
	entNamespace                   // a prelude or host namespace
	entFunc                        // a function
	entVar                         // a module variable
	entConst                       // a namespace constant
	entModule                      // an imported module
)

// entity is a named thing in a module's top-level scope or the prelude.
type entity struct {
	name     string
	kind     entityKind
	decl     *types.Decl
	members  map[string]*entity
	fn       *funcInfo
	sig      *sig
	v        *varInfo
	typ      types.Type
	exported bool
	mod      *moduleInfo
	pos      ast.Pos
	// conversion marks the twelve conversion functions, which also unwrap named primitives.
	conversion value.Kind
	isConv     bool
	// isLen marks len, which takes any value with a size.
	isLen bool
	// isPrintln marks println, which may also be called with no argument.
	isPrintln bool
}

// sig is a function's signature: its type parameters and function type, in which the
// parameter types mention the type parameters.
type sig struct {
	tparams []*types.ParamDecl
	fn      types.Func
	// defaults marks parameters with a default value.
	defaults []bool
	// inferring is set while the return type is being inferred.
	inferring bool
	ownerName string
	name      string
}

func (s *sig) fullName() string {
	if s.ownerName != "" {
		return s.ownerName + "." + s.name
	}
	return s.name
}

type moduleInfo struct {
	m       *loader.Module
	path    string
	index   int
	names   map[string]*entity
	exports map[string]*entity
	vars    []*varInfo
	funcs   []*funcInfo
	types   []*typeInfo
	// script is the state of a script's body (section 10): 0 not checked, 1 in progress, 2
	// done.
	script int
}

type typeInfo struct {
	decl *ast.TypeDecl
	t    *types.Decl
	ent  *entity
	mod  *moduleInfo
	// scope is the block a local type is declared in; nil at the top level.
	scope *scope
}

type funcInfo struct {
	decl    *ast.FuncDecl
	mod     *moduleInfo
	sig     *sig
	tparams map[string]*types.ParamDecl
	state   int
	ent     *entity
}

type varInfo struct {
	decl  *ast.VarDecl
	mod   *moduleInfo
	name  string
	typ   types.Type
	index int
	group *varGroup
	// reached is set for a script's var once the check of the script body has passed its
	// declaration, after which the body sees it as a local.
	reached bool
}

// varGroup ties the names a top-level destructuring binds to one initializer.
type varGroup struct {
	decl  *ast.VarDecl
	state int
	types map[string]types.Type
}

// ---- prelude

var builtinKinds = map[string]value.Kind{
	"int8": value.KindInt8, "int16": value.KindInt16, "int32": value.KindInt32, "int64": value.KindInt64,
	"uint8": value.KindUint8, "uint16": value.KindUint16, "uint32": value.KindUint32, "uint64": value.KindUint64,
	"float32": value.KindFloat32, "float64": value.KindFloat64, "int": value.KindInt64, "byte": value.KindUint8,
	"bool": value.KindBool, "string": value.KindString,
}

// loadPrelude builds the prelude's entities from prelude.ant and the host namespaces.
func (c *Checker) loadPrelude(cfg Config) {
	c.prelude = map[string]*entity{}
	f, err := prelude.Decls()
	if err != nil {
		panic("prelude: " + err.Error())
	}
	c.declareFile(f, "")
	c.bytes = c.prelude["Bytes"].decl
	c.html = c.prelude["Html"].decl
	for _, name := range cfg.Removed {
		delete(c.prelude, name)
	}
	if _, ok := c.prelude["Bytes"]; !ok {
		// The functions go with the namespace; the type stays, as string stays without String.
		c.prelude["Bytes"] = &entity{name: "Bytes", kind: entType, decl: c.bytes, exported: true}
	}
	if _, ok := c.prelude["Html"]; !ok {
		c.prelude["Html"] = &entity{name: "Html", kind: entType, decl: c.html, exported: true}
	}
	for _, ns := range cfg.Namespaces {
		c.declareFile(ns.Decls, ns.Name)
	}
	for _, name := range []string{"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "int", "byte", "string", "bool"} {
		if e, ok := c.prelude[name]; ok {
			e.isConv = true
			e.conversion = builtinKinds[name]
		}
	}
	if e, ok := c.prelude["len"]; ok {
		e.isLen = true
	}
	if e, ok := c.prelude["println"]; ok {
		e.isPrintln = true
	}
}

// declareFile declares the entities of a declarations-mode file into the prelude. In a host
// namespace, plain declarations belong to that namespace.
func (c *Checker) declareFile(f *ast.File, defaultOwner string) {
	saved := c.mod
	c.mod = &moduleInfo{path: f.Path, names: map[string]*entity{}}
	defer func() { c.mod = saved }()
	owner := func(name string) *entity {
		if name == "" {
			return nil
		}
		e, ok := c.prelude[name]
		if !ok {
			e = &entity{name: name, kind: entNamespace, members: map[string]*entity{}}
			c.prelude[name] = e
		}
		if e.members == nil {
			e.members = map[string]*entity{}
		}
		return e
	}
	// types first, so signatures can mention them
	var pending []*typeInfo
	for _, d := range f.Decls {
		td, ok := d.(*ast.TypeDecl)
		if !ok {
			continue
		}
		own := td.Owner
		if own == "" {
			own = defaultOwner
		}
		decl := &types.Decl{Name: td.Name, Pos: td.Pos, Path: c.mod.path, Kind: declKind(td.Body)}
		if own != "" {
			decl.Name = own + "." + td.Name
		}
		for _, tp := range td.TypeParams {
			decl.Params = append(decl.Params, &types.ParamDecl{Name: tp.Name})
		}
		ent := &entity{name: td.Name, kind: entType, decl: decl, pos: td.Pos, exported: true}
		if own != "" {
			owner(own).members[td.Name] = ent
		} else {
			if existing, ok := c.prelude[td.Name]; ok && existing.kind == entNamespace {
				// a namespace whose functions were declared before its type: merge
				ent.members = existing.members
			}
			c.prelude[td.Name] = ent
		}
		pending = append(pending, &typeInfo{decl: td, t: decl, ent: ent})
	}
	c.resolveTypeDecls(pending)
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			own := d.Owner
			if own == "" {
				own = defaultOwner
			}
			fi := &funcInfo{decl: d, mod: c.mod}
			fi.sig = c.signature(d, nil, own)
			ent := &entity{name: d.Name, kind: entFunc, sig: fi.sig, pos: d.Pos, exported: true}
			if own == "" {
				c.prelude[d.Name] = ent
				continue
			}
			o := owner(own)
			o.members[d.Name] = ent
			if o.decl != nil {
				if o.decl.Funcs == nil {
					o.decl.Funcs = map[string]*types.Func{}
				}
				o.decl.Funcs[d.Name] = &fi.sig.fn
			}
		case *ast.VarDecl:
			own := d.Owner
			if own == "" {
				own = defaultOwner
			}
			name := d.Target.(*ast.NameTarget).Name
			ent := &entity{name: name, kind: entConst, typ: c.resolveType(d.Type, nil), pos: d.Pos, exported: true}
			owner(own).members[name] = ent
		}
	}
}

// ---- module declarations

func (c *Checker) declareModule(mi *moduleInfo) {
	c.mod = mi
	f := mi.m.File
	for _, imp := range f.Imports {
		target := c.mods[imp.Module]
		if imp.Namespace != "" {
			mi.names[imp.Namespace] = &entity{name: imp.Namespace, kind: entModule, mod: target, members: target.exports, pos: imp.Pos}
			continue
		}
		for _, n := range imp.Names {
			// Only the target's own exported declarations are visible: a name the target
			// imported is not re-exported (section 10).
			ent, ok := target.exports[n.Name]
			if !ok {
				if _, declared := target.names[n.Name]; declared {
					c.errorf(n.Pos, "%s is not exported by %s", n.Name, imp.Path)
				} else {
					c.errorf(n.Pos, "%s has no declaration %s", imp.Path, n.Name)
				}
				continue
			}
			mi.names[n.Alias] = ent
		}
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.TypeDecl:
			decl := &types.Decl{Name: d.Name, Pos: d.Pos, Path: mi.path, Kind: declKind(d.Body)}
			for _, tp := range d.TypeParams {
				decl.Params = append(decl.Params, &types.ParamDecl{Name: tp.Name})
			}
			ent := &entity{name: d.Name, kind: entType, decl: decl, exported: d.Export, pos: d.Pos, members: map[string]*entity{}}
			mi.names[d.Name] = ent
			mi.types = append(mi.types, &typeInfo{decl: d, t: decl, ent: ent, mod: mi})
			if d.Export {
				mi.exports[d.Name] = ent
			}
		case *ast.FuncDecl:
			fi := &funcInfo{decl: d, mod: mi}
			mi.funcs = append(mi.funcs, fi)
			if d.Owner != "" {
				continue
			}
			ent := &entity{name: d.Name, kind: entFunc, fn: fi, exported: d.Export, pos: d.Pos}
			fi.ent = ent
			mi.names[d.Name] = ent
			if d.Export {
				mi.exports[d.Name] = ent
			}
		case *ast.VarDecl:
			group := &varGroup{decl: d}
			for _, name := range targetNames(d.Target) {
				v := &varInfo{decl: d, mod: mi, name: name.Name, index: len(mi.vars), group: group}
				mi.vars = append(mi.vars, v)
				ent := &entity{name: name.Name, kind: entVar, v: v, exported: d.Export, pos: name.Pos}
				mi.names[name.Name] = ent
				if d.Export {
					mi.exports[name.Name] = ent
				}
			}
		}
	}
}

func targetNames(t ast.Target) []*ast.NameTarget {
	switch t := t.(type) {
	case *ast.NameTarget:
		if t.Slot.Discard {
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

// resolveTypes resolves the bodies of a module's type declarations and attaches the functions
// declared on its types.
func (c *Checker) resolveTypes(mi *moduleInfo) {
	c.mod = mi
	c.resolveTypeDecls(mi.types)
	for _, fi := range mi.funcs {
		d := fi.decl
		if d.Owner == "" {
			continue
		}
		owner, ok := mi.names[d.Owner]
		if !ok || owner.kind != entType {
			continue // the resolver reported it
		}
		if d.Export && !owner.exported {
			c.errorf(d.Pos, "%s.%s is exported but %s is not", d.Owner, d.Name, d.Owner)
		}
		ent := &entity{name: d.Name, kind: entFunc, fn: fi, exported: d.Export, pos: d.Pos}
		fi.ent = ent
		owner.members[d.Name] = ent
	}
}

// declKind is the kind a type declaration's body gives it, known before any body is resolved, so
// that a body naming a later declaration sees whether it is a record, a primitive, or an alias.
// resolveTypeDecl settles which kind of opaque type a declaration without a body is.
func declKind(body ast.Type) types.DeclKind {
	switch b := body.(type) {
	case nil:
		return types.OpaqueDecl
	case *ast.RecordType:
		return types.RecordDecl
	case *ast.TypeName:
		if _, ok := builtinKinds[b.Path[0]]; ok && len(b.Path) == 1 && len(b.Args) == 0 {
			return types.PrimitiveDecl
		}
	}
	return types.AliasDecl
}

// resolveTypeDecls resolves the bodies of type declarations in order, except those a type
// argument list already resolved. A cyclic alias, or a generic one that refers to itself with
// other type arguments, is marked invalid as soon as its cycle is resolved, before a later
// declaration can expand it.
func (c *Checker) resolveTypeDecls(tis []*typeInfo) {
	if c.unresolved == nil {
		c.unresolved = map[*types.Decl]*typeInfo{}
	}
	for _, ti := range tis {
		c.unresolved[ti.t] = ti
	}
	for _, ti := range tis {
		c.resolveTypeOnce(ti)
	}
}

func (c *Checker) resolveTypeOnce(ti *typeInfo) {
	if c.unresolved[ti.t] != ti {
		return
	}
	delete(c.unresolved, ti.t)
	c.resolveTypeDecl(ti)
	c.checkAliasCycle(ti)
	c.checkAliasArgs(ti)
	c.checkCycleUnions(ti)
}

// checkCycleUnions checks again the unions of the declarations ti reaches in its module. A union
// that names a declaration of its own cycle was formed before that declaration's body was known,
// as Array[L[T] | Array[int]] in the body of L, and only now shows whether its members share a
// discriminator (section 4, Unions).
func (c *Checker) checkCycleUnions(ti *typeInfo) {
	own := map[*types.Decl]bool{}
	for _, t := range c.mod.types {
		own[t.t] = true
	}
	if !own[ti.t] {
		return
	}
	reached := map[*types.Decl]bool{ti.t: true}
	order := []*types.Decl{ti.t}
	for i := 0; i < len(order); i++ {
		for _, t := range declTypes(order[i]) {
			types.Walk(t, func(t types.Type) {
				if n, ok := t.(types.Named); ok && own[n.Decl] && !reached[n.Decl] {
					reached[n.Decl] = true
					order = append(order, n.Decl)
				}
			})
		}
	}
	for _, d := range order {
		check := func(t types.Type) bool {
			ok := true
			types.Walk(t, func(t types.Type) {
				if u, isUnion := t.(types.Union); isUnion && ok {
					if err := types.CheckUnion(u.Members); err != nil {
						pos, _ := d.Pos.(ast.Pos)
						c.errorf(pos, "type %s: %s", d.Name, err)
						ok = false
					}
				}
			})
			return ok
		}
		if d.Kind == types.AliasDecl && d.Body != nil && !check(d.Body) {
			d.Body = types.Invalid{}
		}
		for i, f := range d.Fields {
			if !check(f.Type) {
				d.Fields[i].Type = types.Invalid{}
			}
		}
	}
}

// declTypes lists the types a declaration's body consists of.
func declTypes(d *types.Decl) []types.Type {
	var out []types.Type
	if d.Kind == types.AliasDecl && d.Body != nil {
		out = append(out, d.Body)
	}
	for _, f := range d.Fields {
		out = append(out, f.Type)
	}
	return out
}

func (c *Checker) resolveTypeDecl(ti *typeInfo) {
	d, decl := ti.decl, ti.t
	sc := &scope{parent: ti.scope, tparams: map[string]*types.ParamDecl{}, typeDecl: d.Name}
	for i, tp := range d.TypeParams {
		sc.tparams[tp.Name] = decl.Params[i]
	}
	for i, tp := range d.TypeParams {
		if tp.Bound != nil {
			decl.Params[i].Bound = c.resolveType(tp.Bound, sc)
		}
	}
	switch body := d.Body.(type) {
	case nil:
		// The time types of stdlib.md, the opaque types the prelude declares under Time, and
		// Regex.Pattern carry their tag, so is tests them; every other opaque type, a host's
		// included, is a sync type, which is cannot test.
		if (strings.HasPrefix(decl.Name, "Time.") || decl.Name == "Regex.Pattern") && c.mod != nil && c.mod.path == "prelude.ant" {
			decl.Kind = types.TaggedOpaqueDecl
		} else {
			decl.Kind = types.OpaqueDecl
		}
	case *ast.RecordType:
		decl.Kind = types.RecordDecl
		seen := map[string]bool{}
		for _, f := range body.Fields {
			if seen[f.Name] {
				c.errorf(f.Pos, "%s declares the field %s twice", d.Name, f.Name)
			}
			seen[f.Name] = true
			decl.Fields = append(decl.Fields, types.Field{Name: f.Name, Type: c.resolveType(f.Type, sc), Default: f.Default != nil})
		}
	case *ast.TypeName:
		if len(body.Path) == 1 && len(body.Args) == 0 {
			if k, ok := builtinKinds[body.Path[0]]; ok {
				decl.Kind = types.PrimitiveDecl
				decl.Base = k
				return
			}
		}
		t := c.resolveType(body, sc)
		if n, ok := t.(types.Named); ok && (n.Decl.Kind == types.RecordDecl || n.Decl.Kind == types.PrimitiveDecl) {
			c.errorf(d.Pos, "type %s = %s would mint a second tag for one type; declare the fields or the base type again", d.Name, n.Decl.Name)
			decl.Kind = types.AliasDecl
			decl.Body = types.Invalid{}
			return
		}
		decl.Kind = types.AliasDecl
		decl.Body = t
	default:
		decl.Kind = types.AliasDecl
		decl.Body = c.resolveType(d.Body, sc)
	}
}

// checkAliasCycle reports an alias that reaches itself without passing through a record,
// array, map, set, or function. The error goes to the member of the cycle declared last, which
// is the one whose body closes the cycle when the declarations are read in order, whichever
// member a later declaration's use resolved first.
func (c *Checker) checkAliasCycle(ti *typeInfo) {
	if ti.t.Kind != types.AliasDecl {
		return
	}
	visited := map[*types.Decl]bool{}
	var path []*types.Decl
	var walk func(t types.Type) bool
	walk = func(t types.Type) bool {
		switch x := t.(type) {
		case types.Named:
			if x.Decl.Kind != types.AliasDecl {
				return false
			}
			if x.Decl == ti.t {
				return true
			}
			if visited[x.Decl] {
				return false
			}
			visited[x.Decl] = true
			path = append(path, x.Decl)
			if walk(x.Decl.Body) {
				return true
			}
			path = path[:len(path)-1]
		case types.Union:
			// The members as written first: a union flattens the aliases it names, and the
			// path through them says which declarations are in the cycle.
			for _, m := range append(x.Written, x.Members...) {
				if walk(m) {
					return true
				}
			}
		}
		return false
	}
	if !walk(ti.t.Body) {
		return
	}
	last, lastPos := ti.t, ti.decl.Pos
	for _, d := range path {
		if pos, ok := d.Pos.(ast.Pos); ok && (pos.Line > lastPos.Line || pos.Line == lastPos.Line && pos.Col > lastPos.Col) {
			last, lastPos = d, pos
		}
	}
	name := last.Name
	if last == ti.t {
		name = ti.decl.Name
	}
	c.errorf(lastPos, "type %s refers to itself without passing through a record, array, map, set, or function", name)
	last.Body = types.Invalid{}
}

// checkAliasArgs reports a generic alias that refers to itself, directly or through other
// aliases, with anything but its own type parameters in order (section 4, Type declarations):
// each level of N[A, B] = Array[N[B, Box[A]]] | ... would be a new instantiation whose unions
// are never checked. A cycle is complete once its last declaration is resolved, which is ti, so
// every alias ti reaches is checked again.
func (c *Checker) checkAliasArgs(ti *typeInfo) {
	if ti.t.Kind != types.AliasDecl {
		return
	}
	reached := map[*types.Decl]bool{}
	var reach func(t types.Type)
	reach = func(t types.Type) {
		types.Walk(t, func(t types.Type) {
			if n, ok := t.(types.Named); ok && n.Decl.Kind == types.AliasDecl && n.Decl.Body != nil && !reached[n.Decl] {
				reached[n.Decl] = true
				reach(n.Decl.Body)
			}
		})
	}
	reached[ti.t] = true
	reach(ti.t.Body)
	type violation struct {
		decl *types.Decl
		msg  string
	}
	var bad []violation
	for d := range reached {
		if msg := selfArgs(d); msg != "" {
			bad = append(bad, violation{d, msg})
		}
	}
	for _, v := range bad {
		pos, _ := v.decl.Pos.(ast.Pos) // only this module's aliases can still be in a cycle
		c.errorf(pos, "%s", v.msg)
		v.decl.Body = types.Invalid{}
	}
}

// selfArgs describes how the generic alias d refers to itself with other type arguments than
// its own parameters in order, or returns "" when it does not. It follows d's body through the
// aliases it names, tracking which of their parameters stand for which of d's.
func selfArgs(d *types.Decl) string {
	if len(d.Params) == 0 || d.Body == nil {
		return ""
	}
	own := make([]types.Type, len(d.Params))
	for i, p := range d.Params {
		own[i] = types.Param{Decl: p}
	}
	want := types.String(types.Named{Decl: d, Args: own})
	// env maps a parameter of the alias being followed to the parameter of d it stands for.
	type visit struct {
		decl *types.Decl
		env  string
	}
	visited := map[visit]bool{}
	var msg string
	var follow func(t types.Type, env map[*types.ParamDecl]*types.ParamDecl, via *types.Decl)
	follow = func(t types.Type, env map[*types.ParamDecl]*types.ParamDecl, via *types.Decl) {
		types.Walk(t, func(t types.Type) {
			n, ok := t.(types.Named)
			if msg != "" || !ok || n.Decl.Kind != types.AliasDecl || n.Decl.Body == nil {
				return
			}
			next := map[*types.ParamDecl]*types.ParamDecl{}
			key := ""
			for i, p := range n.Decl.Params {
				var stands *types.ParamDecl
				if i < len(n.Args) {
					if a, ok := n.Args[i].(types.Param); ok {
						stands = env[a.Decl]
					}
				}
				next[p] = stands
				key += fmt.Sprintf("%p,", stands)
			}
			if n.Decl == d {
				for i, p := range d.Params {
					if next[n.Decl.Params[i]] != p {
						if via == nil {
							msg = fmt.Sprintf("type %s refers to itself as %s; a generic alias must refer to itself as %s, with its own type parameters in order", d.Name, types.String(n), want)
						} else {
							msg = fmt.Sprintf("type %s refers to itself through %s with other type arguments; a generic alias must refer to itself as %s, with its own type parameters in order", d.Name, via.Name, want)
						}
						return
					}
				}
				return
			}
			if visited[visit{n.Decl, key}] {
				return
			}
			visited[visit{n.Decl, key}] = true
			if via == nil {
				follow(n.Decl.Body, next, n.Decl)
			} else {
				follow(n.Decl.Body, next, via)
			}
		})
	}
	env := map[*types.ParamDecl]*types.ParamDecl{}
	for _, p := range d.Params {
		env[p] = p
	}
	follow(d.Body, env, nil)
	return msg
}

// resolveSignatures resolves every function's parameter and declared return types.
func (c *Checker) resolveSignatures(mi *moduleInfo) {
	c.mod = mi
	for _, fi := range mi.funcs {
		var owner string
		var ownerDecl *types.Decl
		if fi.decl.Owner != "" {
			owner = fi.decl.Owner
			if ent, ok := mi.names[owner]; ok && ent.kind == entType {
				ownerDecl = ent.decl
			}
		}
		fi.sig = c.signature(fi.decl, nil, owner)
		fi.tparams = map[string]*types.ParamDecl{}
		for i, tp := range fi.decl.TypeParams {
			fi.tparams[tp.Name] = fi.sig.tparams[i]
		}
		if fi.ent != nil {
			fi.ent.sig = fi.sig
		}
		if ownerDecl != nil {
			if ownerDecl.Funcs == nil {
				ownerDecl.Funcs = map[string]*types.Func{}
			}
			ownerDecl.Funcs[fi.decl.Name] = &fi.sig.fn
		}
		if fi.decl.Export && fi.decl.Return == nil {
			c.errorf(fi.decl.Pos, "the exported function %s must state its return type", fi.sig.fullName())
		}
	}
}

// signature resolves a function declaration's types. Its return type is nil when inferred.
func (c *Checker) signature(d *ast.FuncDecl, parent *scope, owner string) *sig {
	s := &sig{name: d.Name, ownerName: owner}
	sc := &scope{parent: parent, tparams: map[string]*types.ParamDecl{}}
	for _, tp := range d.TypeParams {
		pd := &types.ParamDecl{Name: tp.Name}
		s.tparams = append(s.tparams, pd)
		sc.tparams[tp.Name] = pd
	}
	for i, tp := range d.TypeParams {
		if tp.Bound != nil {
			s.tparams[i].Bound = c.resolveType(tp.Bound, sc)
		}
	}
	s.fn.Required = len(d.Params)
	for i, p := range d.Params {
		s.fn.Params = append(s.fn.Params, c.resolveType(p.Type, sc))
		s.fn.Names = append(s.fn.Names, p.Name)
		s.defaults = append(s.defaults, p.Default != nil)
		if p.Default != nil && s.fn.Required == len(d.Params) {
			s.fn.Required = i
		}
	}
	if d.Return != nil {
		s.fn.Return = c.resolveType(d.Return, sc)
	}
	return s
}

// ---- type resolution

// scope is a lexical scope of the checker: locals, the types a block declares, and type
// parameters.
type scope struct {
	parent  *scope
	names   map[string]*local
	types   map[string]*entity
	tparams map[string]*types.ParamDecl
	// typeDecl names the type whose body this scope resolves.
	typeDecl string
}

// lookupTypeParam finds a type parameter. outside names the type declared in a block whose
// body reaches past its own parameters to those of the enclosing function, or is "".
func (sc *scope) lookupTypeParam(name string) (p *types.ParamDecl, outside string) {
	for s := sc; s != nil; s = s.parent {
		if s.tparams != nil {
			if p, ok := s.tparams[name]; ok {
				return p, outside
			}
		}
		if s.typeDecl != "" && outside == "" {
			outside = s.typeDecl
		}
	}
	return nil, ""
}

// resolveType turns a type expression into a type, reporting unknown names and wrong
// arities.
func (c *Checker) resolveType(t ast.Type, sc *scope) types.Type {
	switch x := t.(type) {
	case nil:
		return types.Invalid{}
	case *ast.TypeName:
		return c.resolveTypeName(x, sc)
	case *ast.RecordType:
		var fields []types.Field
		seen := map[string]bool{}
		for _, f := range x.Fields {
			if seen[f.Name] {
				c.errorf(f.Pos, "the field %s is listed twice", f.Name)
			}
			seen[f.Name] = true
			fields = append(fields, types.Field{Name: f.Name, Type: c.resolveType(f.Type, sc)})
		}
		return types.Record{Fields: fields}
	case *ast.FuncType:
		f := types.Func{}
		for _, p := range x.Params {
			f.Params = append(f.Params, c.resolveType(p.Type, sc))
		}
		f.Required = len(f.Params)
		f.Return = c.resolveType(x.Return, sc)
		return f
	case *ast.UnionType:
		var members []types.Type
		for _, m := range x.Members {
			members = append(members, c.resolveType(m, sc))
		}
		u, err := types.MakeUnion(members)
		if err != nil {
			c.errorf(x.Pos, "%s", err)
			return types.Invalid{}
		}
		return u
	}
	return types.Invalid{}
}

func (c *Checker) resolveTypeName(x *ast.TypeName, sc *scope) types.Type {
	name := x.Path[0]
	if len(x.Path) == 1 {
		if p, outside := sc.lookupTypeParam(name); p != nil {
			if outside != "" {
				c.errorf(x.Pos, "%s is declared in a block and cannot mention %s, a type parameter of the enclosing function", outside, name)
				return types.Invalid{}
			}
			if len(x.Args) > 0 {
				c.errorf(x.Pos, "the type parameter %s takes no type arguments", name)
			}
			return types.Param{Decl: p}
		}
		switch name {
		case "null":
			return types.Null{}
		case "void":
			return types.Void{}
		case "never":
			return types.Never{}
		case "Array":
			if len(x.Args) != 1 {
				c.errorf(x.Pos, "Array takes one type argument")
				return types.Invalid{}
			}
			return types.Array{Elem: c.resolveType(x.Args[0], sc)}
		case "Map":
			if len(x.Args) != 2 {
				c.errorf(x.Pos, "Map takes two type arguments")
				return types.Invalid{}
			}
			return types.Map{Key: c.resolveType(x.Args[0], sc), Val: c.resolveType(x.Args[1], sc)}
		case "Set":
			if len(x.Args) != 1 {
				c.errorf(x.Pos, "Set takes one type argument")
				return types.Invalid{}
			}
			return types.Set{Elem: c.resolveType(x.Args[0], sc)}
		}
		if k, ok := builtinKinds[name]; ok {
			if len(x.Args) > 0 {
				c.errorf(x.Pos, "%s takes no type arguments", name)
			}
			return types.Prim{Kind: k}
		}
	}
	ent := c.lookupType(sc, name)
	if ent == nil {
		c.errorf(x.Pos, "unknown type %s", name)
		return types.Invalid{}
	}
	for _, part := range x.Path[1:] {
		if ent.kind == entModule {
			next := c.moduleMember(ent, part, x.Pos)
			if next == nil {
				return types.Invalid{}
			}
			ent = next
			continue
		}
		if ent.members == nil {
			c.errorf(x.Pos, "%s has no type %s", ent.name, part)
			return types.Invalid{}
		}
		next, ok := ent.members[part]
		if !ok {
			c.errorf(x.Pos, "%s has no type %s", ent.name, part)
			return types.Invalid{}
		}
		ent = next
	}
	if ent.kind != entType {
		c.errorf(x.Pos, "%s is not a type", x.Path[len(x.Path)-1])
		return types.Invalid{}
	}
	decl := ent.decl
	if len(x.Args) != len(decl.Params) {
		c.errorf(x.Pos, "%s takes %s, got %d", decl.Name, plural(len(decl.Params), "type argument"), len(x.Args))
		return types.Invalid{}
	}
	var args []types.Type
	for _, a := range x.Args {
		args = append(args, c.resolveType(a, sc))
	}
	if ti := c.unresolved[decl]; ti != nil && (len(args) > 0 || decl.Kind == types.AliasDecl) {
		// A later declaration: its body is needed to check the arguments here, or to form a
		// union that names the alias. One whose body is being resolved is in a cycle with this
		// use, which checkAliasCycle and checkAliasArgs cover.
		c.resolveTypeOnce(ti)
	}
	n := types.Named{Decl: decl, Args: args}
	if decl.Kind == types.AliasDecl && len(args) > 0 {
		// Substituting arguments forms unions under the same rules, at the use.
		if _, err := types.Subst(decl.Body, types.Bindings(decl.Params, args)); err != nil {
			c.errorf(x.Pos, "%s: %s", types.String(n), err)
			return types.Invalid{}
		}
	}
	if decl.Kind == types.RecordDecl && len(args) > 0 {
		// The same holds for a field type, as A | B in { v: A | B }.
		for _, f := range decl.Fields {
			if _, err := types.Subst(f.Type, types.Bindings(decl.Params, args)); err != nil {
				c.errorf(x.Pos, "%s: %s", types.String(n), err)
				return types.Invalid{}
			}
		}
	}
	return n
}

// lookupType finds a type name: one declared in an enclosing block, else a module-level or
// prelude entity. A local variable does not hide a type here, since a type position holds no
// value.
func (c *Checker) lookupType(sc *scope, name string) *entity {
	for s := sc; s != nil; s = s.parent {
		if e, ok := s.types[name]; ok {
			return e
		}
	}
	return c.lookupEntity(name)
}

// lookupEntity finds a module-level or prelude name.
func (c *Checker) lookupEntity(name string) *entity {
	if c.mod != nil {
		if e, ok := c.mod.names[name]; ok {
			return e
		}
	}
	if e, ok := c.prelude[name]; ok {
		return e
	}
	return nil
}

// ---- modules

// checkModule checks a module's initializers and function bodies.
func (c *Checker) checkModule(mi *moduleInfo) {
	c.mod = mi
	if mi.m.File.Script {
		c.checkScript(mi)
	}
	for _, ti := range mi.types {
		if ti.scope == nil {
			c.checkDefaults(ti)
		}
	}
	for _, v := range mi.vars {
		c.varType(v, ast.Pos{}, false)
	}
	for _, fi := range mi.funcs {
		c.checkFunc(fi)
	}
}

// checkDefaults checks the field defaults of a record type against the fields' types, in the
// scope of the declaration, outside every function, with the type's own parameters rigid as in
// a generic function's body (section 4, Records).
func (c *Checker) checkDefaults(ti *typeInfo) {
	rt, ok := ti.decl.Body.(*ast.RecordType)
	if !ok || ti.t.Kind != types.RecordDecl {
		return
	}
	savedMod, savedFn, savedFlow, savedRigid := c.mod, c.fn, c.flow, c.rigid
	c.mod = ti.mod
	c.flow = flow{}
	tparams := map[string]*types.ParamDecl{}
	for i, tp := range ti.decl.TypeParams {
		tparams[tp.Name] = ti.t.Params[i]
	}
	c.rigid = withRigid(c.rigid, tparams)
	sc := &scope{parent: ti.scope, names: map[string]*local{}, tparams: tparams, typeDecl: ti.decl.Name}
	for i, f := range rt.Fields {
		if f.Default != nil {
			c.fn = &fnCtx{kind: "init", name: "default of " + ti.decl.Name + "." + f.Name}
			c.check(f.Default, sc, ti.t.Fields[i].Type)
		}
	}
	c.mod, c.fn, c.flow, c.rigid = savedMod, savedFn, savedFlow, savedRigid
}

// checkScript checks a script's body: its top-level var initializers and statements in source
// order, as the body of a function only the runtime calls (section 10, Scripts). The script's
// vars are its locals, so they narrow like locals; its functions, nested bodies declared
// without a position, see them at their declared types, through the module's entities.
func (c *Checker) checkScript(mi *moduleInfo) {
	if mi.script != 0 {
		return
	}
	mi.script = 1
	savedMod, savedFn, savedFlow, savedRigid := c.mod, c.fn, c.flow, c.rigid
	c.mod = mi
	c.fn = &fnCtx{kind: "script", name: "the script"}
	c.flow = flow{}
	c.rigid = nil
	sc := &scope{names: map[string]*local{}}
	for _, s := range mi.m.File.Stmts {
		switch st := s.(type) {
		case *ast.VarDecl:
			c.scriptVar(st, sc)
		case ast.Expr:
			c.synth(st, sc)
		}
	}
	mi.script = 2
	c.mod, c.fn, c.flow, c.rigid = savedMod, savedFn, savedFlow, savedRigid
}

// scriptVar checks a script's top-level var as a local of the script body, and gives the
// module's entity the type it declares, for the script's functions.
func (c *Checker) scriptVar(d *ast.VarDecl, sc *scope) {
	c.varDecl(d, sc)
	for _, nt := range targetNames(d.Target) {
		ent := c.mod.names[nt.Name]
		if ent == nil || ent.kind != entVar || ent.v.decl != d {
			continue
		}
		if ent.v.typ == nil {
			if l := sc.names[nt.Name]; l != nil {
				ent.v.typ = l.typ
			}
		}
		ent.v.reached = true
	}
}

// scriptResult checks what a script body returns: void or an Err (section 10, Scripts). A call
// such as Result.err(e) takes its Ok type from the void the script stands for.
func (c *Checker) scriptResult(e ast.Expr, sc *scope) {
	var t types.Type
	if call, ok := e.(*ast.Call); ok {
		okT := types.Named{Decl: c.prelude["Ok"].decl, Args: []types.Type{types.Void{}}}
		errT := types.Named{Decl: c.prelude["Err"].decl, Args: []types.Type{types.Invalid{}}}
		t = c.call(call, sc, types.Union{Members: []types.Type{okT, errT}})
	} else {
		t = c.synth(e, sc)
	}
	if types.IsInvalid(t) || types.IsNever(t) {
		return
	}
	for _, m := range types.Members(t) {
		switch x := types.Expand(m).(type) {
		case types.Void, types.Never:
			continue
		case types.Named:
			if x.Decl == c.prelude["Err"].decl {
				continue
			}
			if x.Decl == c.prelude["Ok"].decl && len(x.Args) == 1 && types.Equal(x.Args[0], types.Void{}) {
				continue // Result[void, E], which is void or an Err
			}
		}
		c.errorf(e.Position(), "a script returns void or an Err, not %s", c.typeNames(t)[0])
		return
	}
}

// varType returns a module variable's type, inferring it from its initializer on first use.
// direct says the use is in a top-level initializer outside any nested body, where only the
// variables above may be referred to (section 10).
func (c *Checker) varType(v *varInfo, pos ast.Pos, direct bool) types.Type {
	if direct && c.fn != nil && c.fn.kind == "script" && c.fn.initVar == nil && v.mod == c.mod && !v.reached {
		c.errorf(pos, "%s is declared later in the module; a top-level initializer or statement may refer only to the variables above it", v.name)
		return types.Invalid{}
	}
	if v.typ == nil && v.mod.m.File.Script && v.mod.script == 0 {
		// A script's vars take their types from the body, which a function may need first.
		c.checkScript(v.mod)
		if v.typ != nil {
			return v.typ
		}
	}
	if direct && c.fn != nil && c.fn.initVar != nil && c.fn.initVar.mod == v.mod && v.index >= c.fn.initVar.index {
		c.errorf(pos, "%s is declared later in the module; a top-level initializer or statement may refer only to the variables above it", v.name)
		return types.Invalid{}
	}
	if v.typ != nil {
		return v.typ
	}
	g := v.group
	if g.state == 1 {
		c.errorf(pos, "the type of %s depends on itself; annotate it", v.name)
		return types.Invalid{}
	}
	if g.state == 2 {
		if t, ok := g.types[v.name]; ok {
			v.typ = t
			return t
		}
		return types.Invalid{}
	}
	g.state = 1
	savedMod, savedFn, savedFlow := c.mod, c.fn, c.flow
	c.mod = v.mod
	c.fn = &fnCtx{kind: "init", initVar: v, name: "initializer of " + v.name}
	if v.mod.m.File.Script {
		// The script body is in progress and has not reached v, which a function needs now:
		// its initializer is checked out of order, where try still returns from the script.
		c.fn.kind = "script"
	}
	c.flow = flow{}
	sc := &scope{names: map[string]*local{}}
	d := g.decl
	var t types.Type
	if d.Type != nil {
		declared := c.resolveType(d.Type, sc)
		c.check(d.Init, sc, declared)
		t = declared
	} else {
		t = c.synth(d.Init, sc)
	}
	g.types = map[string]types.Type{}
	c.bindTargetTypes(d.Target, t, func(name string, ty types.Type) { g.types[name] = ty })
	g.state = 2
	c.mod, c.fn, c.flow = savedMod, savedFn, savedFlow
	if ty, ok := g.types[v.name]; ok {
		v.typ = ty
		return ty
	}
	v.typ = types.Invalid{}
	return v.typ
}

// ---- functions

// fnCtx is the function, lambda, go block, defer block, or initializer being checked.
type fnCtx struct {
	parent   *fnCtx
	kind     string
	name     string
	declared types.Type
	returns  []types.Type
	// tries holds the try expressions of a body whose return type is inferred, which must
	// come out a Result, or a Result or null (section 8).
	tries   []*ast.Unary
	fi      *funcInfo
	initVar *varInfo
	loops   []*loopCtx
}

type loopCtx struct {
	// breaks holds the state at each break, which the state after the loop joins.
	breaks []flow
}

// local is a local variable or parameter. pending marks a var whose initializer is being
// checked; stated says its type was written.
type local struct {
	name    string
	typ     types.Type
	slot    ast.Slot
	ctx     *fnCtx
	pos     ast.Pos
	pending bool
	// unset marks a parameter with a default while the defaults up to its own are checked.
	unset  bool
	stated bool
	// hidden marks the x of a lambda of arity zero, which is not in scope (section 6).
	hidden bool
}

// checkFunc checks a declared function's body, inferring its return type when it has none.
func (c *Checker) checkFunc(fi *funcInfo) {
	if fi.state != 0 {
		return
	}
	fi.state = 1
	savedMod, savedFn, savedFlow, savedRigid := c.mod, c.fn, c.flow, c.rigid
	c.mod = fi.mod
	s := fi.sig
	ctx := &fnCtx{kind: "func", name: s.fullName(), declared: s.fn.Return, fi: fi}
	c.fn = ctx
	c.flow = flow{}
	// A declared function is top level, so only its own type parameters are rigid in it, even
	// when its return type is inferred on demand from inside another generic function's body.
	c.rigid = withRigid(nil, fi.tparams)
	sc := &scope{names: map[string]*local{}, tparams: fi.tparams}
	c.declareParams(fi.decl.Params, s, sc, ctx)
	if fi.decl.Body != nil {
		s.inferring = s.fn.Return == nil
		t := c.inferBody(fi.decl.Body, sc, ctx)
		if s.fn.Return == nil {
			s.fn.Return = t
		}
		s.inferring = false
	}
	fi.state = 2
	c.mod, c.fn, c.flow, c.rigid = savedMod, savedFn, savedFlow, savedRigid
}

// withRigid is rigid with a function's own type parameters added.
func withRigid(rigid map[*types.ParamDecl]bool, own map[string]*types.ParamDecl) map[*types.ParamDecl]bool {
	if len(own) == 0 {
		return rigid
	}
	out := make(map[*types.ParamDecl]bool, len(rigid)+len(own))
	for p := range rigid {
		out[p] = true
	}
	for _, p := range own {
		out[p] = true
	}
	return out
}

// declareParams binds parameters as locals and checks their defaults.
func (c *Checker) declareParams(params []ast.Param, s *sig, sc *scope, ctx *fnCtx) {
	for i, p := range params {
		l := &local{name: p.Name, typ: s.fn.Params[i], slot: p.Slot, ctx: ctx, pos: p.Pos}
		if !p.Slot.Discard {
			sc.names[p.Name] = l
		}
	}
	// Defaults run in parameter order after the arguments (section 6), so a default may not
	// read its own parameter or a later one that has a default: when that one is omitted too,
	// it is not set yet.
	saved := c.fn
	c.fn = ctx
	for _, p := range params {
		if p.Default != nil && !p.Slot.Discard {
			sc.names[p.Name].unset = true
		}
	}
	for i, p := range params {
		if p.Default != nil {
			c.check(p.Default, sc, s.fn.Params[i])
			if !p.Slot.Discard {
				sc.names[p.Name].unset = false
			}
		}
	}
	c.fn = saved
}

// inferBody checks a function body against its declared return type, or infers one from the
// body's value and its returns.
func (c *Checker) inferBody(body *ast.Block, sc *scope, ctx *fnCtx) types.Type {
	if ctx.declared != nil {
		c.block(body, sc, ctx.declared)
		return ctx.declared
	}
	t := c.block(body, sc, nil)
	all := append([]types.Type{t}, ctx.returns...)
	u, err := types.MakeUnion(all)
	if err != nil {
		c.errorf(body.Pos, "the returns of %s do not form a valid type: %s; annotate the return type", ctx.name, err)
		return types.Invalid{}
	}
	if _, _, isResult := resultParts(types.Without(u, types.Null{})); !isResult && !types.IsInvalid(u) {
		for _, x := range ctx.tries {
			c.errorf(x.Pos, "try needs an enclosing function that returns a Result, but %s returns %s", ctx.name, types.String(u))
		}
	}
	return u
}

// returnTypeOf returns a function's return type, inferring it on demand; a function whose
// inference is in progress is a recursion that needs an annotation.
func (c *Checker) returnTypeOf(ent *entity, pos ast.Pos) types.Type {
	s := ent.sig
	if s.fn.Return != nil {
		return s.fn.Return
	}
	if ent.fn == nil {
		return types.Invalid{}
	}
	if s.inferring || ent.fn.state == 1 {
		c.errorf(pos, "%s is recursive; state its return type", s.fullName())
		return types.Invalid{}
	}
	c.checkFunc(ent.fn)
	if s.fn.Return == nil {
		return types.Invalid{}
	}
	return s.fn.Return
}
