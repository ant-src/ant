<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.merge

```ant signature
func Map.merge[K, V](m: Map[K, V], n: Map[K, V]): Map[K, V]
```

Map.merge returns a new map holding the entries of m and n, where the entries of n win. The keys of
m come first, followed by the keys of n that m lacks.

<!-- end generated -->

## Details

- Returns a new map; `m` and `n` are unchanged, even when one of them is empty.
- A key in both maps keeps its position in `m` and takes its value from `n`.
- The values are not copied: an array or a record held in `m` or `n` is the same one in the result.

## Basic examples

Combine two maps; for the shared key `b`, the value of the second map wins:

```ant
Map.merge(Map{ a: 1, b: 2 }, Map{ b: 20, c: 30 }) // Map{ "a": 1, "b": 20, "c": 30 }
```

Fill in defaults under the settings a user gave:

```ant
let defaults = Map{ theme: "light", lang: "en" }
let user = Map{ lang: "sv" }
Map.merge(defaults, user) // Map{ "theme": "light", "lang": "sv" }
```

## Scope

Merging with an empty map gives a new map equal to the other one, and two empty maps give an empty
map:

```ant
let empty: Map[string, int] = Map{}
Map.merge(Map{ a: 1 }, empty) // Map{ "a": 1 }
Map.merge(empty, Map{ a: 1 }) // Map{ "a": 1 }
Map.merge(empty, empty)       // Map{}
```

The order of the keys comes from `m`, even for the keys whose values come from `n`:

```ant
Map.merge(Map{ b: 0, a: 0 }, Map{ a: 1, z: 9, b: 2 }) // Map{ "b": 2, "a": 1, "z": 9 }
```

Keys of any type work, such as numbers or records:

```ant
type Cell = { row: int, col: int }
Map.merge(Map{ 1: "one" }, Map{ 2: "two" }) // Map{ 1: "one", 2: "two" }
let origin = Cell{ row: 0, col: 0 }
Map.merge(Map{ (origin): "x" }, Map{ (origin): "o" }) // Map{ Cell{ row: 0, col: 0 }: "o" }
```

When the values may be `null`, a stored `null` in `n` still wins:

```ant
let base: Map[string, int | null] = Map{ a: 1 }
let reset: Map[string, int | null] = Map{ a: null }
Map.merge(base, reset)          // Map{ "a": null }
Map.merge(base, reset).has("a") // true
```

Call it through `.`, in a pipe, or with named arguments:

```ant
let m = Map{ a: 1 }
let n = Map{ b: 2 }
m.merge(n)            // Map{ "a": 1, "b": 2 }
m | Map.merge(n)      // Map{ "a": 1, "b": 2 }
Map.merge(n: n, m: m) // Map{ "a": 1, "b": 2 }
```

## Applications

Merge a list of maps from left to right, so later maps override earlier ones:

```ant
let layers = [Map{ port: "80", host: "localhost" }, Map{ port: "8080" }, Map{ host: "example.com" }]
layers.reduce(Map{}, Map.merge) // Map{ "port": "8080", "host": "example.com" }
```

Add or replace one entry without changing the original map:

```ant
let prices = Map{ apple: 3, pear: 4 }
let updated = prices.merge(Map{ pear: 5 })
updated // Map{ "apple": 3, "pear": 5 }
prices  // Map{ "apple": 3, "pear": 4 }
```

Keep the existing values and only add the keys that are missing, by swapping the arguments:

```ant
let counts = Map{ ant: 3 }
let known = ["ant", "bee"].map({ Entry{ key: x, value: 0 } })
Map.merge(Map.fromEntries(known), counts) // Map{ "ant": 3, "bee": 0 }
```

## Properties and relations

The result has every key of `m` and of `n`, and each key of `n` reads its value from `n`:

```ant
let m = Map{ a: 1, b: 2 }
let n = Map{ b: 3, c: 4 }
let r = Map.merge(m, n)
assert(m.keys().every({ r.has(x) }) && n.keys().every({ r.has(x) }))
assert(n.entries().every({ r[x.key] == x.value }))
```

It builds the same map as `Map.fromEntries` over the entries of `m` followed by those of `n`:

```ant
let m = Map{ a: 1, b: 2 }
let n = Map{ b: 3, c: 4 }
let r = Map.fromEntries([...m.entries(), ...n.entries()])
assert(Map.merge(m, n) == r)
assert(Map.merge(m, n).keys() == r.keys())
```

Merging a map with itself gives an equal map, but a new one:

```ant
let m = Map{ a: 1 }
assert(Map.merge(m, m) == m)
assert(!Map.same(Map.merge(m, m), m))
```

## Possible issues

`Map.merge` is not symmetric. When both maps hold a key, the order of the arguments decides which
value stays:

```ant
let x = Map{ a: 1 }
let y = Map{ a: 2 }
Map.merge(x, y) // Map{ "a": 2 }
Map.merge(y, x) // Map{ "a": 1 }
```

It does not change `m`. A call whose result is thrown away has no effect; assign it, as in
`m = m.merge(n)`:

```ant
let m = Map{ a: 1 }
Map.merge(m, Map{ b: 2 })
println(m)
```

```text
Map{ "a": 1 }
```

The values are shared, not copied. Changing an array in the result changes it in `m` too:

```ant
let m = Map{ a: [1] }
let r = Map.merge(m, Map{})
r["a"]?.push(2)
m // Map{ "a": [1, 2] }
```

Both maps must have the same key and value types:

```ant
let a = Map{ a: 1 }
let b = Map{ b: "x" }
let c = Map.merge(a, b) // error: cannot use Map[string, string] where Map[string, int] is expected
```

## See also

[Map.set](set.md) · [Map.fromEntries](fromEntries.md) · [Map.entries](entries.md) ·
[Map.filter](filter.md) · [Map.same](same.md) · [Array.reduce](../Array/reduce.md)
