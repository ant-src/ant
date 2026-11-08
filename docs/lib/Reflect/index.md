<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect

Reflect holds reflection: types as values, written type(T) (section 4, Types as values), taken apart
one level at a time with Reflect.shape, and values boxed with a type they belong to, Reflect.Value,
read along that type and built from checked parts. With these, typed decoding and encoding for a
data format are ordinary library code, and the format keeps every policy.

<!-- end generated -->
