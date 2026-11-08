<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.symlink

```ant signature
func File.symlink(target: string, link: string): Result[void, Error]
```

File.symlink makes a symbolic link at link whose target is target, in the order of ln -s. A relative
target is relative to the link's directory. It is Err when anything is at link.

<!-- end generated -->
