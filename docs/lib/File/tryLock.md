<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# File.tryLock

```ant signature
func File.tryLock(path: string): Result[File.Handle | null, Error]
```

File.tryLock is [`File.lock`](lock.md) without the wait: it is Ok(null) when another handle holds
the lock.

<!-- end generated -->
