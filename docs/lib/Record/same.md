<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Record.same

```ant signature
func Record.same[T: {}](a: T, b: T): bool
```

Record.same reports whether a and b are the same record, not merely equal ones. The bound \{\}
admits any record type.

<!-- end generated -->

## Basic examples

```ant
var e = Error{ message: "x" }
Record.same(e, e)             // true
Record.same(e, Error{ ...e }) // false
```
