<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Url.stringify

```ant signature
func Url.stringify(parts: Url.Parts): string
```

Url.stringify returns the reference parts make, escaping what needs it in every part but query,
which is written as it is, so a program changes a URL by parsing it, replacing a field, and
stringifying it. It writes // before an empty authority only for http, https, ws, wss, ftp and file,
so x:///a comes back as x:/a, the same reference.

<!-- end generated -->
