<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.countWhere

```ant signature
func Array.countWhere[T](xs: Array[T], f: func(T): bool): int
```

Array.countWhere returns how many elements of xs f accepts.

<!-- end generated -->

## Details

- An empty array gives `0`.
- Calls `f` once per element, from first to last. Unlike [Array.some](some.md), it never stops
  early.

## Basic examples

Count the numbers above 2:

```ant
Array.countWhere([1, 2, 3, 4, 5], { x > 2 }) // 3
```

Call it through `.` on any array:

```ant
let words = ["ant", "bee", "", "cat", ""]
words.countWhere({ x.isEmpty() }) // 2
```

## Scope

An empty array counts nothing. Annotate it so that `T` is known:

```ant
let empty: Array[int] = []
empty.countWhere({ true }) // 0
```

A named function works as well as a lambda:

```ant
func isEven(n: int): bool {
  n % 2 == 0
}

[1, 2, 3, 4, 5, 6].countWhere(isEven) // 3
```

In a pipeline, the array is the first argument:

```ant
"the cat and the hat" | String.split(" ") | Array.countWhere({ x == "the" }) // 2
```

Count graphemes of a string by splitting it first:

```ant
"🐜a🐜b🐜".graphemes().countWhere({ x == "🐜" }) // 3
```

An element type that includes `null` lets the callback test for it:

```ant
let scores: Array[int | null] = [90, null, 75, null]
scores.countWhere({ x == null })            // 2
scores.countWhere({ x != null && x >= 80 }) // 1
```

The callback can test a record's fields:

```ant
type Task = { title: string, done: bool }
let tasks = [
  Task{ title: "write", done: true },
  Task{ title: "test", done: false },
  Task{ title: "ship", done: true },
]
tasks.countWhere({ x.done }) // 2
```

The callback tells the members of a union apart with `is`:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
let shapes: Array[Shape] = [Circle{ r: 1.0 }, Square{ side: 2.0 }, Circle{ r: 3.0 }]
shapes.countWhere({ x is Circle }) // 2
```

## Applications

Report how many tests passed:

```ant
type Run = { name: string, ok: bool }
let runs = [Run{ name: "add", ok: true }, Run{ name: "sub", ok: false }, Run{ name: "mul", ok: true }]
let passed = runs.countWhere({ x.ok })
println("${passed}/${len(runs)} passed")
```

```text
2/3 passed
```

Count the vowels in a word:

```ant
let vowels = ["a", "e", "i", "o", "u"]
"formicidae".graphemes().countWhere({ vowels.contains(x) }) // 5
```

Tally the cells of each row of a grid that are walls:

```ant
let grid = ["#..#", "....", "###."]
grid.map({ x.graphemes().countWhere({ x == "#" }) }) // [2, 0, 3]
```

## Properties and relations

`Array.countWhere` is the length of what `Array.filter` keeps, and a fold that adds one per accepted
element:

```ant
let xs = [3, 8, 1, 9, 4]
let f = func(n: int): bool { n > 3 }
assert(xs.countWhere(f) == len(xs.filter(f)))
assert(xs.countWhere(f) == xs.reduce(0, func(n: int, v: int): int { if f(v) { n + 1 } else { n } }))
```

The count lies between 0 and the length. It is above 0 exactly when `Array.some` holds, and equal to
the length exactly when `Array.every` holds:

```ant
let xs = [3, 8, 1, 9, 4]
for f in [func(n: int): bool { n > 3 }, func(n: int): bool { n > 0 }, func(n: int): bool { n > 9 }] {
  let c = xs.countWhere(f)
  assert(0 <= c && c <= len(xs))
  assert((c > 0) == xs.some(f))
  assert((c == len(xs)) == xs.every(f))
}
```

The callback sees every element, in order, once. `Array.some` with the same callback stops at the
first element it accepts:

```ant
let seen: Array[int] = []
[5, 1, 7].countWhere({ seen.push(x); x > 3 }) // 2
seen                                          // [5, 1, 7]

let tried: Array[int] = []
[5, 1, 7].some({ tried.push(x); x > 3 }) // true
tried                                    // [5]
```

A predicate and its negation split the array:

```ant
let xs = [3, 8, 1, 9, 4]
assert(xs.countWhere({ x > 3 }) + xs.countWhere({ !(x > 3) }) == len(xs))
```

## Possible issues

The callback must return `bool`. A number is not a truth value:

```ant
println([1, 2, 3].countWhere({ x % 2 })) // error: cannot use int where bool is expected
```

Compare it explicitly:

```ant
[1, 2, 3].countWhere({ x % 2 == 1 }) // 2
```

`Array.countWhere` takes a predicate, not a value. To count the elements `==` to a value, compare in
the callback:

```ant
println([1, 2, 2, 3].countWhere(2)) // error: cannot use int where func(int): bool is expected
```

```ant
[1, 2, 2, 3].countWhere({ x == 2 }) // 2
```

With a nullable element type, `x` may be `null`, so an operator that needs a number is a compile
error. Test for `null` first:

```ant
let xs: Array[int | null] = [1, null, 3]
println(xs.countWhere({ x > 0 }))
// error: > needs both operands of one type, got int | null and int
```

```ant
let xs: Array[int | null] = [1, null, 3]
xs.countWhere({ x != null && x > 0 }) // 2
```

`NaN` is not `==` to itself, so counting with `==` misses it. Test with `Math.isNaN`:

```ant
let xs = [1.0, Math.NaN, Math.NaN]
xs.countWhere({ x == Math.NaN }) // 0
xs.countWhere({ Math.isNaN(x) }) // 2
```

The call counts the elements as they were when it started. Elements the callback pushes are not
counted:

```ant
let xs = [1, 2, 3]
xs.countWhere({ xs.push(x); true }) // 3
xs                                  // [1, 2, 3, 1, 2, 3]
```

## See also

[Array.filter](filter.md) · [Array.some](some.md) · [Array.every](every.md) · [Array.find](find.md)
· [Array.contains](contains.md) · [Array.reduce](reduce.md) · [Array.groupBy](groupBy.md) ·
[Array.length](length.md)
