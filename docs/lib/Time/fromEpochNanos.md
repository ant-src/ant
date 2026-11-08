<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.fromEpochNanos

```ant signature
func Time.fromEpochNanos(ns: int): Time.Instant
```

Time.fromEpochNanos returns the instant ns nanoseconds after the Unix epoch, or before it when ns is
negative.

<!-- end generated -->
