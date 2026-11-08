<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# eprint

```ant signature
func eprint[T](value: T): void
```

eprint writes String.from(value) to standard error, with no newline. Neither stream is buffered, so
the two interleave in program order.

<!-- end generated -->
