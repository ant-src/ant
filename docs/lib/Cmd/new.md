<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.new

```ant signature
func Cmd.new(program: string, args: Array[string] = [], dir: string | null = null, env: Map[string, string] | null = null): Cmd
```

Cmd.new returns the Cmd of program with args, run in dir and with env when they are given.

<!-- end generated -->
