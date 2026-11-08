<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.roundDuration

```ant signature
func Time.roundDuration(d: Time.Duration, smallestUnit: Time.Unit | null = null, largestUnit: Time.Unit | null = null, increment: int = 1, mode: Time.RoundingMode = Time.HalfExpand{}, relativeTo: Time.Dated | null = null): Time.Duration
```

Time.roundDuration balances d so that no field below largestUnit exceeds its next unit, and rounds
it to a multiple of increment of smallestUnit by mode, dropping the smaller fields. largestUnit
defaults to the larger of the largest nonzero unit of d and smallestUnit; years, months and weeks
need relativeTo, as in [`Time.total`](total.md). increment follows [`Time.round`](round.md) for a
time unit, and may be any for a day or calendar unit when largestUnit is smallestUnit. It panics if
neither unit is given, on years, months or weeks without relativeTo, and on an increment that breaks
those rules.

<!-- end generated -->

## Basic examples

```ant
Time.roundDuration(Time.duration(minutes: 90), largestUnit: Time.Hour{})            // PT1H30M
Time.roundDuration(Time.duration(hours: 1, minutes: 29), smallestUnit: Time.Hour{}) // PT1H
```
