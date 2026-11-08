<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parseInstant

```ant signature
func Time.parseInstant(s: string): Result[Time.Instant, Error]
```

Time.parseInstant parses an instant from s, such as 2026-10-22T14:03:05Z. The text needs a date, a
time, and Z or an offset; a bracketed zone is ignored. Text it cannot parse is an Err.

<!-- end generated -->

## Basic examples

```ant
Time.parseInstant("2026-10-22T16:03:05+02:00") // Ok{ value: 2026-10-22T14:03:05Z }
```
