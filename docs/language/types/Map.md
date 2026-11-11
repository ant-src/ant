<!-- status: reviewed -->

# Map

`Map[K, V]` is the built-in hash map from keys of type `K` to values of type `V`. It keeps its
entries in insertion order. The functions on maps are in the [Map](../../lib/Map/index.md)
namespace.

## Details

- Keys compare the way `==` does, so a record or an array works as a key and matches by its
  contents.
- A map keeps insertion order. Replacing a value keeps the entry's position; a key removed and
  inserted again goes to the end.
- `m[k]` has type `V | null`, and gives `null` for a missing key.
- A map is a reference: passing it to a function or binding it to another name shares it, and `let`
  stops only the rebinding of the name, not changes to the map.
- A key is hashed when it is inserted. Changing it afterwards is a program error, and its entry may
  become unreachable.

## Basic examples

Annotate a map with its key and value types:

```ant
let ages: Map[string, int] = Map{ ada: 36, linus: 54 }
ages // Map{ "ada": 36, "linus": 54 }
```

Without an annotation, the type comes from the literal:

```ant
let codes = Map{ 200: "OK", 404: "Not Found" }
let label: string = codes[404] ?? "unknown"
label // "Not Found"
```

An empty map takes its types from where it goes:

```ant
let seen: Map[string, bool] = Map{}
seen["ant"] = true
seen // Map{ "ant": true }
```

## Scope

### Key types

Any type works as a key. Number, record, and array keys compare by value:

```ant
type Point = { x: int, y: int }
let names = Map{ (Point{ x: 0, y: 0 }): "origin" }
names[Point{ x: 0, y: 0 }] // "origin"
let sizes = Map{ (["usr", "bin"]): 12 }
sizes[["usr", "bin"]] // 12
```

A named primitive type works as a key, since `==` works on it:

```ant
type UserId = string
type Account = { owner: string, balance: int }
let accounts: Map[UserId, Account] = Map{}
accounts[UserId("u-1")] = Account{ owner: "ann", balance: 10 }
accounts[UserId("u-1")]?.owner // "ann"
```

A union key type takes keys of several types:

```ant
let mixed: Map[string | int, string] = Map{ 1: "one", two: "2" }
mixed[1]     // "one"
mixed["two"] // "2"
```

### Value types

Values may be arrays, maps, functions, or nullable:

```ant
let teams: Map[string, Array[string]] = Map{ red: ["ann", "bob"], blue: [] }
teams["red"]?.length() // 2
let nested: Map[string, Map[string, int]] = Map{ en: Map{ one: 1 } }
nested["en"]?.size() // 1
let limits: Map[string, int | null] = Map{ cpu: 4, memory: null }
limits.has("memory") // true
let steps: Map[string, func(int): int] = Map{ double: { x * 2 } }
steps["double"]?.(21) // 42
```

### Fields and parameters

A map is an ordinary type in a record field, a parameter, or a return type. A field default of
`Map{}` builds a new map for every record:

```ant
type Cache = { hits: int, entries: Map[string, string] = Map{} }
let cache = Cache{ hits: 0 }
cache.entries["home"] = "<h1>Home</h1>"
cache // Cache{ hits: 0, entries: Map{ "home": "<h1>Home</h1>" } }
```

### Type aliases

An alias names a map type that recurs:

```ant
type Inventory = Map[string, int]
let stock: Inventory = Map{ apples: 3 }
stock.size() // 1
```

### Generic functions

A function generic over `K` and `V` accepts any map:

```ant
func invert[K, V](m: Map[K, V]): Map[V, K] {
  m
    | Map.entries()
    | Array.map({ Entry{ key: x.value, value: x.key } })
    | Map.fromEntries()
}

println(invert(Map{ a: 1, b: 2 }))
```

```text
Map{ 1: "a", 2: "b" }
```

### Iteration

A `for` loop visits the entries in insertion order, each as an [Entry](../../lib/Map/types/Entry.md)
record:

```ant
let prices = Map{ tea: 3, cake: 5 }
for { key, value } in prices {
  println("${key} costs ${value}")
}
```

```text
tea costs 3
cake costs 5
```

### In a union

`is Map` tests whether a value of a union is a map:

```ant
func describe(v: Map[string, int] | Array[int]): string {
  if v is Map { "a map of ${v.size()}" } else { "an array of ${v.length()}" }
}

println(describe(Map{ a: 1, b: 2 }))
println(describe([1, 2, 3]))
```

```text
a map of 2
an array of 3
```

### Rendering

`println` and `String.from` render a map as a literal, with string keys always quoted:

```ant
let empty: Map[string, int] = Map{}
println(Map{ a: [1], b: [] })
println(Map{ 1: "one", 2: "two" })
println(empty)
```

```text
Map{ "a": [1], "b": [] }
Map{ 1: "one", 2: "two" }
Map{}
```

## Applications

### Counting

A function that fills a map it is given changes the caller's map:

```ant
func count(words: Array[string], into: Map[string, int]) {
  for word in words {
    into[word] = (into[word] ?? 0) + 1
  }
}

let counts: Map[string, int] = Map{}
count(["ant", "bee", "ant"], counts)
count(["bee"], counts)
println(counts)
```

```text
Map{ "ant": 2, "bee": 2 }
```

### Grouping

`Array.groupBy` returns a map from each key to its elements, in first-seen order:

```ant
type Order = { customer: string, total: int }
let orders = [
  Order{ customer: "ann", total: 5 },
  Order{ customer: "bob", total: 3 },
  Order{ customer: "ann", total: 7 },
]
let byCustomer: Map[string, Array[Order]] = orders.groupBy({ x.customer })
for { key, value } in byCustomer {
  let total = value.map({ x.total }).sum() ?? 0
  println("${key}: ${total}")
}
```

```text
ann: 12
bob: 3
```

## Properties and relations

Two maps are `==` when they have the same keys and equal values per key, in any order. `Map.same`
asks whether they are one map:

```ant
let a = Map{ x: 1, y: 2 }
let b = Map{ y: 2, x: 1 }
assert(a == b)
assert(a.keys() != b.keys())
assert(!Map.same(a, b))
let c = a
assert(Map.same(a, c))
```

`==` compares nested maps by their contents too:

```ant
let a = Map{ en: Map{ hi: "hello" } }
let b = Map{ en: Map{ hi: "hello" } }
assert(a == b)
```

## Possible issues

### A read is nullable

`m[k]` is `V | null`, so it does not fit where a `V` is expected. Default it with `??` or narrow it
with a guard:

```ant
let ages = Map{ ada: 36 }
let age: int = ages["ada"] // error: cannot use int | null where int is expected
```

### Type arguments are invariant

A map with a narrower value type is not assignable to one with a wider value type, since a write
through the wider type could store a value the narrower map does not allow:

```ant
let counts: Map[string, int] = Map{ a: 1 }
let maybe: Map[string, int | null] = counts
// error: cannot use Map[string, int] where Map[string, int | null] is expected
```

Build the map with the wider type from the start:

```ant
let maybe: Map[string, int | null] = Map{ a: 1 }
maybe["b"] = null
maybe // Map{ "a": 1, "b": null }
```

### Two map types in one union

A union holds at most one map type. Widen the key or value type instead:

```ant
type Table = Map[string, int] | Map[int, int]
// error: a union may contain each discriminator at most once: Map[string, int] and Map[int, int] are both a map; write Map[string | int, int]
```

### An empty literal needs a type

`Map{}` with nothing to give it a type is a compile error:

```ant
let seen = Map{} // error: cannot infer the key and value types of an empty map; annotate it
```

### Shared maps

Binding a map to another name does not copy it. A change through one name shows through the other:

```ant
let m = Map{ a: 1 }
let n = m
n["b"] = 2
m // Map{ "a": 1, "b": 2 }
```

### Changing a key

Changing an array or record after using it as a key is a program error: memory stays safe, but the
entry may become unreachable.

```ant
let path = ["usr"]
let sizes = Map{ (path): 10 }
// a program error: the key of an entry in sizes changes
path.push("bin")
```

### A stored null reads like a missing key

When `V` includes `null`, `m[k]` gives `null` both for a stored `null` and for a missing key.
`Map.has` tells a missing key from a stored `null`:

```ant
let limits: Map[string, int | null] = Map{ memory: null }
limits["memory"]     // null
limits["disk"]       // null
limits.has("memory") // true
limits.has("disk")   // false
```

### A bare word is a string key

A bare word before `:` in a map literal is a string key, not a variable, even when a variable of
that name is in scope. Parenthesize the name to use the variable. See
[Map literals](../literals/maps.md) for details.

```ant
let name = "ada"
Map{ name: 1 }   // Map{ "name": 1 }
Map{ (name): 1 } // Map{ "ada": 1 }
```

## See also

- [Map](../../lib/Map/index.md): the functions on maps
- [Entry](../../lib/Map/types/Entry.md): the record a map yields
- [Map literals](../literals/maps.md), [Indexing](../operators/indexing.md)
- [Array](Array.md), [Set](Set.md): the other built-in containers
- [Collections](../../guide/collections.md)
