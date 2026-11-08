<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.groupBy

```ant signature
func Array.groupBy[T, K](xs: Array[T], key: func(T): K): Map[K, Array[T]]
```

Array.groupBy returns a map from each key to the elements of xs with that key. Keys are in
first-seen order and each group keeps the order of xs.

<!-- end generated -->

## Details

- `Array.groupBy` calls `key` on the elements of `xs` and puts each element in the group of its key.
  The map's keys are in the order they first appear, and each group keeps the order of `xs`. Each
  key is the key of at least one element, so an empty `xs` gives an empty map
  ([spec: Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)).
- Keys compare the way `==` does, so a record or an array can be a key, and the map keeps insertion
  order
  ([spec: Arrays, maps, and sets](https://github.com/ant-src/spec/blob/master/ant-lang.md#arrays-maps-and-sets)).
- `m[k]` on the result reads `Array[T] | null`: `null` for a key no element has
  ([spec: Map](https://github.com/ant-src/spec/blob/master/ant-lang.md#map)).
- `Array.groupBy` reads the elements once, when it is called, leaves `xs` untouched, and returns new
  arrays
  ([spec: Standard library](https://github.com/ant-src/spec/blob/master/ant-lang.md#11-standard-library)).

## Basic examples

Group numbers by their remainder. The key `1` comes first because `1` is the first element:

```ant
Array.groupBy([1, 2, 3, 4, 5], { x % 2 }) // Map{ 1: [1, 3, 5], 0: [2, 4] }
```

Call it through `.` on the array to group words by their first letter:

```ant
var fruit = ["apple", "avocado", "banana", "blueberry", "cherry"]
fruit.groupBy({ String.slice(x, 0, 1) }) // Map{ "a": ["apple", "avocado"], "b": ["banana", "blueberry"], "c": ["cherry"] }
```

## Scope

An empty array gives an empty map. Annotate it so that `T` is known:

```ant
var empty: Array[int] = []
empty.groupBy({ x }) // Map{}
```

A named function works as well as a lambda, and so does a function of the prelude:

```ant
func parity(n: int): string {
  if n % 2 == 0 { "even" } else { "odd" }
}

Array.range(1, 7).groupBy(parity)                                    // Map{ "odd": [1, 3, 5], "even": [2, 4, 6] }
["ant", "bee", "wasp", "moth", "fly"] | Array.groupBy(String.length) // Map{ 3: ["ant", "bee", "fly"], 4: ["wasp", "moth"] }
```

Name the arguments when that reads better:

```ant
Array.groupBy(xs: [3, 1, 2], key: { x > 1 }) // Map{ true: [3, 2], false: [1] }
```

Group records by a field:

```ant
type Person = { name: string, team: string }
var people = [
  Person{ name: "Ann", team: "red" },
  Person{ name: "Bob", team: "blue" },
  Person{ name: "Cid", team: "red" },
]
people.groupBy({ x.team })["red"] // [Person{ name: "Ann", team: "red" }, Person{ name: "Cid", team: "red" }]
```

A record or an array can be the key, since keys compare by `==`:

```ant
type Cell = { row: int, col: int }
[1, 2, 3, 4].groupBy({ Cell{ row: x % 2, col: 0 } }) // Map{ Cell{ row: 1, col: 0 }: [1, 3], Cell{ row: 0, col: 0 }: [2, 4] }
[1, 2, 3, 4].groupBy({ [x % 2] })                    // Map{ [1]: [1, 3], [0]: [2, 4] }
```

`null` is a key like any other, and a nullable element can be grouped on whether it is `null`:

```ant
var xs: Array[int | null] = [1, null, 2, null]
xs.groupBy({ x })         // Map{ 1: [1], null: [null, null], 2: [2] }
xs.groupBy({ x == null }) // Map{ false: [1, 2], true: [null, null] }
```

Use `is` in the key to group the members of a union by type:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
var shapes: Array[Shape] = [Circle{ r: 1.0 }, Square{ side: 2.0 }, Circle{ r: 3.0 }]
shapes.groupBy({ if x is Circle { "circle" } else { "square" } }) // Map{ "circle": [Circle{ r: 1.0 }, Circle{ r: 3.0 }], "square": [Square{ side: 2.0 }] }
```

Group the characters of a string by splitting it into graphemes first:

```ant
String.graphemes("🐜a🐜b").groupBy({ x }) // Map{ "🐜": ["🐜", "🐜"], "a": ["a"], "b": ["b"] }
```

## Applications

Count how often each word occurs:

```ant
var words = ["the", "ant", "and", "the", "bee"]
words.groupBy({ x }).map({ len(x.value) }) // Map{ "the": 2, "ant": 1, "and": 1, "bee": 1 }
```

Print a roster by team:

```ant
type Person = { name: string, team: string }
var people = [
  Person{ name: "Ann", team: "red" },
  Person{ name: "Bob", team: "blue" },
  Person{ name: "Cid", team: "red" },
]
for entry in people.groupBy({ x.team }).entries() {
  var names = entry.value.map({ x.name })
  println("${entry.key}: ${String.join(names, ", ")}")
}
```

```text
red: Ann, Cid
blue: Bob
```

Total the amounts of each category:

```ant
type Expense = { category: string, amount: int }
var expenses = [
  Expense{ category: "food", amount: 12 },
  Expense{ category: "rent", amount: 900 },
  Expense{ category: "food", amount: 30 },
]
expenses.groupBy({ x.category }).map({ x.value.map({ x.amount }).sum() ?? 0 }) // Map{ "food": 42, "rent": 900 }
```

## Properties and relations

Flattening the groups gives every element once. The order changes when keys interleave:

```ant
var xs = [5, 3, 8, 1]
var groups = xs.groupBy({ x > 4 })
assert(Array.flatten(Map.values(groups)) == [5, 8, 3, 1])
assert(len(Array.flatten(Map.values(groups))) == len(xs))
```

The keys are the distinct keys of the elements, in first-seen order, and every element sits under
its own key:

```ant
var xs = [5, 3, 8, 1]
var key = func(n: int): bool { n > 4 }
var groups = xs.groupBy(key)
assert(Map.keys(groups) == xs.map(key).unique())
for entry in groups.entries() {
  assert(entry.value.every({ key(x) == entry.key }))
}
```

Each group is what `Array.filter` keeps for its key:

```ant
var xs = [1, 2, 3, 4, 5, 6]
var groups = xs.groupBy({ x % 3 })
for entry in groups.entries() {
  assert(entry.value == xs.filter({ x % 3 == entry.key }))
}
```

The groups are new arrays, so changing one leaves `xs` as it was:

```ant
var xs = [5, 3, 8, 1]
var big: Array[int] = xs.groupBy({ x > 4 })[true] ?? []
big.push(100)
assert(xs == [5, 3, 8, 1])
```

## Possible issues

Indexing the result reads `Array[T] | null`, because a key may have no group. Calling a function on
it without narrowing is a compile error:

<!-- test: error has no field or function length on every member -->

```ant
var groups = [1, 2, 3].groupBy({ x % 2 })
println(groups[1].length())
```

Supply an empty group with `??`, and annotate the variable so that `[]` has an element type:

```ant
var groups = [1, 2, 3].groupBy({ x % 2 })
var odd: Array[int] = groups[1] ?? []
var sevens: Array[int] = groups[7] ?? []
odd    // [1, 3]
sevens // []
```

Without the annotation, `[]` has no element type, which is a compile error:

<!-- test: error cannot infer the element type of an empty array -->

```ant
var groups = [1, 2, 3].groupBy({ x % 2 })
var odd = groups[1] ?? []
println(odd)
```

The keys are in first-seen order, not sorted. Sort them when the order matters:

```ant
var groups = [1, 2, 3, 4, 5].groupBy({ x % 2 })
Map.keys(groups)        // [1, 0]
Map.keys(groups).sort() // [0, 1]
```

`NaN` is not `==` to itself, so every `NaN` key starts a group of its own:

```ant
[1.0, Math.NaN, Math.NaN].groupBy({ x }) // Map{ 1.0: [1.0], NaN: [NaN], NaN: [NaN] }
```

A key is hashed when the map gets it. When `key` returns an element itself, such as an array,
mutating that element afterwards mutates the key, and the group may become unreachable
([spec: Arrays, maps, and sets](https://github.com/ant-src/spec/blob/master/ant-lang.md#arrays-maps-and-sets)).
Return a copy, as in `{ [...x] }`, or do not mutate the elements:

```ant
var rows = [[1], [2], [1]]
var groups = rows.groupBy({ [...x] })
rows[0].push(9)
groups[[1]] // [[1, 9], [1]]
```

## See also

[Array.uniqueBy](uniqueBy.md) · [Array.filter](filter.md) · [Array.countWhere](countWhere.md) ·
[Array.chunk](chunk.md) · [Array.sortBy](sortBy.md) · [Map.map](../Map/map.md) ·
[Map.keys](../Map/keys.md) · [Map.values](../Map/values.md)
