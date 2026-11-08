<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.items

```ant signature
func Reflect.items(d: Reflect.Value): Array[Reflect.Value]
```

Reflect.items returns an array's elements or a set's members, each boxed at the element type, in a
new array read when it is called. It panics if d's type is neither.

<!-- end generated -->
