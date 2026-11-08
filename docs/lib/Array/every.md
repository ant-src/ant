<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.every

```ant signature
func Array.every[T](xs: Array[T], f: func(T): bool): bool
```

Array.every reports whether f accepts every element of xs.

<!-- end generated -->

## Details

- `Array.every` returns `true` when `f` returns `true` for every element of `xs`, and `false` when
  it returns `false` for some element. An empty array has no element for `f` to reject, so `f`
  accepts every element of it, and the result is `true`
  ([spec: Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)).
- `Array.every` calls `f` once per element, from the first, and stops at the first element `f`
  rejects, since that element decides the result. Elements after it are not tested
  ([spec: Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)).
- `f` has the function type `func(T): bool`: one parameter of the element type, and a `bool` result
  ([spec: Function types](https://github.com/ant-src/spec/blob/master/ant-lang.md#function-types)).
- A lambda `{ ... }` passed as `f` takes its parameter type from that function type, so `x` is a `T`
  inside it
  ([spec: Anonymous functions](https://github.com/ant-src/spec/blob/master/ant-lang.md#anonymous-functions)).
  A declared function, a function on a type, or a function value of the same type works as well, and
  a closure sees the current value of the variables it captures
  ([spec: Closures and first-class values](https://github.com/ant-src/spec/blob/master/ant-lang.md#closures-and-first-class-values)).

## Basic examples

```ant
Array.every([2, 4, 6], { x % 2 == 0 }) // true
Array.every([2, 3, 6], { x % 2 == 0 }) // false
```

Call it through `.` on any array:

```ant
var ages = [34, 21, 19]
ages.every({ x >= 18 }) // true
ages.every({ x >= 21 }) // false
```

## Scope

The callback runs from the first element and stops at the first element it rejects, so a side effect
shows which elements it tested:

```ant
var seen: Array[int] = []
[2, 4, 5, 6].every({ seen.push(x); x % 2 == 0 }) // false
seen                                             // [2, 4, 5]
```

When `f` accepts every element, it tests them all:

```ant
var seen: Array[int] = []
[2, 4, 6].every({ seen.push(x); x % 2 == 0 }) // true
seen                                          // [2, 4, 6]
```

An empty array gives `true`, whatever `f` tests. Annotate it so that `T` is known:

```ant
var empty: Array[int] = []
empty.every({ x > 100 }) // true
```

Pass a declared function by name instead of a lambda:

```ant
func isPositive(n: int): bool {
  n > 0
}

[1, 2, 3].every(isPositive)  // true
[1, -2, 3].every(isPositive) // false
```

Library functions are values too. `String.isLetter` accepts letters in any script, and rejects `""`:

```ant
["Ant", "Émile", "Ωμέγα"].every(String.isLetter) // true
["Ant", "R2D2"].every(String.isLetter)           // false
["Ant", ""].every(String.isLetter)               // false
```

Records are read through their fields:

```ant
type Task = { title: string, done: bool }
var tasks = [Task{ title: "write", done: true }, Task{ title: "test", done: false }]
tasks.every({ x.done })                      // false
tasks.every({ String.length(x.title) <= 5 }) // true
```

With a union, test the tag with `is`. Inside `&&`, `x` is narrowed to the member:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect
var shapes: Array[Shape] = [Circle{ radius: 1.0 }, Circle{ radius: 2.5 }]
shapes.every({ x is Circle })                   // true
shapes.every({ x is Circle && x.radius > 2.0 }) // false
```

An element type that includes `null` needs the `null` case handled before `x` is compared:

```ant
var scores: Array[int | null] = [90, null, 75]
scores.every({ x != null })            // false
scores.every({ x == null || x >= 50 }) // true
```

For nested arrays, each nested lambda binds its own `x`:

```ant
var grid = [[1, 2, 3], [4, 5, 6]]
grid.every({ x.length() == 3 })    // true
grid.every({ x.every({ x > 0 }) }) // true
```

`Map.values` and `Set.values` return an array, so test the values of a map or the members of a set
through them:

```ant
var stock = Map{ apples: 4, pears: 0 }
stock.values().every({ x > 0 })                        // false
Set.of(["a", "b"]).values().every({ x.length() == 1 }) // true
```

In a pipeline, the array is the first argument:

```ant
"a b c" | String.split(" ") | Array.every({ String.length(x) == 1 }) // true
```

A closure sees later assignments to the variables it captures:

```ant
var limit = 10
var under: func(int): bool = { x < limit }
[3, 9].every(under) // true
limit = 5
[3, 9].every(under) // false
```

## Applications

Check that an array is sorted, by comparing each element with the one before it:

```ant
func isSorted(xs: Array[int]): bool {
  Array.range(1, xs.length()).every({ xs[x - 1] <= xs[x] })
}

isSorted([1, 3, 3, 8]) // true
isSorted([1, 4, 2])    // false
isSorted([])           // true
```

Validate a form before saving it:

```ant
type Field = { name: string, value: string, required: bool }

func validate(fields: Array[Field]): Result[void, string] {
  if fields.every({ !x.required || x.value != "" }) {
    Result.ok(void)
  } else {
    var missing = fields.filter({ x.required && x.value == "" }).map({ x.name })
    Result.err("missing: ${String.join(missing, ", ")}")
  }
}

println(validate([Field{ name: "email", value: "ann@example.com", required: true }]))
println(validate([
  Field{ name: "email", value: "", required: true },
  Field{ name: "phone", value: "", required: false },
]))
```

```text
Ok{ value: void }
Err{ error: "missing: email" }
```

Check that a string holds only hexadecimal digits:

```ant
var hex = Set.of(String.chars("0123456789abcdef"))
String.chars("c0ffee").every({ hex.has(x) }) // true
String.chars("coffee").every({ hex.has(x) }) // false
```

## Properties and relations

`Array.every` is true exactly when `Array.some` finds no rejected element, when `Array.findIndex`
finds none, and when `Array.countWhere` and `Array.filter` keep the whole array:

```ant
var f: func(int): bool = { x % 3 != 0 }
for xs in [[1, 2, 4], [1, 3, 4], []] {
  assert(xs.every(f) == !xs.some({ !f(x) }))
  assert(xs.every(f) == (xs.findIndex({ !f(x) }) == null))
  assert(xs.every(f) == (xs.countWhere(f) == xs.length()))
  assert(xs.every(f) == (xs.filter(f) == xs))
}
```

Over a concatenation, `every` holds when it holds on both parts:

```ant
var f: func(int): bool = { x > 0 }
var xs = [1, 2]
var ys = [3, -4]
assert(xs.concat(ys).every(f) == (xs.every(f) && ys.every(f)))
assert(xs.every(f) && !ys.every(f))
```

Every member of `a` being in `b` is what `Set.isSubset` reports:

```ant
var a = Set.of([1, 2])
var b = Set.of([1, 2, 3])
assert(a.values().every({ b.has(x) }) == a.isSubset(b))
assert(b.values().every({ a.has(x) }) == b.isSubset(a))
```

## Possible issues

An empty array passes every test, so "all checks passed" on no checks at all is `true`. Test for
emptiness when that matters:

```ant
var results: Array[bool] = []
results.every({ x })                       // true
!results.isEmpty() && results.every({ x }) // false
```

`f` must return a `bool`. Ant has no truthy values, so a lambda that returns an `int` is a compile
error:

<!-- test: error cannot use int where bool is expected -->

```ant
var counts = [1, 2, 3]
println(counts.every({ x }))
```

Write the comparison instead:

```ant
[1, 2, 3].every({ x != 0 }) // true
```

`f` takes one parameter. A function of two parameters does not fit `func(T): bool`:

<!-- test: error cannot use func(int, int): bool where func(int): bool is expected -->

```ant
var counts = [1, 2, 3]
println(counts.every(func(a: int, b: int): bool { a < b }))
```

To compare neighbors, iterate over indices, as `isSorted` above does.

A lambda stored in an unannotated `var` has no expected function type. Annotate the variable:

<!-- test: error a lambda needs an expected function type -->

```ant
var positive = { x > 0 }
println([1, 2].every(positive))
```

A `NaN` element fails `x > 0.0` but passes `!(x <= 0.0)`. Test for it with `Math.isNaN`:

```ant
var xs = [1.0, Math.NaN]
xs.every({ x > 0.0 })                   // false
xs.every({ !(x <= 0.0) })               // true
xs.every({ !Math.isNaN(x) && x > 0.0 }) // false
```

## See also

[Array.some](some.md) · [Array.find](find.md) · [Array.findIndex](findIndex.md) ·
[Array.filter](filter.md) · [Array.countWhere](countWhere.md) · [Array.contains](contains.md) ·
[Set.isSubset](../Set/isSubset.md)
