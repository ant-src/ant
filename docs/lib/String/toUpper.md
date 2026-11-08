<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toUpper

```ant signature
func String.toUpper(s: string): string
```

String.toUpper returns s in upper case by Unicode's full case mappings, so the length can change.
The mappings that depend on a language, such as Turkish, are not applied.

<!-- end generated -->

## Basic examples

```ant
String.toUpper("straße") // "STRASSE"
```
