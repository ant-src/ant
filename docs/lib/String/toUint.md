<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.toUint

```ant signature
func String.toUint(s: string, base: int = 10): uint64 | null
```

String.toUint parses s as an unsigned integer in base, or returns null when it is not one. s is
digits of base as for [`String.toInt`](toInt.md), with no sign. A number that does not fit a uint64
is null too. base defaults to 10, and it panics if base is outside 2 ..= 36.

<!-- end generated -->

## Details

- `String.toUint` reads the digits that [`String.toInt`](toInt.md) reads: `0` to `9`, and for a base
  above ten the letters `a` to `z` in either case. It takes no sign, not even `+`, and no prefix
  such as `0x`. Anything else, the empty string included, gives `null`
  ([spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- The result covers the whole `uint64` range, from `0` up to `Math.maxUint64`. A larger number is
  `null`.
- `base` defaults to `10`. A base outside `2 ..= 36` panics.
- The result is `uint64 | null`. Narrow it with `??`, `?? return`, or `if` before you use it as a
  number ([spec: Narrowing](https://github.com/ant-src/spec/blob/master/ant-lang.md#narrowing)).
- For a narrower kind, such as `byte` or `uint16`, convert the result. The conversion panics when
  the value is out of range
  ([spec: Primitives](https://github.com/ant-src/spec/blob/master/ant-lang.md#primitives)).

## Basic examples

Parse a decimal number:

```ant
String.toUint("42") // 42
```

The largest `uint64` parses, though it does not fit an `int`:

```ant
String.toUint("18446744073709551615") // 18446744073709551615
```

A sign is not a digit, so a negative number is `null`:

```ant
String.toUint("-1") // null
```

## Scope

Pass a base from 2 to 36. Above ten, the letters `a` to `z` are digits, in either case:

```ant
String.toUint("101", 2)               // 5
String.toUint("755", 8)               // 493
String.toUint("ff", 16)               // 255
String.toUint("FF", 16)               // 255
String.toUint("ffffffffffffffff", 16) // 18446744073709551615
String.toUint("zz", 36)               // 1295
```

A digit that is not one of the base is `null`, and so is a number one past `Math.maxUint64`:

```ant
String.toUint("12", 2)                // null
String.toUint("g", 16)                // null
String.toUint("18446744073709551616") // null
```

Leading zeros are digits like any other:

```ant
String.toUint("007") // 7
String.toUint("0")   // 0
```

No sign, no prefix, no space, no `_` separator, and no empty string:

```ant
String.toUint("+1")    // null
String.toUint("-0")    // null
String.toUint("0x10")  // null
String.toUint(" 1")    // null
String.toUint("1_000") // null
String.toUint("")      // null
```

Only ASCII digits count. A digit of another script, or a full-width one, is `null`:

```ant
String.toUint("٣") // null
String.toUint("１") // null
```

Call it through `.`, in a pipe, with named arguments, or pass it as a function value:

```ant
"ff".toUint(16)                                 // 255
"42".toUint()                                   // 42
"777" | String.toUint(base: 8)                  // 511
String.toUint(s: "1010", base: 2)               // 10
Array.map(["1", "x", "30"], String.toUint)      // [1, null, 30]
Array.map(["a", "b"], { String.toUint(x, 16) }) // [10, 11]
```

## Named arguments

### base

`base` picks the radix, from 2 to 36, and defaults to `10`. The same text reads differently in each:

```ant
String.toUint("100")           // 100
String.toUint("100", base: 2)  // 4
String.toUint("100", base: 16) // 256
```

## Applications

Read a hex color into its three channels. `?? return null` ends the function when the text is not
hex:

```ant
func rgb(hex: string): Array[byte] | null {
  if String.length(hex) != 6 {
    return null
  }
  var n = String.toUint(hex, 16) ?? return null
  [
    byte(Bits.shiftRight(n, 16)),
    byte(Bits.and(Bits.shiftRight(n, 8), uint64(255))),
    byte(Bits.and(n, uint64(255))),
  ]
}

println(rgb("ff8000"))
println(rgb("00zz00"))
```

```text
[255, 128, 0]
null
```

Parse a 64-bit identifier written in hex, such as a trace ID, whose high bit is set and so does not
fit an `int`:

```ant
var id = String.toUint("9f86d081884c7d65", 16) ?? panic("not hex")
println(id)
println(String.fromInt(id, 16))
```

```text
11495104353665842533
9f86d081884c7d65
```

Read a port number, checking its range before converting to `uint16`:

```ant
func port(s: string): uint16 | null {
  var n = String.toUint(s) ?? return null
  if n > uint64(65535) {
    return null
  }
  uint16(n)
}

println(port("8080"))
println(port("70000"))
println(port("-1"))
```

```text
8080
null
null
```

Read a Unix permission mode in octal, the way `chmod` takes it:

```ant
for mode in ["755", "640", "abc"] {
  var bits = String.toUint(mode, 8)
  if bits == null {
    println("${mode}: not octal")
  } else {
    println("${mode}: ${bits}")
  }
}
```

```text
755: 493
640: 416
abc: not octal
```

## Properties and relations

[`String.fromInt`](fromInt.md) writes what `String.toUint` reads, in every base:

```ant
for n in [uint64(0), uint64(1), uint64(255), Math.maxUint64] {
  for base in [2, 8, 10, 16, 36] {
    assert(String.toUint(String.fromInt(n, base), base) == n)
  }
}
```

On text without a sign, `String.toUint` agrees with [`String.toInt`](toInt.md) wherever the number
fits an `int`:

```ant
for s in ["0", "42", "9223372036854775807"] {
  var u = String.toUint(s) ?? panic("unreachable")
  assert(String.toInt(s) == int(u))
}
```

Above `Math.maxInt`, only `String.toUint` gives a number:

```ant
String.toInt("9223372036854775808")  // null
String.toUint("9223372036854775808") // 9223372036854775808
```

## Possible issues

The result may be `null`, so it is not a `uint64` until you narrow it:

<!-- test: error cannot use uint64 | null where uint64 is expected -->

```ant
var n: uint64 = String.toUint("5")
```

An integer literal is always an `int`, so `?? 0` gives `uint64 | int`, and arithmetic on it is a
compile error. Write the default as a `uint64`:

<!-- test: error + needs both operands of one type -->

```ant
var n = String.toUint("5") ?? 0
var m = n + 1
```

```ant
var n = String.toUint("5") ?? uint64(0)
println(n + uint64(1))
```

```text
6
```

Converting the result to a narrower kind panics when it does not fit. Check the range first, as
`port` does above:

<!-- test: panics 300 out of range for byte -->

```ant
var n = String.toUint("300") ?? panic("not a number")
println(byte(n))
```

A base outside `2 ..= 36` panics, even for text that would parse:

```ant
String.toUint("1", 1)  // panic: String.toUint needs a base from 2 to 36, got 1
String.toUint("1", 37) // panic: String.toUint needs a base from 2 to 36, got 37
```

## See also

- [`String.toInt`](toInt.md): parse a signed `int`.
- [`String.toFloat`](toFloat.md): parse a decimal number.
- [`String.fromInt`](fromInt.md): write an integer in a base.
- [`String.isDigit`](isDigit.md): test for decimal digits of any script.
- [`Math.toInt`](../Math/toInt.md): a `float64` as an `int`, or `null`.
