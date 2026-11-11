<!-- status: reviewed -->

# Raw strings

Backticks write a raw string: `` `C:\temp` `` is a `string` whose value is exactly the text between
the backticks. A raw string has no escapes and no interpolation.

## Details

- `\`, `"`, and `${` stand for themselves, so a raw string suits regular expression sources, Windows
  paths, and templates for other tools.
- A raw string may span lines. Each line break is one `\n` in the value, whatever line endings the
  source file uses, and the indentation of the following lines is part of the value.
- A raw string cannot contain a backtick. Write that text as a string, or join raw strings with `+`.

## Basic examples

A backslash is a backslash:

```ant
`C:\temp`          // "C:\\temp"
`C:\temp`.length() // 7
```

Double quotes need no escape:

```ant
`say "hi"` // "say \"hi\""
```

`${` does not interpolate:

```ant
let name = "Ann"
`hello, ${name}` // "hello, \${name}"
"hello, ${name}" // "hello, Ann"
```

## Scope

### Empty

Two backticks make the empty string:

```ant
``           // ""
``.isEmpty() // true
```

### Escape sequences stay as written

`\n`, `\t`, and `\u{...}` are two or more characters, not an escape:

```ant
`\n`.length() // 2
"\n".length() // 1
`\u{1F600}`   // "\\u{1F600}"
"\u{1F600}"   // "😀"
```

### Unicode

Any character but a backtick goes in as written:

```ant
`naïve 😀`.length() // 7
```

### Several lines

A raw string spans lines. Each line break is one `\n`:

```ant
let poem = `roses
violets`
poem         // "roses\nviolets"
poem.lines() // ["roses", "violets"]
```

### A string like any other

A raw string compares, concatenates, and calls functions as any `string` does:

```ant
`a\b` == "a\\b"    // true
`a\b` + "/c"       // "a\\b/c"
`x,y,z`.split(",") // ["x", "y", "z"]
```

## Applications

### Regular expression sources

A fixed pattern reads best as a regex literal. A pattern built at run time goes to `Regex.compile`
as a string, and its raw parts keep their backslashes single, as in a regex literal:

```ant
let n = 4
let digits = Result.unwrap(Regex.compile(`^\d{` + "${n}" + `}$`))
digits.test("2026") // true
digits.test("26")   // false
```

A pattern full of `/` reads more plainly as a raw string than as a literal, which needs `\/` for
each one:

```ant
let origin = Result.unwrap(Regex.compile(`^https?://[^/"]+/`))
origin.find("see https://ant-lang.org/docs here")?.text // null
origin.find("https://ant-lang.org/docs")?.text          // "https://ant-lang.org/"
```

### Windows paths

```ant
let dir = `C:\Users\ann\Documents`
dir.split(`\`) // ["C:", "Users", "ann", "Documents"]
```

A raw string has no `${...}`, so join it with `+` to put a value in. A backslash right before the
closing backtick is just a backslash:

```ant
let user = "ann"
`C:\Users\` + user // "C:\\Users\\ann"
```

### Templates for other tools

A shell script keeps its own `${...}` for the shell, and ends with a line break:

```ant
let script = `#!/bin/sh
echo "deploying ${APP}"
`
print(script)
```

```text
#!/bin/sh
echo "deploying ${APP}"
```

JSON with quotes inside reads as written:

```ant
let text = `{"name": "Ann", "tags": ["admin"]}`
let value = Result.unwrap(Json.parse(text))
println(value)
```

```text
Map{ "name": "Ann", "tags": ["admin"] }
```

## Properties and relations

A raw string equals the string with each `\` and `"` escaped, `${` written `\${`, and each line
break written `\n`:

```ant
assert(`a\b "c" ${d}` == "a\\b \"c\" \${d}")
assert(`one
two` == "one\ntwo")
```

## Possible issues

### A backtick ends the string

A backslash does not escape a backtick. The second backtick ends the string, and the third starts a
new one that is never closed:

```ant
let quote = `it\`s` // error: unterminated raw string
```

Join raw strings with a quoted backtick instead:

```ant
`it` + "`" + `s` // "it`s"
```

Or write the text as a string, where a backtick needs no escape:

```ant
"use `ant fmt`" // "use `ant fmt`"
```

A line break does not end a raw string either, so a missing closing backtick is the same error:

```ant
let path = `C:\temp // error: unterminated raw string
```

### Indentation is part of the value

A multiline raw string inside a block keeps the spaces in front of each line:

```ant
func usage(): string {
  `usage:
    ant run <file>`
}
usage() // "usage:\n    ant run <file>"
```

## See also

[String literals](strings.md) · [Regular expression literals](regex.md) ·
[string](../types/string.md) · [String functions](../../lib/String/index.md) ·
[Regex.compile](../../lib/Regex/compile.md) · [Strings guide](../../guide/strings.md)
