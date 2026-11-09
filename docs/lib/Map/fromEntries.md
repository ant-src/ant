<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.fromEntries

```ant signature
func Map.fromEntries[K, V](entries: Array[Entry[K, V]]): Map[K, V]
```

Map.fromEntries returns a new map holding entries, in order. When a key repeats, the later entry's
value wins.

<!-- end generated -->

## Details

- Returns a new map; `entries` is unchanged.
- When a key repeats, the map keeps the key at its first position and holds the value of its last
  entry.
- An empty array gives an empty map.

## Basic examples

Build a map from an array of `Entry` records:

```ant
Map.fromEntries([Entry{ key: "ann", value: 31 }, Entry{ key: "bob", value: 27 }]) // Map{ "ann": 31, "bob": 27 }
```

A repeated key takes the value of its last entry and keeps its first position:

```ant
let votes = [Entry{ key: "red", value: 1 }, Entry{ key: "blue", value: 2 }, Entry{ key: "red", value: 3 }]
Map.fromEntries(votes) // Map{ "red": 3, "blue": 2 }
```

## Scope

An empty array gives an empty map. Annotate the array so its element type is known:

```ant
let none: Array[Entry[string, int]] = []
Map.fromEntries(none) // Map{}
```

Keys of any type work, such as integers or records. Keys compare with `==`:

```ant
type Cell = { row: int, col: int }
Map.fromEntries([Entry{ key: 404, value: "Not Found" }])                              // Map{ 404: "Not Found" }
Map.fromEntries([Entry{ key: Cell{ row: 0, col: 1 }, value: "x" }])                   // Map{ Cell{ row: 0, col: 1 }: "x" }
Map.fromEntries([Entry{ key: [1, 2], value: "a" }, Entry{ key: [1, 2], value: "b" }]) // Map{ [1, 2]: "b" }
```

A value may be `null` when the value type allows it:

```ant
let settings: Map[string, int | null] = Map.fromEntries([Entry{ key: "limit", value: null }])
settings                   // Map{ "limit": null }
Map.has(settings, "limit") // true
```

Build the entries with `Array.map`:

```ant
Map.fromEntries(["ant", "beetle"].map({ Entry{ key: x, value: String.length(x) } })) // Map{ "ant": 3, "beetle": 6 }
```

Call it in a pipe:

```ant
[Entry{ key: "a", value: 1 }] | Map.fromEntries() // Map{ "a": 1 }
```

Name the argument:

```ant
Map.fromEntries(entries: [Entry{ key: "a", value: 1 }]) // Map{ "a": 1 }
```

## Applications

Index records by a field:

```ant
type User = { id: int, name: string }
let users = [User{ id: 7, name: "Ann" }, User{ id: 9, name: "Bob" }]
let byId = Map.fromEntries(users.map({ Entry{ key: x.id, value: x } }))
byId[9] // User{ id: 9, name: "Bob" }
```

Swap the keys and values of a map:

```ant
let codes = Map{ red: "#f00", green: "#0f0" }
Map.fromEntries(codes.entries().map({ Entry{ key: x.value, value: x.key } })) // Map{ "#f00": "red", "#0f0": "green" }
```

Sort a map by value. `Map.entries` gives an array to sort, and the new map keeps the sorted order:

```ant
let prices = Map{ pen: 2, ink: 15, pad: 4 }
prices.entries().sortBy({ x.value }) | Map.fromEntries() // Map{ "pen": 2, "pad": 4, "ink": 15 }
```

Parse `key=value` settings:

```ant
func parsePair(s: string): Entry[string, string] {
  let parts = String.split(s, "=")
  Entry{ key: parts[0], value: parts[1] }
}
String.split("host=localhost;port=8080", ";").map(parsePair) | Map.fromEntries() // Map{ "host": "localhost", "port": "8080" }
```

Map each word to its last index with `Array.entries`:

```ant
let words = ["to", "be", "or", "not", "to", "be"]
Map.fromEntries(Array.entries(words).map({ Entry{ key: x.value, value: x.index } })) // Map{ "to": 4, "be": 5, "or": 2, "not": 3 }
```

## Properties and relations

`Map.fromEntries` and `Map.entries` round-trip. The result is equal to the source map but is a new
map:

```ant
let m = Map{ a: 1, b: 2 }
let copy = Map.fromEntries(Map.entries(m))
assert(copy == m)
assert(!Map.same(copy, m))
```

The map has at most one entry per element, and its keys are the distinct keys in the order they
first appear:

```ant
let es = [Entry{ key: "a", value: 1 }, Entry{ key: "b", value: 2 }, Entry{ key: "a", value: 3 }]
let m = Map.fromEntries(es)
assert(Map.size(m) <= len(es))
assert(Map.keys(m) == Array.unique(es.map({ x.key })))
```

## Possible issues

A literal for each element names its type, `Entry`. A bare `{ key: ..., value: ... }` does not
compile:

```ant
let m = Map.fromEntries([{ key: "a", value: 1 }])
// error: a record or map literal names its type, as in Point{ x: 1 } or Map{ "a": 1 }
```

An empty array literal gives no element type to infer:

```ant
let m = Map.fromEntries([]) // error: cannot infer the element type of an empty array; annotate it
```

`fromEntries` belongs to `Map`, not to `Array`, so it cannot be called through `.` on an array. Use
a pipe instead:

```ant
let es = [Entry{ key: "a", value: 1 }]
let m = es.fromEntries() // error: Array[Entry[string, int]] has no field or function fromEntries
```

A repeated key drops the earlier values without a warning. Swapping a map whose values repeat loses
keys:

```ant
let owners = Map{ pen: "Ann", ink: "Bob", pad: "Ann" }
let byOwner = Map.fromEntries(owners.entries().map({ Entry{ key: x.value, value: x.key } }))
byOwner           // Map{ "Ann": "pad", "Bob": "ink" }
Map.size(byOwner) // 2
```

To keep every key, group them with [`Array.groupBy`](../Array/groupBy.md) instead.

## See also

[Map.entries](entries.md), [Entry](types/Entry.md), [Map.map](map.md), [Map.filter](filter.md),
[Map.merge](merge.md), [Array.groupBy](../Array/groupBy.md), [Array.entries](../Array/entries.md)
