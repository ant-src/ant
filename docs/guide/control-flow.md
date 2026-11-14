<!-- status: reviewed -->

# Control flow

Ant has few control-flow constructs, and each of them is an expression. A block gives the value of
its last expression, `if` gives the value of the branch it runs, and `return`, `break`, and
`continue` fit anywhere a value does. Loops are the exception: `for` and `while` do work and give no
value, apart from a `while true` that only `return` or a panic can leave. This chapter shows how
these parts fit together, which one to pick for a job, and where each of them stops.

## Blocks give a value

A block is a sequence of expressions between braces. Its value is the value of the last one, so a
function ends in its result instead of a `return`:

```ant
func shippingCost(weightKg: int, express: bool): int {
  let base = 5 + weightKg * 2
  if express { base * 2 } else { base }
}

println(shippingCost(3, false))
println(shippingCost(3, true))
```

```text
11
22
```

The body of `shippingCost` binds `base` and then ends in an `if`. The `if` is the last expression,
so its value is the function's result, and each branch is a block whose value is its only
expression.

A block that ends in a declaration, an assignment, or a loop has type `void`, apart from the
`while true` described under [Loops give no value](#loops-give-no-value). A function that promises a
value must end in that value, even after a loop that computed it:

```ant
func total(prices: Array[int]): int {
  var sum = 0
  for p in prices { sum += p }
  // error: cannot use void where int is expected
}
```

The checker reports the loop, since it is the last expression. End the body with `sum`.

Braces are a block only after `func(...)`, `else`, `go`, or `defer`, and at the end of an `if`,
`while`, or `for` header or of a `select` clause's head. After a type name they build a value, such
as `Point{ x: 1 }`, and anywhere else they are a [lambda](../language/literals/lambdas.md). Ant has
no free-standing block: to group a few steps under their own scope, put them in a function.

## `if`

`if` runs one of two branches by a `bool` condition. As a statement, its branches do work; as an
expression, they give the values to choose from. Ant has no `?:` operator, because an `if`
expression does that job:

```ant
let stock = 3
let label = if stock == 0 { "sold out" } else if stock < 5 { "low stock" } else { "in stock" }
label // "low stock"
```

An `if` without `else` gives `void` when its condition is `false`, so it cannot fill a place that
expects a value:

```ant
let stock = 3
let note: string = if stock < 5 { "order more" }
// error: an if without else has type void when the condition is false, which cannot be used where string is expected
```

Add an `else` with the value for the other case.

The condition takes no parentheses, as in Go, and the `{` stays on the line of the header. A record
or map literal in a header needs parentheses, because its `{` would otherwise start the body:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 0, y: 0 }
if p == (Point{ x: 0, y: 0 }) { println("origin") }
```

```text
origin
```

An `if` that tests for `null` or for a union member narrows the variable in its branch, and a guard
that leaves with `return` narrows it for the rest of the block. [Null](null.md) and
[Unions](unions.md) cover narrowing.

## Loops

### `for`

`for x in e` runs its body once for each element of `e`. It walks an array, a map, a set, a string,
`Bytes`, or a sequence. Over a map, each element is an `Entry`, which the loop can destructure:

```ant
let stock = Map{ "tea": 4, "coffee": 0, "cocoa": 12 }
for { key, value } in stock {
  println("${key}: ${value}")
}

for c in "añb" {
  println(c)
}
```

```text
tea: 4
coffee: 0
cocoa: 12
a
ñ
b
```

A map yields its entries in insertion order. A string yields each code point as a one-character
string.

Ant has no C-style `for (i = 0; i < n; i++)`. To count, loop over
[`Iter.range`](../lib/Iter/range.md), which yields the integers from its first argument up to but
excluding its second:

```ant
for i in Iter.range(1, 4) {
  println("round ${i}")
}
```

```text
round 1
round 2
round 3
```

To get each element's index as well, loop over [`entries`](../lib/Array/entries.md), which pairs
each element with its index as an `Indexed` record with the fields `index` and `value`:

```ant
let names = ["Ann", "Bob"]
for { index, value } in names.entries() {
  println("${index + 1}. ${value}")
}
```

```text
1. Ann
2. Bob
```

Each iteration binds a fresh `let`. A closure created in the body keeps the value of its own
iteration, and assigning the loop variable is a compile error, since it would not change the
collection:

```ant
let multiples: Array[func(): int] = []
for i in Iter.range(0, 3) {
  multiples.push({ i * 10 })
}
println(multiples.map({ x() }))
```

```text
[0, 10, 20]
```

```ant
let prices = [10, 20]
for p in prices { p = p * 2 }
// error: cannot assign to p, a for variable; declare a var to hold a value that changes
```

To change the elements, assign `xs[index]` in a loop over `xs.entries()`, or build a new array with
[`Array.map`](../lib/Array/map.md).

Over an array, `for` reads the length before each iteration, so it visits the elements the body
pushes. That makes an array a work queue:

```ant
let queue = ["docs"]
let children = Map{ "docs": ["docs/guide", "docs/lib"], "docs/lib": ["docs/lib/Array"] }
for dir in queue {
  for child in children[dir] ?? [] { queue.push(child) }
}
println(queue)
```

```text
["docs", "docs/guide", "docs/lib", "docs/lib/Array"]
```

### `while`

`while` repeats its body as long as its condition is `true`. Use it when the loop runs on a
condition rather than over a collection:

```ant
var balance = 1000
var years = 0
while balance < 2000 {
  balance += balance / 10
  years += 1
}
println("${years} years, ${balance}")
```

```text
8 years, 2142
```

Ant has no `do ... while`. A `while true` loop whose body leaves with `return` or `break` covers it.

### Loops give no value

`for` and `while` have type `void`, except a `while true` with no `break`: only `return` or a panic
can leave it, so it has type `never` and can end a function that returns a value. That is the
`do ... while` replacement above when its body leaves with `return`:

```ant
func firstSquareAbove(limit: int): int {
  var n = 1
  while true {
    if n * n > limit { return n * n }
    n += 1
  }
}

println(firstSquareAbove(50))
```

```text
64
```

Any other loop that builds a result updates a `var` or pushes into an array declared before it, and
the code after the loop uses that.

A transformation of one collection into another reads better as a pipeline. Pick a loop when the
work needs an early exit, a `try`, or several updates per element; pick
[`some`](../lib/Array/some.md), [`every`](../lib/Array/every.md), [`find`](../lib/Array/find.md), or
[`filter`](../lib/Array/filter.md) when the result is the question you are asking:

```ant
let orders = [12, 40, 7, 55]
let large = orders | Array.filter({ x > 30 }) | Array.map({ x * 2 })
large                   // [80, 110]
orders.some({ x > 50 }) // true
orders.find({ x < 10 }) // 7
orders.every({ x > 0 }) // true
```

## `return`, `break`, and `continue`

`return` leaves a function, `break` leaves the innermost loop, and `continue` goes on to its next
iteration. All three have type `never`, so they can stand where a value is expected: as a branch of
an `if` expression, or on the right of `??`.

```ant
let ages = Map{ "Ann": 31, "Bob": 17 }
for name in ["Ann", "Cid", "Bob"] {
  let age = ages[name] ?? continue
  if age < 18 { break }
  println("${name} is ${age}")
}
```

```text
Ann is 31
```

`ages[name]` is `null` for `"Cid"`, so `?? continue` skips to the next name. Below that line, `age`
is an `int`.

`return` takes the whole expression to its right: `x ?? return y + 1` returns `y + 1`. A `return` at
the end of a line returns `void`.

`break` and `continue` have no labels. To leave nested loops at once, put them in a function and
`return` from it:

```ant
func findPair(xs: Array[int], target: int): Array[int] | null {
  for a in xs {
    for b in xs {
      if a < b && a + b == target { return [a, b] }
    }
  }
  null
}

println(findPair([3, 9, 4, 6], 10))
println(findPair([1, 2], 10))
```

```text
[4, 6]
null
```

## Where they stop

A function body, a script body, a lambda, a `go` block, and a `defer` block are boundaries. `return`
leaves the innermost one, and `break` and `continue` never cross one to reach a loop outside it.

### Lambdas

A `return` in a lambda leaves the lambda, not the function that contains it. The function keeps
running with the lambda's value:

```ant
func labels(scores: Array[int]): Array[string] {
  scores.map({
    if x < 0 { return "invalid" }
    "${x} points"
  })
}

println(labels([10, -1, 7]))
```

```text
["10 points", "invalid", "7 points"]
```

This is how `return` in a JavaScript `forEach` callback behaves. When the body must leave the
function, or stop early, write a `for` loop instead of a callback. A `break` in a lambda has no loop
to leave, even when the lambda sits inside one:

```ant
for row in [[1, 2], [3, -1]] {
  let doubled = row.map({
    if x < 0 { break }
    // error: 'break' outside a loop
    x * 2
  })
}
```

### `go` and `defer`

A `go` block runs on its own goroutine, and a `defer` block runs later, when its block exits.
Neither has a caller waiting for a result, so `return` in them takes no value and ends only that
block:

```ant
go {
  return 1
  // error: 'return' takes no value in a go block
}
```

The [Concurrency](concurrency.md) chapter shows how a goroutine hands back a result.

### `defer` runs when its block exits

`defer` registers a block to run when the enclosing block exits: at its end, or by `return`,
`break`, `continue`, or a panic. Deferred blocks run in reverse order. A `defer` in a loop body
belongs to that iteration, so it runs before the next one starts, even when `continue` ends the
iteration early:

```ant
for name in ["a.txt", "b.tmp", "c.txt"] {
  println("open ${name}")
  defer { println("close ${name}") }
  if name.endsWith(".tmp") { continue }
  println("read ${name}")
}
```

```text
open a.txt
read a.txt
close a.txt
open b.tmp
close b.tmp
open c.txt
read c.txt
close c.txt
```

To keep a resource open for the whole loop, acquire it and `defer` its cleanup before the loop. The
[Errors](errors.md) chapter covers `defer` with `Result` and panics.

## Picking a construct

| You want to                                   | Write                                          |
| --------------------------------------------- | ---------------------------------------------- |
| Choose between two values                     | `if c { a } else { b }`                        |
| Leave early on bad input                      | a guard: `if bad { return ... }`               |
| Fall back when a value is `null`              | `x ?? default`, or `x ?? return ...`           |
| Do something for each element                 | `for x in xs { ... }`                          |
| Count from `a` up to `b`                      | `for i in Iter.range(a, b) { ... }`            |
| Loop with the index                           | `for { index, value } in xs.entries() { ... }` |
| Repeat until a condition changes              | `while cond { ... }`                           |
| Transform a collection                        | a pipeline of `Array.filter`, `Array.map`, ... |
| Ask whether any, every, or which element fits | `some`, `every`, `find`                        |
| Stop at the first failure                     | a `for` loop with `try`                        |
| Leave nested loops                            | a function and `return`                        |

## See also

- Reference: [`if`](../language/keywords/if.md), [`for`](../language/keywords/for.md),
  [`while`](../language/keywords/while.md), [`return`](../language/keywords/return.md),
  [`break`](../language/keywords/break.md), [`continue`](../language/keywords/continue.md),
  [`defer`](../language/keywords/defer.md), [`go`](../language/keywords/go.md)
- [`??`](../language/operators/null-coalescing.md), [Iter.range](../lib/Iter/range.md)
- [Collections and sequences](collections.md) for writing a sequence that `for` can walk
- [Idiomatic Ant](idiomatic.md)
