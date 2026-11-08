<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.shape

```ant signature
func Reflect.shape(t: Reflect.ErasedType): Reflect.Shape
```

Reflect.shape returns one level of t's structure: aliases expanded, unions flattened, and the
children as erased type values, so a recursive type is a finite graph that a program walks as deep
as it needs.

<!-- end generated -->
