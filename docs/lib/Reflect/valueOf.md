<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.valueOf

```ant signature
func Reflect.valueOf[T](t: Reflect.Type[T], value: T): Reflect.Value
```

Reflect.valueOf boxes value with its type t. When t is a union, the box's type is the member value
belongs to, found from its kind or tag. The type comes first so that it fixes T before the value is
checked against it: Reflect.valueOf(type(Array\[int\]), \[\]) checks.

<!-- end generated -->
