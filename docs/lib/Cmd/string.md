<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.string

```ant signature
func Cmd.string(c: Cmd | Cmd.Pipeline): string
```

Cmd.string returns c as a line for a log: the program and its arguments, each quoted as a POSIX
shell needs it, and a pipeline's stages joined by " | ". env and dir are left out, so a secret in
env never reaches the log.

<!-- end generated -->
