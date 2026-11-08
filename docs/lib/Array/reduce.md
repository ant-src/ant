<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.reduce

```ant signature
func Array.reduce[T, A](xs: Array[T], init: A, f: func(A, T): A): A
```

Array.reduce folds xs from the left, starting from init and calling f with the accumulator and each
element in turn. It returns init when xs is empty.

<!-- end generated -->

## Details

- `Array.reduce` calls `f(acc, x)` for each element `x` of `xs`, from the first to the last. `acc`
  is `init` on the first call and the result of the previous call after that. The result of the last
  call is the result of `reduce`. The doc comment above states this order; the spec gives only the
  signature.
- On an empty array, `f` is never called and the result is `init`, as the doc comment says.
- The accumulator type `A` may differ from the element type `T`. A generic call solves its type
  parameters from the expected type first, then from the non-lambda arguments left to right, then
  from the lambdas. An anonymous function with annotated parameters is not a lambda, so `f` can fix
  `A` too, and a lambda passed as `init` is checked last
  ([spec: Inference](https://github.com/ant-src/spec/blob/master/ant-lang.md#inference)).
- `f` takes two parameters, so it is any function value of type `func(A, T): A`, such as an
  anonymous function with annotated parameters, a named function, or a variable holding one. A
  lambda `{ ... }` has at most one parameter and does not fit
  ([spec: Anonymous functions](https://github.com/ant-src/spec/blob/master/ant-lang.md#anonymous-functions)).
- `Array.reduce` reads the elements once, when it is called. A callback that pushes to `xs` changes
  `xs` but not what the call visits
  ([spec: Standard library](https://github.com/ant-src/spec/blob/master/ant-lang.md#11-standard-library)).

## Basic examples

Add up an array:

```ant
Array.reduce([1, 2, 3, 4], 0, func(acc: int, x: int) { acc + x }) // 10
```

Multiply instead, starting from 1:

```ant
[1, 2, 3, 4].reduce(1, func(acc: int, x: int) { acc * x }) // 24
```

## Scope

An empty array gives `init` back:

```ant
var empty: Array[int] = []
empty.reduce(42, func(acc: int, x: int) { acc + x }) // 42
```

The fold goes from the left, so an operator that is not associative shows the order:

```ant
[1, 2, 3].reduce(0, func(acc: int, x: int) { acc - x }) // -6
```

```ant
var calls: Array[string] = []
[1, 2, 3].reduce(0, func(acc: int, x: int) {
  calls.push("f(${acc}, ${x})")
  acc + x
})
println(calls)
```

```text
["f(0, 1)", "f(1, 2)", "f(3, 3)"]
```

There is no right fold. Reverse the array first:

```ant
Array.reverse([1, 2, 3]).reduce(0, func(acc: int, x: int) { acc * 10 + x }) // 321
```

The accumulator may have another type than the elements. Here it is an `int` and the elements are
strings:

```ant
["ant", "beetle", "fly"].reduce(0, func(n: int, s: string) { n + String.length(s) }) // 12
```

Join strings by starting from `""`:

```ant
["a", "n", "t"].reduce("", func(s: string, x: string) { s + x }) // "ant"
```

A named function with the right shape works as `f`, and so does a generic one such as `Math.max`:

```ant
func add(a: int, b: int): int {
  a + b
}

Array.reduce([5, 6, 7], 0, add)                 // 18
[3, 9, 4].reduce(Math.minInt, Math.max)         // 9
[2.5, 0.5, 1.5].reduce(Math.Infinity, Math.min) // 0.5
```

A record accumulator carries several values through one pass:

```ant
type Stats = { sum: int, count: int }
var s = [4, 8, 6].reduce(Stats{ sum: 0, count: 0 }, func(a: Stats, x: int) {
  Stats{ sum: a.sum + x, count: a.count + 1 }
})
s // Stats{ sum: 18, count: 3 }
```

A nullable accumulator starts from `null`. Annotate the result so that `A` is `int | null` rather
than `null`:

```ant
var longest: string | null = ["ant", "termite", "bee"].reduce(null, func(best: string | null, w: string): string | null {
  if best == null || String.length(w) > String.length(best) { w } else { best }
})
longest // "termite"
```

A closure reads the variables it captures:

```ant
var rate = 2
[1, 2, 3].reduce(0, func(acc: int, x: int) { acc + x * rate }) // 12
```

Call it through `.` or in a pipe:

```ant
[1, 2, 3].reduce(0, func(a: int, b: int) { a + b })          // 6
[1, 2, 3] | Array.reduce(10, func(a: int, b: int) { a + b }) // 16
```

A direct call may name `init` and `f`:

```ant
Array.reduce([1, 2, 3], init: 100, f: func(a: int, b: int) { a - b }) // 94
```

## Applications

Count the words of a text in a map. The callback changes the map and returns it:

```ant
var counts: Map[string, int] = "the ant and the bee".split(" ").reduce(Map{}, func(m: Map[string, int], w: string) {
  m[w] = (m[w] ?? 0) + 1
  m
})
counts // Map{ "the": 2, "ant": 1, "and": 1, "bee": 1 }
```

Total the lines of an order:

```ant
type Line = { item: string, qty: int, price: int }
var order = [
  Line{ item: "pen", qty: 3, price: 150 },
  Line{ item: "pad", qty: 1, price: 400 },
]
var total = order.reduce(0, func(sum: int, l: Line) { sum + l.qty * l.price })
println("total: ${total} cents")
```

```text
total: 850 cents
```

Apply a list of changes to a starting state. Each step is a union member:

```ant
type Deposit = { amount: int }
type Withdraw = { amount: int }
type Op = Deposit | Withdraw
var ops: Array[Op] = [Deposit{ amount: 100 }, Withdraw{ amount: 30 }, Deposit{ amount: 5 }]
ops.reduce(0, func(balance: int, op: Op) {
  if op is Deposit { balance + op.amount } else { balance - op.amount }
}) // 75
```

Compose an array of functions into one:

```ant
var steps: Array[func(int): int] = [{ x + 1 }, { x * 10 }, { x - 3 }]
var pipeline = steps.reduce({ x }, func(f: func(int): int, g: func(int): int): func(int): int {
  { g(f(x)) }
})
pipeline(4) // 47
```

## Properties and relations

Folding with `+` from `0` is what `Array.sum` gives on a non-empty array:

```ant
var xs = [3, 1, 4, 1, 5]
assert(xs.reduce(0, func(a: int, b: int) { a + b }) == Array.sum(xs))
```

A fold that pushes each element into a new array copies `xs`, and one that puts each in front
reverses it:

```ant
var xs = [1, 2, 3]
var empty: Array[int] = []
assert(xs.reduce(empty, func(a: Array[int], x: int) { a.concat([x]) }) == xs)
assert(xs.reduce(empty, func(a: Array[int], x: int) { [x].concat(a) }) == Array.reverse(xs))
```

`Array.map` is a fold that appends `f(x)`:

```ant
var xs = [1, 2, 3]
var empty: Array[int] = []
var double: func(int): int = { x * 2 }
assert(xs.reduce(empty, func(a: Array[int], x: int) { a.concat([double(x)]) }) == xs.map(double))
```

Folding `xs.concat(ys)` is folding `ys` from the result of folding `xs`:

```ant
var xs = [1, 2]
var ys = [3, 4]
var f = func(a: int, b: int) { a * 10 + b }
assert(xs.concat(ys).reduce(0, f) == ys.reduce(xs.reduce(0, f), f))
```

A callback that pushes to `xs` does not change what the call visits:

```ant
var xs = [1, 2, 3]
xs.reduce(0, func(a: int, x: int) { xs.push(x); a + x }) // 6
xs                                                       // [1, 2, 3, 1, 2, 3]
```

## Possible issues

`init` is required; there is no form that starts from the first element. For the largest or smallest
element without a starting value, use `Array.max` or `Array.min`, which return `null` on an empty
array.

A lambda has one parameter, `x`, so it cannot be the two-parameter `f`. Write `func` with annotated
parameters:

<!-- test: error a lambda has at most one parameter -->

```ant
var total = [1, 2, 3].reduce(0, { x })
```

The type of `init` fixes `A`. The integer literal `0` is an `int`, so summing floats from `0` is a
compile error:

<!-- test: error cannot use func(float64, float64): float64 where func(int, float64): int is expected -->

```ant
var total = [1.5, 2.5].reduce(0, func(a: float64, b: float64) { a + b })
```

Start from `0.0`:

```ant
[1.5, 2.5].reduce(0.0, func(a: float64, b: float64) { a + b }) // 4.0
```

Without an annotation, `null` as `init` makes `A` the type `null`, and a callback that returns
anything else does not fit:

<!-- test: error where func(null, int): null is expected -->

```ant
var last = [1, 2].reduce(null, func(a: int | null, x: int): int | null { x })
```

An array, a map, or a record is a reference. When `f` mutates the accumulator, it mutates the value
`init` refers to, so a variable passed as `init` changes too:

```ant
var seen: Array[int] = []
var copy = [1, 2].reduce(seen, func(a: Array[int], x: int) { a.push(x); a })
seen                   // [1, 2]
Array.same(copy, seen) // true
```

Integer overflow in `f` panics, as it does anywhere else:

```ant
[Math.maxInt, 1].reduce(0, func(a: int, b: int) { a + b }) // panic: integer overflow
```

## See also

[Array.sum](sum.md) · [Array.map](map.md) · [Array.filter](filter.md) ·
[Array.countWhere](countWhere.md) · [Array.groupBy](groupBy.md) · [Array.flatMap](flatMap.md) ·
[Array.reverse](reverse.md) · [Array.max](max.md) · [Array.min](min.md) · [Math.max](../Math/max.md)
