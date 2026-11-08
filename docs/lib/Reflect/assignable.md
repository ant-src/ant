<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.assignable

```ant signature
func Reflect.assignable(from: Reflect.ErasedType, to: Reflect.ErasedType): bool
```

Reflect.assignable reports whether a value of type from may be used where to is expected (section 4,
Assignability), on two closed types. It is the relation the readers, the builders, and
[`Reflect.to`](to.md) use, exposed so that a library can test before calling a function that would
panic.

<!-- end generated -->
