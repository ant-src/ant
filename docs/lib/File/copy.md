<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.copy

```ant signature
func File.copy(from: string, to: string, replace: bool = true): Result[void, Error]
```

File.copy copies the contents of the file at from to to, and then sets to's mode to from's. It
replaces a file at to unless replace is false, when it is Err if anything is at to; it is Err when
to is a directory, or from is not a regular file, such as a FIFO. It follows a symbolic link at
either end, does not copy times, and is Err before changing anything when from and to are the same
file.

<!-- end generated -->
