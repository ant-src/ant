<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.entries

```ant signature
func Array.entries[T](xs: Array[T]): Array[Indexed[T]]
```

Array.entries returns each element of xs paired with its index, in order.

<!-- end generated -->

## Details

- Indices start at 0, so an empty `xs` gives `[]`.
- Returns a new array, so later pushes or removals on `xs` do not change it. Each `value` is the
  element itself, not a copy.

## Basic examples

Pair each element with its index:

```ant
Array.entries(["a", "b", "c"]) // [Indexed{ index: 0, value: "a" }, Indexed{ index: 1, value: "b" }, Indexed{ index: 2, value: "c" }]
```

Read the fields of one entry:

```ant
var es = Array.entries([10, 20, 30])
es[2].index // 2
es[2].value // 30
```

Loop over the entries and destructure each one:

```ant
for { index, value } in ["ant", "bee"].entries() {
  println("${index}: ${value}")
}
```

```text
0: ant
1: bee
```

## Scope

An empty array has no entries. Annotate it so that `T` is known:

```ant
var empty: Array[int] = []
Array.entries(empty) // []
```

`null` elements get an index like any other:

```ant
Array.entries(["x", null]) // [Indexed{ index: 0, value: "x" }, Indexed{ index: 1, value: null }]
```

Records, nested arrays, and union members are carried as they are:

```ant
type Point = { x: int, y: int }
Array.entries([Point{ x: 1, y: 2 }]) // [Indexed{ index: 0, value: Point{ x: 1, y: 2 } }]
Array.entries([[1], []])             // [Indexed{ index: 0, value: [1] }, Indexed{ index: 1, value: [] }]
var r: Array[Result[int, string]] = [Result.ok(1), Result.err("no")]
Array.entries(r)[1] // Indexed{ index: 1, value: Err{ error: "no" } }
```

A string is not an array. Split it into graphemes first:

```ant
String.graphemes("🐜🐝").entries() // [Indexed{ index: 0, value: "🐜" }, Indexed{ index: 1, value: "🐝" }]
```

Call it through `.` or in a pipe:

```ant
[1, 2].entries()         // [Indexed{ index: 0, value: 1 }, Indexed{ index: 1, value: 2 }]
[1, 2] | Array.entries() // [Indexed{ index: 0, value: 1 }, Indexed{ index: 1, value: 2 }]
```

Rename the fields as you destructure them:

```ant
for { index: rank, value: name } in ["Ann", "Bob"].entries() {
  println("#${rank + 1} ${name}")
}
```

```text
#1 Ann
#2 Bob
```

It works in a generic function, where the element type is a type parameter:

```ant
func everyOther[T](items: Array[T]): Array[T] {
  items.entries().filter({ x.index % 2 == 0 }).map({ x.value })
}

println(everyOther(["a", "b", "c", "d", "e"]))
println(everyOther([1, 2]))
```

```text
["a", "c", "e"]
[1]
```

## Applications

Find the positions of the elements a test accepts:

```ant
var temps = [3, 8, 2, 9, 7]
temps.entries().filter({ x.value > 5 }).map({ x.index }) // [1, 3, 4]
```

Weight each element by its position:

```ant
[4, 5, 6].entries().map({ x.index * x.value }) // [0, 5, 12]
```

Find the first position of the largest element, together with its value (`Array.maxBy` keeps the
first of equal keys):

```ant
var scores = [70, 95, 88, 95]
Array.maxBy(scores.entries(), { x.value }) // Indexed{ index: 1, value: 95 }
```

Build a map from each name to its position:

```ant
var names = ["Ann", "Bob", "Cid"]
Map.fromEntries(names.entries().map({ Entry{ key: x.value, value: x.index } })) // Map{ "Ann": 0, "Bob": 1, "Cid": 2 }
```

Compare each element with the one before it:

```ant
var prices = [10, 12, 11, 15]
for { index, value } in prices.entries() {
  if index > 0 && value > prices[index - 1] {
    println("day ${index}: up ${value - prices[index - 1]}")
  }
}
```

```text
day 1: up 2
day 3: up 4
```

## Properties and relations

There is one entry per element, the values are `xs`, and the indices are `Array.range(0, len(xs))`:

```ant
var xs = ["p", "q", "r"]
var es = Array.entries(xs)
assert(len(es) == len(xs))
assert(es.map({ x.value }) == xs)
assert(es.map({ x.index }) == Array.range(0, len(xs)))
assert(es.every({ xs[x.index] == x.value }))
```

The result is a new array, so pushing to `xs` afterwards leaves it as it was:

```ant
var xs = [1, 2, 3]
var es = xs.entries()
xs.push(4)
assert(len(es) == 3)
```

An entry is an ordinary record and compares by `==` with one you write:

```ant
assert(["a"].entries()[0] == Indexed{ index: 0, value: "a" })
```

## Possible issues

The fields are `index` and `value`. A map's entries are `Entry` records with `key` and `value`, so a
destructuring written for a map does not fit:

```ant
for { key, value } in ["a", "b"].entries() { // error: Indexed[string] has no field key
  println("${key}: ${value}")
}
```

Indices start at 0. Add 1 when you number lines for a reader:

```ant
["first", "second"].entries().map({ "${x.index + 1}. ${x.value}" }) // ["1. first", "2. second"]
```

An entry holds the same record as `xs`, so a change made through `xs` shows in the entry:

```ant
type P = { x: int }
var xs = [P{ x: 1 }]
var es = xs.entries()
xs[0].x = 9
es[0].value.x // 9
```

Copy the element first, for example with a record spread, if the entry must stay unchanged.

The other way round, assigning an entry's `value` does not change `xs`:

```ant
var xs = [5]
var es = xs.entries()
es[0].value = 7
xs // [5]
```

An entry is a new record, so assigning its `value` replaces the field in the entry only. To change
the array, write `xs[e.index] = ...`.

`Array.entries` takes an array, not a string:

```ant
println(Array.entries("abc")) // error: cannot use string where Array[T] is expected
```

## See also

[Indexed](types/Indexed.md) · [Array.range](range.md) · [Array.map](map.md) ·
[Array.findIndex](findIndex.md) · [Array.indexOf](indexOf.md) · [Map.entries](../Map/entries.md) ·
[Map.fromEntries](../Map/fromEntries.md) · [Iter.fromArray](../Iter/fromArray.md)
