<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fromInt

```ant signature
func String.fromInt[T: Integer](x: T, base: int = 10): string
```

String.fromInt returns x written in base, with lower-case digits and a - in front when negative.
base defaults to 10, and it panics if base is outside 2 ..= 36.

<!-- end generated -->

## Basic examples

```ant
String.fromInt(255, 16) // "ff"
```
