<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parseDuration

```ant signature
func Time.parseDuration(s: string): Result[Time.Duration, Error]
```

Time.parseDuration parses an ISO 8601 duration from s, such as P1Y2M3DT4H5M6.5S, -PT1H or PT0S. It
takes PnYnMnWnDTnHnMnS with an optional leading sign, where only the last time component may have a
fraction, and keeps the fields as written; text it cannot parse is an Err.

<!-- end generated -->
