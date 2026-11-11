<!-- status: reviewed -->

# Set

`Set[T]` holds each value of type `T` at most once, compared with `==`, and keeps the members in the
order they were first added.

## Details

- A set has no literal and no indexing. `Set.of` builds one from an array; the empty set is
  `Set.of([])`, which takes its element type from the expected type.
- A set is a reference value: assigning or passing it shares one set, and a change through one name
  shows through every other. `Set.same` tells whether two names hold one set.
- Adding a member that is already there changes nothing, so it keeps its position. A member removed
  and added again goes to the end.
- `==` compares the members, in any order.
- A member is hashed when it is added. Changing a member, such as a record, array, map, or set,
  after adding it is a program error: the set may no longer find it.

## Basic examples

Build a set from an array; duplicates collapse to the first of each value:

```ant
let tags = Set.of(["red", "blue", "red"])
tags                   // Set.of(["red", "blue"])
tags.has("blue")       // true
tags.size()            // 2
Set.has(tags, "green") // false
```

A set grows in place. `add` and `remove` tell whether they changed it:

```ant
let seen: Set[int] = Set.of([])
seen.add(3)    // true
seen.add(3)    // false
seen.remove(3) // true
seen           // Set.of([])
```

## Scope

### Member types

A member can be a record, an array, a union, or `null`. Members compare by contents, so equal values
collapse into one:

```ant
type Point = { x: int, y: int }
let points = Set.of([Point{ x: 0, y: 0 }, Point{ x: 0, y: 0 }, Point{ x: 1, y: 2 }])
let pairs = Set.of([[1, 2], [2, 1], [1, 2]])
let maybe: Set[string | null] = Set.of(["a", null, null])
points.size()                   // 2
pairs                           // Set.of([[1, 2], [2, 1]])
maybe.has(null)                 // true
points.has(Point{ x: 1, y: 2 }) // true
```

### Looping

`for` visits the members in insertion order:

```ant
let langs = Set.of(["Ant", "Go", "Lua"])
for lang in langs {
  println(lang)
}
```

```text
Ant
Go
Lua
```

A member added during the loop is visited; a member removed before the loop reaches it is not:

```ant
let s = Set.of([1, 2, 3])
for x in s {
  if x == 1 {
    s.remove(2)
    s.add(4)
  }
  println(x)
}
```

```text
1
3
4
```

`Set.values` gives the members as a new array, in the same order:

```ant
let langs = Set.of(["Go", "Ant"])
langs.values()                // ["Go", "Ant"]
langs.values() | Array.sort() // ["Ant", "Go"]
```

### Size

`len` counts the members, as `Set.size` does:

```ant
let s = Set.of([1, 2, 2])
len(s)      // 2
s.isEmpty() // false
```

### Order

A repeated `add` keeps the member where it is; a removed member comes back at the end:

```ant
let s = Set.of(["a", "b", "c"])
s.add("a")    // false
s             // Set.of(["a", "b", "c"])
s.remove("a") // true
s.add("a")    // true
s             // Set.of(["b", "c", "a"])
```

### Combining sets

`union`, `intersect`, and `difference` build a new set and keep the order of their first argument.
`isSubset` tells whether every member of one set is in another:

```ant
let a = Set.of([1, 2, 3])
let b = Set.of([3, 4, 2])
a.union(b)              // Set.of([1, 2, 3, 4])
a.intersect(b)          // Set.of([2, 3])
a.difference(b)         // Set.of([1])
Set.of([2]).isSubset(a) // true
```

### Sets in unions

`is Set` tells a set apart from the other members of a union:

```ant
func describe(v: int | Set[int]): string {
  if v is Set { "${v.size()} members" } else { "one number" }
}

describe(Set.of([1, 2, 2])) // "2 members"
describe(7)                 // "one number"
```

### Generic functions

A type parameter keeps the member type, so one function serves every set:

```ant
func shared[T](a: Array[T], b: Array[T]): Set[T] {
  Set.of(a).intersect(Set.of(b))
}

shared([1, 2, 3], [3, 2])       // Set.of([2, 3])
shared(["ant"], ["bee"]).size() // 0
```

### Rendering

A set renders as the call that builds it, its strings quoted:

```ant
println(Set.of(["ant", "bee"]))
println(String.from(Set.of([1, 2])))
```

```text
Set.of(["ant", "bee"])
Set.of([1, 2])
```

## Applications

### Keep the first of each line

`add` returns `true` only for a new member, so it filters and records in one step:

```ant
let lines = ["b", "a", "b", "c", "a"]
let seen: Set[string] = Set.of([])
for line in lines {
  if seen.add(line) { println(line) }
}
```

```text
b
a
c
```

### Fast membership tests

Build a set once, then ask it many times:

```ant
let stopWords = Set.of(["a", "an", "the"])
"the cat and a hat".split(" ")
  | Array.filter({ !stopWords.has(x) })
  | String.join(" ") // "cat and hat"
```

### Remove duplicates

`Set.of` and `values` dedupe an array and keep its order:

```ant
Set.of([3, 1, 3, 2, 1]).values() // [3, 1, 2]
```

## Properties and relations

`==` compares members in any order; `Set.same` compares identity:

```ant
let s = Set.of([1, 2])
let alias = s
let copy = Set.of(s.values())
assert(s == Set.of([2, 1]))
assert(s == copy)
assert(Set.same(s, alias))
assert(!Set.same(s, copy))
assert(len(s) == s.size())
```

A union holds what each side holds, and a set is a subset of its union with another:

```ant
let a = Set.of([1, 2])
let b = Set.of([2, 3])
let both = a.union(b)
assert(both.values().every({ a.has(x) || b.has(x) }))
assert(a.isSubset(both) && b.isSubset(both))
assert(a.intersect(b).isSubset(a))
```

## Possible issues

### A second name shares the set

Binding a set to a new name does not copy it. Copy with `Set.of(s.values())` when you need a
separate set:

```ant
let original = Set.of([1, 2])
let shared = original
let copy = Set.of(original.values())
shared.add(3) // true
original      // Set.of([1, 2, 3])
copy          // Set.of([1, 2])
```

### No indexing

A set has no positions. Ask with `has`, or index its `values`:

```ant
let s = Set.of([10, 20])
let first = s[0] // error: Set[int] cannot be indexed
```

```ant
let s = Set.of([10, 20])
s.values()[0] // 10
```

### The empty set needs a type

`Set.of([])` takes its element type from where it goes. With nothing to say it, annotate:

```ant
let empty = Set.of([]) // error: cannot infer the element type of an empty array; annotate it
```

```ant
let empty: Set[int] = Set.of([])
empty.isEmpty() // true
```

### `Set{ }` is not a set

A set has no brace literal. `Set{ ... }` reads as a record literal:

```ant
let s = Set{ 1, 2 } // error: a record field must be a name
```

### No widening of the member type

A `Set[int]` is not a `Set[int | null]`. Copy into a set of the wider type:

```ant
let small: Set[int] = Set.of([1])
let wide: Set[int | null] = small // error: cannot use Set[int] where Set[int | null] is expected
```

```ant
let small: Set[int] = Set.of([1])
let wide: Set[int | null] = Set.of([...small.values()])
wide.add(null) // true
```

### Two set types in one union

`is Set` cannot tell two set types apart at run time, so a union holds only one. Put the union
inside:

```ant
let ids: Set[int] | Set[string] = Set.of([1])
// error: a union may contain each discriminator at most once: Set[int] and Set[string] are both a set; write Set[int | string]
```

## See also

[Set functions](../../lib/Set/index.md) · [Set literals](../literals/sets.md) ·
[Set.of](../../lib/Set/of.md) · [Set.same](../../lib/Set/same.md) · [for](../keywords/for.md) ·
[len](../../lib/prelude/len.md) · [Array](Array.md) · [Map](Map.md) ·
[Collections](../../guide/collections.md)
