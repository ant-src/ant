<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map

Map holds the functions of the built-in map type Map\[K, V\], a hash map keyed by == that keeps
insertion order. m\[k\] reads V \| null and m\[k\] = v inserts or replaces; a function that builds a
map from another keeps its order.

<!-- end generated -->

The `Map` namespace holds the functions on maps. Each function that takes a map takes it first, so
`Map.keys(m)`, `m.keys()`, and `m | Map.keys()` are the same call. `Map.fromEntries` builds a map
from an array and is called only through the namespace.

## Details

- `m[k]` and [Map.get](get.md) return `null` for a missing key; neither panics. When `V` includes
  `null`, only [Map.has](has.md) tells a missing key from a stored `null`.
- A map keeps insertion order. Replacing a value keeps the entry's position; a key removed and
  inserted again goes to the end.
- `m[k] = v`, [Map.set](set.md), and [Map.remove](remove.md) change the map in place. Every other
  function returns a new value and leaves its arguments untouched.
- [Map.map](map.md) and [Map.filter](filter.md) call `f` once per entry, in insertion order.
- Each [Entry](types/Entry.md) that a `for` loop visits or [Map.entries](entries.md) returns is a
  new record. Assigning to its fields does not change the map.
- A map is a reference: copying a variable shares the map. Two maps are `==` when they have the same
  keys and equal values per key, in any order.

## Basic examples

`Map{ ... }` builds a map, `m[k]` reads a value, and `m[k] = v` inserts or replaces one:

```ant
let ages = Map{ ada: 36, linus: 54 }
ages["ada"]   // 36
ages["grace"] // null
ages["grace"] = 85
ages // Map{ "ada": 36, "linus": 54, "grace": 85 }
```

A `for` loop visits the entries in insertion order. Destructure each [Entry](types/Entry.md) into
its key and value:

```ant
let stock = Map{ apples: 3, pears: 0 }
for { key, value } in stock {
  println("${key}: ${value}")
}
```

```text
apples: 3
pears: 0
```

Calls chain through `.`, or the map goes down a pipe:

```ant
let prices = Map{ tea: 3, cake: 5, soup: 7 }
prices.filter({ x.value > 4 }).keys() // ["cake", "soup"]
prices | Map.values() | Array.sum()   // 15
```

## Scope

A single word before `:` is a string key. Any other expression is evaluated as the key, and a
parenthesized name is the variable. A bare name alone is shorthand for `"name": name`:

```ant
let name = "grace"
Map{ name }               // Map{ "name": "grace" }
Map{ 1: "one", 2: "two" } // Map{ 1: "one", 2: "two" }
```

An empty literal needs a type, from an annotation or from where it goes:

```ant
let empty: Map[string, int] = Map{}
empty            // Map{}
empty.isEmpty()  // true
empty["missing"] // null
```

Keys compare by `==`, so records and arrays can be keys and match by contents:

```ant
type Point = { x: int, y: int }
let grid = Map{ (Point{ x: 0, y: 0 }): "origin" }
grid[Point{ x: 0, y: 0 }] // "origin"
let routes = Map{ (["a", "b"]): 2 }
routes[["a", "b"]] // 2
```

When `V` includes `null`, a stored `null` and a missing key read the same. Ask `Map.has`:

```ant
let settings: Map[string, string | null] = Map{ theme: null }
settings["theme"]     // null
settings["font"]      // null
settings.has("theme") // true
settings.has("font")  // false
```

Values can be any type, including arrays and other maps:

```ant
let teams = Map{ red: ["ann", "bob"], blue: [] }
teams["blue"] // []
let nested = Map{ en: Map{ hi: "hello" } }
nested["en"]?.["hi"] // "hello"
```

## Functions by job

### Size and lookup

- [Map.size](size.md): the number of entries
- [Map.isEmpty](isEmpty.md): whether there are none
- [Map.has](has.md): whether there is an entry for a key
- [Map.get](get.md): the value for a key, or `null`; the same as `m[k]`

### Changing a map in place

- [Map.set](set.md): store a value under a key; the same as `m[k] = v`
- [Map.remove](remove.md): remove a key and return its value, or `null`

### Keys, values, and entries

- [Map.keys](keys.md): the keys in insertion order
- [Map.values](values.md): the values in insertion order
- [Map.entries](entries.md): the entries, as [Entry](types/Entry.md) records
- [Map.fromEntries](fromEntries.md): a map from an array of entries; later entries win

### Transforming and combining

- [Map.map](map.md): the same keys, each value computed from its entry
- [Map.filter](filter.md): the entries a predicate accepts
- [Map.merge](merge.md): the entries of two maps; the second map's entries win

### Identity

- [Map.same](same.md): whether two maps are the same map, not only equal

## Applications

Count words. `?? 0` starts a missing count at zero, and the counts keep first-seen order:

```ant
let counts: Map[string, int] = Map{}
for word in ["ant", "bee", "ant", "fly", "ant"] {
  counts[word] = (counts[word] ?? 0) + 1
}
counts // Map{ "ant": 3, "bee": 1, "fly": 1 }
```

Index records by a field to look them up later:

```ant
type User = { id: string, name: string }
let users = [User{ id: "u1", name: "Ada" }, User{ id: "u2", name: "Linus" }]
let byId = users.map({ Entry{ key: x.id, value: x } }) | Map.fromEntries()
byId["u2"]?.name // "Linus"
```

Fill in options from defaults. The keys of the defaults come first, and the given values win:

```ant
let defaults = Map{ color: "auto", width: "80" }
let given = Map{ width: "120" }
Map.merge(defaults, given) // Map{ "color": "auto", "width": "120" }
```

Walk a graph stored as a map from each node to its neighbors. `?? []` gives a node with no entry an
empty list:

```ant
let edges = Map{ a: ["b", "c"], b: ["c"] }
for node in ["a", "b", "c"] {
  for next in edges[node] ?? [] {
    println("${node} -> ${next}")
  }
}
```

```text
a -> b
a -> c
b -> c
```

## Properties and relations

`m[k]` and `Map.get` agree, and `len` and `Map.size` agree:

```ant
let m = Map{ a: 1, b: 2 }
assert(m["a"] == m.get("a"))
assert(len(m) == m.size())
assert(m.isEmpty() == (m.size() == 0))
```

`==` ignores order, while `Map.keys` shows it:

```ant
let m = Map{ a: 1, b: 2 }
let n = Map{ b: 2, a: 1 }
assert(m == n)
assert(m.keys() != n.keys())
```

The entries rebuild an equal map, and keys and values line up with them:

```ant
let m = Map{ x: 1, y: 2, z: 3 }
assert(Map.fromEntries(m.entries()) == m)
assert(m.entries().map({ x.key }) == m.keys())
assert(m.entries().map({ x.value }) == m.values())
```

## Possible issues

`m[k]` has type `V | null`, so it does not fit where a `V` is expected. Give a fallback with `??`:

```ant
let ages = Map{ ada: 36 }
let age: int = ages["ada"] // error: cannot use int | null where int is expected
```

```ant
let ages = Map{ ada: 36 }
let age: int = ages["ada"] ?? 0
age // 36
```

A bare word before `:` is a string, not a variable. Parenthesize the name to use its value:

```ant
let key = "id"
Map{ key: 1 }   // Map{ "key": 1 }
Map{ (key): 1 } // Map{ "id": 1 }
```

A key is hashed when it goes in. Changing a record or array after using it as a key is a program
error: the entry may become unreachable. Use a copy, or a key you do not change.

An empty `Map{}` with nothing to give it a type is a compile error:

```ant
let seen = Map{} // error: cannot infer the key and value types of an empty map; annotate it
```

In the header of a `for`, a map literal must be parenthesized:

```ant
for e in Map{ a: 1 } { // error: a record or map literal in the for header must be parenthesized
  println(e)
}
```

```ant
for e in (Map{ a: 1 }) {
  println(e)
}
```

```text
Entry{ key: "a", value: 1 }
```

Copying a map variable copies the reference. A change through one name shows through the other:

```ant
let m = Map{ a: 1 }
let n = m
n["b"] = 2
m              // Map{ "a": 1, "b": 2 }
Map.same(m, n) // true
```

To get a separate map, merge with an empty one:

```ant
let m = Map{ a: 1 }
let n = Map.merge(m, Map{})
n["b"] = 2
m // Map{ "a": 1 }
```

Assigning to the fields of an entry from a loop or from `Map.entries` changes only that record. Set
the value through the map:

```ant
let stock = Map{ apples: 3, pears: 0 }
for e in stock {
  e.value = 10
}
stock // Map{ "apples": 3, "pears": 0 }
for { key } in stock {
  stock[key] = 10
}
stock // Map{ "apples": 10, "pears": 10 }
```

A loop reads the map as it goes. It skips the removed `b` and visits the added `d`:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
for { key, value } in m {
  if value == 1 {
    m.remove("b")
  }
  if key == "c" {
    m["d"] = 4
  }
  println(key)
}
```

```text
a
c
d
```

## See also

- [Entry](types/Entry.md), the record a map yields
- [Array](../Array/index.md), [Set](../Set/index.md): the other built-in containers
- [Array.groupBy](../Array/groupBy.md): a map from each key to its elements
- [len](../prelude/len.md): the number of entries, as for any container
