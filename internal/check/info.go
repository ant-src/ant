package check

import (
	"sort"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/types"
)

// Info is what the checker learned about a program beyond its errors, for an editor: the type of
// every expression, what every name refers to, and the signature every call applies. It is filled
// when Config.Info is set, and read once Program has returned.
type Info struct {
	// Types holds the type of each expression as it was checked: for a variable, the type at
	// that use, narrowed (section 7). An expression checked more than once, as a lambda body may
	// be, keeps the last type.
	Types map[ast.Expr]types.Type
	// Objects maps each identifier to what it refers to.
	Objects map[*ast.Ident]*Object
	// Members maps each field access to the member or field it reads.
	Members map[*ast.Field]*Object
	// TypeNames maps each type name to what each part of its path refers to: a namespace or a
	// module, then the type or type parameter.
	TypeNames map[*ast.TypeName][]*Object
	// Decls maps each declaring node to its object: a *ast.FuncDecl, a *ast.TypeDecl, a
	// *ast.NameTarget, a *ast.Param, the *ast.Lambda whose x it is, a *ast.Import that binds a
	// namespace, or a *ast.ImportName.
	Decls map[ast.Node]*Object
	// Calls maps each call to the signature it applied.
	Calls map[*ast.Call]*CallInfo
	// Lambdas maps each lambda to the type of its x, or nil for a lambda of arity zero.
	Lambdas map[*ast.Lambda]types.Type
	// Entries maps each named entry of a record literal to the field it sets.
	Entries map[*ast.RecordEntry]*Object
	// Prelude lists the prelude's names, sorted: functions, types, and namespaces.
	Prelude []*Object
	// Modules maps each module's path to its names.
	Modules map[string]*ModuleInfo

	objects map[*entity]*Object
	locals  map[*local]*Object
	fields  map[fieldKey]*Object
	anon    map[string]*Object // fields of anonymous record types, by the type's text and the name
	tparams map[*types.ParamDecl]*Object
}

// ModuleInfo lists a module's top-level names and exports.
type ModuleInfo struct {
	Path    string
	Names   []*Object
	Exports []*Object
}

// CallInfo is what a call applied: the function type, the parameter names and which have
// defaults for a declared function, the callee when it is a named one, and which parameter each
// argument fills.
type CallInfo struct {
	Fn       types.Func
	Names    []string
	Defaults []bool
	Callee   *Object
	// Slots gives, for each argument in order, the index of the parameter it fills, or -1.
	Slots []int
	// Pipe is set when the call is the right side of a pipe, whose first argument is the left
	// side (section 5, Pipes).
	Pipe bool
}

// Kind says what an Object is.
type Kind int

const (
	Var       Kind = iota // a variable, top-level or local, or a lambda's x
	Param                 // a parameter
	Func                  // a function, declared or of a namespace or type
	TypeKind              // a declared type
	TypeParam             // a type parameter
	Namespace             // a prelude or host namespace
	Module                // an imported module
	Const                 // a namespace constant
	Field                 // a field of a record type
)

func (k Kind) String() string {
	return [...]string{"variable", "parameter", "function", "type", "type parameter", "namespace", "module", "constant", "field"}[k]
}

// Object is a named thing: what an identifier, a member, or a type name refers to.
type Object struct {
	Name string
	Kind Kind
	// Decl is the node that declares it: a *ast.FuncDecl, *ast.TypeDecl, *ast.VarDecl of the
	// prelude or a module, a *ast.NameTarget, a *ast.Param, a *ast.FieldType, a *ast.Lambda for
	// its x, or nil for a namespace with no declaration, a type parameter, or the field of an
	// anonymous record type.
	Decl ast.Node
	// Path is the module the declaration is in: "prelude.d.ant" for the prelude, the host
	// namespace's name for a host declaration, or "" when there is no declaration.
	Path     string
	Exported bool
	// Owner names the type or namespace a function, type, or constant belongs to, or "".
	Owner string

	ent   *entity
	loc   *local
	typ   types.Type
	param *types.ParamDecl
	info  *Info
}

type fieldKey struct {
	decl *types.Decl
	name string
}

// NewInfo returns an empty Info for Config.
func NewInfo() *Info {
	return &Info{
		Types: map[ast.Expr]types.Type{}, Objects: map[*ast.Ident]*Object{}, Members: map[*ast.Field]*Object{},
		TypeNames: map[*ast.TypeName][]*Object{}, Decls: map[ast.Node]*Object{}, Calls: map[*ast.Call]*CallInfo{},
		Lambdas: map[*ast.Lambda]types.Type{}, Entries: map[*ast.RecordEntry]*Object{}, Modules: map[string]*ModuleInfo{},
		objects: map[*entity]*Object{}, locals: map[*local]*Object{}, fields: map[fieldKey]*Object{},
		anon: map[string]*Object{}, tparams: map[*types.ParamDecl]*Object{},
	}
}

// Type is the object's type: a variable's or parameter's declared type, a function's type with
// its return type inferred, a type's own, a type parameter's, a constant's, or a field's; nil
// for a namespace or a module.
func (o *Object) Type() types.Type {
	switch {
	case o.loc != nil:
		return o.loc.typ
	case o.param != nil:
		return types.Param{Decl: o.param}
	case o.ent != nil:
		switch e := o.ent; e.kind {
		case entVar:
			if e.v != nil && e.v.typ != nil {
				return e.v.typ
			}
			return types.Invalid{}
		case entConst:
			return e.typ
		case entFunc:
			if e.sig == nil {
				return nil
			}
			fn := e.sig.fn
			if fn.Return == nil {
				fn.Return = types.Invalid{}
			}
			return fn
		case entType:
			var args []types.Type
			for _, p := range e.decl.Params {
				args = append(args, types.Param{Decl: p})
			}
			return types.Named{Decl: e.decl, Args: args}
		}
		return nil
	}
	return o.typ
}

// TypeParams lists a generic function's or type's type parameters, or nil.
func (o *Object) TypeParams() []*types.ParamDecl {
	if o.ent == nil {
		return nil
	}
	switch o.ent.kind {
	case entFunc:
		if o.ent.sig != nil {
			return o.ent.sig.tparams
		}
	case entType:
		return o.ent.decl.Params
	}
	return nil
}

// Defaults reports, for a declared function, which parameters have a default.
func (o *Object) Defaults() []bool {
	if o.ent != nil && o.ent.sig != nil {
		return o.ent.sig.defaults
	}
	return nil
}

// Doc is the declaration's doc comment, or "".
func (o *Object) Doc() string {
	switch d := o.declaration().(type) {
	case *ast.FuncDecl:
		return d.Doc
	case *ast.TypeDecl:
		return d.Doc
	case *ast.VarDecl:
		return d.Doc
	}
	return ""
}

// Sig is the declaration's signature as written (doc.md section 1), or "".
func (o *Object) Sig() string {
	switch d := o.declaration().(type) {
	case *ast.FuncDecl:
		return d.Sig
	case *ast.TypeDecl:
		return d.Sig
	case *ast.VarDecl:
		return d.Sig
	}
	return ""
}

// declaration is the declaration that carries the doc and signature: for a top-level var, the
// var declaration its name target belongs to.
func (o *Object) declaration() ast.Node {
	if o.ent != nil && o.ent.kind == entVar && o.ent.v != nil {
		return o.ent.v.decl
	}
	return o.Decl
}

// TypeDecl is the declaration of a type, with its fields and kind, or nil.
func (o *Object) TypeDecl() *types.Decl {
	if o.ent != nil && o.ent.kind == entType {
		return o.ent.decl
	}
	return nil
}

// Members lists what a namespace, a module, or a type holds: its functions, types, and
// constants, and for a record type its fields, sorted by name. Only the exports of a module.
func (o *Object) Members() []*Object {
	if o.ent == nil || o.info == nil {
		return nil
	}
	var out []*Object
	switch o.ent.kind {
	case entModule:
		for _, e := range o.ent.mod.exports {
			out = append(out, o.info.object(e))
		}
	case entNamespace, entType:
		for _, e := range o.ent.members {
			out = append(out, o.info.object(e))
		}
		if o.ent.kind == entType && o.ent.decl.Kind == types.RecordDecl {
			for _, f := range o.ent.decl.Fields {
				out = append(out, o.info.field(o.ent.decl, f))
			}
		}
	}
	sortObjects(out)
	return out
}

func sortObjects(objs []*Object) {
	sort.Slice(objs, func(i, j int) bool { return objs[i].Name < objs[j].Name })
}

// ---- recording

// object is the Object of an entity, made once.
func (in *Info) object(e *entity) *Object {
	if in == nil || e == nil {
		return nil
	}
	if o, ok := in.objects[e]; ok {
		return o
	}
	o := &Object{Name: e.name, ent: e, info: in, Exported: e.exported}
	switch e.kind {
	case entType:
		o.Kind = TypeKind
	case entNamespace:
		o.Kind = Namespace
	case entFunc:
		o.Kind = Func
	case entVar:
		o.Kind = Var
	case entConst:
		o.Kind = Const
	case entModule:
		o.Kind = Module
		if e.mod != nil {
			o.Path = e.mod.path
		}
	}
	in.objects[e] = o
	return o
}

// declared fills in where an entity is declared.
func (in *Info) declared(e *entity, decl ast.Node, path, owner string) {
	o := in.object(e)
	if o == nil {
		return
	}
	o.Decl, o.Path, o.Owner = decl, path, owner
	if decl != nil {
		in.Decls[decl] = o
	}
}

// localObject is the Object of a local, made once; path is its module.
func (in *Info) localObject(l *local, path string) *Object {
	if in == nil || l == nil {
		return nil
	}
	if o, ok := in.locals[l]; ok {
		return o
	}
	o := &Object{Name: l.name, Kind: Var, loc: l, Decl: l.node, Path: path, info: in}
	if _, isParam := l.node.(*ast.Param); isParam {
		o.Kind = Param
	}
	in.locals[l] = o
	if l.node != nil {
		in.Decls[l.node] = o
	}
	return o
}

// field is the Object of a named record type's field, made once.
func (in *Info) field(decl *types.Decl, f types.Field) *Object {
	if in == nil {
		return nil
	}
	key := fieldKey{decl, f.Name}
	if o, ok := in.fields[key]; ok {
		return o
	}
	o := &Object{Name: f.Name, Kind: Field, typ: f.Type, Owner: decl.Name, Path: decl.Path, info: in}
	in.fields[key] = o
	return o
}

// fieldDecl points a field object at its declaration once the declaring type is known.
func (in *Info) fieldDecl(decl *types.Decl, td *ast.TypeDecl) {
	if in == nil {
		return
	}
	rt, ok := td.Body.(*ast.RecordType)
	if !ok {
		return
	}
	for i := range rt.Fields {
		f := &rt.Fields[i]
		for _, tf := range decl.Fields {
			if tf.Name == f.Name {
				o := in.field(decl, tf)
				o.Decl = f
				in.Decls[f] = o
			}
		}
	}
}

// typeParamObject is the Object of a type parameter, made once.
func (in *Info) typeParamObject(p *types.ParamDecl) *Object {
	if in == nil {
		return nil
	}
	if o, ok := in.tparams[p]; ok {
		return o
	}
	o := &Object{Name: p.Name, Kind: TypeParam, param: p, info: in}
	in.tparams[p] = o
	return o
}

// typeParam records where a type parameter is declared.
func (in *Info) typeParam(p *types.ParamDecl, decl *ast.TypeParam, path string) {
	if in == nil {
		return
	}
	o := in.typeParamObject(p)
	o.Decl, o.Path = decl, path
	in.Decls[decl] = o
}

// builtinType is the Object of a built-in type name with no declaration: a primitive, null,
// void, never, Array, Map, or Set.
func (in *Info) builtinType(name string, t types.Type) *Object {
	return &Object{Name: name, Kind: TypeKind, typ: t, info: in}
}

// alias records that node, an import's name, stands for the object of e.
func (in *Info) alias(node ast.Node, e *entity) {
	if in != nil && e != nil {
		in.Decls[node] = in.object(e)
	}
}

// fieldOf is the Object of the field name of a type: of a named record type's declaration, of
// the member of a union that declares it, or a nameless one for an anonymous record.
func (in *Info) fieldOf(recv types.Type, name string) *Object {
	if in == nil {
		return nil
	}
	for _, m := range types.Members(recv) {
		switch r := types.Expand(m).(type) {
		case types.Named:
			if r.Decl.Kind != types.RecordDecl {
				continue
			}
			for _, f := range r.FieldsOf() {
				if f.Name == name {
					return in.field(r.Decl, f)
				}
			}
		case types.Record:
			for _, f := range r.Fields {
				if f.Name == name {
					// an anonymous record type is its shape (section 4), so one object per
					// shape and name
					key := types.String(r) + "\x00" + name
					if o, ok := in.anon[key]; ok {
						return o
					}
					o := &Object{Name: name, Kind: Field, typ: f.Type, info: in}
					in.anon[key] = o
					return o
				}
			}
		}
	}
	return nil
}

func (in *Info) recordType(e ast.Expr, t types.Type) {
	if in != nil && e != nil && t != nil {
		in.Types[e] = t
	}
}

func (in *Info) recordIdent(id *ast.Ident, o *Object) {
	if in != nil && o != nil {
		in.Objects[id] = o
	}
}

func (in *Info) recordMember(f *ast.Field, o *Object) {
	if in != nil && o != nil {
		in.Members[f] = o
	}
}

func (in *Info) recordCall(x *ast.Call, ci *CallInfo) {
	if in != nil {
		ci.Pipe = x.Pipe
		in.Calls[x] = ci
	}
}

// finish fills the prelude and module listings once the check is done.
func (in *Info) finish(c *Checker) {
	if in == nil {
		return
	}
	for _, e := range c.prelude {
		in.Prelude = append(in.Prelude, in.object(e))
	}
	sortObjects(in.Prelude)
	for _, mi := range c.mods {
		m := &ModuleInfo{Path: mi.path}
		for _, e := range mi.names {
			m.Names = append(m.Names, in.object(e))
		}
		for _, e := range mi.exports {
			m.Exports = append(m.Exports, in.object(e))
		}
		sortObjects(m.Names)
		sortObjects(m.Exports)
		in.Modules[mi.path] = m
	}
}
