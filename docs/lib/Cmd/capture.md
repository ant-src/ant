<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.capture

```ant signature
func Cmd.capture(c: Cmd | Cmd.Pipeline, input: string | Bytes = "", ctx: Context | null = null): Result[Cmd.Output, Error]
```

Cmd.capture runs c with input as its standard input, waits for it to exit, and captures its status
and outputs. A nonzero status is Ok; Err means a stage could not be started or run to its end, and
when ctx is done every stage is killed and the result is Err with ctx's reason.

<!-- end generated -->
