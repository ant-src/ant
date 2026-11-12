<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeRecord

```ant signature
func Reflect.makeRecord(t: Reflect.ErasedType, fields: Map[string, Reflect.Value]): Reflect.Value
```

Reflect.makeRecord returns a new record of the named record type t with the given fields, each field
left out taking its default as a literal evaluates it. The record carries the tag, so is and == work
on it as on any literal. It panics if t is not a named record type, a key is not one of its fields,
a field's type is not assignable to the field's type, or a field with no default is left out.

<!-- end generated -->

## Details

- Each value in `fields` keeps its own type: an `int` is not converted for a `float64` field. Box
  numbers at the field's type, for example with [`Reflect.makeScalar`](makeScalar.md).

## Basic examples

Build a `Point` from boxed field values:

```ant
type Point = { x: float64, y: float64 }
let t = Reflect.erase(type(Point))
let x = Reflect.valueOf(type(float64), 1.5)
let y = Reflect.valueOf(type(float64), -2.0)
Reflect.makeRecord(t, Map{ x, y }) // Reflect.Value(Point{ x: 1.5, y: -2.0 }: Point)
```

Unbox the result with [`Reflect.to`](to.md) to use it as an ordinary record:

```ant
type Point = { x: float64, y: float64 }
let fields = Map{ x: Reflect.valueOf(type(float64), 3.0), y: Reflect.valueOf(type(float64), 4.0) }
let p = Reflect.makeRecord(Reflect.erase(type(Point)), fields) | Reflect.to(type(Point))
p   // Point{ x: 3.0, y: 4.0 }
p.x // 3.0
```

Leave out a field that has a default:

```ant
type Server = { host: string, port: int = 8080, tags: Array[string] = [] }
let host = Reflect.valueOf(type(string), "example.com")
Reflect.makeRecord(Reflect.erase(type(Server)), Map{ host }) // Reflect.Value(Server{ host: "example.com", port: 8080, tags: [] }: Server)
```

## Scope

An empty map builds a record type with no fields, or one whose fields all have defaults:

```ant
type Marker = {}
type Options = { verbose: bool = false, depth: int = 3 }
Reflect.makeRecord(Reflect.erase(type(Marker)), Map{})  // Reflect.Value(Marker{}: Marker)
Reflect.makeRecord(Reflect.erase(type(Options)), Map{}) // Reflect.Value(Options{ verbose: false, depth: 3 }: Options)
```

A generic record type takes its type arguments from `t`, and each field is checked at its
substituted type:

```ant
type Box[T] = { value: T }
let n = Reflect.valueOf(type(int), 7)
Reflect.makeRecord(Reflect.erase(type(Box[int])), Map{ value: n }) // Reflect.Value(Box{ value: 7 }: Box[int])
```

A field of a union type takes a box of any member:

```ant
type Reply = { code: int | null }
let t = Reflect.erase(type(Reply))
Reflect.makeRecord(t, Map{ code: Reflect.valueOf(type(int), 200) })   // Reflect.Value(Reply{ code: 200 }: Reply)
Reflect.makeRecord(t, Map{ code: Reflect.valueOf(type(null), null) }) // Reflect.Value(Reply{ code: null }: Reply)
```

A field of a named primitive type takes a box built with [`Reflect.makeTagged`](makeTagged.md):

```ant
type Celsius = float64
type Reading = { sensor: string, value: Celsius }
let value = Reflect.makeTagged(Reflect.erase(type(Celsius)), Reflect.valueOf(type(float64), 21.5))
let sensor = Reflect.valueOf(type(string), "a1")
Reflect.makeRecord(Reflect.erase(type(Reading)), Map{ sensor, value }) | Reflect.to(type(Reading)) // Reading{ sensor: "a1", value: Celsius(21.5) }
```

A field of record type takes a box of a record, built by reflection or not. A field of anonymous
record type takes any named record with the fields it lists:

```ant
type Point = { x: float64, y: float64 }
type Label = { text: string, at: { x: float64 } }
let at = Reflect.valueOf(type(Point), Point{ x: 1.0, y: 2.0 })
let text = Reflect.valueOf(type(string), "origin")
Reflect.makeRecord(Reflect.erase(type(Label)), Map{ text, at }) // Reflect.Value(Label{ text: "origin", at: Point{ x: 1.0, y: 2.0 } }: Label)
```

## Applications

Decode a JSON object into any flat record type. [`Reflect.shape`](shape.md) lists the fields,
[`Reflect.makeScalar`](makeScalar.md) checks and boxes each value at its field's type, and a field
that is missing from the object takes its default:

```ant
type Server = { host: string, port: int = 8080, debug: bool = false }

func decode[T](t: Reflect.Type[T], text: string): Result[T, Error] {
  let erased = Reflect.erase(t)
  let shape = Reflect.shape(erased)
  if !(shape is Reflect.Record) {
    return Result.fail("${erased} is not a record type")
  }
  let json = try Json.parse(text)
  if !(json is Map) {
    return Result.fail("expected an object")
  }
  let fields: Map[string, Reflect.Value] = Map{}
  for f in shape.fields {
    let v = json.get(f.name)
    if v == null && f.hasDefault {
      continue
    }
    if v is Array || v is Map {
      return Result.fail("${f.name}: expected ${f.type}")
    }
    let boxed = Reflect.makeScalar(f.type, v)
    if boxed == null {
      return Result.fail("${f.name}: expected ${f.type}")
    }
    fields[f.name] = boxed
  }
  Result.ok(Reflect.makeRecord(erased, fields) | Reflect.to(t))
}

println(decode(type(Server), `{"host": "example.com", "port": 9000}`))
println(decode(type(Server), `{"host": "example.com", "port": "http"}`))
```

```text
Ok{ value: Server{ host: "example.com", port: 9000, debug: false } }
Err{ error: Error{ message: "port: expected int" } }
```

List the fields of a record that differ from their defaults, as an encoder that leaves out default
values does. Build one record of defaults from the record's required fields, then compare field by
field:

```ant
type Server = { host: string, port: int = 8080, debug: bool = false }

func changed[T](t: Reflect.Type[T], value: T): Array[string] {
  let shape = Reflect.shape(Reflect.erase(t))
  if !(shape is Reflect.Record) {
    return []
  }
  let d = Reflect.valueOf(t, value)
  let required = shape.fields
    | Array.filter({ !x.hasDefault })
    | Array.map({ Entry{ key: x.name, value: d.field(x.name) } })
    | Map.fromEntries()
  let defaults = Reflect.makeRecord(Reflect.erase(t), required)
  shape.fields
    | Array.filter({ x.hasDefault && d.field(x.name) != defaults.field(x.name) })
    | Array.map({ x.name })
}

println(changed(type(Server), Server{ host: "a" }))
println(changed(type(Server), Server{ host: "a", debug: true }))
```

```text
[]
["debug"]
```

## Properties and relations

The result is boxed at `t` and is `==` to the literal with the same fields:

```ant
type Point = { x: int, y: int }
let t = Reflect.erase(type(Point))
let fields = Map{ x: Reflect.valueOf(type(int), 1), y: Reflect.valueOf(type(int), 2) }
let d = Reflect.makeRecord(t, fields)
assert(Reflect.type(d) == t)
assert(d == Reflect.valueOf(type(Point), Point{ x: 1, y: 2 }))
```

The record carries its tag, so `is` picks it out of a union:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
let r = Reflect.valueOf(type(float64), 1.0)
let figure: Circle | Square = Reflect.makeRecord(Reflect.erase(type(Circle)), Map{ r })
  | Reflect.to(type(Circle | Square))
assert(figure is Circle)
```

Rebuilding a record from its own fields gives an equal record, but a new one:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let d = Reflect.valueOf(type(Point), p)
let fields = Map{ x: d.field("x"), y: d.field("y") }
let copy = Reflect.makeRecord(Reflect.erase(type(Point)), fields) | Reflect.to(type(Point))
assert(copy == p)
assert(!Record.same(copy, p))
```

A default is evaluated for each record, so two records do not share a default array:

```ant
type Server = { host: string, tags: Array[string] = [] }
let t = Reflect.erase(type(Server))
let host = Reflect.valueOf(type(string), "a")
let a = Reflect.makeRecord(t, Map{ host }) | Reflect.to(type(Server))
let b = Reflect.makeRecord(t, Map{ host }) | Reflect.to(type(Server))
a.tags.push("x")
assert(b.tags.isEmpty())
```

A default is evaluated when `Reflect.makeRecord` runs, so a default that reads changing state can
fill the same field differently in two records:

```ant
var next = 1

func id(): int {
  next += 1
  next
}

type Job = { name: string, id: int = id() }

let t = Reflect.erase(type(Job))
let name = Reflect.valueOf(type(string), "build")
let a = Reflect.makeRecord(t, Map{ name }) | Reflect.to(type(Job))
let b = Reflect.makeRecord(t, Map{ name }) | Reflect.to(type(Job))
println(a.id)
println(b.id)
```

```text
2
3
```

## Possible issues

An anonymous record type has no tag to give the record, so it panics. Build a named record type that
the anonymous one accepts:

```ant
Reflect.makeRecord(Reflect.erase(type({ x: int })), Map{})
// panic: Reflect.makeRecord needs a named record type, got the anonymous record type { x: int }; build a named record type assignable to it
```

A union type panics, even `T | null` around a record type. Pass the record type itself:

```ant
type Point = { x: int, y: int }
Reflect.makeRecord(Reflect.erase(type(Point | null)), Map{})
// panic: Reflect.makeRecord needs a named record type, got Point | null
```

A field with no default must be given, and every key must name a field:

```ant
type Point = { x: int, y: int }
let t = Reflect.erase(type(Point))
let one = Reflect.valueOf(type(int), 1)
Reflect.makeRecord(t, Map{ x: one })                 // panic: Reflect.makeRecord: Point is missing the field y, which has no default
Reflect.makeRecord(t, Map{ x: one, y: one, z: one }) // panic: Reflect.makeRecord: Point has no field z
```

A box is not converted to the field's type. An `int` does not fit a `float64` field; box the number
at the field's type instead, as `Reflect.makeScalar` does for a value read from data:

```ant
type Point = { x: float64, y: float64 }
let t = Reflect.erase(type(Point))
let y = Reflect.valueOf(type(float64), 0.0)
Reflect.makeRecord(t, Map{ x: Reflect.valueOf(type(int), 1), y })
// panic: Reflect.makeRecord: a value of type int is not assignable to the field x of Point, whose type is float64
Reflect.makeRecord(t, Map{ x: Reflect.valueOf(type(float64), 1.0), y }) // Reflect.Value(Point{ x: 1.0, y: 0.0 }: Point)
```

## See also

[Reflect.shape](shape.md) · [Reflect.field](field.md) · [Reflect.to](to.md) ·
[Reflect.valueOf](valueOf.md) · [Reflect.makeScalar](makeScalar.md) ·
[Reflect.makeTagged](makeTagged.md) · [Reflect.makeArray](makeArray.md) ·
[Reflect.makeMap](makeMap.md) · [Record.same](../Record/same.md)
