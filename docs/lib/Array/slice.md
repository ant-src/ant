<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.slice

```ant signature
func Array.slice[T](xs: Array[T], from: int, to: int): Array[T]
```

Array.slice returns a new array of the elements of xs from index from up to but excluding to. It
panics if from \> to, or if either is outside 0 ..= length.

<!-- end generated -->

## Details

- The range is half-open: the result starts at `from` and stops before `to`.
- Panics if `from > to`, or if either is below 0 or above the length. There are no negative indices
  counting from the end.
- Returns a new array; `xs` is unchanged.

## Basic examples

Take the elements at indices 1 and 2. Index 3, the `to`, is left out:

```ant
Array.slice([10, 20, 30, 40], 1, 3) // [20, 30]
```

Take the first two elements, or everything after the first:

```ant
let xs = ["a", "b", "c", "d"]
Array.slice(xs, 0, 2)       // ["a", "b"]
Array.slice(xs, 1, len(xs)) // ["b", "c", "d"]
```

## Scope

An empty range gives `[]`, wherever it sits, the end included:

```ant
let xs = [10, 20, 30]
Array.slice(xs, 0, 0) // []
Array.slice(xs, 2, 2) // []
Array.slice(xs, 3, 3) // []
```

The range `0` to the length covers the whole array:

```ant
Array.slice([10, 20, 30], 0, 3) // [10, 20, 30]
```

An empty array has one range, `0` to `0`:

```ant
let empty: Array[int] = []
Array.slice(empty, 0, 0) // []
```

Records and `null` elements are sliced like any other value:

```ant
type Point = { x: int, y: int }
let path = [Point{ x: 0, y: 0 }, Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]
Array.slice(path, 1, 2)               // [Point{ x: 1, y: 2 }]
Array.slice([1, null, 3, null], 1, 3) // [null, 3]
```

Slice the characters a reader sees by splitting a string into graphemes first:

```ant
Array.slice(String.graphemes("🐜🐝🐞🦋"), 1, 3) // ["🐝", "🐞"]
```

Call it through `.` or in a pipe:

```ant
[1, 2, 3, 4].slice(1, 3)         // [2, 3]
[1, 2, 3, 4] | Array.slice(2, 4) // [3, 4]
```

## Named arguments

Name `from` and `to` to make a call read on its own, in either order:

```ant
Array.slice([1, 2, 3, 4], from: 1, to: 3) // [2, 3]
Array.slice([1, 2, 3, 4], to: 2, from: 0) // [1, 2]
```

## Applications

Take at most `n` elements. [Math.min](../Math/min.md) keeps `to` within the length:

```ant
func take[T](xs: Array[T], n: int): Array[T] {
  Array.slice(xs, 0, Math.min(n, len(xs)))
}

println(take([1, 2, 3, 4, 5], 3))
println(take([1, 2], 3))
```

```text
[1, 2, 3]
[1, 2]
```

Drop a header row and keep the data rows:

```ant
let lines = ["name,age", "Ann,31", "Bob,27"]
for line in Array.slice(lines, 1, len(lines)) {
  println(line)
}
```

```text
Ann,31
Bob,27
```

Show one page of results, clamping the last page to the length:

```ant
func page(items: Array[string], number: int, size: int): Array[string] {
  let from = Math.min(number * size, len(items))
  let to = Math.min(from + size, len(items))
  Array.slice(items, from, to)
}

let names = ["Ann", "Bob", "Cid", "Dee", "Eve"]
println(page(names, 0, 2))
println(page(names, 2, 2))
println(page(names, 3, 2))
```

```text
["Ann", "Bob"]
["Eve"]
[]
```

Sum each window of three readings:

```ant
let readings = [3, 6, 9, 6, 3]
for i in Array.range(0, len(readings) - 2) {
  let window = Array.slice(readings, i, i + 3)
  println(Array.sum(window) ?? 0)
}
```

```text
18
21
18
```

## Properties and relations

The result holds `to - from` elements, and element `i` of it is element `from + i` of `xs`:

```ant
let xs = [5, 6, 7, 8, 9]
let from = 1
let to = 4
let part = Array.slice(xs, from, to)
assert(len(part) == to - from)
assert(Array.every(Array.range(0, len(part)), { part[x] == xs[from + x] }))
```

Two slices that meet at the same index put the array back together:

```ant
let xs = [1, 2, 3, 4, 5]
let i = 2
assert(Array.concat(Array.slice(xs, 0, i), Array.slice(xs, i, len(xs))) == xs)
```

Slicing a slice is one slice with the offsets added:

```ant
let xs = [1, 2, 3, 4, 5, 6]
assert(Array.slice(Array.slice(xs, 1, 5), 1, 3) == Array.slice(xs, 2, 4))
```

Destructuring with a rest binds everything after the head as a new array, the same as slicing from
`1`. Unlike the slice, it panics on an empty array:

```ant
let xs = [1, 2, 3]
let [_, ...tail] = xs
assert(tail == Array.slice(xs, 1, len(xs)))
```

A slice of the whole array is a copy: equal, but a new array, so writing to it leaves `xs` as it
was:

```ant
let xs = [1, 2, 3]
let copy = Array.slice(xs, 0, len(xs))
assert(copy == xs)
assert(!Array.same(copy, xs))
copy[0] = 9
xs // [1, 2, 3]
```

## Possible issues

`to` is excluded, so a range that names the last index leaves the last element out:

```ant
let xs = [10, 20, 30]
Array.slice(xs, 0, 2)       // [10, 20]
Array.slice(xs, 0, len(xs)) // [10, 20, 30]
```

`to` past the length panics, unlike in languages that clamp it. Clamp it yourself, as `take` above
does:

```ant
Array.slice([1, 2, 3], 0, 5) // panic: Array.slice: the range 0..5 is not within 0..3
```

A negative index panics. It does not count from the end:

```ant
Array.slice([1, 2, 3], -1, 3) // panic: Array.slice: the range -1..3 is not within 0..3
```

Write the length out to take from the end:

```ant
let xs = [1, 2, 3]
Array.slice(xs, len(xs) - 1, len(xs)) // [3]
```

`from` after `to` panics rather than giving `[]`:

```ant
Array.slice([1, 2, 3], 2, 1) // panic: Array.slice: the range 2..1 is not within 0..3
```

The new array holds the same elements, not copies of them. Writing into an element that is itself an
array or a record shows through the original:

```ant
let grid = [[1, 2], [3, 4]]
let top = Array.slice(grid, 0, 1)
top[0][0] = 99
grid // [[99, 2], [3, 4]]
```

## See also

[Array.get](get.md), [Array.chunk](chunk.md), [Array.concat](concat.md), [Array.range](range.md),
[Array.removeAt](removeAt.md), [String.slice](../String/slice.md), [Bytes.slice](../Bytes/slice.md)
