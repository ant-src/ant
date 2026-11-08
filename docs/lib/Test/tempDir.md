<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Test.tempDir

```ant signature
func Test.tempDir(): string
```

Test.tempDir returns a new directory, removed with its contents when the current test ends. Panics
if the directory cannot be made.

<!-- end generated -->
