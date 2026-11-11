<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.toPlainTime

```ant signature
func Time.toPlainTime(t: Time.ZonedDateTime | Time.PlainDateTime): Time.PlainTime
```

Time.toPlainTime returns the time of day of t, as its wall clock shows it.

<!-- end generated -->

## Details

- The result is the time the wall clock shows: for a `Time.ZonedDateTime`, the clock of its zone. It
  drops the date, the zone, and the offset.
- It keeps the fraction of a second, down to the nanosecond.
- A `Time.Instant` has no wall clock. Put it in a zone with [Time.inZone](inZone.md) first.

## Basic examples

```ant
let meeting = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(14, minute: 30))
Time.toPlainTime(meeting) // 14:30:00
```

```ant
let zoned = Time.parseZoned("2026-10-22T16:03:05.5+02:00[Europe/Stockholm]").unwrap()
Time.toPlainTime(zoned) // 16:03:05.5
```

## Scope

Call it through `.` on a `Time.ZonedDateTime` or a `Time.PlainDateTime`:

```ant
let late = Time.parsePlainDateTime("2026-10-22T23:59:59.999999999").unwrap()
late.toPlainTime() // 23:59:59.999999999
```

One instant shows a different time in each zone:

```ant
let instant = Time.parseInstant("2026-10-22T14:03:05Z").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
instant.inZone(Time.utc).toPlainTime() // 14:03:05
instant.inZone(tokyo).toPlainTime()    // 23:03:05
```

A value typed as the union of the two may hold either. Pass `Time.toPlainTime` by name:

```ant
let zoned = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let local = Time.plainDateTime(Time.plainDate(2026, 1, 6), Time.plainTime(10))
let values: Array[Time.ZonedDateTime | Time.PlainDateTime] = [zoned, local]
values.map(Time.toPlainTime) // [16:03:05, 10:00:00]
```

In a pipe, the value is the first argument:

```ant
"2026-10-22T09:15:00+02:00[Europe/Stockholm]"
  | Time.parseZoned()
  | Result.map(Time.toPlainTime) // Ok{ value: 09:15:00 }
```

## Applications

Read the time of day on the host's clock:

```ant
let now = Time.nowIn(Time.localZone()).toPlainTime()
assert(now.hour() < 24 && now.minute() < 60)
```

Show when a flight leaves and lands, each on its local clock:

```ant
let departs = Time.parseZoned("2026-10-22T10:05:00+02:00[Europe/Stockholm]").unwrap()
let arrives = departs
  .add(Time.duration(hours: 11, minutes: 40))
  .withZone(Time.zone("Asia/Tokyo").unwrap())
println("Departs ${departs.toPlainTime()} Stockholm time")
println("Arrives ${arrives.toPlainTime()} Tokyo time")
```

```text
Departs 10:05:00 Stockholm time
Arrives 04:45:00 Tokyo time
```

Find the shifts that started after 07:00, whatever their date:

```ant
let starts = [
  "2026-10-22T06:55:00+02:00[Europe/Stockholm]",
  "2026-10-22T14:10:00+02:00[Europe/Stockholm]",
  "2026-10-23T22:05:00+02:00[Europe/Stockholm]",
]
let opening = Time.plainTime(7)
let late = starts
  | Array.map({ Time.parseZoned(x).unwrap().toPlainTime() })
  | Array.filter({ x.compare(opening) > 0 })
println(late)
```

```text
[14:10:00, 22:05:00]
```

## Properties and relations

A `Time.ZonedDateTime` splits into its date and its time, which
[Time.plainDateTime](plainDateTime.md) puts back together as its
[Time.toPlainDateTime](toPlainDateTime.md):

```ant
let zoned = Time.parseZoned("2026-10-22T16:03:05+02:00[Europe/Stockholm]").unwrap()
let rebuilt = Time.plainDateTime(zoned.toPlainDate(), zoned.toPlainTime())
assert(rebuilt == zoned.toPlainDateTime())
assert(zoned.toPlainTime() == zoned.toPlainDateTime().toPlainTime())
```

The time has the fields of the value it came from:

```ant
let zoned = Time.parseZoned("2026-10-22T16:03:05.25+02:00[Europe/Stockholm]").unwrap()
let time = zoned.toPlainTime()
assert(time.hour() == zoned.hour())
assert(time.minute() == zoned.minute())
assert(time.millisecond() == zoned.millisecond())
```

## Possible issues

A `Time.Instant` has no wall clock, and a `Time.PlainTime` is already a time, so passing either is a
compile error:

```ant
let epoch = Time.fromEpochMillis(0)
println(Time.toPlainTime(epoch))
// error: cannot use Time.Instant where Time.ZonedDateTime | Time.PlainDateTime is expected
let alarm = Time.plainTime(7)
println(Time.toPlainTime(alarm))
// error: cannot use Time.PlainTime where Time.ZonedDateTime | Time.PlainDateTime is expected
```

In an overlap the clock shows one time twice, so two different instants give equal times. In New
York 01:30 comes twice on November 1, 2026:

```ant
let first = Time.parseZoned("2026-11-01T01:30:00-04:00[America/New_York]").unwrap()
let second = Time.parseZoned("2026-11-01T01:30:00-05:00[America/New_York]").unwrap()
first == second                             // false
first.toPlainTime() == second.toPlainTime() // true
```

A day does not always start at midnight. In São Paulo the clocks jumped from 00:00 to 01:00 on
November 4, 2018, so that day began at 01:00:

```ant
let noon = Time.parseZoned("2018-11-04T12:00:00-02:00[America/Sao_Paulo]").unwrap()
noon.startOfDay().toPlainTime() // 01:00:00
```

## See also

[Time.toPlainDate](toPlainDate.md) · [Time.toPlainDateTime](toPlainDateTime.md) ·
[Time.plainTime](plainTime.md) · [Time.plainDateTime](plainDateTime.md) · [Time.inZone](inZone.md) ·
[Time.startOfDay](startOfDay.md) · [Time](index.md)
