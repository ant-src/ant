<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.cacheDir

```ant signature
func Os.cacheDir(): Result[string, Error]
```

Os.cacheDir returns the directory where the user's programs keep what can be made again:
$XDG_CACHE_HOME or ~/.cache on Unix, ~/Library/Caches on macOS, %LocalAppData% on Windows.

<!-- end generated -->
