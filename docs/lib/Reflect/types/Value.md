<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Value

```ant signature
type Reflect.Value
```

Reflect.Value is a value together with a type it belongs to, the one way to hold a value whose type
is known only at run time. [`Reflect.valueOf`](../valueOf.md) boxes, [`Reflect.to`](../to.md)
unboxes, the readers take one apart along its type, and the builders make new ones from checked
parts. In a checked program the value always belongs to the type, which is never a union or an
alias. Two are == when their types and values are, and one renders as Reflect.Value(42: int).

<!-- end generated -->
