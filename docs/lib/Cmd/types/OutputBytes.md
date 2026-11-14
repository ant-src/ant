<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.OutputBytes

```ant signature
type Cmd.OutputBytes = { status: int, statuses: Array[int], stdout: Bytes, stderr: string }
```

Cmd.OutputBytes is what [`Cmd.captureBytes`](../captureBytes.md) collects: as [`Cmd.Output`](Output.md), with standard output as the last stage wrote it.

<!-- end generated -->
