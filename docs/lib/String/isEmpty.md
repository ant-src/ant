<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.isEmpty

```ant signature
func String.isEmpty(s: string): bool
```

String.isEmpty reports whether s is "".

<!-- end generated -->

## Details

- `String.isEmpty` returns `true` when `s` is `""`, and `false` for any other string.
- A string of white space, or of one invisible code point, is not empty.
- `len(s)` and `String.length(s)` count the code points of `s`
  ([spec: Prelude](https://github.com/ant-src/spec/blob/master/ant-lang.md#prelude)).

## Basic examples

```ant
String.isEmpty("")    // true
String.isEmpty("ant") // false
```

Call it through `.` on any string:

```ant
var name = "Ann"
name.isEmpty() // false
```

## Scope

White space is text. Only `""` is empty:

```ant
" ".isEmpty()  // false
"\n".isEmpty() // false
```

A code point that shows nothing, such as a zero width space or a lone combining accent, still
counts:

```ant
"\u{200B}".isEmpty() // false
"\u{301}".isEmpty()  // false
```

In a pipeline, the string is the first argument:

```ant
"" | String.isEmpty() // true
```

As a function value, it maps an array of strings to an array of `bool`s:

```ant
Array.map(["", "a", ""], String.isEmpty) // [true, false, true]
```

For a string that may be `null`, `?.` yields `null` instead of calling:

```ant
var nickname: string | null = null
nickname?.isEmpty()         // null
nickname?.isEmpty() ?? true // true
```

## Named arguments

The parameter is named `s`:

```ant
String.isEmpty(s: "x") // false
```

## Applications

Test what `String.slice` or `String.trim` returns:

```ant
String.slice("abc", 1, 1).isEmpty() // true
"  \t\n".trim().isEmpty()           // true
```

Drop the empty fields that a doubled or trailing separator leaves:

```ant
String.split("a,,b,", ",")                          // ["a", "", "b", ""]
String.split("a,,b,", ",").filter({ !x.isEmpty() }) // ["a", "b"]
```

Fall back to a default when a name is blank:

```ant
func greet(name: string): string {
  var trimmed = name.trim()
  if trimmed.isEmpty() {
    "Hello, stranger"
  } else {
    "Hello, ${trimmed}"
  }
}

println(greet("Ann"))
println(greet("   "))
```

```text
Hello, Ann
Hello, stranger
```

Reject empty input with a `Result`:

```ant
type EmptyInput = {}

func parseName(s: string): Result[string, EmptyInput] {
  if String.isEmpty(s) {
    return Result.err(EmptyInput{})
  }
  Result.ok(s)
}

println(parseName("Ann"))
println(parseName(""))
```

```text
Ok{ value: "Ann" }
Err{ error: EmptyInput{} }
```

## Properties and relations

`String.isEmpty` agrees with `len`, `String.length`, `==` against `""`, and
[`String.chars`](chars.md):

```ant
for s in ["", "a", " ", "日本", "👨‍👩‍👧"] {
  assert(s.isEmpty() == (len(s) == 0))
  assert(s.isEmpty() == (String.length(s) == 0))
  assert(s.isEmpty() == (s == ""))
  assert(s.isEmpty() == s.chars().isEmpty())
}
```

The concatenation of two strings is empty exactly when both are:

```ant
for a in ["", "x"] {
  for b in ["", "y"] {
    assert((a + b).isEmpty() == (a.isEmpty() && b.isEmpty()))
  }
}
```

## Possible issues

A string is not a condition. Ask `String.isEmpty` explicitly:

<!-- test: error cannot use string where bool is expected -->

```ant
var s = ""
if s {
  println("has text")
}
```

A blank string is not empty. Trim it first to treat white space as nothing:

```ant
var input = "   "
input.isEmpty()        // false
input.trim().isEmpty() // true
```

`String.isSpace` does not treat `""` as blank either:

```ant
"".isSpace()   // false
"  ".isSpace() // true
```

`trim().isEmpty()` covers both the empty and the blank string.

A nullable string, such as the result of `Map` indexing, is not a `string`, so passing it is a
compile error:

<!-- test: error cannot use string | null where string is expected -->

```ant
var user = Map{ name: "" }
println(String.isEmpty(user["name"]))
```

Narrow it first:

```ant
var user = Map{ name: "" }
var v = user["name"]
if v != null {
  println(v.isEmpty())
}
```

```text
true
```

Or give a default with `??`:

```ant
var user = Map{ name: "" }
(user["name"] ?? "").isEmpty() // true
```

A named primitive over `string` is not a `string`. Unwrap it with `string(...)`:

<!-- test: error cannot use Name where string is expected -->

```ant
type Name = string
var n = Name("")
println(String.isEmpty(n))
```

```ant
type Name = string
var n = Name("")
String.isEmpty(string(n)) // true
```

## See also

[String.length](length.md) · [String.trim](trim.md) · [String.isSpace](isSpace.md) ·
[String.chars](chars.md) · [Array.isEmpty](../Array/isEmpty.md) · [Map.isEmpty](../Map/isEmpty.md) ·
[Set.isEmpty](../Set/isEmpty.md)
