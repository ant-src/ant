<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parsePlainTime

```ant signature
func Time.parsePlainTime(s: string): Result[Time.PlainTime, Error]
```

Time.parsePlainTime parses a time of day from s, such as 14:03:05.123. It takes HH:MM, HH:MM:SS or
up to nine digits of fraction, alone or after a date, and ignores the rest of a longer string,
though not of one with Z and no bracketed zone; text it cannot parse is an Err.

<!-- end generated -->
