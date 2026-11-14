<!-- status: reviewed -->

# `let`

`let` declares a fixed binding: a name that keeps the value it is declared with.

## Details

- A `let` needs an initializer.
- Assigning a `let` name is a compile error wherever the assignment is written: in a nested block, a
  lambda, a `go` or `defer` block, or another function.
- `let` fixes the binding, not the value. The fields, elements, and entries of a `let` value can
  still change.
- Function parameters, a lambda's `x`, and every name a `for` binds are `let` bindings too. Only a
  [var](var.md) can be reassigned.

## Basic examples

```ant
let limit = 100
limit * 2 // 200
```

Write the type after the name when the initializer does not give the one you want, such as an empty
array:

```ant
let scores: Array[int] = []
scores.length() // 0
```

A program that uses only fixed names:

```ant
let name = "Ant"
let greeting = "Hello, ${name}"
println(greeting)
```

```text
Hello, Ant
```

## Scope

The type may be wider than the initializer's, such as a union or a nullable type:

```ant
let id: int | string = 7
id is int // true
let nickname: string | null = null
nickname ?? "none" // "none"
```

A record destructuring binds fields by name, and `{ field: name }` renames one:

```ant
type Point = { x: int, y: int }
let { x, y: top } = Point{ x: 3, y: 4 }
x + top // 7
```

An array destructuring binds elements by position, and `...` binds the rest as a new array:

```ant
let [first, ...rest] = [1, 2, 3]
first // 1
rest  // [2, 3]
```

A `let` in a block is visible to the end of that block. An inner block, such as a function body, may
declare the same name, which hides the outer one there:

```ant
let label = "outer"
func describe(): string {
  let label = "inner"
  label
}
describe() // "inner"
label      // "outer"
```

A function value bound with `let` can call itself, since its body runs only after the binding
exists, as long as its type is written, here by the return type `: int`:

```ant
let factorial = func(n: int): int {
  if n <= 1 { 1 } else { n * factorial(n - 1) }
}
println(factorial(5))
```

```text
120
```

A `select` head can bind its result with `let`:

```ant
let ready: Channel[string] = Channel.new(1)
Channel.send(ready, "go")
let message = select {
  let m = Channel.receive(ready) { m ?? "closed" }
  Time.sleep(1s) { "timeout" }
}
println(message)
```

```text
go
```

A library exports a `let` with `export let`. The importer reads it and cannot assign it:

```ant [config.ant]
export let MaxRetries = 3
```

```ant
import "./config.ant" { MaxRetries }

for attempt in Array.range(1, MaxRetries + 1) {
  println("attempt ${attempt}")
}
```

```text
attempt 1
attempt 2
attempt 3
```

## Applications

Name an intermediate result that the code uses twice:

```ant
let words = "the quick brown fox jumps" | String.split(" ")
let longest = words.map({ x.length() }).max()
words.filter({ x.length() == longest }) // ["quick", "brown", "jumps"]
```

Each iteration of a `for` binds a fresh `let`, so a closure made in the body keeps that iteration's
value:

```ant
let handlers: Array[func(): int] = []
for i in [1, 2, 3] {
  handlers.push({ i * 10 })
}
handlers.map({ x() }) // [10, 20, 30]
```

A top-level nullable `let` stays narrowed inside a function that tests it, since nothing can
reassign it:

```ant
let greeting: string | null = "hi"

func shout(): string {
  if greeting == null {
    return ""
  }
  greeting.toUpper()
}

println(shout())
```

```text
HI
```

With `var greeting`, the call `greeting.toUpper()` is a compile error: only a `let` narrows inside a
function that tests it.

## Properties and relations

A `let` fixes the binding, not the value. A `let` array still grows:

```ant
let xs = [1]
xs.push(2)
assert(xs == [1, 2])
```

A `let` record's fields can still be assigned:

```ant
type Counter = { count: int }
let c = Counter{ count: 0 }
c.count += 1
assert(c.count == 1)
```

## Possible issues

Assigning a `let` is a compile error. Declare the name with [var](var.md) when it changes:

```ant
let total = 0
for n in [1, 2, 3] {
  total += n // error: cannot assign to total, a let; declare it with var to reassign it
}
```

```ant
var total = 0
for n in [1, 2, 3] {
  total += n
}
println(total)
```

```text
6
```

An assignment inside a lambda counts too, even if the lambda never runs:

```ant
let count = 0
let bump: func(): void = { count = count + 1 }
// error: cannot assign to count, a let; declare it with var to reassign it
```

A parameter, a lambda's `x`, and a `for` variable are `let` bindings, so assigning one is a compile
error. Bind a `var` to hold a value that changes:

```ant
func double(n: int): int {
  n = n * 2 // error: cannot assign to n, a parameter; declare a var to hold a value that changes
  n
}
for s in ["a"] {
  s = s.toUpper() // error: cannot assign to s, a for variable; declare a var to hold a value that changes
}
```

A `let` in an array destructuring assignment is an assignment too, so a swap needs two `var`s:

```ant
var a = 1
let b = 2
[a, b] = [b, a] // error: cannot assign to b, a let; declare it with var to reassign it
```

A function value that calls itself needs its type written. Without a return type, its body would be
inferred from a call to itself:

```ant
let factorial = func(n: int) {
  if n <= 1 { 1 } else { n * factorial(n - 1) } // error: factorial is used by a function in its own initializer; state its type
}
```

Declaring a name twice in one scope is an error, even with `let` both times:

```ant
let limit = 10
let limit = 20 // error: limit is already declared in this scope
```

A `let` may not take a prelude name, in any scope:

```ant
let println = "hi" // error: println is a prelude name and cannot be redeclared
```

An array destructuring with more positions than the array has elements panics:

```ant
let [a, b] = [1] // panic: a destructuring needs 2 elements, got 1
```

## See also

[var](var.md) · [for](for.md) · [select](select.md) · [export](export.md) · [func](func.md) ·
[Assignment](../operators/assignment.md) · [Variables tour](../../tour/02-variables.md) ·
[Idiomatic Ant](../../guide/idiomatic.md)
