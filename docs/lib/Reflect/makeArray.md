<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeArray

```ant signature
func Reflect.makeArray(t: Reflect.ErasedType, items: Array[Reflect.Value]): Reflect.Value
```

Reflect.makeArray returns a new array of the array type t holding the values of items. It panics if
t is not an array type, or an item's type is not assignable to the element type.

<!-- end generated -->
