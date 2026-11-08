<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.rel

```ant signature
func Path.rel(base: string, target: string): Result[string, Error]
```

Path.rel returns target as a path relative to base, working on the cleaned text. It is Err when no
relative path leads there: when one path is absolute and the other is not, when they are on
different Windows volumes, or when base climbs out with .. further than target does.

<!-- end generated -->

## Basic examples

```ant
Path.rel("/a/b", "/a/c/d") // Ok{ value: "../c/d" }
```
