<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel

```ant signature
type Channel[T]
```

Channel carries values of T from goroutine to goroutine, blocking the side that is ahead. Channels
compare by identity.

<!-- end generated -->
