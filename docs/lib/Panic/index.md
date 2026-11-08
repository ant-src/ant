<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Panic

```ant signature
type Panic = { message: string, trace: Array[Panic.Frame] }
```

Panic is a panic that [Panic.catch](catch.md) stopped: its message and the goroutine's stack when it
began, innermost frame first.

<!-- end generated -->
