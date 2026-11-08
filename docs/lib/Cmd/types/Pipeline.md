<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.Pipeline

```ant signature
type Cmd.Pipeline = { cmds: Array[Cmd] }
```

Cmd.Pipeline is commands run at once, each one's standard output the next one's standard input, made
by [Cmd.pipe](../pipe.md). Every function of Cmd that runs a Cmd runs a pipeline too: the first
stage reads its input, the last writes its output, and every stage writes standard error. Its status
is that of the rightmost stage that failed, or 0; a stage before the last that SIGPIPE ended does
not count, since only a later stage that stopped reading ends it so. A pipeline of no commands
panics wherever it is run.

<!-- end generated -->
