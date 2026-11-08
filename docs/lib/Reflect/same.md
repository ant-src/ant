<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.same

```ant signature
func Reflect.same(a: Reflect.Value, b: Reflect.Value): bool
```

Reflect.same reports whether a and b hold one record, array, map, or set, as Record.same,
Array.same, Map.same, and Set.same do, and whether they are == for any other kind. A walker tests
each child against the boxes on its path with it to find a cycle, which the readers never detect.

<!-- end generated -->
