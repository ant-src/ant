<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.compare

```ant signature
func Time.compare[T: Time.Point](a: T, b: T): int
```

Time.compare returns -1, 0 or 1 as a is before, at or after b. Instants and ZonedDateTimes compare
by instant, so two zoned values at one moment in different zones compare as 0 though they are not
==; the plain types compare by fields. The comparison operators do not apply to time values.

<!-- end generated -->
