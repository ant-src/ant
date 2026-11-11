<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.get

```ant signature
func Map.get[K, V](m: Map[K, V], k: K): V | null
```

Map.get returns the value m holds for k, or null when k is missing. It is the same as m\[k\]; use
[`Map.has`](has.md) when V itself includes null.

<!-- end generated -->

## Details

- It never panics and never adds an entry.
- Keys compare with `==`, so a record, an array, or a union value works as a key.
- A record, array, map, or set it returns is the value stored in `m`, not a copy.

## Basic examples

Read the value for a key:

```ant
let ages = Map{ ada: 36, linus: 54 }
Map.get(ages, "ada") // 36
```

A missing key gives `null`:

```ant
let ages = Map{ ada: 36, linus: 54 }
Map.get(ages, "grace") // null
```

Give a missing key a default with `??`:

```ant
let ages = Map{ ada: 36, linus: 54 }
ages.get("grace") ?? 0 // 0
```

## Scope

An empty map has no entries. Annotate it so that `K` and `V` are known:

```ant
let empty: Map[string, int] = Map{}
empty.get("a") // null
```

Records and arrays are keys by value, so a new record equal to the stored key finds its entry:

```ant
type Point = { x: int, y: int }
let labels = Map{ Point{ x: 0, y: 0 }: "origin" }
labels.get(Point{ x: 0, y: 0 }) // "origin"
labels.get(Point{ x: 1, y: 0 }) // null
let cells = Map{ [0, 1]: "b1" }
cells.get([0, 1]) // "b1"
```

Keys of a union type tell their members apart, so `1` and `"1"` are two keys:

```ant
type Id = int | string
let ids: Map[Id, string] = Map{ 1: "number", "1": "text" }
ids.get(1)   // "number"
ids.get("1") // "text"
```

Look up a key of a named primitive type with a value of that type:

```ant
type UserId = string
let owners = Map{ UserId("u-1"): "Ann" }
owners.get(UserId("u-1")) // "Ann"
```

String keys compare by code points. A precomposed `é` and an `e` followed by a combining accent look
the same but are two keys:

```ant
let words = Map{ "café": 1 }
words.get("café")        // 1
words.get("cafe\u{301}") // null
```

Call it through `.` or in a pipe:

```ant
let ages = Map{ ada: 36, linus: 54 }
ages.get("linus")     // 54
ages | Map.get("ada") // 36
```

## Applications

Count words, starting each count at 0 with `??`:

```ant
let counts: Map[string, int] = Map{}
for word in "to be or not to be".split(" ") {
  counts[word] = (counts.get(word) ?? 0) + 1
}
println(counts)
```

```text
Map{ "to": 2, "be": 2, "or": 1, "not": 1 }
```

Return early from a function when a key is missing, with `?? return`:

```ant
type User = { name: string, email: string }
let users = Map{ "u-1": User{ name: "Ann", email: "ann@example.com" } }

func emailOf(id: string): string {
  let user = users.get(id) ?? return "unknown user ${id}"
  user.email
}

println(emailOf("u-1"))
println(emailOf("u-9"))
```

```text
ann@example.com
unknown user u-9
```

Remember results that are slow to compute:

```ant
let cache: Map[int, int] = Map{}

func fib(n: int): int {
  if n < 2 {
    return n
  }
  let known = cache.get(n)
  if known != null {
    return known
  }
  let result = fib(n - 1) + fib(n - 2)
  cache[n] = result
  result
}

println(fib(80))
```

```text
23416728348467685
```

Group records by a field, adding to the array that `Map.get` returns:

```ant
type Task = { owner: string, title: string }
let tasks = [
  Task{ owner: "ann", title: "docs" },
  Task{ owner: "bob", title: "tests" },
  Task{ owner: "ann", title: "release" },
]
let byOwner: Map[string, Array[string]] = Map{}
for task in tasks {
  let titles = byOwner.get(task.owner)
  if titles == null {
    byOwner[task.owner] = [task.title]
  } else {
    titles.push(task.title)
  }
}
println(byOwner)
```

```text
Map{ "ann": ["docs", "release"], "bob": ["tests"] }
```

## Properties and relations

`Map.get` is the same as indexing with `m[k]`:

```ant
let ages = Map{ ada: 36 }
assert(ages.get("ada") == ages["ada"])
assert(ages.get("grace") == ages["grace"])
```

After `Map.set`, it returns the value just stored; after `Map.remove`, it returns `null`:

```ant
let ages = Map{ ada: 36 }
ages.set("ada", 37)
assert(ages.get("ada") == 37)
ages.remove("ada")
assert(ages.get("ada") == null)
```

When `V` has no `null`, it returns `null` exactly when `Map.has` is false:

```ant
let ages = Map{ ada: 36 }
assert((ages.get("ada") != null) == ages.has("ada"))
assert((ages.get("grace") != null) == ages.has("grace"))
```

It reads without changing the map, so a missing key is not added:

```ant
let ages = Map{ ada: 36 }
ages.get("grace")
assert(ages.size() == 1)
```

## Possible issues

The result may be `null`, so using it as a `V` without narrowing is a compile error:

```ant
let ages = Map{ ada: 36 }
println(ages.get("ada") + 1)
// error: + needs both operands of one type, got int | null and int
```

Use `??` or a test against `null` first, as in the examples above.

When `V` includes `null`, a stored `null` and a missing key look the same. Use [`Map.has`](has.md)
to tell them apart:

```ant
let scores: Map[string, int | null] = Map{ ann: null }
scores.get("ann") // null
scores.get("bob") // null
scores.has("ann") // true
scores.has("bob") // false
```

The key must have the map's key type:

```ant
let ages = Map{ ada: 36 }
println(ages.get(1)) // error: cannot use int where string is expected
```

A key of a named primitive type needs that type; a plain string does not convert:

```ant
type UserId = string
let owners = Map{ UserId("u-1"): "Ann" }
println(owners.get("u-1")) // error: cannot use string where UserId is expected
```

A key mutated after it was inserted may no longer find its entry. Do not change a record or an array
while it is a key; remove the entry, change the key, and insert it again.

## See also

[Map.has](has.md) · [Map.set](set.md) · [Map.remove](remove.md) · [Map.keys](keys.md) ·
[Map.values](values.md) · [Array.find](../Array/find.md) ·
[indexing](../../language/operators/indexing.md) ·
[?? operator](../../language/operators/null-coalescing.md) ·
[?. operator](../../language/operators/optional-chaining.md)
