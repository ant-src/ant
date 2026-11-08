<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.some

```ant signature
func Array.some[T](xs: Array[T], f: func(T): bool): bool
```

Array.some reports whether f accepts at least one element of xs.

<!-- end generated -->

## Details

- `f` has the function type `func(T): bool`: one parameter of the element type, and a `bool` result
  ([spec: Function types](https://github.com/ant-src/spec/blob/master/ant-lang.md#function-types)).
- `f` is any function value of that type: a lambda with its implicit `x`, a declared function, a
  library function, or a closure, which sees the current value of the variables it captures
  ([spec: Closures and first-class values](https://github.com/ant-src/spec/blob/master/ant-lang.md#closures-and-first-class-values)).

## Basic examples

```ant
Array.some([1, 3, 4], { x % 2 == 0 }) // true
Array.some([1, 3, 5], { x % 2 == 0 }) // false
```

Call it through `.` on any array:

```ant
var temps = [18, 24, 31]
temps.some({ x > 30 }) // true
temps.some({ x < 0 })  // false
```

## Scope

Pass a declared function by name instead of a lambda:

```ant
func isNegative(n: int): bool {
  n < 0
}

[3, -1, 4].some(isNegative) // true
[3, 1, 4].some(isNegative)  // false
```

Library functions are values too. `String.isEmpty` finds a blank entry:

```ant
["Ann", "", "Bob"].some(String.isEmpty) // true
["Ann", "Bob"].some(String.isEmpty)     // false
```

Strings compare by their text, Unicode included:

```ant
["mañana", "Ωμέγα", "🐜"].some({ x == "🐜" })             // true
["mañana", "Ωμέγα"].some({ String.startsWith(x, "Ω") }) // true
```

Records are read through their fields:

```ant
type Order = { id: int, paid: bool }
var orders = [Order{ id: 1, paid: true }, Order{ id: 2, paid: false }]
orders.some({ !x.paid })   // true
orders.some({ x.id > 10 }) // false
```

With a union, test the tag with `is`. Inside `&&`, `x` is narrowed to the member:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
var shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
shapes.some({ x is Rect })                     // true
shapes.some({ x is Circle && x.radius > 2.0 }) // false
```

An element type that includes `null` can be searched for `null` itself, or for a value once `null`
is ruled out:

```ant
var scores: Array[int | null] = [90, null, 75]
scores.some({ x == null })           // true
scores.some({ x != null && x > 80 }) // true
scores.some({ x != null && x > 95 }) // false
```

For nested arrays, each nested lambda binds its own `x`:

```ant
var grid = [[0, 0], [0, 7]]
grid.some({ x.some({ x != 0 }) }) // true
grid.some({ x.isEmpty() })        // false
```

The values of a map or the members of a set are arrays through `Map.values` and `Set.values`:

```ant
var stock = Map{ apples: 4, pears: 0 }
stock.values().some({ x == 0 })                 // true
Set.of([2, 4, 6]).values().some({ x % 2 == 1 }) // false
```

In a pipeline, the array is the first argument:

```ant
"red green blue" | String.split(" ") | Array.some({ String.length(x) > 4 }) // true
```

A closure sees later assignments to the variables it captures:

```ant
var target = 3
var hit: func(int): bool = { x == target }
[1, 2, 3].some(hit) // true
target = 9
[1, 2, 3].some(hit) // false
```

## Applications

Check a password for at least one digit:

```ant
func hasDigit(s: string): bool {
  String.chars(s).some({ String.isDigit(x) })
}

hasDigit("hunter2") // true
hasDigit("hunter")  // false
```

Refuse a deploy when any check failed:

```ant
type Check = { name: string, ok: bool }

func deploy(checks: Array[Check]): Result[void, string] {
  if checks.some({ !x.ok }) {
    var failed = checks.filter({ !x.ok }).map({ x.name })
    Result.err("failed: ${String.join(failed, ", ")}")
  } else {
    Result.ok(void)
  }
}

println(deploy([Check{ name: "test", ok: true }, Check{ name: "lint", ok: true }]))
println(deploy([Check{ name: "test", ok: false }, Check{ name: "lint", ok: true }]))
```

```text
Ok{ value: void }
Err{ error: "failed: test" }
```

Find whether an array has a duplicate, by comparing each index with the first index of its value:

```ant
func hasDuplicate(xs: Array[string]): bool {
  Array.range(0, xs.length()).some({ xs.indexOf(xs[x]) != x })
}

hasDuplicate(["a", "b", "a"]) // true
hasDuplicate(["a", "b", "c"]) // false
```

Check whether an array shares an element with a set:

```ant
var mine = Set.of(["go", "ant", "zig"])
["rust", "ant"].some({ mine.has(x) }) // true
["c", "lua"].some({ mine.has(x) })    // false
```

## Properties and relations

`Array.some` is true exactly when `Array.every` fails for the opposite test, when `Array.findIndex`
finds an index, and when `Array.countWhere` and `Array.filter` keep at least one element:

```ant
var f: func(int): bool = { x % 3 == 0 }
for xs in [[1, 2, 4], [1, 3, 4], []] {
  assert(xs.some(f) == !xs.every({ !f(x) }))
  assert(xs.some(f) == (xs.findIndex(f) != null))
  assert(xs.some(f) == (xs.countWhere(f) > 0))
  assert(xs.some(f) == !xs.filter(f).isEmpty())
}
```

Testing for a value with `==` is what `Array.contains` does:

```ant
var xs = ["a", "b", "c"]
assert(xs.some({ x == "b" }) == xs.contains("b"))
assert(xs.some({ x == "z" }) == xs.contains("z"))
```

Over a concatenation, `some` holds when it holds on either part:

```ant
var f: func(int): bool = { x < 0 }
var xs = [1, 2]
var ys = [3, -4]
assert(xs.concat(ys).some(f) == (xs.some(f) || ys.some(f)))
assert(!xs.some(f) && ys.some(f))
```

## Possible issues

An empty array gives `false` for `some`, even for a test that accepts everything, while `every`
gives `true`, even for a test that rejects everything:

```ant
var errors: Array[string] = []
errors.some({ true })   // false
errors.every({ false }) // true
```

`f` must return a `bool`. Ant has no truthy values, so a lambda that returns an `int` is a compile
error:

<!-- test: error cannot use int where bool is expected -->

```ant
var counts = [0, 1, 2]
println(counts.some({ x }))
```

Write the comparison instead:

```ant
[0, 1, 2].some({ x != 0 }) // true
```

`Array.some` takes a predicate, not a value. Passing the value itself is a compile error; use
[`Array.contains`](contains.md) to test by `==`:

<!-- test: error cannot use int where func(int): bool is expected -->

```ant
var xs = [1, 2, 3]
println(xs.some(2))
```

```ant
[1, 2, 3].contains(2) // true
```

A lambda stored in an unannotated `var` has no expected function type. Annotate the variable:

<!-- test: error a lambda needs an expected function type -->

```ant
var big = { x > 100 }
println([1, 200].some(big))
```

`NaN` is not `==` to any value, itself included, because floats compare by IEEE 754 rules
([spec: Expressions](https://github.com/ant-src/spec/blob/master/ant-lang.md#expressions)), so
`x == Math.NaN` never finds it. Test for it with `Math.isNaN`:

```ant
var xs = [1.0, Math.NaN]
xs.some({ x == Math.NaN }) // false
xs.some({ Math.isNaN(x) }) // true
```

## See also

[Array.every](every.md) · [Array.find](find.md) · [Array.findIndex](findIndex.md) ·
[Array.filter](filter.md) · [Array.countWhere](countWhere.md) · [Array.contains](contains.md) ·
[Set.has](../Set/has.md)
