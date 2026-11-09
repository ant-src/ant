<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.toZoned

```ant signature
func Time.toZoned(t: Time.PlainDateTime, z: Time.Zone, disambiguation: Time.Disambiguation = Time.Compatible{}): Time.ZonedDateTime
```

Time.toZoned places the wall-clock time t in zone z. A time that a daylight-saving change skipped or
repeated is resolved by disambiguation, by default Time.Compatible, which takes the later instant in
a gap and the earlier in an overlap; [`Time.possibleInstants`](possibleInstants.md) shows the
choices.

<!-- end generated -->

## Basic examples

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let t = Result.unwrap(Time.parsePlainDateTime("2024-03-10T02:30"))
Time.toZoned(t, ny) // 2024-03-10T03:30:00-04:00[America/New_York]
```
