<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.run

```ant signature
func Parallel.run(fs: Array[func(Context): void], ctx: Context | null = null): void
```

Parallel.run calls every function of fs at once with the scope context, or none when ctx is done
already. Its functions return void, unlike other callbacks whose value the library ignores, so that
they need not share a result type; each hands back its result by writing a field.

<!-- end generated -->
