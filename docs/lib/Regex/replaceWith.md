<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.replaceWith

```ant signature
func Regex.replaceWith(p: Regex.Pattern, s: string, f: func(Regex.Match): string): string
```

Regex.replaceWith returns s with every match of p replaced by what f returns for it. f is called
with each match in order, and its result is used as written, with no $ expansion.

<!-- end generated -->

## Basic examples

```ant
Regex.replaceWith(/[a-z]+/, "ab1cd", { String.toUpper(x.text) }) // "AB1CD"
```
