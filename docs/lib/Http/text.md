<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Http.text

```ant signature
func Http.text(r: Http.Response | Http.Request): string
```

Http.text returns the body of r decoded as UTF-8, with bytes that are not valid UTF-8 as U+FFFD, as all text from outside is. It decodes on every call, so a program that needs the text twice binds it once.

<!-- end generated -->
