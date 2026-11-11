<!-- status: reviewed -->

# `func`

`func` declares a function: a named one at the top level of a module, an anonymous one anywhere an
expression goes, and a function on a type as `func Type.name`.

## Details

- A named `func` is allowed only at the top level of a module. Inside a block, bind an anonymous
  function with `let`.
- The body is a block: its last expression is the result. Use `return` only to leave early.
- The checker infers the return type, except for an exported or recursive function, which states it.
- Parameters are `let` bindings. Assigning one is a compile error, but the body can change the array
  or record a parameter refers to, and the caller sees the change.
- A function is a value. A closure captures variables by reference and sees later assignments to
  them.

## Basic examples

Declare a function and call it:

```ant
func square(n: int): int { n * n }
square(7) // 49
```

The return type may be left out; the checker infers it from the body:

```ant
func greet(name: string) { "Hello, ${name}" }
greet("Ann") // "Hello, Ann"
```

Leave early with `return`; the last expression is the ordinary result:

```ant
func clamp(value: int, low: int, high: int): int {
  if value < low { return low }
  if value > high { return high }
  value
}
clamp(15, 0, 10) // 10
clamp(-3, 0, 10) // 0
clamp(4, 0, 10)  // 4
```

## Scope

### Results

A function without a result returns `void`. Call it for its effect:

```ant
func report(name: string, score: int) {
  println("${name}: ${score}")
}

report("Ann", 42)
```

```text
Ann: 42
```

Return several values as a record, and destructure them at the call site:

```ant
type Stats = { low: int, high: int }

func spread(xs: Array[int]): Stats {
  Stats{ low: xs.min() ?? 0, high: xs.max() ?? 0 }
}

let { low, high } = spread([4, 9, 2])
println("${low}..${high}")
```

```text
2..9
```

### Parameters

A parameter refers to the caller's value, so a change to an array or record through it is visible
outside:

```ant
func addDefault(tags: Array[string]) {
  if !tags.contains("new") { tags.push("new") }
}

let tags = ["sale"]
addDefault(tags)
println(tags)
```

```text
["sale", "new"]
```

A parameter of a function type takes a named function or a lambda:

```ant
func square(n: int): int { n * n }
func twice(n: int, f: func(int): int): int { f(f(n)) }
twice(3, square)    // 81
twice(3, { x + 1 }) // 5
```

Ant has no variadic parameters. Take an array instead:

```ant
func sum(xs: Array[int]): int { xs.sum() ?? 0 }
sum([1, 2, 3]) // 6
```

### Anonymous functions

`func` without a name is an expression. Bind it with `let`, or pass it where a function is expected:

```ant
let add = func(a: int, b: int): int { a + b }
add(2, 3)                                                // 5
["a", "bb"].map(func(s: string): string { s.toUpper() }) // ["A", "BB"]
```

Use an anonymous function where a lambda's single `x` is not enough, such as a nested callback that
needs both values:

```ant
let rows = [[1, 2], [3]]
rows.map(func(row: Array[int]): Array[int] { row.map({ x * row.length() }) }) // [[2, 4], [3]]
```

### Recursion

A recursive function states its return type:

```ant
func factorial(n: int): int {
  if n <= 1 { 1 } else { n * factorial(n - 1) }
}
factorial(5) // 120
```

Inside a block, a recursive function is a `let` binding whose type is stated, by a return type on
the `func` or a type on the binding:

```ant
func countdown(from: int): Array[int] {
  let step = func(n: int): Array[int] {
    if n == 0 { [0] } else { [n].concat(step(n - 1)) }
  }
  step(from)
}

println(countdown(3))
```

```text
[3, 2, 1, 0]
```

### Closures

A function returned from a function keeps the variables it captures:

```ant
func counter(): func(): int {
  var n = 0
  func(): int {
    n += 1
    n
  }
}

let next = counter()
println(next())
println(next())
```

```text
1
2
```

A closure sees later assignments to a captured variable:

```ant
var rate = 2
let scale = func(n: int): int { n * rate }
rate = 10
println(scale(3))
```

```text
30
```

### Generics

Type parameters go in square brackets after the name. A call never writes type arguments; the
checker solves them from the arguments:

```ant
type Pair[A, B] = { first: A, second: B }

func swap[A, B](p: Pair[A, B]): Pair[B, A] {
  Pair{ first: p.second, second: p.first }
}

swap(Pair{ first: 1, second: "one" }) // Pair{ first: "one", second: 1 }
```

A bound limits what a type parameter accepts and what the body may do with it. With `Ordered`, the
body may compare:

```ant
func bigger[T: Ordered](a: T, b: T): T { if a > b { a } else { b } }
bigger(3, 8)         // 8
bigger("ant", "bee") // "bee"
```

A record bound accepts any record type with those fields:

```ant
type User = { name: string, age: int }
type City = { name: string, population: int }

func names[T: { name: string }](xs: Array[T]): Array[string] { xs.map({ x.name }) }

names([User{ name: "Ann", age: 31 }])        // ["Ann"]
names([City{ name: "Oslo", population: 0 }]) // ["Oslo"]
```

### Functions on types

`func Type.name` declares a function in the namespace of a type declared in the same module. When
its first parameter takes the type, a value of the type calls it through `.`:

```ant
type Point = { x: float64, y: float64 }

func Point.origin(): Point { Point{ x: 0.0, y: 0.0 } }

func Point.distance(a: Point, b: Point): float64 {
  Math.sqrt((a.x - b.x) * (a.x - b.x) + (a.y - b.y) * (a.y - b.y))
}

let p = Point{ x: 3.0, y: 4.0 }
Point.distance(p, Point.origin())  // 5.0
p.distance(Point.origin())         // 5.0
Point.origin() | Point.distance(p) // 5.0
```

A function on a type is a value like any other:

```ant
type Point = { x: float64, y: float64 }
func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }
[Point{ x: 3.0, y: 4.0 }, Point{ x: 0.0, y: 1.0 }].map(Point.norm) // [5.0, 1.0]
```

A named primitive type takes functions too, such as a parser:

```ant
type UserId = string

func UserId.parse(s: string): UserId | null {
  if s.startsWith("u-") { UserId(s) } else { null }
}

UserId.parse("u-17") // UserId("u-17")
UserId.parse("17")   // null
```

## Named arguments

A parameter with a default value may be left out. Arguments may be named, after the positional ones,
in any order:

```ant
func connect(host: string, port: int = 443, tls: bool = true): string {
  let scheme = if tls { "https" } else { "http" }
  "${scheme}://${host}:${port}"
}

connect("example.com")                             // "https://example.com:443"
connect("example.com", 8080)                       // "https://example.com:8080"
connect("example.com", tls: false)                 // "http://example.com:443"
connect(port: 80, host: "example.com", tls: false) // "http://example.com:80"
```

Named arguments also work through `.` when the checker knows the one function called:

```ant
type Price = { amount: int }

func Price.label(p: Price, currency: string = "USD"): string { "${p.amount} ${currency}" }

Price{ amount: 12 }.label()                // "12 USD"
Price{ amount: 12 }.label(currency: "EUR") // "12 EUR"
```

## Applications

Validate input with guards and return a `Result`:

```ant
type Order = { id: string, quantity: int }

func parseOrder(id: string, quantity: string): Result[Order, Error] {
  let n = quantity.toInt()
  if n == null { return Result.fail("order ${id}: quantity must be an integer") }
  if n <= 0 { return Result.fail("order ${id}: quantity must be positive") }
  Result.ok(Order{ id, quantity: n })
}

parseOrder("a1", "3") // Ok{ value: Order{ id: "a1", quantity: 3 } }
parseOrder("a2", "0") // Err{ error: Error{ message: "order a2: quantity must be positive" } }
```

Build a function from settings, then pass it to `filter`:

```ant
func between(low: int, high: int): func(int): bool {
  { x >= low && x <= high }
}

let teen = between(13, 19)
[9, 14, 19, 21].filter(teen) // [14, 19]
```

## Properties and relations

A call through `.` and a pipe call the same function on a type as the namespace form:

```ant
type Point = { x: float64, y: float64 }
func Point.scale(p: Point, k: float64): Point { Point{ x: p.x * k, y: p.y * k } }
let p = Point{ x: 1.0, y: 2.0 }
assert(p.scale(3.0) == Point.scale(p, 3.0))
assert((p | Point.scale(3.0)) == Point.scale(p, 3.0))
```

A named argument binds the same parameter as its position:

```ant
func area(width: int, height: int): int { width * height }
assert(area(height: 2, width: 5) == area(5, 2))
```

## Possible issues

A recursive function without a return type is a compile error:

```ant
func fact(n: int) {
  if n <= 1 { 1 } else { n * fact(n - 1) } // error: fact is recursive; state its return type
}
```

An exported function states its return type:

```ant
export func greet(name: string) { "Hello, ${name}" }
// error: the exported function greet must state its return type
```

Assigning a parameter is a compile error. Bind a new name instead:

```ant
func shout(s: string): string {
  s = s.toUpper()
  // error: cannot assign to s, a parameter; declare a var to hold a value that changes
  s
}
```

A named `func` inside a block is a compile error. Bind an anonymous function with `let`:

```ant
func total(xs: Array[int]): int {
  func double(n: int): int { n * 2 }
  // error: a named function is only allowed at the top level; bind it with let or var
  xs.map(double).sum()
}
```

A call through a function value is positional:

```ant
let add = func(a: int, b: int): int { a + b }
add(a: 1, b: 2)
// error: a call through a function value is positional; named arguments apply only to a declared function
```

A function on a type whose first parameter does not take the type is reached only through the type's
name:

```ant
type Point = { x: float64, y: float64 }
func Point.origin(): Point { Point{ x: 0.0, y: 0.0 } }
let p = Point{ x: 1.0, y: 1.0 }
p.origin() // error: Point.origin does not take a Point first, so a value does not reach it
```

`return` inside a lambda leaves only the lambda, not the enclosing function. Here it gives the
lambda's result for one element, and `check` goes on past the `map`:

```ant
func check(xs: Array[int]): string {
  let marks = xs.map({
    if x < 0 { return "negative" }
    "ok"
  })
  "checked: ${marks | String.join(", ")}"
}

check([1, -2, 3]) // "checked: ok, negative, ok"
```

For an early exit, use `find` or `some`, or a `for` loop where `return` leaves the function:

```ant
func firstNegative(xs: Array[int]): int | null {
  for x in xs {
    if x < 0 { return x }
  }
  null
}

firstNegative([1, -2, 3])  // -2
[1, -2, 3].find({ x < 0 }) // -2
```

One type parameter takes one type, so mixed arguments do not solve it:

```ant
func bigger[T: Ordered](a: T, b: T): T { if a > b { a } else { b } }
bigger(1, 2.0) // error: cannot use float64 where int is expected
```

## See also

[return](return.md), [let](let.md), [var](var.md), [type](type.md), [interface](interface.md),
[export](export.md), [lambdas](../literals/lambdas.md), [function types](../types/functions.md),
[calls through `.`](../operators/dot-calls.md), [pipe](../operators/pipe.md),
[Functions](../../guide/functions.md), [Generics](../../guide/generics.md)
