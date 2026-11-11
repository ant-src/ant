<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toUpper

```ant signature
func String.toUpper(s: string): string
```

String.toUpper returns s in upper case by Unicode's full case mappings, so the length can change.
The mappings that depend on a language, such as Turkish, are not applied.

<!-- end generated -->

## Details

- One code point can become several: `"ß"` becomes `"SS"`. Code points without an upper-case form,
  such as digits, stay as they are.
- The rules for Turkish and Lithuanian, which depend on the language, are not applied, so `"i"`
  becomes `"I"`, not `"İ"`.

## Basic examples

```ant
String.toUpper("hello")  // "HELLO"
String.toUpper("straße") // "STRASSE"
```

Call it through `.` on any string:

```ant
let lang = "ant"
lang.toUpper() // "ANT"
```

## Scope

The empty string stays empty, and text with no letters comes back unchanged:

```ant
"".toUpper()       // ""
"123 !?".toUpper() // "123 !?"
```

Letters with accents and letters of other alphabets have upper-case forms too:

```ant
"élan".toUpper()   // "ÉLAN"
"σοφος".toUpper()  // "ΣΟΦΟΣ"
"привет".toUpper() // "ПРИВЕТ"
```

Scripts without case, such as Japanese, pass through:

```ant
"日本語".toUpper() // "日本語"
```

A ligature expands into its letters, so the result can be longer than `s`:

```ant
"ﬁ".toUpper()               // "FI"
"ﬃ".toUpper()               // "FFI"
"ﬁ".toUpper().length()      // 2
"straße".toUpper().length() // 7
```

A combining accent is a code point of its own and stays after the letter, which is mapped:

```ant
"e\u{301}".toUpper() == "E\u{301}" // true
```

In a pipeline, the string is the first argument:

```ant
"warning" | String.toUpper() // "WARNING"
```

`String.toUpper` is a function value too, so it can go straight to `Array.map`:

```ant
["get", "post"].map(String.toUpper) // ["GET", "POST"]
```

## Applications

Name a constant from a phrase:

```ant
func constName(phrase: string): string {
  phrase
    | String.fields()
    | Array.map(String.toUpper)
    | String.join("_")
}

constName("max retry count") // "MAX_RETRY_COUNT"
```

Take the initials of a name:

```ant
func initials(name: string): string {
  name
    | String.fields()
    | Array.map({ x.slice(0, 1).toUpper() })
    | String.concat()
}

initials("ada lovelace") // "AL"
```

Capitalize the first character a reader sees, with [`String.graphemes`](graphemes.md), so an accent
written as its own code point stays with its letter:

```ant
func capitalize(s: string): string {
  let gs = s.graphemes()
  if gs.isEmpty() {
    return s
  }
  gs[0].toUpper() + String.concat(gs.slice(1, gs.length()))
}

println(capitalize("e\u{301}lan") == "E\u{301}lan")
println(capitalize("ant"))
println(capitalize("") == "")
```

```text
true
Ant
true
```

## Properties and relations

On these strings, upper-casing twice gives the same result as upper-casing once:

```ant
for s in ["straße", "ﬁne", "σοφος", "Ant 2", ""] {
  assert(s.toUpper().toUpper() == s.toUpper())
}
```

Upper-casing `"hello"` gives a string that [`String.isUpper`](isUpper.md) accepts. A digit or a
space is not an upper-case letter, so it makes `isUpper` false:

```ant
assert("hello".toUpper().isUpper())
assert(!"hello 1".toUpper().isUpper())
```

Upper-casing and then lower-casing with [`String.toLower`](toLower.md) does not give back the
original when the mapping expands:

```ant
"straße".toUpper().toLower() // "strasse"
"Ant".toUpper().toLower()    // "ant"
```

## Possible issues

The length can change, so a code point index found in `s` may not point at the same text in
`String.toUpper(s)`:

```ant
let s = "straße!"
s.indexOf("!")           // 6
s.toUpper().indexOf("!") // 7
```

`String.toUpper` gives the upper-case letter, not the title-case one. The digraph `"ǆ"` becomes
`"Ǆ"`, not `"ǅ"`:

```ant
"ǆemal".toUpper() // "ǄEMAL"
```

Turkish rules are not applied: the dotted `"i"` becomes the dotless `"I"`:

```ant
"istanbul".toUpper() // "ISTANBUL"
```

To compare without regard to case, use [`String.equalFold`](equalFold.md) or
[`String.foldCase`](foldCase.md), which are made for it:

```ant
String.equalFold("Straße", "STRASSE") // true
```

The function on the right of a pipe needs its parentheses:

```ant
println("ant" | String.toUpper) // error: the right side of a pipe must be a call
```

A nullable string, such as the result of `Map` indexing, is a compile error. Narrow it first, or
give a default with `??`:

```ant
let user = Map{ name: "ann" }
println(String.toUpper(user["name"])) // error: cannot use string | null where string is expected
```

```ant
let user = Map{ name: "ann" }
(user["name"] ?? "").toUpper() // "ANN"
```

## See also

[String.toLower](toLower.md) · [String.foldCase](foldCase.md) · [String.equalFold](equalFold.md) ·
[String.isUpper](isUpper.md) · [String.graphemes](graphemes.md) · [String.normalize](normalize.md)
