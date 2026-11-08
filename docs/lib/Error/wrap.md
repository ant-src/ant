<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Error.wrap

```ant signature
func Error.wrap(e: Error, context: string): Error
```

Error.wrap returns a new Error whose message is context, ": ", and the message of e, leaving e
untouched. An empty context is joined like any other.

<!-- end generated -->

## Basic examples

```ant
var e = Error{ message: "no such file" }
Error.wrap(e, "load config") // Error{ message: "load config: no such file" }
```
