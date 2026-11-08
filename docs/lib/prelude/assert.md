<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# assert

```ant signature
func assert(cond: bool, message: string = ""): void
```

assert panics when cond is false, with a message that quotes the condition, shows the values it was
computed from, and ends with message when it is not empty. message is evaluated only then. assert is
not a value: it can only be called by name.

<!-- end generated -->
