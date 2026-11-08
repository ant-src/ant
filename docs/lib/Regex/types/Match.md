<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.Match

```ant signature
type Regex.Match = { text: string, start: int, end: int, groups: Array[string | null], named: Map[string, string | null] }
```

Regex.Match is one match of a whole pattern. text is the matched text and start and end its code
point indices into the searched string; groups[0] is the whole match and groups[n] the nth group,
null when the group took no part; named has an entry for each named group, in the pattern's order,
null likewise.

<!-- end generated -->
