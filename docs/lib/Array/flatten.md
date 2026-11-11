<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.flatten

```ant signature
func Array.flatten[T](xs: Array[Array[T]]): Array[T]
```

Array.flatten returns a new array holding the elements of each array in xs, concatenated in order.

<!-- end generated -->

## Details

- Removes one level of nesting, joining the inner arrays in order.
- Returns a new array; `xs` and its inner arrays are unchanged.

## Basic examples

Join three arrays into one:

```ant
Array.flatten([[1, 2], [3], [4, 5]]) // [1, 2, 3, 4, 5]
```

Join arrays of strings:

```ant
Array.flatten([["c"], ["a", "b"]]) // ["c", "a", "b"]
```

## Scope

Empty inner arrays add nothing, and an empty outer array gives an empty result:

```ant
[[1], [], [2]].flatten() // [1, 2]
let empties: Array[Array[string]] = [[], []]
empties.flatten() // []
let none: Array[Array[int]] = []
none.flatten() // []
```

Strings and records work like any other element:

```ant
type Point = { x: int, y: int }
[["é"], ["日本", "🐜"]].flatten()                        // ["é", "日本", "🐜"]
[[Point{ x: 1, y: 2 }], [Point{ x: 3, y: 4 }]].flatten() // [Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]
```

Inner arrays that differ only in holding `null` need an expected type. Without one,
`[[1, null], [null]]` would be an `Array[int | null] | Array[null]`, and a union may hold only one
array member:

```ant
let xs: Array[Array[int | null]] = [[1, null], [null]]
xs.flatten() // [1, null, null]
```

Inner arrays of the same type need none:

```ant
[[1, null], [2, null]].flatten() // [1, null, 2, null]
```

An expected type fixes `T`, so inner arrays of different element types flatten into a union:

```ant
let mixed: Array[int | string] = Array.flatten([[1], ["x"]])
mixed // [1, "x"]
```

Only one level is removed. An array of arrays of arrays gives an array of arrays:

```ant
[[[1, 2]], [[3]]].flatten()           // [[1, 2], [3]]
[[[1, 2]], [[3]]].flatten().flatten() // [1, 2, 3]
```

Call it through `.` on an array, or with a pipe:

```ant
[[1], [2, 3]].flatten()         // [1, 2, 3]
[[1], [2, 3]] | Array.flatten() // [1, 2, 3]
```

Name the argument `xs`:

```ant
Array.flatten(xs: [[1], [2]]) // [1, 2]
```

## Applications

List every word of a few lines:

```ant
let lines = ["the ant", "and the bee"]
lines.map({ x.split(" ") }).flatten() // ["the", "ant", "and", "the", "bee"]
```

Collect the values of a map of groups back into one array, group by group:

```ant
let byLetter = ["ant", "bee", "ape"].groupBy({ x.slice(0, 1) })
byLetter.values().flatten() // ["ant", "ape", "bee"]
```

Print every cell of a table, one per line:

```ant
let table = [["name", "age"], ["Ann", "31"]]
for cell in table.flatten() {
  println(cell)
}
```

```text
name
age
Ann
31
```

Build a command line from optional groups of flags:

```ant
func flags(verbose: bool, output: string | null): Array[string] {
  let groups: Array[Array[string]] = [["build"]]
  if verbose {
    groups.push(["-v"])
  }
  if output != null {
    groups.push(["-o", output])
  }
  groups.flatten()
}

println(String.join(flags(true, "out.txt"), " "))
println(String.join(flags(false, null), " "))
```

```text
build -v -o out.txt
build
```

## Properties and relations

The length of the result is the sum of the inner lengths:

```ant
let xss = [[1, 2], [], [3, 4, 5]]
assert(len(xss.flatten()) == (xss.map(len).sum() ?? 0))
```

Flattening two arrays is [Array.concat](concat.md), and flattening any number is the spread literal:

```ant
let a = [1, 2]
let b = [3]
let c = [4]
assert([a, b].flatten() == Array.concat(a, b))
assert([a, b, c].flatten() == [...a, ...b, ...c])
```

The result keeps the order of the inner arrays and of the elements within each, so it matches a
left-to-right [Array.reduce](reduce.md) with [Array.concat](concat.md):

```ant
let xss = [[3, 1], [], [2], [5, 4]]
let empty: Array[int] = []
assert(xss.flatten() == xss.reduce(empty, Array.concat))
```

Flattening after a map is [Array.flatMap](flatMap.md):

```ant
let xs = [1, 2, 3]
assert(xs.map({ [x, x * 10] }).flatten() == xs.flatMap({ [x, x * 10] }))
```

Flattening the chunks of [Array.chunk](chunk.md) gives back the original array:

```ant
let xs = [1, 2, 3, 4, 5]
assert(xs.chunk(2).flatten() == xs)
```

The result is a new array, even when there is one inner array:

```ant
let inner = [1, 2]
let flat = [inner].flatten()
assert(flat == inner)
assert(!Array.same(flat, inner))
```

Writing to the result leaves the inner arrays as they were:

```ant
let grid = [[1], [2]]
let flat = grid.flatten()
flat[0] = 9
flat // [9, 2]
grid // [[1], [2]]
```

## Possible issues

The result holds the elements themselves, not copies. When an element is an array or a record,
writing to it through the result shows through the original:

```ant
let nested = [[[1]], [[2]]]
let flat = nested.flatten()
flat[0][0] = 99
nested // [[[99]], [[2]]]
```

`Array.flatten` takes an array of arrays. A flat array is a compile error:

```ant
let xs = Array.flatten([1, 2]) // error: cannot use Array[int] where Array[Array[T]] is expected
```

Inner arrays of different element types do not flatten without an expected type that holds both:

```ant
let xs = Array.flatten([[1], ["x"]])
// error: the elements do not form a valid type: a union may contain each discriminator at most once: Array[int] and Array[string] are both an array; give the array an expected type
```

Inner arrays that hold `null` in some places and not in others need an expected type too:

```ant
let xs = Array.flatten([[1, null], [null]])
// error: the elements do not form a valid type: a union may contain each discriminator at most once: Array[int | null] and Array[null] are both an array; give the array an expected type
```

An empty literal leaves `T` unsolved. Annotate the argument or the result:

```ant
let xs = Array.flatten([]) // error: cannot infer the element type of an empty array; annotate it
```

It removes one level only, so a deeper array does not fit an expected flat type:

```ant
let xs: Array[int] = Array.flatten([[[1]], [[2]]])
// error: cannot use Array[int] where int is expected
// error: cannot use Array[int] where int is expected
```

## See also

[Array.flatMap](flatMap.md), [Array.concat](concat.md), [Array.extend](extend.md),
[Array.chunk](chunk.md), [Array.groupBy](groupBy.md), [String.concat](../String/concat.md),
[Bytes.concat](../Bytes/concat.md)
