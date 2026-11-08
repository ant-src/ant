<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.program

```ant signature
func Os.program(): string
```

Os.program returns the path of the program's root module as the command was given it, as a shell's
$0 is: ./tool.ant when a \#! line runs it, tool.ant after ant run tool.ant. It is "" in the REPL.

<!-- end generated -->
