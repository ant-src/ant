<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.nullable

```ant signature
func Reflect.nullable[T](t: Reflect.Type[T]): Reflect.Type[T | null]
```

Reflect.nullable returns type(T | null) for a T held as a type value. Unions flatten, so a T that
already includes null gives the same type.

<!-- end generated -->
