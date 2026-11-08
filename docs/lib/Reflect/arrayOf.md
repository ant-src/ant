<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.arrayOf

```ant signature
func Reflect.arrayOf[T](t: Reflect.Type[T]): Reflect.Type[Array[T]]
```

Reflect.arrayOf returns type(Array\[T\]) for a T held as a type value, which a generic function
cannot write itself, since type(...) may not mention a type parameter.

<!-- end generated -->
