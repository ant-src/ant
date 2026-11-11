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
let orders = [Order{ id: 1, items: ["pen", "ink"] }, Order{ id: 2, items: [] }, Order{ id: 3, items: ["pad"] }]
Array.flatMap(orders, { x.items }) // ["pen", "ink", "pad"]
```

## Scope

The arrays `f` returns may have different lengths:

```ant
[1, 2, 3].flatMap({ Array.filled(x, x) }) // [1, 2, 2, 3, 3, 3]
```

An empty `xs` gives an empty array, and `f` is never called:

```ant
let none: Array[int] = []
none.flatMap({ [x, x] }) // []
```

`f` runs once per element, in order, including the elements it drops; a side effect shows the calls:

```ant
let calls: Array[string] = []
let out = ["a", "b", "c"].flatMap(func(s: string): Array[string] {
  calls.push(s)
  if s == "b" { [] } else { [s, s] }
})
calls // ["a", "b", "c"]
out   // ["a", "a", "c", "c"]
```

An `f` that returns its argument flattens an array of arrays:

```ant
[[1, 2], [3]].flatMap({ x }) // [1, 2, 3]
```

A named function that returns an array works as `f`, such as `String.graphemes` on Unicode text:

```ant
["hi", "日本"].flatMap(String.graphemes) // ["h", "i", "日", "本"]
```

A union element is narrowed with `is` inside `f`:

```ant
type Circle = { r: int }
type Square = { side: int }
type Shape = Circle | Square
let shapes: Array[Shape] = [Circle{ r: 1 }, Square{ side: 2 }]
shapes.flatMap({ if x is Circle { [x.r] } else { [x.side, x.side] } }) // [1, 2, 2]
```

Drop the `null`s of an array; `x` is narrowed in each branch:

```ant
let names: Array[string | null] = ["ann", null, "bo"]
let kept: Array[string] = names.flatMap({ if x == null { [] } else { [x] } })
kept // ["ann", "bo"]
```

An expected type fixes `U`, so `f` can return elements of a union:

```ant
let mixed: Array[int | string] = [1, 2].flatMap({ [x, "${x}"] })
mixed // [1, "1", 2, "2"]
```

Call it through `.` on an array, or with a pipe:

```ant
["a b", "c"].flatMap({ x.split(" ") })         // ["a", "b", "c"]
["a b", "c"] | Array.flatMap({ x.split(" ") }) // ["a", "b", "c"]
```

Name the arguments to pass them in any order:

```ant
Array.flatMap(f: { [x, x] }, xs: [1, 2]) // [1, 1, 2, 2]
```

`Array.flatMap` visits the elements `xs` had at the call; elements that `f` pushes are not visited:

```ant
let a = [1, 2]
a.flatMap({ a.push(x + 10); [x] }) // [1, 2]
a                                  // [1, 2, 11, 12]
```

## Applications

Split lines into words:

```ant
let text = "the quick fox\njumps over"
text.lines().flatMap({ x.split(" ") }) // ["the", "quick", "fox", "jumps", "over"]
```

Map and filter in one pass: return `[]` to drop an element and `[v]` to keep `v`. Here `f` drops the
strings that are not integers:

```ant
let parsed = ["1", "x", "3"].flatMap(func(s: string): Array[int] {
  let n = s.toInt()
  if n == null { [] } else { [n] }
})
parsed // [1, 3]
```

Build every pair of two arrays. The inner lambda binds its own `x`, so the outer element goes
through a `func` parameter:

```ant
[1, 2].flatMap(func(n: int) { ["x", "y"].map({ "${n}${x}" }) }) // ["1x", "1y", "2x", "2y"]
```

Expand each order line into one entry per unit:

```ant
type Line = { sku: string, qty: int }
let lines = [Line{ sku: "A", qty: 2 }, Line{ sku: "B", qty: 1 }]
for sku in lines.flatMap({ Array.filled(x.qty, x.sku) }) {
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
let xs = [1, 2, 3]
let f: func(int): Array[int] = { [x, -x] }
assert(xs.flatMap(f) == xs.map(f).flatten())
```

The length of the result is the sum of the lengths `f` returns:

```ant
let xs = [1, 2, 3]
let f: func(int): Array[int] = { Array.filled(x, 0) }
let total = xs.map({ f(x).length() }).sum() ?? 0
assert(xs.flatMap(f).length() == total)
```

An `f` that wraps each element in a one-element array gives back a copy of `xs`, and one that wraps
`g(x)` is `Array.map(xs, g)`:

```ant
let xs = [1, 2, 3]
assert(xs.flatMap({ [x] }) == xs)
assert(!Array.same(xs.flatMap({ [x] }), xs))
assert(xs.flatMap({ [x * 2] }) == xs.map({ x * 2 }))
```

An `f` that returns `[x]` or `[]` is [Array.filter](filter.md):

```ant
let xs = [1, 2, 3, 4]
let evens: Array[int] = xs.flatMap({ if x % 2 == 0 { [x] } else { [] } })
assert(evens == xs.filter({ x % 2 == 0 }))
```

## Possible issues

`f` must return an array. A lambda that returns a single value is a compile error; use
[Array.map](map.md) for that:

```ant
let doubled = [1, 2].flatMap({ x * 2 })
// error: the lambda returns int where Array[U] is expected
```

When one branch of `f` returns `[]`, the empty array does not fix `U`, so the call is a compile
error. Annotate the result:

```ant
let odd = [1, 2, 3].flatMap({ if x % 2 == 0 { [] } else { [x] } })
// error: cannot infer the element type of an empty array; annotate it
```

```ant
let odd: Array[int] = [1, 2, 3].flatMap({ if x % 2 == 0 { [] } else { [x] } })
odd // [1, 3]
```

The result holds the elements of the arrays `f` returns, not copies of them. An element that is
itself an array or a record is the same value in both, so writing to it shows through:

```ant
let grid = [[1], [2]]
let flat = [0].flatMap({ grid })
flat[0][0] = 9
grid // [[9], [2]]
```

`Array.flatMap` flattens one level only. Arrays nested deeper stay arrays:

```ant
[[[1], [2]], [[3]]].flatMap({ x }) // [[1], [2], [3]]
```

## See also

[Array.map](map.md), [Array.flatten](flatten.md), [Array.filter](filter.md),
[Array.concat](concat.md), [Array.reduce](reduce.md), [String.graphemes](../String/graphemes.md),
[String.split](../String/split.md)
