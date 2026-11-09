<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.union

```ant signature
func Set.union[T](a: Set[T], b: Set[T]): Set[T]
```

Set.union returns a new set holding the members of a, then those of b that a lacks.

<!-- end generated -->

## Details

- Returns a new set. Neither argument changes, even when the other one is empty.
- The members that `b` adds follow in the order of `b`. A member of both keeps its position in `a`.

## Basic examples

Combine two sets; a member of both appears once:

```ant
Set.union(Set.of([1, 2]), Set.of([3, 2])) // Set.of([1, 2, 3])
```

The members of `a` come first, and `b` adds only what `a` lacks:

```ant
Set.values(Set.union(Set.of([3, 1]), Set.of([2, 1, 0]))) // [3, 1, 2, 0]
```

## Scope

An empty set on either side gives the other one's members:

```ant
let none: Set[int] = Set.of([])
Set.union(none, Set.of([1])) // Set.of([1])
Set.union(Set.of([1]), none) // Set.of([1])
Set.union(none, none)        // Set.of([])
```

A set joined with itself has the same members:

```ant
let xs = Set.of([1, 2])
Set.union(xs, xs) // Set.of([1, 2])
```

Records and arrays match by `==`, and `null` can be a member like any other value:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
Set.union(Set.of([p]), Set.of([Point{ x: 1, y: 2 }])) // Set.of([Point{ x: 1, y: 2 }])
Set.union(Set.of([[1], [2]]), Set.of([[1]]))          // Set.of([[1], [2]])
Set.union(Set.of([1, null]), Set.of([null, 2]))       // Set.of([1, null, 2])
```

An expected type fixes `T`, so sets of different member types can join into a union:

```ant
let mixed: Set[int | string] = Set.union(Set.of([1]), Set.of(["x"]))
mixed // Set.of([1, "x"])
```

Call it through `.` on a set, or with a pipe, which passes its left side as the first argument:

```ant
Set.of([1]).union(Set.of([2]))       // Set.of([1, 2])
Set.of([1]) | Set.union(Set.of([2])) // Set.of([1, 2])
```

Name the arguments to pass them in any order:

```ant
Set.union(b: Set.of([2]), a: Set.of([1])) // Set.of([1, 2])
```

## Applications

List every tag used by two posts, each once, in the order they first appear:

```ant
let first = Set.of(["ant", "lang", "docs"])
let second = Set.of(["docs", "tests"])
println(String.join(Set.values(Set.union(first, second)), ", "))
```

```text
ant, lang, docs, tests
```

Collect the members of several sets with a loop:

```ant
let groups = [Set.of(["ann", "bob"]), Set.of(["bob", "cy"]), Set.of(["ann", "dee"])]
var everyone: Set[string] = Set.of([])
for g in groups {
  everyone = Set.union(everyone, g)
}
println(Set.values(everyone))
```

```text
["ann", "bob", "cy", "dee"]
```

## Properties and relations

`Set.union` is commutative as sets, since `==` on sets ignores order, but the order of the members
differs:

```ant
let a = Set.of([1, 2])
let b = Set.of([2, 3])
assert(Set.union(a, b) == Set.union(b, a))
assert(Set.values(Set.union(a, b)) != Set.values(Set.union(b, a)))
```

Both arguments are subsets of the result:

```ant
let a = Set.of([1, 2])
let b = Set.of([2, 3])
let u = Set.union(a, b)
assert(Set.isSubset(a, u))
assert(Set.isSubset(b, u))
```

Its size is the sum of the sizes less the members the two share:

```ant
let a = Set.of([1, 2, 3])
let b = Set.of([3, 4])
assert(Set.size(Set.union(a, b)) == Set.size(a) + Set.size(b) - Set.size(Set.intersect(a, b)))
```

The result equals `Set.of` over the values of `a` followed by the values of `b`:

```ant
let a = Set.of([3, 1])
let b = Set.of([1, 2])
assert(Set.values(Set.union(a, b)) == Set.values(Set.of([...Set.values(a), ...Set.values(b)])))
```

The result is a new set, even when one side is empty, and adding to it leaves `a` as it was:

```ant
let a = Set.of([1, 2])
let u = Set.union(a, Set.of([]))
assert(u == a)
assert(!Set.same(u, a))
Set.add(u, 9) // true
a             // Set.of([1, 2])
u             // Set.of([1, 2, 9])
```

## Possible issues

`Set.union` does not change `a`. To add members to a set in place, call [Set.add](add.md) for each,
or assign the result back:

```ant
var seen = Set.of([1])
Set.union(seen, Set.of([2])) // Set.of([1, 2])
seen                         // Set.of([1])
seen = Set.union(seen, Set.of([2]))
seen // Set.of([1, 2])
```

Strings match by code point, so two spellings of the same text stay two members:

```ant
Set.size(Set.union(Set.of(["é"]), Set.of(["é"]))) // 2
```

Two sets of different member types do not join without an expected type that holds both:

```ant
let both = Set.union(Set.of([1]), Set.of(["x"])) // error: cannot use string where int is expected
```

Both arguments are sets. Turn an array into a set with [Set.of](of.md) first:

```ant
let more = Set.union(Set.of([1]), [2]) // error: cannot use Array[int] where Set[int] is expected
```

Two empty sets built from literals leave `T` open. Annotate one side, or the result:

```ant
let none = Set.union(Set.of([]), Set.of([]))
// error: cannot infer the element type of an empty array; annotate it
```

## See also

[Set.intersect](intersect.md), [Set.difference](difference.md), [Set.isSubset](isSubset.md),
[Set.add](add.md), [Set.of](of.md), [Set.values](values.md), [Map.merge](../Map/merge.md),
[Array.concat](../Array/concat.md)
