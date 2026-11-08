<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.post

```ant signature
func Http.post(url: string, body: string | Bytes, headers: Map[string, string] = Map{}, timeout: Time.Duration | null = null, ctx: Context | null = null): Result[Http.Response, Error]
```

Http.post sends body to url in a POST request and reads the whole response. It is
[Http.request](request.md) with the method "POST".

<!-- end generated -->
