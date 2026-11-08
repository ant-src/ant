<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.slice

```ant signature
func String.slice(s: string, from: int, to: int): string
```

String.slice returns the code points of s from index from up to but excluding index to. It panics if
from > to, or if either is outside 0 ..= String.length(s). A code point range can cut through a
grapheme; see [String.graphemes](graphemes.md).

<!-- end generated -->
