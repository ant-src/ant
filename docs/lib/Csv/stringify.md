<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Csv.stringify

```ant signature
func Csv.stringify(rows: Array[Array[string]], sep: string = ","): string
```

Csv.stringify writes rows as CSV with sep between fields, a line per row ending in \\n. A field is
quoted when it holds sep, a quote, or a line break, or begins with white space, and a row of one
empty field is written "". It panics if sep is not a separator [`Csv.parse`](parse.md) accepts.

<!-- end generated -->

## Basic examples

```ant
Csv.stringify([["a", "b,c"], ["1", "2"]]) // "a,\"b,c\"\n1,2\n"
```
