<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.lines

```ant signature
func String.lines(s: string): Array[string]
```

String.lines splits s at every \n, dropping a \r just before one. The piece after the last \n is
omitted when it is empty, so a final newline adds no empty line.

<!-- end generated -->

## Basic examples

```ant
String.lines("a\r\nb\n") // ["a", "b"]
String.lines("\n")       // [""]
String.lines("")         // []
```
