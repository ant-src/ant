<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toFloat

```ant signature
func String.toFloat(s: string): float64 | null
```

String.toFloat parses s as a decimal number, or returns null when it is not one. s is ASCII digits
with an optional sign, fraction after ., and exponent after e or E; the result is rounded to the
nearest float64, and a value beyond its range is null. It also reads NaN, and Inf or Infinity with
an optional sign, in any case, so what [String.from](from.md) writes reads back. White space, _, and
a prefix such as 0x are not numbers.

<!-- end generated -->

## Basic examples

```ant
String.toFloat(".5")        // 0.5
String.toFloat("1e+3")      // 1000.0
String.toFloat("-Infinity") // -Infinity
```
