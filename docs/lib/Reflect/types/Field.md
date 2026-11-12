<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Field

```ant signature
type Reflect.Field = { name: string, type: Reflect.ErasedType, hasDefault: bool }
```

Reflect.Field is one field of a record shape: its name, its type, and whether the declaration gives
it a default, so that a literal or [`Reflect.makeRecord`](../makeRecord.md) may leave it out.

<!-- end generated -->

## Details

- A `Reflect.Field` comes from the `fields` of a [`Reflect.Record`](Shape.md) shape, which you reach
  by narrowing [`Reflect.shape`](../shape.md) with `is Reflect.Record`. The fields come in
  declaration order.
- `type` is the field's type with the record's type arguments substituted: in `Box[int]`, a field
  declared `T` has type `int`.
- `hasDefault` says only whether a literal may leave the field out. Reflection does not give the
  default value; build a record with [`Reflect.makeRecord`](../makeRecord.md) to get it.
- Every call to `Reflect.shape` builds new `Reflect.Field` records, so changing one changes nothing
  else.

## Basic examples

List the fields of a record type with their types and defaults:

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

A `Reflect.Field` is an ordinary record and prints as one:

```ant
type Point = { x: int, y: int }

let s = Reflect.shape(Reflect.erase(type(Point)))
if s is Reflect.Record {
  println(s.fields[0])
  println(s.fields[1].name)
}
```

```text
Reflect.Field{ name: "x", type: int, hasDefault: false }
y
```

## Scope

In a generic record, `type` is the type argument, not the parameter:

```ant
type Box[T] = { value: T, label: string | null }

func fieldsOf(t: Reflect.ErasedType): Array[Reflect.Field] {
  let s = Reflect.shape(t)
  if s is Reflect.Record { s.fields } else { [] }
}

println(fieldsOf(Reflect.erase(type(Box[int])))[0])
println(fieldsOf(Reflect.erase(type(Box[bool])))[0])
println(fieldsOf(Reflect.erase(type(Box[int])))[1])
```

```text
Reflect.Field{ name: "value", type: int, hasDefault: false }
Reflect.Field{ name: "value", type: bool, hasDefault: false }
Reflect.Field{ name: "label", type: string | null, hasDefault: false }
```

An anonymous record type has fields too. It cannot declare defaults, so `hasDefault` is always
false:

```ant
let s = Reflect.shape(Reflect.erase(type({ id: int, done: bool })))
if s is Reflect.Record {
  println(s.fields.map({ "${x.name} ${x.hasDefault}" }) | String.join(", "))
}
```

```text
id false, done false
```

A record type with no fields has an empty `fields` array:

```ant
type Marker = {}

let s = Reflect.shape(Reflect.erase(type(Marker)))
if s is Reflect.Record {
  println(s.fields.length())
}
```

```text
0
```

`type` is a `Reflect.ErasedType`, one level deep. Pass it to `Reflect.shape` to look inside a field
that holds a record:

```ant
type Point = { x: int, y: int }
type Line = { from: Point, to: Point }

let s = Reflect.shape(Reflect.erase(type(Line)))
if s is Reflect.Record {
  let to = s.fields[1]
  println(to.type)
  let inner = Reflect.shape(to.type)
  if inner is Reflect.Record {
    println(inner.fields.map({ x.name }) | String.join(", "))
  }
}
```

```text
Point
x, y
```

You can write a `Reflect.Field` with a literal, as for any record:

```ant
let f = Reflect.Field{ name: "id", type: Reflect.erase(type(int)), hasDefault: false }
f      // Reflect.Field{ name: "id", type: int, hasDefault: false }
f.type // int
```

## Applications

Find the fields a caller must give, the ones without a default:

```ant
type Server = { host: string, port: int = 8080, user: string, debug: bool = false }

func required(t: Reflect.ErasedType): Array[string] {
  let s = Reflect.shape(t)
  if !(s is Reflect.Record) {
    return []
  }
  s.fields | Array.filter({ !x.hasDefault }) | Array.map({ x.name })
}

println(required(Reflect.erase(type(Server))))
```

```text
["host", "user"]
```

Check a set of input keys against a record type before building a value, reporting unknown and
missing fields:

```ant
type Server = { host: string, port: int = 8080 }

func checkKeys(t: Reflect.ErasedType, keys: Array[string]): Array[string] {
  let s = Reflect.shape(t)
  if !(s is Reflect.Record) {
    return ["${t} is not a record type"]
  }
  let names = s.fields.map({ x.name })
  let unknown = keys
    | Array.filter({ !names.contains(x) })
    | Array.map({ "unknown field ${x}" })
  let missing = s.fields
    | Array.filter({ !x.hasDefault && !keys.contains(x.name) })
    | Array.map({ "missing field ${x.name}" })
  [...unknown, ...missing]
}

let t = Reflect.erase(type(Server))
println(checkKeys(t, ["host"]))
println(checkKeys(t, ["port", "hots"]))
```

```text
[]
["unknown field hots", "missing field host"]
```

Find the fields of a given type, here every `string` field:

```ant
type User = { name: string, age: int, email: string, admin: bool = false }

let s = Reflect.shape(Reflect.erase(type(User)))
if s is Reflect.Record {
  let text = Reflect.erase(type(string))
  println(s.fields | Array.filter({ x.type == text }) | Array.map({ x.name }))
}
```

```text
["name", "email"]
```

Read each field of a boxed record by name, as an encoder does:

```ant
type Server = { host: string, port: int = 8080 }

func describe(d: Reflect.Value): string {
  let s = Reflect.shape(Reflect.type(d))
  if !(s is Reflect.Record) {
    return "${d}"
  }
  s.fields
    | Array.map({ "${x.name}=${Reflect.field(d, x.name)}" })
    | String.join(" ")
}

println(describe(Reflect.valueOf(type(Server), Server{ host: "localhost" })))
```

```text
host=Reflect.Value("localhost": string) port=Reflect.Value(8080: int)
```

## Properties and relations

`type` compares with `==` to the type value of the field's declared type, and the fields come in
declaration order:

```ant
type Point = { x: int, y: float64 }

let s = Reflect.shape(Reflect.erase(type(Point)))
if s is Reflect.Record {
  assert(s.fields.map({ x.name }) == ["x", "y"])
  assert(s.fields[0].type == Reflect.erase(type(int)))
  assert(s.fields[1].type == Reflect.erase(type(float64)))
}
```

Each call to `Reflect.shape` gives new fields, so changing one does not change the next call's:

```ant
type Point = { x: int, y: int }

let t = Reflect.erase(type(Point))
let first = Reflect.shape(t)
if first is Reflect.Record {
  first.fields[0].name = "renamed"
}
let second = Reflect.shape(t)
if second is Reflect.Record {
  assert(second.fields[0].name == "x")
}
```

## Possible issues

`fields` is only on the `Reflect.Record` member of the shape. Narrow the shape with `is` first:

```ant
type Server = { host: string }
let fs = Reflect.shape(Reflect.erase(type(Server))).fields
// error: Reflect.Shape has no field fields on every member: narrow the value first
```

`hasDefault` does not hold the default. To read the defaults, build a record from the required
fields with [`Reflect.makeRecord`](../makeRecord.md) and read its fields:

```ant
type Server = { host: string, port: int = 8080 }

let t = Reflect.erase(type(Server))
let host = Reflect.valueOf(type(string), "")
let defaults = Reflect.makeRecord(t, Map{ host })
println(Reflect.field(defaults, "port"))
```

```text
Reflect.Value(8080: int)
```

`Json.stringify` panics on a `Reflect.Field`, since its `type` is a type value, which has no JSON
form. Encode `String.from(f.type)` instead:

```ant
let f = Reflect.Field{ name: "id", type: Reflect.erase(type(int)), hasDefault: false }
Json.stringify(f)                                              // panic: Json.stringify: a type has no JSON form
Json.stringify(Map{ name: f.name, type: String.from(f.type) }) // "{\"name\":\"id\",\"type\":\"int\"}"
```

## See also

[Reflect.shape](../shape.md) · [Reflect.Shape](Shape.md) · [Reflect.ErasedType](ErasedType.md) ·
[Reflect.field](../field.md) · [Reflect.makeRecord](../makeRecord.md)
