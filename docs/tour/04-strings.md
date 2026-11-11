<!-- status: reviewed -->

# Strings

A string is text between double quotes. It can hold any expression inside `${...}`, span several
lines, and hold any Unicode text. Run this program, then change the name or the count.

```ant
let name = "Ada"
let count = 3

println("Hello, ${name}. You have ${count * 2} new messages.")

let card = "Name:  ${name}
Count: ${count}"
println(card)

let path = `C:\Users\${name}`
println(path)

let word = "añt 🐜"
println(len(word))
for c in word {
  print("[${c}]")
}
println("")
println(word.toUpper())
```

```text
Hello, Ada. You have 6 new messages.
Name:  Ada
Count: 3
C:\Users\${name}
5
[a][ñ][t][ ][🐜]
AÑT 🐜
```

## Interpolation

`${...}` takes any expression, not only a name, and renders its value as text, whatever its type.

```ant
let xs = [1, 2, 3]
"sum: ${xs[0] + xs[1] + xs[2]}" // "sum: 6"
"items: ${xs}"                  // "items: [1, 2, 3]"
"\${n}".chars()                 // ["$", "{", "n", "}"]
"$5"                            // "$5"
```

A `$` needs the backslash only when a `{` follows it.

## Multiline and raw strings

A line break inside the quotes is part of the string, and so is the indentation of the next line.
That is why `card` above prints its second line flush left: it is written flush left.

A string between backticks is raw. It has no escapes and no interpolation, so backslashes and `${`
stay as written. Use it for regular expressions, Windows paths, and shell scripts. A raw string
cannot hold a backtick. Write text that has one as a double-quoted string.

```ant
"a\tb"      // "a\tb"
`a\tb`      // "a\\tb"
len("a\tb") // 3
len(`a\tb`) // 4
```

## Unicode

`len`, indices, and slices count code points, not bytes, so `ñ` and `🐜` count as one each. There is
no character type: `for` gives each character as a string of one code point. `String.toBytes` gives
the UTF-8 bytes when you need them.

```ant
len("añt")         // 3
"añt".chars()      // ["a", "ñ", "t"]
"añt".toBytes()    // Bytes(61 c3 b1 74)
"straße".toUpper() // "STRASSE"
```

Changing case can change the length: `String.toUpper` turns `ß` into `SS`.

What a reader sees as one character can be several code points. A family emoji is three people
joined by two invisible joiners. `String.graphemes` splits text the way a reader sees it.

```ant
let family = "👨‍👩‍👧"
len(family)             // 5
len(family.graphemes()) // 1
```

## Learn more

- [Strings](../guide/strings.md) in the language guide: Unicode, bytes, and text from outside.
- [String literals](../language/literals/strings.md) and
  [raw strings](../language/literals/raw-strings.md): every escape.
- [String](../lib/String/index.md): the functions on strings, such as
  [String.chars](../lib/String/chars.md), [String.toUpper](../lib/String/toUpper.md),
  [String.graphemes](../lib/String/graphemes.md), and [String.toBytes](../lib/String/toBytes.md).
- [len](../lib/prelude/len.md) and [for](../language/keywords/for.md).
- [Bytes](../lib/Bytes/index.md): binary data.

Next: [Collections](05-collections.md).
