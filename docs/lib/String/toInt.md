<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toInt

```ant signature
func String.toInt(s: string, base: int = 10): int | null
```

String.toInt parses s as an integer in base, or returns null when it is not one. s is digits of base
with an optional sign and no prefix; only 0 to 9 and the letters a to z, in either case, count as
digits. A number that does not fit an int is null too. base defaults to 10, and it panics if base is
outside 2 ..= 36.

<!-- end generated -->

## Details

- Reads digits of `base` with an optional `+` or `-`. Above base ten the letters `a` to `z` count,
  in either case. Digits of other scripts, such as `٣`, do not.
- Anything else is `null`: an empty string, a lone sign, white space, `_`, `.`, an exponent, or a
  prefix such as `0x`. A number outside the range of `int` is `null` too.
- A `base` outside `2 ..= 36` panics.

## Basic examples

Parse a decimal number:

```ant
String.toInt("42") // 42
```

A string that is not a number gives `null`:

```ant
String.toInt("forty-two") // null
```

Pass a base to read other radixes. Letters count in either case:

```ant
String.toInt("-FF", 16) // -255
String.toInt("ff", 16)  // 255
```

## Scope

A sign is optional, and `+` is allowed. Leading zeros are fine:

```ant
String.toInt("+7")  // 7
String.toInt("-0")  // 0
String.toInt("007") // 7
```

Every base from 2 to 36 works. A digit must be below the base:

```ant
String.toInt("101", 2) // 5
String.toInt("777", 8) // 511
String.toInt("z", 36)  // 35
String.toInt("Z", 36)  // 35
String.toInt("2", 2)   // null
String.toInt("a")      // null
```

The whole string must be digits. Empty text, a lone sign, white space, separators, fractions, and
exponents are all `null`:

```ant
String.toInt("")      // null
String.toInt("-")     // null
String.toInt(" 42")   // null
String.toInt("42\n")  // null
String.toInt("1_000") // null
String.toInt("1.0")   // null
String.toInt("1e3")   // null
```

Only ASCII digits count. Digits of other scripts and full-width digits are `null`:

```ant
String.toInt("٣")  // null
String.toInt("０") // null
```

The whole range of `int` parses, and one past either end is `null`:

```ant
String.toInt("9223372036854775807")  // 9223372036854775807
String.toInt("9223372036854775808")  // null
String.toInt("-9223372036854775808") // -9223372036854775808
String.toInt("-9223372036854775809") // null
```

Call it through `.` on a string, in a pipe, or pass it as a function value:

```ant
"42".toInt()                      // 42
"ff" | String.toInt(base: 16)     // 255
["1", "x", "3"].map(String.toInt) // [1, null, 3]
```

## Named arguments

### base

`base` picks the radix, from 2 to 36, and defaults to `10`. Name it to make the call read clearly:

```ant
String.toInt("777", base: 8)     // 511
String.toInt("777")              // 777
String.toInt(s: "1010", base: 2) // 10
```

## Applications

Read a port from the command line, with a default when the argument is missing or not a number:

<!-- test: args 8080 -->

```ant
let args = Os.args()
let port = if len(args) > 0 { args[0].toInt() ?? 3000 } else { 3000 }
println("listening on ${port}")
```

```text
listening on 8080
```

Sum the numbers in a list of fields, skipping the ones that are not numbers:

```ant
let fields = ["12", "n/a", "30", "", "-2"]
var total = 0
for f in fields {
  let n = f.toInt()
  if n != null {
    total += n
  }
}
println(total)
```

```text
40
```

Split a CSS color such as `#ff8000` into its channels:

```ant
type Rgb = { r: int, g: int, b: int }

func parseColor(s: string): Rgb | null {
  if len(s) != 7 || !s.startsWith("#") {
    return null
  }
  let r = s.slice(1, 3).toInt(16) ?? return null
  let g = s.slice(3, 5).toInt(16) ?? return null
  let b = s.slice(5, 7).toInt(16) ?? return null
  Rgb{ r: r, g: g, b: b }
}

println(parseColor("#ff8000"))
println(parseColor("#FF80zz"))
```

```text
Rgb{ r: 255, g: 128, b: 0 }
null
```

Parse a version such as `1.12.3` and compare two of them:

```ant
func version(s: string): Array[int] | null {
  let parts: Array[int] = []
  for p in s.split(".") {
    let n = p.toInt() ?? return null
    parts.push(n)
  }
  parts
}

println(version("1.12.3"))
println(version("1.x"))
```

```text
[1, 12, 3]
null
```

## Properties and relations

`String.toInt` reads back what [`String.fromInt`](fromInt.md) writes, in the same base:

```ant
for n in [0, 1, -1, 255, -4096, Math.maxInt, Math.minInt] {
  for base in [2, 10, 16, 36] {
    assert(String.fromInt(n, base).toInt(base) == n)
  }
}
```

In base 10 it reads back what [`String.from`](from.md) and interpolation write:

```ant
for n in [0, 42, -7, Math.maxInt] {
  assert(String.from(n).toInt() == n)
  assert("${n}".toInt() == n)
}
```

Letters parse the same in either case, so the result does not change under
[`String.toUpper`](toUpper.md):

```ant
for s in ["beef", "C0FFEE", "-7f"] {
  assert(s.toInt(16) == s.toUpper().toInt(16))
}
```

## Possible issues

A base outside `2 ..= 36` panics. There is no base `0` that guesses the base from a prefix:

```ant
String.toInt("10", 0)  // panic: String.toInt needs a base from 2 to 36, got 0
String.toInt("10", 37) // panic: String.toInt needs a base from 2 to 36, got 37
```

A prefix such as `0x` or `0b` is not a number, even in the matching base. Strip it first with
[`String.trimPrefix`](trimPrefix.md):

```ant
String.toInt("0x1f", 16)                  // null
String.toInt("0x1f".trimPrefix("0x"), 16) // 31
```

White space around the digits makes the result `null`, which bites on a line read from input. Trim
it first with [`String.trim`](trim.md):

```ant
String.toInt(" 42\n")        // null
String.toInt(" 42\n".trim()) // 42
```

The result may be `null`, so using it as an `int` is a compile error:

```ant
let n = String.toInt("42")
println(n + 1) // error: + needs both operands of one type, got int | null and int
```

Narrow it, or give a default with [`??`](../../language/operators/null-coalescing.md):

```ant
let n = String.toInt("42")
if n != null { n + 1 }       // 43
(String.toInt("x") ?? 0) + 1 // 1
```

A number too large for `int` is `null`, not a rounded or wrapped value. Use
[`String.toUint`](toUint.md) for the `uint64` range above `Math.maxInt`:

```ant
String.toInt("18446744073709551615")  // null
String.toUint("18446744073709551615") // 18446744073709551615
```

A number with a fraction is `null`. Parse it with [`String.toFloat`](toFloat.md) and convert the
result with [`Math.toInt`](../Math/toInt.md) when it is whole:

```ant
String.toInt("2.0") // null
let f = String.toFloat("2.0")
if f != null { Math.toInt(f) } // 2
```

## See also

[String.toUint](toUint.md) · [String.toFloat](toFloat.md) · [String.fromInt](fromInt.md) ·
[String.from](from.md) · [String.isDigit](isDigit.md) · [String.trim](trim.md) ·
[Math.toInt](../Math/toInt.md) · [Null](../../guide/null.md)
