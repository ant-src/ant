<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toLower

```ant signature
func String.toLower(s: string): string
```

String.toLower returns s in lower case by Unicode's full case mappings. The string is mapped whole,
so a final Greek sigma becomes ς, and the mappings that depend on a language, such as Turkish, are
not applied.

<!-- end generated -->

## Basic examples

```ant
String.toLower("ΣΟΦΟΣ") // "σοφος"
```
