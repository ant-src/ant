<!-- status: reviewed -->

# `type`

`type` declares a type name. The body decides what the name is: a record type, a named primitive
type, or an alias for any other type.

## Details

- A record body, or a numeric, `bool`, or `string` body, creates a new type, with the name as its
  tag. Two declarations with the same body are two different types.
- Any other body, such as a union or a function type, makes the name an alias, interchangeable with
  the body it names. An alias has no run-time identity, so a record literal, a constructor call, and
  `is` do not accept it.
- A `type` may appear at the top level or in any block. A type is visible throughout its block, so
  types may refer to each other in any order.
- A type may refer to itself through a record field, an array, a map, a set, or a function type.
- A type name is not a value. It appears as a record literal's head, after `is`, before `.` to name
  a function declared on it, and, for a named primitive type, as its constructor.

## Basic examples

A record type lists named fields, and a literal names the type:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 3, y: 4 }
p.x + p.y // 7
```

A named primitive type wraps a primitive. Call the type name to tag a value, and the base type name
to get it back:

```ant
type UserId = string
let id = UserId("u-1")
id         // UserId("u-1")
string(id) // "u-1"
```

A union body makes an alias, which stands for the union written out:

```ant
type Id = int | string
let ids: Array[Id] = [7, "x-9"]
ids.map({ x is int }) // [true, false]
```

## Scope

### Field defaults

A record field may give a default after `=`. A literal may then leave the field out:

```ant
type Server = { addr: string, workers: int = 4 }
Server{ addr: ":8080" }             // Server{ addr: ":8080", workers: 4 }
Server{ addr: ":9090", workers: 8 } // Server{ addr: ":9090", workers: 8 }
```

### Tagged unions

A union of named record types is a tagged union. `is` tests the tag and narrows the value, and a
chain that tests every member needs no final `else`:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle {
    3.0 * s.radius * s.radius
  } else if s is Rect {
    s.width * s.height
  }
}

println(area(Rect{ width: 2.0, height: 3.0 }))
```

```text
6.0
```

A union of empty records is an enumeration:

```ant
type Red = {}
type Amber = {}
type Green = {}
type Light = Red | Amber | Green

func next(l: Light): Light {
  if l is Red { Green{} } else if l is Green { Amber{} } else { Red{} }
}
next(Green{}) // Amber{}
```

A field that every member declares can be read through the union without narrowing:

```ant
type Song = { title: string, seconds: int }
type Podcast = { title: string, episode: int }
type Track = Song | Podcast
let queue: Array[Track] = [Song{ title: "Intro", seconds: 90 }, Podcast{ title: "News", episode: 4 }]
queue.map({ x.title }) // ["Intro", "News"]
```

### Named primitive types

A named primitive type keeps arithmetic and comparison of its base, between two values of the same
type:

```ant
type Cents = int
let total = Cents(250) + Cents(199)
total              // Cents(449)
total > Cents(400) // true
```

The type name is a function value, so it can tag every element of an array:

```ant
type Tag = string
["new", "sale"].map(Tag) // [Tag("new"), Tag("sale")]
```

### Functions on a type

A type name is also the namespace for the functions declared on it with [`func Type.name`](func.md).
A dot call on a value reaches them:

```ant
type Point = { x: int, y: int }
func Point.sum(p: Point): int { p.x + p.y }
Point{ x: 1, y: 2 }.sum() // 3
```

### Generic types

A type takes type parameters in brackets. A literal infers them from its fields:

```ant
type Pair[A, B] = { first: A, second: B }
let p = Pair{ first: 1, second: "one" }
p.second // "one"
```

### Function types

An alias can name a function type:

```ant
type Predicate[T] = func(T): bool
let even: Predicate[int] = { x % 2 == 0 }
[1, 2, 3, 4].filter(even) // [2, 4]
```

### Recursive types

A record type can refer to itself through a field. A generic one passes its own parameters, in
order:

```ant
type Tree = { value: int, children: Array[Tree] }
let t = Tree{ value: 1, children: [Tree{ value: 2, children: [] }] }
t.children[0].value // 2

type List[T] = { value: T, next: List[T] | null }
let l = List{ value: "a", next: List{ value: "b", next: null } }
l.next?.value // "b"
```

### Local types

A `type` in a block keeps a shape next to its one use. Its values may leave the block:

```ant
func tally(words: Array[string]): Array[string] {
  type Count = { word: string, n: int }
  let counts: Map[string, Count] = Map{}
  for w in words {
    let c = counts[w] ?? Count{ word: w, n: 0 }
    c.n += 1
    counts[w] = c
  }
  counts.values() | Array.sortBy({ -x.n }) | Array.map({ "${x.word}: ${x.n}" })
}

println(tally(["a", "b", "a"]))
```

```text
["a: 2", "b: 1"]
```

### Exported types

A library exports a type with `export type`. An importer uses the name as its own:

```ant [money.ant]
export type Cents = int

export func dollars(n: int): Cents {
  Cents(n * 100)
}
```

```ant
import "./money.ant" { Cents, dollars }

let price: Cents = dollars(12) + Cents(50)
println(int(price))
```

```text
1250
```

## Applications

Keep two kinds of identifier apart, so that passing one where the other belongs fails to compile:

```ant
type UserId = string
type OrderId = string
type Order = { id: OrderId, buyer: UserId }

func ordersBy(orders: Array[Order], user: UserId): Array[OrderId] {
  orders.filter({ x.buyer == user }).map({ x.id })
}

let orders = [
  Order{ id: OrderId("o-1"), buyer: UserId("ann") },
  Order{ id: OrderId("o-2"), buyer: UserId("bob") },
]
println(ordersBy(orders, UserId("ann")))
```

```text
[OrderId("o-1")]
```

Model the outcomes of an operation as a tagged union, so each caller handles every one:

```ant
type Found = { line: string }
type Missing = {}
type Lookup = Found | Missing

func search(lines: Array[string], word: string): Lookup {
  let line = lines.find({ x.contains(word) })
  if line == null {
    return Missing{}
  }
  Found{ line }
}

let result = search(["alpha", "beta", "gamma"], "bet")
println(if result is Found { "found ${result.line}" } else { "not found" })
```

```text
found beta
```

## Properties and relations

Two records are equal when they carry the same tag and their fields are equal:

```ant
type Point = { x: int, y: int }
assert(Point{ x: 1, y: 2 } == Point{ x: 1, y: 2 })
```

A spread copies the fields into a literal of another record type with the same fields, which
converts between them:

```ant
type Point = { x: int, y: int }
type Vec = { x: int, y: int }
let v = Vec{ ...Point{ x: 1, y: 2 } }
assert(v == Vec{ x: 1, y: 2 })
```

An alias is interchangeable with its body:

```ant
type Id = int | string
let a: Id = 5
let b: int | string = a
assert(a == b)
```

## Possible issues

Two record types with the same fields are still different types:

```ant
type Point = { x: int, y: int }
type Size = { x: int, y: int }
let s: Size = Point{ x: 1, y: 2 } // error: cannot use Point where Size is expected
```

A declared record type accepts only its own records, even when another record has the same fields:

```ant
type Named = { name: string }
type Person = { name: string, age: int }
let user = Person{ name: "ada", age: 36 }
let n: Named = user // error: cannot use Person where Named is expected
```

An anonymous record type in an annotation accepts any record with at least its fields:

```ant
type Person = { name: string, age: int }
func greet(n: { name: string }): string { "hi ${n.name}" }
greet(Person{ name: "ada", age: 36 }) // "hi ada"
```

A named primitive value is not its base type. Unwrap it with the base type name:

```ant
type UserId = string
let id = UserId("u-1")
let raw: string = id // error: cannot use UserId where string is expected
```

```ant
type UserId = string
let id = UserId("u-1")
string(id).length() // 3
```

Arithmetic needs both operands of one named primitive type:

```ant
type Meters = float64
let d = Meters(3.0) * 2.0 // error: * needs both operands of one type, got Meters and float64
```

An alias of a record or named primitive type is a compile error, since it would make a second tag
for one type. Declare the fields again:

```ant
type Point = { x: int, y: int }
type Coord = Point
// error: type Coord = Point would mint a second tag for one type; declare the fields or the base type again
```

An alias has no run-time identity, so it cannot be called or tested with `is`:

```ant
type Id = int | string
let a = Id(3) // error: Id is a type, not a function
let b: Id = 3
let c = b is Id // error: Id is an alias and has no run-time identity; test its members instead
```

A literal must give every field without a default, and no field the type does not declare:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1 }             // error: Point literal is missing the field y
let q = Point{ x: 1, y: 2, z: 3 } // error: Point has no field z
```

A field cannot be assigned through a union. Narrow the value first:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64 }
type Shape = Circle | Rect
let s: Shape = Rect{ width: 1.0 }
s.width = 2.0 // error: cannot assign a field through the union Shape; narrow first
```

```ant
type Circle = { radius: float64 }
type Rect = { width: float64 }
type Shape = Circle | Rect
let s: Shape = Rect{ width: 1.0 }
if s is Rect {
  s.width = 2.0
}
println(s)
```

```text
Rect{ width: 2.0 }
```

A type is not a value. Use `type(T)` for a type as a value:

```ant
type Point = { x: int }
let t = Point
// error: Point is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

```ant
type Point = { x: int }
type(Point) // Point
```

A type that refers to itself must pass through a record, array, map, set, or function type:

```ant
type Loop = Loop | null
// error: type Loop refers to itself without passing through a record, array, map, set, or function
```

A local type cannot mention a type parameter of the enclosing function:

```ant
func wrap[T](v: T): Array[T] {
  type Box = { value: T }
  // error: Box is declared in a block and cannot mention T, a type parameter of the enclosing function
  [v]
}
```

## See also

[`interface`](interface.md) · [`func`](func.md) · [`export`](export.md) · [`is`](../operators/is.md)
· [Record literals](../literals/records.md) · [Records tour](../../tour/10-records.md) ·
[Unions tour](../../tour/11-unions.md) · [Records guide](../../guide/records.md) ·
[Unions guide](../../guide/unions.md)
