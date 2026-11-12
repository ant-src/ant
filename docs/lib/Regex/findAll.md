<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.findAll

```ant signature
func Regex.findAll(p: Regex.Pattern, s: string): Array[Regex.Match]
```

Regex.findAll returns every match of p in s, left to right and without overlap. An empty match that
abuts the match before it is skipped.

<!-- end generated -->

## Details

- The result is `[]` when `p` never matches.
- A pattern that can match the empty string, such as `/a*/`, gives empty matches between the others,
  except right after a match.
- `start` and `end` count code points, so `s.slice(m.start, m.end)` is `m.text`.

## Basic examples

The text of every match:

```ant
Regex.findAll(/\d+/, "a1b22") | Array.map({ x.text }) // ["1", "22"]
```

Each match is a `Regex.Match` record:

```ant
Regex.findAll(/\d+/, "a1b22")
// [Regex.Match{ text: "1", start: 1, end: 2, groups: ["1"], named: Map{} }, Regex.Match{ text: "22", start: 3, end: 5, groups: ["22"], named: Map{} }]
```

No match gives an empty array:

```ant
Regex.findAll(/\d+/, "none") // []
```

## Scope

### Groups

`groups[0]` is the whole match and `groups[n]` the `n`th group:

```ant
let dates = Regex.findAll(/(\d{4})-(\d{2})/, "2024-01, 2025-11")
dates.map({ x.groups })    // [["2024-01", "2024", "01"], ["2025-11", "2025", "11"]]
dates.map({ x.groups[1] }) // ["2024", "2025"]
```

A group that took no part in a match is `null`:

```ant
Regex.findAll(/(a)|(b)/, "ab") | Array.map({ x.groups }) // [["a", "a", null], ["b", null, "b"]]
```

Named groups also fill `named`:

```ant
let pairs = Regex.findAll(/(?<key>\w+)=(?<value>\w*)/, "a=1 b=")
pairs.map({ x.named }) // [Map{ "key": "a", "value": "1" }, Map{ "key": "b", "value": "" }]
```

### Flags

`i` ignores case, `m` makes `^` and `$` match at every line, and `s` lets `.` match a line break:

```ant
Regex.findAll(/cat/i, "Cat CAT cat") | Array.map({ x.text })          // ["Cat", "CAT", "cat"]
Regex.findAll(/^\w+/m, "one two\nthree four") | Array.map({ x.text }) // ["one", "three"]
Regex.findAll(/^\w+/, "one two\nthree four") | Array.map({ x.text })  // ["one"]
Regex.findAll(/a.b/s, "a\nb axb") | Array.map({ x.text })             // ["a\nb", "axb"]
```

### Unicode

`start` and `end` are code point indices, and `\pL` matches a letter of any script:

```ant
let words = Regex.findAll(/\pL+/, "héllo 世界!")
words.map({ x.text })           // ["héllo", "世界"]
words.map({ [x.start, x.end] }) // [[0, 5], [6, 8]]
```

### Empty input and empty matches

An empty `s` has one empty match for a pattern that matches the empty string, and none otherwise:

```ant
Regex.findAll(/a*/, "").length() // 1
Regex.findAll(/a+/, "").length() // 0
```

The empty match at index 4 of `"baaab"` abuts `"aaa"`, so it is skipped:

```ant
Regex.findAll(/a*/, "baaab") | Array.map({ [x.start, x.end] }) // [[0, 0], [1, 4], [5, 5]]
```

### Compiled patterns

A pattern built from data works the same as a literal:

```ant
let word = "ant"
let p = Result.unwrap(Regex.compile(Regex.escape(word), "i"))
Regex.findAll(p, "Ant, ANT, and ants").length() // 3
```

## Applications

Sum the numbers in a line of text:

```ant
let line = "3 apples, 12 pears, 7 figs"
let total = Regex.findAll(/\d+/, line)
  | Array.map({ x.text.toInt() ?? 0 })
  | Array.sum()
println(total ?? 0)
```

```text
22
```

Collect the distinct hashtags of a post:

```ant
Regex.findAll(/#\w+/, "#ant is fun #lang #ant") | Array.map({ x.text }) | Array.unique()
// ["#ant", "#lang"]
```

Read `key=value` settings into a map, with named groups:

```ant
let text = "host=localhost port=8080 debug=true"
let settings = Regex.findAll(/(?<key>\w+)=(?<value>\S+)/, text)
  | Array.map({ Entry{ key: x.named["key"] ?? "", value: x.named["value"] ?? "" } })
  | Map.fromEntries()
println(settings["port"])
```

```text
8080
```

Report where each `TODO` sits in a line:

```ant
let source = "a := 1 // TODO rename; b := 2 // TODO test"
for m in Regex.findAll(/TODO \w+/, source) {
  println("${m.start}: ${m.text}")
}
```

```text
10: TODO rename
33: TODO test
```

## Properties and relations

Each match's `text` is the slice of `s` between `start` and `end`:

```ant
let s = "x 12 y 345 é 6"
assert(Regex.findAll(/\d+/, s).every({ s.slice(x.start, x.end) == x.text }))
```

The first match is what [Regex.find](find.md) returns:

```ant
let s = "ab12cd345"
assert(Regex.findAll(/\d+/, s).first() == Regex.find(/\d+/, s))
```

The array is empty exactly when [Regex.test](test.md) is false:

```ant
let s = "no digits"
assert(Regex.findAll(/\d/, s).isEmpty() == !Regex.test(/\d/, s))
```

## Possible issues

Matches do not overlap, so `/aa/` matches `"aaaa"` twice, not three times:

```ant
Regex.findAll(/aa/, "aaaa") | Array.map({ x.start }) // [0, 2]
```

A pattern that can match the empty string, such as `/x*/`, also matches between the characters. Use
`+` instead of `*` to keep only non-empty matches:

```ant
Regex.findAll(/x*/, "axb") | Array.map({ x.text }) // ["", "x", ""]
Regex.findAll(/x+/, "axb") | Array.map({ x.text }) // ["x"]
```

`\w` is ASCII, so `/\w+/` breaks words at accented letters; use `\pL` for letters of any script:

```ant
Regex.findAll(/\w+/, "héllo wörld") | Array.map({ x.text })  // ["h", "llo", "w", "rld"]
Regex.findAll(/\pL+/, "héllo wörld") | Array.map({ x.text }) // ["héllo", "wörld"]
```

There is no `g` flag; `Regex.findAll` already finds every match:

```ant
let all = Regex.findAll(/a/g, "aa")
// error: there is no g flag: Regex.findAll and Regex.replace work on every match
```

A group may be `null`, so a group is not a `string` until you handle the `null`:

```ant
let m = Regex.findAll(/(a)|(b)/, "b")[0]
let second: string = m.groups[1] // error: cannot use string | null where string is expected
```

## See also

- [Regex.find](find.md): the first match only.
- [Regex.test](test.md): whether there is a match at all.
- [Regex.replaceWith](replaceWith.md): replace every match with a computed string.
- [Regex.split](split.md): the pieces between the matches.
- [Regex.Match](types/Match.md)
