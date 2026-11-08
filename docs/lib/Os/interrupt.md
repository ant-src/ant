<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.interrupt

```ant signature
func Os.interrupt(): Context
```

Os.interrupt returns a context that is done, with the reason "interrupt", when the process is
interrupted by Ctrl-C, SIGINT, or SIGTERM. It is the same context at every call. Once it has been
called, the first interrupt no longer ends the run; a second does, as does the first after the
program cancels the context itself.

<!-- end generated -->
