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
let path = "/usr/local/bin"
path.count("/") // 3
```

## Scope

Occurrences do not overlap, so the characters of one match are not part of another:

```ant
"aaaa".count("aa")    // 2
"banana".count("ana") // 1
```

An empty string holds no occurrences, a `t` longer than `s` none either, and `s` itself one:

```ant
"".count("a")        // 0
"ab".count("abc")    // 0
"five".count("five") // 1
```

`t` may be any string, an emoji or an escape such as `\n` included:

```ant
"🐜🐜🐜".count("🐜")    // 3
"a\nb\nc\n".count("\n") // 3
```

In a pipeline, the string is the first argument:

```ant
"2026-11-08" | String.count("-") // 2
```

## Applications

Count the lines of a text that ends each line with `\n`:

```ant
let text = "one\ntwo\nthree\n"
println("${text.count("\n")} lines")
```

```text
3 lines
```

A last line without `\n` is not counted; `String.lines` counts it:

```ant
"one\ntwo".count("\n")      // 1
"one\ntwo".lines().length() // 2
```

Check that the brackets of an expression balance in number:

```ant
func sameCount(s: string): bool {
  s.count("(") == s.count(")")
}

sameCount("(1 + 2) * (3 - 4)") // true
sameCount("f(g(x)")            // false
```

Count a word in any case by folding both sides first:

```ant
let text = "Ant, ANT, ant."
text.foldCase().count("ant".foldCase()) // 3
```

## Properties and relations

The count is above `0` exactly when `String.contains` holds:

```ant
for t in ["e", "ee", "x"] {
  assert(("cheese".count(t) > 0) == "cheese".contains(t))
}
```

Counting a single code point gives the same as `Array.countWhere` over `String.chars`:

```ant
let s = "mississippi"
assert(s.count("s") == s.chars().countWhere({ x == "s" }))
```

Splitting at `t` gives one piece more than the count:

```ant
for t in ["a", "an", "x"] {
  assert("banana".split(t).length() == "banana".count(t) + 1)
}
```

## Possible issues

An empty `t` panics:

```ant
"abc".count("") // panic: String.count needs a substring that is not empty
```

To count code points, use `String.length`; to count what a reader sees as characters,
`s.graphemes().length()`:

```ant
"e\u{301}".length()             // 2
"e\u{301}".graphemes().length() // 1
```

`String.count` skips an occurrence that overlaps one it counts, so `"abababa"` holds three `"aba"`
but counts two. To count every position where `t` starts, test each one:

```ant
func countOverlapping(s: string, t: string): int {
  var n = 0
  for i in Iter.range(0, s.length()) {
    if s.slice(i, s.length()).startsWith(t) {
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
let text = "other then the"
text.count("the")                        // 3
text.fields().countWhere({ x == "the" }) // 1
```

The match is case-sensitive:

```ant
"Hello hello".count("hello") // 1
```

`String.count` does not normalize. A precomposed "é" does not contain the code point "e", while "e"
followed by a combining accent does:

```ant
"café".count("e")        // 0
"cafe\u{301}".count("e") // 1
```

It counts code point sequences, not graphemes, so a part of an emoji joined from several people is
found:

```ant
"👨‍👩‍👧".count("👩")          // 1
"👨‍👩‍👧".graphemes().length() // 1
```

## See also

[String.contains](contains.md) · [String.indexOf](indexOf.md) · [String.lastIndexOf](lastIndexOf.md)
· [String.split](split.md) · [String.replace](replace.md) · [String.fields](fields.md) ·
[String.lines](lines.md) · [String.length](length.md) · [String.graphemes](graphemes.md) ·
[String.foldCase](foldCase.md) · [Array.countWhere](../Array/countWhere.md)
