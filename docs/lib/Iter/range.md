<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter.range

```ant signature
func Iter.range(from: int, to: int): Seq[int]
```

Iter.range returns the sequence of integers from from, inclusive, up to to, exclusive. It is empty
when to \<= from.

<!-- end generated -->

## Details

- It never counts down: when `to` is at or below `from`, the sequence is empty.
- It is lazy: it computes nothing until a `for` loop or [`Iter.toArray`](toArray.md) runs it, and it
  stops as soon as the loop breaks. Each run starts again at `from`.
- Unlike [`Array.range`](../Array/range.md), it has no size limit, since it never holds the integers
  in memory.

## Basic examples

Loop over the integers from 0 up to, but not including, 3:

```ant
for i in Iter.range(0, 3) {
  println(i)
}
```

```text
0
1
2
```

Collect the sequence into an array with [`Iter.toArray`](toArray.md):

```ant
Iter.range(2, 5) | Iter.toArray() // [2, 3, 4]
```

## Scope

Equal bounds give an empty sequence, and so do bounds in the wrong order:

```ant
Iter.range(3, 3) | Iter.toArray() // []
Iter.range(5, 2) | Iter.toArray() // []
```

Negative bounds work like any others:

```ant
Iter.range(-3, 1) | Iter.toArray() // [-3, -2, -1, 0]
```

The bounds can be anywhere in the `int` range, up to and including `Math.minInt` and `Math.maxInt`:

```ant
Iter.range(Math.maxInt - 2, Math.maxInt) | Iter.toArray()
// [9223372036854775805, 9223372036854775806]
Iter.range(Math.minInt, Math.maxInt) | Iter.take(2) | Iter.toArray()
// [-9223372036854775808, -9223372036854775807]
```

A pipe passes `from`, and `.` chains other `Iter` functions on the result:

```ant
1 | Iter.range(4) | Iter.toArray()                                // [1, 2, 3]
Iter.range(0, 10).filter({ x % 3 == 0 }).map({ x * x }).toArray() // [0, 9, 36, 81]
```

## Applications

Loop over the indices of an array; an empty array gives no iterations:

```ant
let names = ["Ann", "Bob", "Cy"]
for i in Iter.range(0, len(names)) {
  println("${i + 1}. ${names[i]}")
}
```

```text
1. Ann
2. Bob
3. Cy
```

Sum the integers from 1 to 100:

```ant
var total = 0
for i in Iter.range(1, 101) {
  total += i
}
total // 5050
```

Search without a known end: the range is huge, but the loop returns at the first match, so only a
few integers are ever produced:

```ant
func firstSquareOver(n: int): int {
  for i in Iter.range(0, Math.maxInt) {
    if i * i > n {
      return i
    }
  }
  -1
}
firstSquareOver(50)   // 8
firstSquareOver(1000) // 32
```

Take the first five multiples of 7 from a range of a thousand integers; the pipeline stops after the
fifth:

```ant
Iter.range(0, 1000) | Iter.filter({ x % 7 == 0 }) | Iter.take(5) | Iter.toArray()
// [0, 7, 14, 21, 28]
```

## Properties and relations

Collected, `Iter.range` gives the same integers as [`Array.range`](../Array/range.md):

```ant
assert(Iter.range(-2, 4).toArray() == Array.range(-2, 4))
assert(Iter.range(4, -2).toArray() == Array.range(4, -2))
```

The number of integers is `to - from` when that is positive, and 0 otherwise:

```ant
assert(len(Iter.range(3, 10).toArray()) == 10 - 3)
assert(len(Iter.range(10, 3).toArray()) == 0)
```

A range can be run more than once, and each run gives the same integers:

```ant
let r = Iter.range(0, 3)
assert(r.toArray() == r.toArray())
r.toArray() // [0, 1, 2]
```

## Possible issues

The work happens only when the sequence runs. A callback in [`Iter.map`](map.md) runs once per
integer the pipeline pulls, and not at all when the sequence is built:

```ant
let s = Iter.range(0, 100).map(func(x: int): int {
  println("visit ${x}")
  x * 10
})
println("built")
println(s.take(2).toArray())
```

```text
built
visit 0
visit 1
[0, 10]
```

A sequence is a function, not a collection. Printing it does not show its integers, and `len` and
indexing do not accept it; collect it with [`Iter.toArray`](toArray.md) first:

```ant
println(Iter.range(0, 3))
println(Iter.range(0, 3) | Iter.toArray())
```

```text
func
[0, 1, 2]
```

```ant
let r = Iter.range(0, 3)
let n = len(r)
// error: len needs an array, a map, a set, a string, Bytes, or a channel, got Seq[int]
```

The bounds are `int`; a float is a compile error, even one with no fraction:

```ant
let xs = Iter.range(0, 2.0) | Iter.toArray() // error: cannot use float64 where int is expected
```

To count down, map each integer to its distance from the top; the sequence stays lazy:

```ant
Iter.range(0, 3).map({ 3 - x }).toArray() // [3, 2, 1]
```

Reversing the collected array also works, but it builds the whole array first:

```ant
Iter.range(1, 4).toArray() | Array.reverse() // [3, 2, 1]
```

## See also

[Array.range](../Array/range.md) · [Iter.toArray](toArray.md) · [Iter.take](take.md) ·
[Iter.filter](filter.md) · [Iter.map](map.md) · [Iter.fromArray](fromArray.md) · [Seq](types/Seq.md)
· [for](../../language/keywords/for.md)
