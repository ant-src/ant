<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Duration

```ant signature
type Time.Duration
```

Time.Duration is an amount of time in ten fields, years down to nanoseconds, all of one sign and
kept as given. 90 minutes stays 90 minutes until [`Time.roundDuration`](../roundDuration.md)
balances it, and two durations are == only when all ten fields are, so PT90M and PT1H30M are not. It
renders in ISO 8601 form, as P1Y2M3DT4H5M6.5S, -PT1H or PT0S.

<!-- end generated -->
