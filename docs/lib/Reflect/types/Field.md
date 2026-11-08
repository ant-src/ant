<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Field

```ant signature
type Reflect.Field = { name: string, type: Reflect.ErasedType, hasDefault: bool }
```

Reflect.Field is one field of a record shape: its name, its type, and whether the declaration gives
it a default, so that a literal or [`Reflect.makeRecord`](../makeRecord.md) may leave it out.

<!-- end generated -->
