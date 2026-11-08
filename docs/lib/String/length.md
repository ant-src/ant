<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.length

```ant signature
func String.length(s: string): int
```

String.length returns the number of code points in s, not the graphemes a reader sees.

<!-- end generated -->

## Basic examples

```ant
String.length("e\u{301}") // 2
```
