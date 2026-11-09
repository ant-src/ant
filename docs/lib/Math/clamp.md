<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.clamp

```ant signature
func Math.clamp[T: Ordered](x: T, lo: T, hi: T): T
```

Math.clamp returns x held within lo ..= hi. It panics if lo \> hi or a bound is NaN; a NaN x gives
NaN.

<!-- end generated -->

## Details

- Both bounds are inclusive: the result is `lo` when `x` is below it, `hi` when `x` is above it, and
  `x` otherwise.

## Basic examples

Hold a number within 0 to 10:

```ant
Math.clamp(15, 0, 10) // 10
Math.clamp(-3, 0, 10) // 0
Math.clamp(7, 0, 10)  // 7
```

Hold a fraction within 0.0 to 1.0:

```ant
Math.clamp(1.5, 0.0, 1.0) // 1.0
```

## Scope

A value equal to a bound is already within the range, and `lo` may equal `hi`:

```ant
Math.clamp(10, 0, 10) // 10
Math.clamp(8, 3, 3)   // 3
```

The infinities clamp like any other float, and an infinite bound leaves that side open:

```ant
Math.clamp(Math.Infinity, 0.0, 1.0)              // 1.0
Math.clamp(-Math.Infinity, 0.0, Math.Infinity)   // 0.0
Math.clamp(250.0, -Math.Infinity, Math.Infinity) // 250.0
```

A `NaN` `x` passes through:

```ant
Math.clamp(Math.NaN, 0.0, 1.0) // NaN
```

Every integer type works, with all three arguments of that type:

```ant
Math.clamp(byte(200), byte(10), byte(100)) // 100
Math.clamp(Math.minInt, -1, 1)             // -1
```

Strings clamp by the order of `<`:

```ant
Math.clamp("zebra", "a", "m") // "m"
Math.clamp("cat", "a", "m")   // "cat"
```

Call it in a pipe, where the value on the left becomes `x`:

```ant
15 | Math.clamp(0, 10) // 10
```

Pass it to `map` in a lambda to clamp every element:

```ant
[-5, 3, 12].map({ Math.clamp(x, 0, 10) }) // [0, 3, 10]
```

## Named arguments

Name the arguments to make the bounds plain at the call site, in any order:

```ant
Math.clamp(x: 7, lo: 0, hi: 5) // 5
Math.clamp(hi: 5, lo: 0, x: 7) // 5
```

## Applications

Keep a requested page number within the pages that exist:

```ant
func page(requested: int, pages: int): int {
  Math.clamp(requested, 1, pages)
}
page(0, 12)  // 1
page(40, 12) // 12
page(5, 12)  // 5
```

Draw a progress bar from a fraction that may overshoot:

```ant
func bar(done: float64): string {
  let filled = int(Math.round(Math.clamp(done, 0.0, 1.0) * 10.0))
  String.repeat("#", filled) + String.repeat(".", 10 - filled)
}
bar(0.35) // "####......"
bar(1.2)  // "##########"
bar(-0.1) // ".........."
```

Move a cursor through an array without running off either end:

```ant
let items = ["a", "b", "c"]
var cursor = 0
cursor = Math.clamp(cursor - 1, 0, len(items) - 1)
cursor // 0
cursor = Math.clamp(cursor + 5, 0, len(items) - 1)
cursor // 2
```

Limit each color channel of a record to the range a byte holds:

```ant
type Rgb = { r: int, g: int, b: int }
func channel(v: int): int { Math.clamp(v, 0, 255) }
func Rgb.brighten(c: Rgb, by: int): Rgb {
  Rgb{ r: channel(c.r + by), g: channel(c.g + by), b: channel(c.b + by) }
}
Rgb.brighten(Rgb{ r: 200, g: 120, b: 10 }, 80) // Rgb{ r: 255, g: 200, b: 90 }
```

## Properties and relations

The result always lies within `lo ..= hi`:

```ant
for x in [-100, -1, 0, 4, 9, 100] {
  let c = Math.clamp(x, 0, 9)
  assert(c >= 0 && c <= 9)
}
```

Clamping twice to one range changes nothing the second time:

```ant
let once = Math.clamp(42, 0, 10)
assert(Math.clamp(once, 0, 10) == once)
```

For integers, it is `Math.max` followed by `Math.min`:

```ant
for x in [-3, 5, 15] {
  assert(Math.clamp(x, 0, 10) == Math.min(Math.max(x, 0), 10))
}
```

## Possible issues

Bounds in the wrong order panic rather than being swapped:

```ant
Math.clamp(5, 10, 0) // panic: Math.clamp needs lo <= hi, got 10 and 0
```

A `NaN` bound panics, though a `NaN` `x` does not:

```ant
Math.clamp(0.5, Math.NaN, 1.0) // panic: Math.clamp needs bounds that are not NaN, got NaN and 1.0
```

An integer literal is always an `int`, so it cannot bound a `float64`. Write the bounds as float
literals:

```ant
let ratio = 1.5
let r = Math.clamp(ratio, 0.0, 1) // error: cannot use int where float64 is expected
```

A named primitive type is not a member of `Ordered`, so clamp its payload and tag the result:

```ant
type Celsius = float64
let t = Celsius(40.0)
let held = Math.clamp(t, Celsius(-10.0), Celsius(35.0))
// error: T would be Celsius, which is not a member of its bound Ordered; unwrap it with float64(...)
```

```ant
type Celsius = float64
let t = Celsius(40.0)
Celsius(Math.clamp(float64(t), -10.0, 35.0)) // Celsius(35.0)
```

## See also

- [Math.min](min.md) and [Math.max](max.md): hold a value on one side only.
- [Math.round](round.md)
