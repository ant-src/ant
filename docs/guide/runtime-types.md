<!-- status: reviewed -->

# Types at run time

Ant checks types with one command and runs programs with another. `ant check` reads every annotation
and reports compile errors; `ant run` discards the annotations and runs the values. What survives
into a running program is small: each value knows its kind, a record or a named primitive knows its
tag, and a type you write as `type(T)` becomes a value the program can take apart. This chapter
shows what each of these gives you, and how they combine into typed decoding of data such as JSON.

## Checked, then erased

An annotation is advice for the checker. It does not convert a value, and `ant run` does not test
it:

```ant
let port: int = "8080" // error: cannot use string where int is expected
println(port)
println(port + 1)
```

`ant check` rejects the first line. `ant run` never checks, so it runs the program anyway: the first
`println` prints the string, and the program panics only at the `+`, where the operation itself
needs two values of one kind:

```sh
ant run main.ant
```

```text
8080
panic: + needs two numbers of one kind or two strings, got string and int
  at module main.ant (main.ant:3:14)
```

So run `ant check` in your editor and in continuous integration. No flag makes `ant run` check
first; the checker is what reports these mistakes before the program runs.

## What a value knows

Every value carries its concrete kind, such as `int`, `string`, array, or map. A record, and a value
of a named primitive type such as `type UserId = string`, also carries the tag of its declaration.
`is` tests these at run time, which is how a union narrows. A JSON tree is a union of six kinds:

```ant
func describe(v: Json.Value): string {
  if v is Array { return "array of ${v.length()}" }
  if v is Map { return "object with ${v.keys() | String.join(", ")}" }
  if v is float64 { return "number" }
  "other"
}

println(describe(Result.unwrap(Json.parse(`{"a": 1, "b": [2]}`))))
```

```text
object with a, b
```

`is` has only kinds and tags to look at, and a tag names a declaration, not its type arguments. So
`is` takes a kind or a declared name, written without arguments. Test the elements instead:

```ant
let v: Json.Value = [1.0, 2.0]
let ok = v is Array[float64]
// error: 'is' tests a kind or a tag, written without type arguments: write is Array, then test the elements
```

`is` does not take an alias either, since no value carries an alias's tag. With
`type Id = int | string`, test the members `int` and `string`, not `Id`.

## Types as values

Inside an expression a type name is not a value. `type(T)` is: it makes the type `T` a value of type
`Reflect.Type[T]`. It takes any type an annotation can hold, except a type parameter of the
enclosing function (below) and an interface.

```ant
type Point = { x: int, y: int }
type MaybePoint = Point | null

type(Point)                             // Point
type(Array[Point] | null)               // Array[Point] | null
type(int) == type(int64)                // true
type(MaybePoint) == type(null | Point)  // true
Reflect.shape(Reflect.erase(type(int))) // Reflect.Int{ bits: 64, signed: true }
```

Two type values are `==` when they denote the same type, with aliases expanded and union members
compared as a set. They also hash, so a type value can be a map key, which is how a library keeps a
table of one handler per type.

`type` asks for a type, not for the type of a value. Unlike `typeof x` in JavaScript or `type(x)` in
Python, `type(config)` on a variable is a compile error. A value already tells you its kind through
`is`.

```ant
type Config = { name: string }
let config = Config{ name: "app" }
let t = type(config) // error: type(...) takes a type, not a value; config is a value
```

Nor does it take an interface:

```ant
interface Shape { area(s: Self): float64 }
let t = type(Shape) // error: type(...) cannot reach an interface: Shape is one, and an interface has no run-time identity
```

An interface has no run-time identity, so a decoder targets a union of the concrete types instead.

### In generic code

A type parameter is erased too: nothing is bound to `T` at run time, so `type(T)` inside a generic
function is a compile error. Take a `Reflect.Type[T]` parameter instead, and let the caller write
`type(Config)`. The parameter also fixes `T`, so the caller gets a `Config` back without an
annotation.

```ant
func emptyList[T](t: Reflect.Type[T]): Reflect.Type[Array[T]] {
  type(Array[T])
  // error: type(...) cannot mention T, a type parameter of the enclosing function or type, since nothing is bound to it at run time; take a Reflect.Type[T] parameter and build the type from it with Reflect.arrayOf, Reflect.setOf, Reflect.mapOf, or Reflect.nullable
}
```

The builders `Reflect.arrayOf`, `Reflect.setOf`, `Reflect.mapOf`, and `Reflect.nullable` compose a
larger type from the one you hold, and keep its static type:

```ant
func listOf[T](t: Reflect.Type[T]): Reflect.Type[Array[T]] {
  Reflect.arrayOf(t)
}

listOf(type(int)) == type(Array[int]) // true
Reflect.nullable(type(string))        // string | null
```

A few library functions take a type value to choose what to read or write, such as
[Bytes.read](../lib/Bytes/read.md), where `type(uint16)` picks the width and the result type.

## Reflect

The [Reflect](../lib/Reflect/index.md) namespace does three things with type values.

**It takes a type apart, one level at a time.** `Reflect.shape` returns a member of `Reflect.Shape`,
which you narrow with `is`. It works on a `Reflect.ErasedType`, a type value whose static type is
not known, because the type of a field or an element is not known statically one level down.
`Reflect.erase` gives up the static type of a `Reflect.Type[T]`, once, where the walk starts:

```ant
type Server = { host: string, port: int = 8080, tags: Array[string] = [] }

let s = Reflect.shape(Reflect.erase(type(Server)))
if s is Reflect.Record {
  for f in s.fields {
    println("${f.name}: ${f.type}, default: ${f.hasDefault}")
  }
}
```

```text
host: string, default: false
port: int, default: true
tags: Array[string], default: true
```

**It boxes a value with its type, and reads the box along that type.** `Reflect.valueOf` makes a
`Reflect.Value`; `Reflect.field`, `Reflect.items`, `Reflect.entries`, `Reflect.payload`, and
`Reflect.scalar` read its parts as boxes of their own. `Reflect.to` gives the value back at a static
type, and panics when the box's type is not assignable to it.

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Point | null), Point{ x: 1, y: 2 })
d                           // Reflect.Value(Point{ x: 1, y: 2 }: Point)
Reflect.field(d, "x")       // Reflect.Value(1: int)
Reflect.to(d, type(Point))  // Point{ x: 1, y: 2 }
Reflect.to(d, type(string)) // panic: Reflect.to: Point is not assignable to string
```

The box made at `Point | null` holds a `Point`: a box always has the one member its value belongs
to, so code that reads it never chooses between members. A walk over the readers is the start of an
encoder:

```ant
type Celsius = float64
type Reading = { sensor: string, value: Celsius, tags: Array[string] = [] }

func render(d: Reflect.Value): string {
  let s = Reflect.shape(Reflect.type(d))
  if s is Reflect.Record {
    let parts = s.fields.map({ "${x.name}=${render(Reflect.field(d, x.name))}" })
    return "{${parts | String.join(" ")}}"
  }
  if s is Reflect.Array {
    return "[${Reflect.items(d).map(render) | String.join(" ")}]"
  }
  if s is Reflect.Tagged { return render(Reflect.payload(d)) }
  String.from(Reflect.scalar(d))
}

let r = Reading{ sensor: "a1", value: Celsius(21.5), tags: ["roof"] }
println(render(Reflect.valueOf(type(Reading), r)))
```

```text
{sensor=a1 value=21.5 tags=[roof]}
```

**It builds new values from boxed parts.** `Reflect.makeRecord`, `Reflect.makeArray`,
`Reflect.makeSet`, `Reflect.makeMap`, and `Reflect.makeTagged` build a value of a type known only at
run time, and check each part against its slot. `Reflect.makeScalar` boxes a leaf, and returns
`null` when the leaf does not fit the type. The other builders panic when a part does not fit, since
that is a bug in the decoder, not in the input; `Reflect.assignable` tests a part first. Only
`makeScalar` reports bad input, by returning `null`. The next section uses them.

## Typed decoding

`Json.parse` gives a `Json.Value`: a tree of `null`, `bool`, `float64`, `string`, arrays, and maps.
An annotation cannot turn that tree into your record, since annotations convert nothing:

```ant
type Config = { name: string, port: int }

let tree = Result.unwrap(Json.parse(`{"name": "app", "port": 8080}`))
let config: Config = tree // error: cannot use Json.Value where Config is expected
```

A decoder builds the record. There are two ways to write one.

### A function per type

For one or two types, write the function. Narrow the tree with `is`, convert numbers with
`Math.toInt`, and return an error for input that does not fit:

```ant
type Server = { host: string, port: int }

func decodeServer(v: Json.Value): Result[Server, Error] {
  if !(v is Map) { return Result.fail("want an object") }
  let host = v["host"]
  if !(host is string) { return Result.fail("host: want a string") }
  let port = v["port"]
  if !(port is float64) { return Result.fail("port: want a number") }
  let n = Math.toInt(port)
  if n == null { return Result.fail("port: want an integer") }
  Result.ok(Server{ host, port: n })
}

func parseServer(text: string): Result[Server, Error] {
  decodeServer(try Json.parse(text))
}

println(parseServer(`{"host": "a.example", "port": 8080}`))
println(parseServer(`{"host": "a.example", "port": "80"}`))
```

```text
Ok{ value: Server{ host: "a.example", port: 8080 } }
Err{ error: Error{ message: "port: want a number" } }
```

Every rule is in plain sight, and the checker checks the record literal at the end. The cost is a
function per type, which you update with every field.

### One decoder for every type

For many types, write one decoder that walks the target type with `Reflect.shape` and builds the
value with the builders. The module below handles records, arrays, `T | null`, named primitives, and
scalars, and reports where in the input a value does not fit:

```ant [decode.ant]
let nullType = Reflect.erase(type(null))

func decodeValue(v: Json.Value, t: Reflect.ErasedType, path: string): Result[Reflect.Value, Error] {
  let s = Reflect.shape(t)
  if s is Reflect.Record {
    if !(v is Map) { return Result.fail("${path}: want an object for ${t}") }
    let fields: Map[string, Reflect.Value] = Map{}
    for f in s.fields {
      if v.has(f.name) {
        fields[f.name] = try decodeValue(v[f.name], f.type, "${path}.${f.name}")
      } else if !f.hasDefault {
        return Result.fail("${path}: missing field ${f.name}")
      }
    }
    return Result.ok(Reflect.makeRecord(t, fields))
  }
  if s is Reflect.Array {
    if !(v is Array) { return Result.fail("${path}: want an array for ${t}") }
    let items: Array[Reflect.Value] = []
    for { index, value } in v.entries() {
      items.push(try decodeValue(value, s.element, "${path}[${index}]"))
    }
    return Result.ok(Reflect.makeArray(t, items))
  }
  if s is Reflect.Union {
    let other = s.members.find({ x != nullType })
    if other == null || s.members.length() != 2 || !s.members.contains(nullType) {
      return Result.fail("${path}: cannot decode ${t}")
    }
    return decodeValue(v, if v == null { nullType } else { other }, path)
  }
  if s is Reflect.Tagged {
    return Result.ok(Reflect.makeTagged(t, try decodeValue(v, s.base, path)))
  }
  if v is Array || v is Map { return Result.fail("${path}: want ${t}") }
  let leaf = Reflect.makeScalar(t, v)
  if leaf == null { return Result.fail("${path}: want ${t}, got ${Json.stringify(v)}") }
  Result.ok(leaf)
}

export func decode[T](text: string, t: Reflect.Type[T]): Result[T, Error] {
  let tree = try Json.parse(text)
  let d = try decodeValue(tree, Reflect.erase(t), "$")
  Result.ok(Reflect.to(d, t))
}
```

The caller names the type once, and gets a value of that type back:

```ant
import "./decode.ant" { decode }

type Port = int
type Server = { host: string, port: Port = Port(8080), tags: Array[string] = [] }
type Config = { name: string, servers: Array[Server] }

let text = `{"name": "prod", "servers": [{"host": "a.example"}, {"host": "b.example", "port": 9000}]}`
let config = Result.unwrap(decode(text, type(Config)))
for server in config.servers {
  println(server)
}

println(decode(`{"name": "prod", "servers": [{"port": 1}]}`, type(Config)))
println(decode(`{"name": "prod", "servers": [{"host": "b", "port": 80.5}]}`, type(Config)))
println(decode(`[1, null, 3]`, type(Array[int | null])))
println(decode(`["a", 1]`, type(Array[int | string])))
```

```text
Server{ host: "a.example", port: Port(8080), tags: [] }
Server{ host: "b.example", port: Port(9000), tags: [] }
Err{ error: Error{ message: "$.servers[0]: missing field host" } }
Err{ error: Error{ message: "$.servers[0].port: want int, got 80.5" } }
Ok{ value: [1, null, 3] }
Err{ error: Error{ message: "$[0]: cannot decode int | string" } }
```

A few things make this short:

- `Reflect.makeScalar` is the whole test for a leaf. It converts the JSON number `8080` to an `int`
  and returns `null` for `80.5`, so the decoder needs no branch per number type.
- `Reflect.makeRecord` fills an omitted field from its default, so `port` is `8080` where the input
  leaves it out. The decoder reports only fields that have no default.
- `Reflect.to` at the end turns the box back into a `Config`. The builders checked every part on the
  way, so it cannot fail here.

Every policy is the decoder's own: what a missing field means, how a union picks its member, whether
an unknown key is an error. This one ignores unknown keys, decodes only `T | null` unions, and has
no case for maps; add the cases your format needs.

A generic function passes its `Reflect.Type[T]` through, building larger types with the builders:

```ant
import "./decode.ant" { decode }

type Point = { x: int, y: int }

func decodeList[T](text: string, t: Reflect.Type[T]): Result[Array[T], Error] {
  decode(text, Reflect.arrayOf(t))
}

println(decodeList(`[{"x": 1, "y": 2}]`, type(Point)))
```

```text
Ok{ value: [Point{ x: 1, y: 2 }] }
```

Encoding needs this less often. `Json.stringify` already writes a record as an object with its
fields; walk the value with the readers, as `render` does above, when the format needs a policy of
its own, such as leaving out fields equal to their defaults.

## Which to pick

- **Narrow with `is`** when you hold a union and want to act on its members. This is most code.
- **Write a decoder per type** for a few types, or when each type needs its own rules.
- **Walk `Reflect.shape`** when one piece of code must handle types it does not know in advance: a
  decoder or encoder for a format, a table printer, a configuration loader.
- **Take a `Reflect.Type[T]` parameter** in a generic function that needs the type at run time, and
  let the caller write `type(...)`.

## Habits from other languages

- **An annotation is not a cast.** TypeScript's `as Config` has no equivalent. A `Json.Value`
  becomes a `Config` only through code that builds one.
- **`ant run` does not check types.** A program with compile errors runs until an operation cannot
  go on, or to its end. Run `ant check` before you trust a program.
- **`type(x)` is not `typeof x`.** It takes a type and makes it a value. To ask what a value is, use
  `is`.
- **Type arguments are gone at run time.** `xs is Array[int]` is a compile error, and a generic
  function cannot see its `T`. Pass a `Reflect.Type[T]` when it must.
- **Reflection does not walk your variables.** A function value does not know its signature, and a
  variable does not know its annotation. Reflection starts from a type that the program writes.

## See also

- [Reflection](../lib/guides/reflection.md) and [Reflect](../lib/Reflect/index.md)
- [type](../language/keywords/type.md)
- [is](../language/operators/is.md)
- [Json](../lib/Json/index.md) and [Data formats](../lib/guides/data-formats.md)
- [Unions](unions.md)
- [Generics](generics.md)
- [ant check](../tools/check.md) and [ant run](../tools/run.md)
