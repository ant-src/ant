<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.compare

```ant signature
func Time.compare[T: Time.Point](a: T, b: T): int
```

Time.compare returns -1, 0 or 1 as a is before, at or after b. Instants and ZonedDateTimes compare
by instant, so two zoned values at one moment in different zones compare as 0 though they are not
==; the plain types compare by fields. The comparison operators do not apply to time values.

<!-- end generated -->

## Basic examples

Compare two dates:

```ant
let spring = Time.plainDate(2026, 3, 20)
let summer = Time.plainDate(2026, 6, 21)
Time.compare(spring, summer) // -1
Time.compare(summer, spring) // 1
Time.compare(spring, spring) // 0
```

Test whether one moment comes before another:

```ant
let sent = Time.fromEpochMillis(1_700_000_000_000)
let read = Time.fromEpochMillis(1_700_000_060_000)
Time.compare(sent, read) < 0 // true
```

## Scope

Call it through `.` or in a pipe:

```ant
let a = Time.plainDate(2026, 1, 15)
let b = Time.plainDate(2025, 12, 24)
a.compare(b)        // 1
a | Time.compare(b) // 1
```

Two `ZonedDateTime`s compare by instant. The same moment in Stockholm and in New York compares as
`0`, although the two values are not `==`:

```ant
let moment = Time.fromEpochMillis(1_776_866_585_000)
let stockholm = moment.inZone(Time.zone("Europe/Stockholm").unwrap())
let newYork = moment.inZone(Time.zone("America/New_York").unwrap())
stockholm                        // 2026-04-22T16:03:05+02:00[Europe/Stockholm]
newYork                          // 2026-04-22T10:03:05-04:00[America/New_York]
Time.compare(stockholm, newYork) // 0
stockholm == newYork             // false
```

Use `==` when the zone matters too.

The same wall-clock time in two zones is two moments. Nine in the morning comes to Tokyo before it
comes to London:

```ant
let nine = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(9))
let tokyo = nine.toZoned(Time.zone("Asia/Tokyo").unwrap())
let london = nine.toZoned(Time.zone("Europe/London").unwrap())
Time.compare(tokyo, london) // -1
```

A `PlainDateTime` compares by date first, then by time:

```ant
let morning = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(9))
let lateEve = Time.plainDateTime(Time.plainDate(2026, 5, 3), Time.plainTime(23, 30))
Time.compare(lateEve, morning) // -1
```

A `PlainTime` compares by the clock alone, down to the nanosecond:

```ant
Time.compare(Time.plainTime(23, 59), Time.plainTime(0, 1))          // 1
Time.compare(Time.plainTime(12, nanosecond: 1), Time.plainTime(12)) // 1
```

A value of a union such as `Time.Point` passes once `is` narrows it to one type:

```ant
func sameDay(a: Time.Point, b: Time.PlainDate): bool {
  a is Time.PlainDate && Time.compare(a, b) == 0
}

sameDay(Time.plainDate(2026, 5, 4), Time.plainDate(2026, 5, 4)) // true
sameDay(Time.plainTime(9), Time.plainDate(2026, 5, 4))          // false
```

## Applications

`Time.compare` has the shape of a comparator, so it sorts an array of time values directly:

```ant
let dates = [Time.plainDate(2026, 3, 1), Time.plainDate(2025, 12, 24), Time.plainDate(2026, 1, 15)]
dates.sortWith(Time.compare) // [2025-12-24, 2026-01-15, 2026-03-01]
```

Sort records by a time field, newest first, by swapping the arguments:

```ant
type Post = { title: string, published: Time.PlainDate }
let posts = [
  Post{ title: "Hello", published: Time.plainDate(2025, 11, 2) },
  Post{ title: "Release", published: Time.plainDate(2026, 4, 9) },
  Post{ title: "Roadmap", published: Time.plainDate(2026, 1, 20) },
]
let newestFirst = func(a: Post, b: Post): int { Time.compare(b.published, a.published) }
for post in posts.sortWith(newestFirst) {
  println("${post.published} ${post.title}")
}
```

```text
2026-04-09 Release
2026-01-20 Roadmap
2025-11-02 Hello
```

Check that a date falls in a range, both ends included:

```ant
func within(d: Time.PlainDate, start: Time.PlainDate, end: Time.PlainDate): bool {
  Time.compare(d, start) >= 0 && Time.compare(d, end) <= 0
}

let start = Time.plainDate(2026, 6, 1)
let end = Time.plainDate(2026, 8, 31)
within(Time.plainDate(2026, 7, 14), start, end) // true
within(Time.plainDate(2026, 9, 1), start, end)  // false
```

Find the latest of several moments with `Array.reduce`:

```ant
let logins = [
  Time.fromEpochMillis(1_776_000_000_000),
  Time.fromEpochMillis(1_776_900_000_000),
  Time.fromEpochMillis(1_776_500_000_000),
]
logins.reduce(logins[0], func(latest: Time.Instant, t: Time.Instant) {
  if Time.compare(t, latest) > 0 { t } else { latest }
}) // 2026-04-22T23:20:00Z
```

Tell whether a deadline has passed. The program fixes `now` so that it prints the same every run; a
real one calls `Time.now()`:

```ant
let deadline = Time.parseZoned("2026-05-01T17:00:00+02:00[Europe/Stockholm]").unwrap()
let now = Time.parseInstant("2026-05-01T15:30:00Z").unwrap()
let late = Time.compare(now, deadline.toInstant()) > 0
println(if late { "late" } else { "on time" })
```

```text
late
```

## Properties and relations

Swapping the arguments negates the result:

```ant
let a = Time.plainDate(2026, 3, 1)
let b = Time.plainDate(2026, 4, 1)
assert(Time.compare(a, b) == -Time.compare(b, a))
```

For the plain types and `Instant`, `0` means `==`:

```ant
let a = Time.plainTime(8, 30)
let b = Time.plainTime(8, 30, 0)
assert(Time.compare(a, b) == 0)
assert(a == b)
```

Two `ZonedDateTime`s compare as their instants do:

```ant
let a = Time.parseZoned("2026-05-04T09:00:00[Asia/Tokyo]").unwrap()
let b = Time.parseZoned("2026-05-04T09:00:00[Europe/London]").unwrap()
assert(Time.compare(a, b) == Time.compare(a.toInstant(), b.toInstant()))
```

The sign of `Time.until` from `a` to `b` is the opposite of the comparison:

```ant
let a = Time.plainDate(2026, 3, 1)
let b = Time.plainDate(2026, 4, 1)
assert(Time.until(a, b).sign() == -Time.compare(a, b))
assert(Time.until(b, a).sign() == -Time.compare(b, a))
```

## Possible issues

`<` and the other comparison operators are a compile error on time values:

```ant
let a = Time.plainDate(2026, 3, 1)
let b = Time.plainDate(2026, 4, 1)
println(a < b)
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.PlainDate; narrow first
```

Write `Time.compare(a, b) < 0` instead.

For the same reason, `Array.sort`, `Array.sortBy`, `Array.min`, `Array.max`, `Array.minBy` and
`Array.maxBy` do not take time values or time keys:

```ant
let dates = [Time.plainDate(2026, 3, 1), Time.plainDate(2025, 12, 24)]
println(dates.sort())
// error: T would be Time.PlainDate, which is not a member of its bound Ordered
```

Use `Array.sortWith(Time.compare)` to sort; for the earliest or latest, fold with `Time.compare` as
the `Array.reduce` example in Applications does.

Both arguments must be one type. Comparing an `Instant` with a `ZonedDateTime` is a compile error:

```ant
let zoned = Time.parseZoned("2026-05-04T09:00:00[Asia/Tokyo]").unwrap()
let instant = Time.fromEpochMillis(1_777_852_800_000)
println(Time.compare(instant, zoned))
// error: cannot use Time.ZonedDateTime where Time.Instant is expected
```

Convert one of them first. Nine in the morning in Tokyo is midnight UTC:

```ant
let zoned = Time.parseZoned("2026-05-04T09:00:00[Asia/Tokyo]").unwrap()
let instant = Time.fromEpochMillis(1_777_852_800_000)
Time.compare(instant, zoned.toInstant()) // 0
```

A `PlainTime` does not wrap around midnight. A night shift from `22:00` to `02:00` ends at a time
that compares as earlier than its start:

```ant
Time.compare(Time.plainTime(22), Time.plainTime(2)) // 1
```

Give the times their dates with `Time.plainDateTime` to order them across midnight:

```ant
let start = Time.plainDateTime(Time.plainDate(2026, 5, 4), Time.plainTime(22))
let end = Time.plainDateTime(Time.plainDate(2026, 5, 5), Time.plainTime(2))
Time.compare(start, end) // -1
```

## See also

[Time.compareDuration](compareDuration.md) · [Time.until](until.md) · [Time.since](since.md) ·
[Time.toInstant](toInstant.md) · [Array.sortWith](../Array/sortWith.md) ·
[Time.Point](types/Point.md)
