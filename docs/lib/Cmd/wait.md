<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.wait

```ant signature
func Cmd.wait(p: Cmd.Process, ctx: Context | null = null): Result[int, Error]
```

Cmd.wait waits until every stage of p has exited and returns its status. Every call gets the same
outcome. It is Err when ctx is done first, and p goes on; and when the context given to
[`Cmd.start`](start.md) killed p, with that context's reason.

<!-- end generated -->
