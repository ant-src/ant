<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test

Test holds what a test body needs beyond assert, Panic.catch, and defer (ant test). Each function
acts on the current test: the one whose body the calling goroutine runs, or that owns the goroutine.
Called outside any test, each panics.

<!-- end generated -->
