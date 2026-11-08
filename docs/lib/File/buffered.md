<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.buffered

```ant signature
func File.buffered(f: File.Handle, size: int = 4096): File.Handle
```

File.buffered returns a handle over f whose writes collect in a buffer of size bytes, written to f
when a write does not fit, at [File.flush](flush.md), and at [File.close](close.md), and at nothing
else: a buffered handle neither flushed nor closed when the run ends loses what it holds. It is for
a loop of many small writes. f itself is returned when it is buffered with at least size bytes; a
size below 1 gives the default. Reads of the handle read f. The first error writing the buffer out
is kept and returned by every later write and flush, and by the first close. A write to f itself
goes past the buffer.

<!-- end generated -->
