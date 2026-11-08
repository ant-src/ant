<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.run

```ant signature
func Cmd.run(c: Cmd | Cmd.Pipeline, ctx: Context | null = null): Result[int, Error]
```

Cmd.run runs c in the foreground, with the program's own standard input, output and error, so a
pager, an editor or ssh can use the terminal, and returns its status, nonzero included. While it
runs the ant command holds back Ctrl-C, which every stage gets and decides on, and passes it on when
the last stage dies of it. It is Err when a stage could not be started, and when ctx is done, which
kills every stage.

<!-- end generated -->
