<!-- status: reviewed -->

# `if`

`if` runs one of two branches by a `bool` condition. It is an expression: its value is the value of
the branch that runs.

## Details

- The condition must be a `bool`. Ant has no truthy values: a number, a string, or a nullable value
  is not a condition.
- The value of an `if` is the last expression of the branch that runs. When the branches have
  different types and nothing expects one type, the `if` has the union of them.
- An `if` without `else` has an implicit `else` of type `void`, so its type includes `void` and it
  does not fit where a `string` or another non-void type is expected. In HTML, `void` renders
  nothing, so an `if` without `else` shows its branch only when the condition holds. A complete
  chain of `is` tests needs no `else`, and an `else` after one is a compile error.
- A test with `is`, `== null`, or `!= null` on a local variable, a parameter, or a top-level `let`
  narrows it in the branch, and the `else` branch gets the rest. When an `if` has no `else` and its
  branch ends in `return`, `break`, `continue`, or an expression of type `never` such as
  `panic(...)`, the rest of the block gets the rest too.
- A field path such as `user.address` never narrows. Bind it to a local first.

## Basic examples

Choose between two values:

```ant
let n = 7
if n % 2 == 0 { "even" } else { "odd" } // "odd"
```

Run a statement only when the condition holds:

```ant
let items = ["nut", "bolt"]
if items.length() > 1 {
  println("${items.length()} items")
}
```

```text
2 items
```

## Scope

### Chains with `else if`

```ant
let n = 0
if n < 0 { "negative" } else if n == 0 { "zero" } else { "positive" } // "zero"
```

### As a value

Bind the result with `let`, pass it as an argument, or return it as the last expression of a
function:

```ant
func plural(count: int, noun: string): string {
  if count == 1 { "1 ${noun}" } else { "${count} ${noun}s" }
}

let total = 3
let label = if total == 0 { "none" } else { plural(total, "order") }
println(label)
println(plural(1, "order"))
```

```text
3 orders
1 order
```

### Branches of different types

Without an expected type, the `if` has the union of its branch types:

```ant
let input = "42"
let parsed = if input.isEmpty() { "missing" } else { input.toInt() ?? 0 }
parsed is int // true
```

An expected type checks each branch against it:

```ant
let hasDiscount = false
let note: string | null = if hasDiscount { "10% off" } else { null }
note // null
```

### Conditions with `&&`, `||`, and `!`

```ant
let age = 20
let member = false
if age >= 18 && !member { "guest" } else { "other" } // "guest"
if age < 13 || member { "free" } else { "paid" }     // "paid"
```

### Narrowing `null`

A test against `null` narrows a nullable local to its non-null type in the branch:

```ant
type User = { name: string }

func greet(user: User | null): string {
  if user != null { "hello, ${user.name}" } else { "hello, stranger" }
}

println(greet(User{ name: "Ann" }))
println(greet(null))
```

```text
hello, Ann
hello, stranger
```

The narrowing carries through `&&` and `||`, so the right side sees the narrowed type:

```ant
type User = { name: string }

func label(user: User | null): string {
  if user == null || user.name == "" { return "anonymous" }
  user.name
}

println(label(null))
println(label(User{ name: "" }))
println(label(User{ name: "Bob" }))
```

```text
anonymous
anonymous
Bob
```

An assignment in the branch narrows too. After the `if`, `name` is a `string`, because both paths
leave it non-null:

```ant
var name: string | null = null
if name == null { name = "guest" }
println(name.length())
```

```text
5
```

### Guards

When the branch leaves the block with `return`, `break`, `continue`, or `panic`, the narrowing holds
for the rest of the block:

```ant
func parsePort(text: string | null): int {
  if text == null { return 8080 }
  let port = text.toInt() ?? panic("not a port: ${text}")
  if port < 1 || port > 65535 { panic("port out of range: ${port}") }
  port
}

println(parsePort("3000"))
println(parsePort(null))
```

```text
3000
8080
```

In a loop, `continue` skips an element and narrows the rest of the body:

```ant
let lines = ["10", "x", "32"]
var sum = 0
for line in lines {
  let n = line.toInt()
  if n == null { continue }
  sum += n
}
println(sum)
```

```text
42
```

### Tagged unions with `is`

A chain of `is` tests narrows a union member by member. When it covers every member, it needs no
`else`:

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

println(area(Circle{ radius: 1.0 }))
println(area(Rect{ width: 2.0, height: 3.0 }))
```

```text
3.0
6.0
```

The `else` branch gets the union without the tested member:

```ant
type Done = { value: int }
type Failed = { reason: string }
type Outcome = Done | Failed

func describe(o: Outcome): string {
  if o is Done { "got ${o.value}" } else { "failed: ${o.reason}" }
}

println(describe(Done{ value: 3 }))
println(describe(Failed{ reason: "timeout" }))
```

```text
got 3
failed: timeout
```

### In HTML

In HTML, `void` renders nothing, so an `if` without `else` shows its branch only when the condition
holds:

```ant
let loggedIn = false
let page = <p>Hello{if loggedIn { " back" }}</p>
println(String.from(page))
```

```text
<p>Hello</p>
```

### Inside a lambda

```ant
[3, -1, 0].map({ if x > 0 { "+" } else if x < 0 { "-" } else { "0" } }) // ["+", "-", "0"]
```

### Record and map literals in the condition

A record or map literal in the condition goes in parentheses, since its `{` would otherwise start
the branch:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 0, y: 0 }
if p == (Point{ x: 0, y: 0 }) { "origin" } else { "elsewhere" } // "origin"
```

## Applications

Pick a word by count:

```ant
let count = 1
let noun = if count == 1 { "file" } else { "files" }
"${count} ${noun}" // "1 file"
```

Clamp a value into a range:

```ant
func clamp(n: int, low: int, high: int): int {
  if n < low { low } else if n > high { high } else { n }
}

println([-5, 50, 500].map({ clamp(x, 0, 100) }))
```

```text
[0, 50, 100]
```

Read an optional setting with a guard, default it only when it is missing, and report bad input:

```ant
let settings = Map{ "port": "3000", "workers": "many" }

func setting(key: string, fallback: int): Result[int, Error] {
  let text = settings[key]
  if text == null { return Result.ok(fallback) }
  let n = text.toInt()
  if n == null { return Result.fail("${key}: not an integer: ${text}") }
  Result.ok(n)
}

println(setting("port", 8080))
println(setting("timeout", 30))
println(setting("workers", 4))
```

```text
Ok{ value: 3000 }
Ok{ value: 30 }
Err{ error: Error{ message: "workers: not an integer: many" } }
```

## Properties and relations

`if c { a } else { b }` and `if !c { b } else { a }` give the same value:

```ant
for n in [-2, 0, 5] {
  let c = n > 0
  assert((if c { "yes" } else { "no" }) == (if !c { "no" } else { "yes" }))
}
```

`x ?? d` gives what `if x != null { x } else { d }` gives:

```ant
for text in ["7", "seven"] {
  let x = text.toInt()
  assert((x ?? 0) == (if x != null { x } else { 0 }))
}
```

## Possible issues

The condition must be a `bool`. A number is not a test:

```ant
let n = 5
if n { println(n) } // error: cannot use int where bool is expected
```

An `if` without `else` has type `void` when the condition is false, so it does not fit where a
`string` is expected:

```ant
let n = 5
let size: string = if n > 9 { "big" }
// error: an if without else has type void when the condition is false, which cannot be used where string is expected
```

A missing `else` does not mean `null`; add one:

```ant
let n = 5
let size: string | null = if n > 9 { "big" } else { null }
size // null
```

An `else` after a chain of `is` tests that covers every member cannot run:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle {
    3.0 * s.radius * s.radius
  } else if s is Rect {
    s.width * s.height
  } else { 0.0 } // error: this else cannot run: the tests before it cover every case
}
```

A field path does not narrow, since a call between the test and the use may change it:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let user = User{ name: "Ann", address: Address{ city: "Oslo" } }
if user.address != null { println(user.address.city) }
// error: Address | null has no field city: narrow the value first
```

Bind the field to a local, and test the local:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let user = User{ name: "Ann", address: Address{ city: "Oslo" } }
let address = user.address
if address != null { println(address.city) }
```

```text
Oslo
```

The `{` must stay on the line of the condition. A newline ends the header:

```ant
let n = 5
if n > 3 // error: expected '{' after the if header, got newline
{
  println("big")
}
```

A record literal in the condition without parentheses is a compile error:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 0, y: 0 }
if p == Point{ x: 0, y: 0 } { println("origin") }
// error: a record or map literal in the if header must be parenthesized
```

## See also

[is](../operators/is.md) · [?? operator](../operators/null-coalescing.md) ·
[?. operator](../operators/optional-chaining.md) · [logical operators](../operators/logical.md) ·
[while](while.md) · [return](return.md) · [void](../types/void.md) · [never](../types/never.md) ·
[Control flow](../../guide/control-flow.md) · [Null](../../guide/null.md) ·
[Unions](../../guide/unions.md)
