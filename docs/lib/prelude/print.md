<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# print

```ant signature
func print[T](value: T): void
```

print writes String.from(value) to standard output, with no newline. Output is not buffered, so it
comes out in program order. A failed write is ignored, except a broken pipe, which ends the program
with status 141 as SIGPIPE does; File.writeLine(Os.stdout(), …) reports other errors.

<!-- end generated -->
