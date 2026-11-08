<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Net.listen

```ant signature
func Net.listen(addr: string): Result[Net.Listener, Error]
```

Net.listen listens for TCP connections on addr, written "host:port". A port of 0, as in ":0", takes
any free port, which [`Net.address`](address.md) reports.

<!-- end generated -->
