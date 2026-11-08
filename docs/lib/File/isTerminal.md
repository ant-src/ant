<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.isTerminal

```ant signature
func File.isTerminal(f: File.Handle): bool
```

File.isTerminal reports whether f is a terminal, as \[\[ -t 1 \]\] asks of standard output. A
buffered handle answers for the handle under it.

<!-- end generated -->
