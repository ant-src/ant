<!-- status: reviewed -->

# Dates and times

The [Time](../Time/index.md) namespace holds dates, clock times, zones, and spans of time. It keeps
apart what other libraries mix: a moment on the time line is one type, and what a calendar or a
clock shows is another. A zone is never assumed. Every value is immutable, compares by value with
`==`, and renders as its ISO 8601 text, with the zone in brackets on a zoned time and a zone as its
name. You need no import.

## Choose a type

Pick the type by the question the value answers:

| You hold                                         | Type                                                 |
| ------------------------------------------------ | ---------------------------------------------------- |
| when something happened, such as a log timestamp | [Time.Instant](../Time/types/Instant.md)             |
| what time it was in a place                      | [Time.ZonedDateTime](../Time/types/ZonedDateTime.md) |
| a day on the calendar, such as a birthday        | [Time.PlainDate](../Time/types/PlainDate.md)         |
| a time on the clock, such as an alarm            | [Time.PlainTime](../Time/types/PlainTime.md)         |
| a date and a time before its zone is known       | [Time.PlainDateTime](../Time/types/PlainDateTime.md) |
| a time zone                                      | [Time.Zone](../Time/types/Zone.md)                   |
| a span, such as 2 hours or 1 month               | [Time.Duration](../Time/types/Duration.md)           |

Each renders as the text you would write for it:

```ant
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let moment = Time.fromEpochMillis(1792677785000)
moment                                                                  // 2026-10-22T14:03:05Z
moment.inZone(stockholm)                                                // 2026-10-22T16:03:05+02:00[Europe/Stockholm]
Time.plainDate(2026, 10, 22)                                            // 2026-10-22
Time.plainTime(7, 30)                                                   // 07:30:00
Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(7, 30)) // 2026-10-22T07:30:00
2h30m                                                                   // PT2H30M
```

Functions that serve several types take one of four unions: [Point](../Time/types/Point.md) for all
five time types, [Wall](../Time/types/Wall.md) for the ones with fields,
[Dated](../Time/types/Dated.md) for the ones with a date, and [Timed](../Time/types/Timed.md) for
the ones with a clock time.

## Get the current time

```ant
let zone = Time.localZone()
let now = Time.now().inZone(zone)
println(now.zoneOf() == zone)
println(now.toInstant().inZone(zone) == now)
```

```text
true
true
```

- [Time.now](../Time/now.md): the current instant
- [Time.nowIn](../Time/nowIn.md): the current time in a zone
- [Time.localZone](../Time/localZone.md): the zone the host is set to

## Build values

A constructor panics on a value that cannot exist. A constant span of hours and smaller units is a
[duration literal](../../language/literals/durations.md); a span with days, weeks, months, or years,
or one built from a number, comes from `Time.duration`:

```ant
let minutes = 300
Time.plainDate(2026, 2, 28)       // 2026-02-28
Time.plainTime(9)                 // 09:00:00
Time.duration(months: 1, days: 3) // P1M3D
Time.duration(minutes: minutes)   // PT300M
Time.plainDate(2026, 2, 29)       // panic: Time.plainDate: no such date: 2026-02-29
```

- [Time.plainDate](../Time/plainDate.md): a date from a year, a month, and a day
- [Time.plainTime](../Time/plainTime.md): a clock time from an hour and smaller fields
- [Time.plainDateTime](../Time/plainDateTime.md): a date and a time together
- [Time.duration](../Time/duration.md): a span from named fields
- [Time.fromEpochMillis](../Time/fromEpochMillis.md): an instant from milliseconds since 1970
- [Time.fromEpochNanos](../Time/fromEpochNanos.md): an instant from nanoseconds since 1970

## Read and write text

A time value becomes text through interpolation, `String.from`, and `Json.stringify` alike. Text
from outside goes through a `parse` function, which returns a `Result`. Time has no layout strings;
build another form from the fields, such as `"${t.day()}.${t.month()}.${t.year()}"`, as in
[get the current time and print it](../../howto/time/now-format.md):

```ant
let due = Time.plainDate(2026, 10, 22)
"due ${due}"                      // "due 2026-10-22"
Json.stringify(Map{ "due": due }) // "{\"due\":\"2026-10-22\"}"
Time.parsePlainDate("2026-10-22") // Ok{ value: 2026-10-22 }
Time.parsePlainDate("22/10/2026") // Err{ error: Error{ message: "not a date or time: \"22/10/2026\"" } }
Time.parseDuration("1h30m")       // Ok{ value: PT1H30M }
```

- [Time.parseInstant](../Time/parseInstant.md): an instant, such as `2026-10-22T14:03:05Z`
- [Time.parseZoned](../Time/parseZoned.md): a zoned time, such as
  `2026-10-22T16:03:05+02:00[Europe/Stockholm]`
- [Time.parsePlainDate](../Time/parsePlainDate.md): a date, such as `2026-10-22`
- [Time.parsePlainTime](../Time/parsePlainTime.md): a clock time, such as `14:03:05`
- [Time.parsePlainDateTime](../Time/parsePlainDateTime.md): a date and time, such as
  `2026-10-22T14:03:05`
- [Time.parseDuration](../Time/parseDuration.md): a span, in ISO 8601 such as `PT1H30M` or as Go
  writes it, such as `1h30m`

## Work with zones

Place a wall-clock time in a zone to get a moment, then see the same moment elsewhere:

```ant
let stockholm = Result.unwrap(Time.zone("Europe/Stockholm"))
let tokyo = Result.unwrap(Time.zone("Asia/Tokyo"))
let call = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(9)).toZoned(stockholm)
call                 // 2026-10-22T09:00:00+02:00[Europe/Stockholm]
call.withZone(tokyo) // 2026-10-22T16:00:00+09:00[Asia/Tokyo]
call.offset()        // "+02:00"
```

A wall-clock time in a daylight-saving gap does not exist, and `toZoned` moves it forward unless
`disambiguation:` says otherwise. A program that must refuse such a time checks `possibleInstants`:

```ant
let ny = Result.unwrap(Time.zone("America/New_York"))
let gap = Time.plainDateTime(Time.plainDate(2026, 3, 8), Time.plainTime(2, 30))
gap.toZoned(ny)                                 // 2026-03-08T03:30:00-04:00[America/New_York]
gap.toZoned(ny, disambiguation: Time.Earlier{}) // 2026-03-08T01:30:00-05:00[America/New_York]
gap.possibleInstants(ny).length()               // 0
```

- [Time.zone](../Time/zone.md): a zone from an IANA name or an offset, or an `Err`
- `Time.utc`: the zone UTC
- [Time.zoneId](../Time/zoneId.md): the name or offset of a zone
- [Time.inZone](../Time/inZone.md): an instant seen in a zone
- [Time.withZone](../Time/withZone.md): the same moment in another zone
- [Time.toZoned](../Time/toZoned.md): a wall-clock time placed in a zone
- [Time.possibleInstants](../Time/possibleInstants.md): every moment a wall-clock time has in a
  zone, none or two around a daylight-saving change
- [Time.toInstant](../Time/toInstant.md): the moment of a zoned time
- [Time.toPlainDate](../Time/toPlainDate.md): the date of a zoned or plain date and time
- [Time.toPlainTime](../Time/toPlainTime.md): the clock time of a zoned or plain date and time
- [Time.toPlainDateTime](../Time/toPlainDateTime.md): the wall-clock time of a zoned time
- [Time.startOfDay](../Time/startOfDay.md): the first moment of the day in its zone
- [Time.zoneOf](../Time/zoneOf.md): the zone of a zoned time
- [Time.offset](../Time/offset.md): the offset from UTC as text, such as `+02:00`
- [Time.offsetSeconds](../Time/offsetSeconds.md): the offset from UTC in seconds
- [Time.hoursInDay](../Time/hoursInDay.md): the length of the day, 23 or 25 hours on a
  daylight-saving change
- [Time.nextTransition](../Time/nextTransition.md): the next change of the zone's offset
- [Time.previousTransition](../Time/previousTransition.md): the last change before

## Read and change fields

```ant
let day = Time.plainDate(2026, 10, 22)
day.year()                  // 2026
day.dayOfWeek()             // 4
day.daysInMonth()           // 31
day.with(day: 1)            // 2026-10-01
day.with(month: 4, day: 31) // 2026-04-30
```

- [Time.year](../Time/year.md), [Time.month](../Time/month.md), [Time.day](../Time/day.md): the
  parts of the date, the month from 1 to 12
- [Time.dayOfWeek](../Time/dayOfWeek.md): 1 for Monday to 7 for Sunday
- [Time.dayOfYear](../Time/dayOfYear.md): the day's number in its year
- [Time.weekOfYear](../Time/weekOfYear.md): the ISO week number
- [Time.yearOfWeek](../Time/yearOfWeek.md): the year that ISO week belongs to
- [Time.daysInMonth](../Time/daysInMonth.md): the length of the month
- [Time.daysInYear](../Time/daysInYear.md): 365 or 366
- [Time.inLeapYear](../Time/inLeapYear.md): whether the year has a February 29
- [Time.hour](../Time/hour.md), [Time.minute](../Time/minute.md), [Time.second](../Time/second.md):
  the parts of the clock time
- [Time.millisecond](../Time/millisecond.md), [Time.microsecond](../Time/microsecond.md),
  [Time.nanosecond](../Time/nanosecond.md): the parts of the second
- [Time.with](../Time/with.md): a copy with some fields replaced, a day past the month's end moved
  to its last day

## Calculate with dates and times

Years, months, weeks, and days move the calendar; hours and smaller units move the moment on the
time line, whatever the clock shows. A month added to January 31 lands on the last day of February:

```ant
let start = Time.plainDate(2026, 1, 31)
start.add(Time.duration(months: 1))                        // 2026-02-28
start.until(Time.plainDate(2026, 12, 25))                  // P328D
Time.plainTime(14, 37).round(Time.Minute{}, increment: 15) // 14:30:00
Time.compare(start, Time.plainDate(2026, 3, 1))            // -1
```

- [Time.add](../Time/add.md): a value moved forward by a duration
- [Time.subtract](../Time/subtract.md): a value moved back by a duration
- [Time.until](../Time/until.md): the duration from one value to another, negative when the other is
  earlier
- [Time.since](../Time/since.md): the duration from an earlier value to this one
- [Time.round](../Time/round.md): a value rounded to a unit
- [Time.compare](../Time/compare.md): the order of two values, since `<` does not apply to them

## Work with durations

A duration keeps its fields as written, so `90m` and `1h30m` are not `==`. Compare and convert them
by length:

```ant
Time.compareDuration(90m, 1h30m)                  // 0
Time.total(1h30m, Time.Minute{})                  // 90.0
Time.roundDuration(90m, largestUnit: Time.Hour{}) // PT1H30M
```

A month or a year has no fixed length, so give a `relativeTo:` date to count from:

```ant
let month = Time.duration(months: 1)
Time.total(month, Time.Day{}, relativeTo: Time.plainDate(2026, 2, 1)) // 28.0
Time.total(month, Time.Day{})                                         // panic: Time.total: a total with months needs a relativeTo point
```

- [Time.part](../Time/part.md): one field as written
- [Time.total](../Time/total.md): the whole span in one unit, with a fraction
- [Time.roundDuration](../Time/roundDuration.md): a span balanced into larger units and rounded
- [Time.compareDuration](../Time/compareDuration.md): the order of two spans by length
- [Time.addDuration](../Time/addDuration.md): the sum of two spans
- [Time.sign](../Time/sign.md): -1, 0, or 1
- [Time.negate](../Time/negate.md): the span with its sign flipped
- [Time.abs](../Time/abs.md): the span made positive

## Measure, wait, and exchange timestamps

`Time.now` reads the wall clock, which the host can set back. Measure elapsed time with
`Time.monotonic`, which only moves forward:

```ant
let start = Time.monotonic()
Time.sleep(20ms)
println(Time.monotonic() - start >= 20)
```

```text
true
```

- [Time.monotonic](../Time/monotonic.md): milliseconds from a fixed origin, for measuring
- [Time.sleep](../Time/sleep.md): a pause of the calling goroutine for a duration
- [Time.ticker](../Time/ticker.md): a channel that receives the time at every interval
- [Time.epochMillis](../Time/epochMillis.md): milliseconds since 1970, as other systems store time
- [Time.epochNanos](../Time/epochNanos.md): nanoseconds since 1970

## See also

- [Time](../Time/index.md): the namespace, with every type and function
- [Duration literals](../../language/literals/durations.md): `500ms`, `2h30m`, and their rules
- [Context](../Context/index.md): deadlines and timeouts, given as a `Time.Duration`
- How-to: [get the current time and print it](../../howto/time/now-format.md),
  [parse dates and times](../../howto/time/parse.md),
  [convert between time zones](../../howto/time/zones.md),
  [add and round durations](../../howto/time/durations.md),
  [measure elapsed time](../../howto/time/measure.md)
