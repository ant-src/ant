<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeRecord

```ant signature
func Reflect.makeRecord(t: Reflect.ErasedType, fields: Map[string, Reflect.Value]): Reflect.Value
```

Reflect.makeRecord returns a new record of the named record type t with the given fields, each field
left out taking its default as a literal evaluates it. The record carries the tag, so is and == work
on it as on any literal. It panics if t is not a named record type, a key is not one of its fields,
a field's type is not assignable to the field's type, or a field with no default is left out.

<!-- end generated -->
