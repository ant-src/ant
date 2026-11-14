<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.nextTransition

```ant signature
func Time.nextTransition(t: Time.ZonedDateTime): Time.ZonedDateTime | null
```

Time.nextTransition returns the earliest moment after t at which the zone of t changes its offset,
in that zone. It is null when there is none, as for an offset zone or a zone that has stopped
changing.

<!-- end generated -->

## Details

- The result is strictly after `t`: called on a transition, it returns the one after that.
- The result is in the zone of `t` and shows the offset that starts at the transition, so it is the
  first moment of the new offset.
- It is `null` for an offset zone such as `+05:30`, and for a named zone with no change ahead of
  `t`, such as `UTC` or `Asia/Tokyo`. Narrow the result before you use it.

## Basic examples

New York moves its clocks back on November 1, 2026. At 02:00 daylight time the clock shows 01:00
again, now at `-05:00`:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
Time.nextTransition(t) // 2026-11-01T01:00:00-05:00[America/New_York]
```

UTC never changes its offset:

```ant
let t = Time.parseZoned("2026-06-01T12:00[UTC]").unwrap()
Time.nextTransition(t) // null
```

## Scope

Call it through `.`:

```ant
let winter = Time.parseZoned("2026-01-15T09:00[Europe/Stockholm]").unwrap()
winter.nextTransition() // 2026-03-29T03:00:00+02:00[Europe/Stockholm]
```

The southern hemisphere changes the other way round. Sydney moves its clocks forward in October:

```ant
let sydney = Time.parseZoned("2026-06-01T12:00[Australia/Sydney]").unwrap()
sydney.nextTransition() // 2026-10-04T03:00:00+11:00[Australia/Sydney]
```

Lord Howe Island moves its clocks by half an hour:

```ant
let lordHowe = Time.parseZoned("2026-06-01T12:00[Australia/Lord_Howe]").unwrap()
lordHowe.nextTransition() // 2026-10-04T02:30:00+11:00[Australia/Lord_Howe]
```

An offset zone has a fixed offset, so it has no transition:

```ant
let india = Time.fromEpochMillis(0).inZone(Time.zone("+05:30").unwrap())
india.nextTransition() // null
```

A named zone that no longer changes its offset has no next transition either, though it has past
ones. Tokyo last changed its clocks in 1951:

```ant
let tokyo = Time.parseZoned("2026-06-01T12:00[Asia/Tokyo]").unwrap()
tokyo.nextTransition()     // null
tokyo.previousTransition() // 1951-09-09T00:00:00+09:00[Asia/Tokyo]
```

Only the instant and the zone of `t` matter. The same moment in another zone finds that zone's next
change:

```ant
let newYork = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
let berlin = newYork.withZone(Time.zone("Europe/Berlin").unwrap())
berlin.nextTransition() // 2026-10-25T02:00:00+01:00[Europe/Berlin]
```

## Applications

List the clock changes of a year by calling it again on each result:

```ant
let end = Time.parseZoned("2027-01-01T00:00[Europe/London]").unwrap()
var t = Time.parseZoned("2026-01-01T00:00[Europe/London]").unwrap()
while true {
  let next = t.nextTransition()
  if next == null || Time.compare(next, end) >= 0 {
    break
  }
  println("${next.toPlainDate()}: ${next.offset()}")
  t = next
}
```

```text
2026-03-29: +01:00
2026-10-25: +00:00
```

Tell a user how the clocks move next, from the offsets on either side of the change:

```ant
func clockChange(t: Time.ZonedDateTime): string {
  let next = t.nextTransition()
  if next == null {
    return "no clock change ahead"
  }
  let before = next.subtract(1ns).offsetSeconds()
  let minutes = (next.offsetSeconds() - before) / 60
  let way = if minutes > 0 { "forward" } else { "back" }
  "on ${next.toPlainDate()} clocks go ${way} ${Math.abs(minutes)} minutes"
}

println(clockChange(Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()))
println(clockChange(Time.parseZoned("2026-06-01T12:00[Australia/Lord_Howe]").unwrap()))
println(clockChange(Time.parseZoned("2026-06-01T12:00[Asia/Tokyo]").unwrap()))
```

```text
on 2026-11-01 clocks go back 60 minutes
on 2026-10-04 clocks go forward 30 minutes
no clock change ahead
```

Count the days left until the next change:

```ant
let today = Time.parseZoned("2026-10-20T08:00[Europe/Paris]").unwrap()
let next = today.nextTransition()
if next != null {
  let days = Time.until(today.toPlainDate(), next.toPlainDate()).part(Time.Day{})
  println("clocks change in ${days} days")
}
```

```text
clocks change in 5 days
```

## Properties and relations

The transition is after `t`, in the zone of `t`, and the offset changes there:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
let next = t.nextTransition()
assert(next != null)
assert(Time.compare(next, t) > 0)
assert(next.zoneOf() == t.zoneOf())
let justBefore = next.subtract(1ns)
assert(justBefore.offset() != next.offset())
```

[Time.previousTransition](previousTransition.md) looks the other way. A nanosecond after a
transition, the previous one is that transition:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
let next = t.nextTransition()
assert(next != null)
let after = next.add(1ns)
assert(after.previousTransition() == next)
```

The day of a transition is not 24 hours long. In New York the autumn change makes a 25-hour day:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
let next = t.nextTransition()
assert(next != null)
assert(next.hoursInDay() == 25.0)
```

## Possible issues

The result may be `null`, so calling a function on it before narrowing is a compile error:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
println(t.nextTransition().offset())
// error: Time.ZonedDateTime | null has no field or function offset on every member: narrow the value first
```

Narrow it with `if`, or use `?.` and `??`:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
t.nextTransition()?.offset() ?? "none" // "-05:00"
```

Called on a transition, it does not return that transition but the next one:

```ant
let fallBack = Time.parseZoned("2026-11-01T01:00:00-05:00[America/New_York]").unwrap()
fallBack.nextTransition() // 2027-03-14T03:00:00-04:00[America/New_York]
```

When the clocks go back, the wall-clock time of the transition happens twice, and the transition is
the second time. Turning it into a wall-clock time and back gives the first, an hour earlier:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
let next = t.nextTransition()
if next != null {
  let again = next.toPlainDateTime().toZoned(next.zoneOf())
  println(again)
  println(again == next)
}
```

```text
2026-11-01T01:00:00-04:00[America/New_York]
false
```

Keep the `Time.ZonedDateTime` itself, or its instant, to keep the moment.

## See also

[Time.previousTransition](previousTransition.md) · [Time.hoursInDay](hoursInDay.md) ·
[Time.offset](offset.md) · [Time.offsetSeconds](offsetSeconds.md) ·
[Time.possibleInstants](possibleInstants.md) · [Time.zone](zone.md) ·
[Time.ZonedDateTime](types/ZonedDateTime.md)
