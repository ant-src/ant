<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.size

```ant signature
func Map.size[K, V](m: Map[K, V]): int
```

Map.size returns the number of entries in m.

<!-- end generated -->

## Details

- `Map.size` returns the count at the time of the call. Inserting a new key adds one, removing a key
  takes one away, and replacing the value of a key that is already there changes nothing.
- Keys compare the way `==` does, so two equal records or arrays are one key and one entry.
- An entry whose value is `null` is still an entry and counts.
- `len(m)` returns the same count.

## Basic examples

```ant
Map.size(Map{ ada: 36, linus: 54 }) // 2
let none: Map[string, int] = Map{}
Map.size(none) // 0
```

Call it through `.` on any map:

```ant
let ages = Map{ ada: 36, linus: 54, grace: 85 }
ages.size() // 3
```

## Scope

Equal keys are one entry. Setting a record key that is `==` to one already in the map replaces its
value:

```ant
type Point = { x: int, y: int }

let labels = Map{ (Point{ x: 1, y: 2 }): "a" }
labels[Point{ x: 1, y: 2 }] = "b"
labels.size() // 1
```

An entry whose value is `null` counts like any other:

```ant
let seen: Map[string, int | null] = Map{ ada: null, linus: 54 }
seen.size() // 2
```

The size follows the map as it changes:

```ant
let counts: Map[string, int] = Map{}
counts["go"] = 1
counts["ant"] = 1
counts.size() // 2
counts["go"] = 2
counts.size()        // 2
counts.remove("go")  // 2
counts.size()        // 1
counts.remove("zig") // null
counts.size()        // 1
```

In a pipeline, the map is the first argument:

```ant
Map{ ada: 36, linus: 54 } | Map.size() // 2
```

As a function value, it maps an array of maps to their sizes:

```ant
[Map{ a: 1 }, Map{ a: 1, b: 2 }].map(Map.size) // [1, 2]
```

For a map that may be `null`, `?.` yields `null` instead of calling:

```ant
let headers: Map[string, string] | null = null
headers?.size()      // null
headers?.size() ?? 0 // 0
```

## Named arguments

The parameter is named `m`:

```ant
Map.size(m: Map{ a: 1 }) // 1
```

## Applications

Count the distinct words of a text:

```ant
func distinct(text: string): int {
  let counts: Map[string, int] = Map{}
  for word in String.fields(text) {
    counts[word] = (counts[word] ?? 0) + 1
  }
  counts.size()
}

println(distinct("to be or not to be"))
```

```text
4
```

Report how many entries pass a test, out of how many:

```ant
let stock = Map{ apples: 12, pears: 0, plums: 3, figs: 0 }
let out = stock.filter({ x.value == 0 })
println("${out.size()} of ${stock.size()} items out of stock: ${out.keys()}")
```

```text
2 of 4 items out of stock: ["pears", "figs"]
```

## Properties and relations

`Map.size` agrees with `len`, `Map.isEmpty`, and the lengths of `Map.keys`, `Map.values`, and
`Map.entries`:

```ant
let cases: Array[Map[string, int]] = [Map{}, Map{ a: 1 }, Map{ a: 1, b: 2 }]
for m in cases {
  assert(m.size() == len(m))
  assert(m.isEmpty() == (m.size() == 0))
  assert(m.size() == m.keys().length())
  assert(m.size() == m.values().length())
  assert(m.size() == m.entries().length())
}
```

`Map.map` keeps the size, a filter and its complement split it, and `Map.merge` counts each key of
both maps once:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
let n = Map{ c: 30, d: 40 }
assert(m.map({ x.value * 10 }).size() == m.size())
let small = m.filter({ x.value < 2 })
let large = m.filter({ x.value >= 2 })
assert(small.size() + large.size() == m.size())
let shared = m.filter({ n.has(x.key) }).size()
assert(m.merge(n).size() == m.size() + n.size() - shared)
```

## Possible issues

`size` is a function, not a field. Without parentheses it is a compile error:

```ant
let m = Map{ a: 1 }
println(m.size) // error: Map[string, int] is not a record, so it has no field size
```

An empty map literal has no key and value types to infer. Annotate it:

```ant
println(Map{}.size()) // error: cannot infer the key and value types of an empty map; annotate it
```

A nullable map must be narrowed or reached through `?.` first:

```ant
let m: Map[string, int] | null = null
println(m.size())
// error: Map[string, int] | null has no field or function size on every member: narrow the value first
```

`Map.size` takes only maps. Use [`Set.size`](../Set/size.md) for a set,
[`Array.length`](../Array/length.md) for an array, or [`len`](../prelude/len.md), which takes them
all:

```ant
println(Map.size(Set.of([1]))) // error: cannot use Set[int] where Map[K, V] is expected
```

Repeated keys collapse to one entry, so the size can be less than the number of entries passed to
`Map.fromEntries`:

```ant
Map.fromEntries([Entry{ key: "a", value: 1 }, Entry{ key: "a", value: 2 }]).size() // 1
```

A map is a reference value: assigning or passing one shares it. An insert through one name changes
the size seen through the other:

```ant
let a = Map{ x: 1 }
let b = a
b["y"] = 2
a.size() // 2
```

## See also

[Map.isEmpty](isEmpty.md) · [Map.keys](keys.md) · [Map.has](has.md) · [Map.set](set.md) ·
[Map.remove](remove.md) · [len](../prelude/len.md) · [Set.size](../Set/size.md) ·
[Array.length](../Array/length.md)
