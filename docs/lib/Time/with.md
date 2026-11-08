<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.with

```ant signature
func Time.with[T: Time.Wall](t: T, year: int | null = null, month: int | null = null, day: int | null = null, hour: int | null = null, minute: int | null = null, second: int | null = null, millisecond: int | null = null, microsecond: int | null = null, nanosecond: int | null = null, disambiguation: Time.Disambiguation = Time.Compatible{}): T
```

Time.with returns t with the named fields replaced and the others kept. A field out of range is
constrained, so month 13 is December and day 31 in April is April 30. A PlainDate takes only the
date fields and a PlainTime only the time fields, and panics on another; on a ZonedDateTime the new
wall-clock time keeps the offset of t when the zone has it then, and is placed by disambiguation
otherwise.

<!-- end generated -->

## Basic examples

```ant
Time.with(Time.plainDate(2026, 4, 10), day: 31) // 2026-04-30
```
