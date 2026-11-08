<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.plainDate

```ant signature
func Time.plainDate(year: int, month: int, day: int): Time.PlainDate
```

Time.plainDate returns the date with the given year, month and day. It panics if there is no such
date, such as month 13 or February 30; check input from the world against
[Time.daysInMonth](daysInMonth.md) first, or hand its text to
[Time.parsePlainDate](parsePlainDate.md).

<!-- end generated -->
