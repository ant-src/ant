<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Panic.catch

```ant signature
func Panic.catch[T](f: func(): T): Result[T, Panic]
```

Panic.catch calls f and returns Ok of its value, or an Err of the panic that left f once the
deferred blocks inside f have run. A caught panic is not reported. Os.exit and the end of the run
pass through. It belongs at the edge of a unit of work, not in place of a Result.

<!-- end generated -->
