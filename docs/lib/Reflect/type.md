<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.type

```ant signature
func Reflect.type(d: Reflect.Value): Reflect.ErasedType
```

Reflect.type returns the type of d, never a union or an alias.

<!-- end generated -->

## Details

- The result is never a union or an alias: an alias gives the type it stands for, and a union the
  member the value belongs to. The types inside the result stay as written, so a box of `Json.Value`
  holding an array has type `Array[Json.Value]`.
- The type comes from the box, never from the value. A box made at an anonymous record type keeps
  that type even when the record inside is a named one.

## Basic examples

```ant
Reflect.type(Reflect.valueOf(type(int), 42))          // int
Reflect.type(Reflect.valueOf(type(int | null), 42))   // int
Reflect.type(Reflect.valueOf(type(int | null), null)) // null
```

Take the type apart with `Reflect.shape`, which needs no `Reflect.erase` since the result is already
erased:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
Reflect.type(d)                                  // Point
Reflect.shape(Reflect.type(d)) is Reflect.Record // true
```

## Scope

An alias gives the type it stands for:

```ant
type Ids = Array[int]
Reflect.type(Reflect.valueOf(type(Ids), [1, 2])) // Array[int]
```

A named primitive type is a type of its own, not an alias, and stays:

```ant
type UserId = int
Reflect.type(Reflect.valueOf(type(UserId), UserId(7))) // UserId
```

A union gives the member the value belongs to. A `Result` is the record `Ok` or `Err`:

```ant
type Point = { x: int, y: int }
type Target = Point | string
let r: Result[int, Error] = Result.ok(5)
Reflect.type(Reflect.valueOf(type(Target), Point{ x: 1, y: 2 }))   // Point
Reflect.type(Reflect.valueOf(type(Target), "home"))                // string
Reflect.type(Reflect.valueOf(type(Result[int, Error]), r))         // Ok[int]
Reflect.type(Reflect.valueOf(type(Array[int | string]), [1, "a"])) // Array[int | string]
```

Only the top level resolves. In the last case the array's element type stays the union, and each
element's box resolves it on its own:

```ant
let d = Reflect.valueOf(type(Array[int | string]), [1, "a"])
Reflect.items(d).map({ x.type() }) // [int, string]
```

The boxes that `Reflect.items`, `Reflect.entries`, `Reflect.field`, and `Reflect.payload` return
resolve their own type the same way. Here `Json.Value` resolves at every level:

```ant
let doc = Json.parse(`{"tags": ["a", "b"], "count": 2}`).unwrap()
let d = Reflect.valueOf(type(Json.Value), doc)
d.type()                                   // Map[string, Json.Value]
Reflect.entries(d).map({ x.value.type() }) // [Array[Json.Value], float64]
```

A box made at an anonymous record type keeps that type, whatever named record it holds:

```ant
type Point = { x: int, y: int }
Reflect.type(Reflect.valueOf(type({ x: int }), Point{ x: 1, y: 2 })) // { x: int }
```

Functions, opaque types, and numbers of any width give the type they were boxed at:

```ant
Reflect.type(Reflect.valueOf(type(func(int): int), { x + 1 })) // func(int): int
Reflect.type(Reflect.valueOf(type(Bytes), "hi".toBytes()))     // Bytes
Reflect.type(Reflect.valueOf(type(byte), byte(7)))             // byte
```

Call it through `.` on a box, or in a pipe:

```ant
let d = Reflect.valueOf(type(Map[string, int]), Map{ a: 1 })
d.type()                             // Map[string, int]
d | Reflect.type() | Reflect.shape() // Reflect.Map{ key: string, value: int }
```

## Applications

Render any value by its structure. The walker asks each box for its type and branches on the shape:

```ant
type Celsius = float64
type Reading = { sensor: string, value: Celsius, tags: Array[string] }

func render(d: Reflect.Value): string {
  let s = Reflect.shape(d.type())
  if s is Reflect.Record {
    let parts = s.fields.map({ "${x.name}=${render(d.field(x.name))}" })
    return "{${parts | String.join(" ")}}"
  }
  if s is Reflect.Array {
    return "[${d.items().map(render) | String.join(" ")}]"
  }
  if s is Reflect.Tagged {
    return render(d.payload())
  }
  String.from(d.scalar())
}

let r = Reading{ sensor: "a1", value: Celsius(21.5), tags: ["roof", "north"] }
println(render(Reflect.valueOf(type(Reading), r)))
```

```text
{sensor=a1 value=21.5 tags=[roof north]}
```

Look up a per-type handler in a table keyed by type. A union field finds the entry of the member its
value has:

```ant
type Event = { name: string, at: int | string }

let labels: Map[Reflect.ErasedType, string] = Map{}
labels[Reflect.erase(type(int))] = "epoch seconds"
labels[Reflect.erase(type(string))] = "ISO text"

for e in [Event{ name: "boot", at: 1700000000 }, Event{ name: "deploy", at: "2026-11-12" }] {
  let at = Reflect.valueOf(type(Event), e).field("at")
  println("${e.name}: ${labels.get(at.type()) ?? "unknown"}")
}
```

```text
boot: epoch seconds
deploy: ISO text
```

Unbox only when the type fits, and report it otherwise:

```ant
func asInt(d: Reflect.Value): Result[int, Error] {
  if !Reflect.assignable(d.type(), Reflect.erase(type(int))) {
    return Result.fail("want int, got ${d.type()}")
  }
  Result.ok(Reflect.to(d, type(int)))
}

println(asInt(Reflect.valueOf(type(int | string), 3)))
println(asInt(Reflect.valueOf(type(int | string), "3")))
```

```text
Ok{ value: 3 }
Err{ error: Error{ message: "want int, got string" } }
```

## Properties and relations

For a type that is neither a union nor an alias, the result is the erased type of the box:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
assert(Reflect.type(d) == Reflect.erase(type(Point)))
```

The result is always assignable to the type the box was made at, and never a union:

```ant
let t = type(int | string | null)
for d in [Reflect.valueOf(t, 1), Reflect.valueOf(t, "a"), Reflect.valueOf(t, null)] {
  assert(Reflect.assignable(d.type(), Reflect.erase(t)))
  assert(!(Reflect.shape(d.type()) is Reflect.Union))
}
```

`Reflect.to` succeeds exactly when the box's type is assignable to the target:

```ant
let d = Reflect.valueOf(type(int | null), 42)
assert(Reflect.assignable(d.type(), Reflect.erase(type(int))))
assert(Reflect.to(d, type(int)) == 42)
```

Two boxes are `==` when their types are `==` and their values are:

```ant
let a = Reflect.valueOf(type(int | null), 5)
let b = Reflect.valueOf(type(int), 5)
assert(a.type() == b.type())
assert(a == b)
```

## Possible issues

A box of a union value has the member's type, so it is not `==` to the union. Test with
`Reflect.assignable` instead:

```ant
let d = Reflect.valueOf(type(int | null), 4)
d.type() == Reflect.erase(type(int | null))                   // false
Reflect.assignable(d.type(), Reflect.erase(type(int | null))) // true
```

`Reflect.type` does not ask a plain value what it is. Passing one is a compile error; box it with a
type first:

```ant
println(Reflect.type(42)) // error: cannot use int where Reflect.Value is expected
```

The result is erased, so it cannot be compared with a `type(T)` directly. Erase the other side:

```ant
let d = Reflect.valueOf(type(int), 1)
println(d.type() == type(int))
// error: Reflect.ErasedType and Reflect.Type[int] cannot be compared with ==: neither is assignable to the other
```

The type does not narrow to what the value could also be. A `Point` boxed at `{ x: int }` reports
the anonymous type, so a test for `Point` fails:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type({ x: int }), Point{ x: 1, y: 2 })
d.type() == Reflect.erase(type(Point)) // false
```

## See also

[Reflect.valueOf](valueOf.md) · [Reflect.shape](shape.md) · [Reflect.to](to.md) ·
[Reflect.erase](erase.md) · [Reflect.assignable](assignable.md) · [Reflect.field](field.md) ·
[Reflect.items](items.md) · [Reflect.entries](entries.md) · [Reflect.payload](payload.md) ·
[Reflect.ErasedType](types/ErasedType.md) · [Reflect](index.md)
