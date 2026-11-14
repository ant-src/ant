<!-- status: reviewed -->

# Reflection

Reflection in Ant starts from a type the program writes, never from a value. The type expression
`type(T)` makes the type `T` a value of type `Reflect.Type[T]`. The `Reflect` namespace takes such a
value apart one level at a time, boxes a value together with its type, reads a box along that type,
and builds new values from boxed parts. With these, a typed decoder or encoder for a data format is
ordinary library code. A few functions elsewhere take a type value too, to choose what to read or
write.

```ant
type Point = { x: int, y: int }
type(Point)                              // Point
type(Array[Point] | null)                // Array[Point] | null
type(Point | null) == type(null | Point) // true
```

`type` takes a type, never a value: unlike Python's `type(x)` or JavaScript's `typeof x`,
`type(config)` on a variable is a compile error.

```ant
type Config = { name: string }
let config = Config{ name: "app" }
let t = type(config) // error: type(...) takes a type, not a value; config is a value
```

A function that walks a type works on `Reflect.ErasedType`, since the type of a child is not known
statically one level down. `Reflect.erase` turns the caller's `Reflect.Type[T]` into a
`Reflect.ErasedType` at the start, giving up its static type:

```ant
type Point = { x: int, y: int }
type Route = { name: string, points: Array[Point] }

func describe(t: Reflect.ErasedType): string {
  let s = Reflect.shape(t)
  if s is Reflect.Record {
    return s.fields.map({ "${x.name}: ${describe(x.type)}" }) | String.join(", ")
  }
  if s is Reflect.Array {
    return "list of (${describe(s.element)})"
  }
  String.from(t)
}

println(describe(Reflect.erase(type(Route))))
```

```text
name: string, points: list of (x: int, y: int)
```

A box keeps the type it was made at, resolved to one member when that type is a union, and
`Reflect.to` gives the value back at a static type that the caller's `type(T)` fixes. It panics when
the box's type is not assignable to `T`.

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
Reflect.field(d, "y")                               // Reflect.Value(2: int)
Reflect.to(d, type(Point))                          // Point{ x: 1, y: 2 }
Reflect.type(Reflect.valueOf(type(int | null), 42)) // int
```

## Functions by job

### Type values

- [type](../../language/keywords/type.md): `type(T)`, the type `T` as a value
- [Reflect.erase](../Reflect/erase.md): a type value with its static type forgotten
- [Reflect.arrayOf](../Reflect/arrayOf.md): `Array[T]` from a type value for `T`
- [Reflect.setOf](../Reflect/setOf.md): `Set[T]` from a type value for `T`
- [Reflect.mapOf](../Reflect/mapOf.md): `Map[K, V]` from type values for `K` and `V`
- [Reflect.nullable](../Reflect/nullable.md): `T | null` from a type value for `T`

Inside a generic function `type(Array[T])` is a compile error, since `T` is not a closed type, so
the function builds the type from its `Reflect.Type[T]` parameter:

```ant
func listOf[T](t: Reflect.Type[T]): Reflect.Type[Array[T]] {
  Reflect.arrayOf(t)
}

listOf(type(int)) == type(Array[int]) // true
```

### Taking a type apart

- [Reflect.shape](../Reflect/shape.md): one level of a type's structure, as a `Reflect.Shape`
- [Reflect.assignable](../Reflect/assignable.md): whether a value of one type may go where another
  is expected

### Boxing and unboxing

- [Reflect.valueOf](../Reflect/valueOf.md): a value boxed with its type
- [Reflect.to](../Reflect/to.md): the value in a box, at a static type
- [Reflect.type](../Reflect/type.md): the type of a box, never a union or an alias
- [Reflect.same](../Reflect/same.md): whether two boxes hold the same record, array, map, set, or
  `Bytes`, and `==` for any other value

### Reading a box

- [Reflect.scalar](../Reflect/scalar.md): the leaf value in a box
- [Reflect.items](../Reflect/items.md): an array's elements or a set's members, boxed
- [Reflect.entries](../Reflect/entries.md): a map's entries, boxed
- [Reflect.field](../Reflect/field.md): one field of a record, boxed
- [Reflect.payload](../Reflect/payload.md): the base value of a named primitive type, boxed

### Building a value

- [Reflect.makeScalar](../Reflect/makeScalar.md): a leaf boxed at a scalar type, or `null` when it
  does not fit
- [Reflect.makeArray](../Reflect/makeArray.md): a new array from boxed elements
- [Reflect.makeSet](../Reflect/makeSet.md): a new set from boxed members
- [Reflect.makeMap](../Reflect/makeMap.md): a new map from boxed entries
- [Reflect.makeRecord](../Reflect/makeRecord.md): a new record from boxed fields, the rest from
  their defaults
- [Reflect.makeTagged](../Reflect/makeTagged.md): a new value of a named primitive type

### Types

- [Reflect.Type](../Reflect/types/Type.md): a type value whose static type is known
- [Reflect.ErasedType](../Reflect/types/ErasedType.md): a type value whose static type is not known
- [Reflect.Value](../Reflect/types/Value.md): a value together with a type it belongs to
- [Reflect.Shape](../Reflect/types/Shape.md): what `Reflect.shape` returns, one member per kind of
  type
- [Reflect.Field](../Reflect/types/Field.md): one field of a record shape
- [Reflect.Decl](../Reflect/types/Decl.md): the declaration of a named type, shared by all its
  instantiations
- [Reflect.Scalar](../Reflect/types/Scalar.md): what `Reflect.scalar` returns

### Type values in other namespaces

- [Bytes.read](../Bytes/read.md): a number of the given type from a byte buffer
- [Bytes.write](../Bytes/write.md): overwrite bytes with a number of the given type
- [Bytes.append](../Bytes/append.md): append a number of the given type
- [Bytes.sizeOf](../Bytes/sizeOf.md): how many bytes a number type takes

The type value picks the width and the kind of number, and fixes the static type of the result:

```ant
let b = Bytes.new()
b.append(type(uint16), uint16(513), Bytes.BigEndian{})
b                                          // Bytes(02 01)
b.read(type(uint16), 0, Bytes.BigEndian{}) // 513
Bytes.sizeOf(type(float32))                // 4
```

## See also

[Reflect](../Reflect/index.md) · [Types at run time](../../guide/runtime-types.md) ·
[type](../../language/keywords/type.md) · [Data formats](data-formats.md) · [Json](../Json/index.md)
· [Bytes](../Bytes/index.md)
