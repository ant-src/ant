<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.unique

```ant signature
func Array.unique[T](xs: Array[T]): Array[T]
```

Array.unique returns a new array holding the first of each == value in xs, in order.

<!-- end generated -->

## Details

- Keeps the first of each group of `==` values, in the order of `xs`.
- Every `NaN` is kept, since `NaN` equals nothing; `-0.0 == 0.0`, so only the first of the two is
  kept.
- Returns a new array; `xs` is unchanged.

## Basic examples

```ant
Array.unique([3, 1, 3, 2, 1]) // [3, 1, 2]
```

Call it through `.` on any array:

```ant
let tags = ["go", "ant", "go", "web", "ant"]
tags.unique() // ["go", "ant", "web"]
```

## Scope

An empty array gives an empty array. Annotate it so that `T` is known:

```ant
let empty: Array[int] = []
empty.unique() // []
```

An array without repeats comes back equal, and an array of one value repeated gives one element:

```ant
[1, 2, 3].unique()    // [1, 2, 3]
[7, 7, 7, 7].unique() // [7]
```

Strings compare by code points, so case matters and `==` does not normalize. The two spellings of
"é" stay apart:

```ant
["Ant", "ant", "ANT", "ant"].unique() // ["Ant", "ant", "ANT"]
len(["café", "cafe\u{301}"].unique()) // 2
```

Records compare by tag and fields, so a second record with the same fields is a repeat:

```ant
type Point = { x: int, y: int }
let path = [Point{ x: 0, y: 0 }, Point{ x: 1, y: 0 }, Point{ x: 0, y: 0 }]
path.unique() // [Point{ x: 0, y: 0 }, Point{ x: 1, y: 0 }]
```

Members of a tagged union compare by tag too, so equal fields under different tags both stay:

```ant
type Circle = { size: float64 }
type Square = { size: float64 }
type Shape = Circle | Square
let shapes: Array[Shape] = [Circle{ size: 1.0 }, Square{ size: 1.0 }, Circle{ size: 1.0 }]
shapes.unique() // [Circle{ size: 1.0 }, Square{ size: 1.0 }]
```

Nested arrays compare element by element, and two maps or two sets are equal in any order:

```ant
[[1, 2], [2, 1], [1, 2]].unique()                          // [[1, 2], [2, 1]]
[Map{ a: 1, b: 2 }, Map{ b: 2, a: 1 }].unique()            // [Map{ "a": 1, "b": 2 }]
[Set.of(["x", "y"]), Set.of(["y", "x"])].unique().length() // 1
```

`null` is equal to `null`, so an array whose element type includes `null` keeps one:

```ant
let answers: Array[string | null] = ["yes", null, "no", null, "yes"]
answers.unique() // ["yes", null, "no"]
```

In a pipeline, the array is the first argument:

```ant
"b a c a b" | String.split(" ") | Array.unique() // ["b", "a", "c"]
```

## Applications

List the distinct words of a sentence in the order they first appear:

```ant
let text = "the ant and the bee and the fly"
String.split(text, " ").unique() // ["the", "ant", "and", "bee", "fly"]
```

Count the distinct visitors in a log:

```ant
let visits = ["ann", "bob", "ann", "cid", "bob", "ann"]
len(visits.unique()) // 3
```

Merge two lists of email addresses without repeats, keeping the order of the first list:

```ant
let team = ["ann@ant.dev", "bob@ant.dev"]
let guests = ["cid@ant.dev", "ann@ant.dev"]
Array.concat(team, guests).unique() // ["ann@ant.dev", "bob@ant.dev", "cid@ant.dev"]
```

Collect the grid cells a path visits, each once:

```ant
type Cell = { row: int, col: int }

func visited(moves: string): Array[Cell] {
  var at = Cell{ row: 0, col: 0 }
  let cells = [at]
  for m in String.graphemes(moves) {
    if m == "R" {
      at = Cell{ ...at, col: at.col + 1 }
    } else if m == "L" {
      at = Cell{ ...at, col: at.col - 1 }
    }
    cells.push(at)
  }
  cells.unique()
}

for c in visited("RRLLR") {
  println("${c.row},${c.col}")
}
```

```text
0,0
0,1
0,2
```

## Properties and relations

These hold for every value but `NaN`, which is not `==` to itself, so an array holding `NaN` is not
`==` to its own `unique()`.

`Array.unique` keeps the same order as `Set.of`, which also holds the first of each `==` value:

```ant
let xs = [4, 2, 4, 9, 2, 1]
assert(xs.unique() == Set.values(Set.of(xs)))
```

The result has no repeats, so applying `Array.unique` again changes nothing:

```ant
let xs = ["a", "b", "a", "c", "b"]
let u = xs.unique()
assert(u.unique() == u)
assert(u.every({ u.indexOf(x) == u.lastIndexOf(x) }))
```

Every element of `xs` is in the result, and the result is no longer than `xs`:

```ant
let xs = [5, 3, 5, 5, 8]
let u = xs.unique()
assert(xs.every({ u.contains(x) }))
assert(len(u) <= len(xs))
```

An element is kept exactly when no earlier element is `==` to it:

```ant
let xs = [2, 7, 2, 1, 7]
let kept = xs.entries().filter({ xs.indexOf(x.value) == x.index }).map({ x.value })
assert(kept == xs.unique())
```

`Array.unique` is [`Array.uniqueBy`](uniqueBy.md) with the element as its own key:

```ant
let xs = [3, 1, 3, 2, 1]
assert(xs.unique() == xs.uniqueBy({ x }))
```

The result is a new array, and `xs` keeps its repeats:

```ant
let xs = [1, 1, 2]
let u = xs.unique()
assert(!Array.same(xs, u))
assert(xs == [1, 1, 2])
```

## Possible issues

`NaN` is not `==` to itself, so `Array.unique` keeps every `NaN`. `-0.0 == 0.0`, so it keeps only
the first of the two:

```ant
[1.0, Math.NaN, Math.NaN, 1.0].unique() // [1.0, NaN, NaN]
[0.0, -0.0].unique()                    // [0.0]
[-0.0, 0.0].unique()                    // [-0.0]
```

Functions compare by identity, so two lambdas with the same body both stay:

```ant
let inc: func(int): int = { x + 1 }
let same: func(int): int = { x + 1 }
len([inc, inc, same].unique()) // 2
```

The kept element is the first one, the original record and not a copy. Changing it through the
result changes `xs` as well:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let ps = [p, Point{ x: 1, y: 2 }]
let u = ps.unique()
Record.same(u[0], p) // true
u[0].x = 9
ps[0] // Point{ x: 9, y: 2 }
```

`Array.unique` compares whole values. To drop repeats by one field, or without regard to case, use
[`Array.uniqueBy`](uniqueBy.md) with a key:

```ant
["Ant", "ant", "ANT", "bee"].unique()                        // ["Ant", "ant", "ANT", "bee"]
["Ant", "ant", "ANT", "bee"].uniqueBy({ String.toLower(x) }) // ["Ant", "bee"]
```

`Array.unique` drops repeats wherever they are. To collapse only runs of the same value next to each
other, compare each element with the one before it:

```ant
let xs = [1, 1, 2, 1, 1]
let runs = xs.entries().filter({ x.index == 0 || xs[x.index - 1] != x.value })
xs.unique()           // [1, 2]
runs.map({ x.value }) // [1, 2, 1]
```

## See also

[Array.uniqueBy](uniqueBy.md) · [Set.of](../Set/of.md) · [Set.values](../Set/values.md) ·
[Array.contains](contains.md) · [Array.indexOf](indexOf.md) · [Array.groupBy](groupBy.md) ·
[Array.sort](sort.md) · [== operator](../../language/operators/equality.md)
