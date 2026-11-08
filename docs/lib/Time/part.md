<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.part

```ant signature
func Time.part(d: Time.Duration, unit: Time.Unit): int
```

Time.part returns the field of d for unit, as given and not balanced.

<!-- end generated -->

## Basic examples

```ant
Time.part(Time.duration(minutes: 90), Time.Hour{}) // 0
```
