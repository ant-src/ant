<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.Request

```ant signature
type Http.Request = { method: string, path: string, query: string, headers: Map[string, string], body: string, bytes: Bytes, remote: string }
```

Http.Request is a request that [Http.serve](../serve.md) passes to its handler. path is decoded,
query is raw for [Url.parseQuery](../../Url/parseQuery.md), and remote is the client's "ip:port".
The body is there twice, as in [Http.Response](Response.md), and the headers hold one string per
name, Host among them.

<!-- end generated -->
