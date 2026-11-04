package interp

import (
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// Type expressions (section 4, Types as values; section 12, Expressions). A tag keeps its
// declaration, and the runtime turns that syntax into the checker's own types.Type the first
// time a type expression reaches it, so that equality, unions, substitution, and assignability
// are the checker's code. Every resolution runs under Instance.typeMu, since it may build the
// declarations of every type the expression reaches.

// typeBody is what a tag keeps of its declaration: the syntax, the environment its names
// resolve in, and the declaration built from it when reflection first reaches the tag.
type typeBody struct {
	decl *ast.TypeDecl // nil for a tag the library made that the prelude does not declare
	env  *Env          // nil for a prelude or host type, whose names resolve through the prelude
	t    *types.Decl   // cached before its fields, alias body, or base are resolved
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
	v := &types.TypeValue{T: g.resolveType(e.Type, env, nil)}
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
		if t, ok := g.builtinType(x, env, params); ok {
			return t
		}
	}
	var v value.Value
	if x.Head != nil {
		v = g.ident(x.Head, env)
	} else {
		var ok bool
		if v, ok = g.in.cfg.Prelude[name]; !ok {
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

// builtinType resolves a name no declaration binds: a primitive, null, void, never, Array, Map,
// Set, Bytes, or Html (section 12, Expressions).
func (g *Goroutine) builtinType(x *ast.TypeName, env *Env, params map[string]types.Type) (types.Type, bool) {
	name := x.Path[0]
	arity := func(n int) {
		if len(x.Args) != n {
			g.panicf(x.Pos, "%s takes %d type arguments, got %d", name, n, len(x.Args))
		}
	}
	if k, ok := types.BuiltinKinds[name]; ok {
		arity(0)
		return types.Prim{Kind: k}, true
	}
	switch name {
	case "null":
		arity(0)
		return types.Null{}, true
	case "void":
		arity(0)
		return types.Void{}, true
	case "never":
		arity(0)
		return types.Never{}, true
	case "Array":
		arity(1)
		return types.Array{Elem: g.resolveType(x.Args[0], env, params)}, true
	case "Map":
		arity(2)
		return types.Map{Key: g.resolveType(x.Args[0], env, params), Val: g.resolveType(x.Args[1], env, params)}, true
	case "Set":
		arity(1)
		return types.Set{Elem: g.resolveType(x.Args[0], env, params)}, true
	case "Bytes", "Html":
		arity(0)
		return types.Named{Decl: g.in.builtinDecl(name)}, true
	}
	return nil, false
}

// builtinDecl is the declaration of Bytes or Html, opaque types whose values carry a kind and no
// tag; the prelude declares them without a body.
func (in *Instance) builtinDecl(name string) *types.Decl {
	if d, ok := in.builtinDecls[name]; ok {
		return d
	}
	if in.builtinDecls == nil {
		in.builtinDecls = map[string]*types.Decl{}
	}
	d := &types.Decl{Name: name, Kind: types.OpaqueDecl, Tag: value.KindBytes}
	if name == "Html" {
		d.Tag = value.KindHtml
	}
	in.builtinDecls[name] = d
	return d
}

// preludeType finds the prelude's declaration of a type by its qualified name, for a tag the
// library made in Go, whose body exists only in prelude.d.ant.
func (in *Instance) preludeType(name string) *ast.TypeDecl {
	if in.preludeTypes == nil {
		in.preludeTypes = map[string]*ast.TypeDecl{}
		if f, err := prelude.Decls(); err == nil {
			for _, d := range f.Decls {
				if td, ok := d.(*ast.TypeDecl); ok {
					n := td.Name
					if td.Owner != "" {
						n = td.Owner + "." + td.Name
					}
					in.preludeTypes[n] = td
				}
			}
		}
	}
	return in.preludeTypes[name]
}

// declOf is the declaration of a tag's type, built on the first call and kept on the tag. It is
// cached before its fields, alias body, or base are resolved, so that a recursive type finds
// itself, as the checker's resolveTypeOnce does.
func (g *Goroutine) declOf(tag *value.Tag) *types.Decl {
	body, _ := tag.Body.(*typeBody)
	if body == nil {
		body = &typeBody{decl: g.in.preludeType(tag.Name)}
		tag.Body = body
	}
	if body.t != nil {
		return body.t
	}
	d := &types.Decl{Name: tag.Name, Tag: tag}
	body.t = d
	defer func() {
		// A panic leaves the declaration half built; the next type expression starts over.
		if r := recover(); r != nil {
			body.t = nil
			panic(r)
		}
	}()
	var params map[string]types.Type
	if body.decl != nil {
		for _, tp := range body.decl.TypeParams {
			pd := &types.ParamDecl{Name: tp.Name}
			d.Params = append(d.Params, pd)
			if params == nil {
				params = map[string]types.Type{}
			}
			params[tp.Name] = types.Param{Decl: pd}
		}
	}
	switch tag.Kind {
	case value.RecordTag:
		d.Kind = types.RecordDecl
		if body.decl != nil {
			if rt, ok := body.decl.Body.(*ast.RecordType); ok {
				for _, f := range rt.Fields {
					d.Fields = append(d.Fields, types.Field{Name: f.Name, Type: g.resolveType(f.Type, body.env, params), Default: f.Default != nil})
				}
			}
		}
	case value.PrimitiveTag:
		d.Kind = types.PrimitiveDecl
		d.Base = tag.Base
	case value.AliasTag:
		d.Kind = types.AliasDecl
		if body.decl != nil && body.decl.Body != nil {
			d.Body = g.resolveType(body.decl.Body, body.env, params)
		}
	case value.SyncTag:
		d.Kind = types.OpaqueDecl
	case value.TimeTag, value.RegexTag:
		d.Kind = types.TaggedOpaqueDecl
	}
	return d
}
