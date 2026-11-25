<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Error.Kind

```ant signature
type Error.Kind = Error.NotExist | Error.Exist | Error.Permission
```

Error.Kind is what failed, for the failures of the OS that Go tells apart on every platform. Test it
with is, on the result of the call it is about: a kind survives [`Error.wrap`](../wrap.md).

<!-- end generated -->
