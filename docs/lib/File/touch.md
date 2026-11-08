<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.touch

```ant signature
func File.touch(path: string, time: Time.Instant | null = null): Result[void, Error]
```

File.touch sets the modified and access times of the file at path to time, or to now when time is
null, creating an empty file when nothing is there. It never truncates, and it follows a link.

<!-- end generated -->
