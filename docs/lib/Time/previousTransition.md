<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.previousTransition

```ant signature
func Time.previousTransition(t: Time.ZonedDateTime): Time.ZonedDateTime | null
```

Time.previousTransition returns the latest moment before t at which the zone of t changed its
offset, in that zone. It is null when there is none, as for an offset zone.

<!-- end generated -->

## Details

- The result is strictly before `t`: called on a transition, it returns the one before that.
- The result is in the zone of `t` and shows the offset that starts at the transition, the offset of
  `t` unless `t` is itself a transition.
- It is `null` for UTC, an offset zone such as `+05:30`, and a named zone with no change before `t`.

## Basic examples

New York moved its clocks forward on March 8, 2026. At 02:00 standard time the clock jumped to
03:00, now at `-04:00`:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
Time.previousTransition(t) // 2026-03-08T03:00:00-04:00[America/New_York]
```

UTC has never changed its offset:

```ant
let t = Time.parseZoned("2026-06-01T12:00[UTC]").unwrap()
Time.previousTransition(t) // null
```

## Scope

Call it through `.`:

```ant
let summer = Time.parseZoned("2026-06-01T12:00[Europe/Berlin]").unwrap()
summer.previousTransition() // 2026-03-29T03:00:00+02:00[Europe/Berlin]
```

In the southern hemisphere, June is winter. Sydney moved its clocks back in April:

```ant
let sydney = Time.parseZoned("2026-06-01T12:00[Australia/Sydney]").unwrap()
sydney.previousTransition() // 2026-04-05T02:00:00+10:00[Australia/Sydney]
```

In the hour New York repeats in November, the transition is the start of the repeat:

```ant
let fall = Time.parseZoned("2026-11-01T01:30-05:00[America/New_York]").unwrap()
fall.previousTransition() // 2026-11-01T01:00:00-05:00[America/New_York]
```

Lord Howe Island moves its clocks by half an hour:

```ant
let lordHowe = Time.parseZoned("2026-06-01T12:00[Australia/Lord_Howe]").unwrap()
lordHowe.previousTransition() // 2026-04-05T01:30:00+10:30[Australia/Lord_Howe]
```

A zone that stopped changing its offset long ago still has a previous transition, its last one:

```ant
let tokyo = Time.parseZoned("2026-06-01T12:00[Asia/Tokyo]").unwrap()
let kolkata = Time.parseZoned("2026-06-01T12:00[Asia/Kolkata]").unwrap()
tokyo.previousTransition()   // 1951-09-09T00:00:00+09:00[Asia/Tokyo]
kolkata.previousTransition() // 1945-10-14T23:00:00+05:30[Asia/Kolkata]
```

An offset zone has a fixed offset, so it has no transition:

```ant
let india = Time.fromEpochMillis(0).inZone(Time.zone("+05:30").unwrap())
india.previousTransition() // null
```

Before the first offset change of a named zone there is none either. New York adopted standard time
in 1883:

```ant
let early = Time.parseZoned("1884-01-01T00:00[America/New_York]").unwrap()
let earlier = Time.parseZoned("1800-01-01T00:00[America/New_York]").unwrap()
early.previousTransition()   // 1883-11-18T12:00:00-05:00[America/New_York]
earlier.previousTransition() // null
```

## Applications

Walk back through the recent clock changes by calling it again on each result:

```ant
var t = Time.parseZoned("2026-06-01T12:00[Europe/London]").unwrap()
for _ in Iter.range(0, 4) {
  let previous = t.previousTransition()
  if previous == null {
    break
  }
  println("${previous.toPlainDate()}: ${previous.offset()}")
  t = previous
}
```

```text
2026-03-29: +01:00
2025-10-26: +00:00
2025-03-30: +01:00
2024-10-27: +00:00
```

Tell which way the clocks went at the last change, from the offsets on either side of it:

```ant
func lastChange(t: Time.ZonedDateTime): string {
  let previous = t.previousTransition()
  if previous == null {
    return "the clocks have never changed"
  }
  let before = previous.subtract(Time.duration(nanoseconds: 1)).offsetSeconds()
  let way = if previous.offsetSeconds() > before { "forward" } else { "back" }
  "clocks went ${way} on ${previous.toPlainDate()}"
}

println(lastChange(Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()))
println(lastChange(Time.parseZoned("2026-06-01T12:00[Australia/Sydney]").unwrap()))
println(lastChange(Time.parseZoned("2026-06-01T12:00[UTC]").unwrap()))
```

```text
clocks went forward on 2026-03-08
clocks went back on 2026-04-05
the clocks have never changed
```

Find how long the current offset has been in force:

```ant
let now = Time.parseZoned("2026-06-01T12:00[America/Chicago]").unwrap()
let since = now.previousTransition()
if since != null {
  let days = since.toPlainDate().until(now.toPlainDate()).part(Time.Day{})
  println("${now.offset()} for ${days} days")
}
```

```text
-05:00 for 85 days
```

## Properties and relations

The transition is before `t`, in the zone of `t`, and the offset changes there:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
let previous = t.previousTransition()
assert(previous != null)
assert(Time.compare(previous, t) < 0)
assert(previous.zoneOf() == t.zoneOf())
let justBefore = previous.subtract(Time.duration(nanoseconds: 1))
assert(justBefore.offset() != previous.offset())
```

[Time.nextTransition](nextTransition.md) looks the other way. A nanosecond before a transition, the
next one is that transition:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
let previous = t.previousTransition()
assert(previous != null)
let before = previous.subtract(Time.duration(nanoseconds: 1))
assert(before.nextTransition() == previous)
```

Between `t` and its previous transition the offset does not change, so the transition shows the
offset of `t`:

```ant
let t = Time.parseZoned("2026-12-01T12:00[Europe/Paris]").unwrap()
let previous = t.previousTransition()
assert(previous != null)
assert(previous.offset() == t.offset())
```

## Possible issues

The result may be `null`, so calling a function on it before narrowing is a compile error:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
println(t.previousTransition().toPlainDate())
// error: Time.ZonedDateTime | null has no field or function toPlainDate on every member: narrow the value first
```

Narrow it with `if`, or use `?.` and `??`:

```ant
let t = Time.parseZoned("2026-06-01T12:00[America/New_York]").unwrap()
t.previousTransition()?.offset() ?? "none" // "-04:00"
```

Called on a transition, it does not return that transition but the one before:

```ant
let springForward = Time.parseZoned("2026-03-08T03:00:00-04:00[America/New_York]").unwrap()
springForward.previousTransition() // 2025-11-02T01:00:00-05:00[America/New_York]
```

## See also

[Time.nextTransition](nextTransition.md) · [Time.hoursInDay](hoursInDay.md) ·
[Time.offset](offset.md) · [Time.offsetSeconds](offsetSeconds.md) ·
[Time.possibleInstants](possibleInstants.md) · [Time.zone](zone.md) ·
[Time.ZonedDateTime](types/ZonedDateTime.md)
