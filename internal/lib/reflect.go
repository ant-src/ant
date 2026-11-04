package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// The Reflect namespace (section 11, Reflect): type values, shapes, and reflected values. The
// opaque types are tags that is cannot test, as the sync types are; the shapes are record tags
// whose values the functions build.
var (
	ReflectTypeTag       = value.NewTag("Reflect.Type", value.SyncTag, nil)
	ReflectErasedTypeTag = value.NewTag("Reflect.ErasedType", value.SyncTag, nil)
	ReflectDeclTag       = value.NewTag("Reflect.Decl", value.SyncTag, nil)
	ReflectValueTag      = value.NewTag("Reflect.Value", value.SyncTag, nil)

	ShapeNullTag   = value.NewTag("Reflect.Null", value.RecordTag, nil)
	ShapeVoidTag   = value.NewTag("Reflect.Void", value.RecordTag, nil)
	ShapeNeverTag  = value.NewTag("Reflect.Never", value.RecordTag, nil)
	ShapeBoolTag   = value.NewTag("Reflect.Bool", value.RecordTag, nil)
	ShapeIntTag    = value.NewTag("Reflect.Int", value.RecordTag, []string{"bits", "signed"})
	ShapeFloatTag  = value.NewTag("Reflect.Float", value.RecordTag, []string{"bits"})
	ShapeStringTag = value.NewTag("Reflect.String", value.RecordTag, nil)
	ShapeArrayTag  = value.NewTag("Reflect.Array", value.RecordTag, []string{"element"})
	ShapeMapTag    = value.NewTag("Reflect.Map", value.RecordTag, []string{"key", "value"})
	ShapeSetTag    = value.NewTag("Reflect.Set", value.RecordTag, []string{"member"})
	ShapeRecordTag = value.NewTag("Reflect.Record", value.RecordTag, []string{"name", "decl", "args", "fields"})
	ShapeFieldTag  = value.NewTag("Reflect.Field", value.RecordTag, []string{"name", "type", "hasDefault"})
	ShapeNamedTag  = value.NewTag("Reflect.Named", value.RecordTag, []string{"name", "decl", "args", "base"})
	ShapeUnionTag  = value.NewTag("Reflect.Union", value.RecordTag, []string{"members"})
	ShapeFuncTag   = value.NewTag("Reflect.Func", value.RecordTag, []string{"params", "result"})
	ShapeOpaqueTag = value.NewTag("Reflect.Opaque", value.RecordTag, []string{"name", "decl", "args"})
)

// typeArg takes a type value, a Reflect.Type[T] or a Reflect.ErasedType.
func typeArg(g *interp.Goroutine, who string, v value.Value) *types.TypeValue {
	return typedArg[*types.TypeValue](g, who, "a type value", v)
}

func init() {
	register("Reflect", func() value.Value {
		notYet := func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			g.Panic("%s is not implemented yet", who)
			return nil
		}
		// A type builder combines closed types, so the union it forms cannot fail.
		nullable := func(g *interp.Goroutine, who string, t types.Type) types.Type {
			u, err := types.MakeUnion([]types.Type{t, types.Null{}})
			if err != nil {
				g.Panic("%s: %s", who, err)
			}
			return u
		}
		return namespace("Reflect",
			constant("Type", ReflectTypeTag),
			constant("ErasedType", ReflectErasedTypeTag),
			constant("Decl", ReflectDeclTag),
			constant("Value", ReflectValueTag),
			constant("Shape", value.NewTag("Reflect.Shape", value.AliasTag, nil)),
			constant("Null", ShapeNullTag),
			constant("Void", ShapeVoidTag),
			constant("Never", ShapeNeverTag),
			constant("Bool", ShapeBoolTag),
			constant("Int", ShapeIntTag),
			constant("Float", ShapeFloatTag),
			constant("String", ShapeStringTag),
			constant("Array", ShapeArrayTag),
			constant("Map", ShapeMapTag),
			constant("Set", ShapeSetTag),
			constant("Record", ShapeRecordTag),
			constant("Field", ShapeFieldTag),
			constant("Named", ShapeNamedTag),
			constant("Union", ShapeUnionTag),
			constant("Func", ShapeFuncTag),
			constant("Opaque", ShapeOpaqueTag),
			constant("Scalar", value.NewTag("Reflect.Scalar", value.AliasTag, nil)),
			fn("Reflect.erase", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return typeArg(g, who, a[0])
			}),
			fn("Reflect.arrayOf", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return &types.TypeValue{T: types.Array{Elem: typeArg(g, who, a[0]).T}}
			}),
			fn("Reflect.setOf", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return &types.TypeValue{T: types.Set{Elem: typeArg(g, who, a[0]).T}}
			}),
			fn("Reflect.mapOf", "k, v", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return &types.TypeValue{T: types.Map{Key: typeArg(g, who, a[0]).T, Val: typeArg(g, who, a[1]).T}}
			}),
			fn("Reflect.nullable", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return &types.TypeValue{T: nullable(g, who, typeArg(g, who, a[0]).T)}
			}),
			fn("Reflect.shape", "t", notYet),
			fn("Reflect.assignable", "from, to", notYet),
			fn("Reflect.default", "t, name", notYet),
			fn("Reflect.valueOf", "t, value", notYet),
			fn("Reflect.to", "d, t", notYet),
			fn("Reflect.type", "d", notYet),
			fn("Reflect.scalar", "d", notYet),
			fn("Reflect.items", "d", notYet),
			fn("Reflect.entries", "d", notYet),
			fn("Reflect.field", "d, name", notYet),
			fn("Reflect.payload", "d", notYet),
			fn("Reflect.makeNumber", "t, n", notYet),
			fn("Reflect.makeArray", "t, items", notYet),
			fn("Reflect.makeSet", "t, members", notYet),
			fn("Reflect.makeMap", "t, entries", notYet),
			fn("Reflect.makeRecord", "t, fields", notYet),
			fn("Reflect.makeTagged", "t, payload", notYet),
		)
	})
}
