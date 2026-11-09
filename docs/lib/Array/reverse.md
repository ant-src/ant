<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.reverse

```ant signature
func Array.reverse[T](xs: Array[T]): Array[T]
```

Array.reverse returns a new array holding the elements of xs in reverse order.

<!-- end generated -->

## Details

- Returns a new array; `xs` is unchanged.
- The new array holds the same elements, not copies, so a nested array, map, set, or record is
  shared with `xs`.
- To reverse text, reverse its graphemes, not its code points.

## Basic examples

Reverse an array of numbers:

```ant
Array.reverse([1, 2, 3]) // [3, 2, 1]
```

The original array keeps its order:

```ant
let xs = ["a", "b", "c"]
Array.reverse(xs) // ["c", "b", "a"]
xs                // ["a", "b", "c"]
```

## Scope

An empty array and an array of one element come back with the same elements:

```ant
let none: Array[int] = []
Array.reverse(none)  // []
Array.reverse(["a"]) // ["a"]
```

Records, `null`, and nested arrays move as whole elements:

```ant
type Point = { x: int, y: int }
Array.reverse([Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]) // [Point{ x: 3, y: 4 }, Point{ x: 1, y: 2 }]
Array.reverse([1, null, 3])                               // [3, null, 1]
Array.reverse([[1, 2], [3]])                              // [[3], [1, 2]]
```

Call it through `.` on an array, or with a pipe:

```ant
[1, 2, 3].reverse()         // [3, 2, 1]
[1, 2, 3] | Array.reverse() // [3, 2, 1]
```

Name the argument:

```ant
Array.reverse(xs: [1, 2]) // [2, 1]
```

## Applications

Reverse a string by its graphemes, and join them back:

```ant
let reversed = String.join(Array.reverse(String.graphemes("noe\u{301}l")), "")
reversed == "le\u{301}on" // true
```

Check whether a word is a palindrome:

```ant
func isPalindrome(word: string): bool {
  let gs = String.graphemes(String.toLower(word))
  gs == Array.reverse(gs)
}

isPalindrome("Level") // true
isPalindrome("Ant")   // false
```

Show the latest entries of a log first:

```ant
let log = ["start", "load config", "listen on :8080"]
for line in Array.reverse(log) {
  println(line)
}
```

```text
listen on :8080
load config
start
```

Read a path from the leaf up to the root:

```ant
let parts = String.split("usr/local/bin", "/")
String.join(parts.reverse(), " < ") // "bin < local < usr"
```

## Properties and relations

Reversing twice gives back the original array:

```ant
let xs = [1, 2, 3, 4]
assert(Array.reverse(Array.reverse(xs)) == xs)
```

The length stays the same, and element `i` of the result is element `length - 1 - i` of `xs`:

```ant
let xs = [5, 6, 7, 8]
let ys = Array.reverse(xs)
assert(len(ys) == len(xs))
for i in Array.range(0, len(xs)) {
  assert(ys[i] == xs[len(xs) - 1 - i])
}
```

The result is a new array, even when the order cannot change:

```ant
let xs = [1]
let ys = Array.reverse(xs)
assert(ys == xs)
assert(!Array.same(ys, xs))
```

The first element of the result is the last element of `xs`:

```ant
let xs = [1, 2, 3]
assert(Array.first(Array.reverse(xs)) == Array.last(xs))
```

Reversing a concatenation reverses each part and swaps them:

```ant
let xs = [1, 2]
let ys = [3, 4, 5]
assert(Array.reverse(Array.concat(xs, ys)) == Array.concat(Array.reverse(ys), Array.reverse(xs)))
```

## Possible issues

`Array.reverse` does not change its argument. A call whose result is dropped does nothing:

```ant
var xs = [1, 2, 3]
Array.reverse(xs)
println(xs)
xs = Array.reverse(xs)
println(xs)
```

```text
[1, 2, 3]
[3, 2, 1]
```

The result shares its elements with `xs`. Writing into an element that is an array or a record shows
through the original:

```ant
let grid = [[1], [2]]
let flipped = Array.reverse(grid)
flipped[0][0] = 9
grid // [[1], [9]]
```

`Array.reverse` takes an array, not a string; split the string with
[String.graphemes](../String/graphemes.md) first, as below:

```ant
Array.reverse("noel") // error: cannot use string where Array[T] is expected
```

Reversing the code points of a string moves an accent onto another letter. Split with
[String.graphemes](../String/graphemes.md) instead of [String.chars](../String/chars.md), as in
Applications above:

```ant
String.join(Array.reverse(String.chars("noe\u{301}l")), "") == "l\u{301}eon" // true
```

Reversing a stable sort also reverses the order of equal keys. To sort in descending order and keep
equal elements in their original order, sort by a negated key or with [Array.sortWith](sortWith.md):

```ant
type Player = { name: string, score: int }
let players = [Player{ name: "Ann", score: 2 }, Player{ name: "Bob", score: 1 }, Player{ name: "Cid", score: 2 }]
Array.reverse(Array.sortBy(players, { x.score })).map({ x.name }) // ["Cid", "Ann", "Bob"]
Array.sortBy(players, { -x.score }).map({ x.name })               // ["Ann", "Cid", "Bob"]
```

The sorts treat `-0.0` and `0.0` as equal and keep them in their original order, so reversing a sort
swaps them too. Dividing by each zero shows which is which:

```ant
let zs = [-0.0, 0.0]
Array.sort(zs).map({ 1.0 / x })                // [-Infinity, Infinity]
Array.reverse(Array.sort(zs)).map({ 1.0 / x }) // [Infinity, -Infinity]
```

## See also

[Array.sort](sort.md), [Array.sortBy](sortBy.md), [Array.sortWith](sortWith.md),
[Array.slice](slice.md), [Array.concat](concat.md), [Array.same](same.md),
[String.graphemes](../String/graphemes.md), [String.chars](../String/chars.md)
