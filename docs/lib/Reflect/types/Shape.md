<!-- status: reviewed -->
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

## Details

- `Reflect.Shape` is a union of fifteen record types. Narrow a shape with `is` before reading a
  field; after a chain of `is` tests that covers all fifteen, no `else` is needed.
- An alias has no shape of its own: `Reflect.shape` expands it. A named primitive type such as
  `type Celsius = float64` is a type of its own, so its shape is `Reflect.Tagged`, not
  `Reflect.Float`.
- Every call to `Reflect.shape` builds a new shape. Changing it changes nothing else.

## Basic examples

The shape of a primitive type has no children, or only its size:

```ant
Reflect.shape(type(bool).erase())    // Reflect.Bool{}
Reflect.shape(type(int).erase())     // Reflect.Int{ bits: 64, signed: true }
Reflect.shape(type(float32).erase()) // Reflect.Float{ bits: 32 }
```

The shape of a container holds its children as type values:

```ant
Reflect.shape(type(Array[string]).erase())    // Reflect.Array{ element: string }
Reflect.shape(type(Map[string, int]).erase()) // Reflect.Map{ key: string, value: int }
```

Narrow a shape with `is` to read its fields:

```ant
let s = Reflect.shape(type(Array[int]).erase())
if s is Reflect.Array {
  println(s.element)
}
```

```text
int
```

## Scope

### Null, void, never, and string

Four shapes are empty records:

```ant
Reflect.shape(type(null).erase())   // Reflect.Null{}
Reflect.shape(type(void).erase())   // Reflect.Void{}
Reflect.shape(type(never).erase())  // Reflect.Never{}
Reflect.shape(type(string).erase()) // Reflect.String{}
```

### Numbers

`Reflect.Int` gives the width and sign of an integer type, and `Reflect.Float` the width of a float
type. `byte` is an unsigned 8-bit integer:

```ant
Reflect.shape(type(int8).erase())    // Reflect.Int{ bits: 8, signed: true }
Reflect.shape(type(uint16).erase())  // Reflect.Int{ bits: 16, signed: false }
Reflect.shape(type(byte).erase())    // Reflect.Int{ bits: 8, signed: false }
Reflect.shape(type(float64).erase()) // Reflect.Float{ bits: 64 }
```

### Arrays, maps, and sets

A child may be a container itself. The shape stops at the first level:

```ant
Reflect.shape(type(Set[string]).erase())           // Reflect.Set{ member: string }
Reflect.shape(type(Map[string, Set[int]]).erase()) // Reflect.Map{ key: string, value: Set[int] }
```

### Records

A record shape gives the declaration, the type arguments, and the fields in declaration order.
`hasDefault` says whether the field has a default:

```ant
type Point = { x: int, y: int = 0 }
Reflect.shape(type(Point).erase())
// Reflect.Record{ decl: Point, args: [], fields: [Reflect.Field{ name: "x", type: int, hasDefault: false }, Reflect.Field{ name: "y", type: int, hasDefault: true }] }
```

In a generic record, the fields have the type arguments substituted:

```ant
type Box[T] = { value: T }
Reflect.shape(type(Box[int]).erase())
// Reflect.Record{ decl: Box, args: [int], fields: [Reflect.Field{ name: "value", type: int, hasDefault: false }] }
```

An anonymous record type has no declaration, so `decl` is `null`:

```ant
Reflect.shape(type({ name: string }).erase())
// Reflect.Record{ decl: null, args: [], fields: [Reflect.Field{ name: "name", type: string, hasDefault: false }] }
```

### Named primitive types

A named primitive type gives its declaration, its type arguments, and its base:

```ant
type Celsius = float64
type Id[T] = string
type User = { name: string }
Reflect.shape(type(Celsius).erase())  // Reflect.Tagged{ decl: Celsius, args: [], base: float64 }
Reflect.shape(type(Id[User]).erase()) // Reflect.Tagged{ decl: Id, args: [User], base: string }
```

### Unions

A union lists its members, flattened and deduplicated, in the order written. An alias of a union
expands to its members:

```ant
type Circle = { r: float64 }
type Rect = { w: float64, h: float64 }
type Figure = Circle | Rect
Reflect.shape(type(int | null).erase())             // Reflect.Union{ members: [int, null] }
Reflect.shape(type(Figure | null | Circle).erase()) // Reflect.Union{ members: [Circle, Rect, null] }
Reflect.shape(type(Result[int, Error]).erase())     // Reflect.Union{ members: [Ok[int], Err[Error]] }
```

### Function types

```ant
Reflect.shape(type(func(int, string): bool).erase()) // Reflect.Func{ params: [int, string], result: bool }
Reflect.shape(type(func(): void).erase())            // Reflect.Func{ params: [], result: void }
```

### Opaque types

A type whose values reflection does not take apart gives only its declaration and type arguments:

```ant
Reflect.shape(type(Bytes).erase())         // Reflect.Opaque{ decl: Bytes, args: [] }
Reflect.shape(type(Regex.Pattern).erase()) // Reflect.Opaque{ decl: Regex.Pattern, args: [] }
Reflect.shape(type(Channel[int]).erase())  // Reflect.Opaque{ decl: Channel, args: [int] }
```

### Recursive types

A recursive type has a finite shape, because its children are type values:

```ant
type Tree = { value: int, children: Array[Tree] }
Reflect.shape(type(Tree).erase())
// Reflect.Record{ decl: Tree, args: [], fields: [Reflect.Field{ name: "value", type: int, hasDefault: false }, Reflect.Field{ name: "children", type: Array[Tree], hasDefault: false }] }
```

## Applications

### Describe a type in words

A chain of `is` tests over all fifteen members needs no `else`. Each branch recurses into the
children it has:

```ant
func describe(t: Reflect.ErasedType): string {
  let s = Reflect.shape(t)
  if s is Reflect.Null {
    "null"
  } else if s is Reflect.Void {
    "nothing"
  } else if s is Reflect.Never {
    "no value"
  } else if s is Reflect.Bool {
    "a boolean"
  } else if s is Reflect.Int {
    if s.signed { "an integer" } else { "a non-negative integer" }
  } else if s is Reflect.Float {
    "a number"
  } else if s is Reflect.String {
    "text"
  } else if s is Reflect.Array {
    "a list of ${describe(s.element)}"
  } else if s is Reflect.Map {
    "a table from ${describe(s.key)} to ${describe(s.value)}"
  } else if s is Reflect.Set {
    "a set of ${describe(s.member)}"
  } else if s is Reflect.Record {
    if s.decl == null { "a record" } else { "a ${s.decl}" }
  } else if s is Reflect.Tagged {
    "a ${s.decl} (${describe(s.base)})"
  } else if s is Reflect.Union {
    s.members | Array.map(describe) | String.join(" or ")
  } else if s is Reflect.Func {
    "a function"
  } else if s is Reflect.Opaque {
    "a ${t}"
  }
}

type UserId = string
println(describe(type(Map[UserId, Array[int | null]]).erase()))
println(describe(type(Set[byte]).erase()))
println(describe(type(Bytes).erase()))
```

```text
a table from a UserId (text) to a list of an integer or null
a set of a non-negative integer
a Bytes
```

### List the paths of a configuration

Walk the fields of nested records to list every setting a configuration type has:

```ant
type Server = { host: string, port: int = 8080 }
type Config = { name: string, server: Server, backup: Server | null }

func paths(t: Reflect.ErasedType, prefix: string): Array[string] {
  let s = Reflect.shape(t)
  if !(s is Reflect.Record) { return [prefix] }
  s.fields | Array.flatMap({ paths(x.type, "${prefix}.${x.name}") })
}

for p in paths(type(Config).erase(), "config") {
  println(p)
}
```

```text
config.name
config.server.host
config.server.port
config.backup
```

`backup` is a union, not a record, so the walk stops there and lists it as one path.

### Find a recursive type

A walk over a recursive type must stop. Keep the types on the current path in a set, since type
values hash, and stop at one already on it:

```ant
type Tree = { value: int, children: Array[Tree] }
type Point = { x: int, y: int }
type Line = { from: Point, to: Point }
type Dir = { entries: Map[string, Dir] }

func children(t: Reflect.ErasedType): Array[Reflect.ErasedType] {
  let s = Reflect.shape(t)
  if s is Reflect.Record {
    s.fields | Array.map({ x.type })
  } else if s is Reflect.Array {
    [s.element]
  } else if s is Reflect.Map {
    [s.key, s.value]
  } else if s is Reflect.Set {
    [s.member]
  } else if s is Reflect.Tagged {
    [s.base]
  } else if s is Reflect.Union {
    s.members
  } else {
    []
  }
}

func recursive(t: Reflect.ErasedType, path: Set[Reflect.ErasedType]): bool {
  if path.has(t) { return true }
  path.add(t)
  defer { path.remove(t) }
  children(t).some({ recursive(x, path) })
}

println(recursive(type(Tree).erase(), Set.of([])))
println(recursive(type(Line).erase(), Set.of([])))
println(recursive(type(Json.Value).erase(), Set.of([])))
println(recursive(type(Dir).erase(), Set.of([])))
```

```text
true
false
true
true
```

`Line` reaches `Point` twice, but never inside itself, so it is not recursive.

## Properties and relations

An alias has the shape of its body:

```ant
type Names = Array[string]
assert(Reflect.shape(type(Names).erase()) == Reflect.shape(type(Array[string]).erase()))
```

Two instantiations of a generic type share one declaration and differ in `args`:

```ant
type Box[T] = { value: T }
let a = Reflect.shape(type(Box[int]).erase())
let b = Reflect.shape(type(Box[string]).erase())
if a is Reflect.Record && b is Reflect.Record {
  assert(a.decl == b.decl)
  assert(a.args != b.args)
}
```

A shape is a new value on every call, so changing one does not change the next:

```ant
let s = Reflect.shape(type(Array[int]).erase())
if s is Reflect.Array {
  s.element = type(string).erase()
}
assert(Reflect.shape(type(Array[int]).erase()) == Reflect.Array{ element: type(int).erase() })
```

## Possible issues

A field is reachable only on a narrowed shape:

```ant
let s = Reflect.shape(type(Array[int]).erase())
println(s.element) // error: Reflect.Shape has no field element on every member: narrow the value first
```

The order of union members is for display. Two unions with the same members are the same type, but
their shapes list the members in the order written, so the shapes differ. Compare type values, not
shapes:

```ant
type(int | null).erase() == type(null | int).erase()                               // true
Reflect.shape(type(int | null).erase()) == Reflect.shape(type(null | int).erase()) // false
```

A chain of `is` tests that leaves out a member has an implicit `else` of type `void`:

```ant
func label(t: Reflect.ErasedType): string {
  let s = Reflect.shape(t)
  if s is Reflect.Int {
    "integer"
  } else if s is Reflect.String {
    // error: an if without else has type void when the condition is false, which cannot be used where string is expected
    "text"
  }
}
```

Test for a particular opaque type by comparing type values, not by the name `decl` renders as:

```ant
func isBytes(t: Reflect.ErasedType): bool {
  t == type(Bytes).erase()
}

isBytes(type(Bytes).erase())       // true
isBytes(type(Array[byte]).erase()) // false
```

## Neat examples

`Reflect.Shape` is itself a union, so its shape lists the fifteen members:

```ant
Reflect.shape(type(Reflect.Shape).erase())
// Reflect.Union{ members: [Reflect.Null, Reflect.Void, Reflect.Never, Reflect.Bool, Reflect.Int, Reflect.Float, Reflect.String, Reflect.Array, Reflect.Map, Reflect.Set, Reflect.Record, Reflect.Tagged, Reflect.Union, Reflect.Func, Reflect.Opaque] }
```

## See also

[Reflect.shape](../shape.md) · [Reflect.ErasedType](ErasedType.md) · [Reflect.Field](Field.md) ·
[Reflect.Decl](Decl.md) · [Reflect.erase](../erase.md) · [Reflect.assignable](../assignable.md) ·
[Reflect.makeRecord](../makeRecord.md)
