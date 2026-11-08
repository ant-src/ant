<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.kill

```ant signature
func Cmd.kill(p: Cmd.Process, signal: string = "SIGTERM"): Result[void, Error]
```

Cmd.kill sends the signal named, "SIGTERM", "SIGKILL", "SIGINT", "SIGHUP", "SIGQUIT", "SIGUSR1" or
"SIGUSR2", to p's process group, and does not wait. It is Ok when p has exited already, and Err for
a name the host does not have.

<!-- end generated -->
