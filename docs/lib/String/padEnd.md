<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.padEnd

```ant signature
func String.padEnd(s: string, width: int, pad: string = " "): string
```

String.padEnd returns s with pad repeated on the right until it is width code points long. It pads
and panics as [String.padStart](padStart.md) does.

<!-- end generated -->

## Basic examples

```ant
String.padEnd("7", 4, "ab") // "7aba"
```
