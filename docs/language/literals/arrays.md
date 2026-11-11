<!-- status: reviewed -->

# Array literals

Square brackets build an array: `[1, 2, 3]` is an `Array[int]`. A `...` before an element spreads
another array's elements into the new one.

## Details

- Each evaluation of a literal builds a new array. Elements evaluate left to right.
- Without an expected type, the element type is the union of the elements' types, so `[1, "a"]` is
  an `Array[int | string]`.
- An empty literal `[]` needs an expected type, such as an annotation or a parameter type.
- A spread `...xs` takes an array only. It copies the elements into the new array, but an element
  that is an array, a record, a map or a set is still shared.

## Basic examples

```ant
[1, 2, 3]       // [1, 2, 3]
["ant", "bee"]  // ["ant", "bee"]
[1, 2].length() // 2
```

An empty array takes its element type from an annotation:

```ant
let names: Array[string] = []
names.push("Ann")
names // ["Ann"]
```

A spread copies the elements of another array in place:

```ant
let low = [1, 2]
let high = [8, 9]
[...low, 5, ...high] // [1, 2, 5, 8, 9]
```

## Scope

Elements of different types make a union element type:

```ant
let mixed = [1, "two", null]
mixed[1] is string // true
mixed              // [1, "two", null]
```

An expected type checks each element against it instead. An `int` is not a `float64`, so a float
array writes every element as a float:

```ant
let weights: Array[float64] = [1.0, 2.5]
weights.sum() // 3.5
```

An annotation with a union element type lets the array hold values of every member later:

```ant
let ids: Array[int | string] = [1, 2]
ids.push("u-3")
ids // [1, 2, "u-3"]
```

Arrays nest. An inner `[]` takes its type from the outer one's expected type:

```ant
let grid: Array[Array[int]] = [[], [1, 2]]
grid[1][0] // 1
grid       // [[], [1, 2]]
```

Elements can be records, or values of any other type:

```ant
type Point = { lat: int, lon: int }
let path = [Point{ lat: 0, lon: 0 }, Point{ lat: 3, lon: 4 }]
path.map({ x.lat + x.lon }) // [0, 7]
```

A lambda element needs an expected function type, which an annotation gives:

```ant
let steps: Array[func(int): int] = [{ x + 1 }, { x * 2 }]
steps.map({ x(5) }) // [6, 10]
```

A long literal goes one element per line, with a trailing comma:

```ant
let planets = [
  "Mercury",
  "Venus",
  "Earth",
]
println(planets.length())
```

```text
3
```

The elements evaluate left to right:

```ant
func note(n: int): int {
  println("element ${n}")
  n
}
let xs = [note(1), note(2), note(3)]
println(xs)
```

```text
element 1
element 2
element 3
[1, 2, 3]
```

## Applications

A spread widens an array to a union element type, as a new array:

```ant
let ints = [1, 2]
let cells: Array[int | string] = [...ints, "total"]
cells // [1, 2, "total"]
```

Prepend and append without changing the original:

```ant
let path = ["usr", "local"]
let full = ["", ...path, "bin"] | String.join("/")
full // "/usr/local/bin"
path // ["usr", "local"]
```

Build a fresh default in a function, since every call evaluates the literal again:

```ant
func emptyStack(): Array[int] { [] }
let a = emptyStack()
let b = emptyStack()
a.push(1)
b // []
```

## Properties and relations

A spread of one array is a new array: equal, but separate:

```ant
let xs = [1, 2, 3]
let copy = [...xs]
assert(copy == xs)
copy.push(4)
assert(xs == [1, 2, 3])
```

Binding an array to a second name does not copy it, so both names see a change:

```ant
let xs = [1, 2]
let same = xs
same.push(3)
assert(xs == [1, 2, 3])
```

Two literals with the same elements are `==`:

```ant
assert([1, [2, 3]] == [1, [2, 3]])
assert([1, 2] != [2, 1])
```

## Possible issues

An empty literal with no expected type is a compile error. Annotate the name:

```ant
let queue = [] // error: cannot infer the element type of an empty array; annotate it
```

An `int` and a `float64` in one literal make an `Array[int | float64]`, not an array of floats. An
element then does not fit where a number of one type is expected:

```ant
let readings = [1, 2.5]
let first: float64 = readings[0] // error: cannot use int | float64 where float64 is expected
```

A spread takes only an array, not a string, a map, or a sequence. Convert first, with
`String.chars`, `Map.keys`, or `Iter.toArray`:

```ant
let word = "ant"
let letters = ["<", ...word] // error: a spread in an array literal needs an array, got string
```

```ant
let word = "ant"
["<", ...word.chars()] // ["<", "a", "n", "t"]
```

An array literal is not a set: duplicates stay. Use `Set.of` to keep each value once:

```ant
[1, 1, 2]                // [1, 1, 2]
Set.of([1, 1, 2]).size() // 2
```

A spread copies only the outer array. Inner arrays stay shared:

```ant
let grid = [[1], [2]]
let copy = [...grid]
copy[0].push(9)
grid // [[1, 9], [2]]
```

Reading past the end of an array panics:

```ant
[1, 2, 3][3] // panic: index 3 out of range for an array of length 3
```

## See also

[Array](../types/Array.md) · [Array functions](../../lib/Array/index.md) · [Map literals](maps.md) ·
[Set literals](sets.md) · [Record literals](records.md) · [Indexing](../operators/indexing.md) ·
[Collections tour](../../tour/05-collections.md)
