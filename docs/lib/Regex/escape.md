<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.escape

```ant signature
func Regex.escape(s: string): string
```

Regex.escape returns s with every metacharacter escaped, so a pattern compiled from it matches s
literally.

<!-- end generated -->

## Basic examples

```ant
Regex.escape("a.b*c") // "a\\.b\\*c"
```
