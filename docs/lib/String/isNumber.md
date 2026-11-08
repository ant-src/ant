<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isNumber

```ant signature
func String.isNumber(s: string): bool
```

String.isNumber reports whether every code point of s is a number (N): Nd, Nl such as Ⅻ, or No such
as ½. It is false on "".

<!-- end generated -->
