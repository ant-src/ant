<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.Match

```ant signature
type Regex.Match = { text: string, start: int, end: int, groups: Array[string | null], named: Map[string, string | null] }
```

Regex.Match is one match of a whole pattern. text is the matched text and start and end its code
point indices into the searched string; groups\[0\] is the whole match and groups\[n\] the nth
group, null when the group took no part; named has an entry for each named group, in the pattern's
order, null likewise.

<!-- end generated -->

## Details

- [Regex.find](../find.md) returns `Regex.Match | null`, so narrow the result before reading a
  field. [Regex.findAll](../findAll.md) and the callback of [Regex.replaceWith](../replaceWith.md)
  give matches that need no narrowing.
- `groups` has one entry more than the pattern has groups. A group that took no part is `null`; a
  group that matched empty text is `""`.
- `named` has an entry for every named group, matched or not. Indexing it with a name the pattern
  does not have also gives `null`.

## Basic examples

[Regex.find](../find.md) returns a `Regex.Match` for the leftmost match:

```ant
Regex.find(/\d+/, "ab12") // Regex.Match{ text: "12", start: 2, end: 4, groups: ["12"], named: Map{} }
```

Narrow the result, then read the fields:

```ant
let m = Regex.find(/(\w+)@(\w+)/, "mail ann@home now")
if m == null {
  return
}
println(m.text)
println("${m.start}..${m.end}")
println(m.groups)
```

```text
ann@home
5..13
["ann@home", "ann", "home"]
```

Read one field through `?.`, which gives `null` when nothing matched:

```ant
Regex.find(/\d+/, "ab12")?.text // "12"
Regex.find(/\d+/, "abc")?.text  // null
```

## Scope

Named groups appear both in `groups`, by number, and in `named`, by name, in the pattern's order:

```ant
let m = Regex.find(/(?<year>\d{4})-(?<month>\d\d)/, "on 2026-11")
m?.groups // ["2026-11", "2026", "11"]
m?.named  // Map{ "year": "2026", "month": "11" }
```

A group that took no part is `null`, and one that matched empty text is `""`:

```ant
Regex.find(/(a)(b)?(c?)/, "a")?.groups // ["a", "a", null, ""]
```

A named group in the branch that did not match is in `named` with `null`:

```ant
Regex.find(/(?<num>\d+)|(?<word>[a-z]+)/, "hello")?.named // Map{ "num": null, "word": "hello" }
```

As in RE2, groups are numbered by their opening parentheses, so an outer group comes first:

```ant
Regex.find(/((\w)\w*)@/, "ann@home")?.groups // ["ann@", "ann", "a"]
```

An empty match has empty `text` and `start` equal to `end`:

```ant
Regex.find(/x*/, "abc") // Regex.Match{ text: "", start: 0, end: 0, groups: [""], named: Map{} }
```

Indices count code points, not bytes. The `é` before the match counts as one:

```ant
let m = Regex.find(/\d+/, "café 42")
m?.start // 5
m?.end   // 7
```

[Regex.findAll](../findAll.md) returns an array of matches, read with `x` in a callback:

```ant
let ms = Regex.findAll(/\d+/, "a1 b22 c333")
ms.map({ x.text })                // ["1", "22", "333"]
ms.map({ "${x.start}-${x.end}" }) // ["1-2", "4-6", "8-11"]
```

[Regex.replaceWith](../replaceWith.md) passes each match to its function:

```ant
Regex.replaceWith(/\d+/, "3 apples and 12 pears", { "[${x.text}]" }) // "[3] apples and [12] pears"
```

Destructure the fields you need:

```ant
let m = Regex.find(/[A-Z]+/, "see HTTP spec")
if m == null {
  return
}
let { text, start } = m
println("${text} at ${start}")
```

```text
HTTP at 4
```

`Regex.Match` is a record type, so a literal builds one, for example to test a function that takes a
match:

```ant
func describe(m: Regex.Match): string {
  "${m.text} at ${m.start}"
}

println(describe(Regex.Match{ text: "ab", start: 3, end: 5, groups: ["ab"], named: Map{} }))
```

```text
ab at 3
```

`Json.stringify` writes a match as an object:

```ant
Json.stringify(Regex.find(/(?<k>\w)=(\d)/, "x=1")) // "{\"text\":\"x=1\",\"start\":0,\"end\":3,\"groups\":[\"x=1\",\"x\",\"1\"],\"named\":{\"k\":\"x\"}}"
```

## Applications

Parse a date into a record with named groups. A guard rejects text that does not match:

```ant
type Date = { year: int, month: int, day: int }

func part(m: Regex.Match, name: string): int {
  (m.named[name] ?? "").toInt() ?? 0
}

func parseDate(s: string): Date | null {
  let m = Regex.find(/^(?<y>\d{4})-(?<m>\d\d)-(?<d>\d\d)$/, s)
  if m == null {
    return null
  }
  Date{ year: part(m, "y"), month: part(m, "m"), day: part(m, "d") }
}

println(parseDate("2026-11-12"))
println(parseDate("12.11.2026"))
```

```text
Date{ year: 2026, month: 11, day: 12 }
null
```

Collect `key=value` pairs into a map:

```ant
let text = "host=example.org port=8080 debug=on"
let config = Regex.findAll(/(\w+)=(\S+)/, text)
  | Array.map({ Entry{ key: x.groups[1] ?? "", value: x.groups[2] ?? "" } })
  | Map.fromEntries()
println(config)
```

```text
Map{ "host": "example.org", "port": "8080", "debug": "on" }
```

Mark where each match is, using `start` and `end`:

```ant
let line = "let x = 10 + 200"
var marks = ""
for m in Regex.findAll(/\d+/, line) {
  marks = marks.padEnd(m.start, " ") + "^".repeat(m.end - m.start)
}
println(line)
println(marks)
```

```text
let x = 10 + 200
        ^^   ^^^
```

Reorder names, with an optional middle initial that may take no part:

```ant
let names = "Doe, John; Roe, Jane Q."
let flipped = Regex.replaceWith(/(\w+), (\w+)( \w\.)?/, names, func(m: Regex.Match): string {
  let middle = m.groups[3] ?? ""
  "${m.groups[2] ?? ""}${middle} ${m.groups[1] ?? ""}"
})
println(flipped)
```

```text
John Doe; Jane Q. Roe
```

## Properties and relations

`groups[0]` is `text`, and slicing the searched string at `start` and `end` gives `text`:

```ant
let s = "naïve café, 2026"
for m in Regex.findAll(/\pL+|\d+/, s) {
  assert(m.groups[0] == m.text)
  assert(s.slice(m.start, m.end) == m.text)
  assert(m.end - m.start == m.text.length())
}
```

Every named group is also a numbered group, so `named` holds values that `groups` has too:

```ant
let m = Regex.find(/(?<a>\d)(\d)(?<b>\d)/, "123")
if m == null {
  return
}
assert(m.named["a"] == m.groups[1])
assert(m.named["b"] == m.groups[3])
assert(m.named.size() == 2)
```

[Regex.test](../test.md) is true exactly when [Regex.find](../find.md) gives a match:

```ant
for s in ["a1", "bc", ""] {
  assert(Regex.test(/\d/, s) == (Regex.find(/\d/, s) != null))
}
```

## Possible issues

`Regex.find` can return `null`, so a field read without narrowing is a compile error:

```ant
let m = Regex.find(/\d+/, "x42")
println(m.text) // error: Regex.Match | null has no field text: narrow the value first
```

A group is `string | null`, so it needs a default or a check before a string function:

```ant
let m = Regex.find(/(\d+)/, "x42")
if m != null {
  let n = m.groups[1].toInt() // error: string | null has no field or function toInt on every member: narrow the value first
}
```

```ant
let m = Regex.find(/(\d+)/, "x42")
if m != null {
  println((m.groups[1] ?? "").toInt())
}
```

```text
42
```

Indexing past the last group panics. Use `get` for a group that may not exist:

```ant
let m = Regex.find(/(a)/, "a")
m?.groups.get(2) // null
m?.groups[2]     // panic: index 2 out of range for an array of length 2
```

A misspelled name in `named` gives `null`, as a group that took no part does. Use `has` to tell them
apart:

```ant
let m = Regex.find(/(?<year>\d{4})/, "in 2026")
m?.named["yaer"]     // null
m?.named.has("yaer") // false
m?.named.has("year") // true
```

`start` and `end` count code points, so they do not match what a reader sees as characters when the
text has combining marks. `"e\u{301}"` is one character on screen but two code points:

```ant
let m = Regex.find(/\d/, "e\u{301}1")
m?.start // 2
```

## See also

[Regex.find](../find.md) · [Regex.findAll](../findAll.md) · [Regex.replaceWith](../replaceWith.md) ·
[Regex.Pattern](Pattern.md) · [String.slice](../../String/slice.md) ·
[String.graphemes](../../String/graphemes.md)
