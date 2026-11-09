<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.flatMap

```ant signature
func Array.flatMap[T, U](xs: Array[T], f: func(T): Array[U]): Array[U]
```

Array.flatMap applies f to each element of xs and returns the arrays it returns, concatenated in
order.

<!-- end generated -->

## Details

- The result holds the arrays `f` returns, joined in the order of `xs`. An empty array from `f` adds
  nothing.
- Calls `f` once per element, from first to last.
- Returns a new array; `xs` is unchanged.

## Basic examples

Follow each number with ten times it:

```ant
Array.flatMap([1, 2, 3], { [x, x * 10] }) // [1, 10, 2, 20, 3, 30]
```

Collect the items of every order into one array:

```ant
type Order = { id: int, items: Array[string] }
var orders = [Order{ id: 1, items: ["pen", "ink"] }, Order{ id: 2, items: [] }, Order{ id: 3, items: ["pad"] }]
Array.flatMap(orders, { x.items }) // ["pen", "ink", "pad"]
```

## Scope

The arrays `f` returns may have different lengths:

```ant
Array.flatMap([1, 2, 3], { Array.filled(x, x) }) // [1, 2, 2, 3, 3, 3]
```

An empty `xs` gives an empty array, and `f` is never called:

```ant
var none: Array[int] = []
Array.flatMap(none, { [x, x] }) // []
```

`f` runs once per element, in order, including the elements it drops; a side effect shows the calls:

```ant
var calls: Array[string] = []
var out = Array.flatMap(["a", "b", "c"], func(s: string): Array[string] {
  calls.push(s)
  if s == "b" { [] } else { [s, s] }
})
calls // ["a", "b", "c"]
out   // ["a", "a", "c", "c"]
```

An `f` that returns its argument flattens an array of arrays:

```ant
Array.flatMap([[1, 2], [3]], { x }) // [1, 2, 3]
```

A named function that returns an array works as `f`, such as `String.graphemes` on Unicode text:

```ant
Array.flatMap(["hi", "日本"], String.graphemes) // ["h", "i", "日", "本"]
```

A union element is narrowed with `is` inside `f`:

```ant
type Circle = { r: int }
type Square = { side: int }
type Shape = Circle | Square
var shapes: Array[Shape] = [Circle{ r: 1 }, Square{ side: 2 }]
Array.flatMap(shapes, { if x is Circle { [x.r] } else { [x.side, x.side] } }) // [1, 2, 2]
```

Drop the `null`s of an array; `x` is narrowed in each branch:

```ant
var names: Array[string | null] = ["ann", null, "bo"]
var kept: Array[string] = Array.flatMap(names, { if x == null { [] } else { [x] } })
kept // ["ann", "bo"]
```

An expected type fixes `U`, so `f` can return elements of a union:

```ant
var mixed: Array[int | string] = Array.flatMap([1, 2], { [x, "${x}"] })
mixed // [1, "1", 2, "2"]
```

Call it through `.` on an array, or with a pipe:

```ant
["a b", "c"].flatMap({ String.split(x, " ") })         // ["a", "b", "c"]
["a b", "c"] | Array.flatMap({ String.split(x, " ") }) // ["a", "b", "c"]
```

Name the arguments to pass them in any order:

```ant
Array.flatMap(f: { [x, x] }, xs: [1, 2]) // [1, 1, 2, 2]
```

`Array.flatMap` visits the elements `xs` had at the call; elements that `f` pushes are not visited:

```ant
var a = [1, 2]
Array.flatMap(a, { Array.push(a, x + 10); [x] }) // [1, 2]
a                                                // [1, 2, 11, 12]
```

## Applications

Split lines into words:

```ant
var text = "the quick fox\njumps over"
String.lines(text).flatMap({ String.split(x, " ") }) // ["the", "quick", "fox", "jumps", "over"]
```

Map and filter in one pass: return `[]` to drop an element and `[v]` to keep `v`. Here `f` drops the
strings that are not integers:

```ant
var parsed = Array.flatMap(["1", "x", "3"], func(s: string): Array[int] {
  var n = String.toInt(s)
  if n == null { [] } else { [n] }
})
parsed // [1, 3]
```

Build every pair of two arrays. The inner lambda binds its own `x`, so the outer element goes
through a `func` parameter:

```ant
Array.flatMap([1, 2], func(n: int) { Array.map(["x", "y"], { "${n}${x}" }) }) // ["1x", "1y", "2x", "2y"]
```

Expand each order line into one entry per unit:

```ant
type Line = { sku: string, qty: int }
var lines = [Line{ sku: "A", qty: 2 }, Line{ sku: "B", qty: 1 }]
for sku in Array.flatMap(lines, { Array.filled(x.qty, x.sku) }) {
  println(sku)
}
```

```text
A
A
B
```

## Properties and relations

It is [Array.map](map.md) followed by [Array.flatten](flatten.md):

```ant
var xs = [1, 2, 3]
var f: func(int): Array[int] = { [x, -x] }
assert(Array.flatMap(xs, f) == Array.flatten(Array.map(xs, f)))
```

The length of the result is the sum of the lengths `f` returns:

```ant
var xs = [1, 2, 3]
var f: func(int): Array[int] = { Array.filled(x, 0) }
var total = Array.sum(Array.map(xs, { Array.length(f(x)) })) ?? 0
assert(Array.length(Array.flatMap(xs, f)) == total)
```

An `f` that wraps each element in a one-element array gives back a copy of `xs`, and one that wraps
`g(x)` is `Array.map(xs, g)`:

```ant
var xs = [1, 2, 3]
assert(Array.flatMap(xs, { [x] }) == xs)
assert(!Array.same(Array.flatMap(xs, { [x] }), xs))
assert(Array.flatMap(xs, { [x * 2] }) == Array.map(xs, { x * 2 }))
```

An `f` that returns `[x]` or `[]` is [Array.filter](filter.md):

```ant
var xs = [1, 2, 3, 4]
var evens: Array[int] = Array.flatMap(xs, { if x % 2 == 0 { [x] } else { [] } })
assert(evens == Array.filter(xs, { x % 2 == 0 }))
```

## Possible issues

`f` must return an array. A lambda that returns a single value is a compile error; use
[Array.map](map.md) for that:

<!-- test: error the lambda returns int where Array[U] is expected -->

```ant
var doubled = Array.flatMap([1, 2], { x * 2 })
```

When one branch of `f` returns `[]`, the empty array does not fix `U`, so the call is a compile
error. Annotate the result:

<!-- test: error cannot infer the element type of an empty array -->

```ant
var odd = Array.flatMap([1, 2, 3], { if x % 2 == 0 { [] } else { [x] } })
```

```ant
var odd: Array[int] = Array.flatMap([1, 2, 3], { if x % 2 == 0 { [] } else { [x] } })
odd // [1, 3]
```

The result holds the elements of the arrays `f` returns, not copies of them. An element that is
itself an array or a record is the same value in both, so writing to it shows through:

```ant
var grid = [[1], [2]]
var flat = Array.flatMap([0], { grid })
flat[0][0] = 9
grid // [[9], [2]]
```

`Array.flatMap` flattens one level only. Arrays nested deeper stay arrays:

```ant
Array.flatMap([[[1], [2]], [[3]]], { x }) // [[1], [2], [3]]
```

## See also

[Array.map](map.md), [Array.flatten](flatten.md), [Array.filter](filter.md),
[Array.concat](concat.md), [Array.reduce](reduce.md), [String.graphemes](../String/graphemes.md),
[String.split](../String/split.md)
