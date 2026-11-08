<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.copyAll

```ant signature
func File.copyAll(from: string, to: string): Result[void, Error]
```

File.copyAll copies the file or the whole tree at from to to, keeping modes and copying the links
inside the tree as links. to must not exist or, for a tree, be an empty directory, which keeps its
own mode; anything else is Err before anything is copied, as is a to that lies inside from. An error
partway through leaves what was copied.

<!-- end generated -->
