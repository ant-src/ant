<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.ErasedType

```ant signature
type Reflect.ErasedType
```

Reflect.ErasedType is a type value whose static type is not known: what
[`Reflect.erase`](../erase.md) returns, and what [`Reflect.shape`](../shape.md) takes apart. Code
that walks a type recurses on erased types, since the element type of an Array\[E\] is not known
statically one level down.

<!-- end generated -->
