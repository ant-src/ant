<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.erase

```ant signature
func Reflect.erase[T](t: Reflect.Type[T]): Reflect.ErasedType
```

Reflect.erase returns t with its static type forgotten, for code that walks types. The conversion is
always written: a Reflect.Type\[T\] is not assignable to a Reflect.ErasedType.

<!-- end generated -->
