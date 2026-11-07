package interp

import (
	"strings"
	"sync"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// Type expressions (section 4, Types as values; section 12, Expressions). A tag keeps its
// declaration, and the runtime turns that syntax into the checker's own types.Type the first
// time a type expression reaches it, so that equality, unions, substitution, and assignability
// are the checker's code. Every resolution runs under Instance.typeMu, since it may build the
// declarations of every type the expression reaches. The declarations live on the Instance, not
// on the tags: a tag the library made is shared by every Instance in the process.

// typeBody is what a tag keeps of its declaration: the syntax and the environment its names
// resolve in. A tag the library made in Go keeps none; its declaration is the prelude's.
type typeBody struct {
	decl *ast.TypeDecl
	env  *Env // nil for a host type, whose names resolve through the prelude
}

// typeExpr evaluates a type expression once per node: names in a type resolve to the same tags
// at every evaluation (section 12, Names and scopes).
func (g *Goroutine) typeExpr(e *ast.TypeExpr, env *Env) value.Value {
	if v, ok := g.in.typeExprs.Load(e); ok {
		return v.(value.Value)
	}
	g.in.typeMu.Lock()
	defer g.in.typeMu.Unlock()
	if v, ok := g.in.typeExprs.Load(e); ok {
		return v.(value.Value)
	}
	defer func() {
		// A panic leaves every declaration this expression began half built, so they are all
		// forgotten, and the next expression that reaches one starts over.
		if r := recover(); r != nil {
			for _, tag := range g.in.newDecls {
				delete(g.in.decls, tag)
			}
			g.in.newDecls, g.in.newDeclSCC = nil, nil
			panic(r)
		}
		g.in.newDecls, g.in.newDeclSCC = nil, nil
	}()
	g.in.typePos = e.Pos
	v := &types.TypeValue{T: g.resolveType(e.Type, env, nil)}
	// The instantiations the expression and the declarations it reached write are checked once
	// those declarations are complete, as the checker does at each use (section 4, Unions).
	// Every declaration was checked as it was completed (declOf), but one completed inside a
	// cycle saw a member of it half built, so all are checked again now that the cycle is whole.
	written := []types.Type{v.T}
	for _, tag := range g.in.newDecls {
		d := g.in.decls[tag]
		g.checkDecl(d, g.in.component(d))
		for _, p := range d.Params {
			if p.Bound != nil {
				written = append(written, p.Bound)
			}
		}
		for _, f := range d.Fields {
			written = append(written, f.Type)
		}
		if d.Body != nil {
			written = append(written, d.Body)
		}
	}
	// The declarations are complete, so what one check learns holds for the next.
	var memo types.ArgsMemo
	for _, t := range written {
		if err := memo.CheckWritten(t); err != nil {
			g.panicf(e.Pos, "%s", err)
		}
	}
	// The bounds of a type's parameters hold wherever it is written, in the expression and in
	// the declarations it reached (section 4, Generics); they are read once every declaration
	// is complete, since a record argument may be one still being built.
	for _, t := range written {
		if err := types.CheckWrittenBounds(t); err != nil {
			g.panicf(e.Pos, "%s", err)
		}
	}
	g.in.typeExprs.Store(e, v)
	return v
}

// resolveType turns type syntax into a type, under typeMu. params binds the type parameters of
// the declaration whose body is being resolved. Where the checker would report an error, it
// panics: an unchecked program reaches here with a name that is not a type.
func (g *Goroutine) resolveType(t ast.Type, env *Env, params map[string]types.Type) types.Type {
	switch x := t.(type) {
	case *ast.TypeName:
		return g.resolveTypeName(x, env, params)
	case *ast.RecordType:
		fields := make([]types.Field, len(x.Fields))
		for i, f := range x.Fields {
			fields[i] = types.Field{Name: f.Name, Type: g.resolveType(f.Type, env, params)}
		}
		return types.Record{Fields: fields}
	case *ast.FuncType:
		f := types.Func{Required: len(x.Params)}
		for _, p := range x.Params {
			f.Params = append(f.Params, g.resolveType(p.Type, env, params))
		}
		f.Return = g.resolveType(x.Return, env, params)
		return f
	case *ast.UnionType:
		members := make([]types.Type, len(x.Members))
		for i, m := range x.Members {
			members[i] = g.resolveType(m, env, params)
			// Forming the union expands an alias member, which needs its arguments valid: an
			// unchecked program may give it ones the checker rejects (section 4, Unions).
			if err := types.CheckWritten(members[i]); err != nil {
				g.panicf(x.Pos, "%s", err)
			}
		}
		u, err := types.MakeUnion(members)
		if err != nil {
			g.panicf(x.Pos, "%s", err)
		}
		return u
	}
	g.panicf(g.pos, "internal error: a type expression holds %T", t)
	return nil
}

// resolveTypeName resolves a named type: a type parameter, a built-in name, or a path that
// evaluates to a tag or a constructor, as the target of is does. A name the resolver left
// without a head is a built-in, a type parameter, or a name of the prelude's own declarations,
// which the prelude table binds.
func (g *Goroutine) resolveTypeName(x *ast.TypeName, env *Env, params map[string]types.Type) types.Type {
	name := x.Path[0]
	if len(x.Path) == 1 && x.Head == nil {
		if p, ok := params[name]; ok {
			if len(x.Args) > 0 {
				g.panicf(x.Pos, "the type parameter %s takes no type arguments", name)
			}
			return p
		}
		if n, ok := types.BuiltinArity(name); ok {
			if len(x.Args) != n {
				g.panicf(x.Pos, "%s takes %d type arguments, got %d", name, n, len(x.Args))
			}
			args := make([]types.Type, len(x.Args))
			for i, a := range x.Args {
				args[i] = g.resolveType(a, env, params)
			}
			return types.Builtin(name, args, g.in.builtinDecl)
		}
	}
	var v value.Value
	if x.Head != nil {
		v = g.typeHead(x, env)
	} else {
		var ok bool
		if v, ok = g.in.cfg.Prelude[name]; !ok {
			if env != nil {
				// A program's own name left without a head is a type parameter of its function.
				g.panicf(x.Pos, "%s is a type parameter of the enclosing function, which nothing is bound to at run time", name)
			}
			g.panicf(x.Pos, "unknown type %s", name)
		}
	}
	for _, part := range x.Path[1:] {
		v = g.field(v, part, x.Pos)
	}
	var tag *value.Tag
	switch t := v.(type) {
	case *value.Tag:
		tag = t
	case *value.Constructor:
		tag = t.Tag
	default:
		g.panicf(x.Pos, "%s is not a type, so type(...) cannot name it; it is %s", strings.Join(x.Path, "."), value.Describe(v))
	}
	d := g.declOf(tag)
	args := make([]types.Type, len(x.Args))
	for i, a := range x.Args {
		args[i] = g.resolveType(a, env, params)
	}
	if len(args) != len(d.Params) {
		g.panicf(x.Pos, "%s takes %d type arguments, got %d", d.Name, len(d.Params), len(args))
	}
	return types.Named{Decl: d, Args: args}
}

// builtinDecl is the declaration of Bytes or Html, opaque types whose values carry a kind and no
// tag; the prelude declares them without a body. Made under typeMu, once per Instance.
func (in *Instance) builtinDecl(name string) *types.Decl {
	if d, ok := in.builtinDecls[name]; ok {
		return d
	}
	if in.builtinDecls == nil {
		in.builtinDecls = map[string]*types.Decl{}
	}
	d := &types.Decl{Name: name, Kind: types.KindOpaqueDecl, Base: value.KindBytes}
	if name == "Html" {
		d.Base = value.KindHtml
	}
	in.builtinDecls[name] = d
	return d
}

// preludeTypes is the prelude's type declarations by qualified name, for a tag the library made
// in Go, whose body exists only in prelude.d.ant. The prelude never changes, so the table is
// built once per process.
var preludeTypes = sync.OnceValue(func() map[string]*ast.TypeDecl {
	table := map[string]*ast.TypeDecl{}
	f, err := prelude.Decls()
	if err != nil {
		return table
	}
	for _, d := range f.Decls {
		if td, ok := d.(*ast.TypeDecl); ok {
			n := td.Name
			if td.Owner != "" {
				n = td.Owner + "." + td.Name
			}
			table[n] = td
		}
	}
	return table
})

// declOf is the declaration of a tag's type, built the first time a type expression reaches the
// tag and kept by the Instance. It is cached before its fields, alias body, or base are
// resolved, so that a recursive type finds itself, as the checker's resolveTypeOnce does.
func (g *Goroutine) declOf(tag *value.Tag) *types.Decl {
	in := g.in
	if d, ok := in.decls[tag]; ok {
		return d
	}
	body, _ := tag.Body.(*typeBody)
	if body == nil {
		td := preludeTypes()[tag.Name]
		if td == nil {
			g.panicf(g.pos, "internal error: the prelude does not declare the type %s", tag.Name)
		}
		body = &typeBody{decl: td}
	}
	d := &types.Decl{Name: tag.Name, Tag: tag, Pos: body.decl.Pos}
	if in.decls == nil {
		in.decls = map[*value.Tag]*types.Decl{}
	}
	in.decls[tag] = d
	in.newDecls = append(in.newDecls, tag)
	if in.newDeclSCC == nil {
		in.newDeclSCC = map[*types.Decl]*declSCC{}
	}
	in.newDeclSCC[d] = &declSCC{}
	var params map[string]types.Type
	for _, tp := range body.decl.TypeParams {
		pd := &types.ParamDecl{Name: tp.Name}
		d.Params = append(d.Params, pd)
		if params == nil {
			params = map[string]types.Type{}
		}
		params[tp.Name] = types.Param{Decl: pd}
	}
	for i, tp := range body.decl.TypeParams {
		if tp.Bound != nil {
			d.Params[i].Bound = g.resolveType(tp.Bound, body.env, params)
		}
	}
	switch tag.Kind {
	case value.RecordTag:
		d.Kind = types.RecordDecl
		if rt, ok := body.decl.Body.(*ast.RecordType); ok {
			for _, f := range rt.Fields {
				d.Fields = append(d.Fields, types.Field{Name: f.Name, Type: g.resolveType(f.Type, body.env, params), Default: f.Default != nil})
			}
		}
	case value.PrimitiveTag:
		d.Kind = types.PrimitiveDecl
		d.Base = tag.Base
	case value.AliasTag:
		d.Kind = types.AliasDecl
		if body.decl.Body != nil {
			d.Body = g.resolveType(body.decl.Body, body.env, params)
		}
	case value.SyncTag:
		d.Kind = types.OpaqueDecl
	case value.TimeTag, value.RegexTag:
		d.Kind = types.TaggedOpaqueDecl
	}
	in.newDeclSCC[d].complete = true
	g.checkDecl(d, in.component(d))
	return d
}

// declSCC is what a type expression knows of the strongly connected component of a declaration
// it began, in the graph of the declarations' bodies and fields.
type declSCC struct {
	complete bool
	comp     map[*types.Decl]bool // the component, once it and every one it reaches are complete
}

// component returns the declarations of d's strongly connected component as the declarations
// built so far name each other, which a check of d need follow alone (types.Components). Only
// the declarations the current type expression began are searched: an earlier one is complete,
// and so is everything it reaches. A component that is complete and reaches only complete ones
// is kept, so a chain of declarations is searched once, not once from each link.
func (in *Instance) component(d *types.Decl) func(*types.Decl) bool {
	if c := in.newDeclSCC[d]; c == nil {
		return func(x *types.Decl) bool { return x == d }
	} else if c.comp != nil {
		return func(x *types.Decl) bool { return c.comp[x] }
	}
	var root map[*types.Decl]bool
	settled := func(x *types.Decl) bool {
		c := in.newDeclSCC[x]
		return c == nil || c.comp != nil
	}
	types.Components(d, types.DeclRefs, settled, func(comp []*types.Decl, _ bool) {
		root = make(map[*types.Decl]bool, len(comp))
		for _, x := range comp {
			root[x] = true
		}
		settle := true
		for _, x := range comp {
			settle = settle && in.newDeclSCC[x].complete
			types.DeclRefs(x, func(y *types.Decl) {
				settle = settle && (root[y] || settled(y))
			})
		}
		if settle {
			for _, x := range comp {
				in.newDeclSCC[x].comp = root
			}
		}
	})
	return func(x *types.Decl) bool { return root[x] }
}

// writtenAsName reports whether the alias d's body is written as one type name, as in
// type B = A. Only that spelling mints a second tag for A (section 4, Records); a union such as
// D | D deduplicates to D when it is formed, but it is not written as D, and the checker accepts it.
func writtenAsName(d *types.Decl) bool {
	b, ok := d.Tag.Body.(*typeBody)
	if !ok || b == nil {
		return false
	}
	_, ok = b.decl.Body.(*ast.TypeName)
	return ok
}

// checkDecl panics where the checker rejects the declaration d itself (section 4, Type
// declarations). declOf runs it as soon as d is complete, as the checker does, before a union or
// an instantiation that names d is formed: an alias that reaches itself through unions alone, a
// generic alias on a cycle, or a generic record type that refers to itself with other arguments,
// would make expanding it endless.
// A reference of d to itself passes through the declarations of its component alone, which in
// reports.
func (g *Goroutine) checkDecl(d *types.Decl, in func(*types.Decl) bool) {
	pos := g.in.typePos
	if n, ok := d.Body.(types.Named); ok && d.Kind == types.AliasDecl && writtenAsName(d) && (n.Decl.Kind == types.RecordDecl || n.Decl.Kind == types.PrimitiveDecl) {
		g.panicf(pos, "type %s = %s would mint a second tag for one type; declare the fields or the base type again", d.Name, n.Decl.Name)
	}
	if _, msg, cyclic := types.AliasCycleErrorIn(d, in); cyclic {
		g.panicf(pos, "%s", msg)
	}
	if _, msg, cyclic := types.GenericRecursiveIn(d, in); cyclic {
		g.panicf(pos, "%s", msg)
	}
	if msg := types.SelfArgsIn(d, in); msg != "" {
		g.panicf(pos, "%s", msg)
	}
	if msg := types.UnionSelfIn(d, in); msg != "" {
		g.panicf(pos, "%s", msg)
	}
	for _, p := range d.Params {
		if p.Bound == nil {
			continue
		}
		if q, ok := types.BareBound(p.Bound); ok {
			g.panicf(pos, "%s", types.BareBoundError(p, q))
		}
	}
}
