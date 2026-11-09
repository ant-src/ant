<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.of

```ant signature
func Set.of[T](xs: Array[T]): Set[T]
```

Set.of returns a new set holding the first of each == value in xs, in order. Set.of(\[\]) is the
empty set, its element type coming from the expected type.

<!-- end generated -->

## Details

- Returns a new set. Adding to or removing from the set leaves `xs` as it was, and pushing to or
  removing from `xs` leaves the set as it was.
- A member is hashed when the set is built. Changing an array, map, set, or record after it became a
  member is a program error: the set may no longer find it.

## Basic examples

Build a set from an array. Duplicates go, and the first of each value keeps its place:

```ant
Set.of([3, 1, 3, 2, 1]) // Set.of([3, 1, 2])
Set.of(["red", "blue"]) // Set.of(["red", "blue"])
```

Call functions on the result through `.`:

```ant
var tags = Set.of(["ant", "go", "ant"])
tags.size()     // 2
tags.has("go")  // true
tags.has("zig") // false
```

## Scope

An empty array gives the empty set. The annotation supplies the element type:

```ant
var seen: Set[int] = Set.of([])
seen        // Set.of([])
seen.size() // 0
```

Records are members by contents, so two records with the same fields are one member:

```ant
type Point = { x: int, y: int }
Set.of([Point{ x: 1, y: 2 }, Point{ x: 1, y: 2 }]) // Set.of([Point{ x: 1, y: 2 }])
```

Members of a tagged union compare by tag and fields, so equal fields under two tags stay two
members:

```ant
type Circle = { r: int }
type Square = { r: int }
type Shape = Circle | Square
var shapes: Array[Shape] = [Circle{ r: 1 }, Square{ r: 1 }, Circle{ r: 1 }]
Set.of(shapes) // Set.of([Circle{ r: 1 }, Square{ r: 1 }])
```

Arrays compare by their elements in order; maps and sets compare in any order:

```ant
Set.of([[1, 2], [2, 1], [1, 2]])                 // Set.of([[1, 2], [2, 1]])
Set.of([Map{ a: 1, b: 2 }, Map{ b: 2, a: 1 }])   // Set.of([Map{ "a": 1, "b": 2 }])
Set.of([Set.of(["x", "y"]), Set.of(["y", "x"])]) // Set.of([Set.of(["x", "y"])])
```

An element type that includes `null` gives a set that can hold `null` once:

```ant
var answers: Array[int | null] = [1, null, null, 1]
Set.of(answers) // Set.of([1, null])
```

An array of mixed values gives a set of their union type:

```ant
var mixed = Set.of([1, "a", 1])
mixed // Set.of([1, "a"])
```

A string is not an array. Split it into code points with `String.chars` first:

```ant
Set.of(String.chars("hello")) // Set.of(["h", "e", "l", "o"])
```

In a pipeline, the array is the argument:

```ant
"b a b c a" | String.split(" ") | Set.of() // Set.of(["b", "a", "c"])
```

The parameter is named `xs`:

```ant
Set.of(xs: [2, 2, 7]) // Set.of([2, 7])
```

`Set.of` is a function value too, so it can go to `Array.map`:

```ant
[[1, 1], [2, 3, 2]].map(Set.of) // [Set.of([1]), Set.of([2, 3])]
```

## Applications

Count the distinct words of a sentence:

```ant
var words = String.split("the cat and the hat and the bat", " ")
Set.of(words).size() // 5
```

Collect the distinct values of a map:

```ant
var ages = Map{ ada: 36, linus: 54, alan: 36 }
Set.of(ages.values()) // Set.of([36, 54])
```

Keep a fixed list of allowed values as a set, and check input against it:

```ant
var allowed = Set.of(["debug", "info", "warn", "error"])
for level in ["info", "loud"] {
  if allowed.has(level) {
    println("ok: ${level}")
  } else {
    println("unknown level: ${level}")
  }
}
```

```text
ok: info
unknown level: loud
```

Find the letters two words share, in the order of the first word:

```ant
var a = Set.of(String.chars("listen"))
var b = Set.of(String.chars("pistol"))
Set.intersect(a, b).values() // ["l", "i", "s", "t"]
```

Start from a set of known members, then add to it while reading:

```ant
var visited = Set.of(["home"])
for page in ["about", "home", "blog", "about"] {
  if visited.add(page) {
    println("new: ${page}")
  }
}
println(visited)
```

```text
new: about
new: blog
Set.of(["home", "about", "blog"])
```

## Properties and relations

The members of `Set.of(xs)`, in order, are what `Array.unique` returns:

```ant
var xs = [4, 2, 4, 9, 2]
assert(Set.of(xs).values() == xs.unique())
```

The set never has more members than `xs` has elements, and every element of `xs` is a member:

```ant
var xs = ["a", "b", "a"]
var s = Set.of(xs)
assert(s.size() <= len(xs))
assert(xs.every({ s.has(x) }))
```

Order does not matter to `==`, so an array and its reverse give equal sets:

```ant
var xs = [1, 2, 3, 2]
assert(Set.of(xs) == Set.of(xs.reverse()))
assert(Set.of(xs).values() != Set.of(xs.reverse()).values())
```

Building a set from the values of another gives an equal, separate set:

```ant
var s = Set.of([1, 2])
var t = Set.of(s.values())
assert(s == t)
assert(!Set.same(s, t))
t.add(3)
assert(!s.has(3))
```

## Possible issues

`Set.of([])` with no expected type has no element type to infer:

```ant
var s = Set.of([]) // error: cannot infer the element type of an empty array; annotate it
```

There is no set literal. `Set{}` is read as a record literal:

```ant
var s: Set[int] = Set{} // error: a record literal needs a record type as its head
```

`Set.of` takes one array, not the members as separate arguments:

```ant
var s = Set.of(1, 2) // error: Set.of takes 1 argument, got 2
```

A set has no indexing. Use `Set.values` for an array to index:

```ant
var s = Set.of([10, 20])
println(s[0]) // error: Set[int] cannot be indexed
```

```ant
var s = Set.of([10, 20])
s.values()[0] // 10
```

`NaN` is not `==` to itself, so every `NaN` becomes a member of its own:

```ant
Set.of([Math.NaN, Math.NaN]).size() // 2
Set.of([0.0, -0.0]).size()          // 1
```

Strings compare by code points, so the two Unicode spellings of "é" are two members. Normalize them
with `String.normalize` first:

```ant
var names = ["café", "cafe\u{301}"]
Set.of(names).size()                                     // 2
Set.of(names.map({ String.normalize(x, "NFC") })).size() // 1
```

## See also

[Set.values](values.md) · [Set.has](has.md) · [Set.add](add.md) · [Set.size](size.md) ·
[Set.union](union.md) · [Array.unique](../Array/unique.md) ·
[Map.fromEntries](../Map/fromEntries.md)
