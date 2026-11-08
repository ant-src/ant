<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.serve

```ant signature
func Http.serve(l: Net.Listener, handler: func(Http.Request): Http.Reply, ctx: Context | null = null): Result[void, Error]
```

Http.serve answers every request that arrives on l by calling handler on a goroutine of its own. It
returns Ok once ctx is done and the requests in flight have finished, waiting at most five seconds
for them, or serves until the run is cancelled when ctx is null; it closes l on return and is Err
when l is closed under it. A panic in handler sends status 500, and a reply status outside 200 to
599, or reply header names that differ only in case, panic in the handler's goroutine.

<!-- end generated -->
