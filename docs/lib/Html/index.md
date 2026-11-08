<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html

```ant signature
type Html
```

Html is the value of an HTML expression, <p>{name}</p>: an immutable list of nodes, each an element,
a text node, or raw markup from [Html.unsafe](unsafe.md). It compares by its nodes with ==, and
String.from renders it as markup, escaping text and attribute values.

<!-- end generated -->
