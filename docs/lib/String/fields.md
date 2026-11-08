<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.fields

```ant signature
func String.fields(s: string): Array[string]
```

String.fields returns the runs of s between Unicode white space, or \[\] when there are none.

<!-- end generated -->

## Basic examples

```ant
String.fields("  a \t b\n") // ["a", "b"]
```
