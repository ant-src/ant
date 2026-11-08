<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.since

```ant signature
func Time.since[T: Time.Point](t: T, other: T, largestUnit: Time.Unit | null = null): Time.Duration
```

Time.since returns the duration from other to t, counted from t. It is Time.negate(Time.until(t,
other)), and for calendar units can differ from Time.until(other, t), whose anchor is other. It
panics as [`Time.until`](until.md) does.

<!-- end generated -->

## Basic examples

```ant
var feb29 = Time.plainDate(2024, 2, 29)
var mar31 = Time.plainDate(2024, 3, 31)
Time.since(mar31, feb29, largestUnit: Time.Month{}) // P1M
Time.until(feb29, mar31, largestUnit: Time.Month{}) // P1M2D
```
