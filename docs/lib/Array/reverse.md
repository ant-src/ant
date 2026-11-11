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
none.reverse()  // []
["a"].reverse() // ["a"]
```

Records, `null`, and nested arrays move as whole elements:

```ant
type Point = { x: int, y: int }
[Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }].reverse() // [Point{ x: 3, y: 4 }, Point{ x: 1, y: 2 }]
[1, null, 3].reverse()                               // [3, null, 1]
[[1, 2], [3]].reverse()                              // [[3], [1, 2]]
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
let reversed = "noe\u{301}l" | String.graphemes() | Array.reverse() | String.join("")
reversed == "le\u{301}on" // true
```

Check whether a word is a palindrome:

```ant
func isPalindrome(word: string): bool {
  let gs = word.toLower().graphemes()
  gs == gs.reverse()
}

isPalindrome("Level") // true
isPalindrome("Ant")   // false
```

Show the latest entries of a log first:

```ant
let log = ["start", "load config", "listen on :8080"]
for line in log.reverse() {
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
let parts = "usr/local/bin".split("/")
String.join(parts.reverse(), " < ") // "bin < local < usr"
```

## Properties and relations

Reversing twice gives back the original array:

```ant
let xs = [1, 2, 3, 4]
assert(xs.reverse().reverse() == xs)
```

The length stays the same, and element `i` of the result is element `length - 1 - i` of `xs`:

```ant
let xs = [5, 6, 7, 8]
let ys = xs.reverse()
assert(len(ys) == len(xs))
for i in Array.range(0, len(xs)) {
  assert(ys[i] == xs[len(xs) - 1 - i])
}
```

The result is a new array, even when the order cannot change:

```ant
let xs = [1]
let ys = xs.reverse()
assert(ys == xs)
assert(!Array.same(ys, xs))
```

The first element of the result is the last element of `xs`:

```ant
let xs = [1, 2, 3]
assert(xs.reverse().first() == xs.last())
```

Reversing a concatenation reverses each part and swaps them:

```ant
let xs = [1, 2]
let ys = [3, 4, 5]
assert(Array.concat(xs, ys).reverse() == Array.concat(ys.reverse(), xs.reverse()))
```

## Possible issues

`Array.reverse` does not change its argument. A call whose result is dropped does nothing:

```ant
var xs = [1, 2, 3]
xs.reverse()
println(xs)
xs = xs.reverse()
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
let flipped = grid.reverse()
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
String.join("noe\u{301}l".chars().reverse(), "") == "l\u{301}eon" // true
```

Reversing a stable sort also reverses the order of equal keys. To sort in descending order and keep
equal elements in their original order, sort by a negated key or with [Array.sortWith](sortWith.md):

```ant
type Player = { name: string, score: int }
let players = [Player{ name: "Ann", score: 2 }, Player{ name: "Bob", score: 1 }, Player{ name: "Cid", score: 2 }]
players.sortBy({ x.score }).reverse().map({ x.name }) // ["Cid", "Ann", "Bob"]
players.sortBy({ -x.score }).map({ x.name })          // ["Ann", "Cid", "Bob"]
```

The sorts treat `-0.0` and `0.0` as equal and keep them in their original order, so reversing a sort
swaps them too. Dividing by each zero shows which is which:

```ant
let zs = [-0.0, 0.0]
zs.sort().map({ 1.0 / x })           // [-Infinity, Infinity]
zs.sort().reverse().map({ 1.0 / x }) // [Infinity, -Infinity]
```

## See also

[Array.sort](sort.md), [Array.sortBy](sortBy.md), [Array.sortWith](sortWith.md),
[Array.slice](slice.md), [Array.concat](concat.md), [Array.same](same.md),
[String.graphemes](../String/graphemes.md), [String.chars](../String/chars.md)
