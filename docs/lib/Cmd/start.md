<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd.start

```ant signature
func Cmd.start(
  c: Cmd | Cmd.Pipeline, stdin: File.Handle | null = null, stdout: File.Handle | null = Os.stdout(),
  stderr: File.Handle | null = Os.stderr(), ctx: Context | null = null,
): Result[Cmd.Process, Error]
```

Cmd.start starts c in the background, every stage in one process group of its own, and returns once
it is running. Its standard streams are the handles given, null being the null device; standard
input is the null device and the outputs are the program's unless given. A terminal as stdin is Err.
Every handle stays the program's, which closes its copy of a pipe end it gave away. When ctx is done
every stage is killed, as when the run is cancelled; a command still running when the script ends is
left running.

<!-- end generated -->
