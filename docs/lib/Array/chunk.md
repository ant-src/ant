<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.chunk

```ant signature
func Array.chunk[T](xs: Array[T], n: int): Array[Array[T]]
```

Array.chunk splits xs into arrays of n elements each, the last holding fewer when the length is not
a multiple of n. It panics if n \< 1. An empty xs gives \[\].

<!-- end generated -->

## Details

- `n` below 1 panics, even when `xs` is empty.

## Basic examples

Split five numbers into pairs. The last chunk holds the one left over:

```ant
Array.chunk([1, 2, 3, 4, 5], 2) // [[1, 2], [3, 4], [5]]
```

When the length is a multiple of `n`, every chunk is full:

```ant
Array.chunk([1, 2, 3, 4, 5, 6], 3) // [[1, 2, 3], [4, 5, 6]]
```

## Scope

An `n` of 1 puts each element in its own chunk:

```ant
["a", "b", "c"].chunk(1) // [["a"], ["b"], ["c"]]
```

An `n` at or above the length gives one chunk with every element:

```ant
[1, 2].chunk(2) // [[1, 2]]
[1, 2].chunk(5) // [[1, 2]]
```

An empty array has no chunks:

```ant
let empty: Array[int] = []
empty.chunk(3) // []
```

Chunk a string into groups of three graphemes by splitting it with `String.graphemes` first:

```ant
"🐜🐝🐞🦋🐛".graphemes().chunk(3) // [["🐜", "🐝", "🐞"], ["🦋", "🐛"]]
```

Records and `null` elements are chunked like any other value:

```ant
type Point = { x: int, y: int }
let points = [Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }, Point{ x: 5, y: 6 }]
points.chunk(2)[1]    // [Point{ x: 5, y: 6 }]
[1, null, 3].chunk(2) // [[1, null], [3]]
```

Call it through `.` or in a pipe:

```ant
[1, 2, 3, 4, 5].chunk(2)         // [[1, 2], [3, 4], [5]]
[1, 2, 3, 4, 5] | Array.chunk(3) // [[1, 2, 3], [4, 5]]
```

## Applications

Sum each group of three readings:

```ant
let readings = [4, 8, 6, 1, 2, 3, 10]
readings.chunk(3).map({ x.sum() ?? 0 }) // [18, 6, 10]
```

Turn a flat list of key and value pairs into records:

```ant
type Setting = { key: string, value: string }
let flat = ["host", "localhost", "port", "8080"]
let settings = flat.chunk(2).map({ Setting{ key: x[0], value: x[1] } })
settings[1] // Setting{ key: "port", value: "8080" }
```

Print a list of names in rows of three:

```ant
let names = ["Ann", "Bob", "Cid", "Dee", "Eve", "Fay", "Gus"]
for row in names.chunk(3) {
  println(String.join(row, "  "))
}
```

```text
Ann  Bob  Cid
Dee  Eve  Fay
Gus
```

Split work into pages of a fixed size and number them:

```ant
func pages(items: Array[string], size: int): Array[string] {
  items
    .chunk(size)
    .entries()
    .map({ "page ${x.index + 1}: ${String.join(x.value, ", ")}" })
}

for line in pages(["a", "b", "c", "d", "e"], 2) {
  println(line)
}
```

```text
page 1: a, b
page 2: c, d
page 3: e
```

## Properties and relations

Flattening the chunks gives back the original array:

```ant
let xs = [1, 2, 3, 4, 5, 6, 7]
assert(xs.chunk(3).flatten() == xs)
```

There are `(length + n - 1) / n` chunks, and every chunk but the last holds `n` elements:

```ant
let xs = [1, 2, 3, 4, 5, 6, 7]
let n = 3
let chunks = xs.chunk(n)
assert(len(chunks) == (len(xs) + n - 1) / n)
assert(chunks.slice(0, len(chunks) - 1).every({ len(x) == n }))
assert(len(chunks[len(chunks) - 1]) <= n)
```

The chunks are new arrays, so changing one leaves `xs` as it was:

```ant
let xs = [1, 2, 3]
let c = xs.chunk(2)
c[0][0] = 9
assert(xs == [1, 2, 3])
```

## Possible issues

An `n` of 0 or below panics, even when the array is empty:

```ant
[1, 2, 3].chunk(0) // panic: Array.chunk needs a size of at least 1, got 0
```

```ant
let empty: Array[int] = []
empty.chunk(-1) // panic: Array.chunk needs a size of at least 1, got -1
```

The last chunk can be shorter than `n`. Code that reads a fixed position in every chunk must check
the length first:

```ant
let flat = ["host", "localhost", "port"]
flat.chunk(2).map({ x[1] })                         // panic: index 1 out of range for an array of length 1
flat.chunk(2).filter({ len(x) == 2 }).map({ x[1] }) // ["localhost"]
```

```ant
let empty: Array[int] = []
empty.chunk(3)      // []
len(empty.chunk(3)) // 0
```

An empty array gives `[]`, not `[[]]`, so a loop over its chunks runs zero times.

## See also

[Array.slice](slice.md), [Array.flatten](flatten.md), [Array.groupBy](groupBy.md),
[Array.entries](entries.md), [String.graphemes](../String/graphemes.md)
