<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.findLastIndex

```ant signature
func Array.findLastIndex[T](xs: Array[T], f: func(T): bool): int | null
```

Array.findLastIndex returns the index of the last element of xs that f accepts, or null when there
is none.

<!-- end generated -->

## Details

- Returns `null` when `f` accepts no element, so an empty array gives `null`.
- Calls `f` from the last element back and stops at the first element `f` accepts.

## Basic examples

Find the position of the last number above 5:

```ant
Array.findLastIndex([3, 8, 1, 9, 4], { x > 5 }) // 3
```

When `f` accepts no element, the result is `null`:

```ant
Array.findLastIndex([3, 8, 1], { x > 50 }) // null
```

Call it through `.` on any array:

```ant
var words = ["ant", "bee", "wasp", "fly"]
words.findLastIndex({ String.length(x) == 3 })     // 3
words.findLastIndex({ String.startsWith(x, "b") }) // 1
```

## Scope

An empty array has no last match. Annotate it so that `T` is known:

```ant
var empty: Array[int] = []
empty.findLastIndex({ x > 0 }) // null
```

A one-element array gives `0` or `null`:

```ant
[7].findLastIndex({ x == 7 }) // 0
[7].findLastIndex({ x == 8 }) // null
```

When `f` accepts every element, the result is the last index:

```ant
[1, 2, 3].findLastIndex({ x > 0 }) // 2
```

Strings are tested whole. Split them into graphemes to search by character:

```ant
var letters = String.graphemes("naïve café")
letters.findLastIndex({ x == "é" || x == "ï" }) // 9
```

An array whose element type includes `null` can look for the last `null`:

```ant
var readings: Array[int | null] = [12, null, 15, null, 9]
readings.findLastIndex({ x == null })           // 3
readings.findLastIndex({ x != null && x > 10 }) // 2
```

Records are tested through their fields:

```ant
type Task = { name: string, done: bool }
var tasks = [
  Task{ name: "plan", done: true },
  Task{ name: "build", done: true },
  Task{ name: "ship", done: false },
]
tasks.findLastIndex({ x.done }) // 1
```

A function on a type works the same way:

```ant
type Task = { name: string, done: bool }

func Task.isOpen(t: Task): bool {
  !t.done
}

var tasks = [
  Task{ name: "plan", done: false },
  Task{ name: "build", done: true },
  Task{ name: "ship", done: false },
]
tasks.findLastIndex(Task.isOpen) // 2
```

Members of a union are tested with `is`:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
var shapes: Array[Shape] = [Circle{ radius: 1.0 }, Square{ side: 2.0 }, Circle{ radius: 3.0 }]
shapes.findLastIndex({ x is Square })                   // 1
shapes.findLastIndex({ x is Circle && x.radius < 2.0 }) // 0
```

A named function works where a lambda does:

```ant
func isEven(n: int): bool {
  n % 2 == 0
}

[1, 2, 3, 4, 5].findLastIndex(isEven) // 3
```

A call can name `f`:

```ant
Array.findLastIndex([4, 9, 2], f: { x > 3 }) // 1
```

In a pipeline, the array is the first argument:

```ant
"a,b,,c,," | String.split(",") | Array.findLastIndex({ x != "" }) // 3
```

## Applications

Find the last error in a log and read the lines after it:

```ant
var log = ["start", "error: disk", "retry", "error: net", "retry", "ok"]
var i = log.findLastIndex({ String.startsWith(x, "error") })
if i != null {
  println(log[i])
  println(Array.slice(log, i + 1, len(log)))
}
```

```text
error: net
["retry", "ok"]
```

Trim trailing blank lines from a text:

```ant
func trimTrailing(lines: Array[string]): Array[string] {
  var last = lines.findLastIndex({ String.trim(x) != "" }) ?? -1
  Array.slice(lines, 0, last + 1)
}

println(trimTrailing(["one", "two", "", "  "]))
println(trimTrailing(["", ""]))
```

```text
["one", "two"]
[]
```

Find the last day a sensor reported a value above its limit:

```ant
type Reading = { day: string, value: float64 }

var readings = [
  Reading{ day: "Mon", value: 18.5 },
  Reading{ day: "Tue", value: 31.0 },
  Reading{ day: "Wed", value: 22.0 },
  Reading{ day: "Thu", value: 30.5 },
  Reading{ day: "Fri", value: 19.0 },
]
var i = readings.findLastIndex({ x.value > 30.0 })
println(if i != null { "last over limit: ${readings[i].day}" } else { "never over limit" })
```

```text
last over limit: Thu
```

## Properties and relations

`f` sees the elements from the last back, and the search stops at the first one it accepts:

```ant
var xs = [3, 8, 1, 9, 4]
var i = xs.findLastIndex({
  println("testing ${x}")
  x > 5
})
println(i)
```

```text
testing 4
testing 9
3
```

When `f` accepts no element, it has run once on each, last first:

```ant
var seen: Array[int] = []
var i = [1, 2, 3].findLastIndex({
  seen.push(x)
  x > 9
})
assert(i == null)
assert(seen == [3, 2, 1])
```

The element at the index is the one [`Array.findLast`](findLast.md) returns:

```ant
var xs = [4, 11, 7, 15, 2]
func f(n: int): bool {
  n > 10
}

var i = xs.findLastIndex(f)
assert(i != null)
assert(xs[i] == xs.findLast(f))
assert(xs.findLastIndex({ x > 100 }) == null && xs.findLast({ x > 100 }) == null)
```

No element after the index is accepted, and the result is `null` exactly when
[`Array.some`](some.md) is `false`:

```ant
var xs = [4, 11, 7, 15, 2]
for limit in [0, 5, 10, 20] {
  var i = xs.findLastIndex({ x > limit })
  assert((i != null) == xs.some({ x > limit }))
  var from = (i ?? -1) + 1
  assert(!Array.slice(xs, from, len(xs)).some({ x > limit }))
}
```

It mirrors [`Array.findIndex`](findIndex.md) on the reversed array:

```ant
var xs = [3, 6, 1, 8, 5]
var j = xs.reverse().findIndex({ x % 2 == 0 })
assert(j != null)
assert(xs.findLastIndex({ x % 2 == 0 }) == len(xs) - 1 - j)
```

With an `==` test it agrees with [`Array.lastIndexOf`](lastIndexOf.md):

```ant
var xs = ["a", "b", "a", "c"]
for v in ["a", "c", "z"] {
  assert(xs.findLastIndex({ x == v }) == xs.lastIndexOf(v))
}
```

## Possible issues

The result is nullable, so it cannot index the array as it is:

<!-- test: error cannot use int | null where int is expected -->

```ant
var xs = [5, 6, 7]
println(xs[xs.findLastIndex({ x > 5 })])
```

Nor can it go straight into arithmetic:

<!-- test: error + needs both operands of one type, got int | null and int -->

```ant
var xs = [5, 6, 7]
println(xs.findLastIndex({ x > 5 }) + 1)
```

Test it against `null`, not for truth. `0` is a valid index, and an `int | null` is not a condition:

<!-- test: error cannot use int | null where bool is expected -->

```ant
var i = [5, 6].findLastIndex({ x == 5 })
if i {
  println(i)
}
```

```ant
var i = [5, 6].findLastIndex({ x == 5 })
i != null // true
i         // 0
```

`f` must return a `bool`. A value that is not one is a compile error, not a truth test:

<!-- test: error cannot use int where bool is expected -->

```ant
println([0, 1, 2].findLastIndex({ x }))
```

`NaN` is not `==` to itself, so an `==` test never finds it. Test with `Math.isNaN`:

```ant
var xs = [Math.NaN, 1.0, Math.NaN, 2.0]
xs.findLastIndex({ x == Math.NaN }) // null
xs.findLastIndex({ Math.isNaN(x) }) // 2
```

Because the search stops early, a callback is no place to count. Here `f` never sees the first `7`,
so it counts one; [`Array.countWhere`](countWhere.md) counts both:

```ant
var sevens = 0
var xs = [2, 7, 4, 7, 1]
var i = xs.findLastIndex({
  if x == 7 {
    sevens += 1
  }
  x == 7
})
i                         // 3
sevens                    // 1
xs.countWhere({ x == 7 }) // 2
```

A callback that pushes to the array does not make `Array.findLastIndex` test the new elements:

```ant
var xs = [1, 2, 3]
xs.findLastIndex({ xs.push(x + 10); x > 5 }) // null
```

## See also

[Array.findIndex](findIndex.md) · [Array.findLast](findLast.md) · [Array.find](find.md) ·
[Array.lastIndexOf](lastIndexOf.md) · [Array.indexOf](indexOf.md) · [Array.some](some.md) ·
[Array.entries](entries.md) · [Array.slice](slice.md)
