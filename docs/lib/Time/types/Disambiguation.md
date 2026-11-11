<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Disambiguation

```ant signature
type Time.Disambiguation = Time.Compatible | Time.Earlier | Time.Later
```

Time.Disambiguation says which instant to take for a wall-clock time that a daylight-saving change
skipped or repeated in a zone. Such a time has no instant in a gap and two in an overlap;
[`Time.possibleInstants`](../possibleInstants.md) lists them.

```ant signature
type Time.Compatible = {}
```

Time.Compatible is the disambiguation that takes the later instant for a wall-clock time in a gap
and the earlier for one in an overlap. It is the default, and what the common date libraries do.

```ant signature
type Time.Earlier = {}
```

Time.Earlier is the disambiguation that takes the earlier instant for a wall-clock time in a gap or
an overlap.

```ant signature
type Time.Later = {}
```

Time.Later is the disambiguation that takes the later instant for a wall-clock time in a gap or an
overlap.

<!-- end generated -->

## Details

- The three members are records with no fields, so a value is written with braces:
  `Time.Compatible{}`, `Time.Earlier{}`, `Time.Later{}`.
- It matters only for a wall-clock time that is skipped or repeated. Every other time has one
  instant, and all three give it. Where a function takes one, `Time.Compatible{}` is the default.
- `Time.parseZoned` uses it only for a string with a time and no offset, and `Time.with` only when
  the value's own offset does not fit the new time. [Time.add](../add.md) takes no disambiguation:
  it places a `ZonedDateTime` moved by days or more with `Compatible`.

## Basic examples

On November 1, 2026, New York moves its clocks back from 02:00 to 01:00, so 01:30 happens twice.
`Time.Earlier` takes the first, with the summer offset, and `Time.Later` the second:

```ant
let ny = Time.zone("America/New_York").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
Time.toZoned(repeated, ny, disambiguation: Time.Earlier{})    // 2026-11-01T01:30:00-04:00[America/New_York]
Time.toZoned(repeated, ny, disambiguation: Time.Later{})      // 2026-11-01T01:30:00-05:00[America/New_York]
Time.toZoned(repeated, ny, disambiguation: Time.Compatible{}) // 2026-11-01T01:30:00-04:00[America/New_York]
```

On March 8, 2026, the clocks jump from 02:00 to 03:00, so 02:30 never happens. `Time.Earlier` gives
01:30 at the winter offset, and `Time.Later` and `Time.Compatible` give 03:30 at the summer offset:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
Time.toZoned(skipped, ny, disambiguation: Time.Earlier{})    // 2026-03-08T01:30:00-05:00[America/New_York]
Time.toZoned(skipped, ny, disambiguation: Time.Later{})      // 2026-03-08T03:30:00-04:00[America/New_York]
Time.toZoned(skipped, ny, disambiguation: Time.Compatible{}) // 2026-03-08T03:30:00-04:00[America/New_York]
```

## Scope

A value of the union renders with its member's name, and compares equal to another of the same
member:

```ant
let d: Time.Disambiguation = Time.Later{}
d                                // Time.Later{}
d is Time.Later                  // true
Time.Earlier{} == Time.Earlier{} // true
String.from(Time.Compatible{})   // "Time.Compatible{}"
```

Test the member with `is` to branch on it:

```ant
func describe(d: Time.Disambiguation): string {
  if d is Time.Earlier {
    "the earlier instant"
  } else if d is Time.Later {
    "the later instant"
  } else {
    "later in a gap, earlier in an overlap"
  }
}

println(describe(Time.Earlier{}))
println(describe(Time.Compatible{}))
```

```text
the earlier instant
later in a gap, earlier in an overlap
```

Keep the members in an array and try each one:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
let all: Array[Time.Disambiguation] = [Time.Compatible{}, Time.Earlier{}, Time.Later{}]
for d in all {
  println("${d}: ${skipped.toZoned(ny, disambiguation: d)}")
}
```

```text
Time.Compatible{}: 2026-03-08T03:30:00-04:00[America/New_York]
Time.Earlier{}: 2026-03-08T01:30:00-05:00[America/New_York]
Time.Later{}: 2026-03-08T03:30:00-04:00[America/New_York]
```

`Time.parseZoned` resolves a string with a time and no offset by it:

```ant
Time.parseZoned("2026-11-01T01:30[America/New_York]", disambiguation: Time.Later{})
// Ok{ value: 2026-11-01T01:30:00-05:00[America/New_York] }
```

`Time.with` uses it when the new wall-clock time is skipped:

```ant
let saturday = Time.parseZoned("2026-03-07T02:30[America/New_York]").unwrap()
saturday.with(day: 8, disambiguation: Time.Earlier{}) // 2026-03-08T01:30:00-05:00[America/New_York]
saturday.with(day: 8)                                 // 2026-03-08T03:30:00-04:00[America/New_York]
```

A zone that changes by half an hour, such as Lord Howe Island, has a half-hour gap:

```ant
let lordHowe = Time.zone("Australia/Lord_Howe").unwrap()
let skipped = Time.parsePlainDateTime("2026-10-04T02:15").unwrap()
skipped.toZoned(lordHowe, disambiguation: Time.Earlier{}) // 2026-10-04T01:45:00+10:30[Australia/Lord_Howe]
skipped.toZoned(lordHowe, disambiguation: Time.Later{})   // 2026-10-04T02:45:00+11:00[Australia/Lord_Howe]
```

## Applications

Read the choice from a configuration string:

```ant
func parseDisambiguation(s: string): Result[Time.Disambiguation, Error] {
  if s == "earlier" { return Result.ok(Time.Earlier{}) }
  if s == "later" { return Result.ok(Time.Later{}) }
  if s == "compatible" { return Result.ok(Time.Compatible{}) }
  Result.fail("unknown disambiguation ${s}")
}

println(parseDisambiguation("later"))
println(parseDisambiguation("first"))
```

```text
Ok{ value: Time.Later{} }
Err{ error: Error{ message: "unknown disambiguation first" } }
```

A night shift that ends at 01:30 on the night the clocks go back lasts four or five hours, depending
on which 01:30 the worker left at:

```ant
let ny = Time.zone("America/New_York").unwrap()
let start = Time.parsePlainDateTime("2026-10-31T21:30").unwrap().toZoned(ny)
let end = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
for d in [Time.Earlier{}, Time.Later{}] {
  println(start.until(end.toZoned(ny, disambiguation: d)))
}
```

```text
PT4H
PT5H
```

## Properties and relations

In an overlap, `Time.Earlier` and `Time.Later` take the earlier and the later of the two instants
[Time.possibleInstants](../possibleInstants.md) lists, and `Time.Compatible` takes the earlier:

```ant
let ny = Time.zone("America/New_York").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
let both = repeated.possibleInstants(ny)
let early = repeated.toZoned(ny, disambiguation: Time.Earlier{})
let late = repeated.toZoned(ny, disambiguation: Time.Later{})
assert(both.length() == 2)
assert(both.some({ x == early.toInstant() }))
assert(both.some({ x == late.toInstant() }))
assert(Time.compare(early, late) == -1)
assert(repeated.toZoned(ny) == early)
```

In a gap, `Time.Compatible` agrees with `Time.Later`:

```ant
let ny = Time.zone("America/New_York").unwrap()
let skipped = Time.parsePlainDateTime("2026-03-08T02:30").unwrap()
assert(skipped.possibleInstants(ny).length() == 0)
assert(skipped.toZoned(ny) == skipped.toZoned(ny, disambiguation: Time.Later{}))
```

## Possible issues

No disambiguation rejects a skipped or repeated time; each gives some instant. To refuse one, test
the length of `Time.possibleInstants` instead:

```ant
func strict(t: Time.PlainDateTime, z: Time.Zone): Result[Time.ZonedDateTime, Error] {
  if t.possibleInstants(z).length() != 1 {
    return Result.fail("${t} is skipped or repeated in ${z}")
  }
  Result.ok(t.toZoned(z))
}

let ny = Time.zone("America/New_York").unwrap()
println(strict(Time.parsePlainDateTime("2026-11-01T01:30").unwrap(), ny))
println(strict(Time.parsePlainDateTime("2026-11-01T12:00").unwrap(), ny))
```

```text
Err{ error: Error{ message: "2026-11-01T01:30:00 is skipped or repeated in America/New_York" } }
Ok{ value: 2026-11-01T12:00:00-05:00[America/New_York] }
```

A member is a type; write braces to make a value of it:

```ant
let ny = Time.zone("America/New_York").unwrap()
let repeated = Time.parsePlainDateTime("2026-11-01T01:30").unwrap()
let at = repeated.toZoned(ny, disambiguation: Time.Later)
// error: Time.Later is a type, not a value; a type appears only as a literal head, after is, or before a dot
```

A string with an offset names its instant, so `disambiguation` has no say:

```ant
Time.parseZoned("2026-11-01T01:30-04:00[America/New_York]", disambiguation: Time.Later{})
// Ok{ value: 2026-11-01T01:30:00-04:00[America/New_York] }
```

`Time.with` keeps the value's own offset when it fits the new time, so moving to the repeated hour
from the summer side stays on the summer offset, whatever the disambiguation:

```ant
let october = Time.parseZoned("2026-10-31T01:30[America/New_York]").unwrap()
october.with(month: 11, day: 1, disambiguation: Time.Later{}) // 2026-11-01T01:30:00-04:00[America/New_York]
```

`Time.add` has no `disambiguation` parameter; adding a day to a time that lands in a gap always
moves forward, as `Time.Compatible` does:

```ant
let saturday = Time.parseZoned("2026-03-07T02:30[America/New_York]").unwrap()
let sunday = Time.add(saturday, Time.duration(days: 1), disambiguation: Time.Earlier{})
// error: Time.add has no parameter disambiguation
```

## See also

[Time.toZoned](../toZoned.md) · [Time.possibleInstants](../possibleInstants.md) ·
[Time.parseZoned](../parseZoned.md) · [Time.with](../with.md) · [Time.add](../add.md) ·
[Time.nextTransition](../nextTransition.md) · [Time.ZonedDateTime](ZonedDateTime.md)
