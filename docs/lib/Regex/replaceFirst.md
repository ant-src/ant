<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.replaceFirst

```ant signature
func Regex.replaceFirst(p: Regex.Pattern, s: string, replacement: string): string
```

Regex.replaceFirst returns s with the first match of p, the first that [`Regex.findAll`](findAll.md)
returns, replaced by replacement, expanded as in [`Regex.replace`](replace.md).

<!-- end generated -->

## Details

- Only the leftmost match is replaced; the rest of `s` comes back as it was.
- When `p` does not match, the result equals `s`.
- `replacement` expands `$0` (the whole match), `$1`, `${1}`, `${name}`, and `$$`. A group that took
  no part, or that does not exist, expands to `""`.
- In a `"..."` string, `${` interpolates. Write `"\${1}"` or a backtick string such as `` `${1}` ``.

## Basic examples

Replace the first run of digits:

```ant
Regex.replaceFirst(/\d+/, "a1b22c333", "#") // "a#b22c333"
```

Rearrange the groups of the first match only:

```ant
Regex.replaceFirst(/(\w+)@(\w+)/, "ann@home bob@work", "$2:$1") // "home:ann bob@work"
```

When nothing matches, the string comes back unchanged:

```ant
Regex.replaceFirst(/x/, "abc", "y") // "abc"
```

## Scope

Named groups expand with `${name}`. A backtick string keeps `${` from interpolating:

```ant
Regex.replaceFirst(/(?<user>\w+)@(?<host>\w+)/, "ann@home bob@work", `${host}/${user}`)
// "home/ann bob@work"
```

`$0` is the whole match:

```ant
Regex.replaceFirst(/\d+/, "a1b22", "[$0]") // "a[1]b22"
```

`$$` writes one `$`:

```ant
Regex.replaceFirst(/\d+/, "cost 5", "$$") // "cost $"
Regex.replaceFirst(/\$/, "a$b", "$$$$")   // "a$$b"
```

A group that took no part expands to `""`:

```ant
Regex.replaceFirst(/(a)|(b)/, "b", "[$1]") // "[]"
```

A pattern that matches the empty string replaces at the start of `s`, and `^` inserts a prefix:

```ant
Regex.replaceFirst(/x*/, "abc", "-") // "-abc"
Regex.replaceFirst(/^/, "abc", "> ") // "> abc"
```

An empty `s` gives `""`, unless the pattern matches the empty string:

```ant
Regex.replaceFirst(/a/, "", "b")  // ""
Regex.replaceFirst(/x*/, "", "b") // "b"
```

Flags apply as in any match. With `i`, the first `É` matches `é`:

```ant
Regex.replaceFirst(/é/i, "CAFÉ café", "e") // "CAFe café"
```

Unicode text works like any other:

```ant
Regex.replaceFirst(/🐜/, "🐜🐝🐜", "🐞") // "🐞🐝🐜"
```

Call it through `.` on a pattern, or with named arguments:

```ant
let digits = /\d+/
digits.replaceFirst("a1b22", "#")                      // "a#b22"
Regex.replaceFirst(p: /o/, s: "foo", replacement: "0") // "f0o"
```

## Applications

Reformat the first date in a log line and leave later dates alone:

```ant
let line = "2026-11-12 saved 2026-11-01 backup"
Regex.replaceFirst(/(\d{4})-(\d{2})-(\d{2})/, line, "$3.$2.$1")
// "12.11.2026 saved 2026-11-01 backup"
```

Bump the first `version` line of a config file, not the one in a later section:

```ant
let config = "name = \"app\"\nversion = \"1.2.3\"\n[dep]\nversion = \"0.9.0\"\n"
print(Regex.replaceFirst(/(?m)^version = ".*"$/, config, `version = "1.3.0"`))
```

```text
name = "app"
version = "1.3.0"
[dep]
version = "0.9.0"
```

Strip leading whitespace, which only the first match can be:

```ant
Regex.replaceFirst(/^\s+/, "  hi  there", "") // "hi  there"
```

## Properties and relations

When `p` matches, the result is the text before the first match, the expanded replacement, and the
text after it, as [`Regex.find`](find.md) locates it:

```ant
let s = "x=1, y=22"
let m = Regex.find(/\d+/, s)
assert(m != null)
let expected = s.slice(0, m.start) + "#" + s.slice(m.end, s.length())
assert(Regex.replaceFirst(/\d+/, s, "#") == expected)
```

When `p` does not match, `s` comes back unchanged:

```ant
let s = "no digits"
assert(!Regex.test(/\d/, s))
assert(Regex.replaceFirst(/\d/, s, "#") == s)
```

When `p` matches at most once, it agrees with [`Regex.replace`](replace.md):

```ant
for s in ["id: 42", "id: none", ""] {
  assert(Regex.findAll(/\d+/, s).length() <= 1)
  assert(Regex.replaceFirst(/\d+/, s, "N") == Regex.replace(/\d+/, s, "N"))
}
```

## Possible issues

`$1x` reads the group name `1x`, which does not exist, so it expands to `""`. Write `${1}x`:

```ant
Regex.replaceFirst(/(\w+)/, "ab cd", "$1x")    // " cd"
Regex.replaceFirst(/(\w+)/, "ab cd", "\${1}x") // "abx cd"
```

In a `"..."` string, `${...}` interpolates before `Regex.replaceFirst` sees it:

```ant
let host = "?"
Regex.replaceFirst(/(?<host>\w+)/, "ann", "${host}") // "?"
Regex.replaceFirst(/(?<host>\w+)/, "ann", `${host}`) // "ann"
```

The pattern comes first, so a string cannot be the receiver of a dot call or the start of a pipe. On
a string, `.replaceFirst` is [`String.replaceFirst`](../String/replaceFirst.md):

```ant
let s = "a.b.c"
let fixed = s.replaceFirst(/\./, "!") // error: cannot use Regex.Pattern where string is expected
```

In a pattern, `.` matches any character, so `/./` replaces the first character. Escape it as `\.`,
or use [`String.replaceFirst`](../String/replaceFirst.md) to replace a fixed string.
[`Regex.escape`](escape.md) does the escaping for a string from data:

```ant
Regex.replaceFirst(/./, "a.b", "!")  // "!.b"
Regex.replaceFirst(/\./, "a.b", "!") // "a!b"
"a.b".replaceFirst(".", "!")         // "a!b"
```

## See also

[`Regex.replace`](replace.md) · [`Regex.replaceWith`](replaceWith.md) · [`Regex.find`](find.md) ·
[`Regex.escape`](escape.md) · [`String.replaceFirst`](../String/replaceFirst.md)
