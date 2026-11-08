<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.plainTime

```ant signature
func Time.plainTime(hour: int = 0, minute: int = 0, second: int = 0, millisecond: int = 0, microsecond: int = 0, nanosecond: int = 0): Time.PlainTime
```

Time.plainTime returns the time of day with the given fields, each 0 when omitted. It panics if a
field is out of range, such as hour 24 or minute 60.

<!-- end generated -->
