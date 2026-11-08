<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.Pattern

```ant signature
type Regex.Pattern
```

Regex.Pattern is a compiled regular expression, from a literal /.../flags or
[`Regex.compile`](../compile.md). It is immutable and compares by its normalized source and flags,
so it serves as a map key.

<!-- end generated -->
