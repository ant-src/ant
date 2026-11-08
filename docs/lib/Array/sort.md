<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.sort

```ant signature
func Array.sort[T: Ordered](xs: Array[T]): Array[T]
```

Array.sort returns a new array holding the elements of xs in ascending order. The sort is stable.

<!-- end generated -->

## Details

- `Array.sort` returns a new array with the elements of `xs` in ascending order. The sort is stable:
  equal elements keep their order. A `NaN` goes after every number
  ([spec: Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)).
- It leaves `xs` unchanged
  ([spec: Standard library](https://github.com/ant-src/spec/blob/master/ant-lang.md#11-standard-library)).
- Numbers compare by value, and strings by code point, as `<` compares them
  ([spec: Expressions](https://github.com/ant-src/spec/blob/master/ant-lang.md#expressions)).
- `T` must be a single member of `Ordered`, which is `Number | string`: one numeric type, or
  `string`. A union such as `Number` or `int | null` is not one, nor is a record or `bool`
  ([spec: Generics](https://github.com/ant-src/spec/blob/master/ant-lang.md#generics)). A named
  primitive type such as `type UserId = string` is not one either: it is assignable only to itself,
  and library functions take the base kinds
  ([spec: Named primitive types](https://github.com/ant-src/spec/blob/master/ant-lang.md#named-primitive-types)).
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

Sort ISO 8601 dates as text, which is date order:

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

Each element of the result is at most the next one, and the result has the same length and the same
elements as `xs`:

```ant
var xs = [5, 3, 8, 1, 3]
var s = xs.sort()
assert(Array.range(1, s.length()).every({ s[x - 1] <= s[x] }))
assert(s.length() == xs.length())
assert(xs.every({ s.contains(x) }))
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

Uppercase sorts before lowercase. For an order that ignores case, sort by a lowercased key:

```ant
var names = ["bob", "alice", "Carol"]
names.sort()                        // ["Carol", "alice", "bob"]
names.sortBy({ String.toLower(x) }) // ["alice", "bob", "Carol"]
```

Numbers held as strings sort as text. Convert them first, or sort by a numeric key:

```ant
["10", "9", "100"].sort()                           // ["10", "100", "9"]
["10", "9", "100"].sortBy({ String.toInt(x) ?? 0 }) // ["9", "10", "100"]
```

An array whose element type is a union, such as `int | null` or `Number`, has no single `T`, so it
is a compile error:

<!-- test: error T would be int | null, which is not a single member of its bound Ordered -->

```ant
var xs: Array[int | null] = [3, null, 1]
println(xs.sort())
```

Records and `bool` are not `Ordered` either:

<!-- test: error T would be Player, which is not a single member of its bound Ordered -->

```ant
type Player = { name: string, score: int }
var players = [Player{ name: "Ann", score: 7 }, Player{ name: "Bob", score: 3 }]
println(players.sort())
```

A named primitive type is not `Ordered`, though `<` works on two of its values:

<!-- test: error T would be Cents, which is not a single member of its bound Ordered -->

```ant
type Cents = int
var cs = [Cents(5), Cents(2)]
println(cs.sort())
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
