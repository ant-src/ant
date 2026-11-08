<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.yearOfWeek

```ant signature
func Time.yearOfWeek(t: Time.Dated): int
```

Time.yearOfWeek returns the year that the ISO 8601 week of t belongs to, which differs from
[Time.year](year.md) around New Year.

<!-- end generated -->

## Basic examples

```ant
Time.yearOfWeek(Time.plainDate(2027, 1, 1)) // 2026
```
