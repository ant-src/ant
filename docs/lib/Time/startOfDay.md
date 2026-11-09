<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.startOfDay

```ant signature
func Time.startOfDay(t: Time.ZonedDateTime): Time.ZonedDateTime
```

Time.startOfDay returns the first instant of the date of t in its zone. That is midnight, except on
a day whose midnight a daylight-saving change skipped.

<!-- end generated -->

## Basic examples

```ant
let t = Result.unwrap(Time.parseZoned("2024-09-08T12:00[America/Santiago]"))
Time.startOfDay(t) // 2024-09-08T01:00:00-03:00[America/Santiago]
```
