<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.read

```ant signature
func Cmd.read(c: Cmd | Cmd.Pipeline, input: string | Bytes = "", ctx: Context | null = null): Result[string, Error]
```

Cmd.read runs c with input as its standard input, waits for it to exit, and returns its standard
output as text without one final newline. A nonzero status is Err with the message "program: exit
status n", followed by ": " and standard error when it wrote any; program is that of the stage that
set the status, with "(stage i)" after it when another stage runs the same program. Err also means a
stage could not be started or ctx was done, which kills every stage.

<!-- end generated -->
