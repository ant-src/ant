<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.stdout

```ant signature
func Os.stdout(): File.Handle
```

Os.stdout returns the program's standard output, which [print](../prelude/print.md) and
[println](../prelude/println.md) write to as well. It is not buffered; File.buffered(Os.stdout())
buffers a program that writes much.

<!-- end generated -->
