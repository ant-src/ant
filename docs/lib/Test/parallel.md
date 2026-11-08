<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test.parallel

```ant signature
func Test.parallel(): void
```

Test.parallel runs the current test alongside the other parallel tests of its parent, once the tests
of that parent that did not call it have ended. It is called once, from the test's body.

<!-- end generated -->
