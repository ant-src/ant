<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.previousTransition

```ant signature
func Time.previousTransition(t: Time.ZonedDateTime): Time.ZonedDateTime | null
```

Time.previousTransition returns the latest moment before t at which the zone of t changed its
offset, in that zone. It is null when there is none, as for an offset zone.

<!-- end generated -->
