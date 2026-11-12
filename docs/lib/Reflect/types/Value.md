<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Value

```ant signature
type Reflect.Value
```

Reflect.Value is a value together with a type it belongs to, the one way to hold a value whose type
is known only at run time. [`Reflect.valueOf`](../valueOf.md) boxes, [`Reflect.to`](../to.md)
unboxes, the readers take one apart along its type, and the builders make new ones from checked
parts. In a checked program the value always belongs to the type, which is never a union or an
alias. Two are == when their types and values are, and one renders as Reflect.Value(42: int).

<!-- end generated -->

## Details

- A `Reflect.Value` has no literal and no fields. [`Reflect.valueOf`](../valueOf.md), the readers,
  and the builders make one; [`Reflect.to`](../to.md) and [`Reflect.scalar`](../scalar.md) take the
  value out.
- Its type is the member the value belongs to, never a union, and never an alias: boxing `42` at
  `int | null` gives a box of type `int`.
- The box refers to its value and does not copy it. A change made to a boxed array or record through
  another reference shows through the box. Nothing changes a value through a box.
- `==` compares the types, then the values, so equal copies are `==`. [`Reflect.same`](../same.md)
  reports whether two boxes hold one record, array, map, or set, and compares any other value with
  `==`.
- `is` cannot test a `Reflect.Value`, so a union holds it beside no other type `is` cannot test, and
  `Json.stringify` panics on one.

## Basic examples

Box a value with its type, and take it back out:

```ant
let d = Reflect.valueOf(type(int), 42)
d                        // Reflect.Value(42: int)
Reflect.type(d)          // int
Reflect.to(d, type(int)) // 42
```

A box renders as its value and its type:

```ant
type Point = { x: int, y: int }
Reflect.valueOf(type(Point), Point{ x: 1, y: 2 }) // Reflect.Value(Point{ x: 1, y: 2 }: Point)
Reflect.valueOf(type(Array[string]), ["a", "b"])  // Reflect.Value(["a", "b"]: Array[string])
```

## Scope

### Scalars

Every scalar type boxes, `null` and `void` included:

```ant
Reflect.valueOf(type(string), "hi")          // Reflect.Value("hi": string)
Reflect.valueOf(type(float32), float32(1.5)) // Reflect.Value(1.5: float32)
Reflect.valueOf(type(bool), true)            // Reflect.Value(true: bool)
Reflect.valueOf(type(null), null)            // Reflect.Value(null: null)
Reflect.valueOf(type(void), void)            // Reflect.Value(void: void)
```

### Unions and aliases

A box made at a union holds the member its value belongs to, and an alias is expanded:

```ant
type Ids = Array[int]
Reflect.valueOf(type(int | null), 42)                  // Reflect.Value(42: int)
Reflect.valueOf(type(string | null), null)             // Reflect.Value(null: null)
Reflect.valueOf(type(Ids), [1, 2])                     // Reflect.Value([1, 2]: Array[int])
Reflect.type(Reflect.valueOf(type(int | string), "a")) // string
```

### Named primitive types

```ant
type UserId = int
Reflect.valueOf(type(UserId), UserId(7)) // Reflect.Value(UserId(7): UserId)
```

### Records through an anonymous record type

A record boxed at an anonymous record type keeps its tag, and the box's type is the anonymous one:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type({ x: int }), Point{ x: 1, y: 2 })
d                     // Reflect.Value(Point{ x: 1, y: 2 }: { x: int })
Reflect.field(d, "x") // Reflect.Value(1: int)
```

### Containers, functions, and opaque types

```ant
Reflect.valueOf(type(Map[string, int]), Map{ "a": 1 }) // Reflect.Value(Map{ "a": 1 }: Map[string, int])
Reflect.valueOf(type(func(int): int), { x * 2 })       // Reflect.Value(func: func(int): int)
Reflect.valueOf(type(Bytes), "hi".toBytes())           // Reflect.Value(Bytes(68 69): Bytes)
```

### Children of a box

The readers return boxes, each at the type the parent's type gives the child:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Array[Point]), [Point{ x: 1, y: 2 }])
d.items()               // [Reflect.Value(Point{ x: 1, y: 2 }: Point)]
d.items()[0].field("y") // Reflect.Value(2: int)
```

## Applications

### A list of values of different types

An array of boxes holds values of any types, and each box still knows its own:

```ant
let row = [
  Reflect.valueOf(type(string), "Ann"),
  Reflect.valueOf(type(int), 31),
  Reflect.valueOf(type(Array[string]), ["admin"]),
]
for d in row {
  println("${d.type()}: ${d.to(type(string | int | Array[string]))}")
}
```

```text
string: Ann
int: 31
Array[string]: ["admin"]
```

### The fields that changed

Compare two records of any type field by field. Boxes are `==` when their types and values are:

```ant
type User = { name: string, age: int, tags: Array[string] }

func changed[T](t: Reflect.Type[T], before: T, after: T): Array[string] {
  let s = t.erase().shape()
  if !(s is Reflect.Record) {
    return []
  }
  let a = Reflect.valueOf(t, before)
  let b = Reflect.valueOf(t, after)
  s.fields | Array.map({ x.name }) | Array.filter({ a.field(x) != b.field(x) })
}

let old = User{ name: "Ann", age: 31, tags: ["admin"] }
let now = User{ ...old, age: 32, tags: ["admin"] }
println(changed(type(User), old, now))
```

```text
["age"]
```

## Properties and relations

`Reflect.to` undoes `Reflect.valueOf` at the same type, and the box's type is that type when it is
neither a union nor an alias:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let d = Reflect.valueOf(type(Point), p)
assert(Reflect.to(d, type(Point)) == p)
assert(Reflect.type(d) == Reflect.erase(type(Point)))
```

Two boxes are `==` when their types are `==` and their values are `==`:

```ant
let a = Reflect.valueOf(type(int), 42)
assert(a == Reflect.valueOf(type(int), 42))
assert(a != Reflect.valueOf(type(int8), int8(42)))
assert(a != Reflect.valueOf(type(int), 43))
```

Equal copies are `==` but not the same:

```ant
let a = Reflect.valueOf(type(Array[int]), [1, 2])
let b = Reflect.valueOf(type(Array[int]), [1, 2])
assert(a == b)
assert(!Reflect.same(a, b))
```

## Possible issues

A box is not its value. Take the value out before you use it:

```ant
let d = Reflect.valueOf(type(int), 1)
let n: int = d // error: cannot use Reflect.Value where int is expected
```

`is` cannot test a box:

```ant
let d = Reflect.valueOf(type(int), 1)
println(d is Reflect.Value) // error: is cannot test Reflect.Value; narrow by testing the other members
```

A union holds at most one of the types `is` cannot test, so it cannot hold both a box and a type
value:

```ant
let x: Reflect.Value | Reflect.ErasedType = Reflect.erase(type(int))
// error: a union may hold at most one type with no discriminator, and Reflect.Value and Reflect.ErasedType both have none
```

`Json.stringify` has no form for a box:

```ant
Json.stringify(Reflect.valueOf(type(int), 1))
// panic: Json.stringify: a reflected value has no JSON form; take the value out with Reflect.to
```

A box refers to its value, so a change made through another reference shows through it:

```ant
let xs = [1]
let d = Reflect.valueOf(type(Array[int]), xs)
xs.push(2)
d // Reflect.Value([1, 2]: Array[int])
```

A box of `NaN` is not `==` to itself, because `NaN` is not:

```ant
let d = Reflect.valueOf(type(float64), Math.NaN)
d == d // false
```

Taking a value out at a type its box's type is not assignable to panics:

```ant
let d = Reflect.valueOf(type(int), 42)
Reflect.to(d, type(string)) // panic: Reflect.to: int is not assignable to string
```

## See also

[Reflect.valueOf](../valueOf.md) · [Reflect.to](../to.md) · [Reflect.type](../type.md) ·
[Reflect.same](../same.md) · [Reflect.scalar](../scalar.md) · [Reflect.items](../items.md) ·
[Reflect.entries](../entries.md) · [Reflect.field](../field.md) · [Reflect.payload](../payload.md) ·
[Reflect.makeRecord](../makeRecord.md) · [Reflect.Type](Type.md) ·
[Reflect.ErasedType](ErasedType.md)
