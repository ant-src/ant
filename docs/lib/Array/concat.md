<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.concat

```ant signature
func Array.concat[T](xs: Array[T], ys: Array[T]): Array[T]
```

Array.concat returns a new array holding the elements of xs followed by those of ys.

<!-- end generated -->

## Details

- Returns a new array with the elements of `xs`, then those of `ys`, duplicates kept. Neither
  argument changes.
- To append to `xs` in place, use [Array.extend](extend.md).

## Basic examples

Join two arrays:

```ant
Array.concat([1, 2], [3, 4]) // [1, 2, 3, 4]
```

The elements of `xs` come first:

```ant
Array.concat(["b"], ["a"]) // ["b", "a"]
```

Each side keeps its order, and an element in both stays twice:

```ant
Array.concat([3, 1, 2], [2, 0]) // [3, 1, 2, 2, 0]
```

## Scope

An empty array on either side gives the other one's elements:

```ant
Array.concat([1, 2], []) // [1, 2]
Array.concat([], [1])    // [1]
var none: Array[int] = []
Array.concat(none, none) // []
```

An array joined with itself repeats it:

```ant
var xs = [1, 2]
Array.concat(xs, xs) // [1, 2, 1, 2]
```

Strings, records, and unions with `null` work alike:

```ant
type Point = { x: int, y: int }
Array.concat(["é"], ["日本"])                              // ["é", "日本"]
Array.concat([Point{ x: 1, y: 2 }], [Point{ x: 3, y: 4 }]) // [Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]
Array.concat([1, null], [null])                            // [1, null, null]
```

An expected type fixes `T`, so arrays of different element types can join into a union:

```ant
var mixed: Array[int | string] = Array.concat([1], ["x"])
mixed // [1, "x"]
```

Call it through `.` on an array, or with a pipe, which passes its left side as the first argument:

```ant
[1, 2].concat([3])      // [1, 2, 3]
[1] | Array.concat([2]) // [1, 2]
```

Name the arguments to pass them in any order:

```ant
Array.concat(ys: [2], xs: [1]) // [1, 2]
```

## Applications

Add a header row in front of data rows without touching the data:

```ant
var rows = [["Ann", "31"], ["Bob", "27"]]
var table = Array.concat([["name", "age"]], rows)
for row in table {
  println(String.join(row, ","))
}
```

```text
name,age
Ann,31
Bob,27
```

Put default flags in front of the ones a user gave:

```ant
var defaults = ["--color", "--quiet"]
var extra = ["--verbose"]
println(String.join(Array.concat(defaults, extra), " "))
```

```text
--color --quiet --verbose
```

## Properties and relations

The length of the result is the sum of the lengths:

```ant
var xs = [1, 2, 3]
var ys = [4, 5]
assert(Array.length(Array.concat(xs, ys)) == Array.length(xs) + Array.length(ys))
```

The result is a new array, even when one side is empty:

```ant
var xs = [1, 2]
var ys = Array.concat(xs, [])
assert(ys == xs)
assert(!Array.same(ys, xs))
```

Writing to the result leaves the inputs as they were:

```ant
var xs = [1, 2]
var ys = [3]
var zs = Array.concat(xs, ys)
Array.push(zs, 4)
zs // [1, 2, 3, 4]
xs // [1, 2]
ys // [3]
```

It is the array literal with two spreads:

```ant
var xs = [1, 2]
var ys = [3]
assert(Array.concat(xs, ys) == [...xs, ...ys])
```

It is a copy of `xs` extended with `ys`:

```ant
var xs = [1, 2]
var ys = [3]
var copy = [...xs]
Array.extend(copy, ys)
assert(copy == Array.concat(xs, ys))
```

It is associative, and [Array.flatten](flatten.md) joins any number of arrays the same way:

```ant
var a = [1]
var b = [2]
var c = [3]
assert(Array.concat(Array.concat(a, b), c) == Array.concat(a, Array.concat(b, c)))
assert(Array.flatten([a, b, c]) == Array.concat(Array.concat(a, b), c))
```

## Possible issues

The result holds the elements of `xs` and `ys`, not copies of them. An element that is itself an
array or a record is the same value in both, so writing to it shows through the original:

```ant
var grid = [[1], [2]]
var more = Array.concat(grid, [[3]])
more[0][0] = 99
grid // [[99], [2]]
```

`+` joins strings, not arrays, so it does not concatenate two arrays. Use `Array.concat` or the
spread literal `[...xs, ...ys]`:

```ant
println([1] + [2])
// error: + applies to a single numeric kind, string, or a named primitive type over one, not to Array[int]; narrow first
```

Two arrays of different element types do not join without an expected type that holds both:

```ant
var both = Array.concat([1], ["x"]) // error: cannot use string where int is expected
```

Two empty literals leave `T` open, which is a compile error. Annotate one side, or the result:

```ant
var zs = Array.concat([], [])
// error: cannot infer the element type of an empty array; annotate it
```

Calling `Array.concat` in a loop builds a new array, with every element so far, each time. To grow
one array, use [Array.extend](extend.md) or [Array.push](push.md) instead:

```ant
var all: Array[int] = []
for n in Array.range(0, 3) {
  Array.extend(all, [n, n])
}
println(all)
```

```text
[0, 0, 1, 1, 2, 2]
```

## See also

[Array.extend](extend.md), [Array.push](push.md), [Array.flatten](flatten.md),
[Array.flatMap](flatMap.md), [Array.slice](slice.md), [Array.same](same.md),
[String.concat](../String/concat.md), [Bytes.concat](../Bytes/concat.md)
