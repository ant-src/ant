<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.size

```ant signature
func Set.size[T](s: Set[T]): int
```

Set.size returns the number of members of s.

<!-- end generated -->

## Details

- Members compare the way `==` does, so two equal records or arrays are one member.
- `len(s)` returns the same count.

## Basic examples

```ant
Set.size(Set.of(["red", "green", "blue"])) // 3
var none: Set[int] = Set.of([])
Set.size(none) // 0
```

Call it through `.` on any set:

```ant
var tags = Set.of(["go", "ant"])
tags.size() // 2
```

## Scope

`Set.of` keeps one of each `==` value, so the size counts distinct values, not array elements:

```ant
Set.of([1, 2, 2, 3, 3, 3]).size() // 3
```

Equal records are one member:

```ant
type Point = { x: int, y: int }

var points = Set.of([Point{ x: 1, y: 2 }, Point{ x: 1, y: 2 }, Point{ x: 0, y: 0 }])
points.size() // 2
```

`null` is a member like any other:

```ant
var answers: Set[bool | null] = Set.of([true, null, null])
answers.size() // 2
```

Strings compare by code points, without normalizing, so two spellings of one letter are two members:

```ant
Set.of(["café", "cafe\u{301}"]).size() // 2
```

The size follows the set as it changes:

```ant
var seen: Set[string] = Set.of([])
seen.add("go")     // true
seen.add("ant")    // true
seen.size()        // 2
seen.add("go")     // false
seen.size()        // 2
seen.remove("go")  // true
seen.size()        // 1
seen.remove("zig") // false
seen.size()        // 1
```

In a pipeline, the set is the first argument:

```ant
Set.of([1, 2]) | Set.size() // 2
```

As a function value, it maps an array of sets to their sizes:

```ant
[Set.of([1]), Set.of([1, 2])].map(Set.size) // [1, 2]
```

For a set that may be `null`, `?.` yields `null` instead of calling:

```ant
var tags: Set[string] | null = null
tags?.size()      // null
tags?.size() ?? 0 // 0
```

## Applications

Count the distinct words of a text:

```ant
func distinct(text: string): int {
  Set.of(String.fields(text)).size()
}

println(distinct("to be or not to be"))
```

```text
4
```

Report whether a list has duplicates:

```ant
func hasDuplicates(xs: Array[string]): bool {
  Set.of(xs).size() < xs.length()
}

println(hasDuplicates(["ada", "linus", "grace"]))
println(hasDuplicates(["ada", "linus", "ada"]))
```

```text
false
true
```

Measure how much two sets of tags overlap, as the share of all tags they have in common:

```ant
var a = Set.of(["go", "rust", "ant"])
var b = Set.of(["ant", "zig", "go", "c"])
var shared = a.intersect(b).size()
var all = a.union(b).size()
println("${shared} of ${all} tags shared")
```

```text
2 of 5 tags shared
```

## Properties and relations

`Set.size` agrees with `len`, `Set.isEmpty`, and the length of `Set.values`:

```ant
var cases: Array[Set[int]] = [Set.of([]), Set.of([1]), Set.of([1, 2, 2])]
for s in cases {
  assert(s.size() == len(s))
  assert(s.isEmpty() == (s.size() == 0))
  assert(s.size() == s.values().length())
}
```

The sizes of a union, an intersection, and a difference add up:

```ant
var a = Set.of([1, 2, 3, 4])
var b = Set.of([3, 4, 5])
assert(a.union(b).size() == a.size() + b.size() - a.intersect(b).size())
assert(a.difference(b).size() + a.intersect(b).size() == a.size())
assert(a.intersect(b).size() <= a.size())
```

## Possible issues

`size` is a function, not a field. Without parentheses it is a compile error:

```ant
var s = Set.of([1])
println(s.size) // error: Set[int] is not a record, so it has no field size
```

`Set.of([])` takes its element type from the expected type. Without one, the empty array has no
element type to infer. Annotate it:

```ant
println(Set.of([]).size()) // error: cannot infer the element type of an empty array; annotate it
```

A nullable set must be narrowed or reached through `?.` first:

```ant
var s: Set[int] | null = null
println(s.size())
// error: Set[int] | null has no field or function size on every member: narrow the value first
```

`Set.size` takes only sets. Use [`Map.size`](../Map/size.md) for a map,
[`Array.length`](../Array/length.md) for an array, or [`len`](../prelude/len.md), which takes them
all:

```ant
println(Set.size([1, 2])) // error: cannot use Array[int] where Set[T] is expected
```

A set is a reference value: assigning or passing one shares it. An add through one name changes the
size seen through the other:

```ant
var a = Set.of([1])
var b = a
b.add(2) // true
a.size() // 2
```

## See also

[Set.isEmpty](isEmpty.md) · [Set.of](of.md) · [Set.add](add.md) · [Set.remove](remove.md) ·
[Set.values](values.md) · [len](../prelude/len.md) · [Map.size](../Map/size.md) ·
[Array.length](../Array/length.md)
