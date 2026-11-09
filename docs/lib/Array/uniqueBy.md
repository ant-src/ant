<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.uniqueBy

```ant signature
func Array.uniqueBy[T, K](xs: Array[T], key: func(T): K): Array[T]
```

Array.uniqueBy returns a new array holding the first element of xs for each == key, in order.

<!-- end generated -->

## Details

- Keeps the first element for each key, in the order of `xs`.
- Calls `key` once per element, from first to last, and never stops early.
- Keys compare with `==`: a `NaN` key never matches, and `-0.0` matches `0.0`.
- Returns a new array; `xs` is unchanged.

## Basic examples

Keep the first word for each first letter:

```ant
var fruit = ["apple", "avocado", "banana", "cherry", "blueberry"]
Array.uniqueBy(fruit, { String.slice(x, 0, 1) }) // ["apple", "banana", "cherry"]
```

Keep the first number of each parity, through `.`:

```ant
[4, 7, 2, 9, 1].uniqueBy({ x % 2 }) // [4, 7]
```

## Scope

An empty array gives an empty array. Annotate it, as Possible issues shows:

```ant
var empty: Array[int] = []
empty.uniqueBy({ x }) // []
```

A key that never repeats keeps every element, and a constant key keeps only the first:

```ant
[3, 1, 2].uniqueBy({ x })     // [3, 1, 2]
[3, 1, 2].uniqueBy({ "all" }) // [3]
```

A named function works as well as a lambda, and so do an anonymous function and a function of the
prelude:

```ant
func decade(year: int): int { year / 10 }

[1994, 1998, 2003, 2007, 2011].uniqueBy(decade)               // [1994, 2003, 2011]
[1, 2, 3].uniqueBy(func(n: int): bool { n > 1 })              // [1, 2]
["ant", "bee", "wasp", "fly"] | Array.uniqueBy(String.length) // ["ant", "wasp"]
```

Name the arguments when that reads better:

```ant
Array.uniqueBy(xs: [1, 2, 3, 4], key: { x > 2 }) // [1, 3]
```

Keep the first record for each value of a field:

```ant
type Order = { id: int, customer: string }
var orders = [
  Order{ id: 1, customer: "Ann" },
  Order{ id: 2, customer: "Bob" },
  Order{ id: 3, customer: "Ann" },
]
orders.uniqueBy({ x.customer }) // [Order{ id: 1, customer: "Ann" }, Order{ id: 2, customer: "Bob" }]
```

A record or an array can be the key, since keys compare by `==`:

```ant
type Cell = { row: int, col: int }
type Hit = { row: int, col: int, damage: int }
var hits = [
  Hit{ row: 0, col: 1, damage: 5 },
  Hit{ row: 2, col: 2, damage: 3 },
  Hit{ row: 0, col: 1, damage: 9 },
]
hits.uniqueBy({ Cell{ row: x.row, col: x.col } }).map({ x.damage }) // [5, 3]
hits.uniqueBy({ [x.row, x.col] }).map({ x.damage })                 // [5, 3]
```

`null` is a key like any other:

```ant
var xs: Array[int | null] = [1, null, 2, null, 1]
xs.uniqueBy({ x })         // [1, null, 2]
xs.uniqueBy({ x == null }) // [1, null]
```

Tell the members of a union apart with `is`:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
var shapes: Array[Shape] = [Circle{ r: 1.0 }, Circle{ r: 3.0 }, Square{ side: 2.0 }]
shapes.uniqueBy({ x is Circle }) // [Circle{ r: 1.0 }, Square{ side: 2.0 }]
```

Work on the characters of a string by splitting it into graphemes first:

```ant
String.graphemes("Ant ANT ant").uniqueBy({ String.toLower(x) }) // ["A", "n", "t", " "]
```

## Applications

Drop repeated email addresses, ignoring case:

```ant
var emails = ["ann@example.com", "Bob@example.com", "ANN@example.com", "bob@example.com"]
emails.uniqueBy(String.toLower) // ["ann@example.com", "Bob@example.com"]
```

Keep the latest reading of each sensor. Reverse to keep the last element for each key instead of the
first, then reverse back:

```ant
type Reading = { sensor: string, value: int }
var readings = [
  Reading{ sensor: "a", value: 1 },
  Reading{ sensor: "b", value: 2 },
  Reading{ sensor: "a", value: 3 },
]
readings.reverse().uniqueBy({ x.sensor }).reverse() // [Reading{ sensor: "b", value: 2 }, Reading{ sensor: "a", value: 3 }]
```

Pick the cheapest offer per product by sorting first, since the first element per key wins:

```ant
type Offer = { product: string, price: int, shop: string }
var offers = [
  Offer{ product: "pen", price: 3, shop: "A" },
  Offer{ product: "ink", price: 9, shop: "A" },
  Offer{ product: "pen", price: 2, shop: "B" },
]
for offer in offers.sortBy({ x.price }).uniqueBy({ x.product }) {
  println("${offer.product}: ${offer.price} at ${offer.shop}")
}
```

```text
pen: 2 at B
ink: 9 at A
```

## Properties and relations

With the identity as the key, `Array.uniqueBy` is `Array.unique`:

```ant
var xs = [3, 1, 3, 2, 1]
assert(xs.uniqueBy({ x }) == xs.unique())
```

The keys of the result are the distinct keys of `xs`, in first-seen order, and each kept element is
the first with its key:

```ant
var xs = ["ant", "bee", "wasp", "fly", "moth"]
var kept = xs.uniqueBy(String.length)
assert(kept.map(String.length) == xs.map(String.length).unique())
for k in kept {
  assert(xs.find({ String.length(x) == String.length(k) }) == k)
}
```

Each kept element is the first of its group in `Array.groupBy`:

```ant
var xs = [5, 3, 8, 1, 6]
var firsts = Map.values(xs.groupBy({ x % 3 })).map({ x[0] })
assert(xs.uniqueBy({ x % 3 }) == firsts)
```

`key` runs once per element, in the order of `xs`, so a side effect in it sees every element,
including the ones that are dropped:

```ant
var seen: Array[int] = []
[3, 1, 4, 1].uniqueBy({ seen.push(x); x % 2 }) // [3, 4]
seen                                           // [3, 1, 4, 1]
```

A `key` that pushes to `xs` does not add to what it visits:

```ant
var xs = [1, 2, 1]
xs.uniqueBy({ xs.push(x + 10); x }) // [1, 2]
xs                                  // [1, 2, 1, 11, 12, 11]
```

`Array.uniqueBy` returns a new array no longer than `xs`, and leaves `xs` as it was:

```ant
var xs = [1, 2, 3, 4]
var kept = xs.uniqueBy({ x % 2 })
kept.push(100)
assert(xs == [1, 2, 3, 4])
assert(len(xs.uniqueBy({ x % 2 })) <= len(xs))
```

## Possible issues

Strings compare by code points, so two spellings of the same character are two keys:

```ant
["\u{e9}", "e\u{301}"].uniqueBy({ x }) // ["é", "é"]
```

`NaN` is not `==` to itself, so every `NaN` key is new and keeps its element:

```ant
[1.0, Math.NaN, Math.NaN].uniqueBy({ x }) // [1.0, NaN, NaN]
```

`-0.0` and `0.0` are `==`, so they share a key:

```ant
[-0.0, 0.0].uniqueBy({ x }) // [-0.0]
```

An empty literal with nothing to fix its element type is a compile error, even before the call. The
checker stops at the first line:

<!-- test: error cannot infer the element type of an empty array -->

```ant
var e = []
e.uniqueBy({ x })
```

Annotate the array:

```ant
var e: Array[int] = []
e.uniqueBy({ x }) // []
```

The first element for each key is kept, not the last. Reverse before and after to keep the last, as
in Applications.

## See also

[Array.unique](unique.md) · [Array.groupBy](groupBy.md) · [Array.sortBy](sortBy.md) ·
[Array.filter](filter.md) · [Array.reverse](reverse.md) · [Set.of](../Set/of.md)
