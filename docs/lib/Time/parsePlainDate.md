<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parsePlainDate

```ant signature
func Time.parsePlainDate(s: string): Result[Time.PlainDate, Error]
```

Time.parsePlainDate parses a date from s, such as 2026-10-22. It takes the date of a longer string
and ignores the rest, though not of one with Z and no bracketed zone; text it cannot parse, or a
date that does not exist, is an Err.

<!-- end generated -->

## Basic examples

```ant
let s = "2026-10-22T16:03:05+02:00[Europe/Stockholm]"
Time.parsePlainDate(s) // Ok{ value: 2026-10-22 }
```
