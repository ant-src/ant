<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.Parts

```ant signature
type Url.Parts = { scheme: string, opaque: string, user: string, host: string, port: string, path: string, query: string, fragment: string }
```

Url.Parts is a URL reference divided as RFC 3986 does, with each absent part "". scheme is in lower
case without its colon, opaque is what follows "scheme:" when that does not start with /, as in a
mailto: address, and user is the userinfo as written. path and fragment are decoded, but path keeps
%2F and %25 as written, while query is the raw text after the ?, for
[Url.parseQuery](../parseQuery.md).

<!-- end generated -->
