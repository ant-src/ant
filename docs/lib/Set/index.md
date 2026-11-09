<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set

Set holds the functions of the built-in set type Set\[T\], a hash set keyed by == that keeps
insertion order. There is no literal: the empty set is Set.of(\[\]).

<!-- end generated -->

## Details

- Records, arrays, maps, and sets are members by contents.
- A set keeps insertion order. Adding a member that is already there keeps its position; a member
  removed and added again goes to the end.
- [Set.add](add.md) and [Set.remove](remove.md) change the set in place and report whether they
  changed it. Every other function returns a new value and leaves its arguments untouched.
- A set is a reference: copying a variable shares the set. Two sets are `==` when they have the same
  members, in any order.
- A member is hashed when it goes in. Changing an array or a record after it became a member is a
  program error: the set may no longer find it.

## Basic examples

`Set.of` builds a set from an array and drops the repeats. `Set.has` asks for a member:

```ant
var tags = Set.of(["red", "blue", "red"])
tags              // Set.of(["red", "blue"])
tags.has("blue")  // true
tags.has("green") // false
len(tags)         // 2
```

`Set.add` and `Set.remove` change the set and return whether they did:

```ant
var seen = Set.of([1, 2])
seen.add(3)    // true
seen.add(3)    // false
seen.remove(1) // true
seen           // Set.of([2, 3])
```

A `for` loop visits the members in insertion order:

```ant
var langs = Set.of(["go", "ant", "go", "zig"])
for lang in langs {
  println(lang)
}
```

```text
go
ant
zig
```

## Scope

There is no set literal. The empty set is `Set.of([])`, with its element type from an annotation or
from where it goes:

```ant
var empty: Set[string] = Set.of([])
empty           // Set.of([])
empty.isEmpty() // true
```

Records and arrays are members by contents:

```ant
type Point = { x: int, y: int }
var visited = Set.of([Point{ x: 0, y: 0 }])
visited.has(Point{ x: 0, y: 0 }) // true
Set.of([[1, 2], [1, 2], [2, 1]]) // Set.of([[1, 2], [2, 1]])
```

A set can hold `null` when its element type includes it:

```ant
var answers: Set[bool | null] = Set.of([true, null, null])
answers           // Set.of([true, null])
answers.has(null) // true
```

Calls chain through `.`, or the set goes down a pipe:

```ant
var a = Set.of([1, 2, 3, 4])
var b = Set.of([3, 4, 5])
a.intersect(b).values()            // [3, 4]
a | Set.difference(b) | Set.size() // 2
```

## Functions by job

### Building a set

- [Set.of](of.md): a new set holding the first of each value of an array, in order

### Size and membership

- [Set.size](size.md): the number of members; the same as `len(s)`
- [Set.isEmpty](isEmpty.md): whether there are none
- [Set.has](has.md): whether a value is a member

### Changing a set in place

- [Set.add](add.md): add a member; `true` when it was not already there
- [Set.remove](remove.md): remove a member; `true` when it was there

### Members as an array

- [Set.values](values.md): the members in insertion order

### Combining sets

- [Set.union](union.md): the members of `a`, then those of `b` that `a` lacks
- [Set.intersect](intersect.md): the members of `a` that `b` has
- [Set.difference](difference.md): the members of `a` that `b` lacks
- [Set.isSubset](isSubset.md): whether every member of `a` is a member of `b`

### Identity

- [Set.same](same.md): whether two sets are the same set, not only equal

## Applications

Print each line the first time it appears. `Set.add` answers the question and records the line in
one call:

```ant
var seen: Set[string] = Set.of([])
for line in ["b", "a", "b", "c", "a"] {
  if seen.add(line) {
    println(line)
  }
}
```

```text
b
a
c
```

Find the required settings a map lacks:

```ant
var required = Set.of(["host", "port", "user"])
var given = Map{ host: "localhost", user: "ada" }
Set.difference(required, Set.of(given.keys())).values() // ["port"]
```

Check that a word uses only allowed letters:

```ant
var allowed = Set.of(String.chars("abcdef"))
Set.of(String.chars("faced")).isSubset(allowed)  // true
Set.of(String.chars("faded!")).isSubset(allowed) // false
```

Copy a set, so that changing the copy leaves the original alone:

```ant
var a = Set.of([1, 2])
var c = Set.of(a.values())
c.add(3) // true
a        // Set.of([1, 2])
```

List the members in sorted order:

```ant
var tags = Set.of(["red", "blue", "green"])
tags.values().sort() // ["blue", "green", "red"]
```

## Properties and relations

`len`, `Set.size`, and `Set.values` agree:

```ant
var s = Set.of(["x", "y", "x"])
assert(len(s) == Set.size(s))
assert(len(Set.values(s)) == Set.size(s))
assert(Set.isEmpty(s) == (Set.size(s) == 0))
```

The members of `Set.of(xs)`, in order, are what `Array.unique` returns:

```ant
var xs = [4, 2, 4, 9, 2]
assert(Set.of(xs).values() == xs.unique())
```

`==` ignores order, while `Set.values` shows it:

```ant
var a = Set.of([1, 2, 3])
var b = Set.of([3, 2, 1])
assert(a == b)
assert(a.values() != b.values())
```

A union holds both sets, and an intersection lies inside each:

```ant
var a = Set.of([1, 2, 3])
var b = Set.of([2, 3, 4])
assert(a.isSubset(a.union(b)) && b.isSubset(a.union(b)))
assert(a.intersect(b).isSubset(a) && a.intersect(b).isSubset(b))
assert(a.difference(b).intersect(b).isEmpty())
```

## Possible issues

`Set.of([])` with nothing to give it a type is a compile error:

```ant
var seen = Set.of([]) // error: cannot infer the element type of an empty array; annotate it
```

A set has no indexing. Take the members as an array first:

```ant
var s = Set.of(["a", "b"])
s[0] // error: Set[string] cannot be indexed
```

```ant
var s = Set.of(["a", "b"])
s.values()[0] // "a"
```

Copying a set variable copies the reference. A change through one name shows through the other:

```ant
var a = Set.of([1])
var b = a
b.add(2)       // true
a              // Set.of([1, 2])
Set.same(a, b) // true
```

## See also

- [Array](../Array/index.md), [Map](../Map/index.md): the other built-in containers
- [Array.unique](../Array/unique.md): the distinct elements of an array, as an array
- [Map.keys](../Map/keys.md): the keys of a map, ready for `Set.of`
