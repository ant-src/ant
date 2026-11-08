<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.possibleInstants

```ant signature
func Time.possibleInstants(t: Time.PlainDateTime, z: Time.Zone): Array[Time.Instant]
```

Time.possibleInstants returns the instants at which zone z shows the wall-clock time t, earliest
first. There are none in a daylight-saving gap, two in an overlap and one otherwise, so a program
that refuses ambiguous times tests the length.

<!-- end generated -->
