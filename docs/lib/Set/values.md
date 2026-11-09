<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.values

```ant signature
func Set.values[T](s: Set[T]): Array[T]
```

Set.values returns the members of s in insertion order.

<!-- end generated -->

## Details

- Members are in the order they were first added. Adding a member that is already there keeps its
  place; a member removed and added again moves to the end.
- Returns a new array. Changing it leaves `s` unchanged, and changing `s` afterwards leaves the
  array unchanged.

## Basic examples

```ant
let tags = Set.of(["red", "blue", "red", "green"])
Set.values(tags) // ["red", "blue", "green"]
```

Call it through `.` on the set:

```ant
Set.of([3, 1, 2]).values() // [3, 1, 2]
```

## Scope

An empty set gives an empty array. Annotate the set so that its member type is known:

```ant
let empty: Set[int] = Set.of([])
empty.values() // []
```

Adding a member that is already there keeps it where it was. Removing it and adding it again moves
it to the end:

```ant
let s = Set.of(["a", "b", "c"])
Set.add(s, "a")    // false
s.values()         // ["a", "b", "c"]
Set.remove(s, "a") // true
Set.add(s, "a")    // true
s.values()         // ["b", "c", "a"]
```

Records and arrays compare by `==`, so equal ones are one member:

```ant
type Point = { x: int, y: int }
Set.of([Point{ x: 1, y: 2 }, Point{ x: 1, y: 2 }]).values() // [Point{ x: 1, y: 2 }]
Set.of([[1], [1], []]).values()                             // [[1], []]
```

A set whose member type includes `null` can hold `null`:

```ant
let s: Set[string | null] = Set.of([null, "a", null])
s.values() // [null, "a"]
```

Members may be any string:

```ant
Set.of(["🐜", "é", ""]).values() // ["🐜", "é", ""]
```

In a pipeline, the set is the first argument:

```ant
Set.of(["b", "a"]) | Set.values() // ["b", "a"]
```

## Applications

Remove duplicates from an array and keep the first of each, in order:

```ant
let visits = ["home", "about", "home", "blog", "about"]
Set.of(visits).values() // ["home", "about", "blog"]
```

Print the members in sorted order. The set keeps insertion order, so sort the array:

```ant
let langs = Set.of(["go", "ant", "c"])
for lang in langs.values().sort() {
  println(lang)
}
```

```text
ant
c
go
```

Join the members into one line:

```ant
let flags = Set.of(["-v", "-x", "-v"])
String.join(flags.values(), " ") // "-v -x"
```

Use the members as a lazy sequence:

```ant
let s = Set.of([1, 2, 3, 4])
Iter.fromArray(s.values()) | Iter.map({ x * 10 }) | Iter.take(2) | Iter.toArray() // [10, 20]
```

## Properties and relations

There is one element per member, each distinct, and each one is in the set:

```ant
let s = Set.of(["a", "b", "c"])
let vs = s.values()
assert(len(vs) == s.size())
assert(vs == vs.unique())
assert(vs.every({ s.has(x) }))
```

`Set.of` of the values gives back an equal set, and the values of `Set.of(xs)` are
`Array.unique(xs)`:

```ant
let s = Set.of([4, 2, 4, 1])
assert(Set.of(s.values()) == s)
let xs = [3, 1, 3, 2, 1]
assert(Set.of(xs).values() == xs.unique())
```

The array is a snapshot. Pushing to it leaves the set alone, and adding to the set leaves the array
alone:

```ant
let s = Set.of(["a"])
let vs = s.values()
vs.push("z")
s.add("b")
assert(vs == ["a", "z"])
assert(s.values() == ["a", "b"])
```

`for` over a set visits the same members in the same order, so a loop can iterate `s` directly:

```ant
let s = Set.of([5, 3, 8])
let seen: Array[int] = []
for x in s {
  seen.push(x)
}
assert(seen == s.values())
```

## Possible issues

Two sets with the same members are `==` in any order, but their values are arrays, and arrays
compare in order:

```ant
let a = Set.of([1, 2])
let b = Set.of([2, 1])
a == b                                 // true
a.values() == b.values()               // false
a.values().sort() == b.values().sort() // true
```

A loop over the set itself reads the set as it goes, so a member added during the loop is visited. A
loop over `s.values()` sees only the members the set had when it was called:

```ant
let s = Set.of([1])
for x in s {
  if x < 3 {
    s.add(x + 1)
  }
}
s.values() // [1, 2, 3]
let t = Set.of([1])
for x in t.values() {
  if x < 3 {
    t.add(x + 1)
  }
}
t.values() // [1, 2]
```

`NaN` is not `==` to itself, so each `NaN` is a separate member:

```ant
Set.of([1.0, Math.NaN, Math.NaN]).values() // [1.0, NaN, NaN]
```

A set has no indexing. Take the values first, and read by position there:

```ant
let s = Set.of(["x", "y"])
println(s[0]) // error: Set[string] cannot be indexed
```

```ant
let s = Set.of(["x", "y"])
s.values()[0] // "x"
```

`Set.values` is a function, not a field. Without the parentheses it is a compile error:

```ant
let tags = Set.of(["x"])
println(tags.values) // error: Set[string] is not a record, so it has no field values
```

## See also

[Set.of](of.md) · [Set.size](size.md) · [Set.has](has.md) · [Set.add](add.md) ·
[Map.keys](../Map/keys.md) · [Array.unique](../Array/unique.md) · [Array.sort](../Array/sort.md) ·
[Iter.fromArray](../Iter/fromArray.md)
