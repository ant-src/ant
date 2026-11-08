<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Once.do

```ant signature
func Once.do[T](once: Once, f: func(): T): void
```

Once.do runs f on the first call for once and discards its value; later calls wait for that run to
finish and do nothing. If f panics, once is complete and every call, waiting or later, panics with
the same message.

<!-- end generated -->
