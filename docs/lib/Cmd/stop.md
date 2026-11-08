<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.stop

```ant signature
func Cmd.stop(p: Cmd.Process, grace: Time.Duration = Time.duration(seconds: 5)): Result[int, Error]
```

Cmd.stop ends p: SIGTERM, then SIGKILL when it is still running after grace, or SIGKILL alone when
grace is zero or negative, and returns its status as [`Cmd.wait`](wait.md) does. It panics when
grace has years, months or weeks.

<!-- end generated -->
