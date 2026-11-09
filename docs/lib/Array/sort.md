<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.sort

```ant signature
func Array.sort[T: Ordered](xs: Array[T]): Array[T]
```

Array.sort returns a new array holding the elements of xs in ascending order. The sort is stable.

<!-- end generated -->

## Details

- Returns a new array; `xs` is unchanged.
- The sort is stable: equal elements keep their order, `-0.0` and `0.0` included.
- Strings sort by code point, and a `NaN` goes after every number.
- To sort by a key, or in another order, use [`Array.sortBy`](sortBy.md) or
  [`Array.sortWith`](sortWith.md).

## Basic examples

```ant
Array.sort([3, 1, 2])                // [1, 2, 3]
Array.sort(["pear", "apple", "fig"]) // ["apple", "fig", "pear"]
```

Call it through `.` on any array:

```ant
var prices = [9.99, 4.5, 12.0]
prices.sort() // [4.5, 9.99, 12.0]
```

An empty array gives an empty array:

```ant
var empty: Array[int] = []
empty.sort() // []
```

## Scope

Strings compare by code point, so every uppercase ASCII letter comes before every lowercase one:

```ant
["pear", "apple", "Fig", "Apple"].sort() // ["Apple", "Fig", "apple", "pear"]
```

Digits in strings compare one code point at a time, not as numbers:

```ant
["10", "9", "100"].sort() // ["10", "100", "9"]
```

Letters outside ASCII sort by code point too, after every ASCII letter. The letter "é", U+00E9, is
above "z":

```ant
["z", "é", "e"].sort() // ["e", "z", "é"]
```

Every numeric type works, the sized ones included:

```ant
[byte(200), byte(3), byte(17)].sort()    // [3, 17, 200]
[int8(-5), int8(100), int8(-100)].sort() // [-100, -5, 100]
[float32(1.5), float32(-2.0)].sort()     // [-2.0, 1.5]
```

Infinities take their place as IEEE 754 orders them, and a `NaN` goes after every number:

```ant
[2.5, Math.NaN, -1.0, Math.Infinity, -Math.Infinity].sort() // [-Infinity, -1.0, 2.5, Infinity, NaN]
```

`-0.0` and `0.0` are equal to the sort, so each stays where it was among the two:

```ant
[0.0, -0.0].sort()            // [0.0, -0.0]
[1.0, 0.0, -0.0, -1.0].sort() // [-1.0, 0.0, -0.0, 1.0]
```

Duplicates stay, side by side:

```ant
[2, 1, 2, 1].sort() // [1, 1, 2, 2]
```

The keys of a map and the members of a set come as arrays, so `Array.sort` takes them:

```ant
Map{ b: 2, c: 3, a: 1 }.keys().sort() // ["a", "b", "c"]
Set.of([30, 10, 20]).values().sort()  // [10, 20, 30]
```

In a pipe, the array is the first argument:

```ant
"kiwi fig apple" | String.split(" ") | Array.sort() // ["apple", "fig", "kiwi"]
```

## Applications

Find the median of a list of numbers:

```ant
func median(xs: Array[float64]): float64 | null {
  if xs.isEmpty() {
    return null
  }
  var s = xs.sort()
  var mid = s.length() / 2
  if s.length() % 2 == 1 {
    s[mid]
  } else {
    (s[mid - 1] + s[mid]) / 2.0
  }
}

println(median([5.0, 1.0, 3.0]))
println(median([4.0, 1.0, 3.0, 2.0]))
```

```text
3.0
2.5
```

Take the three highest scores, highest first:

```ant
var scores = [70, 92, 85, 64, 99]
scores.sort().reverse().slice(0, 3) // [99, 92, 85]
```

List the distinct words of a text in order:

```ant
"the cat and the hat" | String.split(" ") | Array.unique() | Array.sort() // ["and", "cat", "hat", "the"]
```

Dates written as zero-padded YYYY-MM-DD sort as text in date order:

```ant
["2026-03-01", "2025-12-31", "2026-11-08"].sort() // ["2025-12-31", "2026-03-01", "2026-11-08"]
```

Print the entries of a map with its keys in alphabetical order:

```ant
var counts = Map{ pear: 2, apple: 5, fig: 1 }
for k in counts.keys().sort() {
  println("${k}: ${counts[k]}")
}
```

```text
apple: 5
fig: 1
pear: 2
```

## Properties and relations

Each element of the result is at most the next one, and the result holds each element of `xs` as
many times as `xs` does:

```ant
var xs = [5, 3, 8, 1, 3]
var s = xs.sort()
assert(Array.range(1, s.length()).every({ s[x - 1] <= s[x] }))
assert(s.length() == xs.length())
func sameCount(v: int): bool {
  xs.countWhere({ x == v }) == s.countWhere({ x == v })
}
assert(xs.every(sameCount))
```

Sorting twice gives the same result as sorting once, and the order of the input does not matter:

```ant
var xs = ["b", "c", "a"]
assert(xs.sort().sort() == xs.sort())
assert(xs.reverse().sort() == xs.sort())
```

For an array without `NaN`, the first and last elements of the result equal `Array.min` and
`Array.max`:

```ant
var xs = [4, -2, 9, 0]
assert(xs.sort().first() == xs.min())
assert(xs.sort().last() == xs.max())
```

`Array.sort` is `Array.sortBy` with each element as its own key, and `Array.sortWith` with a
comparison that returns a negative number, zero, or a positive one:

```ant
var xs = [3, 1, 2, 1]
assert(xs.sort() == xs.sortBy({ x }))
assert(xs.sort() == xs.sortWith(func(a: int, b: int): int { if a < b { -1 } else if a > b { 1 } else { 0 } }))
```

## Possible issues

`Array.sort` returns a new array and leaves `xs` as it was. A call whose result is dropped does
nothing:

```ant
var xs = [3, 1, 2]
xs.sort()
println(xs)
```

```text
[3, 1, 2]
```

Assign the result to keep it:

```ant
var xs = [3, 1, 2]
xs = xs.sort()
xs // [1, 2, 3]
```

There is no descending flag. Reverse the sorted array, or sort by a negated key with
[`Array.sortBy`](sortBy.md):

```ant
[3, 1, 2].sort().reverse() // [3, 2, 1]
[3, 1, 2].sortBy({ -x })   // [3, 2, 1]
```

Uppercase sorts before lowercase. For an order that ignores case, sort by a case-folded key with
[`String.foldCase`](../String/foldCase.md):

```ant
var names = ["bob", "alice", "Carol"]
names.sort()                         // ["Carol", "alice", "bob"]
names.sortBy({ String.foldCase(x) }) // ["alice", "bob", "Carol"]
```

Nothing normalizes strings before they compare. An "é" written as one code point, U+00E9, sorts
after "z", but the same letter written as "e" and a combining U+0301 sorts with the "e" words:

```ant
var words = ["\u{E9}a", "e\u{301}b", "z"]
words.sort() == ["e\u{301}b", "z", "\u{E9}a"] // true
```

Sort by a key in one normal form with [`String.normalize`](../String/normalize.md). Both spellings
of "é" then sort together, after "z":

```ant
var words = ["\u{E9}a", "e\u{301}b", "z"]
words.sortBy({ String.normalize(x, "NFC") }) == ["z", "\u{E9}a", "e\u{301}b"] // true
```

Numbers held as strings sort as text. Convert them first, or sort by a numeric key:

```ant
["10", "9", "100"].sort()                           // ["10", "100", "9"]
["10", "9", "100"].sortBy({ String.toInt(x) ?? 0 }) // ["9", "10", "100"]
```

An array whose element type is a union of members, such as `int | string`, has no single `T`, so it
is a compile error:

```ant
var xs: Array[int | string] = [3, "a", 1]
println(xs.sort())
// error: T would be int | string, which is not a single member of its bound Ordered; narrow first
```

`null` is not `Ordered`, so an `Array[int | null]` does not sort either:

```ant
var xs: Array[int | null] = [3, null, 1]
println(xs.sort())
// error: T would be int | null, which is not a single member of its bound Ordered; narrow first
```

Drop the nulls first, then sort:

```ant
var xs: Array[int | null] = [3, null, 1]
xs.filter({ x != null }).map({ x ?? 0 }).sort() // [1, 3]
```

`bool` is not `Ordered`:

```ant
println([true, false].sort())
// error: T would be bool, which is not a single member of its bound Ordered; narrow first
```

Records are not `Ordered` either:

```ant
type Player = { name: string, score: int }
var players = [Player{ name: "Ann", score: 7 }, Player{ name: "Bob", score: 3 }]
println(players.sort())
// error: T would be Player, which is not a single member of its bound Ordered; narrow first
```

A named primitive type is not `Ordered`, though `<` works on two of its values:

```ant
type Cents = int
var cs = [Cents(5), Cents(2)]
println(cs.sort())
// error: T would be Cents, which is not a single member of its bound Ordered; narrow first
```

Sort by the base value with [`Array.sortBy`](sortBy.md):

```ant
type Cents = int
var cs = [Cents(5), Cents(2)]
cs.sortBy({ int(x) }) // [Cents(2), Cents(5)]
```

Sort records by a field with [`Array.sortBy`](sortBy.md):

```ant
type Player = { name: string, score: int }
var players = [Player{ name: "Ann", score: 7 }, Player{ name: "Bob", score: 3 }]
players.sortBy({ x.score }).map({ x.name }) // ["Bob", "Ann"]
```

## See also

[Array.sortBy](sortBy.md) · [Array.sortWith](sortWith.md) · [Array.reverse](reverse.md) ·
[Array.min](min.md) · [Array.max](max.md) · [Array.unique](unique.md) ·
[comparison operators](../../language/operators/comparison.md)
