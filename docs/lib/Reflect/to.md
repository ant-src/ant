<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.to

```ant signature
func Reflect.to[T](d: Reflect.Value, t: Reflect.Type[T]): T
```

Reflect.to returns the value in d as a T. It panics if d's type is not assignable to T, which is the
check an assignment would have made statically.

<!-- end generated -->
