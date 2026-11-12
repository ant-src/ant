<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.field

```ant signature
func Reflect.field(d: Reflect.Value, name: string): Reflect.Value
```

Reflect.field returns the field name of the record in d, boxed at the field's type. Through an
anonymous record type it reaches only the fields that type lists. It panics if d's type is not a
record type that lists the field.

<!-- end generated -->

## Details

- The box's type is the field's type with the record's type arguments substituted. When the field's
  type is a union, the box holds the member the value belongs to, never the union.
- A field that holds a record, array, map, or set is boxed as that same value, not a copy.

## Basic examples

Read the field `x` of a boxed point:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
Reflect.field(d, "x") // Reflect.Value(1: int)
Reflect.field(d, "y") // Reflect.Value(2: int)
```

Take the value out with `Reflect.to` when you know its type:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
Reflect.to(Reflect.field(d, "y"), type(int)) // 2
```

## Scope

A field that holds a record gives a box you can read further, through `.`:

```ant
type Point = { x: int, y: int }
type Line = { from: Point, to: Point }
let line = Line{ from: Point{ x: 0, y: 0 }, to: Point{ x: 3, y: 4 } }
let d = Reflect.valueOf(type(Line), line)
d.field("to")            // Reflect.Value(Point{ x: 3, y: 4 }: Point)
d.field("to").field("x") // Reflect.Value(3: int)
```

In a generic record, the field's type is the type argument:

```ant
type Box[T] = { value: T }
let d = Reflect.valueOf(type(Box[string]), Box{ value: "hi" })
d.field("value")                  // Reflect.Value("hi": string)
d.field("value") | Reflect.type() // string
```

A field of a union type is boxed at the member its value belongs to, here `string` or `null`:

```ant
type User = { name: string, email: string | null }
let ann = Reflect.valueOf(type(User), User{ name: "Ann", email: "ann@example.com" })
let bob = Reflect.valueOf(type(User), User{ name: "Bob", email: null })
ann.field("email") // Reflect.Value("ann@example.com": string)
bob.field("email") // Reflect.Value(null: null)
```

A `Result` is a union of the records `Ok` and `Err`, so a box of one is a box of the record, and its
fields are `value` and `error`:

```ant
let r: Result[int, Error] = Result.ok(5)
let d = Reflect.valueOf(type(Result[int, Error]), r)
d                // Reflect.Value(Ok{ value: 5 }: Ok[int])
d.field("value") // Reflect.Value(5: int)
```

Through an anonymous record type, only the fields that type lists are visible, even when the record
has more:

```ant
type User = { name: string, age: int }
let d = Reflect.valueOf(type({ name: string }), User{ name: "Ann", age: 30 })
d.field("name") // Reflect.Value("Ann": string)
d.field("age")  // panic: Reflect.field: { name: string } has no field age
```

A field of a named primitive type is boxed with its tag, so take the base value out with
`Reflect.payload` before `Reflect.scalar`:

```ant
type UserId = int
type User = { id: UserId, name: string }
let d = Reflect.valueOf(type(User), User{ id: UserId(7), name: "Ann" })
d.field("id")                                        // Reflect.Value(UserId(7): UserId)
d.field("id") | Reflect.payload() | Reflect.scalar() // 7
```

Records inside an array come out of `Reflect.items` boxed, ready for `field`:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Array[Point]), [Point{ x: 1, y: 2 }, Point{ x: 5, y: 6 }])
d.items().map({ x.field("y").to(type(int)) }) // [2, 6]
```

Call it in a pipe:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
d | Reflect.field("x") | Reflect.scalar() // 1
```

## Applications

Print every field of a record whose fields are scalars, by walking the fields of its shape:

```ant
type Server = { host: string, port: int, debug: bool }

func describe(d: Reflect.Value) {
  let s = d.type().shape()
  if !(s is Reflect.Record) {
    return
  }
  for f in s.fields {
    println("${f.name} = ${d.field(f.name).scalar()}")
  }
}

describe(Reflect.valueOf(type(Server), Server{ host: "localhost", port: 8080, debug: false }))
```

```text
host = localhost
port = 8080
debug = false
```

List the fields that differ between two values of one record type. Two boxes are `==` when their
types and values are:

```ant
type Config = { host: string, port: int, debug: bool }

func changed[T](t: Reflect.Type[T], a: T, b: T): Array[string] {
  let before = Reflect.valueOf(t, a)
  let after = Reflect.valueOf(t, b)
  let s = Reflect.shape(Reflect.erase(t))
  if !(s is Reflect.Record) {
    return []
  }
  s.fields
    .filter({ Reflect.field(before, x.name) != Reflect.field(after, x.name) })
    .map({ x.name })
}

let current = Config{ host: "localhost", port: 8080, debug: false }
let updated = Config{ host: "localhost", port: 9090, debug: true }
println(changed(type(Config), current, updated))
```

```text
["port", "debug"]
```

Follow a dotted path such as `"server.port"` through nested records:

```ant
type Server = { host: string, port: int }
type App = { name: string, server: Server }

func lookup(d: Reflect.Value, path: string): Reflect.Value {
  path.split(".").reduce(d, Reflect.field)
}

let app = App{ name: "shop", server: Server{ host: "localhost", port: 8080 } }
let d = Reflect.valueOf(type(App), app)
println(lookup(d, "server.port"))
println(lookup(d, "name"))
```

```text
Reflect.Value(8080: int)
Reflect.Value("shop": string)
```

## Properties and relations

Taking a field out with `Reflect.to` gives the same value as reading it directly:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let d = Reflect.valueOf(type(Point), p)
assert(d.field("x").to(type(int)) == p.x)
```

Every field's box has a type assignable to the field's type in the record's shape, and that type is
never a union:

```ant
type User = { name: string, email: string | null }
let d = Reflect.valueOf(type(User), User{ name: "Ann", email: null })
let s = d.type().shape()
assert(s is Reflect.Record)
if s is Reflect.Record {
  assert(s.fields.every({ Reflect.assignable(d.field(x.name).type(), x.type) }))
  assert(s.fields.every({ !(d.field(x.name).type().shape() is Reflect.Union) }))
}
```

A field that holds a record is boxed as that record itself, so two fields that share one record are
the same, by `Reflect.same`:

```ant
type Point = { x: int, y: int }
type Line = { from: Point, to: Point }
let p = Point{ x: 1, y: 2 }
let d = Reflect.valueOf(type(Line), Line{ from: p, to: p })
assert(Reflect.same(d.field("from"), d.field("to")))
```

The box refers to the record, so `field` reads the field as it is when called:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let d = Reflect.valueOf(type(Point), p)
p.x = 10
assert(d.field("x") == Reflect.valueOf(type(int), 10))
```

## Possible issues

A box of anything but a record panics:

```ant
let d = Reflect.valueOf(type(int), 3)
d.field("x") // panic: Reflect.field needs a value of a record type, got one of type int
```

A name that is not a field of the type panics. Names are case-sensitive:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
d.field("X") // panic: Reflect.field: Point has no field X
```

A box of a union value has the member's type, so the fields of the other members are not there. An
`Ok` has no `error`:

```ant
let r: Result[int, Error] = Result.ok(5)
let d = Reflect.valueOf(type(Result[int, Error]), r)
d.field("error") // panic: Reflect.field: Ok[int] has no field error
```

When the name comes from input, test it against the shape's fields first:

```ant
type Point = { x: int, y: int }

func read(d: Reflect.Value, name: string): Reflect.Value | null {
  let s = d.type().shape()
  if !(s is Reflect.Record) || !s.fields.some({ x.name == name }) {
    return null
  }
  Reflect.field(d, name)
}

let d = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
println(read(d, "y"))
println(read(d, "z"))
```

```text
Reflect.Value(2: int)
null
```

## See also

[Reflect.valueOf](valueOf.md), [Reflect.shape](shape.md), [Reflect.to](to.md),
[Reflect.scalar](scalar.md), [Reflect.items](items.md), [Reflect.entries](entries.md),
[Reflect.payload](payload.md), [Reflect.makeRecord](makeRecord.md)
