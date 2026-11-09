<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.log

```ant signature
func Math.log(x: float64): float64
```

Math.log returns the natural logarithm of x. It is -Infinity for 0.0 and NaN for a negative x.

<!-- end generated -->

## Details

- `Math.log` never panics. `-0.0` also gives `-Infinity`, and `Infinity` gives `Infinity`.

## Basic examples

```ant
Math.log(1.0)  // 0.0
Math.log(10.0) // 2.302585092994046
Math.log(0.5)  // -0.6931471805599453
```

The logarithm of `Math.E` is `1.0`:

```ant
Math.log(Math.E) // 1.0
```

## Scope

Zero of either sign gives `-Infinity`, and a negative number gives `NaN`:

```ant
Math.log(0.0)  // -Infinity
Math.log(-0.0) // -Infinity
Math.log(-1.0) // NaN
```

The infinities and `NaN` follow IEEE 754:

```ant
Math.log(Math.Infinity)  // Infinity
Math.log(-Math.Infinity) // NaN
Math.log(Math.NaN)       // NaN
```

The extremes of `float64` stay finite:

```ant
Math.log(5e-324)          // -744.4400719213812
Math.log(Math.maxFloat64) // 709.782712893384
```

An integer goes through `float64` first:

```ant
let n = 100
Math.log(float64(n)) // 4.605170185988092
```

Pass it by name to `Array.map`, call it in a pipe, or name its parameter `x`:

```ant
[1.0, Math.E, 10.0].map(Math.log) // [0.0, 1.0, 2.302585092994046]
100.0 | Math.log()                // 4.605170185988092
Math.log(x: 100.0)                // 4.605170185988092
```

## Applications

Find how many years a deposit takes to double at 5% interest compounded yearly:

```ant
Math.log(2.0) / Math.log(1.05) // 14.206699082890461
```

Take a logarithm in any base by dividing two natural logarithms. The quotient can be off in the last
digit:

```ant
func logBase(b: float64, x: float64): float64 {
  Math.log(x) / Math.log(b)
}

logBase(2.0, 1024.0) // 10.0
logBase(16.0, 2.0)   // 0.25
logBase(3.0, 81.0)   // 4.000000000000001
```

Measure the entropy of a distribution, in nats. A zero probability adds nothing, so skip it rather
than multiply `0.0` by `-Infinity`:

```ant
func entropy(ps: Array[float64]): float64 {
  var total = 0.0
  for p in ps {
    if p > 0.0 {
      total -= p * Math.log(p)
    }
  }
  total
}

println(entropy([0.5, 0.5]))
println(entropy([1.0, 0.0]))
println(entropy([0.25, 0.25, 0.25, 0.25]))
```

```text
0.6931471805599453
0.0
1.3862943611198906
```

Average growth rates with the geometric mean, through the mean of the logarithms:

```ant
func geometricMean(xs: Array[float64]): float64 {
  let logs = xs.map(Math.log)
  Math.exp((Array.sum(logs) ?? 0.0) / float64(xs.length()))
}

println(geometricMean([1.1, 1.2, 0.9]))
```

```text
1.059104500597819
```

## Properties and relations

`Math.log` turns products into sums and undoes `Math.exp`, each up to rounding:

```ant
func close(a: float64, b: float64): bool {
  Math.abs(a - b) <= 1e-12 * Math.max(Math.abs(a), Math.abs(b))
}

let xs = [0.25, 1.5, 3.0, 10.0, 1e6]
for a in xs {
  assert(close(Math.exp(Math.log(a)), a))
  assert(close(Math.log(a * a), 2.0 * Math.log(a)))
  for b in xs {
    assert(close(Math.log(a * b), Math.log(a) + Math.log(b)))
  }
}
```

It is increasing, negative below `1.0` and positive above it:

```ant
let ys = [0.1, 0.5, 1.0, 2.0, 100.0].map(Math.log)
for i in Iter.range(1, ys.length()) {
  assert(ys[i - 1] < ys[i])
}
assert(ys[0] < 0.0 && ys[2] == 0.0 && ys[4] > 0.0)
```

## Possible issues

A negative or zero input does not panic, so the `NaN` or `-Infinity` spreads through the rest of a
computation. Check the input, or the result with [Math.isFinite](isFinite.md):

```ant
let r = Math.log(-4.0) * 2.0
r                // NaN
Math.isFinite(r) // false
```

Dividing logarithms to change the base can miss an exact power in the last digit, so rounding the
quotient down gives the wrong whole power:

```ant
Math.log(1000.0) / Math.log(10.0)             // 2.9999999999999996
Math.floor(Math.log(1000.0) / Math.log(10.0)) // 2.0
```

[Math.log10](log10.md) and [Math.log2](log2.md) avoid the extra rounding of the division, but they
are not exact on every power either:

```ant
Math.log10(1000.0) // 3.0
Math.log10(1e15)   // 14.999999999999998
```

For a whole power, round the result instead of flooring it, or count the power with integer
arithmetic:

```ant
Math.round(Math.log(1000.0) / Math.log(10.0)) // 3.0
Math.round(Math.log10(1e15))                  // 15.0
```

Near `1.0`, the sum `1.0 + x` loses the digits of a tiny `x` before `Math.log` sees them:

```ant
Math.log(1.0 + 1e-17) // 0.0
Math.log(1.0 + 1e-10) // 1.000000082690371e-10
```

An integer argument is a compile error, since a numeric literal without a point is an `int`:

```ant
Math.log(10) // error: cannot use int where float64 is expected
```

`Math.log` is not a function on `float64`, so it cannot be called through `.`:

```ant
2.0.log() // error: float64 has no field or function log
```

## See also

[Math.exp](exp.md) · [Math.log2](log2.md) · [Math.log10](log10.md) · [Math.pow](pow.md) ·
[Math.isFinite](isFinite.md) · [Math](index.md)
