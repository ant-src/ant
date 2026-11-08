<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.replace

```ant signature
func Regex.replace(p: Regex.Pattern, s: string, replacement: string): string
```

Regex.replace returns s with every match of p replaced by replacement. In replacement, $n and $\{n\}
expand to group n, $\{name\} to a named group, $$ to one $, and a group that took no part to "". A
name after $ is taken as long as it can be, so write $\{1\}x, not $1x.

<!-- end generated -->

## Basic examples

```ant
Regex.replace(/(\w+)@(\w+)/, "ann@home bob@work", "$2:$1") // "home:ann work:bob"
```
