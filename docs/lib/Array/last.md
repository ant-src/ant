<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.last

```ant signature
func Array.last[T](xs: Array[T]): T | null
```

Array.last returns the last element of xs, or null when xs is empty.

<!-- end generated -->

## Details

- Returns `null` when `xs` is empty; `xs` is unchanged. To remove the last element, use
  [Array.pop](pop.md).
- When `T` includes `null`, a `null` result does not tell an empty array from a `null` last element;
  use [Array.isEmpty](isEmpty.md).

## Basic examples

```ant
Array.last([10, 20, 30]) // 30
Array.last(["a", "b"])   // "b"
```

An empty array gives `null`. Annotate it so that `T` is known:

```ant
var empty: Array[int] = []
Array.last(empty) // null
```

Call it through `.` on any array:

```ant
var history = ["open", "edit", "save"]
history.last() // "save"
```

## Scope

An array of one element gives that element:

```ant
[42].last() // 42
```

An array of records gives the last record. Read a field of it with `?.`, which yields `null` for an
empty array:

```ant
type Commit = { hash: string, message: string }
var commits = [Commit{ hash: "a1f", message: "init" }, Commit{ hash: "9c2", message: "fix typo" }]
var none: Array[Commit] = []
commits.last()                    // Commit{ hash: "9c2", message: "fix typo" }
commits.last()?.hash              // "9c2"
none.last()?.hash                 // null
none.last()?.hash ?? "no commits" // "no commits"
```

A nested array gives its last inner array, and `?.[i]` indexes into it:

```ant
var rows = [[1, 2], [3, 4, 5]]
rows.last()      // [3, 4, 5]
rows.last()?.[0] // 3
```

Elements of a union keep their tag, so `is` tells which member came last:

```ant
type Passed = { value: int }
type Failed = { reason: string }
type Step = Passed | Failed
var steps: Array[Step] = [Passed{ value: 1 }, Failed{ reason: "timeout" }]
steps.last()           // Failed{ reason: "timeout" }
steps.last() is Failed // true
```

An array of functions gives a function, which `?.(…)` calls when it is there:

```ant
var handlers: Array[func(string): string] = [{ x.toUpper() }, { "<${x}>" }]
var empty: Array[func(string): string] = []
handlers.last()?.("hi") // "<hi>"
empty.last()?.("hi")    // null
```

For the last character of a string, split it into graphemes first, so that a character built from
several code points stays whole:

```ant
String.graphemes("bonjour 👋🏽").last() // "👋🏽"
```

The keys of a map and the values of a set are arrays in insertion order, so `Array.last` gives the
key or member whose first insertion came last. Replacing a map value or re-adding a set member does
not move it:

```ant
Map.keys(Map{ b: 2, a: 1 }).last() // "a"
Set.of([3, 1, 2]).values().last()  // 2
var m = Map{ b: 2, a: 1 }
m["b"] = 5
Map.keys(m).last() // "a"
var s = Set.of([3, 1, 2])
s.add(3)          // false
s.values().last() // 2
```

In a pipe, the array is the first argument:

```ant
"src/lib/geo.ant" | String.split("/") | Array.last() // "geo.ant"
```

## Applications

Take the file name from a path, or a default for an empty one:

```ant
func baseName(path: string): string {
  path.split("/").last() ?? ""
}

println(baseName("docs/lib/Array/last.md"))
println(baseName("README.md"))
```

```text
last.md
README.md
```

Read the top of a stack and leave the stack as it was:

```ant
var stack = ["main", "parse", "expr"]
var top = stack.last()
if top != null {
  println("in ${top}, depth ${stack.length()}")
}
```

```text
in expr, depth 3
```

Keep the latest reading from a series, and report when there is none:

```ant
type Reading = { at: string, celsius: float64 }

func latest(rs: Array[Reading]): string {
  var r = rs.last() ?? return "no data"
  "${r.celsius} at ${r.at}"
}

println(latest([Reading{ at: "09:00", celsius: 14.5 }, Reading{ at: "12:00", celsius: 19.0 }]))
println(latest([]))
```

```text
19.0 at 12:00
no data
```

Stop with a panic when an array that must not be empty is:

```ant
var results: Array[int] = []
results.last() ?? panic("no results") // panic: no results
```

## Properties and relations

`Array.last` agrees with `Array.get` at the last index, with indexing on a non-empty array, with
`Array.findLast` given a function that accepts everything, and with `Array.first` of the reversed
array:

```ant
for xs in [[4, 5, 6], [7], []] {
  assert(xs.last() == xs.get(xs.length() - 1))
  assert(xs.last() == xs.findLast({ true }))
  assert(xs.last() == xs.reverse().first())
  if !xs.isEmpty() {
    assert(xs.last() == xs[xs.length() - 1])
  }
}
```

After `Array.push`, the pushed value is the last element:

```ant
var xs = [1, 2]
xs.push(3)
assert(xs.last() == 3)
```

When the element type does not include `null`, the result is `null` exactly when the array is empty:

```ant
for xs in [["a"], []] {
  assert((xs.last() == null) == xs.isEmpty())
}
```

## Possible issues

The result is `T | null`, not `T`, so using it as a `T` is a compile error:

```ant
var xs = [3, 1, 2]
println(xs.last() + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??`, or narrow a local:

```ant
var xs = [3, 1, 2]
(xs.last() ?? 0) + 1 // 3
```

Testing the call itself does not narrow a second call, because only locals, parameters, and
top-level `let`s narrow:

```ant
var xs = [3, 1, 2]
if xs.last() != null {
  println(xs.last() * 2) // error: * needs both operands of one type, got int | null and int
}
```

Bind the result to a local and test that:

```ant
var xs = [3, 1, 2]
var tail = xs.last()
if tail != null {
  println(tail * 2)
}
```

```text
4
```

When the element type includes `null`, a `null` result does not tell an empty array from one whose
last element is `null`. Ask the array with [`Array.isEmpty`](isEmpty.md) instead:

```ant
var answers: Array[string | null] = ["yes", null]
answers.last()    // null
answers.isEmpty() // false
```

Indexing with `[-1]` does not count from the end. It panics, as any index outside the array does:

```ant
var xs = [1, 2, 3]
xs.last() // 3
xs[-1]    // panic: index -1 out of range for an array of length 3
```

## See also

[Array.first](first.md) · [Array.pop](pop.md) · [Array.get](get.md) · [Array.findLast](findLast.md)
· [Array.push](push.md) · [Array.isEmpty](isEmpty.md) · [Array.slice](slice.md) ·
[?? operator](../../language/operators/null-coalescing.md) ·
[?. operator](../../language/operators/optional-chaining.md) · [Null guide](../../guide/null.md)
