<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toInt

```ant signature
func String.toInt(s: string, base: int = 10): int | null
```

String.toInt parses s as an integer in base, or returns null when it is not one. s is digits of base
with an optional sign and no prefix; only 0 to 9 and the letters a to z, in either case, count as
digits. A number that does not fit an int is null too. base defaults to 10, and it panics if base is
outside 2 ..= 36.

<!-- end generated -->

## Basic examples

```ant
String.toInt("-FF", 16) // -255
String.toInt("0x10")    // null
```
