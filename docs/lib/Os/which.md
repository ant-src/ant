<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.which

```ant signature
func Os.which(program: string): string | null
```

Os.which returns the path of the executable in PATH that a [`Cmd`](../Cmd/index.md) naming program
starts, or null when PATH holds none.

<!-- end generated -->
