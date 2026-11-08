<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.Info

```ant signature
type File.Info = { name: string, size: int, isDir: bool, isSymlink: bool, modified: Time.Instant, mode: int }
```

File.Info describes the file at a path, as [`File.stat`](../stat.md) reports it. modified is the
time of the last write. mode holds the permission bits and the setuid, setgid and sticky bits, at
most 0o7777; on Windows a file is 0o444 when read-only and 0o666 when not, and a directory adds
0o111. isSymlink is true only when stat was told not to follow a link.

<!-- end generated -->
