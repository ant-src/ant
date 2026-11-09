<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.keys

```ant signature
func Map.keys[K, V](m: Map[K, V]): Array[K]
```

Map.keys returns the keys of m in insertion order.

<!-- end generated -->

## Details

- Keys are in the order they were first inserted. Replacing a value keeps a key's place; a key
  removed and inserted again moves to the end.
- Returns a new array. Changing it leaves `m` unchanged, and changing `m` afterwards leaves the
  array unchanged.

## Basic examples

```ant
let ages = Map{ ada: 36, linus: 54, grace: 85 }
Map.keys(ages) // ["ada", "linus", "grace"]
```

Call it through `.` on the map:

```ant
let words = Map{ 3: "three", 1: "one", 2: "two" }
words.keys() // [3, 1, 2]
```

## Scope

An empty map gives an empty array. Annotate the map so that its types are known:

```ant
let empty: Map[string, int] = Map{}
empty.keys() // []
```

Replacing a value keeps the key where it was. Removing a key and setting it again moves it to the
end:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
m["a"] = 10
m.keys()      // ["a", "b", "c"]
m.remove("a") // 10
m["a"] = 1
m.keys() // ["b", "c", "a"]
```

Records and arrays can be keys, and come back as they went in:

```ant
type Point = { x: int, y: int }
let names = Map{ Point{ x: 0, y: 0 }: "origin", Point{ x: 1, y: 2 }: "p" }
names.keys()              // [Point{ x: 0, y: 0 }, Point{ x: 1, y: 2 }]
Map{ [1, 2]: "a" }.keys() // [[1, 2]]
```

A map whose key type includes `null` can hold the key `null`. Write it as `(null)` in a literal:

```ant
let m: Map[string | null, int] = Map{ (null): 0, "a": 1 }
m.keys() // [null, "a"]
```

Keys may be any string:

```ant
Map{ "🐜": 1, "é": 2, "": 3 }.keys() // ["🐜", "é", ""]
```

In a pipeline, the map is the first argument:

```ant
Map{ b: 1, a: 2 } | Map.keys() // ["b", "a"]
```

Name the argument when that reads better:

```ant
Map.keys(m: Map{ x: 1 }) // ["x"]
```

## Applications

Print a map in key order. The map keeps insertion order, so sort the keys first:

```ant
let stock = Map{ pear: 3, apple: 5, plum: 1 }
for k in stock.keys().sort() {
  println("${k}: ${stock[k]}")
}
```

```text
apple: 5
pear: 3
plum: 1
```

Remove entries while walking the keys. The array is a copy, so removing from the map does not
disturb the loop:

```ant
let stock = Map{ apple: 0, pear: 3, plum: 0 }
for k in stock.keys() {
  if stock[k] == 0 {
    stock.remove(k)
  }
}
println(stock)
```

```text
Map{ "pear": 3 }
```

List the words that appear more than once, in the order they first appear:

```ant
let words = ["the", "ant", "and", "the", "bee", "ant"]
words.groupBy({ x }).filter({ len(x.value) > 1 }).keys() // ["the", "ant"]
```

## Properties and relations

The keys are the keys of `Map.entries`, one per entry, each distinct, and each one is in the map:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
let ks = m.keys()
assert(ks == m.entries().map({ x.key }))
assert(len(ks) == m.size())
assert(ks == ks.unique())
assert(ks.every({ m.has(x) }))
```

`Map.keys` and `Map.values` line up: the value at each position belongs to the key at that position:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
let ks = m.keys()
let vs = m.values()
for i in Array.range(0, len(ks)) {
  assert(m[ks[i]] == vs[i])
}
```

The array is a snapshot. Pushing to it leaves the map alone, and inserting into the map leaves the
array alone:

```ant
let m = Map{ a: 1 }
let ks = m.keys()
ks.push("z")
m["b"] = 2
assert(ks == ["a", "z"])
assert(m.keys() == ["a", "b"])
```

`for` over a map visits the same keys in the same order. A loop that does not change the map can
iterate `m` directly instead of `m.keys()`:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
let seen: Array[string] = []
for { key } in m {
  seen.push(key)
}
assert(seen == m.keys())
```

## Possible issues

`for k in m` binds each `Entry`, not each key. Read the key as `k.key`:

```ant
let m = Map{ a: 1 }
for k in m {
  println(k)
  println(k.key)
}
```

```text
Entry{ key: "a", value: 1 }
a
```

Two maps with the same entries are `==` in any order, but their keys are arrays, and arrays compare
in order:

```ant
let m = Map{ a: 1, b: 2 }
let n = Map{ b: 2, a: 1 }
m == n                             // true
m.keys() == n.keys()               // false
m.keys().sort() == n.keys().sort() // true
```

In a map literal, a bare word before `:` is a string key, and `null` is no exception. Write `(null)`
for the value `null`:

```ant
let m: Map[string | null, int] = Map{ null: 0 }
m.keys() // ["null"]
```

`NaN` is not `==` to itself, so each `NaN` key is a separate entry:

```ant
Map{ 1.0: 1, Math.NaN: 2, Math.NaN: 3 }.keys() // [1.0, NaN, NaN]
```

`Map.keys` is a function, not a field. Without the parentheses it is a compile error:

```ant
let m = Map{ a: 1 }
println(m.keys) // error: Map[string, int] is not a record, so it has no field keys
```

## See also

[Map.values](values.md) · [Map.entries](entries.md) · [Map.has](has.md) · [Map.size](size.md) ·
[Set.values](../Set/values.md) · [Array.sort](../Array/sort.md) ·
[Array.groupBy](../Array/groupBy.md)
