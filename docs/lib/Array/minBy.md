<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.minBy

```ant signature
func Array.minBy[T, K: Ordered](xs: Array[T], key: func(T): K): T | null
```

Array.minBy returns the first element of xs with the least key, or null when xs is empty. An element
replaces the one held only when its key is strictly less, so of equal keys the first wins and a NaN
key never replaces the element held.

<!-- end generated -->

## Details

- Returns the element, not its key, and `null` when `xs` is empty.
- Calls `key` once per element, from first to last, and never stops early.
- Of equal keys, the first element wins, and a `NaN` key never wins.

## Basic examples

```ant
Array.minBy(["bb", "a", "c"], { String.length(x) }) // "a"
Array.minBy([3, -5, 2], { Math.abs(x) })            // 2
```

Call it through `.` on any array:

```ant
let words = ["hill", "ant", "colony"]
words.minBy({ String.length(x) }) // "ant"
```

## Scope

An empty array has no element to return. Annotate it so that `T` is known:

```ant
let empty: Array[string] = []
empty.minBy({ String.length(x) }) // null
```

When several elements share the least key, the first of them wins:

```ant
["bee", "ant", "fly"].minBy({ String.length(x) }) // "bee"
```

Pick a record by one of its fields. The whole record comes back, and `?.` reads a field of it:

```ant
type Product = { name: string, price: float64 }
let products = [
  Product{ name: "pen", price: 1.5 },
  Product{ name: "ink", price: 0.9 },
  Product{ name: "pad", price: 2.25 },
]
products.minBy({ x.price })       // Product{ name: "ink", price: 0.9 }
products.minBy({ x.price })?.name // "ink"
```

The key can be a `string`:

```ant
type User = { name: string, age: int }
let users = [User{ name: "Bob", age: 27 }, User{ name: "Ann", age: 31 }]
users.minBy({ x.name }) // User{ name: "Ann", age: 31 }
```

The elements themselves need no order. Here they are union values, and the key gives each one a
number:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square

func area(s: Shape): float64 {
  if s is Circle {
    3.14159 * s.r * s.r
  } else {
    s.side * s.side
  }
}

let shapes: Array[Shape] = [Square{ side: 2.0 }, Circle{ r: 1.0 }, Square{ side: 2.5 }]
shapes.minBy(area) // Circle{ r: 1.0 }
```

Pass a named function, or an anonymous `func` with a typed parameter:

```ant
func distance(p: int): int {
  Math.abs(p - 10)
}

[4, 13, 7].minBy(distance)                    // 13
[4, 13, 7].minBy(func(n: int): int { n % 5 }) // 7
```

A lambda captures the variables around it:

```ant
let target = 20
[12, 25, 17].minBy({ Math.abs(x - target) }) // 17
```

`key` runs once for each element, in order, even after the least key has been seen. A key with a
side effect shows it, and shows that an empty array never calls it:

```ant
let seen: Array[int] = []
[0, 4, 2].minBy({ seen.push(x); x }) // 0
seen                                 // [0, 4, 2]

let empty: Array[int] = []
empty.minBy({ seen.push(x); x }) // null
seen                             // [0, 4, 2]
```

In a pipeline, the array is the first argument:

```ant
"pear,fig,banana" | String.split(",") | Array.minBy({ String.length(x) }) // "fig"
```

## Applications

Find the cheapest offer and say where it comes from:

```ant
type Offer = { shop: string, price: int }
let offers = [Offer{ shop: "north", price: 120 }, Offer{ shop: "south", price: 95 }]
let best = offers.minBy({ x.price })
if best != null {
  println("buy at ${best.shop} for ${best.price}")
}
```

```text
buy at south for 95
```

Snap a reading to the nearest allowed value:

```ant
let steps = [0.0, 0.25, 0.5, 0.75, 1.0]
let reading = 0.61
println(steps.minBy({ Math.abs(x - reading) }) ?? 0.0)
```

```text
0.5
```

Pick the shortest route, with a clear message when there is none:

```ant
type Route = { via: string, km: float64 }
let routes = [Route{ via: "coast", km: 41.5 }, Route{ via: "pass", km: 33.0 }]
let route = routes.minBy({ x.km }) ?? panic("no route")
println("go via ${route.via}")
```

```text
go via pass
```

## Properties and relations

The key of the element found equals, by `==`, what [`Array.min`](min.md) of the keys gives when no
key is `NaN`. With `{ x }` as the key, `Array.minBy` agrees with `Array.min`:

```ant
let xs = ["kiwi", "fig", "banana", "date"]
let found = xs.minBy({ String.length(x) }) ?? panic("empty")
assert(String.length(found) == xs.map({ String.length(x) }).min())
let ns = [5, 3, 8, 3]
assert(ns.minBy({ x }) == ns.min())
```

When no key is `NaN`, the first element of [`Array.sortBy`](sortBy.md), which is stable and
ascending, is the one `Array.minBy` finds. Negating a float key, or a signed integer key above the
type's minimum, gives what [`Array.maxBy`](maxBy.md) finds:

```ant
type Job = { name: string, cost: int }
let jobs = [Job{ name: "a", cost: 4 }, Job{ name: "b", cost: 2 }, Job{ name: "c", cost: 2 }]
assert(jobs.minBy({ x.cost }) == jobs.sortBy({ x.cost }).first())
assert(jobs.minBy({ -x.cost }) == jobs.maxBy({ x.cost }))
assert(jobs.minBy({ x.cost })?.name == "b")
```

## Possible issues

The result can be `null`, and the type checker rejects it as a `T` until you narrow it:

```ant
println(["a", "bb"].minBy({ String.length(x) }).length())
// error: string | null has no field or function length on every member: narrow the value first
```

The key must be one ordered type. A record, a `bool`, or a nullable key is a compile error; key on a
field, or give `null` a number with `??`:

```ant
let scores: Array[int | null] = [70, null, 55]
println(scores.minBy({ x }))
// error: K would be int | null, which is not a single member of its bound Ordered; narrow first
```

```ant
let scores: Array[int | null] = [70, null, 55]
scores.minBy({ x ?? 100 }) // 55
```

A `NaN` key never replaces the element held, so it never wins against a number. But nothing is less
than `NaN` either, so a `NaN` key on the first element is never replaced:

```ant
[2.0, Math.NaN, 0.5].minBy({ x }) // 0.5
[Math.NaN, 2.0, 0.5].minBy({ x }) // NaN
```

Filter out `NaN` keys first when they can occur:

```ant
[Math.NaN, 2.0, 0.5].filter({ !Math.isNaN(x) }).minBy({ x }) // 0.5
```

`Array.minBy` compares with `<`, and `-0.0 < 0.0` is `false`, so between the two zeros the first
wins. [`Array.min`](min.md) holds `-0.0` below `0.0` instead:

```ant
[0.0, -0.0].minBy({ x }) // 0.0
[0.0, -0.0].min()        // -0.0
```

## See also

[Array.maxBy](maxBy.md), [Array.min](min.md), [Array.max](max.md), [Array.sortBy](sortBy.md),
[Array.find](find.md), [Array.reduce](reduce.md), [Math.min](../Math/min.md)
