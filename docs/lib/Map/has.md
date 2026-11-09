<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.has

```ant signature
func Map.has[K, V](m: Map[K, V], k: K): bool
```

Map.has reports whether m has an entry for k. When V includes null, it is the only way to tell a
missing key from a stored null.

<!-- end generated -->

## Details

- Keys compare by [`==`](../../language/operators/equality.md), so a record, array, or map used as a
  key matches by contents.

## Basic examples

```ant
var ages = Map{ ada: 36, linus: 54 }
Map.has(ages, "ada")   // true
Map.has(ages, "grace") // false
```

Call it through `.` on any map:

```ant
var ages = Map{ ada: 36, linus: 54 }
ages.has("linus") // true
```

## Scope

An empty map has nothing. Annotate it so that `K` is known:

```ant
var empty: Map[string, int] = Map{}
empty.has("") // false
```

When `V` includes `null`, indexing gives `null` both for a stored `null` and for a missing key.
`Map.has` tells them apart:

```ant
var settings: Map[string, string | null] = Map{ theme: null }
settings["theme"]     // null
settings["font"]      // null
settings.has("theme") // true
settings.has("font")  // false
```

Keys of any type work. Numbers, records, and arrays compare by value:

```ant
type Point = { x: int, y: int }
var names = Map{ 1: "one", 2: "two" }
var grid = Map{ (Point{ x: 0, y: 0 }): "origin" }
var paths = Map{ ([1, 2]): "a/b" }
names.has(2)                  // true
grid.has(Point{ x: 0, y: 0 }) // true
grid.has(Point{ x: 1, y: 0 }) // false
paths.has([1, 2])             // true
```

A tagged union key matches by tag and fields, so equal fields under another tag do not match:

```ant
type Circle = { size: float64 }
type Square = { size: float64 }
type Shape = Circle | Square
var areas: Map[Shape, float64] = Map{ (Circle{ size: 1.0 }): 3.14 }
areas.has(Circle{ size: 1.0 }) // true
areas.has(Square{ size: 1.0 }) // false
```

A named primitive type compares its tag and its payload:

```ant
type UserId = string
var roles = Map{ (UserId("u-1")): "admin" }
roles.has(UserId("u-1")) // true
```

Strings compare by code points, without normalizing:

```ant
var words = Map{ "café": 1 }
words.has("café")        // true
words.has("cafe\u{301}") // false
```

On a map that may itself be `null`, `?.` gives `null`; `??` turns that into a `bool`:

```ant
var m: Map[string, int] | null = null
m?.has("a")          // null
m?.has("a") ?? false // false
```

In a pipeline, the map is the first argument:

```ant
Map{ a: 1 } | Map.has("a") // true
```

## Applications

Report each field of a JSON object as present, `null`, or missing:

```ant
var doc = Result.unwrap(Json.parse("{\"name\": \"ant\", \"license\": null}"))
if doc is Map {
  for key in ["name", "license", "homepage"] {
    if !doc.has(key) {
      println("${key}: missing")
    } else if doc[key] == null {
      println("${key}: null")
    } else {
      println("${key}: ${doc[key]}")
    }
  }
}
```

```text
name: ant
license: null
homepage: missing
```

Cache a lookup whose answer may be `null`, so that a `null` result is not looked up again:

```ant
var lookups = 0
var cache: Map[int, string | null] = Map{}

func nameOf(id: int): string | null {
  if cache.has(id) {
    return cache[id]
  }
  lookups += 1
  var name = if id == 1 { "ada" } else { null }
  cache[id] = name
  name
}

for id in [1, 2, 1, 2] {
  println(nameOf(id))
}
println("lookups: ${lookups}")
```

```text
ada
null
ada
null
lookups: 2
```

List the required settings a configuration lacks:

```ant
var config = Map{ host: "localhost", port: "8080" }
["host", "port", "user"].filter({ !config.has(x) }) // ["user"]
```

## Properties and relations

`Map.has` agrees with `Map.keys`, and with `Map.get` when `V` excludes `null`:

```ant
var m = Map{ a: 1, b: 2 }
for k in ["a", "b", "c"] {
  assert(m.has(k) == m.keys().contains(k))
  assert(m.has(k) == Set.of(m.keys()).has(k))
  assert(m.has(k) == (m.get(k) != null))
}
```

`Map.set` adds the key and `Map.remove` takes it away:

```ant
var m: Map[string, int] = Map{}
m.set("x", 1)
assert(m.has("x"))
m.remove("x")
assert(!m.has("x"))
assert(m.isEmpty())
```

## Possible issues

`Map.has` does not narrow the value that indexing reads. After the test, `m[k]` is still `V | null`:

```ant
var ages = Map{ ada: 36 }
if ages.has("ada") {
  var age: int = ages["ada"] // error: cannot use int | null where int is expected
  println(age)
}
```

When `V` excludes `null`, read the value once and test it instead:

```ant
var ages = Map{ ada: 36 }
var age = ages["ada"]
if age != null {
  println(age + 1)
}
```

```text
37
```

Testing `m[k] != null` instead of `Map.has` is wrong when `V` includes `null`, because it treats a
stored `null` as missing:

```ant
var settings: Map[string, string | null] = Map{ theme: null }
settings["theme"] != null // false
settings.has("theme")     // true
```

The key must have the map's key type:

```ant
var ages = Map{ ada: 36 }
println(ages.has(36)) // error: cannot use int where string is expected
```

`NaN` is not `==` to itself, so a `NaN` key is never found. `0.0` and `-0.0` are `==`, so they find
each other:

```ant
Map{ (Math.NaN): 1 }.has(Math.NaN) // false
Map{ (0.0): 1 }.has(-0.0)          // true
```

A key is hashed when it is inserted. Mutating an array or record after you use it as a key is a
program error, and `Map.has` may no longer find the entry. Insert a copy, or do not change the key.

## See also

[Map.get](get.md) · [Map.set](set.md) · [Map.remove](remove.md) · [Map.keys](keys.md) ·
[Set.has](../Set/has.md) · [Array.contains](../Array/contains.md) ·
[indexing](../../language/operators/indexing.md) ·
[?. operator](../../language/operators/optional-chaining.md)
