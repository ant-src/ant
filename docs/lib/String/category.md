<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.category

```ant signature
func String.category(c: string): string
```

String.category returns the two-letter Unicode general category of c, such as "Lu", or "Cn" when c
is unassigned. It panics unless c is one code point.

<!-- end generated -->
