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
let xs = [1, 2, 4]
Array.insert(xs, 2, 3)
xs // [1, 2, 3, 4]
```

Insert at the front and at the end:

```ant
let letters = ["b", "c"]
Array.insert(letters, 0, "a")
Array.insert(letters, len(letters), "d")
letters // ["a", "b", "c", "d"]
```

## Scope

An empty array takes a value only at index `0`:

```ant
let empty: Array[string] = []
empty.insert(0, "first")
empty // ["first"]
```

Strings, records, and `null` insert alike:

```ant
type Point = { x: int, y: int }
let words = ["日", "語"]
words.insert(1, "本")
words // ["日", "本", "語"]
let path = [Point{ x: 0, y: 0 }, Point{ x: 2, y: 2 }]
path.insert(1, Point{ x: 1, y: 1 })
path // [Point{ x: 0, y: 0 }, Point{ x: 1, y: 1 }, Point{ x: 2, y: 2 }]
let maybe: Array[int | null] = [1, 2]
maybe.insert(1, null)
maybe // [1, null, 2]
```

An array of arrays takes an array as one element:

```ant
let rows = [[1], [3]]
rows.insert(1, [2, 2])
rows // [[1], [2, 2], [3]]
```

A `let` fixes the binding, not the elements, so a `let` array takes inserts:

```ant
let xs = [2]
xs.insert(0, 1)
xs // [1, 2]
```

Call it through `.` on an array, or with a pipe, which passes its left side as the first argument:

```ant
let xs = [10, 30]
xs.insert(1, 20)
xs | Array.insert(0, 0)
xs // [0, 10, 20, 30]
```

Name the arguments to pass them in any order:

```ant
let xs = ["a", "c"]
Array.insert(xs, v: "b", i: 1)
xs // ["a", "b", "c"]
```

## Applications

Keep an array sorted as values arrive, by inserting each before the first larger element:

```ant
func insertSorted(xs: Array[int], v: int) {
  let at = xs.findIndex({ x > v }) ?? len(xs)
  xs.insert(at, v)
}

let scores: Array[int] = []
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
let queue = ["write", "test", "urgent fix", "deploy"]
queue.insert(0, queue.removeAt(2))
queue // ["urgent fix", "write", "test", "deploy"]
```

Put a header row in front of a table built row by row:

```ant
type Row = { name: string, qty: int }
let items = [Row{ name: "ant", qty: 3 }, Row{ name: "bee", qty: 12 }]
let lines = items.map({ "${x.name}\t${x.qty}" })
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
let xs = [1, 2, 3]
let before = len(xs)
xs.insert(1, 9)
assert(len(xs) == before + 1)
assert(xs[1] == 9)
```

It gives in place what slices and a spread literal build as a new array:

```ant
let xs = [1, 2, 3, 4]
let i = 2
let v = 0
let built = [...xs.slice(0, i), v, ...xs.slice(i, len(xs))]
xs.insert(i, v)
assert(xs == built)
```

At index `len(xs)` it is [Array.push](push.md):

```ant
let a = [1, 2]
let b = [1, 2]
a.insert(len(a), 3)
b.push(3)
assert(a == b)
```

[Array.removeAt](removeAt.md) at the same index undoes it and returns `v`:

```ant
let xs = ["a", "b", "c"]
xs.insert(1, "x")
assert(xs.removeAt(1) == "x")
assert(xs == ["a", "b", "c"])
```

`xs` stays the same array, so every other name for it sees the new element:

```ant
let xs = [1, 3]
let alias = xs
xs.insert(1, 2)
assert(Array.same(alias, xs))
alias // [1, 2, 3]
```

## Possible issues

An index past the length panics, and so does a negative one. Index `len(xs)` is the last valid one:

```ant
[1, 2, 3].insert(4, 0)  // panic: Array.insert: the index 4 is not within 0 ..= 3
[1, 2, 3].insert(-1, 0) // panic: Array.insert: the index -1 is not within 0 ..= 3
```

`Array.insert` returns `void`, not the array. Binding its result to a name of an array type is a
compile error:

```ant
let xs = [2, 3]
let ys: Array[int] = xs.insert(0, 1) // error: cannot use void where Array[int] is expected
```

Without the annotation, `ys` is `void` and no error is reported. To get a new array and keep `xs`,
use a spread literal:

```ant
let xs = [2, 3]
let ys = [1, ...xs]
xs // [2, 3]
ys // [1, 2, 3]
```

`v` is one element, so an array of values is a compile error on an `Array[int]`. Spread it into a
new array instead:

```ant
let xs = [1, 4]
xs.insert(1, [2, 3]) // error: cannot use Array[int] where int is expected
```

```ant
let xs = [1, 4]
let ys = [...xs.slice(0, 1), 2, 3, ...xs.slice(1, len(xs))]
ys // [1, 2, 3, 4]
```

A `for` loop over `xs` iterates by index while the index is below the current length. Inserting at
or before the current index shifts the current element up, so the next iteration reads it again. The
loop below never ends without its counter, and `xs` grows on every pass:

```ant
let xs = [1, 2]
var n = 0
for x in xs {
  n = n + 1
  if n > 3 {
    break
  }
  println(x)
  xs.insert(0, x * 10)
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
let xs = [1, 2]
for x in [...xs] {
  xs.insert(0, x * 10)
}
xs // [20, 10, 1, 2]
```

## See also

[Array.push](push.md), [Array.removeAt](removeAt.md), [Array.extend](extend.md),
[Array.slice](slice.md), [Array.findIndex](findIndex.md), [Array.concat](concat.md)
