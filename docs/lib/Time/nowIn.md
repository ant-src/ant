<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.nowIn

```ant signature
func Time.nowIn(z: Time.Zone): Time.ZonedDateTime
```

Time.nowIn returns the current instant seen in zone z, read from the wall clock as
[`Time.now`](now.md) reads it.

<!-- end generated -->

## Details

- Each call reads the clock again. Read the clock once and keep the value when several parts of a
  program need the same moment.
- The result has the zone z, and its date, time, and offset are what a wall clock in z shows at that
  moment.
- It is [Time.now](now.md) put in a zone with [Time.inZone](inZone.md), in one step.

## Basic examples

Print the time in Tokyo:

<!-- test: check-only prints the clock -->

```ant
let tokyo = Time.zone("Asia/Tokyo").unwrap()
println(Time.nowIn(tokyo))
```

It prints the wall clock in Tokyo with its offset and zone, such as
`2026-11-12T06:26:59.512153+09:00[Asia/Tokyo]`.

The result is in the zone you pass. Tokyo has no daylight saving, so its offset is always the same:

```ant
let tokyo = Time.zone("Asia/Tokyo").unwrap()
Time.nowIn(tokyo).zoneOf() == tokyo // true
Time.nowIn(tokyo).offset()          // "+09:00"
```

## Scope

[Time.utc](zone.md) gives the time in UTC:

```ant
Time.nowIn(Time.utc).offset() // "+00:00"
Time.nowIn(Time.utc).zoneOf() // UTC
```

An offset zone keeps its fixed offset all year:

```ant
let india = Time.zone("+05:30").unwrap()
Time.nowIn(india).offsetSeconds() // 19800
```

[Time.localZone](localZone.md) gives the time on the host's wall clock:

<!-- test: check-only prints the clock and the host's zone -->

```ant
let local = Time.nowIn(Time.localZone())
println("${local.toPlainTime().round(Time.Minute{})} in ${local.zoneOf().zoneId()}")
```

On a machine set to Zurich it prints a line such as `22:27:00 in Europe/Zurich`.

## Applications

Show the time in several offices at once. Read the clock once, so every row shows the same moment:

<!-- test: check-only prints the clock -->

```ant
let offices = ["America/New_York", "Europe/London", "Asia/Singapore"]
let now = Time.now()
for id in offices {
  let local = now.inZone(Time.zone(id).unwrap())
  println("${id}: ${local.toPlainTime().round(Time.Minute{})}")
}
```

It prints one line per office, such as `America/New_York: 16:27:00`, `Europe/London: 21:27:00`, and
`Asia/Singapore: 05:27:00`.

Take today's date as seen in a zone, here to count the days to a deadline:

<!-- test: check-only prints the clock -->

```ant
let stockholm = Time.zone("Europe/Stockholm").unwrap()
let today = Time.nowIn(stockholm).toPlainDate()
let deadline = Time.plainDate(2026, 12, 24)
println("${today.until(deadline).part(Time.Day{})} days to go")
```

On November 11, 2026, in Stockholm, it prints `43 days to go`.

Write a rule as a function of a `Time.ZonedDateTime` and call `Time.nowIn` only where the rule is
used. A test then passes a fixed time:

```ant
func isOpen(t: Time.ZonedDateTime): bool {
  if t.dayOfWeek() > 5 {
    return false
  }
  t.hour() >= 9 && t.hour() < 17
}

let newYork = Time.zone("America/New_York").unwrap()
let monday = Time.parseZoned("2026-11-09T10:00:00-05:00[America/New_York]").unwrap()
let saturday = Time.parseZoned("2026-11-14T10:00:00-05:00[America/New_York]").unwrap()
println(isOpen(monday))
println(isOpen(saturday))

func openNow(): bool {
  isOpen(Time.nowIn(newYork))
}
```

```text
true
false
```

## Properties and relations

`Time.nowIn(z)` is the instant `Time.now()` reads, in zone z:

```ant
let tokyo = Time.zone("Asia/Tokyo").unwrap()
let seen = Time.nowIn(tokyo)
assert(seen.zoneOf() == tokyo)
assert(seen == seen.toInstant().inZone(tokyo))
```

The same instant in two zones shows two wall clocks but one instant:

```ant
let now = Time.now()
let tokyo = now.inZone(Time.zone("Asia/Tokyo").unwrap())
let london = now.inZone(Time.zone("Europe/London").unwrap())
assert(tokyo.toInstant() == london.toInstant())
assert(tokyo.epochMillis() == london.epochMillis())
```

## Possible issues

`Time.nowIn` takes a `Time.Zone`, not its name. Make the zone with [Time.zone](zone.md) first:

```ant
let t = Time.nowIn("Asia/Tokyo") // error: cannot use string where Time.Zone is expected
```

`Time.utc` is a named zone, so it is not equal to the offset zone `+00:00`, although both show the
same wall clock:

```ant
let zero = Time.zone("+00:00").unwrap()
Time.nowIn(Time.utc).zoneOf() == zero // false
Time.nowIn(zero).offset()             // "+00:00"
```

Two calls read the clock twice, so a date from one call and a time from another can mix two moments,
such as a date from before midnight and a time from after. Call `Time.nowIn` once and take every
part from that value:

```ant
let tokyo = Time.zone("Asia/Tokyo").unwrap()
let now = Time.nowIn(tokyo)
let date = now.toPlainDate()
let time = now.toPlainTime()
assert(Time.plainDateTime(date, time) == now.toPlainDateTime())
```

To time a piece of code, use [Time.monotonic](monotonic.md), the reading the library gives for
measuring, not two `Time.nowIn` readings:

<!-- test: check-only the elapsed time changes from run to run -->

```ant
func work() {
  Time.sleep(Time.duration(milliseconds: 20))
}

let start = Time.monotonic()
work()
println("${Time.monotonic() - start} ms")
```

## See also

[Time.now](now.md) · [Time.inZone](inZone.md) · [Time.localZone](localZone.md) ·
[Time.zone](zone.md) · [Time.zoneOf](zoneOf.md) · [Time.toPlainDate](toPlainDate.md) ·
[Time.ZonedDateTime](types/ZonedDateTime.md) · [Time](index.md)
