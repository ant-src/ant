<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.range

```ant signature
func Array.range(from: int, to: int): Array[int]
```

Array.range returns the integers from from up to but excluding to. It returns an empty array when to
\<= from.

<!-- end generated -->

## Details

- The range is half-open: `from` is included and `to` is not.
- When `to` is at or below `from`, the result is empty; it never counts down.
- Panics when `to − from` is above 2^31 − 1.
- Builds the whole array at once. [`Iter.range`](../Iter/range.md) yields the same integers lazily.

## Basic examples

Count from 2 up to, but not including, 5:

```ant
Array.range(2, 5) // [2, 3, 4]
```

Start at 0 to get the indices of an array of a given length:

```ant
Array.range(0, 4) // [0, 1, 2, 3]
```

## Scope

Equal bounds give an empty array, and so do bounds in the wrong order:

```ant
Array.range(3, 3) // []
Array.range(5, 2) // []
```

Negative bounds work like any others:

```ant
Array.range(-3, 2) // [-3, -2, -1, 0, 1]
```

The bounds can sit anywhere in the `int` range, as long as the range itself is small:

```ant
Array.range(Math.maxInt - 2, Math.maxInt) // [9223372036854775805, 9223372036854775806]
```

Call it in a pipe, which passes `from`, or chain other functions on the result through `.`:

```ant
0 | Array.range(3)                       // [0, 1, 2]
Array.range(0, 5).map({ x * x })         // [0, 1, 4, 9, 16]
Array.range(0, 5).filter({ x % 2 == 0 }) // [0, 2, 4]
```

Name the bounds `from` and `to`, in any order:

```ant
Array.range(from: 1, to: 4) // [1, 2, 3]
Array.range(to: 4, from: 1) // [1, 2, 3]
```

## Applications

Visit every index of an array:

```ant
var animals = ["ant", "bee", "wasp"]
for i in Array.range(0, len(animals)) {
  println("${i}: ${animals[i]}")
}
```

```text
0: ant
1: bee
2: wasp
```

Sum the integers from 1 to 100. The upper bound is 101, because `to` is excluded:

```ant
Array.sum(Array.range(1, 101)) // 5050
```

Take the difference between each pair of neighbors, starting at index 1:

```ant
var xs = [3, 1, 4, 1, 5]
Array.range(1, len(xs)).map({ xs[x] - xs[x - 1] }) // [-2, 3, -3, 4]
```

Number lines from 1 instead of 0:

```ant
func numbered(lines: Array[string]): Array[string] {
  Array.range(1, len(lines) + 1).map({ "${x}. ${lines[x - 1]}" })
}

for line in numbered(["dig", "carry", "rest"]) {
  println(line)
}
```

```text
1. dig
2. carry
3. rest
```

Build a grid with a new row per index, so the rows are separate arrays:

```ant
var grid = Array.range(0, 3).map({ [0, 0, 0] })
grid[1][1] = 5
grid // [[0, 0, 0], [0, 5, 0], [0, 0, 0]]
```

## Properties and relations

The result has `to - from` elements when `to` is above `from`, and none otherwise:

```ant
assert(len(Array.range(-5, 5)) == 10)
assert(len(Array.range(5, -5)) == 0)
```

Two adjacent ranges concatenate to one range:

```ant
assert(Array.concat(Array.range(0, 3), Array.range(3, 7)) == Array.range(0, 7))
```

`Array.range` gives the same integers as `Iter.range` collected into an array:

```ant
assert(Iter.toArray(Iter.range(2, 6)) == Array.range(2, 6))
```

Slicing an array with the same bounds picks the elements at the indices of the range:

```ant
var xs = ["a", "b", "c", "d", "e"]
assert(Array.slice(xs, 1, 4) == Array.range(1, 4).map({ xs[x] }))
```

## Possible issues

`to` is excluded. To include the last value, add 1 to it:

```ant
Array.range(1, 5) // [1, 2, 3, 4]
Array.range(1, 6) // [1, 2, 3, 4, 5]
```

`Array.range` never counts down. Reverse an ascending range instead:

```ant
Array.range(5, 0)                // []
Array.reverse(Array.range(1, 6)) // [5, 4, 3, 2, 1]
```

`Array.range` panics when `to − from` is above 2^31 − 1:

```ant
Array.range(0, 3000000000) // panic: Array.range needs a range of at most 2147483647 elements, got 3000000000
```

To loop over a range that large, or to stop early, use the lazy `Iter.range`, which never builds the
array:

```ant
Iter.range(0, 3000000000).filter({ x % 7 == 0 }).take(3).toArray() // [0, 7, 14]
```

Both bounds must be `int`s:

```ant
var xs = Array.range(0, 2.5) // error: cannot use float64 where int is expected
```

## See also

[Iter.range](../Iter/range.md), [Array.filled](filled.md), [Array.entries](entries.md),
[Array.slice](slice.md), [Array.map](map.md), [Array.reverse](reverse.md)
