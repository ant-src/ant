<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeSet

```ant signature
func Reflect.makeSet(t: Reflect.ErasedType, members: Array[Reflect.Value]): Reflect.Value
```

Reflect.makeSet returns a new set of the set type t holding the values of members, duplicates
collapsed. It panics if t is not a set type, or a member's type is not assignable to the member
type.

<!-- end generated -->
