<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.hoursInDay

```ant signature
func Time.hoursInDay(t: Time.ZonedDateTime): float64
```

Time.hoursInDay returns the length in hours of the date of t in its zone: 24.0 on most days, 23.0 or
25.0 on a daylight-saving change, 23.5 on Lord Howe Island.

<!-- end generated -->
