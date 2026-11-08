<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.split

```ant signature
func String.split(s: string, sep: string): Array[string]
```

String.split returns the pieces of s between each occurrence of sep. It panics if sep is empty.

<!-- end generated -->

## Basic examples

```ant
String.split("a,b,,c", ",") // ["a", "b", "", "c"]
String.split("", ",")       // [""]
```
