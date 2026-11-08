<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.inScript

```ant signature
func String.inScript(c: string, script: string): bool
```

String.inScript reports whether c has the Unicode Script property script, named in full as in
Scripts.txt, such as "Latin" or "Common". It reads Script, not Script_Extensions. It panics unless c
is one code point, or if script names no script.

<!-- end generated -->
