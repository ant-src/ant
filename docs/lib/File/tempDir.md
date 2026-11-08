<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.tempDir

```ant signature
func File.tempDir(pattern: string, dir: string | null = null): Result[string, Error]
```

File.tempDir creates a new empty directory, mode 0o700, in dir or the host's temporary directory,
and returns its path. Its name is pattern with the last \* replaced by a random string, or with the
string appended when there is no \*. Nothing removes it for the program.

<!-- end generated -->
