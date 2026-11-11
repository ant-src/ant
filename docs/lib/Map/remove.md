<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.remove

```ant signature
func Map.remove[K, V](m: Map[K, V], k: K): V | null
```

Map.remove removes the entry for k from m and returns its value, or null when k is missing. It
mutates m.

<!-- end generated -->

## Details

- When `V` includes `null`, a result of `null` does not tell a removed `null` from a missing key.
  Ask [Map.has](has.md) first.
- A key removed and set again is a new entry at the end of the map's order.

## Basic examples

Remove an entry and get its value back:

```ant
let ages = Map{ ada: 36, linus: 54 }
Map.remove(ages, "ada") // 36
ages                    // Map{ "linus": 54 }
```

A missing key gives `null` and leaves the map as it was:

```ant
let ages = Map{ ada: 36 }
Map.remove(ages, "grace") // null
ages                      // Map{ "ada": 36 }
```

## Scope

An empty map has nothing to remove:

```ant
let empty: Map[string, int] = Map{}
empty.remove("a") // null
```

The other entries keep their order:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
m.remove("b") // 2
m             // Map{ "a": 1, "c": 3 }
```

Keys compare the way `==` does, so a fresh record with the same fields finds the entry:

```ant
type Point = { x: int, y: int }
let labels = Map{ (Point{ x: 1, y: 2 }): "home" }
labels.remove(Point{ x: 1, y: 2 }) // "home"
labels                             // Map{}
```

String keys compare by code points, with no normalization:

```ant
let m = Map{ "é": 1 }
m.remove("e\u{301}") // null
m.remove("é")        // 1
```

When `V` includes `null`, a stored `null` comes back as `null` too:

```ant
let settings: Map[string, string | null] = Map{ theme: null }
settings.remove("theme") // null
settings                 // Map{}
```

Call it through `.`, with a pipe, or with named arguments in any order:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
m.remove("a")            // 1
m | Map.remove("b")      // 2
Map.remove(k: "c", m: m) // 3
m                        // Map{}
```

## Applications

Take a value out with a fallback, using `??`:

```ant
let stock = Map{ apple: 3 }
stock.remove("apple") ?? 0 // 3
stock.remove("pear") ?? 0  // 0
```

Hand out one-time codes, each accepted once:

```ant
let codes = Map{ "x7k2": "ada", "p9q4": "linus" }

func redeem(code: string): string {
  let user = codes.remove(code) ?? return "invalid or used: ${code}"
  "welcome, ${user}"
}

println(redeem("x7k2"))
println(redeem("x7k2"))
println(redeem("zzzz"))
```

```text
welcome, ada
invalid or used: x7k2
invalid or used: zzzz
```

Move an entry to the end, to keep the most recently used key last:

```ant
func touch(m: Map[string, int], k: string) {
  let v = m.remove(k) ?? return
  m[k] = v
}

let recent = Map{ a: 1, b: 2, c: 3 }
touch(recent, "a")
println(recent.keys())
```

```text
["b", "c", "a"]
```

## Properties and relations

After the call, `m` has no entry for `k`, and the result is what [Map.get](get.md) gave before it:

```ant
let m = Map{ a: 1, b: 2 }
for k in ["a", "z"] {
  let before = m.get(k)
  assert(m.remove(k) == before)
  assert(!m.has(k))
}
```

The size drops by one exactly when the key was there:

```ant
let m = Map{ a: 1, b: 2 }
for k in ["a", "z"] {
  let had = m.has(k)
  let size = m.size()
  m.remove(k)
  assert(m.size() == if had { size - 1 } else { size })
}
```

It removes from the map itself, so every other name for it sees the change:

```ant
let m = Map{ a: 1 }
let alias = m
assert(m.remove("a") == 1)
assert(Map.same(alias, m))
alias // Map{}
```

## Possible issues

The result is `V | null`, not `V`, so it does not go where a `V` is expected:

```ant
let m = Map{ a: 1 }
let n: int = m.remove("a") // error: cannot use int | null where int is expected
```

Give a fallback with `??`, or narrow it first:

```ant
let m = Map{ a: 1 }
let n: int = m.remove("a") ?? 0
n // 1
```

The key must have the map's key type:

```ant
let m = Map{ a: 1 }
m.remove(1) // error: cannot use int where string is expected
```

Setting a key to `null` stores `null`; use Map.remove to drop the entry:

```ant
let m: Map[string, int | null] = Map{ a: 1 }
m["a"] = null
m.size() // 1
m        // Map{ "a": null }
```

`NaN` is not `==` to itself, so an entry keyed by `NaN` cannot be removed:

```ant
let m = Map{ (Math.NaN): 1 }
m.remove(Math.NaN) // null
m                  // Map{ NaN: 1 }
```

Removing entries during a `for` over the map is allowed. An entry removed before the loop reaches it
is not visited:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
let seen: Array[string] = []
for e in m {
  if e.key == "a" {
    m.remove("c")
  }
  seen.push(e.key)
}
seen // ["a", "b"]
```

To remove the entries that match a test, [Map.filter](filter.md) returns a new map instead:

```ant
let m = Map{ a: 1, b: 2, c: 3 }
m.filter({ x.value != 2 }) // Map{ "a": 1, "c": 3 }
```

## See also

[Map.set](set.md) · [Map.get](get.md) · [Map.has](has.md) · [Map.filter](filter.md) ·
[Map.size](size.md) · [Set.remove](../Set/remove.md)
