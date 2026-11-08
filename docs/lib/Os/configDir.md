<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.configDir

```ant signature
func Os.configDir(): Result[string, Error]
```

Os.configDir returns the directory where the user's programs keep their configuration:
$XDG_CONFIG_HOME or ~/.config on Unix, ~/Library/Application Support on macOS, %AppData% on Windows.
A program keeps its files under a directory of its own name there.

<!-- end generated -->
