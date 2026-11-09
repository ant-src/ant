<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Entry

```ant signature
type Entry[K, V] = { key: K, value: V }
```

Entry is a key and its value, as a map yields them.

<!-- end generated -->

## Details

- `for` over a map, [`Map.entries`](../entries.md), and the callbacks of [`Map.map`](../map.md) and
  [`Map.filter`](../filter.md) give one `Entry[K, V]` per key, in insertion order.
  [`Map.fromEntries`](../fromEntries.md) turns entries back into a map.

## Basic examples

Loop over a map and destructure each entry:

```ant
var ages = Map{ ann: 31, bob: 27 }
for { key, value } in ages {
  println("${key} is ${value}")
}
```

```text
ann is 31
bob is 27
```

Get the entries as an array, and read their fields:

```ant
var es = Map{ red: "#f00", green: "#0f0" }.entries()
es[1]       // Entry{ key: "green", value: "#0f0" }
es[1].key   // "green"
es[1].value // "#0f0"
```

Write one with a literal. The checker infers `K` from `key` and `V` from `value`:

```ant
var e = Entry{ key: "x", value: 1.5 }
e // Entry{ key: "x", value: 1.5 }
```

## Scope

Bind the whole entry in a loop instead of destructuring it:

```ant
var stock = Map{ apples: 3, pears: 0 }
for e in stock {
  println(e)
}
```

```text
Entry{ key: "apples", value: 3 }
Entry{ key: "pears", value: 0 }
```

Rename the fields as you destructure them, in a loop or in a `var`:

```ant
var ages = Map{ ann: 31 }
for { key: name, value: age } in ages {
  println("${name}: ${age}")
}
var { key: code, value: text } = Entry{ key: 404, value: "Not Found" }
println(code)
println(text)
```

```text
ann: 31
404
Not Found
```

Annotate a variable with both type arguments. The key and the value may be any type, `null`,
records, arrays, and unions included:

```ant
type Point = { x: int, y: int }
var n: Entry[string, int | null] = Entry{ key: "n", value: null }
n                                                // Entry{ key: "n", value: null }
Entry{ key: Point{ x: 1, y: 2 }, value: [1, 2] } // Entry{ key: Point{ x: 1, y: 2 }, value: [1, 2] }
var r: Entry[int, Result[int, string]] = Entry{ key: 1, value: Result.err("no") }
r.value // Err{ error: "no" }
```

`Map.map` and `Map.filter` pass an `Entry` to their callback, so `x.key` and `x.value` read it:

```ant
var prices = Map{ tea: 3, cake: 5 }
Map.map(prices, { x.value * 2 })       // Map{ "tea": 6, "cake": 10 }
Map.filter(prices, { x.key != "tea" }) // Map{ "cake": 5 }
```

A named function that takes an `Entry` works as the callback:

```ant
func label(e: Entry[string, int]): string {
  "${e.key} costs ${e.value}"
}

var prices = Map{ tea: 3, cake: 5 }
Map.map(prices, label) // Map{ "tea": "tea costs 3", "cake": "cake costs 5" }
```

Return an `Entry` from a generic function:

```ant
func pairOf[K, V](k: K, v: V): Entry[K, V] {
  Entry{ key: k, value: v }
}

println(pairOf("a", true))
println(pairOf(1, [2, 3]))
```

```text
Entry{ key: "a", value: true }
Entry{ key: 1, value: [2, 3] }
```

Test a union member with `is`, written without the type arguments:

```ant
func describe(x: Entry[string, int] | string): string {
  if x is Entry { "${x.key}=${x.value}" } else { x }
}

println(describe(Entry{ key: "port", value: 80 }))
println(describe("none"))
```

```text
port=80
none
```

Copy one with a spread and change a field:

```ant
var e = Entry{ key: "x", value: 1 }
Entry{ ...e, value: 5 } // Entry{ key: "x", value: 5 }
```

`String.from` and `Json.stringify` show both fields:

```ant
var e = Entry{ key: "x", value: 1 }
String.from(e)    // "Entry{ key: \"x\", value: 1 }"
Json.stringify(e) // "{\"key\":\"x\",\"value\":1}"
```

## Applications

Build a map from a list of pairs with `Map.fromEntries`. A later entry with the same key wins:

```ant
var es = [Entry{ key: "a", value: 1 }, Entry{ key: "b", value: 2 }, Entry{ key: "a", value: 3 }]
Map.fromEntries(es) // Map{ "a": 3, "b": 2 }
```

Swap the keys and the values of a map:

```ant
var codes = Map{ ok: 200, missing: 404 }
Map.fromEntries(codes.entries().map({ Entry{ key: x.value, value: x.key } })) // Map{ 200: "ok", 404: "missing" }
```

Index records by a field:

```ant
type User = { id: int, name: string }
var users = [User{ id: 7, name: "Ann" }, User{ id: 9, name: "Bob" }]
var byId = Map.fromEntries(users.map({ Entry{ key: x.id, value: x } }))
byId[9]?.name // "Bob"
```

Find the key with the largest value. `maxBy` gives `null` for an empty map:

```ant
var votes = Map{ tea: 3, coffee: 7, water: 5 }
votes.entries().maxBy({ x.value })?.key // "coffee"
```

## Properties and relations

Entries compare by their fields, so an entry equals a literal with the same key and value:

```ant
assert(Map{ a: 1 }.entries()[0] == Entry{ key: "a", value: 1 })
assert(Entry{ key: "a", value: [1, 2] } == Entry{ key: "a", value: [1, 2] })
assert(Entry{ key: "a", value: 1 } != Entry{ key: "b", value: 1 })
```

Every entry's `value` is what the map stores under its `key`:

```ant
var m = Map{ a: 1, b: 2, c: 3 }
for { key, value } in m {
  assert(m[key] == value)
}
```

`Map.fromEntries` and `Map.entries` undo each other:

```ant
var m = Map{ x: 1, y: 2 }
assert(Map.fromEntries(Map.entries(m)) == m)
```

## Possible issues

A literal gives both fields. Leaving one out is a compile error:

```ant
var e = Entry{ key: 0 }
// error: Entry literal is missing the field value
// error: cannot infer the type parameter V of Entry; annotate the result
```

The type arguments go in an annotation, never in a literal:

```ant
var e = Entry[string, int]{ key: "a", value: 1 } // error: expected ']', got ','
```

`Entry` without type arguments is not a type:

```ant
var e: Entry = Entry{ key: "a", value: 1 } // error: Entry takes 2 type arguments, got 0
```

Type arguments are invariant, so an `Entry[string, int]` does not fit where an
`Entry[string, int | null]` is expected:

```ant
var e = Entry{ key: "a", value: 1 }
var f: Entry[string, int | null] = e
// error: cannot use Entry[string, int] where Entry[string, int | null] is expected
```

`Entry` is not `Indexed`. A map's entries have `key` and `value`, and an `Entry` has no `index`:

```ant
var ages = Map{ ann: 31 }
for { index, value } in ages { // error: Entry[string, int] has no field index
  println("${index}: ${value}")
}
```

Write `m[e.key] = ...` to change a value while looping:

```ant
var m = Map{ a: 1, b: 2 }
for e in m {
  m[e.key] = e.value * 10
}
m // Map{ "a": 10, "b": 20 }
```

An entry holds the value the map stores, not a copy, so changing a stored array through the entry
changes the map:

```ant
var lists = Map{ a: [1] }
for { value } in lists {
  value.push(2)
}
lists // Map{ "a": [1, 2] }
```

## See also

[Map.entries](../entries.md) · [Map.fromEntries](../fromEntries.md) · [Map.map](../map.md) ·
[Map.filter](../filter.md) · [Indexed](../../Array/types/Indexed.md) ·
[Array.entries](../../Array/entries.md) · [for](../../../language/keywords/for.md)
