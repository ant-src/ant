<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.tempFile

```ant signature
func File.tempFile(pattern: string, dir: string | null = null): Result[string, Error]
```

File.tempFile creates a new empty file, mode 0o600, in dir or the host's temporary directory, and
returns its path. Its name is pattern with the last * replaced by a random string, or with the
string appended when there is no *. Nothing removes it for the program.

<!-- end generated -->
