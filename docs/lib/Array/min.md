<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.min

```ant signature
func Array.min[T: Ordered](xs: Array[T]): T | null
```

Array.min returns the least element of xs, or null when xs is empty. It returns NaN when any element
is NaN, and holds -0.0 below 0.0.

<!-- end generated -->

## Details

- Returns `null` when `xs` is empty, so give it a default with `??` before using it as a `T`.
- Returns `NaN` when any element is `NaN`, and holds `-0.0` less than `0.0`.
- Leaves `xs` unchanged.

## Basic examples

```ant
Array.min([3, 1, 2])                // 1
Array.min(["pear", "apple", "fig"]) // "apple"
```

Call it through `.` on any array:

```ant
let temps = [12.5, -3.0, 7.25]
temps.min() // -3.0
```

An empty array gives `null`. Annotate it so that `T` is known:

```ant
let empty: Array[int] = []
Array.min(empty) // null
```

## Scope

Every numeric type works, the sized ones included:

```ant
let levels: Array[byte] = [byte(200), byte(7), byte(64)]
levels.min()               // 7
[-5, 3].min()              // -5
[Math.Infinity, 2.0].min() // 2.0
```

Strings compare by code point, so every upper-case ASCII letter comes before every lower-case one,
and a precomposed accented letter such as `é` (U+00E9) comes after `z`:

```ant
["b", "B", "a"].min() // "B"
["é", "z", "e"].min() // "e"
["zé", "zebra"].min() // "zebra"
```

A decomposed `é` is an `e` followed by a combining accent, so it starts with U+0065 and comes before
`z`:

```ant
["z", "e\u{301}"].min() == "e\u{301}" // true
```

A `NaN` anywhere makes the result `NaN`, wherever it is:

```ant
[1.5, Math.NaN, -2.0].min() // NaN
[Math.NaN, 1.0].min()       // NaN
```

`-0.0` counts as less than `0.0`, in either order:

```ant
[0.0, -0.0].min() // -0.0
[-0.0, 0.0].min() // -0.0
```

The values of a map or a set are an array, so `Array.min` gives the least of them:

```ant
let ages = Map{ ann: 31, bob: 27, cy: 45 }
Map.values(ages).min()           // 27
Set.of([4, 2, 9]).values().min() // 2
```

In a pipe, the array is the first argument:

```ant
"5 12 3 8" | String.split(" ") | Array.map({ String.toInt(x) ?? 0 }) | Array.min() // 3
```

## Applications

Find the length of the shortest word:

```ant
["kiwi", "fig", "banana"].map(String.length).min() // 3
```

ISO 8601 dates sort as strings, so the least string is the earliest date:

```ant
["2026-03-01", "2025-12-31", "2026-01-15"].min() // "2025-12-31"
```

Report the lowest score, or say there is none:

```ant
func lowest(scores: Array[int]): string {
  let low = scores.min()
  if low == null {
    return "no scores yet"
  }
  "lowest: ${low}, ${100 - low} below full marks"
}

println(lowest([88, 72, 95]))
println(lowest([]))
```

```text
lowest: 72, 28 below full marks
no scores yet
```

Shift every value so that the least becomes zero:

```ant
let readings = [14, 9, 21, 12]
let base = readings.min() ?? 0
readings.map({ x - base }) // [5, 0, 12, 3]
```

## Properties and relations

The result is an element of the array, and no element is below it:

```ant
for xs in [[5, 3, 8], [7], [2, 2, 9]] {
  let low = xs.min()
  assert(low != null)
  assert(xs.contains(low))
  assert(xs.every({ low <= x }))
}
```

Without `NaN`, it agrees with the first element of [`Array.sort`](sort.md), and with
[`Math.min`](../Math/min.md) folded over the array by [`Array.reduce`](reduce.md):

```ant
for xs in [[4, 1, 3], [9], [6, 6, 2]] {
  assert(xs.min() == xs.sort().first())
  assert(xs.min() == xs.reduce(Math.maxInt, Math.min))
}
```

The result is `null` exactly when the array is empty:

```ant
for xs in [["a"], []] {
  assert((xs.min() == null) == xs.isEmpty())
}
```

`Array.min` does not change the array:

```ant
let xs = [3, 1, 2]
assert(xs.min() == 1)
assert(xs == [3, 1, 2])
```

## Possible issues

An empty array literal gives `T` nothing to infer from, so it is a compile error:

```ant
println(Array.min([])) // error: cannot infer the element type of an empty array; annotate it
```

Annotate the array so that `T` is known:

```ant
let empty: Array[int] = []
Array.min(empty) // null
```

The result is `T | null`, not `T`, so using it as a `T` is a compile error:

```ant
let xs = [3, 1]
println(xs.min() + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??`, or narrow a local:

```ant
let xs = [3, 1]
(xs.min() ?? 0) + 1 // 2
```

An array of records has no order, so it is a compile error. To find the record with the least field,
use [`Array.minBy`](minBy.md):

```ant
type Item = { name: string, price: int }
let items = [Item{ name: "pen", price: 3 }, Item{ name: "ink", price: 2 }]
println(items.min())
// error: T would be Item, which is not a member of its bound Ordered
```

```ant
type Item = { name: string, price: int }
let items = [Item{ name: "pen", price: 3 }, Item{ name: "ink", price: 2 }]
items.minBy({ x.price }) // Item{ name: "ink", price: 2 }
```

A named primitive type is not assignable to its base, so it does not fit `Ordered`:

```ant
type Meters = float64
let legs = [Meters(3.0), Meters(1.5)]
println(legs.min())
// error: T would be Meters, which is not a member of its bound Ordered; unwrap it with float64(...)
```

Unwrap the elements with the base type's name first:

```ant
type Meters = float64
let legs = [Meters(3.0), Meters(1.5)]
legs.map(float64).min() // 1.5
```

An element type that includes `null` is a union, which is not a single member of `Ordered`:

```ant
let cells: Array[int | null] = [4, null, 2]
println(cells.min())
// error: T would be int | null, which is not a single member of its bound Ordered; narrow first
```

Map each `null` to `Math.maxInt` to get an `Array[int]`:

```ant
let cells: Array[int | null] = [4, null, 2]
cells.map({ x ?? Math.maxInt }).min() // 2
```

`Math.maxInt` is only a stand-in: it wins when every element is `null`, and the result is then
`Math.maxInt` instead of `null`:

```ant
let cells: Array[int | null] = [null, null]
cells.map({ x ?? Math.maxInt }).min() // 9223372036854775807
```

A single `NaN` hides every other value. Drop the `NaN` elements first when they mean "no reading":

```ant
let samples = [2.5, Math.NaN, 0.5]
samples.min()                            // NaN
samples.filter({ !Math.isNaN(x) }).min() // 0.5
```

[`Array.minBy`](minBy.md) with the element itself as the key is not the same function. It keeps the
first of equal keys, so it does not hold `-0.0` below `0.0`, and a `NaN` never replaces the element
it holds:

```ant
[0.0, -0.0].minBy({ x })     // 0.0
[0.0, -0.0].min()            // -0.0
[1.0, Math.NaN].minBy({ x }) // 1.0
[1.0, Math.NaN].min()        // NaN
```

## See also

[Array.max](max.md) · [Array.minBy](minBy.md) · [Array.sort](sort.md) · [Array.sum](sum.md) ·
[Array.reduce](reduce.md) · [Math.min](../Math/min.md) ·
[Comparison operators](../../language/operators/comparison.md) ·
[?? operator](../../language/operators/null-coalescing.md) ·
[Generics guide](../../guide/generics.md)
