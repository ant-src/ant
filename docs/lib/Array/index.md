<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array

Array holds the functions of the built-in array type Array\[T\], dense and growable. xs\[i\] reads
and xs\[i\] = v writes, both panicking when i is out of range; indices are ints from 0, and ranges
are half-open, from inclusive and to exclusive. A function whose doc says it mutates xs changes it
in place; every other returns a new array.

<!-- end generated -->

The `Array` namespace holds the functions on arrays. Each function that takes an array takes it
first, so `Array.map(xs, f)`, `xs.map(f)`, and `xs | Array.map(f)` are the same call. `Array.range`
and `Array.filled` build an array from scratch and are called only through the namespace.

## Details

- `xs[i]` and `xs[i] = v` panic when `i` is out of range; [Array.get](get.md) returns `null`
  instead. Ranges are half-open: `from` included, `to` excluded.
- A function that mutates `xs`, such as [Array.push](push.md), changes it in place. Every other
  function leaves its arguments untouched and returns a new value.
- An array is a reference: copying a variable shares the array.
- Two arrays are `==` when they have the same length and equal elements in order.

## Basic examples

Square brackets build an array, `xs[i]` reads an element, and `xs[i] = v` writes one:

```ant
let xs = [10, 20, 30]
xs[0]            // 10
len(xs)          // 3
Array.get(xs, 5) // null
xs[1] = 25
xs // [10, 25, 30]
```

`Array.push` grows an array in place and returns nothing:

```ant
let names: Array[string] = []
Array.push(names, "Ada")
Array.push(names, "Linus")
names // ["Ada", "Linus"]
```

Calls chain through `.`, or the array goes down a pipe:

```ant
let scores = [72, 95, 88, 40]
scores.filter({ x >= 50 }).sort()            // [72, 88, 95]
scores | Array.map({ x / 10 }) | Array.sum() // 28
```

## Scope

A spread copies the elements of an array into a new literal:

```ant
let xs = [1, 2]
[0, ...xs, 3] // [0, 1, 2, 3]
```

A `for` loop visits the elements in order:

```ant
let steps = ["mix", "bake"]
for step in steps {
  println(step)
}
```

```text
mix
bake
```

`Array.entries` pairs each element with its index, for a loop that needs both:

```ant
let steps = ["mix", "bake"]
for step in steps.entries() {
  println("${step.index}: ${step.value}")
}
```

```text
0: mix
1: bake
```

Elements may be records, `null` where the type allows it, or other arrays:

```ant
type Point = { x: int, y: int }
[Point{ x: 1, y: 2 }] // [Point{ x: 1, y: 2 }]
let maybe: Array[int | null] = [1, null]
maybe                   // [1, null]
[[1, 2], [3]].flatten() // [1, 2, 3]
```

## Functions by job

### Size and access

- [Array.length](length.md): the number of elements
- [Array.isEmpty](isEmpty.md): whether there are none
- [Array.get](get.md): element `i`, or `null` when out of range
- [Array.first](first.md), [Array.last](last.md): the first or last element, or `null`

### Changing an array in place

- [Array.push](push.md): append a value
- [Array.extend](extend.md): append the elements of another array
- [Array.pop](pop.md): remove and return the last element
- [Array.insert](insert.md): put a value at index `i`
- [Array.removeAt](removeAt.md): remove and return element `i`

### Building arrays

- [Array.range](range.md): the `int`s from `from` up to `to`
- [Array.filled](filled.md): `n` copies of one value
- [Array.concat](concat.md): two arrays, one after the other
- [Array.slice](slice.md): the elements from `from` up to `to`
- [Array.flatten](flatten.md): an array of arrays, concatenated
- [Array.entries](entries.md): each element with its index, as [Indexed](types/Indexed.md)

### Transforming

- [Array.map](map.md): apply a function to each element
- [Array.flatMap](flatMap.md): map to arrays and concatenate them
- [Array.filter](filter.md): the elements a predicate accepts
- [Array.reverse](reverse.md): the elements in reverse order
- [Array.unique](unique.md), [Array.uniqueBy](uniqueBy.md): the first of each equal value or key

### Searching

- [Array.find](find.md), [Array.findLast](findLast.md): the first or last element a predicate
  accepts
- [Array.findIndex](findIndex.md), [Array.findLastIndex](findLastIndex.md): its index
- [Array.indexOf](indexOf.md), [Array.lastIndexOf](lastIndexOf.md): the first or last index of a
  value
- [Array.contains](contains.md): whether a value is an element

### Testing and counting

- [Array.some](some.md): whether a predicate accepts any element
- [Array.every](every.md): whether it accepts every element
- [Array.countWhere](countWhere.md): how many elements it accepts

### Ordering

- [Array.sort](sort.md): ascending, stable
- [Array.sortBy](sortBy.md): ascending by a key
- [Array.sortWith](sortWith.md): by a comparison function
- [Array.min](min.md), [Array.max](max.md): the least or greatest element
- [Array.minBy](minBy.md), [Array.maxBy](maxBy.md): the element with the least or greatest key

### Combining and grouping

- [Array.reduce](reduce.md): fold the elements into one value
- [Array.sum](sum.md): the sum of numbers, or `null` when empty
- [Array.chunk](chunk.md): split into arrays of `n` elements
- [Array.groupBy](groupBy.md): a map from each key to its elements

### Identity

- [Array.same](same.md): whether two arrays are the same array, not only equal

## Applications

Count the words of each length. `Array.groupBy` keeps the keys in first-seen order:

```ant
let words = ["ant", "bee", "wasp", "fly", "moth"]
let groups = words.groupBy({ x.length() })
groups.entries().map({ "${x.key}: ${len(x.value)}" }) // ["3: 3", "4: 2"]
```

Build a list of results in a loop, then summarize it:

```ant
let squares: Array[int] = []
for i in Array.range(1, 5) {
  squares.push(i * i)
}
squares       // [1, 4, 9, 16]
squares.sum() // 30
squares.max() // 16
```

## Properties and relations

`len` and `Array.length` agree, and an array is empty when its length is `0`:

```ant
let xs = [4, 5, 6]
assert(len(xs) == Array.length(xs))
assert(xs.isEmpty() == (len(xs) == 0))
```

A call through `.`, a pipe, and a direct call are the same call:

```ant
let xs = [3, 1, 2]
assert(xs.sort() == Array.sort(xs))
assert((xs | Array.sort()) == Array.sort(xs))
```

## Possible issues

Copying an array variable copies the reference, not the elements. A push through one name shows
through the other:

```ant
let xs = [1, 2]
let ys = xs
ys.push(3)
xs                 // [1, 2, 3]
Array.same(xs, ys) // true
```

A spread into a new literal makes a separate array:

```ant
let xs = [1, 2]
let ys = [...xs]
ys.push(3)
xs           // [1, 2]
xs == [1, 2] // true
```

Only the functions marked `mutates` change the array. `Array.sort` returns a sorted copy and leaves
`xs` as it was:

```ant
let xs = [3, 1, 2]
let sorted = xs.sort()
xs     // [3, 1, 2]
sorted // [1, 2, 3]
```

`let` fixes the binding, not the elements, so a mutating function works on a `let` array:

```ant
let xs = [1, 2]
xs.push(3)
xs // [1, 2, 3]
```

Reading or writing past the end panics. `Array.get` returns `null` for an index that may be out of
range:

```ant
let xs = [1, 2, 3]
xs[3]     // panic: index 3 out of range for an array of length 3
xs.get(3) // null
```

A write past the end panics with the same message:

```ant
let xs = [1, 2, 3]
xs[3] = 1 // panic: index 3 out of range for an array of length 3
```

Assigning the loop variable is a compile error, since it would not change the array. `Array.map`
returns the changed elements:

```ant
let xs = [1, 2, 3]
for x in xs {
  x = x * 2
  // error: cannot assign to x, a for variable; declare a var to hold a value that changes
}
```

```ant
let xs = [1, 2, 3]
xs.map({ x * 2 }) // [2, 4, 6]
```

A `for` loop reads the array as it goes, so it visits an element pushed during the loop. `Array.map`
and the other functions that take a callback read the elements once, at the call:

```ant
let a = [1, 2, 3]
let seen: Array[int] = []
for x in a {
  seen.push(x)
  if x < 3 {
    a.push(x + 10)
  }
}
seen // [1, 2, 3, 11, 12]
let b = [1, 2, 3]
b.map({ b.push(x + 10); x * 2 }) // [2, 4, 6]
b                                // [1, 2, 3, 11, 12, 13]
```

## See also

[Map](../Map/index.md), [Set](../Set/index.md), [Iter](../Iter/index.md),
[Indexed](types/Indexed.md), [Array type](../../language/types/Array.md),
[Array literals](../../language/literals/arrays.md),
[Indexing](../../language/operators/indexing.md), [len](../prelude/len.md),
[Collections](../guides/collections.md)
