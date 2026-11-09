<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.sortBy

```ant signature
func Array.sortBy[T, K: Ordered](xs: Array[T], key: func(T): K): Array[T]
```

Array.sortBy returns a new array holding the elements of xs in ascending order of key. The sort is
stable.

<!-- end generated -->

## Details

- Returns a new array; `xs` is unchanged.
- Calls `key` once per element, from first to last.
- The sort is stable: elements with equal keys keep their order, `-0.0` and `0.0` included.
- String keys sort by code point, and a `NaN` key goes after every number key.

## Basic examples

Sort words by length:

```ant
Array.sortBy(["pear", "fig", "banana"], { String.length(x) }) // ["fig", "pear", "banana"]
```

Sort records by a field:

```ant
type Person = { name: string, age: int }
var people = [Person{ name: "Cy", age: 40 }, Person{ name: "Ann", age: 31 }]
people.sortBy({ x.age }).map({ x.name }) // ["Ann", "Cy"]
```

## Scope

An empty array sorts to an empty array:

```ant
var empty: Array[int] = []
empty.sortBy({ x }) // []
```

Elements with equal keys keep their order, so `"Bob"` stays after `"Ann"`:

```ant
type Person = { name: string, age: int }
var people = [
  Person{ name: "Cy", age: 40 },
  Person{ name: "Ann", age: 31 },
  Person{ name: "Bob", age: 31 },
]
people.sortBy({ x.age }).map({ x.name }) // ["Ann", "Bob", "Cy"]
```

A string key sorts uppercase ASCII letters before lowercase ones. Sort by `String.foldCase` to
ignore case:

```ant
["b", "B", "a", "A"].sortBy({ x })                           // ["A", "B", "a", "b"]
["Banana", "apple", "cherry"].sortBy({ String.foldCase(x) }) // ["apple", "Banana", "cherry"]
```

An accented letter sorts by its code point, after every ASCII letter:

```ant
["é", "z", "e", "Z"].sortBy({ x }) // ["Z", "e", "z", "é"]
```

Negate a numeric key to sort in descending order:

```ant
[3, 1, 2].sortBy({ -x }) // [3, 2, 1]
```

The keys `-0.0` and `0.0` are equal, so their elements keep their order in `xs`:

```ant
[0.0, -0.0, -1.0].sortBy({ x }) // [-1.0, 0.0, -0.0]
[-0.0, 0.0, -1.0].sortBy({ x }) // [-1.0, -0.0, 0.0]
```

A `NaN` key goes last, after every number, in either direction:

```ant
[3.0, 0.0 / 0.0, 1.0, -2.0].sortBy({ x }) // [-2.0, 1.0, 3.0, NaN]
[2.0, 0.0 / 0.0, 1.0].sortBy({ -x })      // [2.0, 1.0, NaN]
```

The elements can be arrays themselves:

```ant
[[1, 2, 3], [1], [1, 2]].sortBy({ x.length() }) // [[1], [1, 2], [1, 2, 3]]
```

The element type may include `null`. Give the key a value for it with `??`:

```ant
var scores: Array[int | null] = [3, null, 1]
scores.sortBy({ x ?? -1 }) // [null, 1, 3]
```

Pass a named function instead of a lambda:

```ant
func vowels(s: string): int {
  String.graphemes(s).countWhere({ "aeiou".contains(x) })
}

["queue", "sky", "area"].sortBy(vowels) // ["sky", "area", "queue"]
```

A key can use the local variables around it. Sort numbers by their distance from a target:

```ant
var target = 10
[3, 12, 7, 20].sortBy({ Math.abs(x - target) }) // [12, 7, 3, 20]
```

Call it through `.` or in a pipe:

```ant
["bb", "a", "ccc"].sortBy({ String.length(x) })         // ["a", "bb", "ccc"]
["bb", "a", "ccc"] | Array.sortBy({ String.length(x) }) // ["a", "bb", "ccc"]
```

## Applications

List the keys of a map from the smallest value to the largest, through its entries:

```ant
var stock = Map{ "nut": 3, "bolt": 12, "gear": 7 }
stock.entries().sortBy({ x.value }).map({ x.key }) // ["nut", "gear", "bolt"]
```

Count words and print them from the most frequent, with `Array.groupBy`:

```ant
var words = String.split("the ant and the bee and the ant", " ")
for entry in words.groupBy({ x }).entries().sortBy({ -x.value.length() }) {
  println("${entry.key}: ${entry.value.length()}")
}
```

```text
the: 3
ant: 2
and: 2
bee: 1
```

Sort by two keys: sort by the second key first, then by the first. The second sort is stable, so it
keeps the order of the first among equal keys:

```ant
type Person = { name: string, age: int }
var people = [
  Person{ name: "Cy", age: 31 },
  Person{ name: "Bob", age: 40 },
  Person{ name: "Ann", age: 31 },
]
people.sortBy({ x.name }).sortBy({ x.age }).map({ x.name }) // ["Ann", "Cy", "Bob"]
```

Take the names of the fastest runs:

```ant
type Run = { name: string, ms: float64 }

func fastest(runs: Array[Run], n: int): Array[string] {
  var sorted = runs.sortBy({ x.ms })
  sorted.slice(0, Math.min(n, sorted.length())).map({ x.name })
}

var runs = [
  Run{ name: "parse", ms: 12.5 },
  Run{ name: "check", ms: 30.0 },
  Run{ name: "lex", ms: 4.0 },
  Run{ name: "run", ms: 9.75 },
]
println(fastest(runs, 3))
println(fastest([], 3))
```

```text
["lex", "run", "parse"]
[]
```

`key` runs once per element, from the first to the last. A key that logs what it sees shows the
order:

```ant
var seen: Array[string] = []
["pear", "fig", "kiwi"].sortBy({ Array.push(seen, x); String.length(x) }) // ["fig", "pear", "kiwi"]
seen                                                                      // ["pear", "fig", "kiwi"]
```

## Properties and relations

With the identity as the key, `Array.sortBy` gives what `Array.sort` gives:

```ant
var xs = [4, 9, 2, 9, 1]
assert(xs.sortBy({ x }) == xs.sort())
```

The result has the same length as `xs`, its keys are in ascending order, and `xs` is unchanged:

```ant
var xs = [5, -3, 8, -1]
var ys = xs.sortBy({ Math.abs(x) })
assert(ys.length() == xs.length())
assert(xs == [5, -3, 8, -1])
for i in Array.range(1, ys.length()) {
  assert(Math.abs(ys[i - 1]) <= Math.abs(ys[i]))
}
```

When no key is `NaN`, the first element of the result is what `Array.minBy` returns, since both keep
the first of equal keys:

```ant
var words = ["fig", "kiwi", "pea", "plum"]
assert(words.sortBy({ String.length(x) }).first() == words.minBy({ String.length(x) }))
```

`Array.sortWith` with a comparison of the keys gives the same order:

```ant
var words = ["pear", "fig", "banana", "kiwi"]
var byLength = words.sortWith(func(a: string, b: string) { String.length(a) - String.length(b) })
assert(words.sortBy({ String.length(x) }) == byLength)
```

## Possible issues

`Array.sortBy` returns a new array and does not sort `xs` in place. A call whose result is unused
changes nothing:

```ant
var xs = [3, 1, 2]
xs.sortBy({ x })
println(xs)
xs = xs.sortBy({ x })
println(xs)
```

```text
[3, 1, 2]
[1, 2, 3]
```

The key must be ordered. A `bool` or a record is not, which is a compile error:

```ant
println([1, 2, 3].sortBy({ x > 1 }))
// error: K would be bool, which is not a single member of its bound Ordered; narrow first
```

```ant
type Point = { x: int, y: int }
var points = [Point{ x: 2, y: 1 }, Point{ x: 1, y: 5 }]
println(points.sortBy({ x }))
// error: K would be Point, which is not a single member of its bound Ordered; narrow first
```

Return one field as the key instead, or compare whole records with `Array.sortWith`.

A key that can be `null` has the union type `int | null`, which is not a single member of `Ordered`:

```ant
var scores: Array[int | null] = [3, null, 1]
println(scores.sortBy({ x }))
// error: K would be int | null, which is not a single member of its bound Ordered; narrow first
```

Give `null` a value with `??`, as in `scores.sortBy({ x ?? -1 })`.

Negating a key panics when the negation overflows, as it does for any nonzero unsigned key. Use
`Array.sortWith` for descending order on an unsigned key:

```ant
var sizes: Array[uint8] = [uint8(1), uint8(2)]
println(sizes.sortBy({ -x })) // panic: integer overflow
```

Convert the keys to `int` before subtracting, because `b - a` on unsigned values panics when
`b < a`:

```ant
var sizes: Array[uint8] = [uint8(1), uint8(2)]
sizes.sortWith(func(a: uint8, b: uint8) { int(b) - int(a) }) // [2, 1]
```

A key of a named primitive type is not a member of `Ordered`, because a named primitive is
assignable only to itself, never to its base:

```ant
type Celsius = float64
var temps = [Celsius(3.0), Celsius(1.0)]
println(temps.sortBy({ x }))
// error: K would be Celsius, which is not a single member of its bound Ordered; narrow first
```

Unwrap the key with its base type: `temps.sortBy({ float64(x) })`.

When the first key is `NaN`, the first element of the result differs from what `Array.minBy`
returns: `Array.sortBy` puts the `NaN` key last, while `Array.minBy` keeps the first element unless
a later key is less by `<`:

```ant
[0.0 / 0.0, 2.0, 1.0].sortBy({ x }).first() // 1.0
[0.0 / 0.0, 2.0, 1.0].minBy({ x })          // NaN
```

Negating a key works only for a number; a string key has no negation. For descending order, write a
comparison for `Array.sortWith`. `Array.reverse` of an ascending sort gives descending order too,
but it also reverses the order of equal keys, so the sort is no longer stable:

```ant
func longerFirst(a: string, b: string): int {
  String.length(b) - String.length(a)
}

var words = ["bb", "a", "cc"]
words.sortBy({ String.length(x) }).reverse() // ["cc", "bb", "a"]
words.sortWith(longerFirst)                  // ["bb", "cc", "a"]
```

## See also

[Array.sort](sort.md) · [Array.sortWith](sortWith.md) · [Array.reverse](reverse.md) ·
[Array.minBy](minBy.md) · [Array.maxBy](maxBy.md) · [Array.groupBy](groupBy.md) ·
[Map.entries](../Map/entries.md)
