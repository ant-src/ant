<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.repeat

```ant signature
func String.repeat(s: string, n: int): string
```

String.repeat returns s repeated n times. It panics if n \< 0 or n is above 2^31 - 1.

<!-- end generated -->
