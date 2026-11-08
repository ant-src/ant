<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.subtract

```ant signature
func Time.subtract[T: Time.Point](t: T, d: Time.Duration): T
```

Time.subtract returns t moved back by d, which is Time.add(t, Time.negate(d)). It panics as
[`Time.add`](add.md) does.

<!-- end generated -->
