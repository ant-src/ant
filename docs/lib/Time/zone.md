<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.zone

```ant signature
func Time.zone(id: string): Result[Time.Zone, Error]
```

Time.zone returns the zone named by id, a name of the IANA database in any letter case or a fixed
offset. An alias such as Asia/Calcutta is accepted, and an offset may be written ±HH:MM, ±HHMM or
±HH; anything else, the empty string included, is an Err.

<!-- end generated -->
