<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Json.parse

```ant signature
func Json.parse(text: string): Result[Json.Value, Error]
```

Json.parse parses text as one JSON value. A repeated key keeps its first place and its last value.
It is Err when text is not valid JSON, nests arrays and objects deeper than 10000, or holds a number
too large for a float64.

<!-- end generated -->

## Basic examples

```ant
Json.parse("{\"a\": [1, 2.5]}") // Ok{ value: Map{ "a": [1.0, 2.5] } }
```
