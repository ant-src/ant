<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.equalFold

```ant signature
func String.equalFold(a: string, b: string): bool
```

String.equalFold reports whether a and b are equal without regard to case, as String.foldCase(a) ==
String.foldCase(b).

<!-- end generated -->

## Details

- `String.equalFold(a, b)` is `String.foldCase(a) == String.foldCase(b)`.
- Full case folding maps one character to several where Unicode says so: `"Straße"` and `"STRASSE"`
  are equal.
- There is no Turkish rule: `"İ"` and `"i"` are not equal, nor are `"ı"` and `"I"`.
- It changes case only. It does not remove accents, so `"café"` and `"CAFE"` differ, and it does not
  normalize; use `String.normalize(s, "NFC")` on both first.
- The `i` flag of a regular expression folds one code point at a time, so `/straße/i` does not match
  `"STRASSE"`.

## Basic examples

Compare two words without regard to case:

```ant
String.equalFold("Ant", "ANT")  // true
String.equalFold("Ant", "ants") // false
```

Call it through `.` on the first string:

```ant
"Hello".equalFold("hELLO") // true
```

## Scope

Two empty strings are equal, and an empty string equals no other:

```ant
"".equalFold("")  // true
"".equalFold(" ") // false
```

Case folding can change the length of a string. `"ß"` folds to `"ss"`:

```ant
"Straße".equalFold("STRASSE") // true
"ß".equalFold("SS")           // true
```

It covers every script with case, such as Greek, where the final sigma `ς` and `σ` fold alike:

```ant
"σοφος".equalFold("ΣΟΦΟΣ")       // true
"Ångström".equalFold("ÅNGSTRÖM") // true
```

Compatibility characters that have case fold too. The ligature `ﬁ` folds to `"fi"`, and the Kelvin
sign `K` to `"k"`:

```ant
"ﬁ".equalFold("FI")       // true
"\u{212A}".equalFold("k") // true
```

Characters without case, such as digits, spaces, and punctuation, must match exactly:

```ant
"Room 101!".equalFold("ROOM 101!") // true
"Room 101".equalFold("room  101")  // false
```

In a pipe, the left side is `a`:

```ant
"README.md" | String.equalFold("readme.MD") // true
```

Name the arguments:

```ant
String.equalFold(a: "Yes", b: "YES") // true
```

## Applications

Accept a yes or no answer however the user typed it:

```ant
func confirmed(answer: string): bool {
  let a = answer.trim()
  a.equalFold("y") || a.equalFold("yes")
}

confirmed(" Yes ") // true
confirmed("YES")   // true
confirmed("no")    // false
```

Look up an HTTP header, whose names are case-insensitive:

```ant
type Header = { name: string, value: string }
let headers = [
  Header{ name: "Content-Type", value: "text/html" },
  Header{ name: "X-Request-Id", value: "42" },
]
headers.find({ x.name.equalFold("content-type") })?.value // "text/html"
```

Check a file extension without regard to case:

```ant
func isImage(name: string): bool {
  let dot = name.lastIndexOf(".")
  if dot == null {
    return false
  }
  let ext = name.slice(dot + 1, name.length())
  ["png", "jpg", "gif"].some({ ext.equalFold(x) })
}

isImage("photo.JPG") // true
isImage("notes.txt") // false
isImage("README")    // false
```

## Properties and relations

`equalFold` is reflexive and symmetric:

```ant
let a = "Straße"
let b = "STRASSE"
assert(a.equalFold(a))
assert(a.equalFold(b) == b.equalFold(a))
```

It agrees with comparing the folded strings with `==`:

```ant
let pairs = [["Ant", "aNT"], ["ß", "ss"], ["İ", "i"], ["a", "b"]]
for p in pairs {
  assert(p[0].equalFold(p[1]) == (p[0].foldCase() == p[1].foldCase()))
}
```

## Possible issues

Comparing with `String.toLower` is not the same. Lower case keeps `ß`, so it misses what folding
finds:

```ant
"Straße".toLower() == "STRASSE".toLower() // false
"Straße".equalFold("STRASSE")             // true
```

`equalFold` does not normalize. The same accented letter spelled two ways compares `false` until
both are normalized:

```ant
let composed = "\u{E9}"
let decomposed = "e\u{301}"
composed.equalFold(decomposed) // false
let nfc = decomposed.normalize("NFC")
composed.equalFold(nfc) // true
```

The Turkish dotted and dotless i do not fold to `i` and `I`. The full case folding has no Turkish
rule, so `"İ"` folds to `"i\u{307}"` and `"ı"` stays `"ı"`:

```ant
"İ".equalFold("i") // false
"ı".equalFold("I") // false
```

Folding changes case only, not accents, so `é` does not match `E`:

```ant
"café".equalFold("CAFE") // false
```

To ignore accents as well, strip them first: split each accent from its letter with
[String.normalize](normalize.md) in `NFD`, then drop the code points that [String.isMark](isMark.md)
accepts:

```ant
func stripAccents(s: string): string {
  s.normalize("NFD")
    | String.chars()
    | Array.filter({ !x.isMark() })
    | String.join("")
}

stripAccents("café").equalFold("CAFE") // true
```

A regex with the `i` flag folds one code point at a time, so it can reject what `equalFold` accepts:

```ant
Regex.test(/straße/i, "STRASSE") // false
"straße".equalFold("STRASSE")    // true
```

To group or deduplicate strings without regard to case, use `String.foldCase` as the key instead of
comparing every pair with `equalFold`:

```ant
["Ant", "ANT", "bee", "Bee"].uniqueBy(String.foldCase) // ["Ant", "bee"]
```

## See also

[String.foldCase](foldCase.md) · [String.toLower](toLower.md) · [String.toUpper](toUpper.md) ·
[String.normalize](normalize.md) · [Regex.test](../Regex/test.md) ·
[Array.uniqueBy](../Array/uniqueBy.md)
