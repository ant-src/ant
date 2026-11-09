<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.sum

```ant signature
func Array.sum[T: Number](xs: Array[T]): T | null
```

Array.sum returns the sum of the elements of xs, or null when xs is empty. Write Array.sum(xs) ?? 0
for a zero default. It panics on overflow.

<!-- end generated -->

## Details

- Returns `null` when `xs` is empty, not a zero, so give it a default with `??`.
- Integer overflow panics, as it does for `+`. A `NaN`, or infinities of both signs, give `NaN`.

## Basic examples

```ant
Array.sum([1, 2, 3, 4]) // 10
Array.sum([2.5, 0.25])  // 2.75
```

An empty array gives `null`. Give a default with `??`:

```ant
let empty: Array[int] = []
Array.sum(empty)      // null
Array.sum(empty) ?? 0 // 0
```

Call it through `.` on the array:

```ant
let prices = [4.5, 12.0, 3.25]
prices.sum() // 19.75
```

## Scope

A single element is its own sum:

```ant
[42].sum() // 42
```

Every numeric type works, and the sum keeps the element type:

```ant
[byte(200), byte(50)].sum()          // 250
[int8(-100), int8(-28)].sum()        // -128
[float32(0.5), float32(0.25)].sum()  // 0.75
[int64(1), int64(2), int64(3)].sum() // 6
```

Negative numbers cancel:

```ant
[5, -3, -2].sum() // 0
```

Infinities and `NaN` follow IEEE 754:

```ant
[1.0, Math.Infinity].sum()            // Infinity
[Math.Infinity, -Math.Infinity].sum() // NaN
[1.0, Math.NaN, 2.0].sum()            // NaN
```

The values of a map are an array, so `Array.sum` takes them:

```ant
let stock = Map{ apples: 3, pears: 5, plums: 0 }
stock.values().sum() // 8
```

In a pipe, the array is the first argument:

```ant
[1, 2, 3] | Array.map({ x * x }) | Array.sum() // 14
```

## Applications

Total a field of a list of records by mapping to the field first:

```ant
type Item = { name: string, price: float64, qty: int }
let cart = [
  Item{ name: "pen", price: 1.5, qty: 4 },
  Item{ name: "pad", price: 3.0, qty: 2 },
]
cart.map({ x.price * float64(x.qty) }).sum() ?? 0.0 // 12.0
```

Average a list of scores, with `null` when there are none:

```ant
func average(xs: Array[float64]): float64 | null {
  let total = xs.sum()
  if total == null {
    return null
  }
  total / float64(len(xs))
}

println(average([90.0, 75.0, 84.0]))
println(average([]))
```

```text
83.0
null
```

Count the characters of a list of words:

```ant
let words = ["ant", "colony", "queen"]
words.map({ String.length(x) }).sum() // 14
```

Add up the numbers in a line of text, skipping what does not parse:

```ant
let line = "12 7 x 30"
let ns: Array[int] = []
for part in String.split(line, " ") {
  let n = String.toInt(part)
  if n != null {
    ns.push(n)
  }
}
ns.sum() ?? 0 // 49
```

## Properties and relations

On a non-empty array of integers, `Array.sum` agrees with a fold through `+` that starts at zero:

```ant
let xs = [3, -7, 4, 1]
assert(xs.sum() == xs.reduce(0, func(a: int, b: int): int { a + b }))
```

The integers from 1 to `n` add up to `n * (n + 1) / 2`:

```ant
for n in [1, 10, 100] {
  assert(Array.range(1, n + 1).sum() == n * (n + 1) / 2)
}
```

`n` copies of one integer add up to `n` times it:

```ant
assert(Array.filled(5, 7).sum() == 35)
```

When no partial sum overflows, integer sums do not depend on the order of the elements, and
splitting an array splits its sum:

```ant
let xs = [5, 2, 9, 2]
let ys = [1, 8]
assert(xs.sum() == xs.reverse().sum())
assert(xs.concat(ys).sum() == (xs.sum() ?? 0) + (ys.sum() ?? 0))
```

## Possible issues

The result is `T | null`, not `T`, so using it as a `T` is a compile error:

```ant
let xs = [1, 2, 3]
println(xs.sum() + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??`:

```ant
let xs = [1, 2, 3]
(xs.sum() ?? 0) + 1 // 7
```

The default must have the element type. Numeric types never mix, so `?? 0.0` on an `Array[int]` does
not give a `float64`:

```ant
let xs = [1, 2, 3]
let total: float64 = xs.sum() ?? 0.0 // error: cannot use int where float64 is expected
```

Integer overflow panics. The sum of small types overflows sooner:

```ant
[byte(200), byte(100)].sum()   // panic: Array.sum: integer overflow
[9223372036854775807, 1].sum() // panic: Array.sum: integer overflow
```

Convert the elements to a wider type first, or wrap on purpose with
[`Bits.wrapAdd`](../Bits/wrapAdd.md):

```ant
let xs = [byte(200), byte(100)]
xs.map({ int(x) }).sum()         // 300
xs.reduce(byte(0), Bits.wrapAdd) // 44
```

An array whose element type is a union, such as `int | null` or `Number`, has no single `T`, so it
is a compile error:

```ant
let xs: Array[int | null] = [1, null, 3]
println(xs.sum())
// error: T would be int | null, which is not a single member of its bound Number; narrow first
```

Collect the non-`null` elements into an `Array[int]` first. Testing each one narrows it:

```ant
let xs: Array[int | null] = [1, null, 3]
let ns: Array[int] = []
for v in xs {
  if v != null {
    ns.push(v)
  }
}
ns.sum() // 4
```

Because `T` never includes `null`, a `null` result always means an empty array.

Named primitive types, strings, and records are not `Number`:

```ant
type Cents = int
let fees = [Cents(150), Cents(275)]
println(fees.sum())
// error: T would be Cents, which is not a single member of its bound Number; narrow first
```

Unwrap a named primitive with its base type, and wrap the total again:

```ant
type Cents = int
let fees = [Cents(150), Cents(275)]
Cents(fees.map({ int(x) }).sum() ?? 0) // Cents(425)
```

To join strings, use [`String.join`](../String/join.md):

```ant
String.join(["a", "b", "c"], "") // "abc"
```

One `NaN` makes the result `NaN`. Drop it first with `Math.isNaN`:

```ant
let xs = [1.0, Math.NaN, 3.0]
xs.filter({ !Math.isNaN(x) }).sum() // 4.0
```

## See also

[Array.reduce](reduce.md) · [Array.max](max.md) · [Array.min](min.md) · [Array.map](map.md) ·
[Array.length](length.md) · [Bits.wrapAdd](../Bits/wrapAdd.md) · [Math.isNaN](../Math/isNaN.md) ·
[String.join](../String/join.md) · [arithmetic operators](../../language/operators/arithmetic.md) ·
[?? operator](../../language/operators/null-coalescing.md)
