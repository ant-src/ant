<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parsePlainDateTime

```ant signature
func Time.parsePlainDateTime(s: string): Result[Time.PlainDateTime, Error]
```

Time.parsePlainDateTime parses a date and time from s, such as 2026-10-22T14:03:05, with T or a
space between them. It ignores an offset or zone that follows, though not Z with no bracketed zone;
text it cannot parse is an Err.

<!-- end generated -->
