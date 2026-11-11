<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.map

```ant signature
func Array.map[T, U](xs: Array[T], f: func(T): U): Array[U]
```

Array.map returns a new array holding f applied to each element of xs, in order.

<!-- end generated -->

## Details

- Returns a new array; `xs` is unchanged.
- Calls `f` once per element, from first to last.
- On a sequence, `.map` is [`Iter.map`](../Iter/map.md), which is lazy.

## Basic examples

Double each number:

```ant
Array.map([1, 2, 3], { x * 2 }) // [2, 4, 6]
```

Turn each word into its length. The result has a different element type from `xs`:

```ant
Array.map(["ant", "beetle", "fly"], { x.length() }) // [3, 6, 3]
```

## Scope

An empty array gives an empty array:

```ant
let empty: Array[int] = []
empty.map({ x * 2 }) // []
```

Map over the characters of a string by splitting it into graphemes first:

```ant
"née".graphemes().map(String.toUpper) // ["N", "É", "E"]
```

Read a field of each record:

```ant
type User = { name: string, age: int }
let users = [User{ name: "Ann", age: 31 }, User{ name: "Bob", age: 27 }]
users.map({ x.name }) // ["Ann", "Bob"]
```

Build a record from each element:

```ant
type Item = { name: string, qty: int }
["pen", "ink"].map({ Item{ name: x, qty: 1 } }) // [Item{ name: "pen", qty: 1 }, Item{ name: "ink", qty: 1 }]
```

Replace the `null` elements of a nullable array with `??`, which gives an array that holds no
`null`:

```ant
let scores: Array[int | null] = [90, null, 75]
scores.map({ x ?? 0 }) // [90, 0, 75]
```

Handle each member of a union with `is`, which narrows `x` in each branch:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
shapes.map({ if x is Circle { 3.0 * x.radius * x.radius } else { x.width * x.height } }) // [3.0, 6.0]
```

Pass a named function, a function on a type, or `len` instead of a lambda:

```ant
func square(n: int): int { n * n }
type Point = { x: float64, y: float64 }
func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }
[1, 2, 3].map(square)                     // [1, 4, 9]
[Point{ x: 3.0, y: 4.0 }].map(Point.norm) // [5.0]
["ant", "日本"].map(len)                  // [3, 2]
```

A conversion is a function too. `byte` builds an `Array[byte]`, and a named type over `string` tags
every element:

```ant
type UserId = string
[72, 105].map(byte)      // [72, 105]
["u1", "u2"].map(UserId) // [UserId("u1"), UserId("u2")]
```

An anonymous function names and annotates its parameter:

```ant
[1.5, 2.25].map(func(price: float64): string { "$${price}" }) // ["$1.5", "$2.25"]
```

A closure reads the variables it captures:

```ant
let rate = 3
[10, 20].map({ x * rate }) // [30, 60]
```

Call it through `.` or in a pipe:

```ant
[1, 2, 3].map({ x + 1 })      // [2, 3, 4]
[1, 2, 3] | Array.map({ -x }) // [-1, -2, -3]
```

Name the arguments to pass them in any order:

```ant
Array.map(f: { x * x }, xs: [2, 3]) // [4, 9]
```

Map an array of arrays with a nested `map`:

```ant
let rows = [[1, 2], [3]]
rows.map({ x.map({ x * 10 }) }) // [[10, 20], [30]]
```

## Applications

Format each record as a line of text and join the lines:

```ant
type Account = { owner: string, balance: int }
let accounts = [Account{ owner: "Ann", balance: 120 }, Account{ owner: "Bob", balance: 75 }]
accounts | Array.map({ "${x.owner}: ${x.balance}" }) | String.join("\n") // "Ann: 120\nBob: 75"
```

Parse every field of a line. `String.toInt` returns `null` for text that is not an integer:

```ant
"4,8,x,15".split(",").map(String.toInt) // [4, 8, null, 15]
```

Parse every element and keep the first error: map with a function that returns a `Result`, then
`Result.all` returns the first `Err`, or `Ok` of every value:

```ant
func parse(s: string): Result[int, string] {
  let n = s.toInt()
  if n is int { Result.ok(n) } else { Result.err("not a number: ${s}") }
}
["4", "8", "15"] | Array.map(parse) | Result.all() // Ok{ value: [4, 8, 15] }
["4", "x", "15"] | Array.map(parse) | Result.all() // Err{ error: "not a number: x" }
```

Sort, then project, in one chain:

```ant
type Score = { player: string, points: int }
let scores = [Score{ player: "Ann", points: 7 }, Score{ player: "Bob", points: 12 }]
scores | Array.sortBy({ -x.points }) | Array.map({ x.player }) // ["Bob", "Ann"]
```

Number the elements with `Array.entries`, which pairs each with its index:

```ant
let steps = ["mix", "bake", "serve"]
for line in steps.entries().map({ "${x.index + 1}. ${x.value}" }) {
  println(line)
}
```

```text
1. mix
2. bake
3. serve
```

Render a list of HTML elements. An array of nodes renders its elements in order:

```ant
let tasks = ["write", "test"]
println(<ul>{tasks.map({ <li>{x}</li> })}</ul>)
```

```text
<ul><li>write</li><li>test</li></ul>
```

## Properties and relations

The result has the length of `xs`:

```ant
let xs = ["a", "bb", "ccc"]
assert(xs.map({ x.length() }).length() == xs.length())
```

Mapping with the identity gives an equal array that is not the same array:

```ant
let xs = [1, 2, 3]
let ys = xs.map({ x })
assert(ys == xs)
assert(!Array.same(ys, xs))
```

When `f` and `g` have no side effects, two maps in a row are one map with the composed function:

```ant
let xs = [1, 2, 3]
let f: func(int): int = { x + 1 }
let g: func(int): int = { x * 10 }
assert(xs.map(f).map(g) == xs.map({ g(f(x)) }))
```

Mapping each element to a one-element array and flattening is the same as mapping:

```ant
let xs = [1, 2, 3]
assert(xs.flatMap({ [x * 2] }) == xs.map({ x * 2 }))
```

`f` runs once per element, from the first to the last. A callback that logs what it sees shows the
order:

```ant
let seen: Array[string] = []
["a", "b", "c"].map({ seen.push(x); x.toUpper() }) // ["A", "B", "C"]
seen                                               // ["a", "b", "c"]
```

A callback that pushes to `xs` does not change what the call visits:

```ant
let xs = [1, 2, 3]
xs.map({ xs.push(x + 10); x * 2 }) // [2, 4, 6]
xs                                 // [1, 2, 3, 11, 12, 13]
```

## Possible issues

`Array.map` returns a new array and does not change `xs`. A call whose result is thrown away leaves
`xs` as it was:

```ant
let xs = [1, 2, 3]
xs.map({ x * 2 })
println(xs)
```

```text
[1, 2, 3]
```

Nested lambdas each bind their own `x`, so the inner lambda cannot read the outer element. Here
`x + x` is the inner element doubled:

```ant
[1, 2].map({ [10, 20].map({ x + x }) }) // [[20, 40], [20, 40]]
```

Use `func` for the outer function to name its parameter:

```ant
[1, 2].map(func(n: int) { [10, 20].map({ n + x }) }) // [[11, 21], [12, 22]]
```

`f` takes one argument, the element. It gets no index, and a function of two parameters is a compile
error; use `Array.entries` for the index, as in Applications:

```ant
let xs = ["a", "b"]
let labeled = xs.map(func(s: string, i: int): string { "${i}${s}" })
// error: cannot use func(string, int): string where func(string): U is expected
```

An expected type fixes `U` before the lambda is checked, so a lambda that returns another type is a
compile error. Integer literals are always `int`, even where a `float64` is expected:

```ant
let halves: Array[float64] = [1, 2].map({ x * 2 })
// error: cannot use int where float64 is expected
```

Convert in the lambda instead:

```ant
let halves: Array[float64] = [1, 2].map({ float64(x) / 2.0 })
halves // [0.5, 1.0]
```

A lambda whose body is an empty literal fixes nothing, so `U` stays unsolved and the call is a
compile error:

```ant
let lists = [1, 2].map({ [] })
// error: cannot infer the element type of an empty array; annotate it
```

Annotate the result to fix `U`:

```ant
let lists: Array[Array[int]] = [1, 2].map({ [] })
lists // [[], []]
```

The result holds what `f` returns. When `f` returns an element of `xs` that is itself an array or a
record, both arrays hold the same value, so writing to it shows through:

```ant
let grid = [[1], [2]]
let copy = grid.map({ x })
copy[0][0] = 9
grid // [[9], [2]]
```

## See also

[Array.flatMap](flatMap.md), [Array.filter](filter.md), [Array.reduce](reduce.md),
[Array.entries](entries.md), [Iter.map](../Iter/map.md), [Map.map](../Map/map.md),
[Result.map](../Result/map.md), [Parallel.map](../Parallel/map.md)
