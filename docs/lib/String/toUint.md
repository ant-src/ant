<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toUint

```ant signature
func String.toUint(s: string, base: int = 10): uint64 | null
```

String.toUint parses s as an unsigned integer in base, or returns null when it is not one. s is
digits of base as for [`String.toInt`](toInt.md), with no sign. A number that does not fit a uint64
is null too. base defaults to 10, and it panics if base is outside 2 ..= 36.

<!-- end generated -->

## Basic examples

```ant
String.toUint("18446744073709551615") // 18446744073709551615
String.toUint("-1")                   // null
```
