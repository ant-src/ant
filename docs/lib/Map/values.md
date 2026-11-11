<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.values

```ant signature
func Map.values[K, V](m: Map[K, V]): Array[V]
```

Map.values returns the values of m in insertion order.

<!-- end generated -->

## Details

- The values come in the order their keys were first inserted. Replacing a value keeps its place; a
  key removed and inserted again moves its value to the end.
- Returns a new array: pushing to it leaves `m` unchanged. Each element is the stored value itself,
  so changing a record or an array in it changes what `m` holds.

## Basic examples

```ant
let ages = Map{ ada: 36, linus: 54, grace: 85 }
Map.values(ages) // [36, 54, 85]
```

Call it through `.` on the map:

```ant
let colors = Map{ red: "#f00", green: "#0f0" }
colors.values() // ["#f00", "#0f0"]
```

## Scope

An empty map gives an empty array:

```ant
let empty: Map[string, int] = Map{}
empty.values() // []
```

Two keys with the same value give that value twice:

```ant
Map{ a: 1, b: 1, c: 2 }.values() // [1, 1, 2]
```

Replacing a value keeps its place. Removing a key and setting it again moves its value to the end:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
m["a"] = 10
m.values()    // [10, 2, 3]
m.remove("a") // 10
m["a"] = 1
m.values() // [2, 3, 1]
```

When `V` includes `null`, stored `null` values come through as elements:

```ant
let scores: Map[string, int | null] = Map{ ann: 31, bob: null }
scores.values() // [31, null]
```

Values can be records, arrays, or members of a union:

```ant
type Point = { x: int, y: int }
Map{ home: Point{ x: 0, y: 0 } }.values()   // [Point{ x: 0, y: 0 }]
Map{ evens: [2, 4], odds: [1, 3] }.values() // [[2, 4], [1, 3]]
let mixed: Map[string, int | string] = Map{ a: 1, b: "two" }
mixed.values() // [1, "two"]
```

In a pipeline, the map is the first argument:

```ant
Map{ b: 1, a: 2 } | Map.values() // [1, 2]
```

Name the argument when that reads better:

```ant
Map.values(m: Map{ x: 1 }) // [1]
```

## Applications

`for` over a map yields entries; loop over `values()` when only the values matter:

```ant
let stock = Map{ apple: 3, pear: 0 }
for n in stock.values() {
  println(n)
}
```

```text
3
0
```

Total the values of a map:

```ant
let cart = Map{ apple: 3, pear: 2, plum: 5 }
cart.values().sum() // 10
```

Find the largest value:

```ant
let temps = Map{ mon: 18.5, tue: 21.0, wed: 19.5 }
temps.values().max() // 21.0
```

Count words, then list them from most to least frequent. The map finds each word's record; the
values carry both the word and its count:

```ant
type Count = { word: string, n: int }
let counts: Map[string, Count] = Map{}
for w in ["the", "ant", "the", "hill", "the", "ant"] {
  let c = counts[w] ?? Count{ word: w, n: 0 }
  c.n += 1
  counts[w] = c
}
counts.values().sortBy({ -x.n }).map({ "${x.word}: ${x.n}" })
// ["the: 3", "ant: 2", "hill: 1"]
```

List the distinct values in the order they first appear:

```ant
let owners = Map{ "a.txt": "ann", "b.txt": "bob", "c.txt": "ann" }
owners.values().unique() // ["ann", "bob"]
```

## Properties and relations

The values are the values of `Map.entries`, one per key, and they line up with `Map.keys`:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
let ks = m.keys()
let vs = m.values()
assert(vs == m.entries().map({ x.value }))
assert(vs.length() == m.size())
for i in Array.range(0, ks.length()) {
  assert(m[ks[i]] == vs[i])
}
```

`Map.map` changes the values and keeps their order:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
assert(m.map({ x.value * 10 }).values() == m.values().map({ x * 10 }))
```

The array is new. Pushing to it leaves the map alone, and inserting into the map leaves the array
alone:

```ant
let m = Map{ a: 1 }
let vs = m.values()
vs.push(9)
m["b"] = 2
assert(vs == [1, 9])
assert(m.values() == [1, 2])
```

## Possible issues

The array is new, but its elements are the values the map holds. Changing a record through the array
changes the map:

```ant
type Item = { name: string, qty: int }
let stock = Map{ a: Item{ name: "apple", qty: 1 } }
let items = stock.values()
items[0].qty = 5
stock // Map{ "a": Item{ name: "apple", qty: 5 } }
```

Two maps with the same entries are `==` in any order, but their values are arrays, and arrays
compare in order:

```ant
let m = Map{ a: 1, b: 2 }
let n = Map{ b: 2, a: 1 }
m == n                                 // true
m.values() == n.values()               // false
m.values().sort() == n.values().sort() // true
```

The values do not say which key they came from. Use `Map.entries` when you need both:

```ant
let votes = Map{ tea: 3, coffee: 7 }
votes.values().max()                    // 7
votes.entries().maxBy({ x.value })?.key // "coffee"
```

`Map.values` is a function, not a field. Without the parentheses it is a compile error:

```ant
let m = Map{ a: 1 }
println(m.values) // error: Map[string, int] is not a record, so it has no field values
```

## See also

[Map.keys](keys.md) · [Map.entries](entries.md) · [Map.map](map.md) · [Map.get](get.md) ·
[Set.values](../Set/values.md) · [Array.sum](../Array/sum.md) · [Array.unique](../Array/unique.md)
