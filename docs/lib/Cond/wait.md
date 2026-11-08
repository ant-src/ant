<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cond.wait

```ant signature
func Cond.wait(c: Cond): void
```

Cond.wait releases the mutex of c, waits until signalled, and reacquires the mutex before returning.
It panics if the mutex is not locked. It takes no context.

<!-- end generated -->
