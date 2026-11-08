<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.entries

```ant signature
func Reflect.entries(d: Reflect.Value): Array[Entry[Reflect.Value, Reflect.Value]]
```

Reflect.entries returns a map's entries, keys and values boxed at the key and value types, in a new
array read when it is called. It panics if d's type is not a map type.

<!-- end generated -->
