<!-- status: reviewed -->

# Records instead of classes

Ant has no classes. There is no `class` keyword, no `this`, no constructor, no inheritance, and no
function written inside a type's body. Data is a record, behavior is a function, and code that works
with many types names what it needs in an interface. This page explains why Ant takes a class apart
that way, and what the split costs.

## One tool per job

A class does several jobs at once. It declares a shape of data, groups the functions that work on
it, builds values, hides members, and lets one type stand in for another. Ant gives each job its own
tool:

| Job                                | A class                       | Ant                                      |
| ---------------------------------- | ----------------------------- | ---------------------------------------- |
| Declare the data                   | fields                        | a record type                            |
| Group the functions on it          | methods                       | functions on the type: `func Point.norm` |
| Build a value                      | a constructor                 | a record literal, field defaults         |
| Check a value as it is built       | a constructor                 | a function on the type: `Range.new`      |
| Hide what callers must not use     | `private` members             | a module: what it does not `export`      |
| Let many types share one operation | inheritance, an abstract base | an interface, or a union for a fixed set |
| Keep state behind behavior         | an object with private fields | a record of closures                     |

Each tool does one job, and a type uses only the tools it needs.

## Data is a record

```ant
type Point = { x: float64, y: float64 }

let p = Point{ x: 3.0, y: 4.0 }
println(p)
println(p == Point{ x: 3.0, y: 4.0 })
println(Json.stringify(p))
```

```text
Point{ x: 3.0, y: 4.0 }
true
{"x":3.0,"y":4.0}
```

A record is its tag and its fields. Printing it, comparing it with `==`, and encoding it as JSON all
see the same fields, with no hook to override and no hidden state to leave out. A reader who knows
the fields knows the value.

Assigning or passing a record shares it. Write `Point{ ...p, x: 0.0 }` for a copy with one field
changed.

## Functions live on the type

```ant
type Point = { x: float64, y: float64 }

func Point.origin(): Point { Point{ x: 0.0, y: 0.0 } }

func Point.distance(a: Point, b: Point): float64 {
  Math.sqrt((a.x - b.x) * (a.x - b.x) + (a.y - b.y) * (a.y - b.y))
}

func Point.norm(p: Point): float64 { p.distance(Point.origin()) }

let p = Point{ x: 3.0, y: 4.0 }
println(p.norm())
println(Point.distance(p, Point.origin()))
println([p, Point{ x: 0.0, y: 2.0 }].map(Point.norm))
```

```text
5.0
5.0
[5.0, 2.0]
```

`func Point.distance` declares `distance` in the namespace of `Point`. It is an ordinary function:
the receiver is its first parameter, named and typed like the others, so there is no `this` to bind
or to lose. `p.norm()` is a shorter way to write `Point.norm(p)`, and `Point.norm` is a function
value that `map` takes as it is. A function that builds a `Point` from nothing, such as
`Point.origin`, lives in the same namespace and is called through the type's name.

A type's namespace belongs to the module that declares the type, and only that module adds functions
to it. Code elsewhere that wants a new operation on a `Point`, or on a `string`, writes a plain
function and calls it through a pipe:

```ant
func shout(text: string): string { "${text.toUpper()}!" }

println("hello" | shout())
```

```text
HELLO!
```

## Construction is a literal

A record literal names every field that has no default, and field defaults fill in the rest:

```ant
type Server = { addr: string, workers: int = 4 }

println(Server{ addr: ":8080" })
println(Server{ addr: ":9090", workers: 8 })
```

```text
Server{ addr: ":8080", workers: 4 }
Server{ addr: ":9090", workers: 8 }
```

A literal leaves out a field that has a default, and the default fills it in. Most types need
nothing more. When fields must agree with each other, the type gets a function that checks them
together and returns a `Result`:

```ant
type Range = { low: int, high: int }

func Range.new(low: int, high: int): Result[Range, Error] {
  if low > high {
    return Result.fail("range ${low}..${high}: low is above high")
  }
  Result.ok(Range{ low, high })
}

println(Range.new(1, 10))
println(Range.new(10, 1))
```

```text
Ok{ value: Range{ low: 1, high: 10 } }
Err{ error: Error{ message: "range 10..1: low is above high" } }
```

`Range.new` is a function like any other: it returns a `Result` the caller handles, and its name
says what it builds. It does not stop anyone from writing `Range{ low: 10, high: 1 }` directly; that
is one of the costs below.

## Sharing behavior: interfaces

```ant
interface Shape {
  area(s: Self): float64
}

type Square = { side: float64 }
type Rect = { width: float64, height: float64 }

func Square.area(s: Square): float64 { s.side * s.side }
func Rect.area(r: Rect): float64 { r.width * r.height }

func total(shapes: Array[Shape]): float64 {
  shapes.map({ x.area() }).sum() ?? 0.0
}

println(total([Square{ side: 2.0 }, Rect{ width: 2.0, height: 3.0 }]))
```

```text
10.0
```

`Square` and `Rect` never mention `Shape`. A type satisfies an interface by having the functions it
lists, and nothing declares that it does. So a type can satisfy an interface written after it, in a
module its author never saw, and so can the built-in types:

```ant
interface Sized {
  length(s: Self): int
}

func longest(items: Array[Sized]): int {
  items.map({ x.length() }).max() ?? 0
}

println(longest(["ant", [1, 2, 3, 4], "hi".toBytes()]))
```

```text
4
```

A class hierarchy fixes, when a type is written, every abstraction it takes part in. An interface
lets the code that needs an abstraction name it, at the place it needs it.

When the variants are a fixed set that one module owns, such as the states of a connection, a union
fits better than an interface: `is` narrows a value to one member, and the type checker tells you
when a chain of tests no longer covers every member. See
[Interfaces versus unions](../guide/interfaces.md#interfaces-versus-unions) for how to pick.

## Hiding is the module's job

```ant [geometry.ant]
export type Point = { x: float64, y: float64 }

export func Point.norm(p: Point): float64 { Math.sqrt(p.squared()) }

func Point.squared(p: Point): float64 { p.x * p.x + p.y * p.y }
```

```ant
import "./geometry.ant" { Point }

let p = Point{ x: 3.0, y: 4.0 }
println(p.norm())
println(p.squared()) // error: Point.squared is not exported
```

A module is the unit of privacy. Importing `Point` brings the functions exported on it, so
`p.norm()` works in any module that imports the type. `Point.squared` has no `export`, so no other
module calls it, except a test file. A class marks each member public or private; Ant draws one
line, around the file, and the functions on a type follow the same rule as every other declaration.

## Reuse by composition

A class reuses fields by inheriting them. A record reuses them by holding another record, and a
function that needs only some fields takes an anonymous record type, which accepts any record that
has them:

```ant
type Address = { city: string }
type Customer = { name: string, address: Address }
type Supplier = { name: string, address: Address, terms: int }

func label(party: { name: string, address: Address }): string {
  "${party.name}, ${party.address.city}"
}

println(label(Customer{ name: "Ann", address: Address{ city: "Oslo" } }))
println(label(Supplier{ name: "Acme", address: Address{ city: "Lyon" }, terms: 30 }))
```

```text
Ann, Oslo
Acme, Lyon
```

`Customer` and `Supplier` share no base type, and `label` does not need one. It states the fields it
reads, and any record with those fields passes.

## State behind behavior: closures

Fields hold data anyone can read. When behavior should own its state, a function returns a record of
closures, and the state lives in variables only those closures capture:

```ant
type Counter = { next: func(): int, reset: func(): void }

func counter(start: int): Counter {
  var n = start
  Counter{
    next: {
      n += 1
      n
    },
    reset: { n = start },
  }
}

let c = counter(10)
c.next()
c.next()
println(c.next())
c.reset()
println(c.next())
```

```text
13
11
```

`c.next()` reads the field `next` and calls it. No code outside `counter` can reach `n`. This form
suits a value whose behavior is chosen when it is built, such as a handler and its configuration. A
field that holds a function is data, not a function on the type, so it satisfies no interface.

## Why

- **Data stays plain.** Every record works with `==`, printing, JSON, and reflection the same way,
  because no record can change how they see it.
- **Functions stay functions.** A function on a type is a value, so it passes to `map`, `filter`, or
  `sortBy` as it is. With no `this`, no call depends on how the function was reached.
- **One function, whatever the spelling.** `p.norm()`, `Point.norm(p)`, and `p | Point.norm()` call
  the same function with the same arguments. There is no method beside a function to keep in step,
  which is the [one obvious way](one-way.md) applied to behavior.
- **Types are advisory.** `ant run` discards annotations. A record carries its type's tag, and a
  call through `.` looks up the function among those the tag holds. Interfaces cost nothing at run
  time: the interpreter keeps nothing of them, and the type checker alone checks that every call
  they promise exists. The design note on [advisory types](advisory-types.md) has more.
- **Abstractions come later.** An interface can be added when a second type appears, without editing
  the first.

## What it costs

- **No field is private.** `export` hides a type or a function from other modules, not a field from
  code that holds the record. Any code that has a `Range` can set `high` below `low`. An invariant
  that fields must keep lives in the functions that build and change the record, and in the callers'
  discipline, or behind closures.
- **No extension of other types.** Only the module that declares a type adds functions to it, so
  `func String.shout` is a compile error; write it as the plain function `shout` above.

  ```ant
  func String.shout(s: string): string { s.toUpper() }
  // error: func String.shout: String is not a type declared at the top level of this module
  ```

- **No inherited implementations.** An interface lists signatures and has no default functions. Two
  types that share behavior each declare it, and share the code through a helper function they both
  call.
- **No custom equality, printing, or order.** A record cannot change what `==` compares, how it
  prints, or how it sorts. Sort records with `Array.sortBy` or `Array.sortWith`, or with a `compare`
  of their own behind an interface bound.
- **No bound function values.** `p.norm` outside a call reads a field, so it is a compile error when
  `Point` has no field `norm`. Write `Point.norm` for the function, or `{ p.norm() }` to keep `p`:

  ```ant
  type Point = { x: float64, y: float64 }

  func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }

  let p = Point{ x: 3.0, y: 4.0 }
  let f = p.norm // error: Point has no field norm
  let g: func(): float64 = { p.norm() }
  ```

## See also

The guide chapters [Records](../guide/records.md), [Functions](../guide/functions.md),
[Interfaces](../guide/interfaces.md), [Unions](../guide/unions.md), and
[Modules](../guide/modules.md) show how to use each piece. The tour pages
[Records](../tour/10-records.md) and [Interfaces](../tour/16-interfaces.md) introduce them.
Reference: [`type`](../language/keywords/type.md), [`interface`](../language/keywords/interface.md),
[`Self`](../language/types/Self.md), and [Calls through `.`](../language/operators/dot-calls.md).
The design notes [One obvious way](one-way.md) and [Why no exceptions](errors-are-values.md) cover
related decisions.
