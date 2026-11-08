<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.stringifyQuery

```ant signature
func Url.stringifyQuery(params: Map[string, Array[string]]): string
```

Url.stringifyQuery writes params as k=v pairs joined by &, in the map's order, each key and value
escaped as [Url.escape](escape.md) does, so it round-trips with [Url.parseQuery](parseQuery.md).

<!-- end generated -->

## Basic examples

```ant
Url.stringifyQuery(Map{ "q": ["a b"], "tag": ["x", "y"] }) // "q=a+b&tag=x&tag=y"
```
