<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeTagged

```ant signature
func Reflect.makeTagged(t: Reflect.ErasedType, payload: Reflect.Value): Reflect.Value
```

Reflect.makeTagged returns a new value of the named primitive type t over payload. It panics if t is
not a named primitive type, or payload's type is not its base.

<!-- end generated -->
