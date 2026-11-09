<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.contains

```ant signature
func Array.contains[T](xs: Array[T], v: T): bool
```

Array.contains reports whether some element of xs is == v.

<!-- end generated -->

## Details

- Compares each element with `v` by [`==`](../../language/operators/equality.md), so arrays, maps,
  sets, and records match by contents.
- An empty array gives `false`.
- `NaN` equals nothing, itself included, so an array never contains `NaN`.

## Basic examples

```ant
Array.contains([1, 2, 3], 2) // true
Array.contains([1, 2, 3], 5) // false
```

Call it through `.` on any array:

```ant
var fruit = ["apple", "pear", "plum"]
fruit.contains("pear") // true
fruit.contains("fig")  // false
```

## Scope

An empty array contains nothing. Annotate it so that `T` is known:

```ant
var empty: Array[string] = []
empty.contains("") // false
```

Strings compare by code points. `==` does not normalize, so the two spellings of "é" differ:

```ant
var words = ["café", "naïve"]
words.contains("café")        // true
words.contains("cafe\u{301}") // false
words.contains("CAFÉ")        // false
```

Records compare by tag and fields, so a fresh record with the same fields is found:

```ant
type Point = { x: int, y: int }
var path = [Point{ x: 0, y: 0 }, Point{ x: 1, y: 0 }, Point{ x: 1, y: 1 }]
path.contains(Point{ x: 1, y: 0 }) // true
path.contains(Point{ x: 0, y: 1 }) // false
```

Members of a tagged union compare by tag and fields, so equal fields under another tag do not match:

```ant
type Circle = { size: float64 }
type Square = { size: float64 }
type Shape = Circle | Square
var shapes: Array[Shape] = [Circle{ size: 1.0 }, Square{ size: 2.0 }]
shapes.contains(Square{ size: 2.0 }) // true
shapes.contains(Circle{ size: 2.0 }) // false
```

A named primitive type compares its tag and its payload:

```ant
type UserId = string
var admins = [UserId("u-1"), UserId("u-7")]
admins.contains(UserId("u-7")) // true
```

Nested arrays, maps, and sets compare by contents. Two maps or two sets are equal in any order:

```ant
[[1, 2], [3]].contains([3])                       // true
[[1, 2], [3]].contains([2, 1])                    // false
[Map{ a: 1, b: 2 }].contains(Map{ b: 2, a: 1 })   // true
[Set.of(["x", "y"])].contains(Set.of(["y", "x"])) // true
```

An array whose element type includes `null` can look for `null`:

```ant
var scores: Array[int | null] = [90, null, 75]
scores.contains(null) // true
scores.contains(75)   // true
```

In a pipeline, the array is the first argument:

```ant
"the quick brown fox" | String.split(" ") | Array.contains("fox") // true
```

## Applications

Drop stop words from a sentence:

```ant
var stop = ["a", "an", "the", "of"]
var words = String.split("the art of an ant", " ")
words.filter({ !stop.contains(x) }) // ["art", "ant"]
```

Accept only known values for an option:

```ant
func checkLevel(level: string): Result[string, string] {
  if ["debug", "info", "warn", "error"].contains(level) {
    Result.ok(level)
  } else {
    Result.err("unknown level: ${level}")
  }
}

println(checkLevel("info"))
println(checkLevel("loud"))
```

```text
Ok{ value: "info" }
Err{ error: "unknown level: loud" }
```

Walk a grid and stop at the first cell visited twice:

```ant
type Point = { x: int, y: int }

var moves = ["R", "U", "L", "D", "R"]
var at = Point{ x: 0, y: 0 }
var seen = [at]
for m in moves {
  var dx = if m == "R" { 1 } else if m == "L" { -1 } else { 0 }
  var dy = if m == "U" { 1 } else if m == "D" { -1 } else { 0 }
  at = Point{ x: at.x + dx, y: at.y + dy }
  if seen.contains(at) {
    println("back at ${at}")
    break
  }
  seen.push(at)
}
```

```text
back at Point{ x: 0, y: 0 }
```

## Properties and relations

`Array.contains` agrees with `Array.indexOf`, `Array.some`, and `Set.has`, which all compare by
`==`:

```ant
var xs = [3, 1, 4, 1, 5]
for v in [1, 2, 5] {
  assert(xs.contains(v) == (xs.indexOf(v) != null))
  assert(xs.contains(v) == xs.some({ x == v }))
  assert(xs.contains(v) == Set.of(xs).has(v))
  assert(xs.contains(v) == xs.reverse().contains(v))
}
```

After `Array.push`, the array contains the pushed value:

```ant
var xs = ["a", "b"]
xs.push("c")
assert(xs.contains("c"))
```

`Array.unique` keeps the first of each `==` value. A loop that pushes only what the result does not
yet contain builds the same array:

```ant
var xs = [2, 7, 2, 1, 7]
var ys: Array[int] = []
for v in xs {
  if !ys.contains(v) {
    ys.push(v)
  }
}
assert(ys == xs.unique())
```

## Possible issues

`NaN` is not `==` to itself, so an array never contains it. Test with `Math.isNaN`:

```ant
var xs = [1.0, Math.NaN]
xs.contains(Math.NaN)      // false
xs.some({ Math.isNaN(x) }) // true
[0.0].contains(-0.0)       // true
```

Numeric types never mix, so looking for a `float64` in an `Array[int]` is a compile error:

```ant
var counts = [1, 2, 3]
println(counts.contains(2.0)) // error: cannot use float64 where int is expected
```

Nor is `null` a valid value unless the element type includes it:

```ant
var counts = [1, 2, 3]
println(counts.contains(null)) // error: cannot use null where int is expected
```

A nullable value, such as the result of `Array.get`, `Map` indexing, or `Array.first`, is not an
`int` either, so it cannot be looked up in an `Array[int]` as it is:

```ant
var xs = [1, 2, 3]
var m: int | null = Map{ a: 2 }["a"]
println(xs.contains(m)) // error: cannot use int | null where int is expected
```

Narrow it first:

```ant
var xs = [1, 2, 3]
var m: int | null = Map{ a: 2 }["a"]
m != null && xs.contains(m) // true
```

`Array.contains` takes a value, not a predicate. To test elements with a function, use
[`Array.some`](some.md):

```ant
var counts = [1, 2, 3]
println(counts.contains({ x > 2 }))
// error: a lambda needs an expected function type, but int is expected
```

`==` compares records by their contents at the time of the call. A record changed after it went into
the array matches its new fields, and a copy matches as well as the original. To ask whether the
array holds one particular record, compare identity with [`Record.same`](../Record/same.md):

```ant
type Point = { x: int, y: int }
var p = Point{ x: 1, y: 2 }
var ps = [p]
p.x = 5
ps.contains(Point{ x: 1, y: 2 })           // false
ps.contains(Point{ x: 5, y: 2 })           // true
ps.some({ Record.same(x, Point{ ...p }) }) // false
ps.some({ Record.same(x, p) })             // true
```

Functions compare by identity, so two lambdas with the same body are different values:

```ant
var inc: func(int): int = { x + 1 }
var same: func(int): int = { x + 1 }
[inc].contains(inc)  // true
[inc].contains(same) // false
```

Text that can arrive in either Unicode spelling needs normalizing on both sides first, with
[`String.normalize`](../String/normalize.md):

```ant
var names = ["café"].map({ String.normalize(x, "NFC") })
names.contains(String.normalize("cafe\u{301}", "NFC")) // true
```

## See also

[Array.indexOf](indexOf.md) · [Array.lastIndexOf](lastIndexOf.md) · [Array.some](some.md) ·
[Array.find](find.md) · [Array.unique](unique.md) · [Set.has](../Set/has.md) ·
[Map.has](../Map/has.md) · [String.contains](../String/contains.md) ·
[== operator](../../language/operators/equality.md)
