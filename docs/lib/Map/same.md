<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Map.same

```ant signature
func Map.same[K, V](m: Map[K, V], n: Map[K, V]): bool
```

Map.same reports whether m and n are the same map, not merely equal ones.

<!-- end generated -->

## Details

- `==` compares keys and values, so use `Map.same` to tell a shared map from a copy.

## Basic examples

A map is the same as itself, and as any other name for it:

```ant
var m = Map{ a: 1 }
var n = m
Map.same(m, m) // true
Map.same(m, n) // true
```

An equal map built separately is not the same map:

```ant
var m = Map{ a: 1 }
m == Map{ a: 1 }         // true
Map.same(m, Map{ a: 1 }) // false
```

Call it through `.` on any map:

```ant
var m = Map{ a: 1 }
var n = m
m.same(n) // true
```

## Scope

Two empty maps are equal, and still two maps:

```ant
var a: Map[string, int] = Map{}
var b: Map[string, int] = Map{}
a == b         // true
Map.same(a, b) // false
```

`Map.filter`, `Map.map`, `Map.merge`, and `Map.fromEntries` build a new map, even when nothing
changes:

```ant
var m = Map{ a: 1, b: 2 }
Map.same(m, m.filter({ true }))           // false
Map.same(m, m.merge(Map{}))               // false
Map.same(m, Map.fromEntries(m.entries())) // false
```

A map held in a record field is shared by a spread of the record, since the spread copies the field,
which is a reference. A field default builds a new map at every literal:

```ant
type Config = { host: string, headers: Map[string, string] = Map{} }
var c = Config{ host: "a.example" }
var d = Config{ ...c, host: "b.example" }
var e = Config{ host: "a.example" }
Map.same(c.headers, d.headers) // true
Map.same(c.headers, e.headers) // false
```

A map passed to a function is the same map inside it, so the function can change it:

```ant
func count(m: Map[string, int], word: string): Map[string, int] {
  m[word] = (m[word] ?? 0) + 1
  m
}
var words: Map[string, int] = Map{}
Map.same(count(words, "ant"), words) // true
words                                // Map{ "ant": 1 }
```

The maps may have any key and value types, such as records as keys, or arrays as values:

```ant
type Point = { x: int, y: int }
var names = Map{ (Point{ x: 0, y: 0 }): "origin" }
var labels = names
Map.same(names, labels) // true
var groups = Map{ odd: [1, 3], even: [2] }
Map.same(groups, groups) // true
```

In a pipeline, the left side is `m`:

```ant
var m = Map{ a: 1 }
var n = m
m | Map.same(n) // true
```

Name the arguments, in any order:

```ant
var m = Map{ a: 1 }
Map.same(m: m, n: m)           // true
Map.same(n: m, m: Map{ a: 1 }) // false
```

## Applications

Guard a function that moves every entry of one map into another against being given one map twice:

```ant
func moveAll(from: Map[string, int], to: Map[string, int]) {
  if Map.same(from, to) {
    return
  }
  for e in from {
    to[e.key] = e.value
  }
  for k in from.keys() {
    from.remove(k)
  }
}

var pending = Map{ a: 1, b: 2 }
var stored = Map{ z: 0 }
moveAll(pending, stored)
println("${pending} ${stored}")
moveAll(stored, stored)
println(stored)
```

```text
Map{} Map{ "z": 0, "a": 1, "b": 2 }
Map{ "z": 0, "a": 1, "b": 2 }
```

## Properties and relations

`Map.same` is reflexive and symmetric:

```ant
var m = Map{ a: 1 }
var n = m
assert(Map.same(m, m))
assert(Map.same(m, n) == Map.same(n, m))
```

The same map is `==` to itself unless it holds a `NaN`, since floats compare with IEEE 754 rules:

```ant
var m = Map{ x: Math.NaN }
m == m         // false
Map.same(m, m) // true
```

When two names refer to the same map, a change through one shows through the other. For a copy it
does not:

```ant
var m = Map{ a: 1 }
var n = m
var copy = Map.fromEntries(m.entries())
m["b"] = 2
assert(Map.same(m, n) && n == Map{ a: 1, b: 2 })
assert(!Map.same(m, copy) && copy == Map{ a: 1 })
```

`Reflect.same` reports the same for two reflected maps:

```ant
var m = Map{ a: 1 }
var t = type(Map[string, int])
assert(Reflect.same(Reflect.valueOf(t, m), Reflect.valueOf(t, m)))
assert(!Reflect.same(Reflect.valueOf(t, m), Reflect.valueOf(t, Map{ a: 1 })))
```

## Possible issues

`==` compares contents, so it cannot tell whether two names share a map. Use `Map.same` when that is
the question:

```ant
var m = Map{ a: 1 }
var copy = m.merge(Map{})
m == copy         // true
Map.same(m, copy) // false
```

Both arguments have the one type `Map[K, V]`. A `Map[string, int]` and a `Map[string, int | null]`
are different types, so comparing them is a compile error:

```ant
var m = Map{ a: 1 }
var n: Map[string, int | null] = Map{ a: 1 }
println(Map.same(m, n))
// error: cannot use Map[string, int | null] where Map[string, int] is expected
```

A nullable map, such as a value read from another map, must be narrowed first:

```ant
var m = Map{ a: 1 }
var nested = Map{ inner: m }
println(Map.same(nested["inner"], m))
// error: cannot use Map[string, int] | null where Map[K, V] is expected
```

```ant
var m = Map{ a: 1 }
var nested = Map{ inner: m }
var found = nested["inner"]
found != null && Map.same(found, m) // true
```

## See also

[Array.same](../Array/same.md) · [Record.same](../Record/same.md) · [Set.same](../Set/same.md) ·
[Reflect.same](../Reflect/same.md) · [Map.merge](merge.md) · [Map.fromEntries](fromEntries.md) ·
[== operator](../../language/operators/equality.md)
