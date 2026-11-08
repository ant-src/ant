<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.Reply

```ant signature
type Http.Reply = { status: int, headers: Map[string, string], body: string | Bytes }
```

Http.Reply is what an [Http.serve](../serve.md) handler returns to answer a request. body is sent as
given, a string as its UTF-8 and Bytes as they are, with Content-Length set from it; a status
outside 200 to 599, or two header names that differ only in case, panics.

<!-- end generated -->
