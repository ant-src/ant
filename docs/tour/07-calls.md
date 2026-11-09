<!-- status: reviewed -->

# Calls through . and pipes

Functions live in namespaces such as `Array` and `String`, and you call them by their full name. Two
shorter forms read better when you chain calls: `.` on a value, and the pipe `|`.

```ant
type Point = { x: float64, y: float64 }

func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }

let words = ["ant", "Bee", "cat", "Dove", "eel"]

// A function of a namespace, called by its full name.
println(Array.map(words, { String.toUpper(x) }))

// The same calls through the value: words is an array, x a string.
println(words.map({ x.toUpper() }))

// Calls through . chain from left to right.
println(words.filter({ x.length() == 3 }).map({ x.toLower() }).sort())

// A pipe passes its left side first to any function.
let short = words
  | Array.filter({ x.length() == 3 })
  | String.join(", ")
println(short)

// Your own types work the same way.
let p = Point{ x: 3.0, y: 4.0 }
println(Point.norm(p))
println(p.norm())
```

```text
["ANT", "BEE", "CAT", "DOVE", "EEL"]
["ANT", "BEE", "CAT", "DOVE", "EEL"]
["ant", "bee", "cat", "eel"]
ant, Bee, cat, eel
5.0
5.0
```

`words.map(f)` is `Array.map(words, f)`: `words` is an array, so `.` looks in `Array` and passes
`words` first. In the lambda, `x` is a string, so `x.toUpper()` is `String.toUpper(x)`.

`func Point.norm` declares `norm` on your type `Point`, and `p.norm()` finds it the same way.

`.` only reaches the namespace of the value's own type. [`String.join`](../lib/String/join.md) takes
an array but lives in `String`, so `words.join(", ")` is a compile error:

```ant
let words = ["ant", "bee"]
println(words.join(", ")) // error: Array[string] has no field or function join
```

The pipe has no such limit. `a | f(b)` is `f(a, b)` for any function `f`, so
`words | String.join(", ")` works. A pipeline may continue on the next line when that line starts
with `|`.

The pipe binds looser than every other operator but assignment, so a comparison after a pipe stage
is a compile error:

```ant
let s = " x "
println(s | String.trim() == "x")
// error: '|' binds looser than '==', so parenthesize the pipe, as in (a | f()) == b
```

Put a pipeline in parentheses to use it as an operand: `(s | String.trim()) == "x"`.

Change the program: sort `short` before the join, once by adding a
[`| Array.sort()`](../lib/Array/sort.md) stage and once by calling `.sort()` on the filtered array
before the pipe.

The [functions guide](../guide/functions.md) explains when to pick which form. The reference has
pages for [calls through `.`](../language/operators/dot-calls.md), the
[pipe](../language/operators/pipe.md), [`Array.map`](../lib/Array/map.md), and
[`Array.filter`](../lib/Array/filter.md).

Next: [expressions](08-expressions.md).
