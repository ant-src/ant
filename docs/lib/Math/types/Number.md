<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Number

```ant signature
type Number = Integer | float32 | float64
```

Number is any integer or floating-point type.

<!-- end generated -->

## Details

- `int` and `byte` are members, as other names for `int64` and `uint8`.
- Its main use is as a bound: `[T: Number]` lets one generic function take every numeric type. Each
  call fixes `T` to a single member, so all the `T` arguments share one type and a `T` result has it
  too.
- Inside such a function, the arithmetic and comparison operators apply to `T`. A literal stays an
  `int` or a `float64`, not a `T`.
- A value of type `Number` holds any number, but an operator or a function bounded by `Number` needs
  one numeric type: narrow the value with `is` first.
- A named primitive type over a number, such as `type Meters = float64`, is not a `Number`.

## Basic examples

A function bounded by `Number` takes any numeric type and returns that type:

```ant
func double[T: Number](x: T): T {
  x + x
}

double(21)   // 42
double(1.25) // 2.5
```

Prelude functions such as `Math.abs` and `Array.sum` use the same bound:

```ant
Math.abs(-7)           // 7
Array.sum([0.5, 0.25]) // 0.75
```

## Scope

Every integer and float type is a member, and the result keeps the argument's type:

```ant
func double[T: Number](x: T): T {
  x + x
}

double(byte(100))     // 200
double(int8(-60))     // -120
double(uint64(7))     // 14
double(float32(0.75)) // 1.5
```

An array parameter takes an array of any one numeric type:

```ant
func largest[T: Number](xs: Array[T]): T {
  var best = xs[0]
  for x in xs {
    if x > best {
      best = x
    }
  }
  best
}

largest([3, 9, 4])                     // 9
largest([float32(0.5), float32(-2.0)]) // 0.5
```

Inside the body, `is` tells the members apart:

```ant
func kind[T: Number](x: T): string {
  if x is float32 || x is float64 { "float" } else { "integer" }
}

kind(1)          // "integer"
kind(byte(1))    // "integer"
kind(1.5)        // "float"
kind(float32(2)) // "float"
```

As a value type, `Number` holds numbers of different types side by side:

```ant
let readings: Array[Number] = [1, 2.5, byte(3)]
readings // [1, 2.5, 3]
```

Narrow a `Number` value with `is` to compute with it:

```ant
func describe(n: Number): string {
  if n is int {
    "int ${n + 1}"
  } else if n is float64 {
    "float64 ${n * 2.0}"
  } else {
    "another number, ${n}"
  }
}

describe(41)      // "int 42"
describe(1.25)    // "float64 2.5"
describe(int8(5)) // "another number, 5"
```

## Applications

The mean of an array of any numeric type, converted to `float64` for the division:

```ant
func mean[T: Number](xs: Array[T]): float64 {
  var total = 0.0
  for x in xs {
    total += float64(x)
  }
  total / float64(xs.length())
}

mean([1, 2, 4])                    // 2.3333333333333335
mean([byte(10), byte(20)])         // 15.0
mean([float32(0.5), float32(1.5)]) // 1.0
```

Check that a value lies within a range, for any numeric type:

```ant
func within[T: Number](x: T, lo: T, hi: T): bool {
  x >= lo && x <= hi
}

within(5, 1, 10)                      // true
within(10.5, 0.0, 10.0)               // false
within(byte(255), byte(0), byte(255)) // true
```

A record generic over its number type:

```ant
type Range[T: Number] = { lo: T, hi: T }

func Range.width[T: Number](r: Range[T]): T {
  r.hi - r.lo
}

Range{ lo: 3, hi: 10 }.width()    // 7
Range{ lo: 0.5, hi: 2.0 }.width() // 1.5
```

## Properties and relations

A generic body gives each member type its own arithmetic:

```ant
func double[T: Number](x: T): T {
  x + x
}

assert(double(7) == 7 * 2)
assert(double(2.5) == 2.5 * 2.0)
assert(double(byte(100)) == byte(200))
```

`Math.abs` is bounded by `Number`, so it keeps the type it gets:

```ant
assert(Math.abs(int8(-5)) == int8(5))
assert(Math.abs(float32(-1.5)) == float32(1.5))
```

## Possible issues

An operator on two `Number` values is a compile error, because they may be different types. Narrow
them first:

```ant
let a: Number = 2
let b: Number = 3
let c = a + b
// error: + applies to a single numeric kind, string, or a named primitive type over one, not to Number; narrow first
```

`==` compiles on `Number` values, but numbers of different types are never equal:

```ant
let a: Number = 2
let b: Number = 2.0
let c: Number = 2
a == b // false
a == c // true
```

`is` cannot test `Number` or `Integer` themselves, since an alias has no run-time identity:

```ant
func kind(n: Number): string {
  if n is Integer { "integer" } else { "float" }
  // error: Integer is an alias and has no run-time identity; test its members instead
}
```

Test the members instead, as `kind` in Scope does.

A value typed `Number` cannot be passed to a function bounded by `Number`: `T` would be the whole
union, not one member:

```ant
func double[T: Number](x: T): T {
  x + x
}

let n: Number = 2
let d = double(n) // error: T would be Number, which is not a single member of its bound Number
```

All the arguments typed `T` share one type, so an `int` and a `float64` do not mix. Convert one of
them:

```ant
func add[T: Number](a: T, b: T): T {
  a + b
}

let s = add(1, 2.5) // error: cannot use float64 where int is expected
```

```ant
func add[T: Number](a: T, b: T): T {
  a + b
}

add(float64(1), 2.5) // 3.5
```

A literal inside the body is an `int` or a `float64`, not a `T`:

```ant
func half[T: Number](x: T): T {
  x / 2 // error: / needs both operands of one type, got T and int
}
```

Return a concrete type instead, and convert `x` to it:

```ant
func half[T: Number](x: T): float64 {
  float64(x) / 2.0
}

half(7)       // 3.5
half(byte(9)) // 4.5
```

A narrowed value is a plain member type, no longer a `T`, so it cannot be returned as a `T`:

```ant
func bump[T: Number](x: T): T {
  if x is int { x + 1 } else { x } // error: cannot use int where T is expected
}
```

A string, a `bool`, or a named primitive type over a number is not a member:

```ant
func double[T: Number](x: T): T {
  x + x
}

type Meters = float64
let e = double("ab") // error: T would be string, which is not a member of its bound Number
let g = double(Meters(1.0))
// error: T would be Meters, which is not a member of its bound Number; unwrap it with float64(...)
```

Unwrap a named primitive type, then tag the result again:

```ant
func double[T: Number](x: T): T {
  x + x
}

type Meters = float64
Meters(double(float64(Meters(1.5)))) // Meters(3.0)
```

Arithmetic on `T` panics on overflow, as it does on the member type:

```ant
func add[T: Number](a: T, b: T): T {
  a + b
}

add(byte(200), byte(55)) // 255
add(byte(200), byte(56)) // panic: integer overflow
```

## See also

- [Integer](Integer.md): the integer members only, the bound of `Bits`.
- [Ordered](Ordered.md): `Number` and `string`, the bound of `Math.min`, `Math.max`, and
  `Math.clamp`.
- [Math.abs](../abs.md), [Array.sum](../../Array/sum.md), [Atomic.add](../../Atomic/add.md)
- [int](../../prelude/int.md), [float64](../../prelude/float64.md): conversions between number
  types.
- [is](../../../language/operators/is.md)
