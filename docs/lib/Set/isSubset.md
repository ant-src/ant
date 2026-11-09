<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.isSubset

```ant signature
func Set.isSubset[T](a: Set[T], b: Set[T]): bool
```

Set.isSubset reports whether every member of a is a member of b.

<!-- end generated -->

## Details

- Compares members by [`==`](../../language/operators/equality.md): arrays, maps, and sets match by
  contents, and records by tag and fields.
- The empty set is a subset of every set.
- Neither set changes.

## Basic examples

```ant
Set.isSubset(Set.of([1, 2]), Set.of([3, 2, 1])) // true
Set.isSubset(Set.of([1, 4]), Set.of([3, 2, 1])) // false
```

Call it through `.` on any set:

```ant
var picked = Set.of(["red", "blue"])
var palette = Set.of(["red", "green", "blue"])
picked.isSubset(palette) // true
palette.isSubset(picked) // false
```

## Scope

The empty set is a subset of every set, the empty set included. Annotate it so that `T` is known:

```ant
var none: Set[int] = Set.of([])
none.isSubset(Set.of([1, 2]))    // true
none.isSubset(none)              // true
Set.of([1]).isSubset(Set.of([])) // false
```

The second argument takes its element type from the first, so an inline `Set.of([])` works there,
while the first needs the annotation.

A set is a subset of itself and of any equal set, whatever the order of its members:

```ant
var a = Set.of([1, 2, 3])
a.isSubset(a)                 // true
a.isSubset(Set.of([3, 2, 1])) // true
```

Strings compare by code points. `==` does not normalize, so the two spellings of "é" differ:

```ant
Set.of(["café"]).isSubset(Set.of(["café", "thé"])) // true
Set.of(["cafe\u{301}"]).isSubset(Set.of(["café"])) // false
```

Records compare by tag and fields:

```ant
type Point = { x: int, y: int }
var corners = Set.of([Point{ x: 0, y: 0 }, Point{ x: 1, y: 1 }])
var grid = Set.of([Point{ x: 0, y: 0 }, Point{ x: 0, y: 1 }, Point{ x: 1, y: 0 }, Point{ x: 1, y: 1 }])
corners.isSubset(grid) // true
```

Members of a tagged union compare by tag, so equal fields under another tag do not match:

```ant
type Circle = { size: float64 }
type Square = { size: float64 }
type Shape = Circle | Square
var drawn: Set[Shape] = Set.of([Circle{ size: 1.0 }, Square{ size: 2.0 }])
var small: Set[Shape] = Set.of([Circle{ size: 1.0 }])
var big: Set[Shape] = Set.of([Circle{ size: 2.0 }])
small.isSubset(drawn) // true
big.isSubset(drawn)   // false
```

Nested arrays and maps compare by contents:

```ant
Set.of([[1, 2]]).isSubset(Set.of([[1, 2], [3]]))      // true
Set.of([Map{ a: 1 }]).isSubset(Set.of([Map{ a: 1 }])) // true
```

A set whose element type includes `null` can hold `null` like any other member:

```ant
var answers: Set[bool | null] = Set.of([true, null])
var unanswered: Set[bool | null] = Set.of([null])
unanswered.isSubset(answers)            // true
answers.isSubset(Set.of([true, false])) // false
```

In a pipeline, the candidate subset is the first argument:

```ant
"b a" | String.split(" ") | Set.of() | Set.isSubset(Set.of(["a", "b", "c"])) // true
```

## Applications

Check that a request asks only for permissions a role grants:

```ant
func allowed(granted: Set[string], wanted: Array[string]): bool {
  Set.of(wanted).isSubset(granted)
}

var editor = Set.of(["read", "write", "comment"])
println(allowed(editor, ["read", "write"]))
println(allowed(editor, ["read", "delete"]))
```

```text
true
false
```

Report the required columns a CSV header lacks, and accept it when there are none:

```ant
var required = Set.of(["id", "name", "email"])
var header = Set.of(String.split("id,email,age", ","))
if required.isSubset(header) {
  println("ok")
} else {
  println("missing: ${Set.difference(required, header).values()}")
}
```

```text
missing: ["name"]
```

Test whether a word can be spelled from a set of letters, using each letter any number of times:

```ant
var letters = Set.of(String.chars("aenst"))
for word in ["ants", "nest", "tent", "lens"] {
  println("${word}: ${Set.of(String.chars(word)).isSubset(letters)}")
}
```

```text
ants: true
nest: true
tent: true
lens: false
```

"tent" needs two "t"s and still passes: see [Possible issues](#possible-issues).

## Properties and relations

`a` is a subset of `b` exactly when `a` has no member that `b` lacks, when their intersection is
`a`, and when their union is `b`:

```ant
var b = Set.of([1, 2, 3, 4])
for a in [Set.of([2, 4]), Set.of([4, 5]), Set.of([1, 2, 3, 4])] {
  var sub = a.isSubset(b)
  assert(sub == Set.difference(a, b).isEmpty())
  assert(sub == (Set.intersect(a, b) == a))
  assert(sub == (Set.union(b, a) == b))
  assert(sub == a.values().every({ b.has(x) }))
}
```

Two sets are equal exactly when each is a subset of the other:

```ant
var a = Set.of(["x", "y"])
var b = Set.of(["y", "x"])
assert((a == b) == (a.isSubset(b) && b.isSubset(a)))
```

Every set is a subset of its union with another, and its intersection with another is a subset of
it:

```ant
var a = Set.of([1, 2])
var b = Set.of([2, 3])
assert(a.isSubset(Set.union(a, b)))
assert(Set.intersect(a, b).isSubset(a))
assert(Set.difference(a, b).isSubset(a))
```

## Possible issues

`Set.isSubset` allows the two sets to be equal. For a proper subset, also compare sizes:

```ant
var a = Set.of([1, 2])
var b = Set.of([2, 1])
a.isSubset(b)                        // true
a.isSubset(b) && a.size() < b.size() // false
```

`NaN` is not `==` to itself, so a set that holds `NaN` is not a subset even of itself:

```ant
var xs = Set.of([1.0, Math.NaN])
xs.isSubset(xs)            // false
Set.of([1.0]).isSubset(xs) // true
```

The first set fixes the element type. A set built from records of one member of a union is a set of
that member, not of the union, and a set of the union is not assignable to it:

```ant
type Circle = { size: float64 }
type Square = { size: float64 }
type Shape = Circle | Square
var drawn: Set[Shape] = Set.of([Circle{ size: 1.0 }, Square{ size: 2.0 }])
println(Set.of([Circle{ size: 1.0 }]).isSubset(drawn))
// error: cannot use Set[Shape] where Set[Circle] is expected
```

Annotate the first set with the union type, as in Scope above.

A set holds each value once, so `isSubset` ignores repeats:

```ant
Set.of(String.chars("sees")).isSubset(Set.of(String.chars("se"))) // true
```

To compare how many times each value occurs, count them in a `Map`.

Both sets have the same element type. Numeric types never mix:

```ant
var ids = Set.of([1, 2])
println(ids.isSubset(Set.of([1.0, 2.0]))) // error: cannot use float64 where int is expected
```

`Set.isSubset` asks about sets. To ask whether an array's elements all appear in another array,
build a set from each, or test with [`Array.every`](../Array/every.md):

```ant
var xs = [1, 2, 2]
var ys = [3, 2, 1]
Set.of(xs).isSubset(Set.of(ys)) // true
xs.every({ ys.contains(x) })    // true
```

## See also

[Set.has](has.md) · [Set.intersect](intersect.md) · [Set.difference](difference.md) ·
[Set.union](union.md) · [Set.same](same.md) · [Set.of](of.md) · [Array.every](../Array/every.md) ·
[== operator](../../language/operators/equality.md)
