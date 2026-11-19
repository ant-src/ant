<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.chown

```ant signature
func File.chown(path: string, uid: int, gid: int, follow: bool = true): Result[void, Error]
```

File.chown sets the user and the group that own the file at path to the ids uid and gid; -1 keeps
one as it is. It follows a symbolic link, unless follow is false, when it changes the link itself.
Only root may give a file away. It is Err on Windows.

<!-- end generated -->
