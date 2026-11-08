<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.foldCase

```ant signature
func String.foldCase(s: string): string
```

String.foldCase returns s with Unicode's full case folding applied, a key in which case does not
matter. It does not normalize; see [`String.normalize`](normalize.md).

<!-- end generated -->

## Basic examples

```ant
String.foldCase("Straße") // "strasse"
```
