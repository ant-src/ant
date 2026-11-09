<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.max

```ant signature
func Array.max[T: Ordered](xs: Array[T]): T | null
```

Array.max returns the greatest element of xs, or null when xs is empty. It returns NaN when any
element is NaN.

<!-- end generated -->

## Details

- Returns `null` when `xs` is empty, so give it a default with `??` before using it as a `T`.
- Returns `NaN` when any element is `NaN`, and holds `-0.0` less than `0.0`.
- Leaves `xs` unchanged.

## Basic examples

```ant
Array.max([3, 1, 4, 1, 5]) // 5
Array.max([2.5, -1.0])     // 2.5
```

An empty array gives `null`. Annotate it so that `T` is known:

```ant
var empty: Array[int] = []
Array.max(empty) // null
```

Call it through `.` on any array:

```ant
var temps = [18.5, 21.0, 19.25]
temps.max() // 21.0
```

## Scope

Strings work as well as numbers:

```ant
["pear", "apple", "fig"].max() // "pear"
```

Every numeric type works, the sized ones included:

```ant
[byte(3), byte(200)].max()          // 200
[float32(1.5), float32(-2.0)].max() // 1.5
[int8(-5), int8(-100)].max()        // -5
```

Infinities compare as IEEE 754 orders them:

```ant
[Math.Infinity, 1.0e300].max() // Infinity
[-Math.Infinity, -1.0].max()   // -1.0
```

A single element is the greatest:

```ant
[42].max() // 42
```

The values of a map, and the members of a set, are arrays, so `Array.max` takes them:

```ant
Map{ a: 3, b: 7, c: 5 }.values().max() // 7
Set.of(["b", "c", "a"]).values().max() // "c"
```

In a pipe, the array is the first argument:

```ant
"3 14 1 59" | String.split(" ") | Array.max() // "59"
```

## Applications

Give an empty array a default with `??`:

```ant
func highScore(scores: Array[int]): int {
  scores.max() ?? 0
}

println(highScore([70, 92, 85]))
println(highScore([]))
```

```text
92
0
```

Find the greatest field of a list of records by mapping to the field first:

```ant
type Order = { id: string, total: float64 }
var orders = [Order{ id: "a", total: 12.5 }, Order{ id: "b", total: 40.0 }]
orders.map({ x.total }).max() // 40.0
```

Find the latest date among ISO 8601 strings, which sort as text in date order:

```ant
["2026-03-01", "2025-12-31", "2026-11-08"].max() // "2026-11-08"
```

Scale values to the range 0 to 1 by dividing by the greatest, narrowing a local first:

```ant
func normalize(xs: Array[float64]): Array[float64] {
  var top = xs.max()
  if top == null || top == 0.0 {
    return xs
  }
  xs.map({ x / top })
}

println(normalize([1.0, 2.0, 4.0]))
```

```text
[0.25, 0.5, 1.0]
```

## Properties and relations

`Array.max` agrees with the last element of the sorted array, and with a fold through `Math.max`:

```ant
var xs = [3, -7, 4, 1]
assert(xs.max() == xs.sort().last())
assert(xs.max() == xs.reduce(xs[0], Math.max))
```

The result is at least every element, and is one of them:

```ant
var xs = ["kiwi", "apple", "plum"]
var top = xs.max() ?? ""
assert(xs.every({ x <= top }))
assert(xs.contains(top))
```

The order of the elements does not change the result, and `max` does not change the array:

```ant
var xs = [5, 2, 9, 2]
assert(xs.max() == xs.reverse().max())
assert(xs == [5, 2, 9, 2])
```

`Array.max` is at least `Array.min`:

```ant
var xs = [1.5, -2.0, 0.25]
var hi = xs.max() ?? 0.0
var lo = xs.min() ?? 0.0
assert(lo <= hi)
```

## Possible issues

The result is `T | null`, not `T`, so using it as a `T` is a compile error:

```ant
var xs = [3, 1, 2]
println(xs.max() + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??`, or narrow a local:

```ant
var xs = [3, 1, 2]
(xs.max() ?? 0) + 1 // 4
```

One `NaN` makes the result `NaN`, wherever it is. Drop it first with `Math.isNaN`:

```ant
var xs = [1.0, Math.NaN, 3.0]
xs.max()                            // NaN
xs.filter({ !Math.isNaN(x) }).max() // 3.0
```

`-0.0` and `0.0` are `==`, but `Array.max` holds `-0.0` less than `0.0`, so the result does not
depend on their order:

```ant
[-0.0, 0.0].max() // 0.0
[0.0, -0.0].max() // 0.0
[-0.0].max()      // -0.0
```

Mixing integer and float literals infers a union element type, `int | float64`, which is not a
single member of `Ordered`, so it is a compile error:

```ant
println([1, 2.5].max())
// error: T would be int | float64, which is not a single member of its bound Ordered; narrow first
```

Write every element as a float, or annotate the array:

```ant
var xs: Array[float64] = [1.0, 2.5]
[1.0, 2.5].max() // 2.5
xs.max()         // 2.5
```

An array whose element type is a union, such as `int | null` or `Number`, has no single `T`, so it
is a compile error:

```ant
var xs: Array[int | null] = [1, null, 3]
println(xs.max())
// error: T would be int | null, which is not a single member of its bound Ordered; narrow first
```

Collect the non-`null` elements into an `Array[int]` first. Testing each one narrows it:

```ant
var xs: Array[int | null] = [1, null, 3]
var ns: Array[int] = []
for v in xs {
  if v != null {
    ns.push(v)
  }
}
ns.max() // 3
```

Because `T` never includes `null`, a `null` result always means an empty array.

Records, `bool`, and named primitive types are not `Ordered`:

```ant
type Meters = float64
var runs = [Meters(5.0), Meters(10.0)]
println(runs.max())
// error: T would be Meters, which is not a single member of its bound Ordered; narrow first
```

Unwrap a named primitive with its base type, or pick a record by a key with
[`Array.maxBy`](maxBy.md):

```ant
type Meters = float64
type Run = { who: string, distance: float64 }
var runs = [Meters(5.0), Meters(10.0)]
var log = [Run{ who: "Ann", distance: 5.0 }, Run{ who: "Bob", distance: 10.0 }]
runs.map({ float64(x) }).max() // 10.0
log.maxBy({ x.distance })?.who // "Bob"
```

## See also

[Array.min](min.md) · [Array.maxBy](maxBy.md) · [Array.minBy](minBy.md) · [Array.sort](sort.md) ·
[Array.sum](sum.md) · [Array.reduce](reduce.md) · [Math.max](../Math/max.md) ·
[comparison operators](../../language/operators/comparison.md) ·
[?? operator](../../language/operators/null-coalescing.md)
