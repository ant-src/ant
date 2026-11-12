<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.replace

```ant signature
func Regex.replace(p: Regex.Pattern, s: string, replacement: string): string
```

Regex.replace returns s with every match of p replaced by replacement; the matches are the ones
[`Regex.findAll`](findAll.md) returns. In replacement, $n and $\{n\} expand to group n, $\{name\} to
a named group, $$ to one $, and a group that took no part to "". A $ that starts none of these, such
as one at the end, one before a character that cannot start a name, or a $\{ with no name or no \},
stays as written. A name after $ is taken as long as it can be, so write $\{1\}x, not $1x.

<!-- end generated -->

## Details

- `$` is special in `replacement`: write `$$` for a literal `$`, or use
  [`Regex.replaceWith`](replaceWith.md), which inserts its text as written.
- A group number or name that the pattern does not have expands to `""`; it is not an error.
- A `$` that starts no group reference, such as one at the end, stays as written:
  `Regex.replace(/b/, "ab", "$")` is `"a$"`.
- The matches replaced are the ones [`Regex.findAll`](findAll.md) returns, so an empty match right
  after another match is skipped: `Regex.replace(/a*/, "baaab", "-")` is `"-b-b-"`.

## Basic examples

Replace every match of a pattern:

```ant
Regex.replace(/\d+/, "a1b22c333", "#") // "a#b#c#"
```

Rearrange the groups of each match with `$1`, `$2`:

```ant
Regex.replace(/(\w+)@(\w+)/, "ann@home bob@work", "$2:$1") // "home:ann work:bob"
```

When the pattern does not match, the string comes back unchanged:

```ant
Regex.replace(/x/, "banana", "o") // "banana"
```

## Scope

`$0` is the whole match:

```ant
Regex.replace(/\d+/, "3 apples, 12 pears", "[$0]") // "[3] apples, [12] pears"
```

Braces end a group number, so text can follow it directly:

```ant
Regex.replace(/(\d+)/, "a1b22", `${1}x`)  // "a1xb22x"
Regex.replace(/(\d+)/, "a1b22", "\${1}x") // "a1xb22x"
```

A named group expands by its name, as `$name` or `${name}`:

```ant
let date = /(?<y>\d{4})-(?<m>\d\d)-(?<d>\d\d)/
Regex.replace(date, "2026-11-12", `${d}.${m}.${y}`)  // "12.11.2026"
Regex.replace(/(?<word>\w+)/, "hi there", "<$word>") // "<hi> <there>"
```

`$$` gives one `$`:

```ant
Regex.replace(/\d+/, "cost 5", "$$$0") // "cost $5"
```

A group that took no part in a match expands to `""`:

```ant
Regex.replace(/(a)|(b)/, "ab", "[$1|$2]") // "[a|][|b]"
```

A pattern that matches the empty string matches between the characters too:

```ant
Regex.replace(/x*/, "abc", "-")   // "-a-b-c-"
Regex.replace(/^/m, "a\nb", "> ") // "> a\n> b"
Regex.replace(/(?:)/, "", "x")    // "x"
Regex.replace(/./, "", "x")       // ""
```

Flags apply as in any match. `i` ignores case:

```ant
Regex.replace(/cat/i, "Cat cat CAT", "dog") // "dog dog dog"
```

`\d`, `\w`, and `\s` are ASCII; `\pL` matches a letter of any script:

```ant
Regex.replace(/\w+/, "naïve", "#")        // "#ï#"
Regex.replace(/\pL+/, "naïve ωμέγα", "#") // "# #"
```

The pattern is the first argument, so call it through `.` on a pattern, or with named arguments:

```ant
let vowels = /[aeiou]/
vowels.replace("banana", "_")                           // "b_n_n_"
Regex.replace(p: vowels, s: "banana", replacement: "_") // "b_n_n_"
```

## Applications

Collapse runs of white space into one space, then trim the ends:

```ant
Regex.replace(/\s+/, "  too   many\t spaces \n", " ").trim() // "too many spaces"
```

Turn a title into a URL slug:

```ant
func slug(title: string): string {
  let dashed = Regex.replace(/[^a-z0-9]+/, title.toLower(), "-")
  Regex.replace(/^-|-$/, dashed, "")
}

println(slug("Hello, World! It's 2026!"))
```

```text
hello-world-it-s-2026
```

Mask all but the last four digits of a card number:

```ant
Regex.replace(/\d(?:\d|\s)*(\d{4})/, "card 4111 1111 1111 1234", "**** $1")
// "card **** 1234"
```

Replace user-supplied text literally: [`Regex.escape`](escape.md) quotes the pattern, and doubling
every `$` quotes the replacement:

```ant
func replaceText(s: string, old: string, new: string): string {
  let p = Result.unwrap(Regex.compile(Regex.escape(old)))
  Regex.replace(p, s, new.replace("$", "$$"))
}

println(replaceText("price: 1.5 or 1x5", "1.5", "$2"))
```

```text
price: $2 or 1x5
```

## Properties and relations

Replacing every match with `$0` gives `s` back:

```ant
for s in ["", "a1b22", "no digits"] {
  assert(Regex.replace(/\d+/, s, "$0") == s)
}
```

When there is no match, `s` comes back unchanged:

```ant
let s = "banana"
assert(!Regex.test(/x/, s))
assert(Regex.replace(/x/, s, "y") == s)
```

`Regex.replace` agrees with [`Regex.replaceWith`](replaceWith.md) when the function builds the same
text from the match:

```ant
let p = /(\w+)@(\w+)/
let s = "ann@home bob@work"
assert(p.replace(s, "$2:$1") == p.replaceWith(s, { "${x.groups[2]}:${x.groups[1]}" }))
```

For a pattern that matches only literal text and a replacement with no `$`, it agrees with
[`String.replace`](../String/replace.md):

```ant
let s = "a.b.c"
assert(Regex.replace(/\./, s, "/") == s.replace(".", "/"))
```

[`Regex.replaceFirst`](replaceFirst.md) replaces only the first match:

```ant
Regex.replace(/\d/, "a1b2", "#")      // "a#b#"
Regex.replaceFirst(/\d/, "a1b2", "#") // "a#b2"
```

## Possible issues

A name after `$` runs as long as it can, so `$1x` asks for a group named `1x`, which does not exist,
and expands to `""`. `_` is part of a name too:

```ant
Regex.replace(/(\d+)/, "a1b22", "$1x")    // "ab"
Regex.replace(/(\w+)/, "hi there", "$1_") // " "
```

Write `${1}x` in a raw string instead, as in Scope.

In an ordinary string literal, `${name}` is interpolation, not a group reference:

```ant
let s = Regex.replace(/(?<y>\d+)/, "2026", "${y}!") // error: y is not defined
```

When a variable of that name exists, the code compiles and inserts its value, which is worse. Use a
raw string, `` `${y}!` ``.

The pattern comes first, so piping the string into `Regex.replace` does not compile:

```ant
let s = "foo"
println(s | Regex.replace(/o/, "0"))
// error: cannot use string where Regex.Pattern is expected
// error: cannot use Regex.Pattern where string is expected
```

Pass the string as the second argument, or call `replace` on the pattern: `/o/.replace(s, "0")`.

`Regex.replace` returns a new string and leaves `s` as it was:

```ant
let s = "a b"
Regex.replace(/ /, s, "_") // "a_b"
s                          // "a b"
```

`\d` matches only ASCII digits, and `i` folds case one code point at a time, so `ß` does not match
`SS`:

```ant
Regex.replace(/\d/, "1 ٣", "#")               // "# ٣"
Regex.replace(/straße/i, "STRASSE", "street") // "STRASSE"
```

## See also

[Regex.replaceFirst](replaceFirst.md) · [Regex.replaceWith](replaceWith.md) ·
[Regex.findAll](findAll.md) · [Regex.split](split.md) · [Regex.escape](escape.md) ·
[Regex.compile](compile.md) · [Regex.Match](types/Match.md) · [String.replace](../String/replace.md)
· [Regex literals](../../language/literals/regex.md) ·
[Raw strings](../../language/literals/raw-strings.md)
