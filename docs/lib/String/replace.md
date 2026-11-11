<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.replace

```ant signature
func String.replace(s: string, old: string, new: string): string
```

String.replace returns s with every occurrence of old replaced by new. It panics if old is empty or
the result's length in bytes is above 2^31 - 1.

<!-- end generated -->

## Details

- `old` is plain text, not a pattern, and matches exactly, code point by code point: no case
  folding, no normalization.
- Occurrences are taken from the left and never overlap: in `"aaa"`, `"aa"` occurs once.

## Basic examples

Replace every occurrence of a substring:

```ant
String.replace("a-b-c", "-", "+")             // "a+b+c"
String.replace("hello world", "world", "Ant") // "hello Ant"
```

When `old` does not occur, the string comes back unchanged:

```ant
String.replace("hello", "x", "y") // "hello"
```

Replace with `""` to delete:

```ant
String.replace("1,000,000", ",", "") // "1000000"
```

## Scope

`old` and `new` may have different lengths, and either may be longer than one code point:

```ant
String.replace("a::b::c", "::", "/") // "a/b/c"
String.replace("a.b", ".", "...")    // "a...b"
String.replace("abc", "abc", "")     // ""
String.replace("ab", "abc", "x")     // "ab"
```

The empty string has nothing to replace:

```ant
String.replace("", "a", "b") // ""
```

`old` and `new` may hold any code points: emoji, line breaks, accented letters.

```ant
String.replace("🐜🐝🐜", "🐜", "🐞")          // "🐞🐝🐞"
String.replace("one\ntwo\nthree", "\n", ", ") // "one, two, three"
String.replace("naïve café", "é", "e")        // "naïve cafe"
```

Call it through `.`, in a pipe, or with named arguments:

```ant
"a b c".replace(" ", "_")                      // "a_b_c"
"a b c" | String.replace(" ", "_")             // "a_b_c"
String.replace(s: "a b c", old: " ", new: "_") // "a_b_c"
```

## Applications

Turn Windows line endings into `\n`:

```ant
let text = "one\r\ntwo\r\n"
String.replace(text, "\r\n", "\n") // "one\ntwo\n"
```

Fill the placeholders of a template from a map:

```ant
func fill(template: string, values: Map[string, string]): string {
  var out = template
  for e in values {
    out = String.replace(out, "{" + e.key + "}", e.value)
  }
  out
}

println(fill("Dear {name}, your order {id} has shipped.", Map{ "name": "Ann", "id": "#42" }))
```

```text
Dear Ann, your order #42 has shipped.
```

Escape `&`, `<` and `>` in text. Replace `&` first, so the `&` of the entities added after it stays
as it is:

```ant
func escape(s: string): string {
  s
    | String.replace("&", "&amp;")
    | String.replace("<", "&lt;")
    | String.replace(">", "&gt;")
}

println(escape("a < b && c > d"))
```

```text
a &lt; b &amp;&amp; c &gt; d
```

To build HTML, use an [HTML expression](../../guide/html.md), which escapes `{...}` text itself.

Build a slug from a title by chaining calls through `.`:

```ant
"Hello World Again".toLower().replace(" ", "-") // "hello-world-again"
```

## Properties and relations

Replacing a string with itself, or replacing a string that does not occur, gives `s` back:

```ant
let s = "banana"
assert(String.replace(s, "an", "an") == s)
assert(String.replace(s, "x", "y") == s)
assert(!String.contains(s, "x"))
```

After a replacement, `old` occurs no more, as long as `new` does not bring it back:

```ant
let s = "a--b---c"
assert(!String.contains(String.replace(s, "-", " "), "-"))
```

`String.replace` gives the same as splitting at `old` and joining with `new`:

```ant
for s in ["a,b,,c", "", ",", "abc"] {
  assert(String.replace(s, ",", ";") == String.join(String.split(s, ","), ";"))
}
```

The length changes by the [`String.count`](count.md) of `old` times the difference in length:

```ant
let s = "the cat sat on the mat"
let old = "at"
let new = "ug"
let r = String.replace(s, old, new)
assert(r == "the cug sug on the mug")
let grow = String.length(new) - String.length(old)
assert(String.length(r) == String.length(s) + String.count(s, old) * grow)
```

Only the first occurrence changes with [`String.replaceFirst`](replaceFirst.md); when `old` occurs
once, the two agree:

```ant
String.replace("a-b-c", "-", "+")      // "a+b+c"
String.replaceFirst("a-b-c", "-", "+") // "a+b-c"
assert(String.replace("a-b", "-", "+") == String.replaceFirst("a-b", "-", "+"))
```

## Possible issues

A call returns the new string and leaves `s` as it was, so a call whose result is not used changes
nothing:

```ant
let s = "a b"
s.replace(" ", "_") // "a_b"
s                   // "a b"
```

Assign the result: `s = s.replace(" ", "_")`.

An empty `old` panics:

```ant
String.replace("abc", "", "-") // panic: String.replace needs an old string that is not empty
```

To put something between every character, join the code points instead:

```ant
String.join(String.chars("abc"), "-") // "a-b-c"
```

The match is case-sensitive:

```ant
String.replace("Ant ant ANT", "ant", "bee") // "Ant bee ANT"
```

`String.replace` does not normalize. A precomposed `é` does not contain the code point `e`, while an
`e` followed by a combining accent does. The accent stays behind and now sits on the `E`:

```ant
String.replace("café", "e", "E")        // "café"
String.replace("cafe\u{301}", "e", "E") // "cafÉ"
```

`old` is not a pattern: a `.` matches only a `.`, and `$1` in `new` is written as it is. Use
[`Regex.replace`](../Regex/replace.md) for patterns and groups:

```ant
String.replace("a.b", ".", "!")                   // "a!b"
String.replace("ann@home", "@", "$1")             // "ann$1home"
Regex.replace(/(\w+)@(\w+)/, "ann@home", "$2:$1") // "home:ann"
```

Occurrences do not overlap. The search goes on after the end of each match, so a run of `old` is cut
from the left:

```ant
String.replace("aaa", "aa", "X")  // "Xa"
String.replace("aaaa", "aa", "X") // "XX"
String.count("aaa", "aa")         // 1
```

The text that `new` inserts is not searched again, so a replacement that contains `old` does not
repeat:

```ant
String.replace("a", "a", "aa") // "aa"
```

Chained replacements run one after another, so a swap needs a placeholder that the text does not
hold:

```ant
let s = "left right"
s.replace("left", "right").replace("right", "left")                           // "left left"
s.replace("left", "\u{0}").replace("right", "left").replace("\u{0}", "right") // "right left"
```

## See also

[String.replaceFirst](replaceFirst.md) · [String.count](count.md) · [String.split](split.md) ·
[String.join](join.md) · [String.contains](contains.md) · [String.trimPrefix](trimPrefix.md) ·
[String.normalize](normalize.md) · [Regex.replace](../Regex/replace.md) ·
[Regex.replaceWith](../Regex/replaceWith.md)
