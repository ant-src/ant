<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.weekOfYear

```ant signature
func Time.weekOfYear(t: Time.Dated): int
```

Time.weekOfYear returns the ISO 8601 week of t, 1 to 53. Weeks start on Monday and week 1 holds the
year's first Thursday, so days around New Year can fall in a week of the neighbouring year;
[Time.yearOfWeek](yearOfWeek.md) gives that year.

<!-- end generated -->

## Basic examples

```ant
Time.weekOfYear(Time.plainDate(2027, 1, 1)) // 53
```
