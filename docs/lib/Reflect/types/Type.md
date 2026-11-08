<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Type

```ant signature
type Reflect.Type[T]
```

Reflect.Type is the type T as a value, written type(T). It is opaque: [`Reflect.erase`](../erase.md)
gives the value with its static type forgotten, and [`Reflect.valueOf`](../valueOf.md) boxes a T
under it. Two type values are == when they denote the same type, aliases expanded, and a type value
renders as the type.

<!-- end generated -->
