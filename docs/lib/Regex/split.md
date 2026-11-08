<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.split

```ant signature
func Regex.split(p: Regex.Pattern, s: string): Array[string]
```

Regex.split returns the pieces of s between the matches of p. A pattern that never matches gives
[s], an empty s gives [""], and an empty match at the start or end of s adds no empty piece.

<!-- end generated -->

## Basic examples

```ant
Regex.split(/a*/, "baaab") // ["b", "b"]
```
