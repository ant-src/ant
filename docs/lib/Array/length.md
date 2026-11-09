<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.length

```ant signature
func Array.length[T](xs: Array[T]): int
```

Array.length returns the number of elements in xs.

<!-- end generated -->

## Details

- Returns the length at the time of the call; [Array.push](push.md), [Array.pop](pop.md), and the
  other mutating functions change it.
- `len(xs)` returns the same count.

## Basic examples

```ant
Array.length([10, 20, 30]) // 3
Array.length(["ant"])      // 1
```

Call it through `.` on any array:

```ant
let colors = ["red", "green", "blue"]
colors.length() // 3
```

## Scope

An empty array has length `0`. Annotate it so that `T` is known:

```ant
let empty: Array[string] = []
empty.length() // 0
```

Each element counts once, whatever it holds. A nested array is one element, and `null` is an element
like any other:

```ant
[[1, 2], [3]].length() // 2
let slots: Array[int | null] = [null, 7, null]
slots.length() // 3
```

An element of an `Array[string]` is a whole string, so the length counts strings, not characters:

```ant
let words = ["naïve", "🐜", ""]
words.length() // 3
```

Records and union members are one element each:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
let shapes: Array[Shape] = [Circle{ r: 1.0 }, Square{ side: 2.0 }, Circle{ r: 0.5 }]
shapes.length() // 3
```

In a pipeline, the array is the first argument:

```ant
"the quick brown fox" | String.split(" ") | Array.length() // 4
```

`Array.length` is a function value. Pass it to `Array.map` to get the length of each inner array:

```ant
[[1, 2], [3], []].map(Array.length) // [2, 1, 0]
```

The length follows the array as it grows and shrinks:

```ant
let stack = [1, 2]
stack.push(3)
stack.length() // 3
stack.pop()    // 3
stack.length() // 2
```

## Applications

Read the last element by index:

```ant
let queue = ["a", "b", "c"]
queue[queue.length() - 1] // "c"
```

Number the lines of a list with their indices:

```ant
let tasks = ["write", "test", "ship"]
for i in Array.range(0, tasks.length()) {
  println("${i + 1}. ${tasks[i]}")
}
```

```text
1. write
2. test
3. ship
```

Compute an average, guarding the empty case:

```ant
func average(xs: Array[float64]): float64 | null {
  if xs.length() == 0 {
    return null
  }
  var total = 0.0
  for x in xs {
    total += x
  }
  total / float64(xs.length())
}

println(average([2.0, 4.0, 9.0]))
println(average([]))
```

```text
5.0
null
```

Report a count with the right plural:

```ant
func summary(files: Array[string]): string {
  let n = files.length()
  if n == 1 { "1 file" } else { "${n} files" }
}

println(summary(["a.ant"]))
println(summary(["a.ant", "b.ant"]))
```

```text
1 file
2 files
```

## Properties and relations

`Array.length`, `len`, and `Array.isEmpty` agree:

```ant
for xs in [[1, 2, 3], [], [0]] {
  assert(xs.length() == len(xs))
  assert(xs.isEmpty() == (xs.length() == 0))
}
```

`Array.push` adds one to the length and `Array.pop` on a non-empty array takes one away.
`Array.concat` adds the lengths:

```ant
let xs = [1, 2, 3]
let ys = [4, 5]
let n = xs.length()
xs.push(9)
assert(xs.length() == n + 1)
xs.pop()
assert(xs.length() == n)
assert(xs.concat(ys).length() == xs.length() + ys.length())
```

The functions that keep every element keep the length, and a filter and its complement split it:

```ant
let xs = [5, 3, 8, 1]
assert(xs.map({ x * 2 }).length() == xs.length())
assert(xs.sort().length() == xs.length())
assert(xs.reverse().length() == xs.length())
let even = xs.filter({ x % 2 == 0 })
let odd = xs.filter({ x % 2 != 0 })
assert(even.length() + odd.length() == xs.length())
assert(even.length() == xs.countWhere({ x % 2 == 0 }))
```

`Array.range(0, n)` and `Array.filled(n, v)` have length `n`:

```ant
for n in [0, 1, 5] {
  assert(Array.range(0, n).length() == n)
  assert(Array.filled(n, "x").length() == n)
}
```

## Possible issues

The last valid index is the length minus one. Indexing at the length panics:

```ant
let xs = [1, 2, 3]
xs[xs.length()] // panic: index 3 out of range for an array of length 3
```

`length` is a function, not a field. Without parentheses it is a compile error:

```ant
let xs = [1, 2, 3]
println(xs.length) // error: Array[int] is not a record, so it has no field length
```

`Array.length` takes only arrays. A map, a set, or a string is a compile error; use
[`Map.size`](../Map/size.md), [`Set.size`](../Set/size.md), [`String.length`](../String/length.md),
or [`len`](../prelude/len.md), which takes them all:

```ant
let ages = Map{ ada: 36 }
println(Array.length(ages)) // error: cannot use Map[string, int] where Array[T] is expected
```

A nullable array has no length until it is narrowed:

```ant
let xs: Array[int] | null = null
println(xs.length())
// error: Array[int] | null has no field or function length on every member: narrow the value first
```

Narrow it, or use `?.` with a default:

```ant
let xs: Array[int] | null = null
xs?.length() ?? 0 // 0
```

An array is a reference value: assigning or passing one shares it. A push through one name changes
the length seen through the other:

```ant
let a = [1, 2]
let b = a
b.push(3)
a.length() // 3
```

A `for x in xs` loop reads the length before each step, not once at the start, so elements pushed
during the loop are visited. An unconditional `xs.push(x)` in the loop body never ends:

```ant
let xs = [1, 2]
for x in xs {
  if x < 3 {
    xs.push(x + 1)
  }
}
xs // [1, 2, 2, 3, 3]
```

To fix the count when the loop starts, read the length into a variable first, or loop over
`Array.range(0, xs.length())`:

```ant
let xs = [1, 2]
for i in Array.range(0, xs.length()) {
  xs.push(xs[i] + 1)
}
xs // [1, 2, 2, 3]
```

## See also

[Array.isEmpty](isEmpty.md) · [Array.countWhere](countWhere.md) · [Array.range](range.md) ·
[len](../prelude/len.md) · [String.length](../String/length.md) · [Bytes.length](../Bytes/length.md)
· [Map.size](../Map/size.md) · [Set.size](../Set/size.md) · [Channel.length](../Channel/length.md) ·
[Indexing](../../language/operators/indexing.md)
