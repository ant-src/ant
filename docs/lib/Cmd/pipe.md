<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.pipe

```ant signature
func Cmd.pipe(cmds: Array[Cmd]): Cmd.Pipeline
```

Cmd.pipe returns the pipeline of cmds, in order, holding a copy of the array. It panics when cmds is
empty.

<!-- end generated -->
