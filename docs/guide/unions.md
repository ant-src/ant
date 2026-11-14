<!-- status: reviewed -->

# Unions

A union is a type whose values belong to one of several types: an `int` or a `string`, a `Circle` or
a `Rect`, a `User` or `null`. You write it with `|`, and you find out which member a value holds
with `is`. The type checker follows each test, so in the branch it guards you use the value as the
member it tested.

This chapter covers unions of primitives, tagged unions of records, enumerations, the exhaustiveness
check on chains of `is` tests, and which types a union may hold.

## Unions and is

```ant
type Id = int | string

func describe(id: Id): string {
  if id is int { "number ${id + 1}" } else { "text ${id.toUpper()}" }
}

println(describe(41))
println(describe("a7"))
```

```text
number 42
text A7
```

`Id` holds an `int` or a `string`. `id is int` tests which one, at run time. In the `then` branch
the type checker knows `id` is an `int`, so `id + 1` checks; in the `else` branch it knows `id` is
the rest of the union, a `string`. This is narrowing, and it works the same way for `null` as for
any other member: `x != null` is the test for the nullable union `T | null`.

`is` is an ordinary `bool` expression. It binds like a comparison, combines with `&&`, `||`, and
`!`, and works anywhere a `bool` does:

```ant
let values: Array[int | string | null] = [1, "two", null, 4]
values.countWhere({ x is int })                // 2
values.filter({ !(x is string) })              // [1, null, 4]
values.some({ x is string && x.length() > 5 }) // false
```

On the right of `&&`, `x` is already a `string`, so `x.length()` checks.

A test narrows in the rest of the block too. When the branch that holds one member leaves with
`return`, `break`, `continue`, or `panic`, the code below sees the remaining members. Use that to
handle a special member first and keep the main path unindented:

```ant
type Id = int | string

func normalize(id: Id): string {
  if id is int { return "#${id}" }
  id.trim().toLower()
}

println(normalize(7))
println(normalize("  AB-3 "))
```

```text
#7
ab-3
```

Only a local variable, a parameter, or a top-level `let` narrows. A field such as `order.status`
does not, so bind it to a `let` before you test it; the [Null](null.md) chapter shows the same rule
for `null`.

## Tagged unions

A union of named record types is a tagged union. Every record value carries the name of its type as
a tag, and `is` tests the tag. There is no `kind` field to declare and keep in sync:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle { 3.0 * s.radius * s.radius } else { s.width * s.height }
}

let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
println(shapes.map(area))
println(shapes[1])
```

```text
[3.0, 6.0]
Rect{ width: 2.0, height: 3.0 }
```

A value prints with its tag, and `==` compares the tag and then the fields, so a `Circle` never
equals a `Rect`.

A field that every member declares reads through the union without a test. When the members declare
it with different types, the field has the union of those types, as long as that union is valid (see
[What a union may hold](#what-a-union-may-hold)); otherwise narrow first. After a test removes the
members that lack a field, the rest can read it:

```ant
type Circle = { name: string, radius: float64 }
type Rect = { name: string, width: float64, height: float64 }
type Triangle = { name: string, base: float64, height: float64 }
type Shape = Circle | Rect | Triangle

func height(s: Shape): float64 {
  if s is Circle { return 2.0 * s.radius }
  s.height
}

let shapes: Array[Shape] = [
  Circle{ name: "coin", radius: 1.0 },
  Triangle{ name: "roof", base: 4.0, height: 1.5 },
]
println(shapes.map({ "${x.name}: ${height(x)}" }) | String.join(", "))
```

```text
coin: 2.0, roof: 1.5
```

`s.name` works on any `Shape`. `s.height` works only after the guard, where `s` is a
`Rect | Triangle` and both members declare `height`.

You cannot assign a field through a union, even one every member declares. Narrow first, then
assign:

```ant
type Draft = { title: string }
type Published = { title: string, url: string }
type Post = Draft | Published

func rename(p: Post, title: string) {
  p.title = title // error: cannot assign a field through the union Post; narrow first
}
```

```ant
type Draft = { title: string }
type Published = { title: string, url: string }
type Post = Draft | Published

func rename(p: Post, title: string) {
  if p is Draft { p.title = title } else { p.title = title }
}

let post: Post = Draft{ title: "notes" }
rename(post, "Notes on unions")
println(post)
```

```text
Draft{ title: "Notes on unions" }
```

`Result` is a tagged union from the library: `Ok[T] | Err[E]`. `is Ok` and `is Err` narrow it like
any other union. Use `try` to pass an `Err` up, and the functions of `Result` to transform one
without a test:

```ant
func parse(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("not a number: ${text}")
  Result.ok(n)
}

for text in ["41", "x"] {
  let r = parse(text)
  if r is Ok { println(r.value + 1) } else { println(r.error.message) }
}
```

```text
42
not a number: x
```

### Named primitive types

A named primitive type carries a tag too, so it can share a union with its own base type:

```ant
type UserId = string

func show(id: UserId | string): string {
  if id is UserId { "user ${string(id)}" } else { "raw ${id}" }
}

println(show(UserId("u-1")))
println(show("u-1"))
```

```text
user u-1
raw u-1
```

## Enumerations

A union of empty records is an enumeration. Each member is a type with no fields, and `Red{}` is its
one value:

```ant
type Red = {}
type Amber = {}
type Green = {}
type Light = Red | Amber | Green

func next(light: Light): Light {
  if light is Red {
    Green{}
  } else if light is Green {
    Amber{}
  } else if light is Amber {
    Red{}
  }
}

let light: Light = Red{}
println(next(light))
println(light == Red{})

let actions: Map[Light, string] = Map{ Red{}: "stop", Green{}: "go" }
println(actions[next(Amber{})])
```

```text
Green{}
true
stop
```

Members compare with `==` and work as map keys. Each member is also a type of its own, so a function
can take only `Red` and the type checker rejects the other lights.

Use an enumeration for a fixed set of choices that carry no data. When one choice needs data, give
that member fields. The union stays a tagged union, and the chains that take it apart do not change:

```ant
type Pending = {}
type Shipped = { tracking: string }
type Cancelled = { reason: string }
type Status = Pending | Shipped | Cancelled

func describe(status: Status): string {
  if status is Pending {
    "waiting"
  } else if status is Shipped {
    "shipped, tracking ${status.tracking}"
  } else if status is Cancelled {
    "cancelled: ${status.reason}"
  }
}

let statuses: Array[Status] = [Pending{}, Shipped{ tracking: "Z-9" }, Cancelled{ reason: "stock" }]
for s in statuses {
  println(describe(s))
}
```

```text
waiting
shipped, tracking Z-9
cancelled: stock
```

## Exhaustive chains

The chains in `next` and `describe` have no final `else`, and still return a value. After the last
test, no member of the union is left, so the type checker knows the chain covers every case. This is
the exhaustiveness check, and it is the reason to leave the `else` off.

Add a member to the union, and every chain whose value is used and has no final `else` fails to
check until it handles the new one:

```ant
type Red = {}
type Amber = {}
type Green = {}
type Blinking = {}
type Light = Red | Amber | Green | Blinking

func next(light: Light): Light {
  if light is Red { Green{} } else if light is Green { Amber{} } else if light is Amber { Red{} }
  // error: an if without else has type void when the condition is false, which cannot be used where Light is expected
}
```

A chain used as a statement, such as one that only calls `println`, has type `void` and is not
checked: the new member falls through and does nothing. Return the chain's value, or give it an
expected type, when you want the check.

A final `else` hides the new member instead: `Blinking` would fall into it without a word. Write
`else` when the remaining members really share one answer, and test each member when every member
needs its own.

After a complete chain nothing is left, so an `else` there can never run, and the type checker
rejects it:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle {
    3.0 * s.radius * s.radius
  } else if s is Rect {
    s.width * s.height
  } else {
    // error: this else cannot run: the tests before it cover every case
    0.0
  }
}
```

There is no `switch`, `match`, or pattern matching. A chain of `if ... is` is how you take a union
apart, and the check applies to it.

## What a union may hold

`is` decides at run time, and the runtime sees only a value's kind and its tag, never its type
arguments. So the members of a union must differ in what `is` can see: each primitive type, `null`,
`void`, `Bytes`, and `Html` once; `Array`, `Map`, `Set`, and `func` once each, whatever their type
arguments; and each named record or primitive type once.

Two arrays are one discriminator, and so are two instantiations of one generic record. Put the union
inside instead:

```ant
type Box[T] = { value: T }
type Ids = Array[int] | Array[string]
// error: a union may contain each discriminator at most once: Array[int] and Array[string] are both an array
type Boxes = Box[int] | Box[string]
// error: a union may contain each discriminator at most once: Box[int] and Box[string] are both tag Box; write Box[int | string]
```

```ant
let ids: Array[int | string] = [1, "a2"]
ids.map({ if x is int { x * 10 } else { x.length() } }) // [10, 2]
```

The rule holds for every union the type checker forms, not only those you write. An `if` whose
branches give `Array[int]` and `Array[string]` is the same error, at the `if`. Give it an expected
type, or narrow first:

```ant
let flag = true
let xs = if flag { [1] } else { ["a"] }
// error: the branches of this if do not form a valid type: a union may contain each discriminator at most once: Array[int] and Array[string] are both an array; narrow first or give the whole an expected type
let ys: Array[int | string] = if flag { [1] } else { ["a"] }
```

A few types have no discriminator, so a union holds them only in a narrow way:

- **An anonymous record type** has no tag. A union holds at most one, and never beside a named
  record type: `{ name: string } | null` is fine.
- **A type parameter** shares a union only with `null`, `void`, other type parameters, and named
  types without type arguments: `T | null` and `T | NotFound` are fine, `T | Array[int]` is not.
- **An interface** shares a union only with `null` and `void`. The [Interfaces](interfaces.md)
  chapter shows why and what to write instead.
- **Sync objects, resources, and `Random`**, such as `Mutex` or `File.Handle`: at most one of them,
  narrowed by testing the other members, as `m != null`.

```ant
type Box[T] = { value: T }
type Mixed = Box[int] | { value: string }
// error: a union may not contain an anonymous record type { value: string } alongside the named record type Box[int]: is cannot tell them apart
func pick[T](x: T | Array[int]): int { 1 }
// error: a type parameter may share a union only with null, void, other type parameters, and named types without type arguments, not with Array[int]
```

`is` takes the name a member is tested by, without type arguments: `x is Array`, `s is Box`. An
alias such as `Id` has no run-time identity, so `x is Id` is a compile error that says to test its
members.

## Unions, interfaces, or nullable fields

- **A fixed set of variants that you own**, such as the states of an order or the tokens of a
  parser: a tagged union. Code that handles every variant is checked again when you add one.
- **A set that other code extends**, such as anything a library sorts or prints: an
  [interface](interfaces.md). A new type joins by declaring the functions.
- **One shape whose parts are sometimes missing**, such as a user with an optional phone number: a
  record with a nullable field. Split it into members only when the members need different fields.
- **A value or a failure**: a `Result`, or `T | null` when the failure needs no explanation.

## Habits from other languages

- **No `kind` field.** The tag is already there. A `kind: string` beside it can drift, and `is` on
  the tag is what narrows.
- **A test must be able to go either way.** `is` on a value that can only be that member, or can
  never be it, is a compile error, not a constant `true` or `false`.
- **`filter` keeps the element type.** `shapes.filter({ x is Circle })` is still an `Array[Shape]`.
  Use `flatMap` with a test when you need an `Array[Circle]`, and annotate the result, since a bare
  `[]` fixes no element type:
  `let circles: Array[Circle] = shapes.flatMap({ if x is Circle { [x] } else { [] } })`.
- **A field path does not narrow.** Bind the field to a `let`, test the `let`.
- **JSON drops the tag.** `Json.stringify` writes a record without its type name, so every member of
  an enumeration writes as `{}`. Map each member to a string with an `if ... is` chain, such as
  `if s is Pending { "pending" } else { "shipped" }`, and write that string instead.

```ant
type Pending = {}
type Shipped = { tracking: string }
type Status = Pending | Shipped
type Order = { id: int, status: Status }

let status = Pending{}
let same = status is Pending // error: the test is always true: Pending is already Pending

func tracking(order: Order): string {
  if order.status is Shipped {
    return order.status.tracking
    // error: Status has no field tracking on every member: narrow the value first
  }
  "none"
}
```

```ant
type Pending = {}
type Shipped = { tracking: string }
type Status = Pending | Shipped
type Order = { id: int, status: Status }

func tracking(order: Order): string {
  let status = order.status
  if status is Shipped { status.tracking } else { "none" }
}

let orders = [Order{ id: 1, status: Pending{} }, Order{ id: 2, status: Shipped{ tracking: "Z-9" } }]
println(orders.map(tracking))

let shipped: Array[Shipped] = orders
  .map({ x.status })
  .flatMap({ if x is Shipped { [x] } else { [] } })
println(shipped)
```

```text
["none", "Z-9"]
[Shipped{ tracking: "Z-9" }]
```

## See also

- [Tour: Unions](../tour/11-unions.md)
- [`is`](../language/operators/is.md), [`if`](../language/keywords/if.md), and
  [`type`](../language/keywords/type.md)
- [Null](null.md), for `T | null`
- [Interfaces](interfaces.md), for open sets
- [Records](records.md)
- [Errors](errors.md), for `Result`
