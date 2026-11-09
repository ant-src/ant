<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.filter

```ant signature
func Array.filter[T](xs: Array[T], f: func(T): bool): Array[T]
```

Array.filter returns a new array holding the elements of xs that f accepts, in order.

<!-- end generated -->

## Details

- Returns a new array; `xs` is unchanged, even when `f` accepts every element.
- Calls `f` once per element, from first to last, and never stops early.
- On a sequence, `.filter` is [`Iter.filter`](../Iter/filter.md), which is lazy.

## Basic examples

Keep the even numbers:

```ant
Array.filter([1, 2, 3, 4, 5, 6], { x % 2 == 0 }) // [2, 4, 6]
```

Keep the words longer than three characters:

```ant
Array.filter(["ant", "beetle", "fly", "wasp"], { String.length(x) > 3 }) // ["beetle", "wasp"]
```

## Scope

An empty array gives an empty array, and so does a predicate that accepts nothing:

```ant
var empty: Array[int] = []
Array.filter(empty, { x > 0 })     // []
Array.filter([1, 2, 3], { false }) // []
```

Filter the characters of a string by splitting it into graphemes first:

```ant
Array.filter(String.graphemes("héllo wörld"), { x != "l" }) // ["h", "é", "o", " ", "w", "ö", "r", "d"]
```

Test a record through its fields:

```ant
type Task = { title: string, done: bool }
var tasks = [Task{ title: "write", done: true }, Task{ title: "test", done: false }]
Array.filter(tasks, { !x.done }) // [Task{ title: "test", done: false }]
```

Drop the `null` elements of a nullable array:

```ant
var scores: Array[int | null] = [90, null, 75, null]
Array.filter(scores, { x != null }) // [90, 75]
```

Pick one member of a union with `is`:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
var shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
Array.filter(shapes, { x is Circle }) // [Circle{ radius: 1.0 }]
```

Filter by index through `Array.entries`:

```ant
Array.entries(["a", "b", "c", "d"]).filter({ x.index % 2 == 0 }).map({ x.value }) // ["a", "c"]
```

Pass a named function or a function on a type instead of a lambda:

```ant
func isVowel(s: string): bool { String.contains("aeiou", s) }
Array.filter(String.graphemes("anthill"), isVowel) // ["a", "i"]
Array.filter(["a", "", "b"], String.isEmpty)       // [""]
```

An anonymous function names its parameter, which helps when the body nests another lambda:

```ant
var banned = ["spam", "junk"]
var subjects = ["hello", "cheap spam", "lunch?"]
Array.filter(subjects, func(s: string): bool { !banned.some({ String.contains(s, x) }) }) // ["hello", "lunch?"]
```

A closure reads the variables it captures:

```ant
var limit = 10
var prices = [4, 12, 9, 30]
Array.filter(prices, { x <= limit }) // [4, 9]
```

Call it through `.` or in a pipe:

```ant
[1, 2, 3, 4].filter({ x > 2 })         // [3, 4]
[1, 2, 3, 4] | Array.filter({ x < 3 }) // [1, 2]
```

## Applications

Chain `filter` with `map` to select and then transform:

```ant
type User = { name: string, active: bool }
var users = [
  User{ name: "Ann", active: true },
  User{ name: "Bob", active: false },
  User{ name: "Cid", active: true },
]
users.filter({ x.active }).map({ x.name }) // ["Ann", "Cid"]
```

Keep the lines of a text that are not blank or comments:

```ant
var text = "# config\nhost = localhost\n\nport = 8080\n"
String.split(text, "\n").filter({ String.trim(x) != "" && !String.startsWith(x, "#") }) // ["host = localhost", "port = 8080"]
```

Remove one value from an array by keeping everything else:

```ant
var tags = ["red", "green", "blue"]
tags = tags.filter({ x != "green" })
tags // ["red", "blue"]
```

Split an array in two with a predicate and its negation:

```ant
var temps = [18, 25, 31, 12, 27]
var warm = temps.filter({ x >= 25 })
var cool = temps.filter({ x < 25 })
warm // [25, 31, 27]
cool // [18, 12]
```

Report the failed checks of a test run:

```ant
type Check = { name: string, ok: bool }
var checks = [
  Check{ name: "parse", ok: true },
  Check{ name: "format", ok: false },
  Check{ name: "lint", ok: false },
]
for c in checks.filter({ !x.ok }) {
  println("FAIL ${c.name}")
}
```

```text
FAIL format
FAIL lint
```

## Properties and relations

The result is never longer than `xs`, and every element of it passes `f`:

```ant
var xs = [5, 8, 1, 9, 4]
var f: func(int): bool = { x > 4 }
var kept = Array.filter(xs, f)
assert(len(kept) <= len(xs))
assert(Array.every(kept, f))
```

The length of the result is what `Array.countWhere` counts:

```ant
var xs = ["a", "bb", "ccc", "dd"]
assert(len(Array.filter(xs, { String.length(x) == 2 })) == Array.countWhere(xs, { String.length(x) == 2 }))
```

Filtering by `f` and by its negation splits `xs` with nothing lost:

```ant
var xs = [3, 6, 9, 12, 15]
var yes = Array.filter(xs, { x % 2 == 0 })
var no = Array.filter(xs, { x % 2 != 0 })
assert(len(yes) + len(no) == len(xs))
```

Filtering twice by the same predicate changes nothing the second time:

```ant
var xs = [1, 2, 3, 4, 5]
var once = Array.filter(xs, { x % 2 == 1 })
assert(Array.filter(once, { x % 2 == 1 }) == once)
```

The result is a new array even when it holds every element, so changing it leaves `xs` as it was:

```ant
var xs = [1, 2, 3]
var all = Array.filter(xs, { true })
assert(all == xs)
assert(!Array.same(all, xs))
all[0] = 9
assert(xs == [1, 2, 3])
```

`f` runs once for each element, in order, even after it has accepted or rejected some:

```ant
var seen: Array[int] = []
Array.filter([3, 1, 4, 1], { Array.push(seen, x); x > 1 }) // [3, 4]
seen                                                       // [3, 1, 4, 1]
```

A callback that pushes to `xs` does not change what the call visits:

```ant
var xs = [1, 2, 3]
Array.filter(xs, { Array.push(xs, 10); x > 1 }) // [2, 3]
xs                                              // [1, 2, 3, 10, 10, 10]
```

## Possible issues

`Array.filter` returns a new array and leaves `xs` as it was. A call whose result is thrown away
does not change `xs`; assign the result, as in `xs = xs.filter(...)`:

```ant
var xs = [3, 1, 2]
Array.filter(xs, { x > 1 })
println(xs)
```

```text
[3, 1, 2]
```

The predicate must return a `bool`. Ant has no truthy values, so an `int` is a compile error:

<!-- test: error cannot use int where bool is expected -->

```ant
var counts = [0, 2, 0, 5]
var nonZero = Array.filter(counts, { x })
```

The result keeps the element type of `xs`. Filtering out `null` gives an `Array[int | null]`, not an
`Array[int]`:

<!-- test: error cannot use Array[int | null] where Array[int] is expected -->

```ant
var scores: Array[int | null] = [90, null, 75]
var ints: Array[int] = Array.filter(scores, { x != null })
```

Use `Array.flatMap` instead. Inside the lambda the test on `x` narrows it, so the arrays it returns
hold `int`s:

```ant
var scores: Array[int | null] = [90, null, 75]
var ints: Array[int] = Array.flatMap(scores, { if x == null { [] } else { [x] } })
ints // [90, 75]
```

The same holds for a union: `filter` with `is Circle` still gives an `Array[Shape]`, whose fields
are those every member has:

<!-- test: error Shape has no field radius on every member -->

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
var shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
var radii = shapes.filter({ x is Circle }).map({ x.radius })
```

Use `flatMap` with `is` to get an `Array[Circle]`:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
var shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
var circles: Array[Circle] = shapes.flatMap({ if x is Circle { [x] } else { [] } })
circles.map({ x.radius }) // [1.0]
```

## See also

[Array.find](find.md), [Array.countWhere](countWhere.md), [Array.every](every.md),
[Array.some](some.md), [Array.map](map.md), [Array.entries](entries.md),
[Array.flatMap](flatMap.md), [Array.uniqueBy](uniqueBy.md), [Map.filter](../Map/filter.md),
[Iter.filter](../Iter/filter.md)
