package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/value"
)

// The Reflect namespace (section 11, Reflect): type values, shapes, and reflected values. The
// opaque types are tags that is cannot test, as the sync types are; the shapes are record tags
// whose values Reflect.shape builds. A type value is the checker's own types.Type, so
// equality, union membership, substitution, and assignability are the checker's code.
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

// reflectedArg takes a Reflect.Value.
func reflectedArg(g *interp.Goroutine, who string, v value.Value) *value.Reflected {
	return typedArg[*value.Reflected](g, who, "a reflected value", v)
}

func recordArg(g *interp.Goroutine, who string, v value.Value) *value.Record {
	return typedArg[*value.Record](g, who, "a record", v)
}

func typeValue(t types.Type) *types.TypeValue { return &types.TypeValue{T: t} }

// typeOf is the type a Reflect.Value holds.
func typeOf(d *value.Reflected) types.Type { return d.Type.(*types.TypeValue).T }

func typeArray(ts []types.Type) *value.Array {
	vals := make([]value.Value, len(ts))
	for i, t := range ts {
		vals[i] = typeValue(t)
	}
	return value.NewArray(vals)
}

// box makes a Reflect.Value of v at t, resolved to the member v belongs to when t is a union,
// never an alias (section 11, Reflect, the invariant). An unchecked program whose value carries
// another kind or tag panics here, as a wrong annotation panics at its use.
func box(g *interp.Goroutine, who string, t types.Type, v value.Value) *value.Reflected {
	m, ok := memberOf(t, v)
	if !ok {
		g.Panic("%s: %s does not belong to %s", who, value.Describe(v), types.String(t))
	}
	return &value.Reflected{Type: typeValue(m), Value: v}
}

// memberOf resolves t, aliases expanded, to the member v belongs to, by v's kind or tag alone:
// t itself unless it is a union, and for a union the one member whose discriminator v carries,
// an anonymous record member taking a record whose tag no named member has (section 4, Unions).
func memberOf(t types.Type, v value.Value) (types.Type, bool) {
	t = types.Expand(t)
	u, ok := t.(types.Union)
	if !ok {
		return t, belongs(t, v)
	}
	var anonymous types.Type
	for _, m := range u.Members {
		m = types.Expand(m)
		if _, ok := m.(types.Record); ok {
			anonymous = m
			continue
		}
		if belongs(m, v) {
			return m, true
		}
	}
	if anonymous != nil && belongs(anonymous, v) {
		return anonymous, true
	}
	return nil, false
}

// belongs reports whether v carries the discriminator of t, a type that is not a union: its
// kind, or its tag (section 4, Kinds).
func belongs(t types.Type, v value.Value) bool {
	switch x := types.Expand(t).(type) {
	case types.Prim:
		return value.KindOf(v) == x.Kind
	case types.Null:
		return value.KindOf(v) == value.KindNull
	case types.Void:
		return value.KindOf(v) == value.KindVoid
	case types.Array:
		return value.KindOf(v) == value.KindArray
	case types.Map:
		return value.KindOf(v) == value.KindMap
	case types.Set:
		return value.KindOf(v) == value.KindSet
	case types.Func:
		return value.KindOf(v) == value.KindFunc
	case types.Record:
		_, ok := v.(*value.Record)
		return ok
	case types.Named:
		switch id := x.Decl.Tag.(type) {
		case value.Kind:
			return value.KindOf(v) == id
		case *value.Tag:
			switch y := v.(type) {
			case *value.Record:
				return y.Tag == id
			case *value.Tagged:
				return y.Tag == id
			case value.Time:
				return y.TimeType() == id.Name
			case *value.Regex:
				return id.Kind == value.RegexTag
			case value.Sync:
				return y.SyncName() == id.Name
			case value.TypeValue:
				return id == ReflectTypeTag || id == ReflectErasedTypeTag
			case value.Declaration:
				return id == ReflectDeclTag
			case *value.Reflected:
				return id == ReflectValueTag
			}
		}
	}
	return false
}

// shape takes a type apart one level: aliases expanded, unions flattened (section 11, Reflect).
func shape(t types.Type) value.Value {
	switch x := types.Expand(t).(type) {
	case types.Prim:
		switch {
		case x.Kind.IsInteger():
			return value.NewRecord(ShapeIntTag, []value.Value{int64(x.Kind.Bits()), x.Kind.IsSigned()})
		case x.Kind == value.KindFloat32:
			return value.NewRecord(ShapeFloatTag, []value.Value{int64(32)})
		case x.Kind == value.KindFloat64:
			return value.NewRecord(ShapeFloatTag, []value.Value{int64(64)})
		case x.Kind == value.KindBool:
			return value.NewRecord(ShapeBoolTag, nil)
		}
		return value.NewRecord(ShapeStringTag, nil)
	case types.Null:
		return value.NewRecord(ShapeNullTag, nil)
	case types.Void:
		return value.NewRecord(ShapeVoidTag, nil)
	case types.Never:
		return value.NewRecord(ShapeNeverTag, nil)
	case types.Array:
		return value.NewRecord(ShapeArrayTag, []value.Value{typeValue(x.Elem)})
	case types.Map:
		return value.NewRecord(ShapeMapTag, []value.Value{typeValue(x.Key), typeValue(x.Val)})
	case types.Set:
		return value.NewRecord(ShapeSetTag, []value.Value{typeValue(x.Elem)})
	case types.Record:
		return value.NewRecord(ShapeRecordTag, []value.Value{value.Null, value.Null, value.NewArray(nil), fieldShapes(x.Fields)})
	case types.Named:
		decl := types.DeclValue(x.Decl)
		args := typeArray(x.Args)
		switch x.Decl.Kind {
		case types.RecordDecl:
			return value.NewRecord(ShapeRecordTag, []value.Value{x.Decl.Name, decl, args, fieldShapes(x.FieldsOf())})
		case types.PrimitiveDecl:
			return value.NewRecord(ShapeNamedTag, []value.Value{x.Decl.Name, decl, args, typeValue(types.Prim{Kind: x.Decl.Base})})
		}
		return value.NewRecord(ShapeOpaqueTag, []value.Value{types.String(x), decl, args})
	case types.Union:
		members := make([]types.Type, len(x.Members))
		for i, m := range x.Members {
			members[i] = types.Expand(m)
		}
		return value.NewRecord(ShapeUnionTag, []value.Value{typeArray(members)})
	case types.Func:
		return value.NewRecord(ShapeFuncTag, []value.Value{typeArray(x.Params), typeValue(x.Return)})
	}
	return value.NewRecord(ShapeNeverTag, nil)
}

func fieldShapes(fields []types.Field) *value.Array {
	vals := make([]value.Value, len(fields))
	for i, f := range fields {
		vals[i] = value.NewRecord(ShapeFieldTag, []value.Value{f.Name, typeValue(f.Type), f.Default})
	}
	return value.NewArray(vals)
}

// namedRecord takes a type that must be a named record type, and gives its declaration and tag.
func namedRecord(g *interp.Goroutine, who string, t types.Type) (types.Named, *value.Tag) {
	n, ok := types.Expand(t).(types.Named)
	if !ok || n.Decl.Kind != types.RecordDecl {
		g.Panic("%s needs a named record type, got %s", who, types.String(t))
	}
	tag, ok := n.Decl.Tag.(*value.Tag)
	if !ok {
		g.Panic("%s: %s has no tag to build a value with", who, types.String(t))
	}
	return n, tag
}

// fits panics unless a part's type is assignable to the slot it goes into, naming the part, the
// slot, and the whole type (section 11, Reflect, Builders).
func fits(g *interp.Goroutine, who, part string, d *value.Reflected, want types.Type, slot string, whole types.Type) {
	if !types.Assignable(typeOf(d), want) {
		g.Panic("%s: %s of type %s is not assignable to the %s %s of %s", who, part, d.Type.String(), slot, types.String(want), types.String(whole))
	}
}

func init() {
	register("Reflect", func() value.Value {
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
				return typeValue(types.Array{Elem: typeArg(g, who, a[0]).T})
			}),
			fn("Reflect.setOf", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return typeValue(types.Set{Elem: typeArg(g, who, a[0]).T})
			}),
			fn("Reflect.mapOf", "k, v", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return typeValue(types.Map{Key: typeArg(g, who, a[0]).T, Val: typeArg(g, who, a[1]).T})
			}),
			fn("Reflect.nullable", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				// The members are closed types, so the union cannot fail to form.
				u, err := types.MakeUnion([]types.Type{typeArg(g, who, a[0]).T, types.Null{}})
				if err != nil {
					g.Panic("%s: %s", who, err)
				}
				return typeValue(u)
			}),
			fn("Reflect.shape", "t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return shape(typeArg(g, who, a[0]).T)
			}),
			fn("Reflect.assignable", "from, to", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return types.Assignable(typeArg(g, who, a[0]).T, typeArg(g, who, a[1]).T)
			}),
			fn("Reflect.default", "t, name", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t := typeArg(g, who, a[0]).T
				name := stringArg(g, who, a[1])
				n, tag := namedRecord(g, who, t)
				i, ok := tag.FieldIndex(name)
				if !ok {
					g.Panic("%s: %s has no field %s", who, types.String(t), name)
				}
				v, ok := g.FieldDefault(tag, i)
				if !ok {
					g.Panic("%s: the field %s of %s has no default", who, name, types.String(t))
				}
				return box(g, who, n.FieldsOf()[i].Type, v)
			}),
			fn("Reflect.valueOf", "t, value", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return box(g, who, typeArg(g, who, a[0]).T, a[1])
			}),
			fn("Reflect.to", "d, t", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d := reflectedArg(g, who, a[0])
				t := typeArg(g, who, a[1]).T
				if !types.Assignable(typeOf(d), t) {
					g.Panic("%s: %s is not assignable to %s", who, d.Type.String(), types.String(t))
				}
				return d.Value
			}),
			fn("Reflect.type", "d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return reflectedArg(g, who, a[0]).Type
			}),
			fn("Reflect.scalar", "d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d := reflectedArg(g, who, a[0])
				switch types.Expand(typeOf(d)).(type) {
				case types.Prim, types.Null, types.Void:
					return d.Value
				}
				g.Panic("%s needs a value of a scalar type, got one of type %s", who, d.Type.String())
				return nil
			}),
			fn("Reflect.items", "d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d := reflectedArg(g, who, a[0])
				var elem types.Type
				var items []value.Value
				switch x := types.Expand(typeOf(d)).(type) {
				case types.Array:
					elem, items = x.Elem, arrayArg(g, who, d.Value).Values()
				case types.Set:
					elem, items = x.Elem, setArg(g, who, d.Value).Values()
				default:
					g.Panic("%s needs a value of an array or set type, got one of type %s", who, d.Type.String())
				}
				boxes := make([]value.Value, len(items))
				for i, v := range items {
					boxes[i] = box(g, who, elem, v)
				}
				return value.NewArray(boxes)
			}),
			fn("Reflect.entries", "d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d := reflectedArg(g, who, a[0])
				m, ok := types.Expand(typeOf(d)).(types.Map)
				if !ok {
					g.Panic("%s needs a value of a map type, got one of type %s", who, d.Type.String())
				}
				entries := mapArg(g, who, d.Value).Entries()
				boxes := make([]value.Value, len(entries))
				for i, e := range entries {
					boxes[i] = value.Entry(box(g, who, m.Key, e.Key), box(g, who, m.Val, e.Value))
				}
				return value.NewArray(boxes)
			}),
			fn("Reflect.field", "d, name", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d := reflectedArg(g, who, a[0])
				name := stringArg(g, who, a[1])
				fields := types.FieldsOf(typeOf(d))
				if fields == nil {
					g.Panic("%s needs a value of a record type, got one of type %s", who, d.Type.String())
				}
				for _, f := range fields {
					if f.Name != name {
						continue
					}
					v, ok := recordArg(g, who, d.Value).Field(name)
					if !ok {
						g.Panic("%s: the record has no field %s", who, name)
					}
					return box(g, who, f.Type, v)
				}
				g.Panic("%s: %s has no field %s", who, d.Type.String(), name)
				return nil
			}),
			fn("Reflect.payload", "d", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				d := reflectedArg(g, who, a[0])
				n, ok := types.Expand(typeOf(d)).(types.Named)
				if !ok || n.Decl.Kind != types.PrimitiveDecl {
					g.Panic("%s needs a value of a named primitive type, got one of type %s", who, d.Type.String())
				}
				tagged := typedArg[*value.Tagged](g, who, "a tagged value", d.Value)
				return box(g, who, types.Prim{Kind: n.Decl.Base}, tagged.Payload)
			}),
			fn("Reflect.makeNumber", "t, n", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t := typeArg(g, who, a[0]).T
				p, ok := types.Expand(t).(types.Prim)
				if !ok || !p.Kind.IsNumber() {
					g.Panic("%s needs an integer or float type, got %s", who, types.String(t))
				}
				if !value.KindOf(a[1]).IsNumber() {
					g.Panic("%s needs a number, got %s", who, value.Describe(a[1]))
				}
				v, err := value.Convert(p.Kind, a[1])
				if err != nil {
					return value.Null
				}
				return &value.Reflected{Type: typeValue(p), Value: v}
			}),
			fn("Reflect.makeArray", "t, items", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t := typeArg(g, who, a[0]).T
				at, ok := types.Expand(t).(types.Array)
				if !ok {
					g.Panic("%s needs an array type, got %s", who, types.String(t))
				}
				items := arrayArg(g, who, a[1]).Values()
				vals := make([]value.Value, len(items))
				for i, item := range items {
					d := reflectedArg(g, who, item)
					fits(g, who, "an item", d, at.Elem, "element type", at)
					vals[i] = d.Value
				}
				return &value.Reflected{Type: typeValue(at), Value: value.NewArray(vals)}
			}),
			fn("Reflect.makeSet", "t, members", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t := typeArg(g, who, a[0]).T
				st, ok := types.Expand(t).(types.Set)
				if !ok {
					g.Panic("%s needs a set type, got %s", who, types.String(t))
				}
				set := value.NewSet()
				for _, member := range arrayArg(g, who, a[1]).Values() {
					d := reflectedArg(g, who, member)
					fits(g, who, "a member", d, st.Elem, "member type", st)
					set.Add(d.Value)
				}
				return &value.Reflected{Type: typeValue(st), Value: set}
			}),
			fn("Reflect.makeMap", "t, entries", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t := typeArg(g, who, a[0]).T
				mt, ok := types.Expand(t).(types.Map)
				if !ok {
					g.Panic("%s needs a map type, got %s", who, types.String(t))
				}
				m := value.NewMap()
				for _, entry := range arrayArg(g, who, a[1]).Values() {
					e := recordArg(g, who, entry)
					kv, ok := e.Field("key")
					vv, ok2 := e.Field("value")
					if !ok || !ok2 {
						g.Panic("%s needs entries with a key and a value, got %s", who, value.Describe(entry))
					}
					k, v := reflectedArg(g, who, kv), reflectedArg(g, who, vv)
					fits(g, who, "a key", k, mt.Key, "key type", mt)
					fits(g, who, "a value", v, mt.Val, "value type", mt)
					m.Set(k.Value, v.Value)
				}
				return &value.Reflected{Type: typeValue(mt), Value: m}
			}),
			fn("Reflect.makeRecord", "t, fields", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t := typeArg(g, who, a[0]).T
				n, tag := namedRecord(g, who, t)
				fields := n.FieldsOf()
				vals := make([]value.Value, len(tag.Fields))
				for _, e := range mapArg(g, who, a[1]).Entries() {
					name, ok := e.Key.(string)
					if !ok {
						g.Panic("%s needs field names as keys, got %s", who, value.Describe(e.Key))
					}
					i, ok := tag.FieldIndex(name)
					if !ok {
						g.Panic("%s: %s has no field %s", who, types.String(t), name)
					}
					d := reflectedArg(g, who, e.Value)
					if !types.Assignable(typeOf(d), fields[i].Type) {
						g.Panic("%s: a value of type %s is not assignable to the field %s of %s, whose type is %s", who, d.Type.String(), name, types.String(t), types.String(fields[i].Type))
					}
					vals[i] = d.Value
				}
				for i, v := range vals {
					if v == nil && !interp.HasDefault(tag, i) {
						g.Panic("%s: %s is missing the field %s, which has no default", who, types.String(t), tag.Fields[i])
					}
				}
				g.FillDefaults(tag, vals)
				return &value.Reflected{Type: typeValue(n), Value: value.NewRecord(tag, vals)}
			}),
			fn("Reflect.makeTagged", "t, payload", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t := typeArg(g, who, a[0]).T
				n, ok := types.Expand(t).(types.Named)
				if !ok || n.Decl.Kind != types.PrimitiveDecl {
					g.Panic("%s needs a named primitive type, got %s", who, types.String(t))
				}
				tag, _ := n.Decl.Tag.(*value.Tag)
				d := reflectedArg(g, who, a[1])
				fits(g, who, "a payload", d, types.Prim{Kind: n.Decl.Base}, "base type", n)
				return &value.Reflected{Type: typeValue(n), Value: &value.Tagged{Tag: tag, Payload: d.Value}}
			}),
		)
	})
}
