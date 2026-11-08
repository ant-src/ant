<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeScalar

```ant signature
func Reflect.makeScalar(t: Reflect.ErasedType, v: Reflect.Scalar): Reflect.Value | null
```

Reflect.makeScalar boxes v at the scalar type t, the inverse of Reflect.scalar: a decoder that holds
only a slot's erased type boxes every leaf through it. It returns null where t is not null, void,
bool, a number, or string, or where v is not of t's kind, so the one call tests that a leaf fits the
slot. A number converts to any number type as the conversion functions do, except that a conversion
that would panic returns null: a fraction, a value out of range, NaN, or an infinity for an integer
type, and a finite value beyond the range of a float32 for float32.

<!-- end generated -->
