<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context

```ant signature
type Context
```

Context bounds a wait with a deadline or ends it with a cancellation. A context becomes done once:
when it is cancelled, when its deadline passes, or when its parent is done. Every blocking library
call takes one as its last parameter. Contexts compare by identity.

<!-- end generated -->
