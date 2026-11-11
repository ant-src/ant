<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.toPlainDateTime

```ant signature
func Time.toPlainDateTime(t: Time.ZonedDateTime): Time.PlainDateTime
```

Time.toPlainDateTime returns the date and time of t as its wall clock shows it, without its zone.

<!-- end generated -->

## Details

- The result keeps the date and time the zone shows at that instant and drops the zone and the
  offset. It no longer names a moment on the time line; [Time.toInstant](toInstant.md) keeps the
  moment.

## Basic examples

```ant
let t = Time.parseZoned("2026-10-22T14:03:05+02:00[Europe/Stockholm]").unwrap()
Time.toPlainDateTime(t) // 2026-10-22T14:03:05
```

The same instant in another zone shows another wall-clock time:

```ant
let t = Time.parseZoned("2026-10-22T14:03:05+02:00[Europe/Stockholm]").unwrap()
let ny = Time.zone("America/New_York").unwrap()
Time.toPlainDateTime(t.withZone(ny)) // 2026-10-22T08:03:05
```

## Scope

The fraction of a second is kept, down to the nanosecond:

```ant
let t = Time.parseZoned("2026-10-22T14:03:05.123456789Z[UTC]").unwrap()
t.toPlainDateTime() // 2026-10-22T14:03:05.123456789
```

A `Time.Instant` has no wall clock until you place it in a zone with [Time.inZone](inZone.md):

```ant
let tokyo = Time.zone("Asia/Tokyo").unwrap()
Time.fromEpochMillis(0).inZone(tokyo).toPlainDateTime() // 1970-01-01T09:00:00
```

In a pipe, the zoned value from a parse goes through `Result.map`:

```ant
"2026-12-24T18:00+01:00[Europe/Berlin]"
  | Time.parseZoned()
  | Result.map({ x.toPlainDateTime() }) // Ok{ value: 2026-12-24T18:00:00 }
```

## Applications

Show each event at the local time of the place it happens, without the offset:

```ant
type Event = { name: string, at: Time.ZonedDateTime }

let events = [
  Event{ name: "launch", at: Time.parseZoned("2026-10-22T09:00+02:00[Europe/Stockholm]").unwrap() },
  Event{ name: "review", at: Time.parseZoned("2026-10-22T15:30-04:00[America/New_York]").unwrap() },
]
for e in events {
  println("${e.name}: ${e.at.toPlainDateTime()} local time")
}
```

```text
launch: 2026-10-22T09:00:00 local time
review: 2026-10-22T15:30:00 local time
```

Keep the wall-clock time and change the zone: a call at 23:30 in Stockholm moves to 23:30 in Tokyo.
[Time.withZone](withZone.md) keeps the instant instead:

```ant
let call = Time.parseZoned("2026-10-22T23:30+02:00[Europe/Stockholm]").unwrap()
let tokyo = Time.zone("Asia/Tokyo").unwrap()
call.toPlainDateTime().toZoned(tokyo) // 2026-10-22T23:30:00+09:00[Asia/Tokyo]
call.withZone(tokyo)                  // 2026-10-23T06:30:00+09:00[Asia/Tokyo]
```

Measure on the wall clock instead of the time line. Stockholm turns its clocks back on October 25,
2026, so noon to noon is 25 hours but one calendar day:

```ant
let start = Time.parseZoned("2026-10-24T12:00+02:00[Europe/Stockholm]").unwrap()
let end = Time.parseZoned("2026-10-25T12:00+01:00[Europe/Stockholm]").unwrap()
start.until(end)                                     // PT25H
start.toPlainDateTime().until(end.toPlainDateTime()) // P1D
```

## Properties and relations

The result is the date and the time of `t`, put together by [Time.plainDateTime](plainDateTime.md):

```ant
let t = Time.parseZoned("2026-10-22T14:03:05+02:00[Europe/Stockholm]").unwrap()
assert(t.toPlainDateTime() == Time.plainDateTime(t.toPlainDate(), t.toPlainTime()))
```

[Time.toZoned](toZoned.md) in the zone of `t` gives `t` back when that wall-clock time occurs once:

```ant
let t = Time.parseZoned("2026-10-22T14:03:05+02:00[Europe/Stockholm]").unwrap()
assert(t.toPlainDateTime().toZoned(t.zoneOf()) == t)
```

## Possible issues

Dropping the zone loses which of two repeated times a value was. Stockholm shows 02:30 twice on
October 25, 2026; both give the same result, and `Time.toZoned` picks the earlier one by default:

```ant
let first = Time.parseZoned("2026-10-25T02:30+02:00[Europe/Stockholm]").unwrap()
let second = Time.parseZoned("2026-10-25T02:30+01:00[Europe/Stockholm]").unwrap()
let wall = second.toPlainDateTime()
let stockholm = second.zoneOf()
first.toPlainDateTime() == wall           // true
wall.possibleInstants(stockholm).length() // 2
wall.toZoned(stockholm)                   // 2026-10-25T02:30:00+02:00[Europe/Stockholm]
```

Keep the `Time.ZonedDateTime`, or its [Time.toInstant](toInstant.md), where the two must stay apart.

`Time.toPlainDateTime` takes only a zoned value. A `Time.Instant`, such as the one
[Time.now](now.md) returns, needs a zone first:

```ant
let now = Time.toPlainDateTime(Time.now())
// error: cannot use Time.Instant where Time.ZonedDateTime is expected
```

Read the current time in a zone with [Time.nowIn](nowIn.md), or place the instant in the host's zone
from [Time.localZone](localZone.md):

<!-- test: check-only reads the clock -->

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
println(Time.nowIn(stockholm).toPlainDateTime())
println(Time.now().inZone(Time.localZone()).toPlainDateTime())
```

A `Time.PlainDateTime` already is one, so there is nothing to convert:

```ant
let t = Time.plainDateTime(Time.plainDate(2026, 10, 22), Time.plainTime(9))
let same = Time.toPlainDateTime(t)
// error: cannot use Time.PlainDateTime where Time.ZonedDateTime is expected
```

## See also

[Time.toPlainDate](toPlainDate.md) · [Time.toPlainTime](toPlainTime.md) ·
[Time.toInstant](toInstant.md) · [Time.toZoned](toZoned.md) · [Time.withZone](withZone.md) ·
[Time.inZone](inZone.md) · [Time.plainDateTime](plainDateTime.md) ·
[Time.ZonedDateTime](types/ZonedDateTime.md) · [Time.PlainDateTime](types/PlainDateTime.md) ·
[Time](index.md)
