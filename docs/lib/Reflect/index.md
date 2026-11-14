<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect

Reflect holds reflection: types as values, written type(T), taken apart one level at a time with
Reflect.shape, and values boxed with a type they belong to, Reflect.Value, read along that type and
built from checked parts. With these, typed decoding and encoding for a data format are ordinary
library code, and the format keeps every policy.

<!-- end generated -->

Reflection in Ant starts from a type the program writes, never from a value. `type(T)` makes the
type `T` a value of type `Reflect.Type[T]`. `Reflect.shape` takes such a value apart one level at a
time. `Reflect.valueOf` boxes a value together with its type as a `Reflect.Value`, which the readers
take apart along that type and the builders make from parts they check. `Reflect.to` turns a box
back into a typed value.

## Details

- `type(T)` takes a type, never a value, and the type cannot mention a type parameter. A generic
  function takes a `Reflect.Type[T]` parameter instead, and builds larger types from it with
  [Reflect.arrayOf](arrayOf.md), [Reflect.setOf](setOf.md), [Reflect.mapOf](mapOf.md), and
  [Reflect.nullable](nullable.md).
- Two type values are `==` when they denote the same type: aliases expanded, and union members
  compared as a set. A type value can be a map key or a set member.
- [Reflect.erase](erase.md) forgets the static type and gives a `Reflect.ErasedType`, which
  [Reflect.shape](shape.md), [Reflect.assignable](assignable.md), and the make functions take. A
  `Reflect.Type[T]` never converts to one implicitly.
- A box's type is never a union or an alias. Nothing changes a value through a box: the builders
  make new values, and the readers return boxes that refer to the original children.
- The readers, the builders other than [Reflect.makeScalar](makeScalar.md), and [Reflect.to](to.md)
  panic when a type has the wrong shape or a part does not fit its slot. Test first with
  [Reflect.shape](shape.md) and [Reflect.assignable](assignable.md) when the input may not fit.

## Basic examples

`type(T)` is the type as a value. It renders as the type, and `int` is `int64`:

```ant
type Config = { name: string, port: int }
type(Config)               // Config
type(Array[Config] | null) // Array[Config] | null
type(int) == type(int64)   // true
```

`Reflect.shape` gives one level of a type's structure:

```ant
type Config = { name: string, port: int = 8080 }
let s = Reflect.shape(Reflect.erase(type(Config)))
if s is Reflect.Record {
  for f in s.fields {
    println("${f.name}: ${f.type}, default: ${f.hasDefault}")
  }
}
println(Reflect.shape(Reflect.erase(type(Array[Config]))))
```

```text
name: string, default: false
port: int, default: true
Reflect.Array{ element: Config }
```

`Reflect.valueOf` boxes a value with its type, the readers take it apart, and `Reflect.to` gives it
back at its static type:

```ant
type Config = { name: string, port: int }
let d = Reflect.valueOf(type(Config), Config{ name: "api", port: 80 })
d                           // Reflect.Value(Config{ name: "api", port: 80 }: Config)
Reflect.field(d, "port")    // Reflect.Value(80: int)
Reflect.to(d, type(Config)) // Config{ name: "api", port: 80 }
```

A builder makes a new value from boxed parts, and each field the parts leave out takes its default:

```ant
type Config = { name: string, port: int = 8080 }
let name = Reflect.valueOf(type(string), "api")
let d = Reflect.makeRecord(Reflect.erase(type(Config)), Map{ name })
Reflect.to(d, type(Config)) // Config{ name: "api", port: 8080 }
```

## Scope

### Type values

A union is flattened and compared as a set of members, and an alias is the type it stands for:

```ant
type Names = Array[string]
type(int | null) == type(null | int)         // true
type(Names) == type(Array[string])           // true
type(Names)                                  // Names
Reflect.erase(type(Names)) | Reflect.shape() // Reflect.Array{ element: string }
```

Prelude types can be taken apart too. `Result` is a union of two records:

```ant
type(Result[int, Error]) | Reflect.erase() | Reflect.shape() // Reflect.Union{ members: [Ok[int], Err[Error]] }
```

### Boxes at a union

At a union, a box takes the member its value belongs to, found from the value's kind or tag:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Figure = Circle | Square
let d = Reflect.valueOf(type(Figure | null), Square{ side: 2.0 })
Reflect.type(d)                            // Square
Reflect.valueOf(type(Figure | null), null) // Reflect.Value(null: null)
```

### Leaves

`Reflect.scalar` gives a leaf in its own kind, and `Reflect.makeScalar` boxes a leaf at a scalar
type, converting a number. It returns `null` when the leaf does not fit:

```ant
Reflect.scalar(Reflect.valueOf(type(byte), byte(7)))  // 7
Reflect.makeScalar(Reflect.erase(type(int)), 3.0)     // Reflect.Value(3: int)
Reflect.makeScalar(Reflect.erase(type(int)), 3.5)     // null
Reflect.makeScalar(Reflect.erase(type(string)), true) // null
```

## Functions by job

### Type values

- [Reflect.erase](erase.md): a type value with its static type forgotten
- [Reflect.arrayOf](arrayOf.md), [Reflect.setOf](setOf.md), [Reflect.mapOf](mapOf.md),
  [Reflect.nullable](nullable.md): a larger type from a type value, keeping the static type
- [Reflect.shape](shape.md): one level of a type's structure
- [Reflect.assignable](assignable.md): whether a value of one type may go where another is expected

### Boxing and unboxing

- [Reflect.valueOf](valueOf.md): a value boxed with its type
- [Reflect.to](to.md): the value in a box, at a static type
- [Reflect.type](type.md): the type of a box
- [Reflect.same](same.md): whether two boxes hold one record, array, map, set, or `Bytes`

### Reading

- [Reflect.scalar](scalar.md): the leaf value in a box
- [Reflect.items](items.md): an array's elements or a set's members, boxed
- [Reflect.entries](entries.md): a map's entries, boxed
- [Reflect.field](field.md): one field of a record, boxed
- [Reflect.payload](payload.md): the base value of a named primitive type, boxed

### Building

- [Reflect.makeScalar](makeScalar.md): a leaf boxed at a scalar type, or `null`
- [Reflect.makeArray](makeArray.md), [Reflect.makeSet](makeSet.md), [Reflect.makeMap](makeMap.md): a
  new container from boxed parts
- [Reflect.makeRecord](makeRecord.md): a new record from boxed fields and defaults
- [Reflect.makeTagged](makeTagged.md): a new value of a named primitive type

### Types

- [Reflect.Type](types/Type.md) and [Reflect.ErasedType](types/ErasedType.md): type values
- [Reflect.Value](types/Value.md): a value together with its type
- [Reflect.Shape](types/Shape.md), [Reflect.Field](types/Field.md), and
  [Reflect.Decl](types/Decl.md): what `Reflect.shape` returns
- [Reflect.Scalar](types/Scalar.md): what `Reflect.scalar` returns

## Applications

### A typed JSON decoder

Decode JSON into any type the caller names. The decoder walks the erased type, builds each part with
a builder, and lets the caller's `type(T)` fix the result at the end. It fills a missing field from
its default, tries union members in order, and reports a leaf that does not fit:

```ant
type Celsius = float64
type Reading = { sensor: string, value: Celsius, tags: Array[string] = [], note: string | null = null }

func decodeValue(t: Reflect.ErasedType, j: Json.Value): Result[Reflect.Value, Error] {
  let s = Reflect.shape(t)
  if s is Reflect.Record {
    if !(j is Map) {
      return Result.fail("want an object for ${t}")
    }
    let fields: Map[string, Reflect.Value] = Map{}
    for f in s.fields {
      if !j.has(f.name) {
        if f.hasDefault {
          continue
        }
        return Result.fail("missing field ${f.name}")
      }
      fields[f.name] = try (decodeValue(f.type, j[f.name]) | Result.wrapErr(f.name))
    }
    return Result.ok(Reflect.makeRecord(t, fields))
  }
  if s is Reflect.Array {
    if !(j is Array) {
      return Result.fail("want an array for ${t}")
    }
    let items = try (j.map({ decodeValue(s.element, x) }) | Result.all())
    return Result.ok(Reflect.makeArray(t, items))
  }
  if s is Reflect.Tagged {
    let base = try decodeValue(s.base, j)
    return Result.ok(Reflect.makeTagged(t, base))
  }
  if s is Reflect.Union {
    let member = s.members.find({ decodeValue(x, j).isOk() }) ?? return Result.fail("no member of ${t} fits")
    return decodeValue(member, j)
  }
  if j is Array || j is Map {
    return Result.fail("want ${t}")
  }
  let d = Reflect.makeScalar(t, j) ?? return Result.fail("${j} does not fit ${t}")
  Result.ok(d)
}

func decode[T](text: string, t: Reflect.Type[T]): Result[T, Error] {
  let j = try Json.parse(text)
  let d = try decodeValue(Reflect.erase(t), j)
  Result.ok(Reflect.to(d, t))
}

println(decode(`{"sensor": "a1", "value": 21.5}`, type(Reading)))
println(decode(`[{"sensor": "b2", "value": 19, "note": "roof"}]`, type(Array[Reading])))
println(decode(`{"sensor": "c3"}`, type(Reading)))
println(decode(`{"sensor": 7, "value": 1}`, type(Reading)))
println(decode(`[1, 2.5]`, type(Array[int])))
```

```text
Ok{ value: Reading{ sensor: "a1", value: Celsius(21.5), tags: [], note: null } }
Ok{ value: [Reading{ sensor: "b2", value: Celsius(19.0), tags: [], note: "roof" }] }
Err{ error: Error{ message: "missing field value" } }
Err{ error: Error{ message: "sensor: 7.0 does not fit string" } }
Err{ error: Error{ message: "2.5 does not fit int" } }
```

The result of `decode(text, type(Reading))` is a `Result[Reading, Error]`, so the caller reads its
fields with no further test.

### A query-string encoder

Encode any record as a query string. The function takes the type first, boxes the value once, and
reads each field along the record's type. An array field repeats its key, and a named primitive is
written as its base value:

```ant
type Page = int
type Search = { query: string, tags: Array[string], page: Page, exact: bool }

func leaf(d: Reflect.Value): string {
  let s = Reflect.type(d) | Reflect.shape()
  if s is Reflect.Tagged {
    return leaf(Reflect.payload(d))
  }
  String.from(Reflect.scalar(d))
}

func toQuery[T](t: Reflect.Type[T], value: T): string {
  let d = Reflect.valueOf(t, value)
  let s = Reflect.type(d) | Reflect.shape()
  if !(s is Reflect.Record) {
    panic("${t} is not a record type")
  }
  let pairs: Array[string] = []
  for f in s.fields {
    let v = Reflect.field(d, f.name)
    if Reflect.shape(f.type) is Reflect.Array {
      for item in Reflect.items(v) {
        pairs.push("${f.name}=${leaf(item)}")
      }
    } else {
      pairs.push("${f.name}=${leaf(v)}")
    }
  }
  pairs | String.join("&")
}

let search = Search{ query: "ant", tags: ["lang", "go"], page: Page(2), exact: false }
println(toQuery(type(Search), search))
```

```text
query=ant&tags=lang&tags=go&page=2&exact=false
```

### A table of codecs keyed by type

Type values hash, so a library keeps special cases in a map keyed by type value and falls back to a
general rule for the rest:

```ant
let names: Map[Reflect.ErasedType, string] = Map{
  Reflect.erase(type(Bytes)): "base64",
  Reflect.erase(type(Time.Instant)): "RFC 3339",
}
func format(t: Reflect.ErasedType): string {
  names.get(t) ?? "default"
}
println(format(Reflect.erase(type(Bytes))))
println(format(Reflect.erase(type(int))))
```

```text
base64
default
```

A codec that covers every instantiation of a generic type keys on its `Reflect.Decl` instead. Two
instantiations share one declaration, so `Box[string]` finds the entry made for `Box[int]`:

```ant
type Box[T] = { value: T }

func declOf(t: Reflect.ErasedType): Reflect.Decl | null {
  let s = Reflect.shape(t)
  if s is Reflect.Record { s.decl } else { null }
}

let boxDecl = declOf(Reflect.erase(type(Box[int]))) ?? panic("Box has a declaration")
let codecs: Map[Reflect.Decl, string] = Map{}
codecs[boxDecl] = "box codec"
let d = declOf(Reflect.erase(type(Box[string]))) ?? panic("Box has a declaration")
println(boxDecl)
println(codecs.get(d) ?? "default")
```

```text
Box
box codec
```

## Properties and relations

Boxing a value and taking it out at the same type gives the value back, and a box's type is never a
union: a box made at a union takes the member its value belongs to:

```ant
type Ids = Array[int]
let ids: Ids = [1, 2]
let d = Reflect.valueOf(type(Ids), ids)
assert(Reflect.to(d, type(Ids)) == ids)
let n = Reflect.valueOf(type(int | null), 42)
assert(Reflect.type(n) == Reflect.erase(type(int)))
```

Each type builder gives the same type value as writing the larger type out:

```ant
type Config = { name: string }
assert(Reflect.arrayOf(type(Config)) == type(Array[Config]))
assert(Reflect.nullable(type(Config)) == type(Config | null))
assert(Reflect.mapOf(type(string), type(int)) == type(Map[string, int]))
```

A value assigns to every type that `Reflect.assignable` accepts, so `Reflect.to` succeeds there:

```ant
let d = Reflect.valueOf(type(int), 42)
assert(Reflect.assignable(Reflect.type(d), Reflect.erase(type(int | null))))
assert(Reflect.to(d, type(int | null)) == 42)
```

## Possible issues

`type(T)` takes a type, not a value. It does not ask a value for its type:

```ant
type Config = { name: string }
let config = Config{ name: "api" }
let t = type(config) // error: type(...) takes a type, not a value; config is a value
```

Inside a generic function, `type(T)` cannot name the function's type parameter. Take a
`Reflect.Type[T]` parameter instead:

```ant
func elements[T](xs: Array[T]): string {
  String.from(type(Array[T]))
  // error: type(...) cannot mention T, a type parameter of the enclosing function or type, since nothing is bound to it at run time; take a Reflect.Type[T] parameter and build the type from it with Reflect.arrayOf, Reflect.setOf, Reflect.mapOf, or Reflect.nullable
}
```

```ant
func elements[T](t: Reflect.Type[T], xs: Array[T]): string {
  String.from(Reflect.arrayOf(t))
}
elements(type(int), [1, 2]) // "Array[int]"
```

An `if` cannot choose between two typed type values: `Reflect.Type[int64]` and `Reflect.Type[int32]`
cannot share a union. Erase both branches:

```ant
let wide = true
let t = if wide { type(int64) } else { type(int32) }
// error: the branches of this if do not form a valid type: a union may hold at most one type with no discriminator, and Reflect.Type[int] and Reflect.Type[int32] both have none; narrow first or give the whole an expected type
```

```ant
let wide = true
let t = if wide { Reflect.erase(type(int64)) } else { Reflect.erase(type(int32)) }
t // int
```

`type(T)` cannot reach an interface, either directly or through a field, because an interface has no
run-time identity. Reflect a union of named records instead:

```ant
interface Shape {
  area(s: Self): float64
}
type Scene = { shapes: Array[Shape] }
let t = type(Scene) // error: type(...) cannot reach an interface: Shape is one, and an interface has no run-time identity
```

A generic type needs its type arguments in `type(T)`:

```ant
type Box[T] = { value: T }
let t = type(Box) // error: Box takes 1 type argument, got 0
```

```ant
type Box[T] = { value: T }
type(Box[int]) // Box[int]
```

`Reflect.to` panics when the box's type does not assign to the type asked for. Test first with
`Reflect.assignable` when the box comes from input:

```ant
let d = Reflect.valueOf(type(int), 42)
Reflect.to(d, type(string)) // panic: Reflect.to: int is not assignable to string
```

## See also

[Reflect.shape](shape.md) · [Reflect.valueOf](valueOf.md) · [Reflect.to](to.md) ·
[Reflect.makeRecord](makeRecord.md) · [Reflect.Shape](types/Shape.md) ·
[Reflect.Value](types/Value.md) · [type](../../language/keywords/type.md) ·
[Types at run time](../../guide/runtime-types.md) · [Reflection](../guides/reflection.md) ·
[Json](../Json/index.md)
