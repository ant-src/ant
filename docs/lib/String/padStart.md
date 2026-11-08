<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.padStart

```ant signature
func String.padStart(s: string, width: int, pad: string = " "): string
```

String.padStart returns s with pad repeated on the left until it is width code points long. The
padding is cut to the code points s lacks, and s comes back unchanged when it is already at or
beyond width. pad defaults to a space. It panics if pad is empty or width is above 2^31 - 1.

<!-- end generated -->

## Basic examples

```ant
String.padStart("7", 4, "ab") // "aba7"
```
