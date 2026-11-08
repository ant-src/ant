<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.field

```ant signature
func Reflect.field(d: Reflect.Value, name: string): Reflect.Value
```

Reflect.field returns the field name of the record in d, boxed at the field's type. Through an
anonymous record type it reaches only the fields that type lists. It panics if d's type is not a
record type that lists the field.

<!-- end generated -->
