<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.range

```ant signature
func Array.range(from: int, to: int): Array[int]
```

Array.range returns the integers from from up to but excluding to. It returns an empty array when to
\<= from.

<!-- end generated -->

## Basic examples

```ant
Array.range(2, 5) // [2, 3, 4]
```
