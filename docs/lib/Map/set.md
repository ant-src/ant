<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.set

```ant signature
func Map.set[K, V](m: Map[K, V], k: K, v: V): void
```

Map.set stores v under k in m, replacing any value already there. It mutates m and is the same as
m\[k\] = v. A replaced entry keeps its position; a new one goes to the end.

<!-- end generated -->

## Details

- Every other name for the same map, including a caller that passed it, sees the new entry.
- Keys compare with `==`, so a record or an array works as a key. Do not mutate a key after you set
  it: the entry may become unreachable.

## Basic examples

Add an entry to a map:

```ant
let ages = Map{ ada: 36 }
Map.set(ages, "linus", 54)
ages // Map{ "ada": 36, "linus": 54 }
```

Replace the value of a key that is already there:

```ant
let ages = Map{ ada: 36, linus: 54 }
Map.set(ages, "ada", 37)
ages // Map{ "ada": 37, "linus": 54 }
```

## Scope

Start from an empty map, which needs its type spelled out:

```ant
let counts: Map[string, int] = Map{}
Map.set(counts, "apples", 3)
counts // Map{ "apples": 3 }
```

A replaced entry keeps its place; a key removed and set again goes to the end:

```ant
let order = Map{ a: 1, b: 2, c: 3 }
Map.set(order, "a", 10)
Map.keys(order)        // ["a", "b", "c"]
Map.remove(order, "a") // 10
Map.set(order, "a", 11)
Map.keys(order) // ["b", "c", "a"]
```

Records, arrays, and numbers work as keys. Two keys that are `==` are the same key:

```ant
type Point = { x: int, y: int }
let tiles: Map[Point, string] = Map{}
Map.set(tiles, Point{ x: 1, y: 2 }, "tree")
Map.set(tiles, Point{ x: 1, y: 2 }, "rock")
tiles // Map{ Point{ x: 1, y: 2 }: "rock" }
let paths: Map[Array[int], string] = Map{}
Map.set(paths, [0, 1], "left")
paths[[0, 1]] // "left"
let names: Map[int, string] = Map{}
Map.set(names, 7, "seven")
names // Map{ 7: "seven" }
```

When the value type includes `null`, a stored `null` is an entry like any other, and
[Map.has](has.md) tells it from a missing key:

```ant
let parents: Map[string, string | null] = Map{}
Map.set(parents, "root", null)
parents                     // Map{ "root": null }
Map.has(parents, "root")    // true
Map.has(parents, "missing") // false
```

A `let` fixes the binding, not the map, so a `let` map takes new entries:

```ant
let config = Map{ debug: "off" }
Map.set(config, "debug", "on")
config // Map{ "debug": "on" }
```

Call it through `.` on a map, or with a pipe, which passes its left side as the first argument:

```ant
let m: Map[string, int] = Map{}
m.set("a", 1)
m | Map.set("b", 2)
m // Map{ "a": 1, "b": 2 }
```

## Applications

Count the words of a text:

```ant
let counts: Map[string, int] = Map{}
for word in String.split("the cat saw the dog", " ") {
  Map.set(counts, word, (counts[word] ?? 0) + 1)
}
println(counts)
```

```text
Map{ "the": 2, "cat": 1, "saw": 1, "dog": 1 }
```

Remember results a recursive function has already computed:

```ant
let cache: Map[int, int] = Map{}

func fib(n: int): int {
  if n < 2 { return n }
  let hit = cache[n]
  if hit != null { return hit }
  let v = fib(n - 1) + fib(n - 2)
  Map.set(cache, n, v)
  v
}

println(fib(80))
println(Map.size(cache))
```

```text
23416728348467685
79
```

Index records by a field, the last record with a key winning:

```ant
type User = { id: string, name: string }

let users = [
  User{ id: "u1", name: "Ada" },
  User{ id: "u2", name: "Linus" },
  User{ id: "u1", name: "Ada L." },
]
let byId: Map[string, User] = Map{}
for user in users {
  Map.set(byId, user.id, user)
}
println(Map.size(byId))
println(byId["u1"]?.name)
```

```text
2
Ada L.
```

## Properties and relations

`Map.set(m, k, v)` is the same as the index assignment `m[k] = v`:

```ant
let a = Map{ x: 1 }
let b = Map{ x: 1 }
Map.set(a, "y", 2)
b["y"] = 2
assert(a == b)
```

After the call, [Map.get](get.md) returns `v` for `k`:

```ant
let m = Map{ a: 1 }
Map.set(m, "b", 2)
assert(Map.get(m, "b") == 2)
```

The size grows by one for a new key and stays the same for a key already there:

```ant
let m = Map{ a: 1 }
Map.set(m, "b", 2)
assert(Map.size(m) == 2)
Map.set(m, "a", 3)
assert(Map.size(m) == 2)
```

`m` stays the same map, so every other name for it sees the new entry, including the caller of a
function that sets it:

```ant
func touch(m: Map[string, int]) {
  Map.set(m, "touched", 1)
}

let m = Map{ a: 1 }
let alias = m
touch(m)
assert(Map.same(alias, m))
assert(Map.has(alias, "touched"))
```

## Possible issues

`Map.set` returns `void`, not the map. To build a new map and keep `m`, use [Map.merge](merge.md):

```ant
func withDefault(m: Map[string, int]): Map[string, int] {
  Map.set(m, "default", 0) // error: cannot use void where Map[string, int] is expected
}
```

```ant
let base = Map{ a: 1 }
let more = Map.merge(base, Map{ default: 0 })
base // Map{ "a": 1 }
more // Map{ "a": 1, "default": 0 }
```

The key and the value must have the map's types:

```ant
let ages = Map{ ada: 36 }
Map.set(ages, 7, 54)          // error: cannot use int where string is expected
Map.set(ages, "linus", "old") // error: cannot use string where int is expected
```

A `for` loop over a map visits entries set during the loop, so adding a new key for each entry never
ends. Here a size check stops it:

```ant
let squares = Map{ 1: 1, 2: 4 }
for e in squares {
  if Map.size(squares) >= 6 { break }
  Map.set(squares, e.key * 10, e.value)
}
squares // Map{ 1: 1, 2: 4, 10: 1, 20: 4, 100: 1, 200: 4 }
```

Replacing the value of the entry the loop is on adds no entry, so this loop visits each entry once:

```ant
let prices = Map{ tea: 2, cake: 5 }
for { key, value } in prices {
  Map.set(prices, key, value * 2)
}
prices // Map{ "tea": 4, "cake": 10 }
```

A key is hashed when it is set. If you go on changing an array or record you used as a key, set a
copy instead:

```ant
let path = [0, 1]
let seen: Map[Array[int], bool] = Map{}
Map.set(seen, [...path], true)
path[0] = 9
seen[[0, 1]] // true
```

## See also

[Map.get](get.md), [Map.has](has.md), [Map.remove](remove.md), [Map.merge](merge.md),
[Map.fromEntries](fromEntries.md), [Set.add](../Set/add.md)
