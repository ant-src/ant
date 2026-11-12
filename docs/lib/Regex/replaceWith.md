<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.replaceWith

```ant signature
func Regex.replaceWith(p: Regex.Pattern, s: string, f: func(Regex.Match): string): string
```

Regex.replaceWith returns s with every match of p, as [`Regex.findAll`](findAll.md) returns them,
replaced by what f returns for it. f is called with each match in order, and its result is used as
written, with no $ expansion.

<!-- end generated -->

## Details

- When `p` never matches, the result is `s` and `f` is not called.
- `$` in the result of `f` is an ordinary character. Read groups from the `Regex.Match` instead.

## Basic examples

Uppercase every run of letters:

```ant
Regex.replaceWith(/[a-z]+/, "ab1cd", { x.text.toUpper() }) // "AB1CD"
```

Double every number in a sentence:

```ant
let text = "3 apples, 12 pears"
Regex.replaceWith(/\d+/, text, { "${(x.text.toInt() ?? 0) * 2}" }) // "6 apples, 24 pears"
```

## Scope

A pattern that never matches returns `s` as it is:

```ant
Regex.replaceWith(/\d/, "abc", { "#" }) // "abc"
```

Read a numbered group from `x.groups`. A group that takes no part in a match is `null` there:

```ant
let at = /(\w+)@(\w+)/
Regex.replaceWith(at, "ann@home", { "${x.groups[2] ?? ""}:${x.groups[1] ?? ""}" }) // "home:ann"
Regex.replaceWith(/a|(b)/, "ab", { x.groups[1] ?? "?" })                           // "?b"
```

Read a named group from `x.named`:

```ant
let pair = /(?<key>\w+)=(?<value>\w+)/
Regex.replaceWith(pair, "a=1 b=2", { "${x.named["value"] ?? ""}=${x.named["key"] ?? ""}" }) // "1=a 2=b"
```

`x.start` and `x.end` are code point indices into `s`:

```ant
Regex.replaceWith(/日本/, "日本語 日本", { "[${x.start}, ${x.end})" }) // "[0, 2)語 [4, 6)"
```

A pattern that matches the empty string matches between the characters too, and at both ends:

```ant
Regex.replaceWith(/x*/, "abc", { "-" }) // "-a-b-c-"
```

`f` sees the matches [`Regex.findAll`](findAll.md) returns, which skip an empty match right after
another match, so there is none between `aaa` and the last `b`:

```ant
Regex.replaceWith(/a*/, "baaab", { "<${x.text}>" }) // "<>b<aaa>b<>"
```

Pass a named function instead of a lambda:

```ant
func shout(m: Regex.Match): string { "${m.text.toUpper()}!" }
Regex.replaceWith(/\bno\b/, "no, no way", shout) // "NO!, NO! way"
```

Call it through `.` on a pattern:

```ant
let digits = /\d+/
digits.replaceWith("a1b22", { "#" }) // "a#b#"
```

## Named arguments

Name the arguments to pass them in any order:

```ant
Regex.replaceWith(s: "a1b2", f: { "<${x.text}>" }, p: /\d/) // "a<1>b<2>"
```

## Applications

Turn `snake_case` into `camelCase`:

```ant
Regex.replaceWith(/_([a-z])/, "line_width", { (x.groups[1] ?? "").toUpper() }) // "lineWidth"
```

Capitalize each word:

```ant
Regex.replaceWith(/\b\w/, "the quick fox", { x.text.toUpper() }) // "The Quick Fox"
```

Fill a template from a map, and leave unknown placeholders as they are:

```ant
let vars = Map{ "name": "Ann", "city": "Oslo" }
let filled = Regex.replaceWith(/\{(\w+)\}/, "Hi {name} from {city}, {zip}", {
  vars[x.groups[1] ?? ""] ?? x.text
})
println(filled)
```

```text
Hi Ann from Oslo, {zip}
```

Number the placeholders of a query. `f` runs once per match, in order, so a counter it updates
counts the matches:

```ant
var n = 0
Regex.replaceWith(/\?/, "a = ? and b = ?", { n += 1; "$${n}" }) // "a = $1 and b = $2"
```

## Properties and relations

Replacing each match with its own text gives `s` back:

```ant
let s = "a1b22c"
assert(Regex.replaceWith(/\d+/, s, { x.text }) == s)
```

With a constant result that holds no `$`, it is [`Regex.replace`](replace.md):

```ant
let s = "a1b22c"
assert(Regex.replaceWith(/\d+/, s, { "#" }) == Regex.replace(/\d+/, s, "#"))
```

`f` sees the matches from left to right. A callback that logs what it sees shows the order:

```ant
let seen: Array[string] = []
Regex.replaceWith(/\d+/, "x1y22z333", { seen.push(x.text); "" }) // "xyz"
seen                                                             // ["1", "22", "333"]
```

## Possible issues

The result of `f` is used as written, so `$1` is not a group reference here, as it is in
[`Regex.replace`](replace.md). Read the group from `x.groups`:

```ant
Regex.replaceWith(/(\d)/, "a1b2", { "<$1>" })                   // "a<$1>b<$1>"
Regex.replaceWith(/(\d)/, "a1b2", { "<${x.groups[1] ?? ""}>" }) // "a<1>b<2>"
```

An element of `x.groups` is `string | null`, so a function on `string` needs the `null` handled
first:

```ant
let camel = Regex.replaceWith(/_([a-z])/, "a_b", { x.groups[1].toUpper() })
// error: string | null has no field or function toUpper on every member: narrow the value first
```

`x.groups` holds one element per group of the pattern, plus the whole match. A higher index panics:

```ant
let text = Regex.replaceWith(/(a)/, "a", { x.groups[2] ?? "" })
// panic: index 2 out of range for an array of length 2
```

`f` returns a string, not the match:

```ant
let same = Regex.replaceWith(/\d/, "a1", { x })
// error: cannot use Regex.Match where string is expected
```

The pattern is the first parameter, so a pipe fills `p`, not `s`. Piping `s` is a compile error,
reported at both arguments:

```ant
let s = "a1b2"
let marked = s | Regex.replaceWith(/\d/, { "<${x.text}>" })
// error: cannot use string where Regex.Pattern is expected
// error: cannot use Regex.Pattern where string is expected
```

## See also

[Regex.replace](replace.md), [Regex.replaceFirst](replaceFirst.md), [Regex.findAll](findAll.md),
[Regex.Match](types/Match.md), [String.replace](../String/replace.md)
