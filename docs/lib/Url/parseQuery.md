<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.parseQuery

```ant signature
func Url.parseQuery(query: string): Result[Map[string, Array[string]], Error]
```

Url.parseQuery parses query, k=v pairs joined by &, decoding each key and value as
[`Url.unescape`](unescape.md) does. Keys are in order of first appearance, each with its values in
order, and a key without = has the value "". It is Err where [`Url.unescape`](unescape.md) is.

<!-- end generated -->

## Basic examples

```ant
Url.parseQuery("tag=a&tag=b&q") // Ok{ value: Map{ "tag": ["a", "b"], "q": [""] } }
```
