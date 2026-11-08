<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Decl

```ant signature
type Reflect.Decl
```

Reflect.Decl is a named type's declaration without its type arguments:
Box[int](../../prelude/int.md) and Box[string](../../prelude/string.md) share one. Two are == when
they are the same declaration, and one renders as the declared name. It is the key for anything that
applies to every instantiation of a generic type.

<!-- end generated -->
