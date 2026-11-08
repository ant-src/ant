<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeMap

```ant signature
func Reflect.makeMap(t: Reflect.ErasedType, entries: Array[Entry[Reflect.Value, Reflect.Value]]): Reflect.Value
```

Reflect.makeMap returns a new map of the map type t holding entries, a later entry with an == key
replacing an earlier one. It panics if t is not a map type, or a key's or value's type is not
assignable to its slot.

<!-- end generated -->
