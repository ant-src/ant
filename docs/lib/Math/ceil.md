<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.ceil

```ant signature
func Math.ceil(x: float64): float64
```

Math.ceil returns the least whole number not below x.

<!-- end generated -->

## Details

- The result is a `float64`, not an `int`. Convert it with [int](../prelude/int.md), which panics
  when the value is out of range, `NaN`, or an infinity, or with [Math.toInt](toInt.md), which
  returns `null` instead.
- A whole number, `NaN`, and an infinity come back unchanged. `Math.ceil` never panics.

## Basic examples

Round a number up to a whole number:

```ant
Math.ceil(2.1) // 3.0
Math.ceil(2.9) // 3.0
Math.ceil(3.0) // 3.0
```

Negative numbers round up toward zero:

```ant
Math.ceil(-2.7) // -2.0
```

Convert the result to an `int`:

```ant
int(Math.ceil(2.1)) // 3
```

## Scope

`NaN` and the infinities come back unchanged:

```ant
Math.ceil(Math.NaN)       // NaN
Math.ceil(Math.Infinity)  // Infinity
Math.ceil(-Math.Infinity) // -Infinity
```

A number between -1 and 0 rounds up to `-0.0`, which compares equal to `0.0`:

```ant
Math.ceil(-0.5)        // -0.0
Math.ceil(-0.5) == 0.0 // true
```

Every `float64` from 2^52 up is already whole, so a large number comes back as it is:

```ant
Math.ceil(1e300)           // 1e+300
Math.ceil(Math.maxFloat64) // 1.7976931348623157e+308
```

The smallest positive `float64` still rounds up to 1:

```ant
Math.ceil(4.9e-324) // 1.0
```

Call it in a pipe, or pass it as a function value:

```ant
2.1 | Math.ceil()               // 3.0
[0.2, 1.5, -1.5].map(Math.ceil) // [1.0, 2.0, -1.0]
```

## Applications

Count the pages needed for a list of items, with a fixed number of items per page:

```ant
let items = 23
let perPage = 10
let pages = int(Math.ceil(float64(items) / float64(perPage)))
println("${pages} pages")
```

```text
3 pages
```

Count the paint cans needed for a wall, when each can covers a fixed area:

```ant
let wall = 23.5
let perCan = 5.0
let cans = int(Math.ceil(wall / perCan))
println("${cans} cans")
```

```text
5 cans
```

Charge a parking fee per started hour:

```ant
type Stay = { plate: string, minutes: int }
let rate = 2.5
let stays = [Stay{ plate: "AB-12", minutes: 45 }, Stay{ plate: "CD-34", minutes: 121 }]
for stay in stays {
  let hours = Math.ceil(float64(stay.minutes) / 60.0)
  println("${stay.plate}: ${hours * rate}")
}
```

```text
AB-12: 2.5
CD-34: 7.5
```

## Properties and relations

`Math.ceil(x)` is the negation of [Math.floor](floor.md) of `-x`:

```ant
for x in [2.1, -2.7, 3.0, 0.5] {
  assert(Math.ceil(x) == -Math.floor(-x))
}
```

The result is never below `x`, and less than one above it:

```ant
for x in [2.1, -2.7, 3.0, 0.001] {
  let c = Math.ceil(x)
  assert(c >= x && c - x < 1.0)
}
```

For a positive number with a fraction, `Math.ceil` is one more than [Math.trunc](trunc.md); for a
negative one, they are equal:

```ant
assert(Math.ceil(2.4) == Math.trunc(2.4) + 1.0)
assert(Math.ceil(-2.4) == Math.trunc(-2.4))
```

Applying it twice gives the same result as once:

```ant
let x = 7.25
assert(Math.ceil(Math.ceil(x)) == Math.ceil(x))
```

## Possible issues

`Math.ceil` takes a `float64` only, so an `int` is a compile error. Integer division has already
dropped the fraction, so convert the operands before dividing:

```ant
let items = 23
let pages = Math.ceil(items / 10) // error: cannot use int where float64 is expected
```

```ant
let items = 23
Math.ceil(float64(items) / 10.0) // 3.0
```

A `float32` or a named type over `float64` needs a conversion too:

```ant
type Price = float64
let p = Price(2.5)
Math.ceil(float64(p))            // 3.0
Price(Math.ceil(float64(p)))     // Price(3.0)
Math.ceil(float64(float32(2.5))) // 3.0
```

The result of a float computation may lie just above a whole number, and `Math.ceil` then rounds it
up by a whole step:

```ant
1.1 * 100.0            // 110.00000000000001
Math.ceil(1.1 * 100.0) // 111.0
```

So rounding a price up to whole cents by scaling it by 100 can add a cent to a price that is already
whole:

```ant
Math.ceil(1.1 * 100.0) / 100.0 // 1.11
```

Converting a result too large for an `int` panics. Use [Math.toInt](toInt.md) to get `null` instead:

```ant
int(Math.ceil(1e300))        // panic: 1e+300 out of range for int
Math.toInt(Math.ceil(1e300)) // null
```

## See also

[Math.floor](floor.md) · [Math.round](round.md) · [Math.trunc](trunc.md) · [Math.toInt](toInt.md) ·
[int](../prelude/int.md) · [float64](../prelude/float64.md)
