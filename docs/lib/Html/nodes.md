<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html.nodes

```ant signature
func Html.nodes(h: Html): Array[Html.Node]
```

Html.nodes returns the top-level nodes of h, in order. An element's children are read with
Html.nodes again.

<!-- end generated -->

## Basic examples

```ant
Html.nodes(<p>hi</p>) // [Html.Element{ name: "p", attributes: [], children: hi }]
```
