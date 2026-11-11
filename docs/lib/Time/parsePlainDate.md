<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.parsePlainDate

```ant signature
func Time.parsePlainDate(s: string): Result[Time.PlainDate, Error]
```

Time.parsePlainDate parses a date from s, such as 2026-10-22. It takes the date of a longer string
and ignores the rest, though not of one with Z and no bracketed zone; text it cannot parse, or a
date that does not exist, is an Err.

<!-- end generated -->

## Details

- The date is the one written in the string. An offset or a zone does not move it, so
  `2026-10-22T23:30:00-05:00` gives `2026-10-22`, not the date in UTC.
- A date that does not exist, such as `2026-02-30`, is an `Err`, where `Time.plainDate` panics on
  it.
- A longer string must still have one of the accepted forms after the date, so a bad time or a stray
  word makes it an `Err`.

## Basic examples

```ant
Time.parsePlainDate("2026-10-22") // Ok{ value: 2026-10-22 }
```

A full zoned string gives its date:

```ant
let s = "2026-10-22T16:03:05+02:00[Europe/Stockholm]"
Time.parsePlainDate(s) // Ok{ value: 2026-10-22 }
```

Text that is not a date is an `Err`:

```ant
Time.parsePlainDate("22/10/2026")
// Err{ error: Error{ message: "not a date or time: \"22/10/2026\"" } }
```

## Scope

A date and time, with `T` or a space between them:

```ant
Time.parsePlainDate("2026-10-22T16:03:05.5") // Ok{ value: 2026-10-22 }
Time.parsePlainDate("2026-10-22 16:03")      // Ok{ value: 2026-10-22 }
```

An offset, or `Z` together with a bracketed zone:

```ant
Time.parsePlainDate("2026-10-22T16:03:05+02:00") // Ok{ value: 2026-10-22 }
Time.parsePlainDate("2026-10-22T14:03:05Z[UTC]") // Ok{ value: 2026-10-22 }
```

The calendar annotation `[u-ca=iso8601]` is accepted; any other calendar is an `Err`:

```ant
Time.parsePlainDate("2026-10-22[u-ca=iso8601]") // Ok{ value: 2026-10-22 }
```

```ant
Time.parsePlainDate("2026-10-22[u-ca=gregory]")
// Err{ error: Error{ message: "calendar gregory is not supported; only iso8601 is" } }
```

A year outside 0 to 9999 takes a sign and six digits, the form such a date renders in:

```ant
Time.parsePlainDate("+012026-10-22") // Ok{ value: +012026-10-22 }
Time.parsePlainDate("-000001-02-28") // Ok{ value: -000001-02-28 }
```

A date the calendar lacks is an `Err`, not a panic:

```ant
Time.parsePlainDate("2024-02-29") // Ok{ value: 2024-02-29 }
Time.parsePlainDate("2026-02-29") // Err{ error: Error{ message: "no such date: 2026-02-29" } }
Time.parsePlainDate("2026-13-01") // Err{ error: Error{ message: "no such date: 2026-13-01" } }
```

Pass it as a function value to parse many strings:

```ant
["2026-03-01", "2026-01-15"].map(Time.parsePlainDate)
// [Ok{ value: 2026-03-01 }, Ok{ value: 2026-01-15 }]
```

`Result.all` turns those into one `Result` for the whole list, the first `Err` if any:

```ant
["2026-03-01", "2026-01-15"].map(Time.parsePlainDate) | Result.all()
// Ok{ value: [2026-03-01, 2026-01-15] }
["2026-03-01", "2026-02-30"].map(Time.parsePlainDate) | Result.all()
// Err{ error: Error{ message: "no such date: 2026-02-30" } }
```

In a pipeline, the result feeds the `Result` functions:

```ant
Time.parsePlainDate("2026-10-22") | Result.map(Time.dayOfWeek) // Ok{ value: 4 }
```

## Named arguments

The parameter is named `s`:

```ant
Time.parsePlainDate(s: "2026-10-22") // Ok{ value: 2026-10-22 }
```

## Applications

Check a date typed into a form, and say what is wrong with it:

```ant
func checkBirthday(text: string): string {
  let parsed = Time.parsePlainDate(text.trim())
  if parsed is Err { return "invalid: ${parsed.error.message}" }
  "born on day ${parsed.value.dayOfYear()} of ${parsed.value.year()}"
}

println(checkBirthday(" 1990-07-15 "))
println(checkBirthday("1990-02-30"))
```

```text
born on day 196 of 1990
invalid: no such date: 1990-02-30
```

Parse a list of dates, stopping at the first bad one, with its line number in the error:

```ant
func parseDates(lines: Array[string]): Result[Array[Time.PlainDate], Error] {
  var dates: Array[Time.PlainDate] = []
  for { index, value } in lines.entries() {
    dates.push(try (Time.parsePlainDate(value) | Result.wrapErr("line ${index + 1}")))
  }
  Result.ok(dates)
}

println(parseDates(["2026-01-05", "2026-01-12"]))
println(parseDates(["2026-01-05", "2026-02-30"]))
```

```text
Ok{ value: [2026-01-05, 2026-01-12] }
Err{ error: Error{ message: "line 2: no such date: 2026-02-30" } }
```

Fall back to a default date when the text does not parse:

```ant
Time.parsePlainDate("soon") | Result.unwrapOr(Time.plainDate(2026, 1, 1)) // 2026-01-01
```

Count the days to a deadline read from text:

```ant
let today = Time.plainDate(2026, 11, 11)
let deadline = Result.unwrap(Time.parsePlainDate("2026-12-24"))
Time.until(today, deadline) // P43D
```

Group log timestamps by the date they were written on:

```ant
let stamps = [
  "2026-10-22T09:15:00+02:00",
  "2026-10-22T23:50:00+02:00",
  "2026-10-23T00:10:00+02:00",
]
for { key, value } in stamps.groupBy({ Result.unwrap(Time.parsePlainDate(x)) }) {
  println("${key}: ${value.length()}")
}
```

```text
2026-10-22: 2
2026-10-23: 1
```

## Properties and relations

A date renders as text that parses back to the same date:

```ant
for d in [Time.plainDate(2026, 10, 22), Time.plainDate(0, 2, 29), Time.plainDate(-1, 1, 1)] {
  assert(Result.unwrap(Time.parsePlainDate("${d}")) == d)
}
```

On a date and time, it agrees with `Time.parsePlainDateTime` followed by `Time.toPlainDate`:

```ant
for s in ["2026-10-22T16:03:05", "2026-10-22 00:00", "2026-10-22T23:59:59.999+02:00"] {
  let viaDateTime = Time.parsePlainDateTime(s) | Result.map(Time.toPlainDate)
  assert(Time.parsePlainDate(s) == viaDateTime)
}
```

## Possible issues

The offset does not convert the date. A time late in the evening west of UTC keeps its own date:

```ant
Time.parsePlainDate("2026-10-22T23:30:00-05:00") // Ok{ value: 2026-10-22 }
```

To get the date in another zone, parse the instant and place it in that zone:

```ant
let at = Result.unwrap(Time.parseInstant("2026-10-22T23:30:00-05:00"))
at.inZone(Time.utc).toPlainDate() // 2026-10-23
```

`Z` alone names a moment, not a wall-clock date, so it is an `Err`. Parse such a string with
`Time.parseInstant`:

```ant
Time.parsePlainDate("2026-10-22T14:03:05Z")
// Err{ error: Error{ message: "Z names an exact time; a plain value needs a wall-clock time" } }
```

The rest of a longer string is checked, not skipped, so a bad time makes the whole string an `Err`:

```ant
Time.parsePlainDate("2026-10-22T25:00") // Err{ error: Error{ message: "hour out of range: 25" } }
```

```ant
Time.parsePlainDate("2026-10-22 trailing")
// Err{ error: Error{ message: "not a date or time: \"2026-10-22 trailing\"" } }
```

Only the ISO form parses. Leading spaces, single-digit months and days, and compact dates are all
`Err`; trim the text first when it may carry spaces:

```ant
Time.parsePlainDate(" 2026-10-22").isOk()        // false
Time.parsePlainDate(" 2026-10-22".trim()).isOk() // true
Time.parsePlainDate("2026-1-5").isOk()           // false
Time.parsePlainDate("20261022").isOk()           // false
```

`try` binds tighter than `|`, so `try Time.parsePlainDate(s) | Result.wrapErr(...)` pipes the date,
not the `Result`. Parenthesize the pipeline:

```ant
func load(s: string): Result[Time.PlainDate, Error] {
  let d = try Time.parsePlainDate(s) | Result.wrapErr("start")
  // error: cannot use Time.PlainDate where Result[T, Error] is expected
  Result.ok(d)
}
```

## See also

[Time.plainDate](plainDate.md) · [Time.parsePlainDateTime](parsePlainDateTime.md) ·
[Time.parsePlainTime](parsePlainTime.md) · [Time.parseZoned](parseZoned.md) ·
[Time.parseInstant](parseInstant.md) · [Time.toPlainDate](toPlainDate.md) ·
[Time.PlainDate](types/PlainDate.md)
