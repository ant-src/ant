<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.log2

```ant signature
func Math.log2(x: float64): float64
```

Math.log2 returns the base-2 logarithm of x.

<!-- end generated -->

## Details

- `Math.log2` never panics. `0.0` and `-0.0` give `-Infinity`, a negative `x` gives `NaN`, and `NaN`
  gives `NaN`.

## Basic examples

```ant
Math.log2(8.0)    // 3.0
Math.log2(1024.0) // 10.0
Math.log2(10.0)   // 3.321928094887362
```

Fractions below `1.0` have negative logarithms:

```ant
Math.log2(0.5)  // -1.0
Math.log2(0.25) // -2.0
```

## Scope

`1.0` gives `0.0`, and both zeros give `-Infinity`:

```ant
Math.log2(1.0)  // 0.0
Math.log2(0.0)  // -Infinity
Math.log2(-0.0) // -Infinity
```

A negative number has no real logarithm, so the result is `NaN`. The infinities and `NaN` follow
IEEE 754:

```ant
Math.log2(-1.0)           // NaN
Math.log2(Math.Infinity)  // Infinity
Math.log2(-Math.Infinity) // NaN
Math.log2(Math.NaN)       // NaN
```

The largest finite `float64` and the smallest positive one stay in range:

```ant
Math.log2(Math.maxFloat64) // 1024.0
Math.log2(5e-324)          // -1074.0
```

An integer goes through `float64` first:

```ant
let n = 4096
Math.log2(float64(n)) // 12.0
```

Pass it by name to `Array.map`, call it in a pipe, or name the parameter `x`:

```ant
[1.0, 2.0, 4.0, 8.0].map(Math.log2) // [0.0, 1.0, 2.0, 3.0]
16.0 | Math.log2()                  // 4.0
Math.log2(x: 32.0)                  // 5.0
```

## Applications

Count the bits needed to number `n` distinct values:

```ant
func bitsFor(n: int): int {
  int(Math.ceil(Math.log2(float64(n))))
}

[2, 10, 256, 1000].map(bitsFor) // [1, 4, 8, 10]
```

Measure the Shannon entropy of a distribution, in bits:

```ant
func entropy(ps: Array[float64]): float64 {
  var h = 0.0
  for p in ps {
    if p > 0.0 {
      h -= p * Math.log2(p)
    }
  }
  h
}

entropy([0.5, 0.5])               // 1.0
entropy([0.25, 0.25, 0.25, 0.25]) // 2.0
entropy([0.9, 0.1])               // 0.4689955935892812
```

Find the interval between two frequencies in semitones: an octave doubles the frequency and has 12
semitones:

```ant
func semitones(from: float64, to: float64): float64 {
  12.0 * Math.log2(to / from)
}

semitones(440.0, 880.0)                             // 12.0
Math.round(semitones(440.0, 660.0) * 100.0) / 100.0 // 7.02
```

## Properties and relations

`Math.log2` turns products into sums, `Math.pow(2.0, _)` undoes it, and it equals
`Math.log(x) / Math.log(2.0)`, each up to rounding:

```ant
func close(a: float64, b: float64): bool {
  Math.abs(a - b) <= 1e-12 * Math.max(1.0, Math.max(Math.abs(a), Math.abs(b)))
}

for x in [0.1, 0.5, 3.0, 10.0, 1e6] {
  assert(close(Math.log2(x * 3.0), Math.log2(x) + Math.log2(3.0)))
  assert(close(Math.pow(2.0, Math.log2(x)), x))
  assert(close(Math.log2(x), Math.log(x) / Math.log(2.0)))
}
```

For a positive `int`, the floor of its base-2 logarithm is the index of its highest set bit.
[Bits.leadingZeros](../Bits/leadingZeros.md) gives that index exactly, without going through
`float64`:

```ant
Math.floor(Math.log2(1000.0)) // 9.0
63 - Bits.leadingZeros(1000)  // 9
```

## Possible issues

`Math.log(x) / Math.log(2.0)` rounds twice, so its result can differ from `Math.log2(x)`:

```ant
let x = Math.pow(2.0, 29.0)
Math.log2(x)                // 29.0
Math.log(x) / Math.log(2.0) // 29.000000000000004
```

A large `int` rounds when it becomes a `float64`, and the logarithm rounds again, so the floor of
its logarithm can be one too high. `Bits.leadingZeros` stays exact:

```ant
let n = Bits.shiftLeft(1, 50) - 1
Math.floor(Math.log2(float64(n)))           // 50.0
63 - Bits.leadingZeros(n)                   // 49
Math.floor(Math.log2(float64(Math.maxInt))) // 63.0
63 - Bits.leadingZeros(Math.maxInt)         // 62
```

Converting the result to `int` panics when it has a fraction. Round it first, or use
[Math.toInt](toInt.md) to get `null`:

```ant
int(Math.log2(1000.0))             // panic: 9.965784284662087 has a fraction, so it does not convert to int
int(Math.floor(Math.log2(1000.0))) // 9
Math.toInt(Math.log2(1000.0))      // null
```

An integer argument is a compile error, since a numeric literal without a point is an `int`:

```ant
Math.log2(8) // error: cannot use int where float64 is expected
```

## See also

[Math.log](log.md) · [Math.log10](log10.md) · [Math.exp](exp.md) · [Math.pow](pow.md) ·
[Bits.leadingZeros](../Bits/leadingZeros.md) · [Math.toInt](toInt.md) · [Math](index.md)
