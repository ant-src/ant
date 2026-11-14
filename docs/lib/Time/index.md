<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time

Time holds dates, times, zones, and durations, after JavaScript's Temporal: exact time and
wall-clock time are different types, a zone is named and carried, and nothing assumes UTC or the
host's zone. Dates are in the ISO 8601 calendar, the proleptic Gregorian one, with a year 0 and
negative years. The time types are opaque, immutable, and compared by value. Unlike the other I/O
namespaces, Time is computation and panics on an argument that cannot be.

<!-- end generated -->

## Details

- An [Instant](types/Instant.md) is a moment with no zone. A [ZonedDateTime](types/ZonedDateTime.md)
  is a moment seen in a zone. [PlainDate](types/PlainDate.md), [PlainTime](types/PlainTime.md), and
  [PlainDateTime](types/PlainDateTime.md) are what a calendar or a clock shows, with no zone.
- Constructors panic on a value that cannot exist, such as February 30. Text from outside goes
  through the `parse` functions, which return a `Result`.
- `<` and the other comparison operators do not apply to time values. Order them with
  [Time.compare](compare.md).
- Years, months, weeks, and days move the wall clock; hours and smaller units move the instant. An
  `Instant` has no calendar, so adding a day to it panics.
- A [Duration](types/Duration.md) keeps its fields as given: `90m` and `1h30m` are not `==`.
- A constant span of hours and smaller units is a
  [duration literal](../../language/literals/durations.md) such as `500ms`, `2h30m`, or `1.5s`, of
  type `Time.Duration`, with no import and no arithmetic operators. [Time.duration](duration.md)
  builds computed and calendar spans.
- Every value renders as its RFC 9557 string in `String.from`, interpolation, and `Json.stringify`
  alike. The `parse` functions read it back, and `Time.zone` reads back a zone.

## Basic examples

Build a date and a time, read their fields, and do calendar arithmetic:

```ant
let day = Time.plainDate(2026, 10, 22)
day                                     // 2026-10-22
Time.dayOfWeek(day)                     // 4
Time.add(day, Time.duration(months: 1)) // 2026-11-22
Time.plainTime(14, 30)                  // 14:30:00
```

Write a constant span as a duration literal and move a clock time by it:

```ant
let d = 1h30m
d                        // PT1H30M
Time.plainTime(9).add(d) // 10:30:00
```

Place a wall-clock time in a zone to get a moment, and see that moment elsewhere:

```ant
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let meeting = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(14, 30))
let zoned = Time.toZoned(meeting, stockholm)
zoned                                                        // 2026-10-22T14:30:00+02:00[Europe/Stockholm]
zoned.toInstant()                                            // 2026-10-22T12:30:00Z
zoned.withZone(Result.unwrap(Time.zone("America/New_York"))) // 2026-10-22T08:30:00-04:00[America/New_York]
```

Measure the span between two dates:

```ant
let start = Time.plainDate(2026, 1, 1)
let end = Time.plainDate(2026, 10, 22)
Time.until(start, end)                            // P294D
Time.until(start, end, largestUnit: Time.Month{}) // P9M21D
```

## Scope

Every value parses back from the string it renders as. Text that is not a time, or a date that does
not exist, is an `Err`:

```ant
Time.parseInstant("2026-10-22T14:03:05Z") // Ok{ value: 2026-10-22T14:03:05Z }
Time.parseDuration("PT1H30M")             // Ok{ value: PT1H30M }
Time.parseDuration("2h30m")               // Ok{ value: PT2H30M }
Time.parsePlainDate("2026-02-30")         // Err{ error: Error{ message: "no such date: 2026-02-30" } }
```

A zone is an IANA name, in any letter case and including aliases, or a fixed offset. `Time.utc` is
the named zone UTC, and a named zone never equals an offset zone:

```ant
Time.zone("asia/calcutta") == Time.zone("Asia/Kolkata") // true
Time.zoneId(Result.unwrap(Time.zone("+0530")))          // "+05:30"
Time.zone("Mars/Olympus")                               // Err{ error: Error{ message: "unknown time zone \"Mars/Olympus\"" } }
Time.fromEpochMillis(0).inZone(Time.utc)                // 1970-01-01T00:00:00+00:00[UTC]
Time.utc == Result.unwrap(Time.zone("+00:00"))          // false
```

`Time.with` replaces fields and constrains the result to a real date:

```ant
let day = Time.plainDate(2026, 10, 22)
day.with(month: 13, day: 31)          // 2026-12-31
Time.plainTime(9, 15).with(minute: 0) // 09:00:00
```

A value typed by one of the unions, such as [Dated](types/Dated.md), narrows with `is`:

```ant
let t: Time.Dated = Time.plainDate(2026, 10, 22)
t.year()            // 2026
t is Time.PlainDate // true
```

Calls chain through a pipe:

```ant
Time.plainDate(2026, 10, 22) | Time.add(Time.duration(days: 10)) | Time.month() // 11
```

## Functions by job

### Types

- [Instant](types/Instant.md): a moment on the time line, with no zone or calendar
- [ZonedDateTime](types/ZonedDateTime.md): an instant seen in a zone
- [PlainDate](types/PlainDate.md), [PlainTime](types/PlainTime.md),
  [PlainDateTime](types/PlainDateTime.md): a calendar date, a clock time, and both, with no zone
- [Zone](types/Zone.md): an IANA zone or a fixed offset
- [Duration](types/Duration.md): a span in ten fields, kept as given
- [Point](types/Point.md), [Wall](types/Wall.md), [Dated](types/Dated.md), [Timed](types/Timed.md):
  the unions of time types that functions take
- [Unit](types/Unit.md), [Disambiguation](types/Disambiguation.md),
  [RoundingMode](types/RoundingMode.md): the options of the functions

### Zones

- [Time.zone](zone.md): a zone from its name or offset, or an `Err`
- [Time.localZone](localZone.md): the host's zone
- [Time.zoneId](zoneId.md): the name or offset of a zone
- `Time.utc`: the named zone UTC

### Now, waiting, and the epoch

- [Time.now](now.md): the current instant
- [Time.nowIn](nowIn.md): the current time in a zone
- [Time.monotonic](monotonic.md): milliseconds from an arbitrary origin, for measuring
- [Time.sleep](sleep.md): block the goroutine for a duration
- [Time.ticker](ticker.md): a channel that receives the time every interval
- [Time.fromEpochMillis](fromEpochMillis.md), [Time.fromEpochNanos](fromEpochNanos.md): an instant
  from a count since the Unix epoch
- [Time.epochMillis](epochMillis.md), [Time.epochNanos](epochNanos.md): the count since the epoch

### Construction and text

- [Time.plainDate](plainDate.md), [Time.plainTime](plainTime.md),
  [Time.plainDateTime](plainDateTime.md): build a plain value; panics on one that cannot exist
- [Time.duration](duration.md): build a duration from named fields; a constant span is a literal
  such as `2h30m`
- [Time.parseInstant](parseInstant.md), [Time.parseZoned](parseZoned.md),
  [Time.parsePlainDate](parsePlainDate.md), [Time.parsePlainTime](parsePlainTime.md),
  [Time.parsePlainDateTime](parsePlainDateTime.md), [Time.parseDuration](parseDuration.md): read a
  value from text, or an `Err`

### Conversion

- [Time.inZone](inZone.md): an instant seen in a zone
- [Time.withZone](withZone.md): the same instant in another zone
- [Time.toZoned](toZoned.md): a wall-clock time placed in a zone
- [Time.possibleInstants](possibleInstants.md): every instant a wall-clock time has in a zone
- [Time.toInstant](toInstant.md), [Time.toPlainDate](toPlainDate.md),
  [Time.toPlainTime](toPlainTime.md), [Time.toPlainDateTime](toPlainDateTime.md): one part of a
  value
- [Time.startOfDay](startOfDay.md): the first instant of the day in its zone
- [Time.zoneOf](zoneOf.md), [Time.offset](offset.md), [Time.offsetSeconds](offsetSeconds.md): the
  zone and its offset at that moment
- [Time.hoursInDay](hoursInDay.md): the length of the day in its zone
- [Time.nextTransition](nextTransition.md), [Time.previousTransition](previousTransition.md): the
  nearest change of offset

### Fields

- [Time.year](year.md), [Time.month](month.md), [Time.day](day.md): the date
- [Time.dayOfWeek](dayOfWeek.md), [Time.dayOfYear](dayOfYear.md): the day's place in the week and
  the year
- [Time.weekOfYear](weekOfYear.md), [Time.yearOfWeek](yearOfWeek.md): the ISO week and its year
- [Time.daysInMonth](daysInMonth.md), [Time.daysInYear](daysInYear.md),
  [Time.inLeapYear](inLeapYear.md): the size of the month and the year
- [Time.hour](hour.md), [Time.minute](minute.md), [Time.second](second.md),
  [Time.millisecond](millisecond.md), [Time.microsecond](microsecond.md),
  [Time.nanosecond](nanosecond.md): the time
- [Time.with](with.md): a copy with some fields replaced

### Arithmetic

- [Time.add](add.md), [Time.subtract](subtract.md): move a value by a duration
- [Time.until](until.md), [Time.since](since.md): the duration between two values
- [Time.round](round.md): a value rounded to a unit
- [Time.compare](compare.md): the order of two values

### Durations

- [Time.part](part.md): one field of a duration
- [Time.sign](sign.md), [Time.negate](negate.md), [Time.abs](abs.md): the sign of a duration
- [Time.addDuration](addDuration.md): the sum of two durations
- [Time.compareDuration](compareDuration.md): the order of two durations by length
- [Time.total](total.md): a duration as a number of one unit
- [Time.roundDuration](roundDuration.md): a duration balanced and rounded

## Applications

`Time.now()` is the current Instant, and `Time.nowIn(zone)` is the wall-clock time in that zone:

```ant
let now = Time.nowIn(Time.localZone())
println(now.zoneOf() == Time.localZone())
```

```text
true
```

Compute an age in whole years:

```ant
func age(birthday: Time.PlainDate, today: Time.PlainDate): int {
  Time.until(birthday, today, largestUnit: Time.Year{}).part(Time.Year{})
}
age(Time.plainDate(1990, 10, 23), Time.plainDate(2026, 10, 22)) // 35
```

Schedule a weekly 9:00 meeting in New York. Weeks move the wall clock, so the meeting stays at 9:00
when daylight saving ends and the instant moves an hour:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let first = Time.plainDateTime(Time.plainDate(2026, 10, 26), Time.plainTime(9)).toZoned(ny)
for week in [0, 1, 2] {
  let at = first.add(Time.duration(weeks: week))
  println("${at.toPlainDateTime()} ${at.toInstant()}")
}
```

```text
2026-10-26T09:00:00 2026-10-26T13:00:00Z
2026-11-02T09:00:00 2026-11-02T14:00:00Z
2026-11-09T09:00:00 2026-11-09T14:00:00Z
```

Sort timestamps read from text. `Time.compare` is the ordering:

```ant
let events = ["2026-10-22T14:03:05Z", "2026-03-01T08:00:00Z", "2026-07-15T12:00:00Z"]
  | Array.map({ Result.unwrap(Time.parseInstant(x)) })
  | Array.sortWith(Time.compare)
println(events)
```

```text
[2026-03-01T08:00:00Z, 2026-07-15T12:00:00Z, 2026-10-22T14:03:05Z]
```

Measure how long a piece of work takes with [Time.monotonic](monotonic.md):

```ant
let start = Time.monotonic()
Time.sleep(20ms)
let elapsed = Time.monotonic() - start
println(elapsed >= 20)
```

```text
true
```

Act on a beat with a ticker that stops with its context:

```ant
let ctx = Context.new(timeout: 1s)
defer { ctx.cancel() }
let ticks = Time.ticker(10ms, ctx: ctx)
for i in [1, 2, 3] {
  ticks.receive()
  println("tick ${i}")
}
```

```text
tick 1
tick 2
tick 3
```

## Properties and relations

A value renders as text that parses back to an equal value:

```ant
let t = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(14, 3, 5, 500))
assert(Result.unwrap(Time.parsePlainDateTime("${t}")) == t)
```

Two zoned values at one moment in different zones compare as `0` but are not `==`:

```ant
let moment = Time.fromEpochMillis(0)
let utc = moment.inZone(Time.utc)
let kolkata = moment.inZone(Result.unwrap(Time.zone("Asia/Kolkata")))
assert(Time.compare(utc, kolkata) == 0)
assert(utc != kolkata)
assert(utc.toInstant() == kolkata.toInstant())
```

Durations compare by fields with `==` and by length with `Time.compareDuration`:

```ant
let a = 90m
let b = 1h30m
assert(a != b)
assert(Time.compareDuration(a, b) == 0)
assert(Time.roundDuration(a, largestUnit: Time.Hour{}) == b)
```

## Possible issues

A day added in a zone is not always 24 hours. On the day daylight saving starts, a day is 23 hours:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let sat = Time.plainDateTime(Time.plainDate(2026, 3, 7), Time.plainTime(12)).toZoned(ny)
sat.add(Time.duration(days: 1)) // 2026-03-08T12:00:00-04:00[America/New_York]
sat.add(24h)                    // 2026-03-08T13:00:00-04:00[America/New_York]
```

A constructor panics on a date that does not exist:

```ant
Time.plainDate(2026, 2, 30) // panic: Time.plainDate: no such date: 2026-02-30
```

An `Instant` has no calendar, so calendar units panic on it. Convert it to a `ZonedDateTime` first:

```ant
Time.fromEpochMillis(0).add(Time.duration(days: 1))                  // panic: Time.add: an instant has no calendar, so it takes no days; convert it with Time.inZone first
Time.fromEpochMillis(0).inZone(Time.utc).add(Time.duration(days: 1)) // 1970-01-02T00:00:00+00:00[UTC]
```

`<` does not compare time values:

```ant
let a = Time.plainDate(2026, 1, 1)
let b = Time.plainDate(2026, 2, 1)
let earlier = a < b
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.PlainDate
```

```ant
let a = Time.plainDate(2026, 1, 1)
let b = Time.plainDate(2026, 2, 1)
Time.compare(a, b) < 0 // true
```

A value typed by a union passes to `Time.year`, but not to `Time.add`, `Time.until`, `Time.compare`,
or `Time.with`, which take a single member of the union:

```ant
let t: Time.Dated = Time.plainDate(2026, 10, 22)
let next = Time.add(t, Time.duration(days: 1))
// error: T would be Time.Dated, which is not a single member of its bound Time.Point; narrow first
```

Narrow it with `is` first:

```ant
let t: Time.Dated = Time.plainDate(2026, 10, 22)
if t is Time.PlainDate { Time.add(t, Time.duration(days: 1)) } // 2026-10-23
```

Comparing two duration literals of the same length with `==` is an error, since `==` compares the
fields as written:

```ant
let same = 90m == 1h30m
// error: 90m == 1h30m compares the fields as written, and == is false: compare the lengths with Time.compareDuration(90m, 1h30m)
```

```ant
Time.compareDuration(90m, 1h30m) // 0
```

## See also

- [Context](../Context/index.md): deadlines and cancellation, measured in a `Time.Duration`
- [Channel](../Channel/index.md): receive from a [Time.ticker](ticker.md), or wait in a `select`
- [Json](../Json/index.md): time values as JSON strings
