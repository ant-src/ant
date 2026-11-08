<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Error

```ant signature
type Error = { message: string }
```

Error is the error the standard library returns in an [Err](../Result/index.md): a message for a
person to read. [Error.wrap](wrap.md) and [Result.wrapErr](../Result/wrapErr.md) add context in
front of the message.

<!-- end generated -->
