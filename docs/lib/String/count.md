<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.count

```ant signature
func String.count(s: string, t: string): int
```

String.count returns the number of non-overlapping occurrences of t in s. It panics if t is empty.

<!-- end generated -->

## Basic examples

Count the occurrences of a substring:

```ant
String.count("cheese", "e")    // 3
String.count("banana", "na")   // 2
String.count("banana", "kiwi") // 0
```

Call it through `.` on any string:

```ant
var path = "/usr/local/bin"
path.count("/") // 3
```

## Scope

Occurrences do not overlap, so the characters of one match are not part of another:

```ant
String.count("aaaa", "aa")    // 2
String.count("banana", "ana") // 1
```

An empty string holds no occurrences, a `t` longer than `s` none either, and `s` itself one:

```ant
String.count("", "a")        // 0
String.count("ab", "abc")    // 0
String.count("five", "five") // 1
```

`t` may be any string, an emoji or an escape such as `\n` included:

```ant
String.count("🐜🐜🐜", "🐜")    // 3
String.count("a\nb\nc\n", "\n") // 3
```

In a pipeline, the string is the first argument:

```ant
"2026-11-08" | String.count("-") // 2
```

## Applications

Count the lines of a text that ends each line with `\n`:

```ant
var text = "one\ntwo\nthree\n"
println("${String.count(text, "\n")} lines")
```

```text
3 lines
```

A last line without `\n` is not counted; `String.lines` counts it:

```ant
String.count("one\ntwo", "\n") // 1
len(String.lines("one\ntwo"))  // 2
```

Check that the brackets of an expression balance in number:

```ant
func sameCount(s: string): bool {
  String.count(s, "(") == String.count(s, ")")
}

sameCount("(1 + 2) * (3 - 4)") // true
sameCount("f(g(x)")            // false
```

Count a word in any case by folding both sides first:

```ant
var text = "Ant, ANT, ant."
String.count(String.foldCase(text), String.foldCase("ant")) // 3
```

## Properties and relations

The count is above `0` exactly when `String.contains` holds:

```ant
for t in ["e", "ee", "x"] {
  assert((String.count("cheese", t) > 0) == String.contains("cheese", t))
}
```

Counting a single code point gives the same as `Array.countWhere` over `String.chars`:

```ant
var s = "mississippi"
assert(String.count(s, "s") == String.chars(s).countWhere({ x == "s" }))
```

Splitting at `t` gives one piece more than the count:

```ant
for t in ["a", "an", "x"] {
  assert(len(String.split("banana", t)) == String.count("banana", t) + 1)
}
```

## Possible issues

An empty `t` panics:

```ant
String.count("abc", "") // panic: String.count needs a substring that is not empty
```

To count code points, use `String.length`; to count what a reader sees as characters,
`len(String.graphemes(s))`:

```ant
String.length("e\u{301}")         // 2
len(String.graphemes("e\u{301}")) // 1
```

`String.count` skips an occurrence that overlaps one it counts, so `"abababa"` holds three `"aba"`
but counts two. To count every position where `t` starts, test each one:

```ant
func countOverlapping(s: string, t: string): int {
  var n = 0
  for i in Iter.range(0, String.length(s)) {
    if String.startsWith(String.slice(s, i, String.length(s)), t) {
      n += 1
    }
  }
  n
}

String.count("abababa", "aba")     // 2
countOverlapping("abababa", "aba") // 3
```

`String.count` matches substrings, not words, so "the" is found inside "other" and "then". To count
whole words, split the text into fields first:

```ant
var text = "other then the"
String.count(text, "the")                      // 3
String.fields(text).countWhere({ x == "the" }) // 1
```

The match is case-sensitive:

```ant
String.count("Hello hello", "hello") // 1
```

`String.count` does not normalize. A precomposed "é" does not contain the code point "e", while "e"
followed by a combining accent does:

```ant
String.count("café", "e")        // 0
String.count("cafe\u{301}", "e") // 1
```

It counts code point sequences, not graphemes, so a part of an emoji joined from several people is
found:

```ant
String.count("👨‍👩‍👧", "👩")    // 1
len(String.graphemes("👨‍👩‍👧")) // 1
```

## See also

[String.contains](contains.md) · [String.indexOf](indexOf.md) · [String.lastIndexOf](lastIndexOf.md)
· [String.split](split.md) · [String.replace](replace.md) · [String.fields](fields.md) ·
[String.lines](lines.md) · [String.length](length.md) · [String.graphemes](graphemes.md) ·
[String.foldCase](foldCase.md) · [Array.countWhere](../Array/countWhere.md)
