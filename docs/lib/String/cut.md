<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.cut

```ant signature
func String.cut(s: string, sep: string): String.Cut | null
```

String.cut splits s around the first occurrence of sep, or returns null when sep does not occur. It
panics if sep is empty.

<!-- end generated -->

## Basic examples

```ant
String.cut("k=v=w", "=") // String.Cut{ before: "k", after: "v=w" }
```
