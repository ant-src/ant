<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.glob

```ant signature
func File.glob(pattern: string, dir: string = ".", hidden: bool = false): Result[Array[string], Error]
```

File.glob returns the files and directories below dir whose paths relative to dir match pattern, as
paths relative to dir, cleaned, each once, in lexical order. The syntax is
[`Path.match`](../Path/match.md)'s: wildcards skip names that begin with a dot unless hidden is
true, \*\* does not descend through symbolic links, and a pattern that ends in a separator matches
directories only. An absolute pattern ignores dir. Nothing matching is Ok with an empty array; a
malformed pattern, and a directory the pattern leads into that cannot be read, is Err.

<!-- end generated -->
