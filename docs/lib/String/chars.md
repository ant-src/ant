<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.chars

```ant signature
func String.chars(s: string): Array[string]
```

String.chars returns the code points of s, each as a one-code-point string.

<!-- end generated -->

## Details

- Each element of the result is a string of one code point, in the order the code points appear in
  `s`. Ant has no character type: a character is a string of one code point
  ([String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- The result holds `String.length(s)` elements, and an empty `s` gives `[]`.
- The elements are the strings a `for` loop over `s` binds, in the same order
  ([Loops](https://github.com/ant-src/spec/blob/master/ant-lang.md#loops)).
- A code point is not always what a reader sees as one character: an accent can be a code point of
  its own, and a flag or a family emoji is several. `String.graphemes` splits text into those units.
- `String.chars` returns a new array; changing it leaves `s` as it was, since a string is immutable.

## Basic examples

Split a word into its characters:

```ant
String.chars("ant") // ["a", "n", "t"]
```

Index the array to read a character by position, which a string itself does not allow:

```ant
var cs = String.chars("hello")
cs[1]           // "e"
cs[len(cs) - 1] // "o"
```

## Scope

An empty string gives an empty array:

```ant
String.chars("") // []
```

A letter outside ASCII, such as `é` written as one code point, is one element:

```ant
String.chars("héllo") // ["h", "é", "l", "l", "o"]
```

White space and control characters are code points like any other:

```ant
String.chars("a b\n") // ["a", " ", "b", "\n"]
```

An accent written as a combining mark after its letter is a separate element. `String.codePoint`
shows the two code points:

```ant
String.chars("e\u{301}").map(String.codePoint) // [101, 769]
String.chars("\u{E9}").map(String.codePoint)   // [233]
```

A flag is two regional indicator symbols, and a family emoji is three people joined by zero-width
joiners. `String.graphemes` keeps the family as one element:

```ant
String.chars("🇨🇭")             // ["🇨", "🇭"]
String.chars("👨‍👩‍👧")          // ["👨", "\u{200d}", "👩", "\u{200d}", "👧"]
len(String.graphemes("👨‍👩‍👧")) // 1
```

Call it through `.` or in a pipe:

```ant
"ant".chars()          // ["a", "n", "t"]
"ant" | String.chars() // ["a", "n", "t"]
```

## Applications

Keep the digits of a string, passing a function on a type to `Array.filter`:

```ant
String.chars("a1b22c3").filter(String.isDigit) // ["1", "2", "2", "3"]
```

List the distinct letters of a word, in the order they first appear:

```ant
String.chars("Mississippi").unique() // ["M", "i", "s", "p"]
```

Count how often each character occurs:

```ant
func counts(s: string): Map[string, int] {
  var m: Map[string, int] = Map{}
  for c in String.chars(s) {
    m[c] = (m[c] ?? 0) + 1
  }
  m
}

counts("banana") // Map{ "b": 1, "a": 3, "n": 2 }
```

Compare each character with the next one, which needs indices:

```ant
var cs = String.chars("anthill")
for i in Iter.range(0, len(cs) - 1) {
  if cs[i] == cs[i + 1] {
    println("double ${cs[i]} at ${i}")
  }
}
```

```text
double l at 5
```

Check whether a phrase reads the same backward, ignoring case and anything but letters:

```ant
func isPalindrome(s: string): bool {
  var cs = String.chars(String.toLower(s)).filter(String.isLetter)
  cs == Array.reverse(cs)
}

isPalindrome("Never odd or even") // true
isPalindrome("anthill")           // false
```

Shift each letter by 13 places, and put the characters back together with `String.join`:

```ant
func rot13(s: string): string {
  String.chars(s).map(func(c: string): string {
    var n = String.codePoint(c)
    if n >= 97 && n <= 122 {
      return String.fromCodePoint((n - 97 + 13) % 26 + 97) ?? c
    }
    if n >= 65 && n <= 90 {
      return String.fromCodePoint((n - 65 + 13) % 26 + 65) ?? c
    }
    c
  }) | String.join("")
}

rot13("Hello, Ant") // "Uryyb, Nag"
```

## Properties and relations

The result has `String.length(s)` elements, each of length 1:

```ant
var s = "héllo 🐜"
var cs = String.chars(s)
assert(len(cs) == String.length(s))
assert(cs.every({ String.length(x) == 1 }))
```

Joining the characters gives back the string:

```ant
var s = "e\u{301}🇨🇭 ant"
assert(String.join(String.chars(s), "") == s)
```

A `for` loop over the string visits the same characters in the same order:

```ant
var s = "héllo"
var seen: Array[string] = []
for c in s {
  Array.push(seen, c)
}
assert(seen == String.chars(s))
```

Each character's code point is what `String.toCodePoints` gives:

```ant
var s = "añt🐜"
assert(String.chars(s).map(String.codePoint) == String.toCodePoints(s))
```

## Possible issues

Reversing the characters moves an accent onto another letter, and turns a flag into two loose letter
symbols. Reverse the graphemes instead:

```ant
var s = "e\u{301}x"
String.join(Array.reverse(String.chars(s)), "") == "x\u{301}e"     // true
String.join(Array.reverse(String.graphemes(s)), "") == "xe\u{301}" // true
```

The two spellings of `é` give different characters, so compare after `String.normalize`:

```ant
String.chars("\u{E9}") == String.chars("e\u{301}")                          // false
String.chars(String.normalize("e\u{301}", "NFC")) == String.chars("\u{E9}") // true
```

An index past the last character panics:

```ant
String.chars("ab")[2] // panic: index 2 out of range for an array of length 2
```

To read one character or a range once, `String.slice` gives it as a string, with no array:

```ant
String.slice("hello", 1, 2) // "e"
```

## See also

[String.graphemes](graphemes.md), [String.toCodePoints](toCodePoints.md),
[String.codePoint](codePoint.md), [String.length](length.md), [String.slice](slice.md),
[String.join](join.md), [String.normalize](normalize.md), [Array.reverse](../Array/reverse.md)
