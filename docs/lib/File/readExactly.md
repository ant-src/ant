<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.readExactly

```ant signature
func File.readExactly(f: File.Handle, n: int, ctx: Context | null = null): Result[Bytes | null, Error]
```

File.readExactly reads exactly n bytes from f, waiting until all of them have arrived. It returns
null at end of file, is Err with the message "unexpected end of file" when the input ends partway,
discarding the bytes it read, and is Err when n is below 0. An n of 0 gives empty Bytes without
reading, at the end of the file too.

<!-- end generated -->
