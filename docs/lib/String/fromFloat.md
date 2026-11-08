<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fromFloat

```ant signature
func String.fromFloat(x: float64, digits: int): string
```

String.fromFloat returns x in fixed point with digits decimals, rounded half to even. NaN and the
infinities render as [String.from](from.md) does. It panics if digits < 0.

<!-- end generated -->

## Basic examples

```ant
String.fromFloat(0.125, 2) // "0.12"
```
