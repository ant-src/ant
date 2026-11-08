<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Panic.Frame

```ant signature
type Panic.Frame = { function: string, path: string, line: int, column: int }
```

Panic.Frame is one frame of a panic's trace: the function, the path of its module, and the line and
column, counted from 1, of the panic in the first frame and of a call in the others.

<!-- end generated -->
