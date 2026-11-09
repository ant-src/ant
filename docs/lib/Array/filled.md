<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.filled

```ant signature
func Array.filled[T](n: int, v: T): Array[T]
```

Array.filled returns a new array of n elements, each the same v. Every slot holds v itself, not a
copy. It panics if n \< 0.

<!-- end generated -->

## Details

- Panics when `n` is below 0 or above 2^31 − 1.
- When `v` is an array, map, set, or record, every slot refers to that one value, so a change
  through one slot shows in all of them.

## Basic examples

Make an array of three zeros:

```ant
Array.filled(3, 0) // [0, 0, 0]
```

The result is a new array. Writing a slot changes only that slot:

```ant
let marks = Array.filled(4, "-")
marks[1] = "x"
marks // ["-", "x", "-", "-"]
```

## Scope

An `n` of 0 gives an empty array:

```ant
Array.filled(0, "a") // []
```

`v` can be any value: a string, a float, a `bool`, or a record:

```ant
type Point = { x: int, y: int }
Array.filled(2, "ab")                // ["ab", "ab"]
Array.filled(2, 1.5)                 // [1.5, 1.5]
Array.filled(3, false)               // [false, false, false]
Array.filled(2, Point{ x: 0, y: 0 }) // [Point{ x: 0, y: 0 }, Point{ x: 0, y: 0 }]
```

To fill with `null` and store values later, annotate the element type. The annotation makes the
array an `Array[int | null]`:

```ant
let slots: Array[int | null] = Array.filled(3, null)
slots[0] = 5
slots // [5, null, null]
```

The count comes first, so a pipe passes the count:

```ant
3 | Array.filled("z") // ["z", "z", "z"]
```

The parameters are named `n` and `v`, and named arguments may come in any order:

```ant
Array.filled(n: 2, v: "x") // ["x", "x"]
Array.filled(v: "x", n: 2) // ["x", "x"]
```

## Applications

Count how often each digit occurs, with one counter per digit:

```ant
let counts = Array.filled(5, 0)
for d in [1, 3, 3, 4, 1, 1] {
  counts[d] += 1
}
counts // [0, 3, 0, 2, 1]
```

Pad an array of digits on the left to a fixed width:

```ant
func pad(xs: Array[int], width: int): Array[int] {
  [...Array.filled(width - len(xs), 0), ...xs]
}
pad([7, 8], 5) // [0, 0, 0, 7, 8]
```

Build a grid with a new row per line, then draw it:

```ant
let grid = Array.range(0, 3).map({ Array.filled(4, ".") })
grid[1][2] = "#"
for row in grid {
  println(String.join(row, ""))
}
```

```text
....
..#.
....
```

Give each goroutine its own slot to write a result into, then read the slots after the wait:

```ant
let words = ["ant", "bee", "wasp"]
let lengths: Array[int | null] = Array.filled(len(words), null)
let wg = WaitGroup.new()
for entry in Array.entries(words) {
  WaitGroup.add(wg, 1)
  go {
    defer { WaitGroup.done(wg) }
    lengths[entry.index] = String.length(entry.value)
  }
}
WaitGroup.wait(wg)
println(lengths)
```

```text
[3, 3, 4]
```

`Parallel.map` does this for a whole array.

## Properties and relations

The result has `n` elements, and every one is `v`:

```ant
let xs = Array.filled(4, "a")
assert(len(xs) == 4)
assert(Array.every(xs, { x == "a" }))
```

It equals mapping every index of `Array.range(0, n)` to `v`:

```ant
assert(Array.filled(3, 1) == Array.range(0, 3).map({ 1 }))
```

Joining `n` copies of a string gives the same string as `String.repeat`:

```ant
assert(String.join(Array.filled(3, "ab"), "") == String.repeat("ab", 3))
```

## Possible issues

A negative count panics. Code that computes the count, such as `pad` above, panics when the
computation goes below 0:

```ant
Array.filled(-1, 0) // panic: Array.filled needs a count of at least 0, got -1
```

A count above 2^31 − 1 panics too:

```ant
Array.filled(3000000000, 0) // panic: Array.filled needs a count of at most 2147483647, got 3000000000
```

Every slot holds the same `v`, not a copy. When `v` is an array or a record, a change through one
slot shows in every slot:

```ant
let rows = Array.filled(2, [0, 0])
rows[0][0] = 1
rows                         // [[1, 0], [1, 0]]
Array.same(rows[0], rows[1]) // true
```

Build a new value per slot with `Array.map` over `Array.range` instead:

```ant
let rows = Array.range(0, 2).map({ [0, 0] })
rows[0][0] = 1
rows // [[1, 0], [0, 0]]
```

Without an annotation, `Array.filled(n, null)` is an `Array[null]`, which holds nothing but `null`:

```ant
let slots = Array.filled(3, null)
slots[0] = 5 // error: cannot use int where null is expected
```

An empty array literal as `v` has no element type to infer, so annotate it:

```ant
let rows = Array.filled(2, [])
// error: cannot infer the element type of an empty array; annotate it
```

## See also

[Array.range](range.md), [Array.map](map.md), [Array.same](same.md),
[String.repeat](../String/repeat.md), [Array.push](push.md), [Parallel.map](../Parallel/map.md)
