<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.flatten

```ant signature
func Array.flatten[T](xs: Array[Array[T]]): Array[T]
```

Array.flatten returns a new array holding the elements of each array in xs, concatenated in order.

<!-- end generated -->

## Details

- `Array.flatten` takes an array of arrays and removes one level of nesting
  ([Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)).
- `Array.flatten` changes neither `xs` nor its inner arrays, and returns a new array
  ([Standard library](https://github.com/ant-src/spec/blob/master/ant-lang.md#11-standard-library)).
- Every inner array has the same element type `T`. Type arguments are never written at a call site:
  the expected type or the argument fixes `T`, and an empty `[]` with nothing to fix it is a compile
  error ([Generics](https://github.com/ant-src/spec/blob/master/ant-lang.md#generics),
  [Inference](https://github.com/ant-src/spec/blob/master/ant-lang.md#inference)).

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
Array.flatten([[1], [], [2]]) // [1, 2]
var empties: Array[Array[string]] = [[], []]
Array.flatten(empties) // []
var none: Array[Array[int]] = []
Array.flatten(none) // []
```

Strings and records work like any other element:

```ant
type Point = { x: int, y: int }
Array.flatten([["é"], ["日本", "🐜"]])                           // ["é", "日本", "🐜"]
Array.flatten([[Point{ x: 1, y: 2 }], [Point{ x: 3, y: 4 }]]) // [Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]
```

Inner arrays that hold `null` need an expected type. Without one, `[[1, null], [null]]` would be an
`Array[int | null] | Array[null]`, and a union may hold only one array member
([Unions](https://github.com/ant-src/spec/blob/master/ant-lang.md#unions)):

```ant
var xs: Array[Array[int | null]] = [[1, null], [null]]
Array.flatten(xs) // [1, null, null]
```

An expected type fixes `T`, so inner arrays of different element types flatten into a union:

```ant
var mixed: Array[int | string] = Array.flatten([[1], ["x"]])
mixed // [1, "x"]
```

Only one level is removed. An array of arrays of arrays gives an array of arrays:

```ant
Array.flatten([[[1, 2]], [[3]]])                // [[1, 2], [3]]
Array.flatten(Array.flatten([[[1, 2]], [[3]]])) // [1, 2, 3]
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
var lines = ["the ant", "and the bee"]
Array.flatten(lines.map({ String.split(x, " ") })) // ["the", "ant", "and", "the", "bee"]
```

Collect the values of a map of groups back into one array, group by group:

```ant
var byLetter = Array.groupBy(["ant", "bee", "ape"], { String.slice(x, 0, 1) })
Array.flatten(Map.values(byLetter)) // ["ant", "ape", "bee"]
```

Print every cell of a table, one per line:

```ant
var table = [["name", "age"], ["Ann", "31"]]
for cell in Array.flatten(table) {
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
  var groups: Array[Array[string]] = [["build"]]
  if verbose {
    groups.push(["-v"])
  }
  if output != null {
    groups.push(["-o", output])
  }
  Array.flatten(groups)
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
var xss = [[1, 2], [], [3, 4, 5]]
assert(len(Array.flatten(xss)) == (Array.sum(Array.map(xss, len)) ?? 0))
```

Flattening two arrays is [Array.concat](concat.md), and flattening any number is the spread literal:

```ant
var a = [1, 2]
var b = [3]
var c = [4]
assert(Array.flatten([a, b]) == Array.concat(a, b))
assert(Array.flatten([a, b, c]) == [...a, ...b, ...c])
```

Flattening after a map is [Array.flatMap](flatMap.md):

```ant
var xs = [1, 2, 3]
assert(Array.flatten(Array.map(xs, { [x, x * 10] })) == Array.flatMap(xs, { [x, x * 10] }))
```

Flattening the chunks of [Array.chunk](chunk.md) gives back the original array:

```ant
var xs = [1, 2, 3, 4, 5]
assert(Array.flatten(Array.chunk(xs, 2)) == xs)
```

The result is a new array, even when there is one inner array:

```ant
var inner = [1, 2]
var flat = Array.flatten([inner])
assert(flat == inner)
assert(!Array.same(flat, inner))
```

Writing to the result leaves the inner arrays as they were:

```ant
var grid = [[1], [2]]
var flat = Array.flatten(grid)
flat[0] = 9
flat // [9, 2]
grid // [[1], [2]]
```

## Possible issues

The result holds the elements themselves, not copies. When an element is an array or a record,
writing to it through the result shows through the original:

```ant
var nested = [[[1]], [[2]]]
var flat = Array.flatten(nested)
flat[0][0] = 99
nested // [[[99]], [[2]]]
```

`Array.flatten` takes an array of arrays. A flat array is a compile error:

<!-- test: error cannot use Array[int] where Array[Array[T]] is expected -->

```ant
var xs = Array.flatten([1, 2])
```

Inner arrays of different element types do not flatten without an expected type that holds both:

<!-- test: error Array[int] and Array[string] are both an array -->

```ant
var xs = Array.flatten([[1], ["x"]])
```

Inner arrays that hold `null` in some places and not in others need an expected type too:

<!-- test: error Array[int | null] and Array[null] are both an array -->

```ant
var xs = Array.flatten([[1, null], [null]])
```

An empty literal leaves `T` unsolved. Annotate the argument or the result:

<!-- test: error cannot use Array[never] where Array[Array[T]] is expected -->

```ant
var xs = Array.flatten([])
```

It removes one level only, so a deeper array does not fit an expected flat type:

<!-- test: error cannot use Array[int] where int is expected -->

```ant
var xs: Array[int] = Array.flatten([[[1]], [[2]]])
```

## See also

[Array.flatMap](flatMap.md), [Array.concat](concat.md), [Array.extend](extend.md),
[Array.chunk](chunk.md), [Array.groupBy](groupBy.md), [String.concat](../String/concat.md),
[Bytes.concat](../Bytes/concat.md)
