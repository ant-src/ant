<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toFloat

```ant signature
func String.toFloat(s: string): float64 | null
```

String.toFloat parses s as a decimal number, or returns null when it is not one. s is ASCII digits
with an optional sign, fraction after ., and exponent after e or E; the result is rounded to the
nearest float64, and a value beyond its range is null. It also reads NaN, and Inf or Infinity with
an optional sign, in any case, so what [`String.from`](from.md) writes reads back. White space, \_,
and a prefix such as 0x are not numbers.

<!-- end generated -->

## Details

- Reads ASCII digits with an optional sign, fraction, and exponent, so `.5` and `5.` are numbers. It
  also reads `NaN`, `Inf`, and `Infinity`, in any case.
- Anything else is `null`: white space, `_`, `,`, a prefix such as `0x`, and digits of other
  scripts.
- A value beyond the range of `float64` is `null`, not an infinity; a value too small to represent
  is `0.0`.

## Basic examples

Parse a decimal number:

```ant
String.toFloat("3.14") // 3.14
String.toFloat("-1.5") // -1.5
```

A string that is not a number gives `null`:

```ant
String.toFloat("abc") // null
```

Supply a default with `??`:

```ant
String.toFloat("n/a") ?? 0.0 // 0.0
```

## Scope

A whole number becomes a `float64`; a leading `+` and leading zeros are allowed:

```ant
String.toFloat("42")    // 42.0
String.toFloat("+7")    // 7.0
String.toFloat("00012") // 12.0
```

Either side of the `.` may be missing, but not both:

```ant
String.toFloat(".5") // 0.5
String.toFloat("5.") // 5.0
String.toFloat(".")  // null
```

An exponent follows `e` or `E`, with its own optional sign, and needs at least one digit:

```ant
String.toFloat("1e+3")   // 1000.0
String.toFloat("2.5E-2") // 0.025
String.toFloat("1e")     // null
String.toFloat("e5")     // null
```

`NaN` and the infinities are read in any case, and `Inf` is short for `Infinity`:

```ant
String.toFloat("NaN")       // NaN
String.toFloat("inf")       // Infinity
String.toFloat("-INF")      // -Infinity
String.toFloat("+Infinity") // Infinity
```

`String.from` writes negative zero as `-0.0`, and it reads back with its sign:

```ant
String.toFloat("-0.0") // -0.0
```

The result is the nearest `float64`. A value too small to represent rounds to `0.0`, and a value
beyond the range of `float64` is `null`:

```ant
String.toFloat("18014398509481985") // 18014398509481984.0
String.toFloat("1e-400")            // 0.0
String.toFloat("1e400")             // null
```

The empty string, white space, `_`, `,`, and prefixes are not numbers:

```ant
String.toFloat("")      // null
String.toFloat(" 1")    // null
String.toFloat("1_000") // null
String.toFloat("1,5")   // null
String.toFloat("0x10")  // null
```

Only ASCII digits count, so an Arabic-Indic digit or the minus sign `−` (U+2212) gives `null`:

```ant
String.toFloat("٣")  // null
String.toFloat("−1") // null
```

The result is always `float64`; convert it to get a `float32`:

```ant
float32(String.toFloat("1.5") ?? 0.0) // 1.5
```

Call it with a named argument, or in a pipe:

```ant
String.toFloat(s: "3.25") // 3.25
"2.5" | String.toFloat()  // 2.5
```

## Applications

Sum the numeric fields of a line and count the ones that do not parse. Narrow with `!= null` to use
the number:

```ant
let line = "3.5, 2, oops, 1e1"
var total = 0.0
var bad = 0
for field in String.split(line, ",") {
  let n = String.toFloat(String.trim(field))
  if n != null {
    total += n
  } else {
    bad += 1
  }
}
println("total ${total}, ${bad} bad")
```

```text
total 15.5, 1 bad
```

Read a setting with a fallback when it is missing or malformed:

```ant
let settings = Map{ "ratio": "0.75", "scale": "big" }

func setting(name: string, fallback: float64): float64 {
  String.toFloat(settings[name] ?? "") ?? fallback
}

println(setting("ratio", 1.0))
println(setting("scale", 1.0))
println(setting("depth", 2.0))
```

```text
0.75
1.0
2.0
```

Validate a whole column with [`Array.every`](../Array/every.md):

```ant
let column = ["1.5", "2", "-0.25"]
Array.every(column, { String.toFloat(x) != null }) // true
```

## Properties and relations

`String.toFloat` reads back what [`String.from`](from.md) writes, for every finite value and the
infinities:

```ant
for x in [0.1 + 0.2, -1.5, 1e21, 5e-324, 1.7976931348623157e308, Math.Infinity, -Math.Infinity] {
  assert(String.toFloat(String.from(x)) == x)
}
```

`NaN` reads back too, but `NaN` never equals itself, so test it with
[`Math.isNaN`](../Math/isNaN.md):

```ant
let n = String.toFloat(String.from(Math.NaN))
assert(n != null && Math.isNaN(n))
```

For an integer that fits, it agrees with [`String.toInt`](toInt.md) and a conversion:

```ant
for s in ["0", "-17", "123456"] {
  assert(String.toFloat(s) == float64(String.toInt(s) ?? 0))
}
```

It reads what [`String.fromFloat`](fromFloat.md) writes, up to the rounding `fromFloat` applies:

```ant
String.toFloat(String.fromFloat(2.675, 2)) // 2.67
```

## Possible issues

The result can be `null`, so arithmetic on it does not compile until you remove the `null`:

```ant
let x = String.toFloat("1.5") * 2.0
// error: * needs both operands of one type, got float64 | null and float64
```

```ant
(String.toFloat("1.5") ?? 0.0) * 2.0 // 3.0
```

Surrounding white space makes the whole string invalid. Trim it first with [`String.trim`](trim.md):

```ant
String.toFloat(" 2.5\n")              // null
String.toFloat(String.trim(" 2.5\n")) // 2.5
```

A number beyond the range of `float64` is `null`, not `Infinity`, so you cannot tell it from text
that is not a number:

```ant
String.toFloat("1e309")    // null
String.toFloat("Infinity") // Infinity
```

`"NaN"` parses, so a parsed value can still be `NaN`. It compares unequal to everything, itself
included:

```ant
let v = String.toFloat("nan") ?? 0.0
v == v // false
```

A large integer loses precision as a `float64`. Use [`String.toInt`](toInt.md) when you need the
exact integer:

```ant
String.toFloat("18014398509481985") // 18014398509481984.0
String.toInt("18014398509481985")   // 18014398509481985
```

## See also

- [`String.toInt`](toInt.md) and [`String.toUint`](toUint.md) parse integers.
- [`String.from`](from.md) writes a float that `String.toFloat` reads back.
- [`String.fromFloat`](fromFloat.md) writes a float in fixed point.
- [`Math.isNaN`](../Math/isNaN.md) tests for `NaN`.
