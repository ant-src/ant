<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.remove

```ant signature
func File.remove(path: string): Result[void, Error]
```

File.remove removes the file or empty directory at path. It is Err on a directory that has entries;
[`File.removeAll`](removeAll.md) removes a whole tree.

<!-- end generated -->
