<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.request

```ant signature
func Http.request(method: string, url: string, body: string | Bytes = "", headers: Map[string, string] = Map{}, timeout: Time.Duration | null = null, ctx: Context | null = null): Result[Http.Response, Error]
```

Http.request sends one request with method, body, and headers to url and reads the whole response. A
response of any status is Ok, and no redirect is followed; Err means no response: a malformed url,
an empty method, header names that differ only in case, no connection, the timeout ("timeout"), or
ctx done. timeout bounds the whole call and is no deadline when null; one that is not positive or
has years, months, or weeks is Err; when both timeout and ctx are given, the earlier wins.

<!-- end generated -->
