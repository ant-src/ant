<!-- status: reviewed -->

# Generics

A type parameter lets one function or type work for many types. The type checker solves it at every
use, from the values and the types around it. At run time a call never checks its type arguments or
their bounds.

This chapter covers type parameters and how inference solves them, bounds, why containers do not
widen, and how to pick between the kinds of `type` declaration, including recursive and local ones.

## Type parameters

Type parameters go in square brackets after the name of a function or a type:

```ant
type Page[T] = { number: int, items: Array[T] }

func paginate[T](xs: Array[T], size: int): Array[Page[T]] {
  let pages: Array[Page[T]] = []
  for chunk in xs.chunk(size) {
    pages.push(Page{ number: pages.length() + 1, items: chunk })
  }
  pages
}

let pages = paginate(["a", "b", "c", "d", "e"], 2)
println(pages.length())
println(pages[2])
println(paginate([1.5, 2.5], 10)[0].items)
```

```text
3
Page{ number: 3, items: ["e"] }
[1.5, 2.5]
```

You never write type arguments at a call. In `paginate(["a", …], 2)` the checker reads `T = string`
off the array, and in the second call `T = float64`. The literal `Page{ number: …, items: chunk }`
infers `Page`'s `T` from its fields in the same way. Inside the body, `T` is a type the caller
chooses, so `paginate` can store and pass values of it but cannot assume anything else about them.

Write type arguments only where a type is written: in an annotation such as `Array[Page[T]]`, a
parameter, or a return type.

A function on a generic type lists the type's parameters in its own brackets, and it can add more,
such as `U`:

```ant
type Box[T] = { value: T }

func Box.map[T, U](b: Box[T], f: func(T): U): Box[U] { Box{ value: f(b.value) } }

println(Box{ value: "ant" }.map({ x.length() }))
```

```text
Box{ value: 3 }
```

## How inference fills in types

Inference is local. Each expression gets its type from itself or from the type expected where it
stands: an annotation, a parameter type, a declared return type, a record field. A generic call
takes its type arguments from the expected type first, then from the arguments, and from lambdas
last, so that `x` in a lambda already has a type when the checker reaches it:

```ant
let words = ["ant", "bee", "wasp"]
words.map({ x.length() })            // [3, 3, 4]
words.map({ x.length() }).sum() ?? 0 // 10
```

In `words.map`, the array sets `T` to `string`, and then the body of the lambda sets `U` to `int`.

When nothing fixes a type, annotate. An empty literal has no elements to read a type from, so it
needs the type from its context:

```ant
let seen: Set[string] = Set.of([])
let counts: Map[string, int] = Map{}
let lines: Array[string] = []
println([seen.size(), counts.size(), lines.length()])
```

```text
[0, 0, 0]
```

A generic function can also be a value. The expected function type fixes its type parameters, as the
element type does when it is passed to `map`:

```ant
func first[T](xs: Array[T]): T | null { xs.get(0) }

let firstInt: func(Array[int]): int | null = first
println(firstInt([4, 5]))
println([["x", "y"], ["z"]].map(first))
```

```text
4
["x", "z"]
```

## Bounds

A type parameter without a bound is opaque. The body can compare it with `==`, use it as a `Map` key
or `Set` member, put it in a string, and pass it along, but nothing more. A bound, written
`[T: Bound]`, tells the body what it may do with a `T`, and limits what a call may pass:

- A union bound such as `Ordered` (the numbers and `string`) allows each operator that every member
  supports, so `[T: Ordered]` allows `<` and `>`.
- A record bound such as `{ name: string }` allows reading and assigning its fields.
- An interface bound allows calling the functions the interface lists, through `.`.

```ant
func largest[T: Ordered](xs: Array[T]): T | null {
  var best: T | null = null
  for x in xs {
    if best == null || x > best { best = x }
  }
  best
}

func titles[T: { title: string }](xs: Array[T]): Array[string] { xs.map({ x.title }) }

type Book = { title: string, pages: int }
type Film = { title: string, minutes: int }

println(largest([3, 9, 4]))
println(largest(["pear", "apple"]))
println(titles([Book{ title: "Emma", pages: 474 }]))
println(titles([Film{ title: "Alien", minutes: 117 }]))
```

```text
9
pear
["Emma"]
["Alien"]
```

A solution for a union bound is always a single type that fits one member of the union, never the
union itself. In one call of `largest` `T` is `int`, in the next `string`, and that is what makes
`x > best` sound: both sides are the same member.

Two arguments of one `T` must have the same type. The first argument below fixes `T` as `int`, and
literals never convert, so write `2` or `float64(1)`:

```ant
func bigger[T: Ordered](a: T, b: T): T { if a > b { a } else { b } }

let m = bigger(1, 2.0) // error: cannot use float64 where int is expected
```

### Bound or interface parameter

An interface that mentions `Self` only as the receiver, such as `Priced`, can also be a plain
parameter type; one that takes a second `Self` is a bound only. Use a bound when the caller needs
its own type back. `cheapest` below returns a `T`, so the caller gets a `Book` and can read its
`title`, which a function returning `Priced` could not give:

```ant
interface Priced {
  price(p: Self): int
}

type Book = { title: string, cents: int }

func Book.price(b: Book): int { b.cents }

func cheapest[T: Priced](items: Array[T]): T | null { items.minBy({ x.price() }) }

let books = [Book{ title: "Dune", cents: 999 }, Book{ title: "Emma", cents: 450 }]
println(cheapest(books)?.title)
```

```text
Emma
```

The [interfaces chapter](interfaces.md) covers declaring and satisfying interfaces.

### Narrowing leaves `T` behind

`is` tests a bounded `T` as it tests a value of its bound. The narrowed value is the tested type, a
plain `string` here, and no longer a `T`, so it cannot be returned where a `T` is expected:

```ant
func orFallback[T: Ordered](x: T, fallback: T): T {
  if x is string {
    if x == "" { fallback } else { x } // error: cannot use string where T is expected
  } else {
    x
  }
}
```

Keep a copy taken before the test. It is still a `T`:

```ant
func orFallback[T: Ordered](x: T, fallback: T): T {
  let original = x
  if x is string && x == "" { fallback } else { original }
}

println(orFallback("", "none"))
println(orFallback(7, 0))
```

```text
none
7
```

## Containers do not widen

Type arguments are invariant. A `Circle` is a `Shape`, but an `Array[Circle]` is not an
`Array[Shape]`: if it were, code holding the wider array could push a `Rect` among the circles.

<!-- test: before
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
-->

```ant
let circles = [Circle{ radius: 1.0 }]
let shapes: Array[Shape] = circles // error: cannot use Array[Circle] where Array[Shape] is expected
```

Pick one of two fixes. Copy into a new array, which the expected type widens element by element, or
make the function generic, bounded by the wider type, so that it accepts the array as it is:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle { 3.0 * s.radius * s.radius } else { s.width * s.height }
}

func totalArea[T: Shape](shapes: Array[T]): float64 { shapes.map(area).sum() ?? 0.0 }

let circles = [Circle{ radius: 1.0 }, Circle{ radius: 2.0 }]
let mixed: Array[Shape] = [...circles, Rect{ width: 2.0, height: 3.0 }]
println(totalArea(circles))
println(mixed.map(area).sum())
```

```text
15.0
21.0
```

The copy is right when the code needs a mixed array. The bound is right when the function only reads
the elements: no copy, and the caller's array stays an `Array[Circle]`.

A union bound takes one member at a time, so `totalArea` does not accept `mixed` itself. For a
function that takes both, use a parameter of type `Array[Shape]` and copy at the call.

<!-- test: before
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle { 3.0 * s.radius * s.radius } else { s.width * s.height }
}

func totalArea[T: Shape](shapes: Array[T]): float64 { shapes.map(area).sum() ?? 0.0 }

let mixed: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
-->

```ant
totalArea(mixed) // error: T would be Shape, which is not a single member of its bound Shape; narrow first
```

## Type declarations: which one to pick

`type` declares a name, and its body decides what kind of name it is:

| Body                                       | Gives                                   | Pick it for                                       |
| ------------------------------------------ | --------------------------------------- | ------------------------------------------------- |
| A record, `{ x: int }`                     | A new type, with its name as the tag    | Data with named fields; members of tagged unions  |
| A primitive, `string`, `int`, …            | A new type over that primitive          | Identifiers and units that must not mix           |
| Anything else: a union, a function type, … | An alias, interchangeable with its body | Naming a long type, such as a union or a callback |

A record or a named primitive type is a new type: two declarations with the same body give two types
that do not mix. An alias is only a shorter way to write its body, so an `Id` and an `int | string`
go wherever the other does. A generic alias takes its parameters to the body:

```ant
type Predicate[T] = func(T): bool
type Id = int | string

let short: Predicate[string] = { x.length() < 4 }
let id: Id = 7
let raw: int | string = id
["ant", "beetle", "wasp"].filter(short) // ["ant"]
raw == id                               // true
```

A named primitive type may take a type parameter that its body does not mention. `Id[User]` and
`Id[Order]` are then two types over one tag: the checker keeps them apart, and nothing changes at
run time. Only an expected type fixes the parameter, so annotate:

```ant
type User = { name: string }
type Order = { total: int }
type Id[T] = string

func lookup(id: Id[User]): string { "user ${string(id)}" }

let user: Id[User] = Id("u-1")
println(lookup(user))
```

```text
user u-1
```

An `Id[Order]` does not go where an `Id[User]` is expected:

<!-- test: before
type User = { name: string }
type Order = { total: int }
type Id[T] = string

func lookup(id: Id[User]): string { "user ${string(id)}" }
-->

```ant
let order: Id[Order] = Id("o-7")
lookup(order) // error: cannot use Id[Order] where Id[User] is expected
```

## Recursive types

A record type may refer to itself through a field, and a generic one passes its own type parameters,
unchanged and in order:

```ant
type List[T] = { value: T, next: List[T] | null }

let list = List{ value: "a", next: List{ value: "b", next: null } }
list.next?.value // "b"
```

An alias may refer to itself too, but only through an array, a map, a set, a record, or a function
type. This is how Ant writes data whose shape is known only at run time, since there is no `any`
type:

```ant
type Value = null | bool | float64 | string | Array[Value] | Map[string, Value]

func depth(v: Value): int {
  let children = if v is Array { v } else if v is Map { v.values() } else { return 0 }
  1 + (children.map(depth).max() ?? 0)
}

let doc: Value = Map{ tags: ["a", ["b"]], ok: true }
println(depth(doc))
println(depth("flat"))
```

```text
3
0
```

The prelude's `Json.Value` is such an alias. A recursive alias takes no type parameters. Write the
generic version with a record in the cycle instead, because a named record type is compared by its
tag and the recursion stops there:

```ant
type Leaf[T] = { value: T }
type Nest[T] = Array[Nest[T]] | Leaf[T]
// error: type Nest refers to itself without passing through a record's field; an alias with type parameters may not, so wrap the recursion in a named record type
```

```ant
type Leaf[T] = { value: T }
type Group[T] = { items: Array[Nest[T]] }
type Nest[T] = Leaf[T] | Group[T]

func total(n: Nest[int]): int {
  if n is Leaf { n.value } else { n.items.map(total).sum() ?? 0 }
}

let tree: Nest[int] = Group{
  items: [Leaf{ value: 1 }, Group{ items: [Leaf{ value: 2 }, Leaf{ value: 3 }] }],
}
println(total(tree))
```

```text
6
```

## Local types

A `type` may appear in any block, which keeps a shape next to its one use. It is visible throughout
its block, and its values may leave the block even though code outside cannot name the type:

```ant
func longest(words: Array[string]): string {
  type Scored = { word: string, size: int }
  let scored = words.map({ Scored{ word: x, size: x.length() } })
  scored.maxBy({ x.size })?.word ?? ""
}

println(longest(["ant", "beetle", "wasp"]))
```

```text
beetle
```

Each local type has one tag for the whole run, so it cannot mention a type parameter of the function
around it. Make it a top-level generic type instead:

```ant
func wrap[T](v: T): Array[T] {
  type Box = { value: T }
  // error: Box is declared in a block and cannot mention T, a type parameter of the enclosing function
  [v]
}
```

## Coming from other languages

There is no `any`, `Object`, or `interface{}`. A function that takes every type is generic, and data
of unknown shape is a union written out, such as `Value` above:

```ant
func show(v: any): string { "${v}" } // error: unknown type any
```

```ant
func show[T](v: T): string { "${v}" }

println(show(3))
println(show([1, 2]))
```

```text
3
[1, 2]
```

A call does not take type arguments, as `first<int>(xs)` or `first[int](xs)` would elsewhere. Give
the type by annotating the result or the argument instead:

<!-- test: before
func first[T](xs: Array[T]): T | null { xs.get(0) }
-->

```ant
let n = first[int]([1, 2])
// error: first is generic; call it, or give the expected function type that fixes its type parameters
```

A `var` keeps the type of its initializer, and a later assignment does not widen it. Annotate the
wider type where you declare the name:

```ant
type Idle = {}
type Busy = { job: string }
type State = Idle | Busy

var s = Idle{}
s = Busy{ job: "build" } // error: cannot use Busy where Idle is expected
var t: State = Idle{}
t = Busy{ job: "build" }
```

A literal takes no type from its context: `5` is always an `int` and `2.0` always a `float64`. Use a
conversion for another numeric kind:

```ant
let small: int8 = 5 // error: cannot use int where int8 is expected
let ok = int8(5)
```

## See also

[Generics tour](../tour/15-generics.md) · [`type`](../language/keywords/type.md) ·
[`func`](../language/keywords/func.md) · [`interface`](../language/keywords/interface.md) ·
[Types](types.md) · [Records](records.md) · [Unions](unions.md) · [Interfaces](interfaces.md) ·
[Runtime types](runtime-types.md)
