<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Shape

```ant signature
type Reflect.Shape = Reflect.Null | Reflect.Void | Reflect.Never | Reflect.Bool | Reflect.Int | Reflect.Float | Reflect.String | Reflect.Array | Reflect.Map | Reflect.Set | Reflect.Record | Reflect.Tagged | Reflect.Union | Reflect.Func | Reflect.Opaque
```

Reflect.Shape is one level of a type's structure, as [`Reflect.shape`](../shape.md) gives it:
aliases expanded, unions flattened, and every child a [`Reflect.ErasedType`](ErasedType.md).

```ant signature
type Reflect.Null = {}
```

Reflect.Null is the shape of null.

```ant signature
type Reflect.Void = {}
```

Reflect.Void is the shape of void.

```ant signature
type Reflect.Never = {}
```

Reflect.Never is the shape of never.

```ant signature
type Reflect.Bool = {}
```

Reflect.Bool is the shape of bool.

```ant signature
type Reflect.Int = { bits: int, signed: bool }
```

Reflect.Int is the shape of an integer type: its width in bits and whether it is signed. int is \{
bits: 64, signed: true \} and byte is \{ bits: 8, signed: false \}.

```ant signature
type Reflect.Float = { bits: int }
```

Reflect.Float is the shape of float32 or float64: its width in bits.

```ant signature
type Reflect.String = {}
```

Reflect.String is the shape of string.

```ant signature
type Reflect.Array = { element: Reflect.ErasedType }
```

Reflect.Array is the shape of Array\[element\].

```ant signature
type Reflect.Map = { key: Reflect.ErasedType, value: Reflect.ErasedType }
```

Reflect.Map is the shape of Map\[key, value\].

```ant signature
type Reflect.Set = { member: Reflect.ErasedType }
```

Reflect.Set is the shape of Set\[member\].

```ant signature
type Reflect.Record = { decl: Reflect.Decl | null, args: Array[Reflect.ErasedType], fields: Array[Reflect.Field] }
```

Reflect.Record is the shape of a record type: its declaration, which renders as the declared name
and is null for an anonymous record type, its type arguments, and its fields in declaration order
with their types substituted, so that in Box\[int\] the field value has type int.

```ant signature
type Reflect.Tagged = { decl: Reflect.Decl, args: Array[Reflect.ErasedType], base: Reflect.ErasedType }
```

Reflect.Tagged is the shape of a named primitive type, whose values are tagged primitives: its
declaration, type arguments, and base type. Id\[User\] and Id\[Account\] share a declaration and
differ in args.

```ant signature
type Reflect.Union = { members: Array[Reflect.ErasedType] }
```

Reflect.Union is the shape of a union: its members, flattened and deduplicated, in the order
written. T \| null has a Reflect.Null member, and Result\[int, Error\] the records Ok\[int\] and
Err\[Error\].

```ant signature
type Reflect.Func = { params: Array[Reflect.ErasedType], result: Reflect.ErasedType }
```

Reflect.Func is the shape of a function type: its parameter types and its result type.

```ant signature
type Reflect.Opaque = { decl: Reflect.Decl, args: Array[Reflect.ErasedType] }
```

Reflect.Opaque is the shape of a type whose values reflection does not take apart: Bytes, Html,
Regex.Pattern, the time types, the sync, random, and resource types, and the Reflect types
themselves. decl is its declaration and args are its type arguments; String.from of the type value
renders the type. Code that handles a type specially tests for it by comparing type values, t ==
Reflect.erase(type(Bytes)), not by name.

<!-- end generated -->
