<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Json.stringify

```ant signature
func Json.stringify[T](value: T, indent: string = ""): string
```

Json.stringify returns the compact JSON text of value. Records become objects without their tag,
sets arrays, and time values strings. It panics with "cyclic value" on a value that contains itself,
and panics on what has no JSON form: NaN and the infinities, Bytes, which a program writes with
[`Encoding.toBase64`](../Encoding/toBase64.md) instead, a function, void, a type, or a reference
such as a channel or a Random. A map key must be a string or a named primitive over one, which is
written without its tag; any other key panics, as do two keys that give one property name.

With an empty indent the text is one line. Otherwise each member of an array or object is on a line
of its own, indented by indent once per level, ": " follows a property name, and an empty array or
object stays \[\] or \{\}; the text ends without a newline. It panics on an indent that is not all
spaces and tabs.

<!-- end generated -->

## Basic examples

```ant
Json.stringify(Map{ "a": [1, 2] })            // "{\"a\":[1,2]}"
Json.stringify(Map{ "a": [1] }, indent: "  ") // "{\n  \"a\": [\n    1\n  ]\n}"
```
