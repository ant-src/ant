<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.compile

```ant signature
func Regex.compile(source: string, flags: string = ""): Result[Regex.Pattern, Error]
```

Regex.compile compiles source, in RE2 syntax, with flags, some of i, m, and s in any order. It
returns Err on bad syntax, an unknown flag, or a repeated one. flags defaults to "".

<!-- end generated -->
