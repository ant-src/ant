<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.zoneId

```ant signature
func Time.zoneId(z: Time.Zone): string
```

Time.zoneId returns the identifier of z: its name as the database spells it, or its offset as
±HH:MM.

<!-- end generated -->

## Basic examples

```ant
Time.zoneId(Result.unwrap(Time.zone("asia/calcutta"))) // "Asia/Calcutta"
Time.zoneId(Result.unwrap(Time.zone("+0530")))         // "+05:30"
```
