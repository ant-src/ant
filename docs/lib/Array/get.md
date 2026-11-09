<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.get

```ant signature
func Array.get[T](xs: Array[T], i: int): T | null
```

Array.get returns element i of xs, or null when i is out of range. Unlike xs\[i\], it never panics.

<!-- end generated -->

## Details

- Returns `null` when `i` is out of range, a negative `i` included. `xs[i]` panics instead.
- When `T` includes `null`, a `null` result does not tell a missing element from a stored `null`;
  compare `i` with `len(xs)`.

## Basic examples

```ant
Array.get([10, 20, 30], 1) // 20
Array.get([10, 20, 30], 3) // null
```

Call it through `.` on any array:

```ant
var fruit = ["apple", "pear", "plum"]
fruit.get(0) // "apple"
fruit.get(9) // null
```

## Scope

The last valid index is the length minus one:

```ant
var xs = [10, 20, 30]
xs.get(xs.length() - 1) // 30
xs.get(xs.length())     // null
```

A negative index is out of range. Ant does not count from the end:

```ant
[10, 20, 30].get(-1) // null
```

An empty array has no index in range. Annotate it so that `T` is known:

```ant
var empty: Array[string] = []
empty.get(0) // null
```

Records come back as they are stored, and `?.` reads a field of the result:

```ant
type User = { id: int, name: string }
var users = [User{ id: 1, name: "Ann" }, User{ id: 7, name: "Bob" }]
users.get(1)       // User{ id: 7, name: "Bob" }
users.get(1)?.name // "Bob"
users.get(2)?.name // null
```

Chain it with `?.` to reach into a nested array:

```ant
var grid = [[1, 2], [3, 4, 5]]
grid.get(1)?.get(2) // 5
grid.get(0)?.get(2) // null
grid.get(5)?.get(0) // null
```

Index the graphemes of a string by splitting it with `String.graphemes` first:

```ant
String.graphemes("🐜🐝🐞").get(1) // "🐝"
```

In a pipeline, the array is the first argument:

```ant
"a,b,c" | String.split(",") | Array.get(2) // "c"
```

## Applications

Fall back to a default with `??`:

```ant
var names = ["Ann", "Bob"]
names.get(0) ?? "nobody" // "Ann"
names.get(4) ?? "nobody" // "nobody"
```

Read optional command-line arguments by position:

<!-- test: args report.csv -->

```ant
var args = Os.args()
var input = args.get(0) ?? "data.csv"
var output = args.get(1) ?? "out.txt"
println("${input} -> ${output}")
```

```text
report.csv -> out.txt
```

Look at the neighbors of an element without testing the bounds by hand:

```ant
var temps = [12, 15, 11, 18]
for i in Array.range(0, temps.length()) {
  var left = temps.get(i - 1) ?? temps[i]
  var right = temps.get(i + 1) ?? temps[i]
  println("${i}: ${left} ${temps[i]} ${right}")
}
```

```text
0: 12 12 15
1: 12 15 11
2: 15 11 18
3: 11 18 18
```

Return early from a function when the index is out of range, with `?? return`:

```ant
func weekday(n: int): string {
  var name = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].get(n - 1) ?? return "day ${n}?"
  "day ${n} is ${name}"
}

println(weekday(3))
println(weekday(8))
```

```text
day 3 is Wed
day 8?
```

## Properties and relations

Inside the range, `Array.get` agrees with `xs[i]`; outside it, the result is `null`:

```ant
var xs = [4, 9, 16]
for i in Array.range(-2, 5) {
  if i >= 0 && i < xs.length() {
    assert(xs.get(i) == xs[i])
  } else {
    assert(xs.get(i) == null)
  }
}
```

Index `0` is `Array.first`, and the length minus one is `Array.last`, for an empty array too:

```ant
var xs = ["x", "y", "z"]
var empty: Array[string] = []
assert(xs.get(0) == xs.first())
assert(xs.get(xs.length() - 1) == xs.last())
assert(empty.get(0) == empty.first())
assert(empty.get(-1) == empty.last())
```

A record it returns is the one in the array, not a copy, so a change to it shows in the array:

```ant
type Item = { name: string, stock: int }
var items = [Item{ name: "nut", stock: 0 }]
var item = items.get(0)
if item != null {
  item.stock = 10
}
assert(items[0].stock == 10)
```

## Possible issues

The result may be `null`, so using it as a `T` without narrowing is a compile error:

<!-- test: error cannot use int | null where int is expected -->

```ant
var xs = [10, 20, 30]
var v: int = xs.get(1)
println(v)
```

Use `?.`, `??`, or a test against `null` first, as in the examples above.

Testing the call does not narrow a second call, which is still `User | null`:

<!-- test: error has no field name -->

```ant
type User = { id: int, name: string }
var users = [User{ id: 1, name: "Ann" }]
if users.get(0) != null {
  println(users.get(0).name)
}
```

Store the result in a `var` first, then test the variable, as the record example in Properties and
relations does.

`xs[i]` is not the same: it panics where `Array.get` returns `null`:

```ant
[10, 20, 30][3] // panic: index 3 out of range for an array of length 3
```

When the element type includes `null`, a stored `null` and an index out of range look the same.
Compare the index with the length to tell them apart:

```ant
var scores: Array[int | null] = [90, null]
scores.get(1)       // null
scores.get(5)       // null
1 < scores.length() // true
5 < scores.length() // false
```

The index must be an `int`. Numeric types never mix, so a `float64` is a compile error:

<!-- test: error cannot use float64 where int is expected -->

```ant
var xs = [10, 20, 30]
println(xs.get(1.0))
```

`?.[i]` is not a safe index: it guards against a `null` array, not an index out of range, and it
needs a left side whose type includes `null`. Use `Array.get` for the index:

<!-- test: error needs a nullable left side -->

```ant
var xs = [10, 20, 30]
println(xs?.[5])
```

## See also

[Array.first](first.md) · [Array.last](last.md) · [Array.length](length.md) ·
[Array.indexOf](indexOf.md) · [Array.find](find.md) · [Map.get](../Map/get.md) ·
[?. operator](../../language/operators/optional-chaining.md) ·
[?? operator](../../language/operators/null-coalescing.md)
