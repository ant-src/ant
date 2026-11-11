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
let xs = [1, 2]
Array.extend(xs, [3, 4])
xs // [1, 2, 3, 4]
```

Append the array another call returns:

```ant
let ids = [1, 2]
Array.extend(ids, Array.range(10, 13))
ids // [1, 2, 10, 11, 12]
```

## Scope

An empty `ys` leaves `xs` as it was, and an empty `xs` ends up with the elements of `ys`:

```ant
let xs = [1, 2]
xs.extend([])
xs // [1, 2]
let empty: Array[int] = []
empty.extend([5, 6])
empty // [5, 6]
```

An array extended with itself doubles, because the call reads `ys` as it was at the start:

```ant
let xs = [1, 2]
xs.extend(xs)
xs // [1, 2, 1, 2]
```

Strings, records, and `null` elements append alike:

```ant
type Point = { x: int, y: int }
let words = ["é"]
words.extend("日本".chars())
words // ["é", "日", "本"]
let points = [Point{ x: 1, y: 2 }]
points.extend([Point{ x: 3, y: 4 }])
points // [Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]
let maybe: Array[int | null] = [1]
maybe.extend([null, 2])
maybe // [1, null, 2]
```

A `let` fixes the binding, not the elements, so a `let` array can be extended:

```ant
let xs = [1]
xs.extend([2])
xs // [1, 2]
```

Call it through `.` on an array, or with a pipe, which passes its left side as the first argument:

```ant
let xs = [1]
xs.extend([2])
xs | Array.extend([3])
xs // [1, 2, 3]
```

Name the arguments to pass them in any order:

```ant
let xs = [1]
Array.extend(ys: [2], xs: xs)
xs // [1, 2]
```

## Applications

Collect the words of several lines into one array:

```ant
let lines = ["the quick fox", "jumps over", "the dog"]
let words: Array[string] = []
for line in lines {
  words.extend(line.split(" "))
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
  let args = ["test"]
  if verbose {
    args.extend(["-v", "-count", "1"])
  }
  args.extend(files)
  args
}

println(command(true, ["a.ant", "b.ant"]) | String.join(" "))
println(command(false, []) | String.join(" "))
```

```text
test -v -count 1 a.ant b.ant
test
```

## Properties and relations

The length grows by the length of `ys`:

```ant
let xs = [1, 2, 3]
let ys = [4, 5]
let before = len(xs)
xs.extend(ys)
assert(len(xs) == before + len(ys))
```

It gives in place what [Array.concat](concat.md) returns as a new array:

```ant
let xs = [1, 2]
let ys = [3]
let joined = xs.concat(ys)
xs.extend(ys)
assert(xs == joined)
```

It is [Array.push](push.md) of each element of `ys`, in order:

```ant
let a = [1]
let b = [1]
let ys = [2, 3]
a.extend(ys)
for y in ys {
  b.push(y)
}
assert(a == b)
```

`xs` stays the same array, so every other name for it sees the new elements:

```ant
let xs = [1]
let alias = xs
xs.extend([2])
assert(Array.same(alias, xs))
alias // [1, 2]
```

## Possible issues

`Array.extend` returns `void`, not the extended array. To get a new array and keep `xs`, use
[Array.concat](concat.md) or the spread literal `[...xs, ...ys]`:

```ant
let xs = [1]
let ys = [2]
let all = [...xs, ...ys]
xs            // [1]
all           // [1, 2]
xs.concat(ys) // [1, 2]
```

The elements are appended, not copied. An element that is itself an array or a record is the same
value in both arrays, so writing to it shows through both:

```ant
let grid = [[1]]
let more = [[2]]
grid.extend(more)
more[0][0] = 9
grid // [[1], [9]]
```

A `for` loop over `xs` iterates by index while the index is below the current length, so it also
visits the elements that `Array.extend` appends during the loop. Extending `xs` with each element it
visits never ends. Here a counter stops it after three calls:

```ant
let xs = [1, 2]
var n = 0
for x in xs {
  if n < 3 {
    xs.extend([x * 10])
  }
  n = n + 1
}
n  // 5
xs // [1, 2, 10, 20, 100]
```

Loop over a copy to visit only the elements that were there at the start, or collect into a new
array and join it with [Array.concat](concat.md):

```ant
let xs = [1, 2]
for x in [...xs] {
  xs.extend([x * 10])
}
xs // [1, 2, 10, 20]
let ys = [1, 2]
let more: Array[int] = []
for y in ys {
  more.push(y * 10)
}
ys.concat(more) // [1, 2, 10, 20]
```

`ys` is an array. To append one value, use [Array.push](push.md):

```ant
let xs = [1]
xs.extend(2) // error: cannot use int where Array[int] is expected
```

An `Array[int]` does not extend an `Array[int | null]`, even though each `int` fits, because array
types are invariant. Spell the element type where `ys` is built:

```ant
let xs: Array[int | null] = [1]
let ys = [2]
xs.extend(ys) // error: cannot use Array[int] where Array[int | null] is expected
```

```ant
let xs: Array[int | null] = [1]
let ys: Array[int | null] = [2]
xs.extend(ys)
xs // [1, 2]
```

## See also

[Array.push](push.md), [Array.concat](concat.md), [Array.insert](insert.md),
[Array.flatten](flatten.md), [Array.pop](pop.md), [Array.same](same.md)
