<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.entries

```ant signature
func Map.entries[K, V](m: Map[K, V]): Array[Entry[K, V]]
```

Map.entries returns the entries of m in insertion order.

<!-- end generated -->

## Details

- Replacing a value keeps its entry's position; a key removed and inserted again goes to the end.
- `Map.entries` returns a new array, so later changes to `m` do not change it, and changes to it do
  not change `m`. Each `value` is the stored value itself, not a copy.
- Each `Entry` is a new record, so assigning to its fields does not change `m`.

## Basic examples

List the entries of a map:

```ant
var ages = Map{ ann: 31, bob: 27 }
Map.entries(ages) // [Entry{ key: "ann", value: 31 }, Entry{ key: "bob", value: 27 }]
```

Read the fields of one entry:

```ant
var es = Map{ red: "#f00", green: "#0f0" }.entries()
es[1].key   // "green"
es[1].value // "#0f0"
```

Loop over the entries and destructure each one:

```ant
var stock = Map{ apples: 3, pears: 0 }
for { key, value } in stock.entries() {
  println("${key}: ${value}")
}
```

```text
apples: 3
pears: 0
```

## Scope

An empty map has no entries. Annotate it so that `K` and `V` are known:

```ant
var empty: Map[string, int] = Map{}
Map.entries(empty) // []
```

The order is the order keys were first inserted. Replacing a value keeps its place; a key removed
and inserted again goes to the end:

```ant
var m = Map{ a: 1, b: 2, c: 3 }
m["a"] = 10
Map.remove(m, "b") // 2
m["b"] = 20
m.entries() // [Entry{ key: "a", value: 10 }, Entry{ key: "c", value: 3 }, Entry{ key: "b", value: 20 }]
```

Keys of any type, records included, and `null` values come through as they are:

```ant
type Point = { x: int, y: int }
Map.entries(Map{ 1: "one", 2: "two" })         // [Entry{ key: 1, value: "one" }, Entry{ key: 2, value: "two" }]
Map.entries(Map{ (Point{ x: 1, y: 2 }): "a" }) // [Entry{ key: Point{ x: 1, y: 2 }, value: "a" }]
var maybe: Map[string, int | null] = Map{ a: null }
maybe.entries() // [Entry{ key: "a", value: null }]
```

Call it through `.` or in a pipe:

```ant
var m = Map{ x: 1 }
m.entries()       // [Entry{ key: "x", value: 1 }]
m | Map.entries() // [Entry{ key: "x", value: 1 }]
```

Rename the fields as you destructure them:

```ant
var ages = Map{ ann: 31, bob: 27 }
for { key: name, value: age } in ages.entries() {
  println("${name} is ${age}")
}
```

```text
ann is 31
bob is 27
```

## Applications

Find the key with the largest value:

```ant
var votes = Map{ tea: 3, coffee: 7, water: 5 }
votes.entries().sortBy({ -x.value })[0].key // "coffee"
```

Build a query string:

```ant
var params = Map{ q: "ants", page: "2" }
String.join(params.entries().map({ "${x.key}=${x.value}" }), "&") // "q=ants&page=2"
```

Sort a map by its keys:

```ant
var m = Map{ b: 2, c: 3, a: 1 }
Map.fromEntries(m.entries().sortBy({ x.key })) // Map{ "a": 1, "b": 2, "c": 3 }
```

Swap keys and values:

```ant
var codes = Map{ ok: 200, missing: 404 }
Map.fromEntries(codes.entries().map({ Entry{ key: x.value, value: x.key } })) // Map{ 200: "ok", 404: "missing" }
```

Remove entries while you walk the map. The array is taken before the loop starts, so removing from
`m` does not disturb it:

```ant
var sessions = Map{ s1: 0, s2: 5, s3: 0 }
for { key, value } in sessions.entries() {
  if value == 0 {
    Map.remove(sessions, key)
  }
}
println(sessions)
```

```text
Map{ "s2": 5 }
```

## Properties and relations

The keys and the values of the entries are `Map.keys` and `Map.values`, and `Map.fromEntries` turns
the entries back into an equal map:

```ant
var m = Map{ a: 1, b: 2, c: 3 }
var es = Map.entries(m)
assert(len(es) == Map.size(m))
assert(es.map({ x.key }) == Map.keys(m))
assert(es.map({ x.value }) == Map.values(m))
assert(Map.fromEntries(es) == m)
assert(es.every({ m[x.key] == x.value }))
```

`for` over the map itself visits the same entries in the same order:

```ant
var m = Map{ x: 1, y: 2 }
var seen: Array[Entry[string, int]] = []
for e in m {
  seen.push(e)
}
assert(seen == m.entries())
```

The result is a new array, so inserting into `m` afterwards leaves it as it was:

```ant
var m = Map{ a: 1 }
var es = m.entries()
m["b"] = 2
assert(len(es) == 1)
```

## Possible issues

The fields are `key` and `value`. `Array.entries` gives `Indexed` records with `index` and `value`,
so a destructuring written for an array does not fit:

```ant
var ages = Map{ ann: 31 }
for { index, value } in ages.entries() { // error: Entry[string, int] has no field index
  println("${index}: ${value}")
}
```

An entry holds the same value as the map, so a change made to a stored array shows in the entry:

```ant
var lists = Map{ a: [1] }
var es = lists.entries()
lists["a"]?.push(2)
es[0].value // [1, 2]
```

The other way round, assigning an entry's `value` does not change the map. Write `m[e.key] = ...` to
change it:

```ant
var m = Map{ a: 1 }
var es = m.entries()
es[0].value = 9
m // Map{ "a": 1 }
```

The same holds for the entries a `for` loop over the map visits:

```ant
var m = Map{ a: 1, b: 2 }
for e in m {
  e.value = 0
}
println(m)
```

```text
Map{ "a": 1, "b": 2 }
```

`Map.entries` takes a map, not a set:

```ant
println(Map.entries(Set.of([1]))) // error: cannot use Set[int] where Map[K, V] is expected
```

## See also

[Entry](types/Entry.md) · [Map.keys](keys.md) · [Map.values](values.md) ·
[Map.fromEntries](fromEntries.md) · [Map.map](map.md) · [Map.filter](filter.md) ·
[Array.entries](../Array/entries.md)
