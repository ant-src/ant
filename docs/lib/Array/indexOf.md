<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.indexOf

```ant signature
func Array.indexOf[T](xs: Array[T], v: T): int | null
```

Array.indexOf returns the index of the first element of xs that is == v, or null when there is none.

<!-- end generated -->

## Details

- Compares each element with `v` by [`==`](../../language/operators/equality.md) and returns the
  first matching index.
- Returns `null` when no element matches, so an empty array gives `null`.
- `NaN` equals nothing, itself included, so `NaN` is never found.

## Basic examples

```ant
Array.indexOf([10, 20, 30], 20) // 1
Array.indexOf([10, 20, 30], 40) // null
```

Call it through `.` on any array. A repeated value gives the index of its first occurrence:

```ant
let colors = ["red", "green", "blue", "green"]
colors.indexOf("green") // 1
colors.indexOf("red")   // 0
```

## Scope

An empty array has no index to give. Annotate it so that `T` is known:

```ant
let empty: Array[string] = []
empty.indexOf("") // null
```

Strings compare whole and by code points, with no normalization and no case folding:

```ant
let words = ["Ant", "ant", "café"]
words.indexOf("ant")         // 1
words.indexOf("an")          // null
words.indexOf("cafe\u{301}") // null
```

Records compare by tag and fields, so a fresh record with the same fields is found:

```ant
type Card = { rank: int, suit: string }
let hand = [Card{ rank: 10, suit: "hearts" }, Card{ rank: 1, suit: "spades" }]
hand.indexOf(Card{ rank: 1, suit: "spades" }) // 1
hand.indexOf(Card{ rank: 1, suit: "hearts" }) // null
```

Members of a union compare by tag, then by fields:

```ant
type Pending = {}
type Done = { code: int }
type Failed = { code: int }
type Status = Pending | Done | Failed
let log: Array[Status] = [Pending{}, Failed{ code: 1 }, Done{ code: 1 }]
log.indexOf(Done{ code: 1 }) // 2
log.indexOf(Pending{})       // 0
```

Functions compare by identity, so a lambda with the same body is a different value:

```ant
func inc(x: int): int { x + 1 }
func dec(x: int): int { x - 1 }
let g: func(int): int = { x + 1 }
[inc, dec].indexOf(dec) // 1
[g].indexOf(g)          // 0
[g].indexOf({ x + 1 })  // null
```

Nested arrays and maps compare by contents. Maps are equal in any order:

```ant
[[1, 2], [2, 1]].indexOf([2, 1])                            // 1
[Map{ x: 1 }, Map{ x: 1, y: 2 }].indexOf(Map{ y: 2, x: 1 }) // 1
```

An array whose element type includes `null` can look for `null`:

```ant
let cells: Array[string | null] = ["a", null, "c", null]
cells.indexOf(null) // 1
cells.indexOf("c")  // 2
```

In a pipeline, the array is the first argument:

```ant
"mon tue wed thu fri" | String.split(" ") | Array.indexOf("wed") // 2
```

A direct call can name `v`, through `Array` or through `.`:

```ant
Array.indexOf(["a", "b"], v: "b") // 1
```

## Applications

Turn a weekday name into its number, with a fallback for unknown names:

```ant
let days = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"]
func dayNumber(name: string): int {
  (days.indexOf(name.toLower()) ?? -1) + 1
}

println(dayNumber("Wed"))
println(dayNumber("someday"))
```

```text
3
0
```

Read the value that follows a flag in a list of arguments:

```ant
func flag(args: Array[string], name: string): string | null {
  let i = args.indexOf(name)
  if i == null {
    return null
  }
  args.get(i + 1)
}

let args = ["build", "--out", "dist", "--verbose"]
println(flag(args, "--out"))
println(flag(args, "--verbose"))
println(flag(args, "--quiet"))
```

```text
dist
null
null
```

Remove the first occurrence of a value, and leave the array alone when it is missing:

```ant
let queue = ["ann", "bob", "cy", "bob"]
let i = queue.indexOf("bob")
if i != null {
  queue.removeAt(i)
}
println(queue)
```

```text
["ann", "cy", "bob"]
```

## Properties and relations

The element at the index is `==` to `v`, and no element before it is:

```ant
let xs = [5, 3, 5, 8, 3]
for v in [3, 5, 8] {
  let i = xs.indexOf(v)
  assert(i != null)
  assert(xs[i] == v)
  assert(!xs.slice(0, i).contains(v))
}
```

The result is `null` exactly when [`Array.contains`](contains.md) is `false`, and it agrees with
[`Array.findIndex`](findIndex.md) given an `==` test:

```ant
let xs = ["a", "b", "a"]
for v in ["a", "b", "z"] {
  assert((xs.indexOf(v) == null) == !xs.contains(v))
  assert(xs.indexOf(v) == xs.findIndex({ x == v }))
}
```

With [`Array.lastIndexOf`](lastIndexOf.md), it tells whether a value occurs exactly once:

```ant
let xs = [4, 7, 4, 9]
assert(xs.indexOf(7) == xs.lastIndexOf(7))
assert(xs.indexOf(4) != xs.lastIndexOf(4))
```

## Possible issues

The result is nullable, so it cannot index an array as it is:

```ant
let xs = ["a", "b"]
println(xs[xs.indexOf("b")]) // error: cannot use int | null where int is expected
```

Test it against `null`, not for truth. `0` is a valid index, and an `int | null` is not a condition:

```ant
let i = ["a", "b"].indexOf("a")
if i { // error: cannot use int | null where bool is expected
  println(i)
}
```

`NaN` is not `==` to itself, so `Array.indexOf` never finds it. Use
[`Array.findIndex`](findIndex.md) with `Math.isNaN`:

```ant
let xs = [1.0, Math.NaN]
xs.indexOf(Math.NaN)            // null
xs.findIndex({ Math.isNaN(x) }) // 1
```

`-0.0 == 0.0`, so looking for `-0.0` finds an earlier `0.0`:

```ant
[0.0, -0.0].indexOf(-0.0) // 0
```

Numeric types never mix, so looking for a `float64` in an `Array[int]` is a compile error:

```ant
let xs = [1, 2, 3]
println(xs.indexOf(2.0)) // error: cannot use float64 where int is expected
```

`Array.indexOf` takes a value, not a predicate. To find an element with a function, use
[`Array.findIndex`](findIndex.md):

```ant
let xs = [1, 2, 3]
println(xs.indexOf({ x > 1 }))
// error: a lambda needs an expected function type, but int is expected
```

On a string array, `Array.indexOf` finds a whole element. To find text inside one string, use
[`String.indexOf`](../String/indexOf.md):

```ant
["hello world"].indexOf("world")       // null
String.indexOf("hello world", "world") // 6
```

## See also

[Array.lastIndexOf](lastIndexOf.md) · [Array.contains](contains.md) ·
[Array.findIndex](findIndex.md) · [Array.find](find.md) · [Array.removeAt](removeAt.md) ·
[String.indexOf](../String/indexOf.md) · [Bytes.indexOf](../Bytes/indexOf.md) ·
[== operator](../../language/operators/equality.md)
