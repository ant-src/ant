<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toLower

```ant signature
func String.toLower(s: string): string
```

String.toLower returns s in lower case by Unicode's full case mappings. The string is mapped whole,
so a final Greek sigma becomes ς, and the mappings that depend on a language, such as Turkish, are
not applied.

<!-- end generated -->

## Details

- The result can be longer than `s`: `"İ"` becomes `"i\u{307}"`, an `i` and a combining dot.
- The string is mapped whole: a capital sigma at the end of a word becomes `ς`, elsewhere `σ`.
- The rules for Turkish and Lithuanian, which depend on the language, are not applied.

## Basic examples

Lower-case a string:

```ant
String.toLower("Hello, World") // "hello, world"
String.toLower("ANT")          // "ant"
```

Call it through `.`:

```ant
"README".toLower() // "readme"
```

## Scope

An empty string stays empty, and text with no letters to change, such as digits and spaces, comes
back unchanged:

```ant
String.toLower("")         // ""
String.toLower("room 101") // "room 101"
```

Every script with case is mapped, not only Latin:

```ant
String.toLower("ÅNGSTRÖM") // "ångström"
String.toLower("ПРИВЕТ")   // "привет"
```

The Greek capital sigma becomes `ς` at the end of a word and `σ` elsewhere:

```ant
String.toLower("ΣΟΦΟΣ")      // "σοφος"
String.toLower("ΟΔΟΣ ΣΟΦΟΣ") // "οδος σοφος"
```

A mapping can make the string longer. The capital `İ` with a dot lowers to two code points:

```ant
String.toLower("İ") == "i\u{307}"  // true
String.length(String.toLower("İ")) // 2
```

Compatibility characters with case are lowered too, such as the Kelvin sign and the title-case
digraph `ǅ`:

```ant
String.toLower("\u{212A}") // "k"
String.toLower("ǅ")        // "ǆ"
```

In a pipe, the left side is `s`:

```ant
"Content-Type" | String.toLower() // "content-type"
```

Name the argument:

```ant
String.toLower(s: "YES") // "yes"
```

Pass it as a function value:

```ant
["Ant", "BEE", "Cat"].map(String.toLower) // ["ant", "bee", "cat"]
```

## Applications

Make a slug for a URL from a title:

```ant
func slug(title: string): string {
  String.join(String.fields(String.toLower(title)), "-")
}

slug("Hello  Ant World") // "hello-ant-world"
```

Check a file name for an image extension however it is spelled:

```ant
func isImage(name: string): bool {
  let lower = String.toLower(name)
  [".png", ".jpg", ".gif"].some({ String.endsWith(lower, x) })
}

isImage("Photo.JPG") // true
isImage("notes.txt") // false
```

## Properties and relations

Lower case and upper case are not inverses. `toUpper` turns `ß` into `"SS"`, and lowering that gives
`"ss"`:

```ant
assert(String.toLower(String.toUpper("ß")) == "ss")
assert(String.toLower("ß") != String.toLower(String.toUpper("ß")))
```

## Possible issues

Mapping each character on its own loses the final sigma, since the sigma depends on what follows it:

```ant
String.join(String.chars("ΣΟΦΟΣ").map(String.toLower), "") // "σοφοσ"
String.toLower("ΣΟΦΟΣ")                                    // "σοφος"
```

The Turkish rules are not applied, so the dotless `I` lowers to `i`, not `ı`, and `İ` gains a
combining dot:

```ant
String.toLower("ISTANBUL") // "istanbul"
String.toLower("İ") == "i" // false
```

Comparing lowered strings misses some matches, because lower case keeps `ß`. Compare with
`String.equalFold`, or key with `String.foldCase`:

```ant
String.toLower("Straße") == String.toLower("STRASSE") // false
String.equalFold("Straße", "STRASSE")                 // true
```

## See also

[String.toUpper](toUpper.md) · [String.foldCase](foldCase.md) · [String.equalFold](equalFold.md) ·
[String.isLower](isLower.md) · [String.normalize](normalize.md)
