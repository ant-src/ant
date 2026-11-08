<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parseZoned

```ant signature
func Time.parseZoned(s: string, disambiguation: Time.Disambiguation = Time.Compatible{}): Result[Time.ZonedDateTime, Error]
```

Time.parseZoned parses a zoned date-time from s, such as
2026-10-22T16:03:05+02:00[Europe/Stockholm]. The text needs a bracketed zone. Without an offset, the
wall-clock time is placed in the zone by disambiguation; with Z, the moment is taken and shown in
the zone; an offset the zone does not have at that time is an Err, as is text it cannot parse.

<!-- end generated -->
