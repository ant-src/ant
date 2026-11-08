<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Json.Value

```ant signature
type Json.Value = null | bool | float64 | string | Array[Json.Value] | Map[string, Json.Value]
```

Json.Value is a JSON value: null, a boolean, a number, a string, an array, or an object. Every
number is a float64; [`Math.toInt`](../../Math/toInt.md) converts one to an int when it is whole and
in range.

<!-- end generated -->
