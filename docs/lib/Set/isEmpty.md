<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.isEmpty

```ant signature
func Set.isEmpty[T](s: Set[T]): bool
```

Set.isEmpty reports whether s has no members.

<!-- end generated -->

## Basic examples

```ant
Set.isEmpty(Set.of(["red", "green"])) // false
let none: Set[string] = Set.of([])
Set.isEmpty(none) // true
```

Call it through `.` on any set:

```ant
let tags = Set.of(["urgent"])
tags.isEmpty() // false
```

## Scope

Duplicates collapse, but one member is enough to make a set non-empty:

```ant
Set.of([7, 7, 7]).isEmpty() // false
```

The answer follows the set as `Set.add` and `Set.remove` change it:

```ant
let open: Set[int] = Set.of([])
open.isEmpty()  // true
open.add(42)    // true
open.isEmpty()  // false
open.remove(42) // true
open.isEmpty()  // true
```

The result of a set operation is often what you test, such as whether two sets share anything:

```ant
Set.of([1, 2]).intersect(Set.of([3, 4])).isEmpty() // true
Set.of([1, 2]).intersect(Set.of([2, 3])).isEmpty() // false
```

In a pipeline, the set is the first argument:

```ant
Set.of(["a"]) | Set.isEmpty() // false
```

Pass it to `Array.map` to test each set in an array:

```ant
let groups: Array[Set[int]] = [Set.of([1]), Set.of([])]
groups.map(Set.isEmpty) // [false, true]
```

For a set that may be `null`, `?.` yields `null` instead of calling:

```ant
let picked: Set[string] | null = null
picked?.isEmpty()         // null
picked?.isEmpty() ?? true // true
```

## Applications

Reject a post whose tags are not all allowed:

```ant
let allowed = Set.of(["ant", "go", "docs"])

func check(tags: Array[string]) {
  let unknown = Set.of(tags).difference(allowed)
  if unknown.isEmpty() {
    println("ok")
  } else {
    println("unknown tags: ${unknown.values()}")
  }
}

check(["ant", "docs"])
check(["ant", "rust", "zig"])
```

```text
ok
unknown tags: ["rust", "zig"]
```

Walk a graph level by level until no new nodes turn up:

```ant
let edges = Map{ a: ["b", "c"], b: ["d"], c: ["d"], d: [] }

let seen = Set.of(["a"])
var frontier = Set.of(["a"])
var depth = 0
while !frontier.isEmpty() {
  println("${depth}: ${frontier.values()}")
  let next: Set[string] = Set.of([])
  for node in frontier {
    for to in edges[node] ?? [] {
      if seen.add(to) {
        next.add(to)
      }
    }
  }
  frontier = next
  depth += 1
}
```

```text
0: ["a"]
1: ["b", "c"]
2: ["d"]
```

## Properties and relations

`Set.isEmpty` agrees with `Set.size`, `len`, `Set.values`, and `==` against the empty set:

```ant
let none: Set[int] = Set.of([])
let cases: Array[Set[int]] = [Set.of([]), Set.of([1]), Set.of([1, 2, 1])]
for s in cases {
  assert(s.isEmpty() == (s.size() == 0))
  assert(s.isEmpty() == (len(s) == 0))
  assert(s.isEmpty() == s.values().isEmpty())
  assert(s.isEmpty() == (s == none))
}
```

The empty set is a subset of every set:

```ant
let cases: Array[Set[int]] = [Set.of([]), Set.of([1, 2])]
for s in cases {
  assert(Set.of([]).isSubset(s))
}
```

## Possible issues

A set is not a condition. Ask `Set.isEmpty` explicitly:

```ant
let s = Set.of([1])
if s { // error: cannot use Set[int] where bool is expected
  println("has members")
}
```

`Set.of([])` alone has nothing to take an element type from. Annotate it:

```ant
println(Set.of([]).isEmpty()) // error: cannot infer the element type of an empty array; annotate it
```

A nullable set must be narrowed or reached through `?.` first:

```ant
let s: Set[int] | null = null
println(Set.isEmpty(s)) // error: cannot use Set[int] | null where Set[T] is expected
```

`Set.isEmpty` takes only sets. Arrays, maps, and strings have their own:

```ant
println(Set.isEmpty([1])) // error: cannot use Array[int] where Set[T] is expected
```

## See also

[Set.size](size.md) · [len](../prelude/len.md) · [Set.of](of.md) · [Set.has](has.md) ·
[Set.intersect](intersect.md) · [Set.difference](difference.md) ·
[Array.isEmpty](../Array/isEmpty.md) · [Map.isEmpty](../Map/isEmpty.md) ·
[String.isEmpty](../String/isEmpty.md)
