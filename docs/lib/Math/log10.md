<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.log10

```ant signature
func Math.log10(x: float64): float64
```

Math.log10 returns the base-10 logarithm of x.

<!-- end generated -->

## Details

- `Math.log10` never panics. `0.0` gives `-Infinity`, a negative `x` gives `NaN`, `Infinity` gives
  `Infinity`, and `NaN` gives `NaN`.

## Basic examples

```ant
Math.log10(100.0)  // 2.0
Math.log10(1000.0) // 3.0
Math.log10(2.0)    // 0.3010299956639812
```

Numbers between `0.0` and `1.0` have a negative logarithm:

```ant
Math.log10(0.01) // -2.0
Math.log10(0.5)  // -0.3010299956639812
```

## Scope

`1.0` gives `0.0`, and both zeros give `-Infinity`:

```ant
Math.log10(1.0)  // 0.0
Math.log10(0.0)  // -Infinity
Math.log10(-0.0) // -Infinity
```

A negative number has no real logarithm, and the infinities and `NaN` follow IEEE 754:

```ant
Math.log10(-1.0)           // NaN
Math.log10(Math.Infinity)  // Infinity
Math.log10(-Math.Infinity) // NaN
Math.log10(Math.NaN)       // NaN
```

The result stays small across the whole `float64` range, subnormal numbers included:

```ant
Math.log10(Math.maxFloat64) // 308.25471555991675
Math.log10(1e-300)          // -300.0
Math.log10(5e-324)          // -323.30621534311575
```

An integer goes through `float64` first:

```ant
let n = 1000000
Math.log10(float64(n)) // 6.0
```

Pass it by name to `Array.map`, call it in a pipe, or name its parameter `x`:

```ant
[1.0, 10.0, 100.0].map(Math.log10) // [0.0, 1.0, 2.0]
1e6 | Math.log10()                 // 6.0
Math.log10(x: 1e6)                 // 6.0
```

## Applications

Express a power ratio in decibels:

```ant
func decibels(ratio: float64): float64 {
  10.0 * Math.log10(ratio)
}

decibels(100.0) // 20.0
decibels(0.5)   // -3.010299956639812
```

Find the pH of a solution from its hydrogen-ion concentration:

```ant
let concentration = 1e-7
-Math.log10(concentration) // 7.0
```

Find the order of magnitude of a number, such as the step for an axis tick:

```ant
func magnitude(x: float64): float64 {
  Math.pow(10.0, Math.floor(Math.log10(x)))
}

magnitude(4321.0) // 1000.0
magnitude(0.05)   // 0.01
```

Compare earthquake magnitudes: each step on the scale is ten times the amplitude:

```ant
type Quake = { place: string, amplitude: float64 }

let quakes = [Quake{ place: "A", amplitude: 1000.0 }, Quake{ place: "B", amplitude: 100000.0 }]
for q in quakes {
  println("${q.place}: ${Math.log10(q.amplitude)}")
}
```

```text
A: 3.0
B: 5.0
```

## Properties and relations

The logarithm of a product is the sum of the logarithms, and `Math.pow(10.0, x)` undoes it, each up
to rounding. It is also `Math.log(x) / Math.log(10.0)`:

```ant
func close(a: float64, b: float64): bool {
  Math.abs(a - b) <= 1e-12 * Math.max(1.0, Math.max(Math.abs(a), Math.abs(b)))
}

for x in [0.25, 2.0, 7.0, 42.0, 1e10] {
  assert(close(Math.log10(x * 3.0), Math.log10(x) + Math.log10(3.0)))
  assert(close(Math.pow(10.0, Math.log10(x)), x))
  assert(close(Math.log10(x), Math.log(x) / Math.log(10.0)))
}
```

`Math.log10` is increasing, so it keeps the order of positive numbers:

```ant
let xs = [0.001, 0.5, 1.0, 3.0, 1e9]
let ys = xs.map(Math.log10)
for i in Iter.range(1, ys.length()) {
  assert(ys[i - 1] < ys[i])
}
```

## Possible issues

The result is a `float64` and can land one digit below a whole number:

```ant
Math.log10(1e15) // 14.999999999999998
Math.log10(0.1)  // -0.9999999999999999
```

So counting digits with `Math.floor(Math.log10(x)) + 1.0` can be off by one. For a non-negative
`int`, count with `String.length("${n}")` instead:

```ant
let n = 1000000000000000
Math.floor(Math.log10(float64(n))) + 1.0 // 15.0
String.length("${n}")                    // 16
```

`Math.log(x) / Math.log(10.0)` is a different rounding of the same value, and can differ in the last
digit:

```ant
Math.log10(1000.0)                // 3.0
Math.log(1000.0) / Math.log(10.0) // 2.9999999999999996
```

`0.0` gives `-Infinity`, which panics when converted to an integer:

```ant
int(Math.log10(0.0)) // panic: -Infinity does not convert to int
```

An integer argument is a compile error, since a numeric literal without a point is an `int`:

```ant
Math.log10(100) // error: cannot use int where float64 is expected
```

## See also

[Math.log](log.md) · [Math.log2](log2.md) · [Math.exp](exp.md) · [Math.pow](pow.md) ·
[Math.floor](floor.md) · [Math](index.md)
