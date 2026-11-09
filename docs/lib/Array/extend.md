<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.extend

```ant signature
func Array.extend[T](xs: Array[T], ys: Array[T]): void
```

Array.extend appends the elements of ys to the end of xs, in order, as ys was at the call, so
Array.extend(xs, xs) doubles xs. It mutates xs.

<!-- end generated -->

## Details

- Changes `xs` in place and returns `void`; `ys` is unchanged. To get a new array instead, use
  [Array.concat](concat.md).
- Appends the elements of `ys` as they were at the call, so `Array.extend(xs, xs)` doubles `xs`.

## Basic examples

Append the elements of one array to another:

```ant
var xs = [1, 2]
Array.extend(xs, [3, 4])
xs // [1, 2, 3, 4]
```

Append the array another call returns:

```ant
var ids = [1, 2]
Array.extend(ids, Array.range(10, 13))
ids // [1, 2, 10, 11, 12]
```

## Scope

An empty `ys` leaves `xs` as it was, and an empty `xs` ends up with the elements of `ys`:

```ant
var xs = [1, 2]
Array.extend(xs, [])
xs // [1, 2]
var empty: Array[int] = []
Array.extend(empty, [5, 6])
empty // [5, 6]
```

An array extended with itself doubles, because the call reads `ys` as it was at the start:

```ant
var xs = [1, 2]
Array.extend(xs, xs)
xs // [1, 2, 1, 2]
```

Strings, records, and `null` elements append alike:

```ant
type Point = { x: int, y: int }
var words = ["é"]
Array.extend(words, String.chars("日本"))
words // ["é", "日", "本"]
var points = [Point{ x: 1, y: 2 }]
Array.extend(points, [Point{ x: 3, y: 4 }])
points // [Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]
var maybe: Array[int | null] = [1]
Array.extend(maybe, [null, 2])
maybe // [1, null, 2]
```

A `let` fixes the binding, not the elements, so a `let` array can be extended:

```ant
let xs = [1]
Array.extend(xs, [2])
xs // [1, 2]
```

Call it through `.` on an array, or with a pipe, which passes its left side as the first argument:

```ant
var xs = [1]
xs.extend([2])
xs | Array.extend([3])
xs // [1, 2, 3]
```

Name the arguments to pass them in any order:

```ant
var xs = [1]
Array.extend(ys: [2], xs: xs)
xs // [1, 2]
```

## Applications

Collect the words of several lines into one array:

```ant
var lines = ["the quick fox", "jumps over", "the dog"]
var words: Array[string] = []
for line in lines {
  Array.extend(words, String.split(line, " "))
}
println(len(words))
println(String.join(words, ","))
```

```text
7
the,quick,fox,jumps,over,the,dog
```

Build a command line from fixed parts and optional ones:

```ant
func command(verbose: bool, files: Array[string]): Array[string] {
  var args = ["test"]
  if verbose {
    Array.extend(args, ["-v", "-count", "1"])
  }
  Array.extend(args, files)
  args
}

println(String.join(command(true, ["a.ant", "b.ant"]), " "))
println(String.join(command(false, []), " "))
```

```text
test -v -count 1 a.ant b.ant
test
```

## Properties and relations

The length grows by the length of `ys`:

```ant
var xs = [1, 2, 3]
var ys = [4, 5]
var before = len(xs)
Array.extend(xs, ys)
assert(len(xs) == before + len(ys))
```

It gives in place what [Array.concat](concat.md) returns as a new array:

```ant
var xs = [1, 2]
var ys = [3]
var joined = Array.concat(xs, ys)
Array.extend(xs, ys)
assert(xs == joined)
```

It is [Array.push](push.md) of each element of `ys`, in order:

```ant
var a = [1]
var b = [1]
var ys = [2, 3]
Array.extend(a, ys)
for y in ys {
  Array.push(b, y)
}
assert(a == b)
```

`xs` stays the same array, so every other name for it sees the new elements:

```ant
var xs = [1]
var alias = xs
Array.extend(xs, [2])
assert(Array.same(alias, xs))
alias // [1, 2]
```

## Possible issues

`Array.extend` returns `void`, not the extended array. To get a new array and keep `xs`, use
[Array.concat](concat.md) or the spread literal `[...xs, ...ys]`:

```ant
var xs = [1]
var ys = [2]
var all = [...xs, ...ys]
xs                   // [1]
all                  // [1, 2]
Array.concat(xs, ys) // [1, 2]
```

The elements are appended, not copied. An element that is itself an array or a record is the same
value in both arrays, so writing to it shows through both:

```ant
var grid = [[1]]
var more = [[2]]
Array.extend(grid, more)
more[0][0] = 9
grid // [[1], [9]]
```

A `for` loop over `xs` iterates by index while the index is below the current length, so it also
visits the elements that `Array.extend` appends during the loop. Extending `xs` with each element it
visits never ends. Here a counter stops it after three calls:

```ant
var xs = [1, 2]
var n = 0
for x in xs {
  if n < 3 {
    Array.extend(xs, [x * 10])
  }
  n = n + 1
}
n  // 5
xs // [1, 2, 10, 20, 100]
```

Loop over a copy to visit only the elements that were there at the start, or collect into a new
array and join it with [Array.concat](concat.md):

```ant
var xs = [1, 2]
for x in [...xs] {
  Array.extend(xs, [x * 10])
}
xs // [1, 2, 10, 20]
var ys = [1, 2]
var more: Array[int] = []
for y in ys {
  Array.push(more, y * 10)
}
Array.concat(ys, more) // [1, 2, 10, 20]
```

`ys` is an array. To append one value, use [Array.push](push.md):

```ant
var xs = [1]
Array.extend(xs, 2) // error: cannot use int where Array[int] is expected
```

An `Array[int]` does not extend an `Array[int | null]`, even though each `int` fits, because array
types are invariant. Spell the element type where `ys` is built:

```ant
var xs: Array[int | null] = [1]
var ys = [2]
Array.extend(xs, ys) // error: cannot use Array[int] where Array[int | null] is expected
```

```ant
var xs: Array[int | null] = [1]
var ys: Array[int | null] = [2]
Array.extend(xs, ys)
xs // [1, 2]
```

## See also

[Array.push](push.md), [Array.concat](concat.md), [Array.insert](insert.md),
[Array.flatten](flatten.md), [Array.pop](pop.md), [Array.same](same.md)
