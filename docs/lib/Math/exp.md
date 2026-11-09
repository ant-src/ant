<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.exp

```ant signature
func Math.exp(x: float64): float64
```

Math.exp returns e raised to the power x.

<!-- end generated -->

## Details

- `Math.exp` never panics. A large `x` overflows to `Infinity`, a very negative `x` underflows to
  `0.0`, and `NaN` gives `NaN`.

## Basic examples

```ant
Math.exp(0.0)  // 1.0
Math.exp(1.0)  // 2.718281828459045
Math.exp(-1.0) // 0.36787944117144233
```

## Scope

Negative inputs give values between `0.0` and `1.0`, and `-0.0` gives `1.0`:

```ant
Math.exp(-2.0) // 0.1353352832366127
Math.exp(-0.0) // 1.0
```

The infinities and `NaN` follow IEEE 754:

```ant
Math.exp(Math.Infinity)  // Infinity
Math.exp(-Math.Infinity) // 0.0
Math.exp(Math.NaN)       // NaN
```

An integer goes through `float64` first:

```ant
let n = 3
Math.exp(float64(n)) // 20.085536923187668
```

Pass it by name to `Array.map`, or call it in a pipe:

```ant
[0.0, 1.0, 2.0].map(Math.exp) // [1.0, 2.718281828459045, 7.38905609893065]
1.0 | Math.exp()              // 2.718281828459045
```

## Applications

Grow a deposit at 5% interest compounded continuously for ten years:

```ant
let principal = 1000.0
let rate = 0.05
principal * Math.exp(rate * 10.0) // 1648.7212707001281
```

Find how much carbon-14 is left after two half-lives, from the decay constant:

```ant
let halfLife = 5730.0
let k = Math.log(2.0) / halfLife
Math.exp(-k * 11460.0) // 0.25
```

Squash any score into `0.0 ..= 1.0` with the logistic function:

```ant
func sigmoid(x: float64): float64 {
  1.0 / (1.0 + Math.exp(-x))
}

println([-2.0, 0.0, 2.0].map(sigmoid))
```

```text
[0.11920292202211755, 0.5, 0.8807970779778823]
```

Turn scores into probabilities with softmax. Subtracting the largest score first keeps every
`Math.exp` finite, even for large scores:

```ant
func softmax(xs: Array[float64]): Array[float64] {
  let top = Array.max(xs) ?? 0.0
  let ws = xs.map({ Math.exp(x - top) })
  let total = Array.sum(ws) ?? 0.0
  ws.map({ x / total })
}

println(softmax([1.0, 2.0, 3.0]))
println(softmax([1000.0, 1001.0]))
```

```text
[0.09003057317038046, 0.24472847105479764, 0.6652409557748218]
[0.2689414213699951, 0.7310585786300049]
```

## Properties and relations

`Math.exp` never returns a negative number and never decreases, `Math.log` undoes it, and
`exp(x) * exp(-x)` is `1.0`, each up to rounding for inputs that neither overflow nor underflow:

```ant
func close(a: float64, b: float64): bool {
  Math.abs(a - b) <= 1e-12 * Math.max(Math.abs(a), Math.abs(b))
}

let xs = [-3.0, -1.0, 0.0, 0.3, 2.0]
let ys = xs.map(Math.exp)
for i in Iter.range(0, xs.length()) {
  assert(ys[i] >= 0.0)
  assert(close(Math.log(ys[i]), xs[i]))
  assert(close(ys[i] * Math.exp(-xs[i]), 1.0))
  if i > 0 {
    assert(ys[i - 1] <= ys[i])
  }
}
```

## Possible issues

The result overflows to `Infinity` a little above `709.0`:

```ant
Math.exp(709.0)                // 8.218407461554972e+307
Math.exp(710.0)                // Infinity
Math.isFinite(Math.exp(710.0)) // false
```

So a ratio of two exponentials that both overflow is `Infinity / Infinity`, which is `NaN`, though
the true ratio is about `0.37`. Shift the inputs first, as `softmax` does above:

```ant
let ws = [1000.0, 1001.0].map(Math.exp)
ws[0] / ws[1]             // NaN
Math.exp(1000.0 - 1001.0) // 0.36787944117144233
```

A very negative input underflows to `0.0`, and the logarithm can no longer recover it:

```ant
Math.exp(-746.0)           // 0.0
Math.log(Math.exp(-746.0)) // -Infinity
```

Work with logarithms instead of the exponentials themselves, or shift the inputs as `softmax` does.

Rounding can make `Math.log(Math.exp(x))` differ from `x` in the last digit, so compare with a
tolerance rather than `==`:

```ant
Math.log(Math.exp(0.3)) // 0.30000000000000004
```

`Math.exp(x)` and `Math.pow(Math.E, x)` can differ in the last digit too:

```ant
Math.exp(2.0)         // 7.38905609893065
Math.pow(Math.E, 2.0) // 7.3890560989306495
```

An integer argument is a compile error, since a numeric literal without a point is an `int`:

```ant
Math.exp(1) // error: cannot use int where float64 is expected
```

`Math.exp` is not a function on `float64`, so it cannot be called through `.`:

```ant
2.0.exp() // error: float64 has no field or function exp
```

## See also

[Math.log](log.md) · [Math.pow](pow.md) · [Math.log2](log2.md) · [Math.log10](log10.md) ·
[Math.isFinite](isFinite.md) · [Math](index.md)
