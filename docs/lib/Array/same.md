<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.same

```ant signature
func Array.same[T](xs: Array[T], ys: Array[T]): bool
```

Array.same reports whether xs and ys are the same array, not merely equal ones.

<!-- end generated -->

## Details

- Returns `true` only when `xs` and `ys` are one array, even when two different arrays hold `==`
  elements.
- `==` compares elements, so use `Array.same` to tell a shared array from a copy.

## Basic examples

An array is the same as itself, and as any other name for it:

```ant
var xs = [1, 2]
var ys = xs
Array.same(xs, xs) // true
Array.same(xs, ys) // true
```

An equal array built separately is not the same array:

```ant
var xs = [1, 2]
xs == [1, 2]           // true
Array.same(xs, [1, 2]) // false
```

Call it through `.` on any array:

```ant
var xs = ["a"]
var ys = xs
xs.same(ys) // true
```

## Scope

Two empty arrays are equal, and still two arrays:

```ant
var a: Array[string] = []
var b: Array[string] = []
a == b           // true
Array.same(a, b) // false
```

A spread copies the elements into a new array, and so does a library function that returns an array,
such as `Array.slice` over the whole array or `Array.reverse`:

```ant
var xs = [3, 1, 2]
Array.same(xs, [...xs])                // false
Array.same(xs, xs.slice(0, len(xs)))   // false
Array.same(xs, xs.reverse().reverse()) // false
```

A copy is shallow. The inner arrays of a copied grid are the same arrays as before:

```ant
var grid = [[1, 2], [3]]
var copy = [...grid]
Array.same(grid, copy)       // false
Array.same(grid[0], copy[0]) // true
```

An array held in a record field is shared by a spread of the record, since the spread copies the
field, which is a reference:

```ant
type Post = { title: string, tags: Array[string] }
var p = Post{ title: "Ants", tags: ["insects"] }
var q = Post{ ...p, title: "More ants" }
Array.same(p.tags, q.tags) // true
```

An array passed to a function is the same array inside it, so the function can change it:

```ant
func addOne(xs: Array[int]): Array[int] {
  xs.push(1)
  xs
}
var xs = [0]
Array.same(addOne(xs), xs) // true
xs                         // [0, 1]
```

The arrays may hold any element type, such as records, or functions:

```ant
type Point = { x: int, y: int }
var path = [Point{ x: 0, y: 0 }]
var route = path
Array.same(path, route) // true
var fs: Array[func(int): int] = [{ x + 1 }]
Array.same(fs, fs) // true
```

In a pipeline, the left side is `xs`:

```ant
var xs = [1]
var ys = xs
xs | Array.same(ys) // true
```

Name the arguments, in any order:

```ant
var xs = [1]
Array.same(xs: xs, ys: xs)  // true
Array.same(ys: xs, xs: [1]) // false
```

## Applications

Find rows of a grid that share one array, as `Array.filled` gives when its value is an array, since
it puts the same value in every slot:

```ant
func sharedRows(grid: Array[Array[string]]): bool {
  for i in Array.range(0, len(grid)) {
    for j in Array.range(i + 1, len(grid)) {
      if Array.same(grid[i], grid[j]) {
        return true
      }
    }
  }
  false
}

sharedRows(Array.filled(2, ["."]))           // true
sharedRows(Array.range(0, 2).map({ ["."] })) // false
```

Guard a function that empties one array into another against being given one array twice:

```ant
func moveAll(from: Array[string], to: Array[string]) {
  if Array.same(from, to) {
    return
  }
  to.extend(from)
  while from.pop() != null {}
}

var inbox = ["a", "b"]
var done = ["z"]
moveAll(inbox, done)
println("${inbox} ${done}")
moveAll(done, done)
println(done)
```

```text
[] ["z", "a", "b"]
["z", "a", "b"]
```

## Properties and relations

`Array.same` is reflexive and symmetric:

```ant
var xs = [1, 2]
var ys = xs
assert(Array.same(xs, xs))
assert(Array.same(xs, ys) == Array.same(ys, xs))
```

The same array is `==` to itself unless it holds a `NaN`, since floats compare with IEEE 754 rules:

```ant
var xs = [Math.NaN]
xs == xs           // false
Array.same(xs, xs) // true
```

When two names refer to the same array, a change through one shows through the other. For a copy it
does not:

```ant
var xs = [1]
var ys = xs
var zs = [...xs]
xs.push(2)
assert(Array.same(xs, ys) && ys == [1, 2])
assert(!Array.same(xs, zs) && zs == [1])
```

`Reflect.same` reports the same for two reflected arrays:

```ant
var xs = [1]
var t = type(Array[int])
assert(Reflect.same(Reflect.valueOf(t, xs), Reflect.valueOf(t, xs)))
assert(!Reflect.same(Reflect.valueOf(t, xs), Reflect.valueOf(t, [1])))
```

## Possible issues

`==` compares contents, so it cannot tell whether two names share an array. Use `Array.same` when
that is the question:

```ant
var xs = [1]
var copy = [...xs]
xs == copy           // true
Array.same(xs, copy) // false
```

Both arguments have the one type `Array[T]`. An `Array[int]` and an `Array[int | null]` are
different types, so comparing them is a compile error:

<!-- test: error cannot use Array[int | null] where Array[int] is expected -->

```ant
var xs = [1]
var ys: Array[int | null] = [1]
println(Array.same(xs, ys))
```

A nullable array, such as a value read from a map, must be narrowed first:

<!-- test: error cannot use Array[int] | null where Array[T] is expected -->

```ant
var xs = [1]
var m = Map{ a: xs }
println(Array.same(m["a"], xs))
```

```ant
var xs = [1]
var m = Map{ a: xs }
var found = m["a"]
found != null && Array.same(found, xs) // true
```

## See also

[Record.same](../Record/same.md) · [Map.same](../Map/same.md) · [Set.same](../Set/same.md) ·
[Reflect.same](../Reflect/same.md) · [Array.filled](filled.md) · [Array.slice](slice.md) ·
[== operator](../../language/operators/equality.md)
