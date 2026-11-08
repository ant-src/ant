<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Csv.parse

```ant signature
func Csv.parse(text: string, sep: string = ","): Result[Array[Array[string]], Error]
```

Csv.parse parses text as RFC 4180 CSV with sep between fields. Each row keeps as many fields as
written, nothing is trimmed, and empty lines are dropped; it is Err, naming the line, on a quote
inside an unquoted field or a quoted field that never closes. It panics if sep is not one code point
other than ", \r, \n, U+0000, and U+FFFD.

<!-- end generated -->

## Basic examples

```ant
Csv.parse("a,\"b,c\"\n1,2\n") // Ok{ value: [["a", "b,c"], ["1", "2"]] }
```
