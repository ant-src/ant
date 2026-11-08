<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.normalize

```ant signature
func String.normalize(s: string, form: string): string
```

String.normalize returns s in the Unicode normalization form form, one of "NFC", "NFD", "NFKC", or
"NFKD". It panics on any other form.

<!-- end generated -->
