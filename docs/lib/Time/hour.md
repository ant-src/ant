<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.hour

```ant signature
func Time.hour(t: Time.Timed): int
```

Time.hour returns the hour of t, 0 to 23.

<!-- end generated -->

## Details

- On a `Time.ZonedDateTime`, the hour is the one the wall clock in its zone shows. The same instant
  has a different hour in another zone.
- A `Time.Instant` and a `Time.PlainDate` have no time of day and so no hour. Put an instant in a
  zone with [Time.inZone](inZone.md) first.

## Basic examples

```ant
Time.hour(Time.plainTime(14, minute: 30)) // 14
Time.hour(Time.plainTime())               // 0
```

Call it through `.` on any value with a time:

```ant
let closing = Time.plainTime(17, minute: 45)
closing.hour() // 17
```

## Scope

A `Time.PlainDateTime` has the hour of its time, whatever the date:

```ant
let start = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(9, minute: 15))
start.hour() // 9
```

A `Time.ZonedDateTime` has the hour its zone's wall clock shows:

```ant
let tokyo = Time.parseZoned("2026-10-22T08:00:00+09:00[Asia/Tokyo]").unwrap()
tokyo.hour() // 8
```

A value typed `Time.Timed` may hold any of the three types with a time, and `Time.hour` takes it
without narrowing. Pass `Time.hour` by name to read each one:

```ant
let breakfast = Time.plainTime(7, minute: 30)
let meeting = Time.plainDateTime(Time.plainDate(2026, 1, 6), Time.plainTime(10))
let launch = Time.parseZoned("2026-01-07T21:00:00Z[UTC]").unwrap()
let times: Array[Time.Timed] = [breakfast, meeting, launch]
times.map(Time.hour) // [7, 10, 21]
```

In a pipe, the value is the first argument:

```ant
"09:15" | Time.parsePlainTime() | Result.map(Time.hour) // Ok{ value: 9 }
```

## Applications

Greet by the time of day:

```ant
func greeting(t: Time.Timed): string {
  let h = t.hour()
  if h < 12 {
    "Good morning"
  } else if h < 18 {
    "Good afternoon"
  } else {
    "Good evening"
  }
}

println(greeting(Time.plainTime(8)))
println(greeting(Time.plainTime(13, minute: 5)))
println(greeting(Time.plainTime(22)))
```

```text
Good morning
Good afternoon
Good evening
```

Write a time on a 12-hour clock:

```ant
func twelveHour(t: Time.PlainTime): string {
  let h = t.hour()
  let shown = if h % 12 == 0 { 12 } else { h % 12 }
  let suffix = if h < 12 { "AM" } else { "PM" }
  let minute = "${t.minute()}".padStart(2, "0")
  "${shown}:${minute} ${suffix}"
}

println(twelveHour(Time.plainTime(0, minute: 5)))
println(twelveHour(Time.plainTime(9, minute: 30)))
println(twelveHour(Time.plainTime(12)))
println(twelveHour(Time.plainTime(18, minute: 45)))
```

```text
12:05 AM
9:30 AM
12:00 PM
6:45 PM
```

Count log entries per hour:

```ant
let stamps = ["08:12", "08:47", "09:03", "11:30", "11:31", "11:59"]
let perHour = stamps
  | Array.map({ Time.parsePlainTime(x).unwrap().hour() })
  | Array.groupBy({ x })
  | Map.entries()
  | Array.map({ "${x.key}h: ${x.value.length()}" })
for line in perHour {
  println(line)
}
```

```text
8h: 2
9h: 1
11h: 3
```

Keep the orders placed within business hours, 9:00 to 17:00:

```ant
let placed = [
  Time.plainTime(8, minute: 59),
  Time.plainTime(9),
  Time.plainTime(16, minute: 59),
  Time.plainTime(17),
]
let open = placed.filter({ x.hour() >= 9 && x.hour() < 17 })
println(open)
```

```text
[09:00:00, 16:59:00]
```

## Properties and relations

The hour is between 0 and 23:

```ant
let times = [Time.plainTime(), Time.plainTime(12), Time.plainTime(23, minute: 59, second: 59)]
for t in times {
  assert(t.hour() >= 0)
  assert(t.hour() <= 23)
}
```

`Time.hour` returns the hour a time was made with, and setting the hour with [Time.with](with.md)
sets what `Time.hour` reads:

```ant
let t = Time.plainTime(6, minute: 20)
assert(Time.plainTime(15).hour() == 15)
assert(t.with(hour: 11).hour() == 11)
assert(t.with(hour: 11).minute() == 20)
```

A `Time.PlainDateTime` and a `Time.ZonedDateTime` have the hour of their time:

```ant
let zoned = Time.parseZoned("2026-10-22T23:30:00+02:00[Europe/Stockholm]").unwrap()
let local = zoned.toPlainDateTime()
assert(zoned.hour() == zoned.toPlainTime().hour())
assert(local.hour() == local.toPlainTime().hour())
```

## Possible issues

A `Time.Instant` and a `Time.PlainDate` have no hour, so passing one is a compile error:

```ant
let epoch = Time.fromEpochMillis(0)
println(Time.hour(epoch)) // error: cannot use Time.Instant where Time.Timed is expected
let date = Time.plainDate(2026, 1, 1)
println(Time.hour(date)) // error: cannot use Time.PlainDate where Time.Timed is expected
```

Put the instant in a zone first. The hour depends on the zone:

```ant
let instant = Time.parseInstant("2026-10-21T23:30:00Z").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
instant.inZone(Time.utc).hour() // 23
instant.inZone(tokyo).hour()    // 8
```

`Time.plainTime` panics on hour 24, but `Time.with` constrains the hour to 0 to 23:

```ant
Time.plainTime(24)                      // panic: Time.plainTime: hour out of range: 24
Time.plainTime(9).with(hour: 25).hour() // 23
Time.plainTime(9).with(hour: -3).hour() // 0
```

A `Time.PlainTime` wraps around the clock, so adding hours can give a smaller hour:

```ant
let late = Time.plainTime(23, minute: 30)
late.add(2h).hour() // 1
```

On a daylight-saving change, an hour added is an hour of real time, not of the clock. In Stockholm
the clocks jump from 02:00 to 03:00 on March 29, 2026, so an hour after 01:30 is 03:30:

```ant
let before = Time.parseZoned("2026-03-29T01:30:00+01:00[Europe/Stockholm]").unwrap()
before.add(1h).hour() // 3
```

When the clocks go back, one hour happens twice. In Stockholm 02:00 to 03:00 repeats on October 25,
2026, so two instants an hour apart have the same hour:

```ant
let first = Time.parseZoned("2026-10-25T02:30:00+02:00[Europe/Stockholm]").unwrap()
let second = first.add(1h)
println(first.hour())
println(second.hour())
println(second)
```

```text
2
2
2026-10-25T02:30:00+01:00[Europe/Stockholm]
```

## See also

[Time.minute](minute.md) · [Time.second](second.md) · [Time.hoursInDay](hoursInDay.md) ·
[Time.with](with.md) · [Time.round](round.md) · [Time.plainTime](plainTime.md) ·
[Time.inZone](inZone.md) · [Time](index.md)
