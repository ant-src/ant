<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Unit

```ant signature
type Time.Unit = Time.Year | Time.Month | Time.Week | Time.Day | Time.Hour | Time.Minute | Time.Second | Time.Millisecond | Time.Microsecond | Time.Nanosecond
```

Time.Unit names a field of a duration, from years down to nanoseconds. Years, months, weeks and days
are calendar units that move the wall clock; hours and smaller are exact units that move the
instant.

```ant signature
type Time.Year = {}
```

Time.Year is the unit of a calendar year, which has no fixed length.

```ant signature
type Time.Month = {}
```

Time.Month is the unit of a calendar month, which has no fixed length.

```ant signature
type Time.Week = {}
```

Time.Week is the unit of a calendar week, a calendar unit like Time.Year and Time.Month.

```ant signature
type Time.Day = {}
```

Time.Day is the unit of a calendar day, 24 hours except in a zone on the day of a daylight-saving
change.

```ant signature
type Time.Hour = {}
```

Time.Hour is the unit of an hour, an exact unit.

```ant signature
type Time.Minute = {}
```

Time.Minute is the unit of a minute, an exact unit.

```ant signature
type Time.Second = {}
```

Time.Second is the unit of a second, an exact unit.

```ant signature
type Time.Millisecond = {}
```

Time.Millisecond is the unit of a millisecond, an exact unit.

```ant signature
type Time.Microsecond = {}
```

Time.Microsecond is the unit of a microsecond, an exact unit.

```ant signature
type Time.Nanosecond = {}
```

Time.Nanosecond is the unit of a nanosecond, the smallest unit.

<!-- end generated -->
