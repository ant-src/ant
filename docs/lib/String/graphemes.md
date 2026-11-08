<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.graphemes

```ant signature
func String.graphemes(s: string): Array[string]
```

String.graphemes returns the extended grapheme clusters of s (Unicode UAX \#29), the units a reader
sees as characters.

<!-- end generated -->

## Basic examples

```ant
String.graphemes("🇨🇭!") // ["🇨🇭", "!"]
```
