<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.ZonedDateTime

```ant signature
type Time.ZonedDateTime
```

Time.ZonedDateTime is an instant seen in a zone, the only type that knows both the moment and the
wall-clock time there. It renders with its offset and zone both, as
2024-11-03T01:00:00-04:00\[America/New\_York\], so the moment survives a round trip even where the
wall-clock time is ambiguous; two values are == when they are the same moment in the same zone.

<!-- end generated -->
