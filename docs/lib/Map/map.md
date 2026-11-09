<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.map

```ant signature
func Map.map[K, V, W](m: Map[K, V], f: func(Entry[K, V]): W): Map[K, W]
```

Map.map returns a new map with the keys of m, each mapped to f applied to its entry. It keeps the
order of m.

<!-- end generated -->

## Details

- `f` returns the new value only; the keys stay the same. To change keys, rebuild the map with
  [`Map.fromEntries`](fromEntries.md).
- `f` is called once per entry, in insertion order.

## Basic examples

Multiply each value by ten:

```ant
Map.map(Map{ a: 1, b: 2 }, { x.value * 10 }) // Map{ "a": 10, "b": 20 }
```

Read the key and the value of each entry. The values of the result can have another type:

```ant
Map.map(Map{ a: 1, b: 2 }, { "${x.key}=${x.value}" }) // Map{ "a": "a=1", "b": "b=2" }
```

## Scope

An empty map gives an empty map:

```ant
var empty: Map[string, int] = Map{}
Map.map(empty, { x.value * 2 }) // Map{}
```

The result keeps the order of `m`, which is the order the keys were inserted, not sorted order:

```ant
Map.map(Map{ b: 2, a: 1 }, { x.value + 1 }) // Map{ "b": 3, "a": 2 }
```

Read a field of each record value:

```ant
type User = { name: string, age: int }
var users = Map{ u1: User{ name: "Ann", age: 31 }, u2: User{ name: "Bob", age: 27 } }
Map.map(users, { x.value.name }) // Map{ "u1": "Ann", "u2": "Bob" }
```

Keys can be records too. `x.key` is then the record:

```ant
type Point = { x: int, y: int }
var labels = Map{ Point{ x: 0, y: 0 }: "origin", Point{ x: 1, y: 2 }: "p" }
Map.map(labels, { x.key.x + x.key.y }) // Map{ Point{ x: 0, y: 0 }: 0, Point{ x: 1, y: 2 }: 3 }
```

Replace the `null` values of a map with `??`:

```ant
var scores: Map[string, int | null] = Map{ ann: 90, bob: null }
Map.map(scores, { x.value ?? 0 }) // Map{ "ann": 90, "bob": 0 }
```

An empty literal on the right of `??` takes the type of the left side, so it needs no annotation:

```ant
var tags: Map[string, Array[string] | null] = Map{ ann: ["admin"], bob: null }
Map.map(tags, { x.value ?? [] }) // Map{ "ann": ["admin"], "bob": [] }
```

Handle each member of a union with `is`, which narrows `x.value` in each branch:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
var shapes: Map[string, Shape] = Map{ sun: Circle{ radius: 1.0 }, door: Rect{ width: 1.0, height: 2.0 } }
Map.map(shapes, { if x.value is Circle { "circle" } else { "rect" } }) // Map{ "sun": "circle", "door": "rect" }
```

Pass a named function that takes an `Entry`:

```ant
func describe(e: Entry[string, int]): string { "${e.key} has ${e.value}" }
Map.map(Map{ ant: 3, bee: 1 }, describe) // Map{ "ant": "ant has 3", "bee": "bee has 1" }
```

Call it through `.` or in a pipe:

```ant
var stock = Map{ pen: 3, ink: 0 }
stock.map({ x.value > 0 })       // Map{ "pen": true, "ink": false }
stock | Map.map({ x.value + 1 }) // Map{ "pen": 4, "ink": 1 }
```

Name the arguments to pass them in any order:

```ant
Map.map(f: { x.key }, m: Map{ a: 1, b: 2 }) // Map{ "a": "a", "b": "b" }
```

## Applications

Label each item by whether it is in stock:

```ant
var stock = Map{ pen: 3, ink: 0 }
Map.map(stock, { if x.value > 0 { "in stock" } else { "sold out" } }) // Map{ "pen": "in stock", "ink": "sold out" }
```

Format each entry, then join the values into one line:

```ant
var ages = Map{ ann: 31, bob: 27 }
Map.map(ages, { "${x.key} is ${x.value}" }) | Map.values() | String.join(", ") // "ann is 31, bob is 27"
```

Map, then keep some entries with [`Map.filter`](filter.md):

```ant
var prices = Map{ pen: 1.5, ink: 2.25, pad: 4.0 }
prices | Map.map({ x.value * 2.0 }) | Map.filter({ x.value < 5.0 }) // Map{ "pen": 3.0, "ink": 4.5 }
```

Loop over the result, destructuring each entry:

```ant
var ages = Map{ ann: 31, bob: 27 }
for { key, value } in Map.map(ages, { x.value + 1 }) {
  println("${key} turns ${value}")
}
```

```text
ann turns 32
bob turns 28
```

## Properties and relations

The result has the keys of `m`, in the same order:

```ant
var m = Map{ b: 2, a: 1 }
assert(Map.keys(Map.map(m, { x.value * 2 })) == Map.keys(m))
```

Its values are `f` applied to the entries of `m`, in order:

```ant
var m = Map{ a: 1, b: 2 }
var f: func(Entry[string, int]): int = { x.value * x.value }
assert(Map.values(Map.map(m, f)) == Map.entries(m).map(f))
```

`Map.map` is the same as rebuilding the map from its entries with
[`Map.fromEntries`](fromEntries.md):

```ant
var m = Map{ a: 1, b: 2 }
var f: func(Entry[string, int]): int = { x.value + 10 }
assert(Map.map(m, f) == Map.fromEntries(Map.entries(m).map({ Entry{ key: x.key, value: f(x) } })))
```

`f` sees the entries one at a time, in insertion order:

```ant
var seen: Array[string] = []
Map.map(Map{ b: 2, a: 1 }, func(e: Entry[string, int]): int {
  Array.push(seen, e.key)
  e.value
})
assert(seen == ["b", "a"])
```

Mapping each entry to its value gives an equal map that is not the same map:

```ant
var m = Map{ a: 1, b: 2 }
var n = Map.map(m, { x.value })
assert(n == m)
assert(!Map.same(n, m))
```

## Possible issues

`f` takes the entry, not the value. Treating `x` as the value is a compile error:

```ant
var m = Map{ a: 1 }
var doubled = Map.map(m, { x * 2 })
// error: * needs both operands of one type, got Entry[string, int] and int
```

`Map.map` returns a map, not an array. To get an array of results, map the entries:

```ant
Map.entries(Map{ a: 1, b: 2 }).map({ x.value * 10 }) // [10, 20]
```

A function of two parameters, a key and a value, is a compile error too:

```ant
var m = Map{ a: 1 }
var values = Map.map(m, func(k: string, v: int): int { v })
// error: cannot use func(string, int): int where func(Entry[string, int]): W is expected
```

`Map.map` cannot change the keys. To build a map with new keys, map the entries and rebuild it with
[`Map.fromEntries`](fromEntries.md). When two new keys are equal, the later entry wins:

```ant
var codes = Map{ a: 1, b: 2 }
Map.fromEntries(Map.entries(codes).map({ Entry{ key: x.value, value: x.key } })) // Map{ 1: "a", 2: "b" }
```

An expected type fixes `W` before the lambda is checked. Integer literals are always `int`, even
where a `float64` is expected:

```ant
var m = Map{ a: 1 }
var halves: Map[string, float64] = Map.map(m, { x.value * 2 })
// error: cannot use int where float64 is expected
```

A lambda whose body is an empty literal fixes nothing, so `W` stays unsolved:

```ant
var lists = Map.map(Map{ a: 1 }, { [] })
// error: cannot infer the element type of an empty array; annotate it
```

Annotate the result to fix `W`:

```ant
var lists: Map[string, Array[int]] = Map.map(Map{ a: 1 }, { [] })
lists // Map{ "a": [] }
```

The result holds what `f` returns. When `f` returns a value of `m` that is an array or a record,
both maps hold the same value, so a change to it shows through:

```ant
var groups = Map{ a: [1], b: [2] }
var copy = Map.map(groups, { x.value })
var first = copy["a"]
if first != null { Array.push(first, 9) }
groups // Map{ "a": [1, 9], "b": [2] }
```

## See also

[Map.filter](filter.md), [Map.entries](entries.md), [Map.fromEntries](fromEntries.md),
[Map.values](values.md), [Array.map](../Array/map.md)
