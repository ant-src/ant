<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.possibleInstants

```ant signature
func Time.possibleInstants(t: Time.PlainDateTime, z: Time.Zone): Array[Time.Instant]
```

Time.possibleInstants returns the instants at which zone z shows the wall-clock time t, earliest
first. There are none in a daylight-saving gap, two in an overlap and one otherwise, so a program
that refuses ambiguous times tests the length.

<!-- end generated -->

## Details

- The instants are `Time.Instant` values; use [Time.inZone](inZone.md) to see them in the zone
  again.
- In an offset zone, such as `+05:30`, and in `Time.utc`, every wall-clock time has exactly one
  instant.

## Basic examples

A summer noon in New York happens once:

```ant
let ny = Time.zone("America/New_York").unwrap()
let noon = Time.parsePlainDateTime("2026-06-01T12:00").unwrap()
Time.possibleInstants(noon, ny) // [2026-06-01T16:00:00Z]
```

On March 8, 2026, New York moves its clocks from 02:00 to 03:00, so 02:30 never happens:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
Time.possibleInstants(skipped, ny) // []
```

On November 1, 2026, the clocks go back from 02:00 to 01:00, so 01:30 happens twice, an hour apart:

```ant
let ny = Time.zone("America/New_York").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
Time.possibleInstants(repeated, ny) // [2026-11-01T05:30:00Z, 2026-11-01T06:30:00Z]
```

## Scope

Place each instant back in the zone to see both offsets:

```ant
let ny = Time.zone("America/New_York").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
for at in repeated.possibleInstants(ny) {
  println(at.inZone(ny))
}
```

```text
2026-11-01T01:30:00-04:00[America/New_York]
2026-11-01T01:30:00-05:00[America/New_York]
```

At the edges of a change, 02:00 on the spring day is skipped and 03:00 exists; 01:00 on the autumn
day happens twice and 02:00 once:

```ant
let ny = Time.zone("America/New_York").unwrap()
Time.parsePlainDateTime("2026-03-08T02:00").unwrap().possibleInstants(ny).length() // 0
Time.parsePlainDateTime("2026-03-08T03:00").unwrap().possibleInstants(ny).length() // 1
Time.parsePlainDateTime("2026-11-01T01:00").unwrap().possibleInstants(ny).length() // 2
Time.parsePlainDateTime("2026-11-01T02:00").unwrap().possibleInstants(ny).length() // 1
```

Lord Howe Island moves its clocks by half an hour, so its gap and its overlap are half an hour long:

```ant
let lordHowe = Time.zone("Australia/Lord_Howe").unwrap()
let autumn = Time.parsePlainDateTime("2026-04-05T01:45").unwrap()
let spring = Time.parsePlainDateTime("2026-10-04T02:15").unwrap()
autumn.possibleInstants(lordHowe) // [2026-04-04T14:45:00Z, 2026-04-04T15:15:00Z]
spring.possibleInstants(lordHowe) // []
```

UTC and offset zones never skip or repeat a time:

```ant
let noon = Time.parsePlainDateTime("2026-06-01T12:00").unwrap()
noon.possibleInstants(Time.utc)                     // [2026-06-01T12:00:00Z]
noon.possibleInstants(Time.zone("+05:30").unwrap()) // [2026-06-01T06:30:00Z]
```

## Applications

Refuse a time that does not name exactly one moment, instead of letting [Time.toZoned](toZoned.md)
pick one:

```ant
func placeExactly(t: Time.PlainDateTime, z: Time.Zone): Result[Time.ZonedDateTime, Error] {
  let found = t.possibleInstants(z)
  if found.isEmpty() {
    return Result.fail("${t} does not exist in ${z}")
  }
  if found.length() > 1 {
    return Result.fail("${t} happens twice in ${z}")
  }
  Result.ok(found[0].inZone(z))
}

let ny = Time.zone("America/New_York").unwrap()
for text in ["2026-06-01T12:00", "2026-03-08T02:30", "2026-11-01T01:30"] {
  println(placeExactly(Time.parsePlainDateTime(text).unwrap(), ny))
}
```

```text
Ok{ value: 2026-06-01T12:00:00-04:00[America/New_York] }
Err{ error: Error{ message: "2026-03-08T02:30:00 does not exist in America/New_York" } }
Err{ error: Error{ message: "2026-11-01T01:30:00 happens twice in America/New_York" } }
```

Warn about the half hours of a day that a daily job at that time would miss or run twice:

```ant
let berlin = Time.zone("Europe/Berlin").unwrap()
for date in [Time.plainDate(2026, 3, 29), Time.plainDate(2026, 10, 25)] {
  var t = Time.plainDateTime(date, Time.plainTime())
  while t.toPlainDate() == date {
    let count = t.possibleInstants(berlin).length()
    if count != 1 {
      let what = if count == 0 { "skipped" } else { "repeated" }
      println("${t.toPlainTime()} ${what}")
    }
    t = t.add(Time.duration(minutes: 30))
  }
}
```

```text
02:00:00 skipped
02:30:00 skipped
02:00:00 repeated
02:30:00 repeated
```

## Properties and relations

In an overlap, the two instants are the ones [Time.toZoned](toZoned.md) picks with `Time.Earlier`
and `Time.Later`, and both show the same wall-clock time:

```ant
let ny = Time.zone("America/New_York").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
let both = repeated.possibleInstants(ny)
let earlier = repeated.toZoned(ny, disambiguation: Time.Earlier{}).toInstant()
let later = repeated.toZoned(ny, disambiguation: Time.Later{}).toInstant()
assert(both.length() == 2)
assert(earlier != later)
assert(both.some({ x == earlier }))
assert(both.some({ x == later }))
assert(both.every({ x.inZone(ny).toPlainDateTime() == repeated }))
```

When there is exactly one instant, every disambiguation gives it:

```ant
let ny = Time.zone("America/New_York").unwrap()
let noon = Time.parsePlainDateTime("2026-06-01T12:00").unwrap()
let only = noon.possibleInstants(ny)
assert(only.length() == 1)
assert(noon.toZoned(ny).toInstant() == only[0])
assert(noon.toZoned(ny, disambiguation: Time.Earlier{}).toInstant() == only[0])
assert(noon.toZoned(ny, disambiguation: Time.Later{}).toInstant() == only[0])
```

## Possible issues

A skipped time has no instants, so taking the first one panics:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
skipped.possibleInstants(ny)[0] // panic: index 0 out of range for an array of length 0
```

Check the length first, or use `get`, which gives `null`:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
skipped.possibleInstants(ny).get(0) // null
```

[Time.toZoned](toZoned.md) never fails on a skipped time: it takes an instant anyway, by default the
later one, so the wall-clock time you get back is not the one you gave. Use `possibleInstants` to
notice:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
skipped.toZoned(ny)                    // 2026-03-08T03:30:00-04:00[America/New_York]
skipped.possibleInstants(ny).isEmpty() // true
```

## See also

[Time.toZoned](toZoned.md) · [Time.Disambiguation](types/Disambiguation.md) ·
[Time.inZone](inZone.md) · [Time.parseZoned](parseZoned.md) · [Time.with](with.md) ·
[Time.nextTransition](nextTransition.md) · [Time.hoursInDay](hoursInDay.md) ·
[Time.PlainDateTime](types/PlainDateTime.md)
