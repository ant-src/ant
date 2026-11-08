<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.Output

```ant signature
type Cmd.Output = { status: int, statuses: Array[int], stdout: string, stderr: string, stdoutBytes: Bytes, stderrBytes: Bytes }
```

Cmd.Output is what [Cmd.capture](../capture.md) collects from a command that ran to its end. status
is its status, statuses each stage's own; each output is kept as bytes in stdoutBytes and
stderrBytes and as text in stdout and stderr. The stages' standard error interleaves as it arrived.

<!-- end generated -->
