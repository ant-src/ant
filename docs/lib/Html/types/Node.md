<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html.Node

```ant signature
type Html.Node = Html.Element | Html.Text | Html.Raw
```

Html.Node is one node of an Html value, as [Html.nodes](../nodes.md) returns it.

```ant signature
type Html.Element = { name: string, attributes: Array[Html.Attribute], children: Html }
```

Html.Element is an element: its name as written, its attributes in order, and its children.

```ant signature
type Html.Text = { text: string }
```

Html.Text is a text node, holding its text decoded.

```ant signature
type Html.Raw = { markup: string }
```

Html.Raw is a node of trusted markup, made by [Html.unsafe](../unsafe.md).

<!-- end generated -->
