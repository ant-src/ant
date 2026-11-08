<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.openPipe

```ant signature
func Os.openPipe(): Result[Os.Pipe, Error]
```

Os.openPipe returns a pipe, two handles the program closes: what is written to write is read from
read. A read waits for data. The reading end is at end of file once every writing end is closed, the
copies given to children and the program's own.

<!-- end generated -->
