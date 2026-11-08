<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.scalar

```ant signature
func Reflect.scalar(d: Reflect.Value): Reflect.Scalar
```

Reflect.scalar returns the value in d in its own kind. It panics if d's type is not null, void,
bool, a number, or string.

<!-- end generated -->
