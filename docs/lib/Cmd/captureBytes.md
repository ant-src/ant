<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.captureBytes

```ant signature
func Cmd.captureBytes(c: Cmd | Cmd.Pipeline, input: string | Bytes = "", ctx: Context | null = null): Result[Cmd.OutputBytes, Error]
```

Cmd.captureBytes is [`Cmd.capture`](capture.md) that keeps standard output as the command wrote it, for a command whose output is binary; standard error stays text.

<!-- end generated -->
