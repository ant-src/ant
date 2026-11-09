<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.filter

```ant signature
func Map.filter[K, V](m: Map[K, V], f: func(Entry[K, V]): bool): Map[K, V]
```

Map.filter returns a new map holding the entries of m that f accepts, in the order of m.

<!-- end generated -->

## Details

- Returns a new map; `m` is unchanged, even when `f` accepts every entry.
- Calls `f` once per entry, in insertion order.

## Basic examples

Keep the entries whose value passes a test:

```ant
var stock = Map{ apples: 3, pears: 0, plums: 7 }
Map.filter(stock, { x.value > 0 }) // Map{ "apples": 3, "plums": 7 }
```

Keep the entries whose key passes a test:

```ant
var stock = Map{ apples: 3, pears: 0, plums: 7 }
Map.filter(stock, { String.startsWith(x.key, "p") }) // Map{ "pears": 0, "plums": 7 }
```

## Scope

An empty map gives an empty map, and so does a predicate that accepts nothing:

```ant
var empty: Map[string, int] = Map{}
Map.filter(empty, { true })              // Map{}
Map.filter(Map{ a: 1, b: 2 }, { false }) // Map{}
```

The result keeps the order of `m`, which is the order its keys were first inserted:

```ant
var m = Map{ b: 2, a: 1, c: 3 }
Map.filter(m, { x.key != "a" }) // Map{ "b": 2, "c": 3 }
```

Keys can be records:

```ant
type Point = { x: int, y: int }
var grid = Map{ Point{ x: 0, y: 0 }: "start", Point{ x: 2, y: 1 }: "tree", Point{ x: 0, y: 3 }: "rock" }
Map.filter(grid, { x.key.x == 0 }) // Map{ Point{ x: 0, y: 0 }: "start", Point{ x: 0, y: 3 }: "rock" }
```

Values can be a union; pick one member with `is`:

```ant
type Open = {}
type Closed = { reason: string }
type State = Open | Closed
var doors: Map[string, State] = Map{ front: Open{}, back: Closed{ reason: "locked" } }
Map.filter(doors, { x.value is Closed }) // Map{ "back": Closed{ reason: "locked" } }
```

Drop the entries whose value is `null`:

```ant
var ages = Map{ "Ann": 31, "Bob": null, "Cid": 27 }
Map.filter(ages, { x.value != null }) // Map{ "Ann": 31, "Cid": 27 }
```

Pass a named function that takes an `Entry`:

```ant
func isPositive(e: Entry[string, int]): bool { e.value > 0 }
Map.filter(Map{ a: 1, b: -1, c: 5 }, isPositive) // Map{ "a": 1, "c": 5 }
```

An anonymous function can destructure the entry into its key and value:

```ant
var m = Map{ a: 1, b: 2, c: 3 }
Map.filter(m, func(e: Entry[string, int]): bool {
  var { key, value } = e
  value > 1 && key != "c"
}) // Map{ "b": 2 }
```

Call it through `.` or in a pipe:

```ant
var stock = Map{ apples: 3, pears: 0, plums: 7 }
stock.filter({ x.value < 5 })            // Map{ "apples": 3, "pears": 0 }
stock | Map.filter({ x.key == "plums" }) // Map{ "plums": 7 }
```

## Applications

Hide the keys named in a set before logging form data:

```ant
var hidden = Set.of(["password", "token"])
var form = Map{ user: "ann", password: "hunter2", token: "abc" }
form.filter({ !Set.has(hidden, x.key) }) // Map{ "user": "ann" }
```

Print the words of two or more letters that occur at least three times:

```ant
var counts = Map{ "the": 12, "ant": 3, "a": 9, "hill": 1 }
for e in Map.entries(counts.filter({ x.value >= 3 && String.length(x.key) > 1 })) {
  println("${e.key} ${e.value}")
}
```

```text
the 12
ant 3
```

Split a map in two with a predicate and its negation:

```ant
var scores = Map{ Ann: 91, Bob: 58, Cid: 74 }
var passed = scores.filter({ x.value >= 60 })
var failed = scores.filter({ x.value < 60 })
passed // Map{ "Ann": 91, "Cid": 74 }
failed // Map{ "Bob": 58 }
```

## Properties and relations

`Map.filter` gives the same map as filtering the entries and building a map from them:

```ant
var m = Map{ a: 1, b: 2, c: 3 }
var f: func(Entry[string, int]): bool = { x.value % 2 == 1 }
assert(Map.filter(m, f) == Map.fromEntries(Map.entries(m).filter(f)))
assert(Map.keys(Map.filter(m, f)) == ["a", "c"])
```

Filtering by `f` and by its negation splits `m` with nothing lost:

```ant
var m = Map{ a: 1, b: 2, c: 3, d: 4 }
var f: func(Entry[string, int]): bool = { x.value > 2 }
var yes = Map.filter(m, f)
var no = Map.filter(m, { !f(x) })
assert(Map.size(yes) + Map.size(no) == Map.size(m))
assert(Map.merge(yes, no) == m)
```

`f` runs once per entry, in insertion order:

```ant
var m = Map{ b: 2, a: 1, c: 3 }
var seen: Array[string] = []
var big = Map.filter(m, func(e: Entry[string, int]): bool {
  Array.push(seen, e.key)
  e.value > 1
})
assert(seen == ["b", "a", "c"])
assert(big == Map{ b: 2, c: 3 })
```

Filtering twice by the same predicate changes nothing the second time:

```ant
var m = Map{ a: 1, b: 2, c: 3 }
var once = Map.filter(m, { x.value != 2 })
assert(Map.filter(once, { x.value != 2 }) == once)
```

The result is a new map even when it holds every entry, so changing it leaves `m` as it was:

```ant
var m = Map{ a: 1, b: 2 }
var all = Map.filter(m, { true })
assert(all == m)
assert(!Map.same(all, m))
all["c"] = 3
assert(Map.size(m) == 2)
```

## Possible issues

`Map.filter` returns a new map and leaves `m` as it was. A call whose result is thrown away does not
change `m`; assign the result, or use `Map.remove` to change `m` in place:

```ant
var stock = Map{ apples: 3, pears: 0 }
Map.filter(stock, { x.value > 0 })
println(stock)
```

```text
Map{ "apples": 3, "pears": 0 }
```

The predicate takes one `Entry`, not a key and a value:

```ant
var stock = Map{ apples: 3, pears: 0 }
var inStock = Map.filter(stock, func(k: string, v: int): bool { v > 0 })
// error: cannot use func(string, int): bool where func(Entry[string, int]): bool is expected
```

The predicate must return a `bool`. Ant has no truthy values, so an `int` is a compile error:

```ant
var stock = Map{ apples: 3, pears: 0 }
var inStock = Map.filter(stock, { x.value }) // error: cannot use int where bool is expected
```

The result keeps the value type of `m`. Filtering out `null` gives a `Map[string, int | null]`, not
a `Map[string, int]`:

```ant
var ages = Map{ "Ann": 31, "Bob": null }
var kept = Map.filter(ages, { x.value != null })
var known: Map[string, int] = kept // error: cannot use Map[string, int | null] where Map[string, int] is expected
```

Build the map in a loop instead, where the test on `value` narrows it to `int`:

```ant
var ages = Map{ "Ann": 31, "Bob": null, "Cid": 27 }
var known: Map[string, int] = Map{}
for { key, value } in ages {
  if value != null { known[key] = value }
}
known // Map{ "Ann": 31, "Cid": 27 }
```

## See also

[Map.map](map.md), [Map.entries](entries.md), [Map.fromEntries](fromEntries.md),
[Map.remove](remove.md), [Map.merge](merge.md), [Array.filter](../Array/filter.md),
[Entry](types/Entry.md)
