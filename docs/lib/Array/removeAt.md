<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.removeAt

```ant signature
func Array.removeAt[T](xs: Array[T], i: int): T
```

Array.removeAt removes element i from xs and returns it, shifting later elements down. It mutates
xs. It panics if i is out of range.

<!-- end generated -->

## Details

- Removes the element at `i`, moves every later element down by one, and returns the removed
  element. It changes `xs` in place.
- Panics when `i` is negative or not below the length, so it always panics on an empty array.
- A `for` loop over `xs` does not visit an element removed during the loop.

## Basic examples

Remove the element at index 1. The call returns it:

```ant
let xs = [10, 20, 30]
Array.removeAt(xs, 1) // 20
xs                    // [10, 30]
```

Remove the first element. The others move down, so the old element 1 is now element 0:

```ant
let queue = ["ann", "bob", "cid"]
Array.removeAt(queue, 0) // "ann"
queue[0]                 // "bob"
```

## Scope

Remove the last element with index `length - 1`:

```ant
let xs = [1, 2, 3]
xs.removeAt(len(xs) - 1) // 3
xs                       // [1, 2]
```

Removing the only element leaves an empty array:

```ant
let xs = ["only"]
xs.removeAt(0) // "only"
xs             // []
```

Remove a grapheme from a string by splitting it first and joining the rest:

```ant
let chars = "héllo".graphemes()
chars.removeAt(1)      // "é"
String.join(chars, "") // "hllo"
```

A removed record comes back whole, ready to use:

```ant
type Task = { title: string, done: bool }
let tasks = [Task{ title: "write", done: true }, Task{ title: "test", done: false }]
tasks.removeAt(0).title // "write"
tasks                   // [Task{ title: "test", done: false }]
```

On an array with `null` elements the result is `null` when the removed element is:

```ant
let slots: Array[string | null] = ["a", null, "c"]
slots.removeAt(1) // null
slots             // ["a", "c"]
```

A union element keeps its member, so `is` tells which one was removed:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
shapes.removeAt(1) is Rect // true
```

Call it through `.` or in a pipe:

```ant
let xs = [1, 2, 3, 4]
xs.removeAt(0)         // 1
xs | Array.removeAt(1) // 3
xs                     // [2, 4]
```

## Applications

Remove an element by value: find its index first, and remove only when there is one:

```ant
func removeValue(xs: Array[string], v: string): bool {
  let i = xs.indexOf(v)
  if i == null {
    return false
  }
  xs.removeAt(i)
  true
}

let tags = ["red", "green", "blue"]
removeValue(tags, "green") // true
removeValue(tags, "pink")  // false
tags                       // ["red", "blue"]
```

Move an element to a new position by removing it and inserting it again:

```ant
let playlist = ["intro", "verse", "chorus", "outro"]
playlist.insert(0, playlist.removeAt(2))
playlist // ["chorus", "intro", "verse", "outro"]
```

Serve a queue in order, taking each item from the front:

```ant
let jobs = ["build", "test", "deploy"]
while !jobs.isEmpty() {
  let job = jobs.removeAt(0)
  println("running ${job}")
}
```

```text
running build
running test
running deploy
```

Remove every element that matches by walking the indices from the end, so a removal never moves an
element the loop has yet to see:

```ant
let xs = [1, 4, 6, 7, 8, 3]
var i = len(xs) - 1
while i >= 0 {
  if xs[i] % 2 == 0 {
    xs.removeAt(i)
  }
  i -= 1
}
println(xs)
```

```text
[1, 7, 3]
```

## Properties and relations

The result is the element that was at `i`, and the array is one shorter:

```ant
let xs = [5, 6, 7, 8]
let before = xs[2]
let n = len(xs)
assert(xs.removeAt(2) == before)
assert(len(xs) == n - 1)
```

What is left is the array without element `i`, the same as concatenating the slices on either side:

```ant
let xs = ["a", "b", "c", "d"]
let expected = Array.concat(xs.slice(0, 1), xs.slice(2, len(xs)))
xs.removeAt(1)
assert(xs == expected)
```

`Array.insert` at the same index undoes it:

```ant
let xs = [1, 2, 3]
xs.insert(1, xs.removeAt(1))
assert(xs == [1, 2, 3])
```

Removing at the last index does what `Array.pop` does on a non-empty array:

```ant
let a = [1, 2, 3]
let b = [1, 2, 3]
assert(a.removeAt(len(a) - 1) == b.pop())
assert(a == b)
```

## Possible issues

The index must be below the length. The length itself is out of range:

```ant
[1, 2, 3].removeAt(3) // panic: Array.removeAt: the index 3 is out of range for a length of 3
```

A negative index does not count from the end; it panics:

```ant
[1, 2, 3].removeAt(-1) // panic: Array.removeAt: the index -1 is out of range for a length of 3
```

An empty array has no index to remove. Check the length first, or use `Array.pop`, which returns
`null` instead:

```ant
let empty: Array[int] = []
empty.removeAt(0) // panic: Array.removeAt: the index 0 is out of range for a length of 0
empty.pop()       // null
```

`Array.indexOf` and `Array.findIndex` return `int | null`, and `Array.removeAt` needs an `int`.
Passing the result straight through is a compile error:

```ant
let tags = ["red", "green"]
tags.removeAt(tags.indexOf("green"))
// error: cannot use int | null where int is expected
```

Test for `null` first, as in the applications above.

Removing inside a `for` loop over the same array skips elements. The loop reads by index, and the
element after a removed one moves into the index the loop has already passed:

```ant
let xs = [2, 4, 5]
var i = 0
for x in xs {
  if x % 2 == 0 {
    xs.removeAt(i)
  } else {
    i += 1
  }
}
println(xs)
```

```text
[4, 5]
```

Walk the indices from the end, as in the applications above, or build a new array with
`Array.filter`.

`Array.removeAt` changes the array in place. A second variable that holds the same array sees the
change:

```ant
let original = [1, 2, 3]
let alias = original
alias.removeAt(0) // 1
original          // [2, 3]
```

## See also

[Array.insert](insert.md), [Array.pop](pop.md), [Array.push](push.md), [Array.indexOf](indexOf.md),
[Array.findIndex](findIndex.md), [Array.filter](filter.md), [Array.slice](slice.md),
[Map.remove](../Map/remove.md), [Set.remove](../Set/remove.md)
