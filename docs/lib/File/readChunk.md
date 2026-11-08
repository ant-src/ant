<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readChunk

```ant signature
func File.readChunk(f: File.Handle, max: int, ctx: Context | null = null): Result[Bytes | null, Error]
```

File.readChunk reads at least 1 and at most max bytes from f, returning as soon as there are any. It
returns null at end of file and is Err when max is below 1. It is the read for a stream whose pieces
have no fixed size.

<!-- end generated -->
