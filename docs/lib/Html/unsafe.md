<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Html.unsafe

```ant signature
func Html.unsafe(s: string): Html
```

Html.unsafe returns s as markup, which String.from writes unescaped. It is the one way to turn a
string into markup: pass only markup the program trusts, such as a sanitizer's output.

<!-- end generated -->

## Basic examples

```ant
String.from(Html.unsafe("<b>x</b>")) // "<b>x</b>"
```
