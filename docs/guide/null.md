<!-- status: reviewed -->

# Null

`null` marks an absent value: no user with that id, no element in an empty array, no number in the
text. In Ant, absence is part of a type. A value can be `null` only where its type says so, and the
type checker makes you handle the `null` before you use the value.

This chapter shows how nullable types work, how a test narrows them, what `?.` and `??` do, and when
to return `null` and when to return a `Result`.

## Nullable types

A nullable type is a union with `null`, written `T | null`. A `string` always holds a string; a
`string | null` holds a string or `null`.

```ant
let name: string | null = null
let title: string = null // error: cannot use null where string is expected
```

Many library functions return a nullable type when there may be no answer:

```ant
[3, 8, 5].find({ x > 4 }) // 8
[3, 8, 5].find({ x > 9 }) // null
"42".toInt()              // 42
"4x2".toInt()             // null
let ages = Map{ "Ann": 31 }
ages["Ann"] // 31
ages["Bob"] // null
```

There is no `undefined` and no `Option` wrapper. `null` is the only value of its type, and since a
nullable type is a union, `T | null | null` is `T | null`. Absence does not nest, so a map whose
values are themselves nullable gives `null` both for a missing key and for a stored `null`. Ask the
map instead of the value:

```ant
let nicknames: Map[string, string | null] = Map{ "Ann": null }
nicknames["Ann"]     // null
nicknames["Bob"]     // null
nicknames.has("Ann") // true
nicknames.has("Bob") // false
```

## Narrowing

You cannot use a `User | null` as a `User`. Test it first: after `user != null`, the type checker
knows `user` is a `User`, and the `else` branch gets the rest of the union, here `null`.

```ant
type User = { name: string }

func greet(user: User | null): string {
  if user != null { "hello, ${user.name}" } else { "hello, guest" }
}

println(greet(User{ name: "Ann" }))
println(greet(null))
```

```text
hello, Ann
hello, guest
```

A guard narrows the rest of the block. When the branch of `user == null` leaves, with `return`,
`break`, `continue`, or `panic`, the code below it sees a `User`. `assert(user != null)` narrows the
same way. Prefer the guard when absence ends the work, and keep the main path unindented:

```ant
type User = { id: int, name: string }

let users = [User{ id: 1, name: "Ann" }]

func nameOf(id: int): string {
  let user = users.find({ x.id == id })
  if user == null { return "no user ${id}" }
  user.name.toUpper()
}

println(nameOf(1))
println(nameOf(2))
```

```text
ANN
no user 2
```

Tests combine through `&&`, `||`, and `!`, so the right side of `&&` already sees the narrowed type:

```ant
type User = { name: string, admin: bool }

func canEdit(user: User | null): bool {
  user != null && user.admin
}

println(canEdit(User{ name: "Ann", admin: true }))
println(canEdit(null))
```

```text
true
false
```

An assignment narrows too. After `if c == null { c = ... }`, both branches leave a value in `c`, so
`c` is no longer nullable. For a plain default, `(given ?? "untitled").toUpper()` is the idiomatic
form; the assignment fits when filling the value takes several statements, and this example keeps it
short to show the narrowing:

```ant
func label(given: string | null): string {
  var c = given
  if c == null { c = "untitled" }
  c.toUpper()
}

println(label("notes"))
println(label(null))
```

```text
NOTES
UNTITLED
```

### What narrows

Narrowing applies to a local variable, a parameter, or a top-level `let`. A field path never
narrows, because any call between the test and the use could change the field. Bind the field to a
`let`, then test the `let`:

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }

func city(user: User): string {
  if user.address != null {
    return user.address.city // error: Address | null has no field city: narrow the value first
  }
  "unknown"
}
```

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }

func city(user: User): string {
  let address = user.address
  if address == null { return "unknown" }
  address.city
}

println(city(User{ name: "Ann", address: Address{ city: "Oslo" } }))
println(city(User{ name: "Bob", address: null }))
```

```text
Oslo
unknown
```

When you only need the field or a default, `user.address?.city ?? "unknown"` does it in one
expression; bind a `let` when several uses follow the test.

Inside a lambda, a `go` block, or a `defer` block, an outer `let` or parameter keeps its narrowing,
since nothing can change it. An outer `var` does not, even where the code around the lambda knows
its type: the lambda may run after another assignment. Copy the `var` into a `let` first:

```ant
var best: string | null = null
best = "Ann"
println(best.toUpper())
let names = ["Bob", "Cy"]
let tags = names.map({ "${x} after ${best.toUpper()}" })
// error: string | null has no field or function toUpper on every member: narrow the value first
```

```ant
var best: string | null = null
best = "Ann"
let names = ["Bob", "Cy"]
let winner = best
println(names.map({ "${x} after ${winner.toUpper()}" }) | String.join(", "))
```

```text
Bob after ANN, Cy after ANN
```

The rule also runs the other way: a local assigned inside a lambda, a `go` block, or a `defer` block
never narrows outside it, since the assignment may run at any moment. A goroutine does not hand back
a result by setting a local of its caller; send it through a channel, or write it into an array slot
or a record field and read that into a fresh `let` after the wait:

```ant
type User = { id: int, name: string }
let users = [User{ id: 1, name: "Ann" }]
var hit: User | null = null
go { hit = users.find({ x.id == 1 }) }
if hit != null { println(hit.name) } // error: User | null has no field name: narrow the value first
```

## ?? and ?.

`??` replaces `null` with a value of your choice. `?.` reads a field, indexes, or calls only when
the left side is not `null`, and gives `null` otherwise. It is spelled `a?.b` for a field, `a?.[i]`
for an index, and `f?.(x)` for a call. Together they cover the common cases without a test.

```ant
type Address = { city: string }
type User = { name: string, address: Address | null }
let ann: User | null = User{ name: "Ann", address: Address{ city: "Oslo" } }
let bob: User | null = User{ name: "Bob", address: null }
let nobody: User | null = null
ann?.address?.city                 // "Oslo"
bob?.address?.city                 // null
nobody?.address?.city ?? "unknown" // "unknown"
nobody?.name.length()              // null
let scores: Array[int] | null = [7, 9]
scores?.[0] // 7
let double: (func(int): int) | null = { x * 2 }
double?.(4) // 8
```

Write `?.` on each link that may be `null`, and `.` on the others. A `null` on the left skips the
rest of the chain, so in `nobody?.name.length()` the `.length()` applies to a plain `string` and
needs no `?.` of its own.

The right side of `??` can be `return`, `break`, or `panic(...)`. Then `??` either unwraps the value
or leaves the function, so the `let` it initializes is not nullable, and a variable on its left is
narrowed for the rest of the block:

```ant
func firstWord(text: string): string {
  let words = text.fields()
  let first = words.first() ?? return "(empty)"
  first.toUpper()
}

println(firstWord("hello world"))
println(firstWord("   "))
```

```text
HELLO
(empty)
```

Choose between them by what absence means:

- A sensible default exists: `??` with the default, such as `ages[name] ?? 0`.
- Absence ends the work: `?? return ...`, or an `if` guard when handling it takes several lines.
- Both cases need real work: `if x != null { ... } else { ... }`.
- The value only passes through: `?.`, and let the caller decide.

## Null or Result

Both carry an outcome that is not a plain value. Pick by what the caller needs to know:

- Return `T | null` when absence is an ordinary answer and needs no explanation: a lookup that finds
  nothing, an empty array's first element, text that is not a number.
- Return a `Result` when the caller needs to know why it failed, or when the failure is an error to
  report: a file that cannot be read, input that breaks a rule.

`String.toInt` returns `int | null`, since "not a number" is the whole story. `File.read` returns a
`Result`, since the caller needs to know whether the file is missing or unreadable.

A function that returns `T | null` passes an absence up with `?? return null`, as `try` passes an
`Err`:

```ant
type User = { id: int, name: string, managerId: int }

let users = Map{
  1: User{ id: 1, name: "Ann", managerId: 2 },
  2: User{ id: 2, name: "Bea", managerId: 9 },
}

func managerName(id: int): string | null {
  let user = users[id] ?? return null
  let manager = users[user.managerId] ?? return null
  manager.name
}

println(managerName(1))
println(managerName(2))
println(managerName(7))
```

```text
Bea
null
null
```

Where `null` from a lookup becomes an error for your function, turn it into a `Result` with
`?? return`:

```ant
type User = { id: int, name: string }

let users = Map{ 1: User{ id: 1, name: "Ann" } }

func userName(id: int): Result[string, Error] {
  let user = users[id] ?? return Result.fail("no user ${id}")
  Result.ok(user.name)
}

println(userName(1))
println(userName(2))
```

```text
Ok{ value: "Ann" }
Err{ error: Error{ message: "no user 2", kind: null } }
```

`try` unwraps a `Result` and has no nullable form; `?? return` is its counterpart for `null`.

A default stands for a missing value, not a wrong one. Apply `??` before you parse or validate, so
that a missing setting takes the default and a supplied bad one still reports an error:

```ant
func port(settings: Map[string, string]): Result[int, Error] {
  let text = settings["port"]
  if text == null { return Result.ok(8080) }
  let n = text.toInt() ?? return Result.fail("port: not a number: ${text}")
  Result.ok(n)
}

println(port(Map{}))
println(port(Map{ "port": "9000" }))
println(port(Map{ "port": "oops" }))
```

```text
Ok{ value: 8080 }
Ok{ value: 9000 }
Err{ error: Error{ message: "port: not a number: oops", kind: null } }
```

Writing `(settings["port"] ?? "8080").toInt() ?? 8080` instead would turn `port=oops` into 8080
without a word.

## Coming from other languages

- **No truthiness.** A condition must be a `bool`; `if name { ... }` on a `string | null` is a
  compile error. Write `if name != null`.
- **`?.` and `??` need a nullable left side.** On a plain `string`, they are dead code, and the type
  checker rejects them.
- **One `?.` per nullable link.** `user?.name?.length()` is a compile error when `name` is a
  `string`: the first `?.` already covers the rest of the chain.
- **`??` does not mix with comparisons.** `n ?? 0 > 2` is a syntax error; write `(n ?? 0) > 2`.
- **Compound assignment on a map element reads it first.** `counts[k] += 1` reads an `int | null`;
  write `counts[k] = (counts[k] ?? 0) + 1`.
- **Sums and extremes of an empty array are `null`.** `xs.sum()` is an `int | null`; write
  `xs.sum() ?? 0` for a zero.

```ant
let name: string | null = "Ann"
if name { println(name) } // error: cannot use string | null where bool is expected
let plain = "Ann"
let size = plain?.length() // error: ?. needs a nullable left side, got string
let user: string | null = "Ann"
let first = user?.trim()?.length()
// error: ?. needs a nullable left side, got string: an earlier ?. in the chain already skips the rest of it on null, so write . here
let counts = Map{ "a": 1 }
counts["a"] += 1
// error: += on a map element reads it first, and the read is int | null; write m[k] = (m[k] ?? init) + v
```

The mixed `??` stops the parser, so the type checker reports nothing else:

```ant
let n: int | null = 3
let big = n ?? 0 > 2
// error: '??' cannot mix with '>' without parentheses: write (a ?? b) > c or a ?? (b > c)
```

## See also

- [`??`](../language/operators/null-coalescing.md) and
  [`?.`](../language/operators/optional-chaining.md)
- [`null`](../language/literals/null.md)
- [`is`](../language/operators/is.md), which narrows unions the same way
- [Errors](errors.md), for `Result` and `try`
- [Unions](unions.md)
- [Idiomatic Ant](idiomatic.md)
