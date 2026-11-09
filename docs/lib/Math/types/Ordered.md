<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Ordered

```ant signature
type Ordered = Number | string
```

Ordered is the numbers and string, the built-in types that \< and \> compare.

<!-- end generated -->

## Details

- A call sets `T` to one member, never to a union or to `Ordered` itself; narrow a union value with
  `is` first.
- A named primitive type such as `type Meters = float64` is not a member, though `<` works on two
  values of it.

## Basic examples

Bound a type parameter by `Ordered` to compare its values:

```ant
func bigger[T: Ordered](a: T, b: T): T { if a > b { a } else { b } }
bigger(3, 8)          // 8
bigger("pear", "fig") // "pear"
bigger(2.5, -1.0)     // 2.5
```

The prelude functions that compare take the same bound:

```ant
Math.max(4, 9)          // 9
[3, 1, 2].sort()        // [1, 2, 3]
["kiwi", "apple"].min() // "apple"
```

## Scope

Every sized integer and float type is a member, with both arguments of the one type:

```ant
func bigger[T: Ordered](a: T, b: T): T { if a > b { a } else { b } }
bigger(byte(3), byte(200))          // 200
bigger(int32(-7), int32(-2))        // -2
bigger(float32(1.5), float32(0.25)) // 1.5
```

Inside the body, `T` also takes any other operator that every member takes, such as `+`, which adds
numbers and concatenates strings:

```ant
func twice[T: Ordered](x: T): T { x + x }
twice(21)   // 42
twice("ab") // "abab"
```

Test a `T` against a member with `is` to treat strings apart. The narrowed value is a plain
`string`, no longer a `T`, so keep a copy to return:

```ant
func orDefault[T: Ordered](x: T, fallback: T): T {
  let original = x
  if x is string {
    if x == "" { fallback } else { original }
  } else {
    x
  }
}
orDefault("", "none")   // "none"
orDefault("ok", "none") // "ok"
orDefault(0, 5)         // 0
```

`Ordered` is also a type a value can have, a union like any other. Narrow it before you compare:

```ant
func describe(v: Ordered): string {
  if v is string { "text ${v}" } else if v is int { "int ${v}" } else { "another number" }
}
describe("hi") // "text hi"
describe(7)    // "int 7"
describe(0.5)  // "another number"
```

## Applications

Find the largest element of any array whose elements compare, `null` when it is empty:

```ant
func largest[T: Ordered](xs: Array[T]): T | null {
  var best: T | null = null
  for x in xs {
    if best == null || x > best {
      best = x
    }
  }
  best
}
largest([3, 9, 4])         // 9
largest(["pear", "apple"]) // "pear"
let none: Array[int] = []
largest(none) // null
```

`Array.max` does the same for an array.

Check that an array is sorted, for numbers and strings alike:

```ant
func isSorted[T: Ordered](xs: Array[T]): bool {
  for i in Iter.range(1, len(xs)) {
    if xs[i] < xs[i - 1] {
      return false
    }
  }
  true
}
isSorted([1, 2, 2, 5])    // true
isSorted(["b", "a"])      // false
isSorted([0.5, 1.0, 1.5]) // true
```

Sort records by an `Ordered` key with `sortBy`, since a record itself is not ordered:

```ant
type Person = { name: string, age: int }
let people = [Person{ name: "Ann", age: 31 }, Person{ name: "Bob", age: 27 }]
people.sortBy({ x.age }).map({ x.name })  // ["Bob", "Ann"]
people.sortBy({ x.name }).map({ x.name }) // ["Ann", "Bob"]
```

## Properties and relations

A bounded function returns the caller's type, so the result needs no narrowing:

```ant
func bigger[T: Ordered](a: T, b: T): T { if a > b { a } else { b } }
let n: int = bigger(1, 2)
let s: string = bigger("a", "b")
assert(n == 2 && s == "b")
```

## Possible issues

A value of type `Ordered` is a union, so it cannot set `T`. Narrow it first:

```ant
func bigger[T: Ordered](a: T, b: T): T { if a > b { a } else { b } }
let a: Ordered = 1
let b: Ordered = 2
bigger(a, b) // error: T would be Ordered, which is not a single member of its bound Ordered; narrow first
```

Comparing two `Ordered` values directly is a compile error too, since they may hold different kinds:

```ant
func less(a: Ordered, b: Ordered): bool {
  a < b // error: < applies to a single numeric kind, string, or a named primitive type over one, not to Ordered; narrow first
}
```

An array of mixed numbers and strings does not sort:

```ant
let xs: Array[int | string] = [1, "a"]
let sorted = xs.sort() // error: T would be int | string, which is not a single member of its bound Ordered; narrow first
```

Both arguments need one type. An `int` and a `float64` do not mix:

```ant
let m = Math.max(1, 2.0) // error: cannot use float64 where int is expected
```

`bool` is not ordered:

```ant
let m = Math.max(true, false) // error: T would be bool, which is not a single member of its bound Ordered; narrow first
```

A named primitive type is not a member, though `<` works on two values of it. Compare the payloads:

```ant
type Meters = float64
let a = Meters(3.0)
let b = Meters(5.0)
let closer = Math.min(a, b) // error: T would be Meters, which is not a single member of its bound Ordered; narrow first
```

Unwrap them with `float64` to call `Math.min`, or compare them with `<` to keep the `Meters` type:

```ant
type Meters = float64
let a = Meters(3.0)
let b = Meters(5.0)
Math.min(float64(a), float64(b)) // 3.0
if a < b { a } else { b }        // Meters(3.0)
```

A `NaN` wins `Math.min` and `Math.max`, and `sort` puts it after every number:

```ant
Math.max(1.0, Math.NaN)     // NaN
[2.0, Math.NaN, 1.0].sort() // [1.0, 2.0, NaN]
```

## See also

- [Number](Number.md) and [Integer](Integer.md): the numeric bounds.
- [Math.min](../min.md), [Math.max](../max.md), and [Math.clamp](../clamp.md)
- [Array.sort](../../Array/sort.md) and [Array.sortBy](../../Array/sortBy.md)
- [Array.min](../../Array/min.md) and [Array.max](../../Array/max.md)
- [Math.isNaN](../isNaN.md)
