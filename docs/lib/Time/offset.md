<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.offset

```ant signature
func Time.offset(t: Time.ZonedDateTime): string
```

Time.offset returns the offset of t from UTC as ±HH:MM, such as +02:00, with seconds as well when a
historical offset has some. The string of t rounds its offset to the minute; this does not.

<!-- end generated -->
