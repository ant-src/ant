<!-- status: reviewed -->

# Array

`Array[T]` is an ordered, growable list of values of type `T`, indexed from `0`.

## Details

- An array is a reference value: assigning or passing it shares one array, and a change through one
  name shows through every other. `let` stops the name from being reassigned, not the array from
  changing.
- `==` compares the length and the elements in order. `Array.same` tells whether two names hold one
  array.
- `xs[i]` panics when `i` is out of range, negative included; `xs.get(i)` returns `null` instead.
- The element type is invariant: an `Array[Circle]` is not an `Array[Shape]`, and an `Array[int]` is
  not an `Array[int | null]`. Widen by copying, `[...circles]`.
- A union holds at most one array type, so write `Array[int | string]`, not
  `Array[int] | Array[string]`.

## Basic examples

Annotate an array with `Array[T]`, read an element by index, and ask for its length:

```ant
let primes: Array[int] = [2, 3, 5, 7]
primes[0]       // 2
primes.length() // 4
primes.last()   // 7
```

An array grows in place:

```ant
let queue: Array[string] = []
queue.push("ann")
queue.push("bob")
queue // ["ann", "bob"]
```

Write an element by index:

```ant
let scores = [10, 20, 30]
scores[1] = 25
scores // [10, 25, 30]
```

## Scope

### Element types

An element can be a record, a union, `null`, a function, or another array:

```ant
type Point = { x: int, y: int }
let path = [Point{ x: 0, y: 0 }, Point{ x: 3, y: 4 }]
let cells: Array[int | null] = [1, null, 3]
let grid: Array[Array[int]] = [[1, 2], [3, 4]]
let steps: Array[func(int): int] = [{ x + 1 }, { x * 2 }]
path[1].y                   // 4
cells.filter({ x != null }) // [1, 3]
grid[1][0]                  // 3
steps.map({ x(10) })        // [11, 20]
```

### Looping

`for` visits the elements in order:

```ant
let langs = ["Ant", "Go", "Lua"]
for lang in langs {
  println(lang)
}
```

```text
Ant
Go
Lua
```

`Array.entries` pairs each element with its index:

```ant
let langs = ["Ant", "Go"]
for { index, value } in langs.entries() {
  println("${index}: ${value}")
}
```

```text
0: Ant
1: Go
```

### Length

`len` counts the elements, as `Array.length` does:

```ant
let xs = [4, 5, 6]
len(xs)      // 3
xs.isEmpty() // false
```

### Destructuring

An array destructuring binds elements by position, and `...` binds the rest as a new array:

```ant
let [head, ...tail] = [1, 2, 3]
head // 1
tail // [2, 3]
```

Without a rest, the extra elements are ignored:

```ant
let [first] = [1, 2, 3]
first // 1
```

### Arrays in unions

`is Array` tells an array apart from the other members of a union:

```ant
func describe(v: int | Array[int]): string {
  if v is Array { "${v.length()} numbers" } else { "one number" }
}

describe([1, 2]) // "2 numbers"
describe(7)      // "one number"
```

### Generic functions

A type parameter keeps the element type, so one function serves every array:

```ant
func lastTwo[T](xs: Array[T]): Array[T] {
  let from = if xs.length() < 2 { 0 } else { xs.length() - 2 }
  xs.slice(from, xs.length())
}

lastTwo(["a", "b", "c"]) // ["b", "c"]
lastTwo([1])             // [1]
```

### Rendering

An array renders with square brackets, its strings quoted:

```ant
let words = ["ant", "bee"]
println(words)
println(String.from([1, 2]))
```

```text
["ant", "bee"]
[1, 2]
```

### Map keys and set members

An array compares by its elements, so it can be a map key or a set member:

```ant
let names = Map{ ([0, 0]): "origin", ([1, 0]): "east" }
names[[1, 0]]                   // "east"
Set.of([[1, 2], [1, 2]]).size() // 1
```

A map hashes a key when it is inserted, so do not change an array after you use it as a key or a
member: the entry may become unreachable. Copy it with `[...xs]` first if it keeps changing.

## Applications

### A stack

`push` and `pop` work at the end of the array:

```ant
let stack: Array[string] = []
stack.push("a")
stack.push("b")
stack.pop() // "b"
stack       // ["a"]
```

### Collect into a caller's array

A function changes the array its caller passed, since both hold the same one:

```ant
func record(log: Array[string], event: string) {
  log.push("${log.length() + 1}. ${event}")
}

let log: Array[string] = []
record(log, "start")
record(log, "stop")
println(log)
```

```text
["1. start", "2. stop"]
```

### A grid

Build each row afresh, so the rows are separate arrays:

```ant
let board = Array.range(0, 3) | Array.map({ Array.filled(3, ".") })
board[1][1] = "x"
board.map({ x | String.join("") }) // ["...", ".x.", "..."]
```

## Properties and relations

`==` compares elements; `Array.same` compares identity. A spread makes a copy that is equal but not
the same:

```ant
let xs = [1, 2, 3]
let alias = xs
let copy = [...xs]
assert(xs == copy)
assert(Array.same(xs, alias))
assert(!Array.same(xs, copy))
assert(len(xs) == xs.length())
```

Order matters to `==`:

```ant
assert([1, 2] != [2, 1])
assert([[1], [2]] == [[1], [2]])
```

## Possible issues

### A second name shares the array

Binding an array to a new name does not copy it. Copy with a spread when you need a separate array:

```ant
let original = [1, 2]
let shared = original
let copy = [...original]
shared.push(3)
original // [1, 2, 3]
copy     // [1, 2]
```

### An empty literal needs a type

An empty `[]` has no element to infer the type from, so annotate it:

```ant
let xs = [] // error: cannot infer the element type of an empty array; annotate it
```

```ant
let xs: Array[int] = []
xs.length() // 0
```

### `Array.filled` repeats one value

`Array.filled(n, v)` puts the same `v` in every slot. When `v` is an array, every row is that one
array, and a write to one row shows in all of them:

```ant
let grid = Array.filled(2, Array.filled(3, 0))
grid[0][1] = 5
grid // [[0, 5, 0], [0, 5, 0]]
```

Build each row with `Array.map`, as in [A grid](#a-grid).

### Index out of range

Indexing past the end, or with a negative index, panics. Use `get` when the index may miss:

```ant
let xs = [1, 2]
xs[2]     // panic: index 2 out of range for an array of length 2
xs[-1]    // panic: index -1 out of range for an array of length 2
xs.get(2) // null
```

Writing past the end panics too. Add elements with `push` or `insert`:

```ant
let xs = [1]
xs[1] = 2 // panic: index 1 out of range for an array of length 1
```

### Destructuring too few elements

An array destructuring panics, as indexing does, when the array has fewer elements than the fixed
positions:

```ant
let [first, second] = [1] // panic: a destructuring needs 2 elements, got 1
```

### No widening of the element type

An `Array[Circle]` is not an `Array[Shape]`: a function holding the wider type could store a `Rect`
among the circles. Copy into a new array of the wider type:

```ant
type Circle = { radius: float64 }
type Rect = { w: float64, h: float64 }
type Shape = Circle | Rect
let circles = [Circle{ radius: 1.0 }]
let shapes: Array[Shape] = circles // error: cannot use Array[Circle] where Array[Shape] is expected
```

```ant
type Circle = { radius: float64 }
type Rect = { w: float64, h: float64 }
type Shape = Circle | Rect
let circles = [Circle{ radius: 1.0 }]
let shapes: Array[Shape] = [...circles]
shapes.length() // 1
```

### Two array types in one union

`is Array` cannot tell two array types apart at run time, so a union holds only one. Put the union
inside:

```ant
let ids: Array[int] | Array[string] = [1]
// error: a union may contain each discriminator at most once: Array[int] and Array[string] are both an array; write Array[int | string]
```

### The loop variable is not the element

Assigning the `for` variable would not change the array, so it is a compile error. Write by index,
or build a new array with `map`:

```ant
let xs = [1, 2]
for x in xs { x = x * 2 }
// error: cannot assign to x, a for variable; declare a var to hold a value that changes
```

```ant
let xs = [1, 2]
xs.map({ x * 2 }) // [2, 4]
```

## See also

[Array functions](../../lib/Array/index.md) · [Array literals](../literals/arrays.md) ·
[Indexing](../operators/indexing.md) · [for](../keywords/for.md) · [len](../../lib/prelude/len.md) ·
[Array.same](../../lib/Array/same.md) · [Array.filled](../../lib/Array/filled.md) · [Map](Map.md) ·
[Set](Set.md) · [Collections](../../guide/collections.md)
