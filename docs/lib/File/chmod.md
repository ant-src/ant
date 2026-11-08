<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.chmod

```ant signature
func File.chmod(path: string, mode: int): Result[void, Error]
```

File.chmod sets the mode of the file at path to mode exactly, following a symbolic link. It is Err
for a mode outside 0 ..= 0o7777. On Windows only 0o200 counts: without it the file becomes
read-only.

<!-- end generated -->
