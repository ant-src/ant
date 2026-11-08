<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.findAll

```ant signature
func Regex.findAll(p: Regex.Pattern, s: string): Array[Regex.Match]
```

Regex.findAll returns every match of p in s, left to right and without overlap. An empty match that
abuts the match before it is skipped.

<!-- end generated -->

## Basic examples

```ant
Regex.findAll(/\d+/, "a1b22") | Array.map({ x.text }) // ["1", "22"]
```
