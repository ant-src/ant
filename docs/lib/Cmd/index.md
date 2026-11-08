<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cmd

```ant signature
type Cmd = { program: string, args: Array[string] = [], dir: string | null = null, env: Map[string, string] | null = null }
```

Cmd is a program to run, as a value: a Cmd says what to run, and a function of Cmd says how. Nothing
runs it through a shell. program is looked up in PATH when it has no path separator, and is taken
relative to dir otherwise; the child runs in dir and with env as its whole environment when they are
given. A child ended by signal n has the status 128 + n, as a shell reports it.

<!-- end generated -->
