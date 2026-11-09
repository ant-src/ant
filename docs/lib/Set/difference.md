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
var none: Set[int] = Set.of([])
Set.difference(none, Set.of([1]))    // Set.of([])
Set.difference(Set.of([1, 2]), none) // Set.of([1, 2])
```

A set minus itself is empty:

```ant
var s = Set.of([1, 2])
Set.difference(s, s) // Set.of([])
```

Members compare with `==`, so strings compare by their code points, and records and arrays by their
contents:

```ant
type Point = { x: int, y: int }
var points = Set.of([Point{ x: 1, y: 2 }, Point{ x: 0, y: 0 }])
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
var before = Set.of(["ann", "bob", "cy"])
var after = Set.of(["bob", "cy", "dee"])
println("removed: ${Set.values(Set.difference(before, after))}")
println("added: ${Set.values(Set.difference(after, before))}")
```

```text
removed: ["ann"]
added: ["dee"]
```

Report the required settings that a configuration does not set:

```ant
var required = Set.of(["host", "port", "user"])
var config = Map{ host: "localhost", user: "ann" }
for name in Set.difference(required, Set.of(Map.keys(config))) {
  println("missing: ${name}")
}
```

```text
missing: port
```

## Properties and relations

The result has no member of `b`, and every member of it is a member of `a`:

```ant
var a = Set.of([1, 2, 3])
var b = Set.of([2, 5])
var d = Set.difference(a, b)
assert(Set.isEmpty(Set.intersect(d, b)))
assert(Set.isSubset(d, a))
```

The difference and the intersection split `a` in two:

```ant
var a = Set.of([1, 2, 3])
var b = Set.of([2, 5])
assert(Set.union(Set.difference(a, b), Set.intersect(a, b)) == a)
```

The difference is empty exactly when `a` is a subset of `b`:

```ant
var a = Set.of([1, 2])
var b = Set.of([1, 2, 3])
assert(Set.isEmpty(Set.difference(a, b)) == Set.isSubset(a, b))
assert(!Set.isEmpty(Set.difference(b, a)))
```

The members that are in only one of two sets are the union of both differences:

```ant
var a = Set.of([1, 2, 3])
var b = Set.of([3, 4])
Set.union(Set.difference(a, b), Set.difference(b, a)) // Set.of([1, 2, 4])
```

It is a filter over the values of `a`:

```ant
var a = Set.of([1, 2, 3])
var b = Set.of([2])
assert(Set.values(Set.difference(a, b)) == Array.filter(Set.values(a), { !Set.has(b, x) }))
```

The result is a new set, even when `b` is empty, and adding to it leaves `a` as it was:

```ant
var a = Set.of([1, 2])
var none: Set[int] = Set.of([])
var d = Set.difference(a, none)
assert(d == a)
assert(!Set.same(d, a))
Set.add(d, 3) // true
a             // Set.of([1, 2])
d             // Set.of([1, 2, 3])
```

## Possible issues

`Set.difference` is not symmetric. Swapping the arguments gives the other side:

```ant
var a = Set.of([1, 2])
var b = Set.of([2, 3])
Set.difference(a, b) // Set.of([1])
Set.difference(b, a) // Set.of([3])
```

`-` works on numbers only, so it does not subtract sets:

```ant
println(Set.of([1, 2]) - Set.of([1]))
// error: - applies to a single numeric kind or a named primitive type over one, not to Set[int]; narrow first
```

Both sets need the same element type:

```ant
var d = Set.difference(Set.of([1]), Set.of(["x"])) // error: cannot use string where int is expected
```

Two sets built from empty arrays, `Set.of([])`, leave `T` open, which is a compile error. Annotate
one side, or the result:

```ant
var d = Set.difference(Set.of([]), Set.of([]))
// error: cannot infer the element type of an empty array; annotate it
```

To remove members from a set in place, call [Set.remove](remove.md) for each one instead of building
a new set:

```ant
var s = Set.of([1, 2, 3])
for x in [2, 3] {
  Set.remove(s, x)
}
println(s)
```

```text
Set.of([1])
```

## See also

[Set.intersect](intersect.md), [Set.union](union.md), [Set.isSubset](isSubset.md),
[Set.remove](remove.md), [Set.has](has.md), [Array.filter](../Array/filter.md)
