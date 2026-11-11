<!-- status: reviewed -->

# Comparison operators

`<`, `<=`, `>`, and `>=` order two numbers or two strings and give a `bool`.

## Details

- Both operands must have one type: a single numeric kind such as `int` or `float64`, `string`, or
  one named primitive type over such a base. Anything else is a compile error.
- Strings compare by code point, not by locale or by case: `"Z" < "a"` is `true`.
- A comparison with a `NaN` operand is `false`, and `-0.0` is not less than `0.0`.
- Comparisons bind tighter than `==` and `!=`, and looser than arithmetic, so `a + 1 < b` needs no
  parentheses.

## Basic examples

Compare two numbers:

```ant
3 < 5      // true
5 <= 5     // true
2.5 > 10.0 // false
```

Compare two strings:

```ant
"apple" >= "banana" // false
"pear" > "fig"      // true
```

## Scope

### Every numeric kind

Both operands share one kind, whatever its width. A literal is an `int` or a `float64`, so convert
it to compare with a sized value:

```ant
let small = int8(3)
let ratio = float32(0.5)
small < int8(4)         // true
ratio >= float32(0.25)  // true
uint8(200) > uint8(100) // true
```

### Strings by code point

Ordering follows code points, so a prefix comes first, digits are not read as numbers, and an
accented letter sorts after `z`. A decomposed `é`, an `e` followed by a combining accent, is not
equal to the composed one and sorts before it:

```ant
"" < "a"          // true
"ab" < "abc"      // true
"Zebra" < "apple" // true
"10" < "9"        // true
"é" > "z"         // true
"e\u{301}" < "é"  // true
```

### Named primitive types

Two values of one named primitive type compare as their base, and the result is a plain `bool`:

```ant
type Meters = float64
Meters(3.0) < Meters(4.5) // true
```

### Generic code

A type parameter bounded by `Ordered`, the numbers and `string`, takes the comparison operators:

```ant
func bigger[T: Ordered](a: T, b: T): T {
  if a > b { a } else { b }
}

println(bigger(3, 7))
println(bigger("pear", "fig"))
```

```text
7
pear
```

### Unions after narrowing

A union operand compares once `is` or a `null` check narrows it to one kind:

```ant
let input: int | string = 42
let limit: int | null = 10
if input is int && limit != null && input > limit {
  println("over the limit")
}
```

```text
over the limit
```

### Precedence

Arithmetic binds tighter, and the result of a comparison feeds `==`, `&&`, and `||`:

```ant
1 + 2 < 2 * 2    // true
1 < 2 == 3 < 4   // true
0 < 5 && 5 <= 10 // true
```

## Applications

### A range check

Write a bounded test as two comparisons joined by `&&`:

```ant
func inRange(x: int, lo: int, hi: int): bool { lo <= x && x < hi }

inRange(5, 0, 10)  // true
inRange(10, 0, 10) // false
```

### Filter by a threshold

A comparison in a lambda picks the elements past a threshold, finds the first one, or tests them
all:

```ant
let scores = [72, 45, 91, 50]
scores.filter({ x >= 50 }) // [72, 91, 50]
scores.find({ x > 90 })    // 91
scores.every({ x > 40 })   // true
```

### The larger value or a value within a range

`Math.max` and `Math.min` pick one of two values, and `Math.clamp` holds a value within a range, so
they replace a hand-written `if a > b { a } else { b }`:

```ant
Math.max(3, 7)          // 7
Math.min("pear", "fig") // "fig"
Math.clamp(150, 0, 100) // 100
```

### Order records by a field

Records have no `<`; compare their fields:

```ant
type Version = { major: int, minor: int }

func Version.isOlder(a: Version, b: Version): bool {
  if a.major != b.major { a.major < b.major } else { a.minor < b.minor }
}

let v1 = Version{ major: 1, minor: 2 }
let v2 = Version{ major: 1, minor: 10 }
println(v1.isOlder(v2))
```

```text
true
```

## Properties and relations

`a > b` is `b < a`, and for numbers other than `NaN`, `!(a < b)` is `a >= b`:

```ant
let pairs = [[1, 2], [2, 2], [3, 2]]
for p in pairs {
  let a = p[0]
  let b = p[1]
  assert((a > b) == (b < a))
  assert(!(a < b) == (a >= b))
  assert((a <= b) == (a < b || a == b))
}
```

## Possible issues

### Mixed kinds do not compare

There is no implicit conversion, so an `int` and a `float64` differ:

```ant
let count = 3
let average = 2.5
let above = count > average // error: > needs both operands of one type, got int and float64
```

Convert one side first:

```ant
let count = 3
let average = 2.5
float64(count) > average // true
```

A named primitive type does not mix with its base either:

```ant
type Meters = float64
let short = Meters(1.0) < 2.0 // error: < needs both operands of one type, got Meters and float64
```

### Comparisons do not chain

`a < b < c` compares a `bool` with a number. Join two comparisons with `&&` instead:

```ant
let x = 2
let ok = 1 < x < 3 // error: < needs both operands of one type, got bool and int
```

### A union or a nullable value needs narrowing

`Number`, `int | null`, and other unions are not one kind:

```ant
func smaller(a: Number, b: Number): bool { a < b }
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Number; narrow first
let limit: int | null = 10
let over = limit > 5 // error: > needs both operands of one type, got int | null and int
```

Supply a default with `??`, or test for `null` first:

```ant
let limit: int | null = 10
(limit ?? 0) > 5 // true
```

### Records, arrays, and time values have no order

Only numbers and strings compare with `<`:

```ant
let first = [1, 2] < [1, 3]
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Array[int]; narrow first
let earlier = Time.fromEpochMillis(0) < Time.fromEpochMillis(1)
// error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.Instant; narrow first
```

Use `Time.compare` for time values, and `Array.sortBy` or `Array.sortWith` to order records.

### `NaN` and negative zero

Every comparison with `NaN` is `false`, so `!(x < y)` does not imply `x >= y`. `-0.0` and `0.0` are
equal, so neither is less than the other:

```ant
Math.NaN < 1.0        // false
Math.NaN >= 1.0       // false
Math.isNaN(0.0 / 0.0) // true
-0.0 < 0.0            // false
```

### Case and digits in strings

Code point order puts every uppercase ASCII letter before every lowercase one, and compares digit
strings character by character. Fold the case or parse the number when that is the order you want:

```ant
"apple" < "Banana"                       // false
"apple".foldCase() < "Banana".foldCase() // true
"10" < "9"                               // true
```

## See also

- [Equality operators](equality.md)
- [is](is.md)
- [&&, ||, and !](logical.md)
- [Arithmetic operators](arithmetic.md)
- [?? operator](null-coalescing.md)
- [Math.isNaN](../../lib/Math/isNaN.md)
- [Math.max](../../lib/Math/max.md)
- [Math.min](../../lib/Math/min.md)
- [Math.clamp](../../lib/Math/clamp.md)
- [Array.sortBy](../../lib/Array/sortBy.md)
- [Array.sortWith](../../lib/Array/sortWith.md)
- [Time.compare](../../lib/Time/compare.md)
- [String.foldCase](../../lib/String/foldCase.md)
