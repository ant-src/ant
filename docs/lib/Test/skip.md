<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test.skip

```ant signature
func Test.skip(reason: string): never
```

Test.skip ends the current test as skipped, after its deferred blocks run. It is not a panic, so
Panic.catch does not stop it. A test that failed before stays failed.

<!-- end generated -->
