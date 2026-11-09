<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.findLast

```ant signature
func Array.findLast[T](xs: Array[T], f: func(T): bool): T | null
```

Array.findLast returns the last element of xs that f accepts, or null when there is none.

<!-- end generated -->

## Details

- Returns `null` when `f` accepts no element, so an empty array gives `null`.
- Calls `f` from the last element back and stops at the first element `f` accepts.
- When `T` includes `null`, a `null` result does not tell a missing element from a found `null`; use
  [Array.findLastIndex](findLastIndex.md).

## Basic examples

Find the last even number:

```ant
Array.findLast([1, 4, 7, 8, 9], { x % 2 == 0 }) // 8
```

When no element matches, the result is `null`:

```ant
Array.findLast([1, 3, 5], { x % 2 == 0 }) // null
```

Call it through `.` on any array:

```ant
var words = ["ant", "bee", "beetle", "wasp"]
words.findLast({ String.startsWith(x, "b") }) // "beetle"
```

## Scope

An empty array has no last match. Annotate it so that `T` is known:

```ant
var empty: Array[string] = []
empty.findLast({ x != "" }) // null
```

When every element matches, the result is the last element:

```ant
[3, 1, 2].findLast({ x > 0 }) // 2
```

`f` can test a record's fields:

```ant
type Order = { id: int, total: float64 }
var orders = [Order{ id: 1, total: 9.5 }, Order{ id: 2, total: 42.0 }, Order{ id: 3, total: 7.25 }]
orders.findLast({ x.total > 8.0 }) // Order{ id: 2, total: 42.0 }
```

A test with `is` finds the last member of one type in a union. The result has type `Shape | null`,
not `Circle | null`: the test inside `f` does not narrow the result:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
var shapes: Array[Shape] = [Circle{ radius: 1.0 }, Square{ side: 2.0 }, Circle{ radius: 3.0 }]
shapes.findLast({ x is Circle }) // Circle{ radius: 3.0 }
shapes.findLast({ x is Square }) // Square{ side: 2.0 }
```

A named function works as well as a lambda:

```ant
func isVowel(s: string): bool {
  String.contains("aeiou", s)
}

String.graphemes("rhythm and blues").findLast(isVowel) // "e"
```

A lambda can read local variables:

```ant
var limit = 10
[4, 12, 9, 15, 3].findLast({ x < limit }) // 3
```

In a pipeline, the array is the first argument:

```ant
"a,bb,ccc,dd" | String.split(",") | Array.findLast({ String.length(x) == 2 }) // "dd"
```

A callback that records its argument shows the order: from the last element back, stopping at the
first match:

```ant
var tested: Array[int] = []
var found = [1, 4, 7, 8, 9].findLast({
  tested.push(x)
  x % 2 == 0
})
println(found)
println(tested)
```

```text
8
[9, 8]
```

With no match, `f` sees every element, last first:

```ant
var tested: Array[int] = []
var found = [1, 3, 5].findLast({
  tested.push(x)
  x % 2 == 0
})
println(found)
println(tested)
```

```text
null
[5, 3, 1]
```

Both parameters can be named, in any order:

```ant
Array.findLast(f: { x > 1 }, xs: [1, 2, 3]) // 3
```

## Applications

Find the most recent error in a log:

```ant
type LogLine = { level: string, message: string }
var log = [
  LogLine{ level: "info", message: "start" },
  LogLine{ level: "error", message: "disk full" },
  LogLine{ level: "info", message: "retry" },
  LogLine{ level: "error", message: "timeout" },
  LogLine{ level: "info", message: "done" },
]
log.findLast({ x.level == "error" })?.message ?? "no errors" // "timeout"
```

Read the last value a setting was given, where later lines override earlier ones:

```ant
var lines = ["port=80", "host=localhost", "port=8080"]
var port = lines.findLast({ String.startsWith(x, "port=") })
if port != null {
  println(String.trimPrefix(port, "port="))
}
```

```text
8080
```

Find the last stop on a route within driving range:

```ant
type Stop = { name: string, km: int }
var route = [Stop{ name: "Ash", km: 40 }, Stop{ name: "Birch", km: 95 }, Stop{ name: "Cedar", km: 160 }]
var range = 120
var stop = route.findLast({ x.km <= range })
println(stop?.name ?? "stay home")
```

```text
Birch
```

## Properties and relations

`Array.findLast` on an array gives what `Array.find` gives on its reverse:

```ant
var xs = [5, 2, 8, 3, 6]
for limit in [0, 4, 7, 10] {
  assert(xs.findLast({ x > limit }) == xs.reverse().find({ x > limit }))
}
```

It agrees with `Array.findLastIndex` and with the last element of `Array.filter`:

```ant
var xs = [5, 2, 8, 3, 6]
for limit in [0, 4, 7, 10] {
  var i = xs.findLastIndex({ x < limit })
  var v = xs.findLast({ x < limit })
  assert(v == (if i != null { xs[i] } else { null }))
  assert(v == xs.filter({ x < limit }).last())
}
```

`f` runs once for each element from the end through the match, so it runs `length − i` times when
the match is at index `i`, and `length` times when there is none:

```ant
var xs = [5, 2, 8, 3, 6]
for limit in [0, 4, 7, 10] {
  var calls = 0
  xs.findLast({
    calls += 1
    x < limit
  })
  var i = xs.findLastIndex({ x < limit })
  assert(calls == xs.length() - (i ?? 0))
}
```

The result is `null` exactly when `Array.some` is `false`, as long as `T` does not include `null`:

```ant
var xs = ["ant", "bee", "wasp"]
for n in [2, 3, 4, 5] {
  assert((xs.findLast({ String.length(x) == n }) != null) == xs.some({ String.length(x) == n }))
}
```

## Possible issues

The result may be `null`, so using it as a `T` is a compile error:

```ant
var last = [1, 2, 3].findLast({ x > 1 })
println(last + 1) // error: + needs both operands of one type, got int | null and int
```

Narrow it first, or give a fallback with `??`:

```ant
var last = [1, 2, 3].findLast({ x > 1 })
(last ?? 0) + 1 // 4
```

When the element type includes `null`, a `null` result does not say whether `f` found a `null` or
found nothing. Use [`Array.findLastIndex`](findLastIndex.md), whose result is an index:

```ant
var readings: Array[int | null] = [3, null, 5]
readings.findLast({ x == null })      // null
readings.findLast({ x == 7 })         // null
readings.findLastIndex({ x == null }) // 1
readings.findLastIndex({ x == 7 })    // null
```

`Array.findLast` takes a predicate, not a value. To look for a value by `==`, use
[`Array.lastIndexOf`](lastIndexOf.md) or [`Array.contains`](contains.md):

```ant
var xs = [1, 2, 3]
println(xs.findLast(2)) // error: cannot use int where func(int): bool is expected
```

## See also

[Array.find](find.md) · [Array.findLastIndex](findLastIndex.md) · [Array.findIndex](findIndex.md) ·
[Array.last](last.md) · [Array.filter](filter.md) · [Array.some](some.md) ·
[Array.lastIndexOf](lastIndexOf.md) · [?? operator](../../language/operators/null-coalescing.md) ·
[?. operator](../../language/operators/optional-chaining.md)
