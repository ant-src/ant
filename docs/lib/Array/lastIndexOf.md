<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.lastIndexOf

```ant signature
func Array.lastIndexOf[T](xs: Array[T], v: T): int | null
```

Array.lastIndexOf returns the index of the last element of xs that is == v, or null when there is
none.

<!-- end generated -->

## Details

- Compares each element with `v` by [`==`](../../language/operators/equality.md) and returns the
  last matching index.
- Returns `null` when no element matches, so an empty array gives `null`.
- `NaN` equals nothing, itself included, so `NaN` is never found.

## Basic examples

```ant
Array.lastIndexOf([1, 2, 3, 2, 1], 2) // 3
Array.lastIndexOf([1, 2, 3], 5)       // null
```

Call it through `.` on any array:

```ant
let path = ["home", "docs", "home", "src"]
path.lastIndexOf("home") // 2
path.indexOf("home")     // 0
```

## Scope

An empty array has no last match. Annotate it so that `T` is known:

```ant
let empty: Array[string] = []
empty.lastIndexOf("") // null
```

When the value occurs once, the first and the last index are the same:

```ant
[4, 5, 6].lastIndexOf(5) // 1
[4, 5, 6].indexOf(5)     // 1
```

Strings compare by code points. `==` does not normalize, so the two spellings of "é" differ:

```ant
let words = ["café", "cafe\u{301}", "café"]
words.lastIndexOf("café")        // 2
words.lastIndexOf("cafe\u{301}") // 1
```

Records compare by tag and fields, so a fresh record with the same fields is found:

```ant
type Point = { x: int, y: int }
let trail = [Point{ x: 0, y: 0 }, Point{ x: 1, y: 0 }, Point{ x: 0, y: 0 }, Point{ x: 0, y: 1 }]
trail.lastIndexOf(Point{ x: 0, y: 0 }) // 2
trail.lastIndexOf(Point{ x: 5, y: 5 }) // null
```

Members of a tagged union compare by tag and fields, so equal fields under another tag do not match:

```ant
type Circle = { size: float64 }
type Square = { size: float64 }
type Shape = Circle | Square
let shapes: Array[Shape] = [Square{ size: 2.0 }, Circle{ size: 2.0 }, Square{ size: 2.0 }]
shapes.lastIndexOf(Square{ size: 2.0 }) // 2
shapes.lastIndexOf(Circle{ size: 2.0 }) // 1
```

Nested arrays and maps compare by contents:

```ant
[[1, 2], [3], [1, 2]].lastIndexOf([1, 2])                       // 2
[Map{ a: 1 }, Map{ a: 1, b: 2 }].lastIndexOf(Map{ b: 2, a: 1 }) // 1
```

An array whose element type includes `null` can look for the last `null`:

```ant
let readings: Array[int | null] = [12, null, 15, null, 9]
readings.lastIndexOf(null) // 3
readings.lastIndexOf(15)   // 2
```

In a pipeline, the array is the first argument:

```ant
"a/b/a/c" | String.split("/") | Array.lastIndexOf("a") // 2
```

## Named arguments

`v` has no default, but a call can still name it, as any declared function's arguments can be named:

```ant
Array.lastIndexOf([7, 8, 7], v: 7) // 2
```

## Applications

A command line passes arguments through groups separated by `--`. Keep the last group, the arguments
meant for the innermost program:

```ant
let args = ["ant", "run", "--", "tool.ant", "--", "-v", "input.txt"]
let i = args.lastIndexOf("--")
if i != null {
  println(Array.slice(args, i + 1, len(args)))
}
```

```text
["-v", "input.txt"]
```

Find which step of an undo history last reached a state:

```ant
type State = { x: int, y: int }

let history = [State{ x: 0, y: 0 }, State{ x: 1, y: 0 }, State{ x: 0, y: 0 }, State{ x: 2, y: 3 }]
let step = history.lastIndexOf(State{ x: 0, y: 0 })
if step != null {
  println("last at origin: step ${step}")
}
```

```text
last at origin: step 2
```

Report the values that occur more than once, by comparing the first and the last index:

```ant
func repeated(xs: Array[string]): Array[string] {
  xs.unique().filter({ xs.indexOf(x) != xs.lastIndexOf(x) })
}

println(repeated(["ant", "bee", "ant", "fly", "bee", "wasp"]))
```

```text
["ant", "bee"]
```

## Properties and relations

`Array.lastIndexOf` agrees with [`Array.findLastIndex`](findLastIndex.md) given an `==` test, and
returns `null` exactly when [`Array.contains`](contains.md) is `false`:

```ant
let xs = ["a", "b", "a", "c"]
for v in ["a", "b", "c", "z"] {
  assert(xs.lastIndexOf(v) == xs.findLastIndex({ x == v }))
  assert((xs.lastIndexOf(v) != null) == xs.contains(v))
}
```

The element at the index is `==` to `v`, and no later element is:

```ant
let xs = [3, 1, 4, 1, 5, 9, 1]
let i = xs.lastIndexOf(1)
assert(i != null)
assert(xs[i] == 1)
assert(!Array.slice(xs, i + 1, len(xs)).contains(1))
```

It mirrors [`Array.indexOf`](indexOf.md) on the reversed array:

```ant
let xs = [3, 1, 4, 1, 5]
let j = xs.reverse().indexOf(1)
assert(j != null)
assert(xs.lastIndexOf(1) == len(xs) - 1 - j)
```

The first index never comes after the last:

```ant
let xs = [2, 7, 2, 1, 7]
for v in xs {
  assert((xs.indexOf(v) ?? 0) <= (xs.lastIndexOf(v) ?? 0))
}
```

## Possible issues

The result is nullable, so it cannot index the array as it is:

```ant
let xs = [5, 6, 5]
println(xs[xs.lastIndexOf(5)]) // error: cannot use int | null where int is expected
```

Test it against `null`, not for truth. `0` is a valid index, and an `int | null` is not a condition:

```ant
let i = [5, 6].lastIndexOf(5)
if i { // error: cannot use int | null where bool is expected
  println(i)
}
```

```ant
let i = [5, 6].lastIndexOf(5)
i != null // true
i         // 0
```

`NaN` is not `==` to itself, so `Array.lastIndexOf` never finds it. Use
[`Array.findLastIndex`](findLastIndex.md) with `Math.isNaN`:

```ant
let xs = [Math.NaN, 1.0, Math.NaN, 0.0]
xs.lastIndexOf(Math.NaN)            // null
xs.findLastIndex({ Math.isNaN(x) }) // 2
xs.lastIndexOf(-0.0)                // 3
```

Numeric types never mix, so looking for a `float64` in an `Array[int]` is a compile error:

```ant
let counts = [1, 2, 1]
println(counts.lastIndexOf(1.0)) // error: cannot use float64 where int is expected
```

The same holds the other way. Write `2.0`: an integer literal is always an `int`.

```ant
let prices = [1.5, 2.0, 1.5]
println(prices.lastIndexOf(2)) // error: cannot use int where float64 is expected
```

`Array.lastIndexOf` takes a value, not a predicate. To test elements with a function, use
[`Array.findLastIndex`](findLastIndex.md):

```ant
let counts = [1, 2, 3]
println(counts.lastIndexOf({ x > 1 }))
// error: a lambda needs an expected function type, but int is expected
```

Functions compare by identity, so two lambdas with the same body are different values:

```ant
let inc: func(int): int = { x + 1 }
let same: func(int): int = { x + 1 }
[inc, same, inc].lastIndexOf(inc)  // 2
[inc, same, inc].lastIndexOf(same) // 1
```

## See also

[Array.indexOf](indexOf.md) · [Array.findLastIndex](findLastIndex.md) ·
[Array.findLast](findLast.md) · [Array.contains](contains.md) · [Array.slice](slice.md) ·
[String.lastIndexOf](../String/lastIndexOf.md) · [== operator](../../language/operators/equality.md)
