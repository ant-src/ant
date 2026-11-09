<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.insert

```ant signature
func Array.insert[T](xs: Array[T], i: int, v: T): void
```

Array.insert inserts v into xs so that it becomes element i, shifting later elements up. It mutates
xs. It panics if i is outside 0 ..= length, so i equal to the length appends.

<!-- end generated -->

## Details

- Changes `xs` in place and returns `void`. The elements at `i` and after move up one index.
- `i` may be any index from `0` to the length, both included; the length appends, like
  [Array.push](push.md). Any other `i`, a negative one included, panics.
- Inserts one element. To insert several, build a new array with a spread literal.

## Basic examples

Insert a value in the middle:

```ant
var xs = [1, 2, 4]
Array.insert(xs, 2, 3)
xs // [1, 2, 3, 4]
```

Insert at the front and at the end:

```ant
var letters = ["b", "c"]
Array.insert(letters, 0, "a")
Array.insert(letters, len(letters), "d")
letters // ["a", "b", "c", "d"]
```

## Scope

An empty array takes a value only at index `0`:

```ant
var empty: Array[string] = []
Array.insert(empty, 0, "first")
empty // ["first"]
```

Strings, records, and `null` insert alike:

```ant
type Point = { x: int, y: int }
var words = ["日", "語"]
Array.insert(words, 1, "本")
words // ["日", "本", "語"]
var path = [Point{ x: 0, y: 0 }, Point{ x: 2, y: 2 }]
Array.insert(path, 1, Point{ x: 1, y: 1 })
path // [Point{ x: 0, y: 0 }, Point{ x: 1, y: 1 }, Point{ x: 2, y: 2 }]
var maybe: Array[int | null] = [1, 2]
Array.insert(maybe, 1, null)
maybe // [1, null, 2]
```

An array of arrays takes an array as one element:

```ant
var rows = [[1], [3]]
Array.insert(rows, 1, [2, 2])
rows // [[1], [2, 2], [3]]
```

A `let` fixes the binding, not the elements, so a `let` array takes inserts:

```ant
let xs = [2]
Array.insert(xs, 0, 1)
xs // [1, 2]
```

Call it through `.` on an array, or with a pipe, which passes its left side as the first argument:

```ant
var xs = [10, 30]
xs.insert(1, 20)
xs | Array.insert(0, 0)
xs // [0, 10, 20, 30]
```

Name the arguments to pass them in any order:

```ant
var xs = ["a", "c"]
Array.insert(xs, v: "b", i: 1)
xs // ["a", "b", "c"]
```

## Applications

Keep an array sorted as values arrive, by inserting each before the first larger element:

```ant
func insertSorted(xs: Array[int], v: int) {
  var at = Array.findIndex(xs, { x > v }) ?? len(xs)
  Array.insert(xs, at, v)
}

var scores: Array[int] = []
for s in [42, 7, 99, 7, 63] {
  insertSorted(scores, s)
}
println(scores)
```

```text
[7, 7, 42, 63, 99]
```

Move an element to a new position with [Array.removeAt](removeAt.md):

```ant
var queue = ["write", "test", "urgent fix", "deploy"]
Array.insert(queue, 0, Array.removeAt(queue, 2))
queue // ["urgent fix", "write", "test", "deploy"]
```

Put a header row in front of a table built row by row:

```ant
type Row = { name: string, qty: int }
var items = [Row{ name: "ant", qty: 3 }, Row{ name: "bee", qty: 12 }]
var lines = items.map({ "${x.name}\t${x.qty}" })
lines.insert(0, "name\tqty")
for line in lines {
  println(line)
}
```

```text
name	qty
ant	3
bee	12
```

## Properties and relations

The length grows by one, and `v` is element `i`:

```ant
var xs = [1, 2, 3]
var before = len(xs)
Array.insert(xs, 1, 9)
assert(len(xs) == before + 1)
assert(xs[1] == 9)
```

It gives in place what slices and a spread literal build as a new array:

```ant
var xs = [1, 2, 3, 4]
var i = 2
var v = 0
var built = [...Array.slice(xs, 0, i), v, ...Array.slice(xs, i, len(xs))]
Array.insert(xs, i, v)
assert(xs == built)
```

At index `len(xs)` it is [Array.push](push.md):

```ant
var a = [1, 2]
var b = [1, 2]
Array.insert(a, len(a), 3)
Array.push(b, 3)
assert(a == b)
```

[Array.removeAt](removeAt.md) at the same index undoes it and returns `v`:

```ant
var xs = ["a", "b", "c"]
Array.insert(xs, 1, "x")
assert(Array.removeAt(xs, 1) == "x")
assert(xs == ["a", "b", "c"])
```

`xs` stays the same array, so every other name for it sees the new element:

```ant
var xs = [1, 3]
var alias = xs
Array.insert(xs, 1, 2)
assert(Array.same(alias, xs))
alias // [1, 2, 3]
```

## Possible issues

An index past the length panics, and so does a negative one. Index `len(xs)` is the last valid one:

```ant
Array.insert([1, 2, 3], 4, 0)  // panic: Array.insert: the index 4 is not within 0..3
Array.insert([1, 2, 3], -1, 0) // panic: Array.insert: the index -1 is not within 0..3
```

`Array.insert` returns `void`, not the array. Binding its result to a name of an array type is a
compile error:

<!-- test: error cannot use void where Array[int] is expected -->

```ant
var xs = [2, 3]
var ys: Array[int] = Array.insert(xs, 0, 1)
```

Without the annotation, `ys` is `void` and no error is reported. To get a new array and keep `xs`,
use a spread literal:

```ant
var xs = [2, 3]
var ys = [1, ...xs]
xs // [2, 3]
ys // [1, 2, 3]
```

`v` is one element, so an array of values is a compile error on an `Array[int]`. Spread it into a
new array instead:

<!-- test: error cannot use Array[int] where int is expected -->

```ant
var xs = [1, 4]
Array.insert(xs, 1, [2, 3])
```

```ant
var xs = [1, 4]
var ys = [...Array.slice(xs, 0, 1), 2, 3, ...Array.slice(xs, 1, len(xs))]
ys // [1, 2, 3, 4]
```

A `for` loop over `xs` iterates by index while the index is below the current length. Inserting at
or before the current index shifts the current element up, so the next iteration reads it again. The
loop below never ends without its counter, and `xs` grows on every pass:

```ant
var xs = [1, 2]
var n = 0
for x in xs {
  n = n + 1
  if n > 3 {
    break
  }
  println(x)
  Array.insert(xs, 0, x * 10)
}
println(xs)
```

```text
1
1
1
[10, 10, 10, 1, 2]
```

Loop over a copy instead:

```ant
var xs = [1, 2]
for x in [...xs] {
  Array.insert(xs, 0, x * 10)
}
xs // [20, 10, 1, 2]
```

## See also

[Array.push](push.md), [Array.removeAt](removeAt.md), [Array.extend](extend.md),
[Array.slice](slice.md), [Array.findIndex](findIndex.md), [Array.concat](concat.md)
