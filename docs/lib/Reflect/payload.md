<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.payload

```ant signature
func Reflect.payload(d: Reflect.Value): Reflect.Value
```

Reflect.payload returns the base value of the named primitive in d, boxed at the base type. It
panics if d's type is not a named primitive type.

<!-- end generated -->
