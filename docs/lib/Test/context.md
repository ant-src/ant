<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test.context

```ant signature
func Test.context(): Context
```

Test.context returns the current test's context, done when its body ends, before its deferred
blocks. Give it to blocking calls that must stop when the test does.

<!-- end generated -->
