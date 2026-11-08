<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.sortWith

```ant signature
func Array.sortWith[T](xs: Array[T], cmp: func(T, T): int): Array[T]
```

Array.sortWith returns a new array holding the elements of xs ordered by cmp. cmp(a, b) \< 0 puts a
before b, and the sort is stable.

<!-- end generated -->

## Details

- `Array.sortWith` returns a new array with the elements of `xs` in the order `cmp` gives:
  `cmp(a, b) < 0` puts `a` before `b`, `cmp(a, b) > 0` puts `a` after `b`, and `cmp(a, b) == 0`
  keeps their order. The sort is stable, so elements that `cmp` ranks equal keep their order from
  `xs` ([spec: Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)).
- `Array.sortWith` calls `cmp` as often as sorting needs, in no promised order
  ([spec: Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)).
- When `cmp` is not a consistent order, the result still holds the elements of `xs`, in some order
  ([spec: Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)).
- `xs` is left untouched, and `Array.sortWith` reads its elements once, when it is called
  ([spec: Standard library](https://github.com/ant-src/spec/blob/master/ant-lang.md#11-standard-library)).
- `cmp` takes two parameters, so it is a `func` or a named function, never a lambda, which has the
  single parameter `x`
  ([spec: Anonymous functions](https://github.com/ant-src/spec/blob/master/ant-lang.md#anonymous-functions)).
- `T` has no bound: `cmp` decides the order, so records, unions, and any other type sort. For an
  `Ordered` element or key, [Array.sort](sort.md) and [Array.sortBy](sortBy.md) need no comparator
  ([spec: Generics](https://github.com/ant-src/spec/blob/master/ant-lang.md#generics)).

## Basic examples

Sort numbers in descending order:

```ant
Array.sortWith([3, 1, 2], func(a: int, b: int) { b - a }) // [3, 2, 1]
```

Call it through `.` on any array. `xs` keeps its order:

```ant
var xs = [5, 3, 8, 1]
xs.sortWith(func(a: int, b: int) { a - b }) // [1, 3, 5, 8]
xs                                          // [5, 3, 8, 1]
```

## Scope

An empty array sorts to an empty array. Annotate it so that `T` is known:

```ant
var empty: Array[string] = []
empty.sortWith(func(a: string, b: string) { 0 }) // []
```

A named function works as well as an anonymous one:

```ant
func byLength(a: string, b: string): int {
  String.length(a) - String.length(b)
}

["ccc", "a", "bb"].sortWith(byLength) // ["a", "bb", "ccc"]
```

In a pipeline, the array is the first argument:

```ant
func byLength(a: string, b: string): int {
  String.length(a) - String.length(b)
}

"the quick brown fox" | String.split(" ") | Array.sortWith(byLength) // ["the", "fox", "quick", "brown"]
```

The sort is stable: `"the"` and `"fox"` tie on length, and so do `"quick"` and `"brown"`, and each
pair keeps its order.

A record sorts by the fields its comparator compares. Here a lower priority number comes first, and
ties go by name:

```ant
type Job = { priority: int, name: string }
var jobs = [
  Job{ priority: 2, name: "lint" },
  Job{ priority: 1, name: "test" },
  Job{ priority: 1, name: "build" },
]
jobs.sortWith(func(a: Job, b: Job) {
  if a.priority != b.priority {
    a.priority - b.priority
  } else if a.name < b.name {
    -1
  } else if a.name > b.name {
    1
  } else {
    0
  }
}).map({ x.name }) // ["build", "test", "lint"]
```

A function on the record's type, such as `Job.compare`, is a named function, so it passes as `cmp`
directly:

```ant
type Job = { priority: int, name: string }

func Job.compare(a: Job, b: Job): int { a.priority - b.priority }

var jobs = [Job{ priority: 2, name: "lint" }, Job{ priority: 1, name: "test" }]
jobs.sortWith(Job.compare).map({ x.name }) // ["test", "lint"]
```

A generic function can sort any type that has its own `compare`, through an interface bound
([spec: Interfaces](https://github.com/ant-src/spec/blob/master/ant-lang.md#interfaces)):

```ant
interface Compare {
  compare(a: Self, b: Self): int
}

type Job = { priority: int, name: string }

func Job.compare(a: Job, b: Job): int { a.priority - b.priority }

func sorted[T: Compare](xs: Array[T]): Array[T] {
  xs.sortWith(func(a: T, b: T): int { a.compare(b) })
}

sorted([Job{ priority: 2, name: "lint" }, Job{ priority: 1, name: "test" }]).map({ x.name }) // ["test", "lint"]
```

The comparator tells the members of a union apart with `is`. This one puts every `Circle` before
every `Square` and keeps the rest of the order:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
func rank(s: Shape): int {
  if s is Circle { 0 } else { 1 }
}

var shapes: Array[Shape] = [Square{ side: 2.0 }, Circle{ r: 1.0 }, Square{ side: 1.0 }, Circle{ r: 3.0 }]
shapes.sortWith(func(a: Shape, b: Shape) { rank(a) - rank(b) }) // [Circle{ r: 1.0 }, Circle{ r: 3.0 }, Square{ side: 2.0 }, Square{ side: 1.0 }]
```

An element type that includes `null` sorts once the comparator says where `null` goes. Here it goes
last:

```ant
var scores: Array[int | null] = [70, null, 95, 80]
scores.sortWith(func(a: int | null, b: int | null) {
  if a == null && b == null {
    0
  } else if a == null {
    1
  } else if b == null {
    -1
  } else {
    b - a
  }
}) // [95, 80, 70, null]
```

The comparator is a closure, so it sees the variables it captures:

```ant
var descending = true
var cmp = func(a: int, b: int): int {
  if descending { b - a } else { a - b }
}
[2, 3, 1].sortWith(cmp) // [3, 2, 1]
descending = false
[2, 3, 1].sortWith(cmp) // [1, 2, 3]
```

## Applications

Sort words case-insensitively:

```ant
func caseless(a: string, b: string): int {
  var x = String.toLower(a)
  var y = String.toLower(b)
  if x < y { -1 } else if x > y { 1 } else { 0 }
}

["banana", "Cherry", "apple"].sortWith(caseless) // ["apple", "banana", "Cherry"]
```

A library function that returns a comparison result fits directly. `Time.compare` returns `-1`, `0`,
or `1`:

```ant
var dates = [Time.plainDate(2026, 3, 1), Time.plainDate(2025, 12, 24), Time.plainDate(2026, 1, 15)]
for d in dates.sortWith(Time.compare) {
  println(d)
}
```

```text
2025-12-24
2026-01-15
2026-03-01
```

Rank a leaderboard: the highest score first, and players with one score in the order they joined:

```ant
type Player = { name: string, score: int }
var players = [
  Player{ name: "Ann", score: 40 },
  Player{ name: "Bob", score: 55 },
  Player{ name: "Cid", score: 40 },
]
for p in players.sortWith(func(a: Player, b: Player) { b.score - a.score }) {
  println("${p.name} ${p.score}")
}
```

```text
Bob 55
Ann 40
Cid 40
```

## Properties and relations

For `int` elements whose differences do not overflow, the comparator `a - b` makes `Array.sortWith`
equal to `Array.sort`, and with a comparator on a key it is `Array.sortBy` on that key:

```ant
var xs = [4, 1, 3, 1, 2]
assert(xs.sortWith(func(a: int, b: int) { a - b }) == xs.sort())

var words = ["bee", "ant", "wasp", "fly"]
assert(words.sortWith(func(a: string, b: string) { String.length(a) - String.length(b) })
  == words.sortBy({ String.length(x) }))
```

Swapping the arguments of `cmp` reverses the order of unequal elements, so for distinct elements it
is the reverse of the ascending sort:

```ant
var xs = [4, 1, 3, 5, 2]
assert(xs.sortWith(func(a: int, b: int) { b - a }) == xs.sort().reverse())
```

The result has the same elements as `xs`, and sorting it again changes nothing:

```ant
var xs = [4, 1, 3, 1, 2]
var asc = func(a: int, b: int): int { a - b }
var ys = xs.sortWith(asc)
assert(ys.sort() == xs.sort())
assert(ys.sortWith(asc) == ys)
```

## Possible issues

A lambda has one parameter, and `cmp` takes two. Write a `func` with named parameters:

<!-- test: error a lambda has at most one parameter -->

```ant
println([3, 1, 2].sortWith({ 0 }))
```

The comparator returns an `int`, not a `bool`. A test like `a < b` is a compile error:

<!-- test: error cannot use func(int, int): bool where func(int, int): int is expected -->

```ant
println([3, 1, 2].sortWith(func(a: int, b: int) { a < b }))
```

Return a negative, zero, or positive `int`:

```ant
[3, 1, 2].sortWith(func(a: int, b: int) { if a < b { -1 } else if a > b { 1 } else { 0 } }) // [1, 2, 3]
```

`a - b` is short, but integer overflow panics. Near the ends of the range, compare instead of
subtracting:

```ant
[Math.maxInt, Math.minInt].sortWith(func(a: int, b: int) { a - b }) // panic: integer overflow
[Math.maxInt, Math.minInt].sortWith(func(a: int, b: int) {
  if a < b { -1 } else if a > b { 1 } else { 0 }
}) // [-9223372036854775808, 9223372036854775807]
```

The comparator returns an `int` for floats too, so `a - b` on `float64` elements is a compile error:

<!-- test: error cannot use func(float64, float64): float64 where func(float64, float64): int is expected -->

```ant
println([1.5, 0.5].sortWith(func(a: float64, b: float64) { a - b }))
```

Compare with `<` and `>` and return `-1`, `0`, or `1`:

```ant
[1.5, 0.5].sortWith(func(a: float64, b: float64) { if a < b { -1 } else if a > b { 1 } else { 0 } }) // [0.5, 1.5]
```

A comparator that is not a consistent order does not fail, but it promises no order: the result
holds the elements of `xs` in some order. Here `cmp` says every element goes after every other:

```ant
var xs = [4, 1, 3, 2]
var ys = xs.sortWith(func(a: int, b: int) { 1 })
assert(ys.sort() == xs.sort())
```

`cmp` runs in no promised order and as often as sorting needs, so a side effect in it, such as
counting or logging the pairs it sees, shows how one run happened to sort, not a rule to rely on.
Keep `cmp` a plain comparison.

The rule that puts `NaN` after every number belongs to `Array.sort` and `Array.sortBy`, not to
`Array.sortWith`: the comparator decides where `NaN` goes. For a plain ascending sort of floats that
may hold `NaN`, use `Array.sort`:

```ant
[1.5, Math.NaN, 0.5].sort() // [0.5, 1.5, NaN]
```

For any other order, place `NaN` in the comparator with `Math.isNaN`
([spec: Math and Bits](https://github.com/ant-src/spec/blob/master/ant-lang.md#math-and-bits)). This
one puts `NaN` last and the other values in descending order:

```ant
[1.5, Math.NaN, 0.5].sortWith(func(a: float64, b: float64) {
  if Math.isNaN(a) && Math.isNaN(b) {
    0
  } else if Math.isNaN(a) {
    1
  } else if Math.isNaN(b) {
    -1
  } else if a > b {
    -1
  } else if a < b {
    1
  } else {
    0
  }
}) // [1.5, 0.5, NaN]
```

## See also

[Array.sort](sort.md) · [Array.sortBy](sortBy.md) · [Array.reverse](reverse.md) ·
[Array.minBy](minBy.md) · [Array.maxBy](maxBy.md) · [Time.compare](../Time/compare.md)
