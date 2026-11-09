<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.min

```ant signature
func Math.min[T: Ordered](a: T, b: T): T
```

Math.min returns the smaller of a and b.

<!-- end generated -->

## Details

- Strings compare by code point, so every uppercase ASCII letter comes before every lowercase one.
- If either argument is `NaN`, the result is `NaN`.
- `-0.0` counts as smaller than `0.0`, in either order.

## Basic examples

```ant
Math.min(3, 7)      // 3
Math.min(2.5, -1.0) // -1.0
Math.min(-4, -4)    // -4
```

## Scope

Every integer and float type works, and the result keeps its type:

```ant
Math.min(uint8(200), uint8(10))      // 10
Math.min(int8(-128), int8(127))      // -128
Math.min(float32(1.5), float32(0.5)) // 0.5
Math.min(Math.minInt, Math.maxInt)   // -9223372036854775808
```

Strings compare by code point:

```ant
Math.min("pear", "apple")  // "apple"
Math.min("Zebra", "apple") // "Zebra"
Math.min("", "a")          // ""
Math.min("é", "z")         // "z"
```

Special float values:

```ant
Math.min(-Math.Infinity, 1.0) // -Infinity
Math.min(Math.Infinity, 5.0)  // 5.0
Math.min(Math.NaN, 1.0)       // NaN
Math.min(1.0, Math.NaN)       // NaN
1.0 / Math.min(0.0, -0.0)     // -Infinity
1.0 / Math.min(-0.0, 0.0)     // -Infinity
```

Pass `Math.min` as a function value:

```ant
[18.5, 21.0, 16.25, 19.0].reduce(Math.Infinity, Math.min) // 16.25
```

## Named arguments

```ant
Math.min(a: 4, b: 2) // 2
Math.min(b: 2, a: 4) // 2
```

## Applications

Cap every value at a limit:

```ant
let requested = [5, 12, 30]
requested.map({ Math.min(x, 10) }) // [5, 10, 10]
```

Take at most the first three elements, without a panic on a shorter array:

```ant
func firstThree(xs: Array[string]): Array[string] {
  xs.slice(0, Math.min(3, xs.length()))
}

firstThree(["a", "b", "c", "d"]) // ["a", "b", "c"]
firstThree(["a"])                // ["a"]
```

The smallest of three values, with one call nested in another:

```ant
let a = 8
let b = 3
let c = 5
Math.min(a, Math.min(b, c)) // 3
```

For more values, put them in an array and use [Array.min](../Array/min.md). It returns `T | null`,
with `null` for an empty array:

```ant
let a = 8
let b = 3
let c = 5
let d = 6
[a, b, c, d].min()      // 3
[a, b, c, d].min() ?? 0 // 3
```

The overlap of two time windows, with [Math.max](max.md):

```ant
type Window = { start: int, end: int }

func overlap(a: Window, b: Window): int {
  Math.max(0, Math.min(a.end, b.end) - Math.max(a.start, b.start))
}

overlap(Window{ start: 0, end: 10 }, Window{ start: 6, end: 15 }) // 4
overlap(Window{ start: 0, end: 5 }, Window{ start: 6, end: 15 })  // 0
```

## Properties and relations

`Math.min` returns one of its arguments, is symmetric, and is never larger than either:

```ant
for [a, b] in [[3, 7], [-2, -9], [5, 5]] {
  assert(Math.min(a, b) == a || Math.min(a, b) == b)
  assert(Math.min(a, b) == Math.min(b, a))
  assert(Math.min(a, b) <= a && Math.min(a, b) <= b)
}
```

It is [Math.max](max.md) with the order reversed, and equals [Array.min](../Array/min.md) on two
elements:

```ant
let a = 4
let b = 11
assert(Math.min(a, b) == -Math.max(-a, -b))
assert(Math.min(a, b) == [a, b].min())
```

## Possible issues

Both arguments need the same type. An `int` and a `float64` are a compile error; convert one:

```ant
let count = 3
let ratio = 2.5
let n = Math.min(count, ratio) // error: cannot use float64 where int is expected
```

```ant
let count = 3
let ratio = 2.5
Math.min(float64(count), ratio) // 2.5
```

A value typed `Number` or a union must be narrowed first:

```ant
let a: Number = 3
let b: Number = 4
let n = Math.min(a, b)
// error: T would be Number, which is not a single member of its bound Ordered; narrow first
```

A nullable value is a union with `null`. Give it a fallback with `??`:

```ant
let limit: int | null = null
let n = Math.min(limit, 10)
// error: T would be int | null, which is not a single member of its bound Ordered; narrow first
```

```ant
let limit: int | null = null
Math.min(limit ?? 10, 10) // 10
```

A named primitive type is not `Ordered`. Unwrap it, then tag the result again:

```ant
type Meters = float64
let n = Math.min(Meters(2.0), Meters(3.0))
// error: T would be Meters, which is not a single member of its bound Ordered; narrow first
```

```ant
type Meters = float64
let a = Meters(2.0)
let b = Meters(3.0)
Meters(Math.min(float64(a), float64(b))) // Meters(2.0)
```

One `NaN` makes the result `NaN`, so a running minimum over data with a `NaN` ends as `NaN`:

```ant
[1.0, Math.NaN, 0.5].reduce(Math.Infinity, Math.min) // NaN
```

## See also

[Math.max](max.md) · [Math.clamp](clamp.md) · [Array.min](../Array/min.md) ·
[Array.minBy](../Array/minBy.md) · [Math namespace](index.md)
