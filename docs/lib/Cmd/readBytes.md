<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.readBytes

```ant signature
func Cmd.readBytes(c: Cmd | Cmd.Pipeline, input: string | Bytes = "", ctx: Context | null = null): Result[Bytes, Error]
```

Cmd.readBytes is [Cmd.read](read.md) that returns standard output as the command wrote it.

<!-- end generated -->
