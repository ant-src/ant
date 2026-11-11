<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.difference

```ant signature
func Set.difference[T](a: Set[T], b: Set[T]): Set[T]
```

Set.difference returns a new set holding the members of a that b lacks, in the order of a.

<!-- end generated -->

## Details

- Returns a new set. Neither argument changes.
- Members match by `==`, so records and arrays match by their contents. The order of `b` does not
  matter.

## Basic examples

Remove the members of one set from another:

```ant
Set.difference(Set.of([1, 2, 3, 4]), Set.of([2, 4])) // Set.of([1, 3])
```

Members of `b` that `a` lacks do not matter:

```ant
Set.difference(Set.of(["a", "b"]), Set.of(["b", "z"])) // Set.of(["a"])
```

The result keeps the insertion order of `a`:

```ant
Set.difference(Set.of([3, 1, 2]), Set.of([1])) // Set.of([3, 2])
```

## Scope

An empty set on either side:

```ant
let none: Set[int] = Set.of([])
Set.difference(none, Set.of([1]))    // Set.of([])
Set.difference(Set.of([1, 2]), none) // Set.of([1, 2])
```

A set minus itself is empty:

```ant
let s = Set.of([1, 2])
Set.difference(s, s) // Set.of([])
```

Members compare with `==`, so strings compare by their code points, and records and arrays by their
contents:

```ant
type Point = { x: int, y: int }
let points = Set.of([Point{ x: 1, y: 2 }, Point{ x: 0, y: 0 }])
Set.difference(Set.of(["é", "E", "日本"]), Set.of(["e", "日本"])) // Set.of(["é", "E"])
Set.difference(points, Set.of([Point{ x: 0, y: 0 }]))             // Set.of([Point{ x: 1, y: 2 }])
Set.difference(Set.of([[1], [2]]), Set.of([[2]]))                 // Set.of([[1]])
```

A union with `null` works alike:

```ant
Set.difference(Set.of([1, null, 2]), Set.of([null])) // Set.of([1, 2])
```

Call it through `.` on a set, or with a pipe, which passes its left side as the first argument:

```ant
Set.of([1, 2]).difference(Set.of([1]))       // Set.of([2])
Set.of([1, 2]) | Set.difference(Set.of([2])) // Set.of([1])
```

Name the arguments to pass them in any order:

```ant
Set.difference(b: Set.of([2]), a: Set.of([1, 2])) // Set.of([1])
```

## Applications

Find what was removed and what was added between two versions of a list:

```ant
let before = Set.of(["ann", "bob", "cy"])
let after = Set.of(["bob", "cy", "dee"])
println("removed: ${before.difference(after).values()}")
println("added: ${after.difference(before).values()}")
```

```text
removed: ["ann"]
added: ["dee"]
```

Report the required settings that a configuration does not set:

```ant
let required = Set.of(["host", "port", "user"])
let config = Map{ host: "localhost", user: "ann" }
for name in required.difference(Set.of(config.keys())) {
  println("missing: ${name}")
}
```

```text
missing: port
```

## Properties and relations

The result has no member of `b`, and every member of it is a member of `a`:

```ant
let a = Set.of([1, 2, 3])
let b = Set.of([2, 5])
let d = Set.difference(a, b)
assert(Set.intersect(d, b).isEmpty())
assert(d.isSubset(a))
```

The difference and the intersection split `a` in two:

```ant
let a = Set.of([1, 2, 3])
let b = Set.of([2, 5])
assert(Set.union(Set.difference(a, b), Set.intersect(a, b)) == a)
```

The difference is empty exactly when `a` is a subset of `b`:

```ant
let a = Set.of([1, 2])
let b = Set.of([1, 2, 3])
assert(Set.difference(a, b).isEmpty() == a.isSubset(b))
assert(!Set.difference(b, a).isEmpty())
```

The members that are in only one of two sets are the union of both differences:

```ant
let a = Set.of([1, 2, 3])
let b = Set.of([3, 4])
Set.union(Set.difference(a, b), Set.difference(b, a)) // Set.of([1, 2, 4])
```

It is a filter over the values of `a`:

```ant
let a = Set.of([1, 2, 3])
let b = Set.of([2])
assert(Set.difference(a, b).values() == a.values().filter({ !b.has(x) }))
```

The result is a new set, even when `b` is empty, and adding to it leaves `a` as it was:

```ant
let a = Set.of([1, 2])
let none: Set[int] = Set.of([])
let d = Set.difference(a, none)
assert(d == a)
assert(!Set.same(d, a))
d.add(3) // true
a        // Set.of([1, 2])
d        // Set.of([1, 2, 3])
```

## Possible issues

`Set.difference` is not symmetric. Swapping the arguments gives the other side:

```ant
let a = Set.of([1, 2])
let b = Set.of([2, 3])
Set.difference(a, b) // Set.of([1])
Set.difference(b, a) // Set.of([3])
```

`-` works on numbers only, so it does not subtract sets:

```ant
println(Set.of([1, 2]) - Set.of([1]))
// error: - applies to a single numeric kind or a named primitive type over one, not to Set[int]
```

Both sets need the same element type:

```ant
let d = Set.difference(Set.of([1]), Set.of(["x"])) // error: cannot use string where int is expected
```

Two sets built from empty arrays, `Set.of([])`, leave `T` open, which is a compile error. Annotate
one side, or the result:

```ant
let d = Set.difference(Set.of([]), Set.of([]))
// error: cannot infer the element type of an empty array; annotate it
```

To remove members from a set in place, call [Set.remove](remove.md) for each one instead of building
a new set:

```ant
let s = Set.of([1, 2, 3])
for x in [2, 3] {
  s.remove(x)
}
println(s)
```

```text
Set.of([1])
```

## See also

[Set.intersect](intersect.md), [Set.union](union.md), [Set.isSubset](isSubset.md),
[Set.remove](remove.md), [Set.has](has.md), [Array.filter](../Array/filter.md)
