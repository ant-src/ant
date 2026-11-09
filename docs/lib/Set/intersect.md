<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.intersect

```ant signature
func Set.intersect[T](a: Set[T], b: Set[T]): Set[T]
```

Set.intersect returns a new set holding the members of a that b has, in the order of a.

<!-- end generated -->

## Details

- Returns a new set; neither `a` nor `b` changes, even when `b` has every member of `a`.
- The order of `b` does not matter.

## Basic examples

Keep the numbers that both sets hold:

```ant
Set.intersect(Set.of([1, 2, 3, 4]), Set.of([3, 4, 5])) // Set.of([3, 4])
```

The result follows the order of the first set:

```ant
Set.intersect(Set.of(["c", "a", "b"]), Set.of(["b", "c", "x"])) // Set.of(["c", "b"])
```

## Scope

An empty set on either side, or two sets with no member in common, give the empty set:

```ant
var none: Set[int] = Set.of([])
Set.intersect(none, Set.of([1, 2]))        // Set.of([])
Set.intersect(Set.of([1, 2]), none)        // Set.of([])
Set.intersect(Set.of([1, 2]), Set.of([3])) // Set.of([])
```

Records match by their fields, not by identity:

```ant
type Point = { x: int, y: int }
var visited = Set.of([Point{ x: 0, y: 0 }, Point{ x: 1, y: 2 }])
var targets = Set.of([Point{ x: 1, y: 2 }, Point{ x: 5, y: 5 }])
Set.intersect(visited, targets) // Set.of([Point{ x: 1, y: 2 }])
```

`null` is a member like any other value:

```ant
var a: Set[string | null] = Set.of(["a", null])
var b: Set[string | null] = Set.of([null, "b"])
Set.intersect(a, b) // Set.of([null])
```

Call it through `.` or in a pipe:

```ant
Set.of([1, 2, 3]).intersect(Set.of([2]))          // Set.of([2])
Set.of([1, 2, 3]) | Set.intersect(Set.of([3, 1])) // Set.of([1, 3])
```

## Applications

Find the words two sentences share:

```ant
var first = Set.of(String.fields("the quick brown fox"))
var second = Set.of(String.fields("the lazy brown dog"))
Set.intersect(first, second) // Set.of(["the", "brown"])
```

Intersect any number of sets by folding with `Array.reduce`:

```ant
var sets = [Set.of([1, 2, 3, 4]), Set.of([2, 3, 4]), Set.of([4, 3, 9])]
Array.reduce(sets, sets[0], Set.intersect) // Set.of([3, 4])
```

Keep only the permissions a role grants:

```ant
var grants = Set.of(["read", "write"])
var asked = Set.of(["read", "delete", "write"])
var allowed = Set.intersect(asked, grants)
allowed // Set.of(["read", "write"])
```

Check whether two sets share a member:

```ant
!Set.isEmpty(Set.intersect(Set.of([1, 2]), Set.of([2, 3]))) // true
!Set.isEmpty(Set.intersect(Set.of([1, 2]), Set.of([3, 4]))) // false
```

## Properties and relations

The result is a subset of both sets:

```ant
var a = Set.of([1, 2, 3, 4])
var b = Set.of([2, 4, 6])
var both = Set.intersect(a, b)
assert(Set.isSubset(both, a))
assert(Set.isSubset(both, b))
```

Swapping the arguments gives an equal set, since `==` ignores order, but the members come in the
order of the new first argument:

```ant
var a = Set.of([3, 1, 2])
var b = Set.of([1, 3])
Set.intersect(a, b) == Set.intersect(b, a) // true
Set.values(Set.intersect(a, b))            // [3, 1]
Set.values(Set.intersect(b, a))            // [1, 3]
```

`Set.intersect` keeps the members of `a` that `Set.difference(a, b)` removes:

```ant
var a = Set.of(["x", "y", "z"])
var b = Set.of(["y", "w"])
assert(Set.intersect(a, b) == Set.difference(a, Set.difference(a, b)))
```

The result is never larger than either set:

```ant
var a = Set.of([1, 2, 3, 4, 5])
var b = Set.of([4, 5, 6])
var n = Set.size(Set.intersect(a, b))
assert(n <= Set.size(a) && n <= Set.size(b))
```

The result is a new set even when it equals `a`, so adding to it leaves `a` as it was:

```ant
var a = Set.of([1, 2])
var same = Set.intersect(a, a)
assert(same == a)
assert(!Set.same(same, a))
Set.add(same, 9) // true
a                // Set.of([1, 2])
```

## Possible issues

`Set.intersect` does not change `a`. A call whose result is thrown away keeps every member; assign
the result:

```ant
var tags = Set.of(["red", "green", "blue"])
Set.intersect(tags, Set.of(["red"]))
println(tags)
```

```text
Set.of(["red", "green", "blue"])
```

Both arguments must be sets. Turn an array into a set with `Set.of` first:

```ant
var ids = Set.of([1, 2, 3])
var kept = Set.intersect(ids, [2, 3]) // error: cannot use Array[int] where Set[int] is expected
```

Both sets must have the same element type. A `Set[int]` is not a `Set[int | null]`:

```ant
var a: Set[int | null] = Set.of([1, null])
var b: Set[int] = Set.of([1])
var both = Set.intersect(a, b) // error: cannot use Set[int] where Set[int | null] is expected
```

## See also

[Set.union](union.md), [Set.difference](difference.md), [Set.isSubset](isSubset.md),
[Set.has](has.md), [Set.of](of.md), [Array.filter](../Array/filter.md)
