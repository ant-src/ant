<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.fromEpochMillis

```ant signature
func Time.fromEpochMillis(ms: int): Time.Instant
```

Time.fromEpochMillis returns the instant ms milliseconds after the Unix epoch, or before it when ms
is negative. It panics if the result is beyond the range of an instant, 100 million days each way.

<!-- end generated -->
