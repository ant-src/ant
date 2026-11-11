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
[1, 2].concat([]) // [1, 2]
[].concat([1])    // [1]
let none: Array[int] = []
none.concat(none) // []
```

An array joined with itself repeats it:

```ant
let xs = [1, 2]
xs.concat(xs) // [1, 2, 1, 2]
```

Strings, records, and unions with `null` work alike:

```ant
type Point = { x: int, y: int }
["é"].concat(["日本"])                              // ["é", "日本"]
[Point{ x: 1, y: 2 }].concat([Point{ x: 3, y: 4 }]) // [Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]
[1, null].concat([null])                            // [1, null, null]
```

An expected type fixes `T`, so arrays of different element types can join into a union:

```ant
let mixed: Array[int | string] = Array.concat([1], ["x"])
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
let rows = [["Ann", "31"], ["Bob", "27"]]
let table = [["name", "age"]].concat(rows)
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
let defaults = ["--color", "--quiet"]
let extra = ["--verbose"]
println(defaults.concat(extra) | String.join(" "))
```

```text
--color --quiet --verbose
```

## Properties and relations

The length of the result is the sum of the lengths:

```ant
let xs = [1, 2, 3]
let ys = [4, 5]
assert(xs.concat(ys).length() == xs.length() + ys.length())
```

The result is a new array, even when one side is empty:

```ant
let xs = [1, 2]
let ys = xs.concat([])
assert(ys == xs)
assert(!Array.same(ys, xs))
```

Writing to the result leaves the inputs as they were:

```ant
let xs = [1, 2]
let ys = [3]
let zs = xs.concat(ys)
zs.push(4)
zs // [1, 2, 3, 4]
xs // [1, 2]
ys // [3]
```

It is the array literal with two spreads:

```ant
let xs = [1, 2]
let ys = [3]
assert(xs.concat(ys) == [...xs, ...ys])
```

It is a copy of `xs` extended with `ys`:

```ant
let xs = [1, 2]
let ys = [3]
let copy = [...xs]
copy.extend(ys)
assert(copy == xs.concat(ys))
```

It is associative, and [Array.flatten](flatten.md) joins any number of arrays the same way:

```ant
let a = [1]
let b = [2]
let c = [3]
assert(Array.concat(Array.concat(a, b), c) == Array.concat(a, Array.concat(b, c)))
assert(Array.flatten([a, b, c]) == Array.concat(Array.concat(a, b), c))
```

## Possible issues

The result holds the elements of `xs` and `ys`, not copies of them. An element that is itself an
array or a record is the same value in both, so writing to it shows through the original:

```ant
let grid = [[1], [2]]
let more = grid.concat([[3]])
more[0][0] = 99
grid // [[99], [2]]
```

`+` joins strings, not arrays, so it does not concatenate two arrays. Use `Array.concat` or the
spread literal `[...xs, ...ys]`:

```ant
println([1] + [2])
// error: + applies to a single numeric kind, string, or a named primitive type over one, not to Array[int]
```

Two arrays of different element types do not join without an expected type that holds both:

```ant
let both = Array.concat([1], ["x"]) // error: cannot use string where int is expected
```

Two empty literals leave `T` open, which is a compile error. Annotate one side, or the result:

```ant
let zs = Array.concat([], [])
// error: cannot infer the element type of an empty array; annotate it
```

Calling `Array.concat` in a loop builds a new array, with every element so far, each time. To grow
one array, use [Array.extend](extend.md) or [Array.push](push.md) instead:

```ant
let all: Array[int] = []
for n in Array.range(0, 3) {
  all.extend([n, n])
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
