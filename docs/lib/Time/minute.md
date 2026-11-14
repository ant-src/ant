<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.minute

```ant signature
func Time.minute(t: Time.Timed): int
```

Time.minute returns the minute of t, 0 to 59.

<!-- end generated -->

## Details

- The minute is the one the wall clock shows, 0 to 59. On a `Time.ZonedDateTime` it depends on the
  zone: a zone with a half-hour or quarter-hour offset shows a different minute for the same
  instant.

## Basic examples

```ant
Time.minute(Time.plainTime(14, 5))      // 5
Time.minute(Time.plainTime(hour: 9))    // 0
Time.minute(Time.plainTime(23, 59, 59)) // 59
```

Call it through `.` on any value with a time:

```ant
let alarm = Time.plainTime(hour: 6, minute: 45)
alarm.minute() // 45
```

## Scope

A `Time.PlainDateTime` has the minute of its time:

```ant
let meeting = Time.plainDateTime(Time.plainDate(2026, 11, 11), Time.plainTime(10, 30))
meeting.minute() // 30
```

A `Time.ZonedDateTime` has the minute its zone's wall clock shows:

```ant
let call = Time.parseZoned("2026-11-11T16:20:00+01:00[Europe/Stockholm]").unwrap()
call.minute() // 20
```

A value typed `Time.Timed` may hold any of the three types with a time, and `Time.minute` takes it
without narrowing. Pass `Time.minute` by name to read each one:

```ant
let lunch = Time.plainDateTime(Time.plainDate(2026, 11, 11), Time.plainTime(12, 15))
let launch = Time.parseZoned("2026-11-11T18:45:00Z[UTC]").unwrap()
let times: Array[Time.Timed] = [Time.plainTime(8, 5), lunch, launch]
times.map(Time.minute) // [5, 15, 45]
```

Read the minute of a parsed time:

```ant
"07:42:10" | Time.parsePlainTime() | Result.map(Time.minute) // Ok{ value: 42 }
```

## Named arguments

The parameter is named `t`:

```ant
Time.minute(t: Time.plainTime(13, 37)) // 37
```

## Applications

Write a time as `H:MM`, padding the minute to two digits:

```ant
func clock(t: Time.PlainTime): string {
  let mm = String.from(t.minute()).padStart(2, "0")
  "${t.hour()}:${mm}"
}

clock(Time.plainTime(9, 5))   // "9:05"
clock(Time.plainTime(17, 30)) // "17:30"
```

Count the minutes since midnight:

```ant
func minutesOfDay(t: Time.PlainTime): int {
  t.hour() * 60 + t.minute()
}

minutesOfDay(Time.plainTime(0, 0))   // 0
minutesOfDay(Time.plainTime(8, 30))  // 510
minutesOfDay(Time.plainTime(23, 59)) // 1439
```

Keep the events that start on the hour or the half hour:

```ant
let starts = [
  Time.plainTime(9, 0),
  Time.plainTime(9, 20),
  Time.plainTime(10, 30),
  Time.plainTime(11, 45),
]
println(starts.filter({ x.minute() % 30 == 0 }))
```

```text
[09:00:00, 10:30:00]
```

Read the current minute. [Time.nowIn](nowIn.md) gives the wall clock of a zone:

<!-- test: check-only output changes between runs -->

```ant
let now = Time.nowIn(Time.localZone())
println(now.minute())
```

## Properties and relations

Setting the minute with [Time.with](with.md) sets what `Time.minute` reads:

```ant
let t = Time.plainTime(10, 15)
for m in Array.range(0, 60) {
  assert(t.with(minute: m).minute() == m)
}
```

The minute stays 0 to 59 however many minutes are added:

```ant
let start = Time.plainTime(10, 15)
for i in Array.range(0, 200) {
  let m = start.add(Time.duration(minutes: i)).minute()
  assert(m >= 0 && m <= 59)
}
```

Adding an hour leaves the minute as it was, and adding a minute carries into the hour at 59:

```ant
let t = Time.plainTime(9, 59)
assert(t.add(1h).minute() == 59)
assert(t.add(1m).minute() == 0)
assert(t.add(1m).hour() == 10)
```

A `Time.PlainDateTime` and a `Time.ZonedDateTime` have the minute of their time:

```ant
let zoned = Time.parseZoned("2026-11-11T21:07:00+01:00[Europe/Stockholm]").unwrap()
let local = zoned.toPlainDateTime()
assert(zoned.minute() == zoned.toPlainTime().minute())
assert(local.minute() == local.toPlainTime().minute())
```

## Possible issues

A `Time.Instant` has no wall clock, so passing one is a compile error:

```ant
println(Time.minute(Time.now())) // error: cannot use Time.Instant where Time.Timed is expected
```

Put the instant in a zone first. The minute depends on the zone: the same instant is on the hour in
UTC, at half past in Kolkata, and at a quarter to in Kathmandu:

```ant
let instant = Time.parseInstant("2026-11-11T09:00:00Z").unwrap()
let kolkata = Time.zone("Asia/Kolkata").unwrap()
let kathmandu = Time.zone("Asia/Kathmandu").unwrap()
instant.inZone(Time.utc).minute()  // 0
instant.inZone(kolkata).minute()   // 30
instant.inZone(kathmandu).minute() // 45
```

A `Time.PlainDate` has no time:

```ant
println(Time.minute(Time.plainDate(2026, 11, 11))) // error: cannot use Time.PlainDate where Time.Timed is expected
```

`Time.minute` reads the clock, not a span. A duration literal keeps its minutes as written, and
[Time.part](part.md) reads them:

```ant
let d = 90m
d.part(Time.Minute{}) // 90
```

The seconds do not round the minute up. Round first with [Time.round](round.md) to read the nearest
minute:

```ant
let t = Time.plainTime(10, 29, 50)
t.minute()                      // 29
t.round(Time.Minute{}).minute() // 30
```

## See also

[Time.hour](hour.md) · [Time.second](second.md) · [Time.with](with.md) · [Time.round](round.md) ·
[Time.plainTime](plainTime.md) · [Time.inZone](inZone.md) · [Time.nowIn](nowIn.md) ·
[Time.part](part.md) · [Time](index.md)
