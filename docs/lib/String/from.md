<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.from

```ant signature
func String.from[T](value: T): string
```

String.from renders any value as text, in the format print, println, and string interpolation use. A
string renders as itself at the top level and quoted with escapes inside a container; a float
renders as the shortest decimal that reads back to it, with .0 when integral.

<!-- end generated -->

## Basic examples

```ant
String.from(["a", "b"]) // "[\"a\", \"b\"]"
String.from(1.0)        // "1.0"
```
