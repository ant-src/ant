<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Path.match

```ant signature
func Path.match(pattern: string, path: string, hidden: bool = false): Result[bool, Error]
```

Path.match reports whether the whole of path matches the glob pattern, element by element. \* is any
run of characters within an element, ? one character, \[a-z\] one of a set, \[^a-z\] or \[!a-z\] one
outside it, and \*\* as a whole element zero or more elements (one or more as the last). On Unix \\
escapes; on Windows it separates, and \[\*\] escapes. A wildcard does not match a dot that begins an
element unless hidden is true. It looks only at the text and is Err on a malformed pattern.

<!-- end generated -->

## Basic examples

```ant
Path.match("**/*.ant", "src/main.ant") // Ok{ value: true }
Path.match("*", ".env")                // Ok{ value: false }
```
