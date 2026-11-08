<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Once.value

```ant signature
func Once.value[T](f: func(): T): func(): T
```

Once.value returns a function that runs f on its first call and returns the cached value on every
call. If f panics, every call, waiting or later, panics with the same message.

<!-- end generated -->
