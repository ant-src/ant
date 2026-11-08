<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.count

```ant signature
func String.count(s: string, t: string): int
```

String.count returns the number of non-overlapping occurrences of t in s. It panics if t is empty.

<!-- end generated -->

## Basic examples

```ant
String.count("aaaa", "aa") // 2
```
