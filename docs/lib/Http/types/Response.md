<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.Response

```ant signature
type Http.Response = { status: int, headers: Map[string, string], body: string, bytes: Bytes }
```

Http.Response is the response to a client request. Its body is there twice: as received in bytes,
and as text in body, with bytes that are not valid UTF-8 as U+FFFD. The headers hold one string per
name, a repeated header joined with ", ".

<!-- end generated -->
