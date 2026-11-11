<!-- status: reviewed -->

# &&, ||, and !

`a && b` is `true` when both operands are `true`. `a || b` is `true` when either operand is `true`.
`!a` is the negation of `a`. All three take and give `bool`.

## Details

- `&&` and `||` short-circuit: `a && b` evaluates `b` only when `a` is `true`, and `a || b` only
  when `a` is `false`.
- The operands must be `bool`. There is no truthiness: an `int`, a `string`, or a nullable value is
  a compile error, not a test.
- `&&` binds tighter than `||`, and both bind looser than comparison and equality, so
  `a < b && c == d` needs no parentheses. `!` is a prefix operator and binds tighter than all of
  them.
- `??` does not mix with `&&` or `||` without parentheses.
- A null test or an `is` test narrows the variable through `&&`, `||`, and `!`, so the right operand
  of `x != null && …` sees `x` as non-null.

## Basic examples

```ant
true && false // false
true || false // true
!true         // false
```

Comparisons combine without parentheses:

```ant
let age = 30
age >= 18 && age < 65 // true
age < 18 || age >= 65 // false
```

## Scope

### Short-circuit evaluation

The right operand runs only when the left one does not decide the result:

```ant
func loud(name: string, value: bool): bool {
  println("evaluated ${name}")
  value
}

println(false && loud("right", true))
println(true || loud("right", true))
println(true && loud("right", false))
```

```text
false
true
evaluated right
false
```

### Guarding the right operand

Short-circuiting lets the left operand protect the right one. Here `xs[0]` runs only when the array
is not empty:

```ant
let xs: Array[int] = []
!xs.isEmpty() && xs[0] > 1 // false
```

### Narrowing through `&&` and `||`

After `name != null &&`, the right operand sees `name` as a `string`. After `name == null ||`, it
does too, since the right side runs only when the test failed:

```ant
let name: string | null = "Ann"
name != null && name.length() > 2 // true
name == null || name.isEmpty()    // false
```

The same holds for `is` on a union:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square

func isLarge(s: Shape): bool {
  s is Circle && s.r > 10.0 || s is Square && s.side > 10.0
}

println(isLarge(Circle{ r: 11.0 }))
println(isLarge(Square{ side: 4.0 }))
```

```text
true
false
```

### Guards with `||`

An `if` with `||` and an early `return` leaves the variable narrowed below it:

```ant
func describe(text: string | null): string {
  if text == null || text.isEmpty() { return "empty" }
  "${text.length()} characters"
}

println(describe(null))
println(describe(""))
println(describe("ant"))
```

```text
empty
empty
3 characters
```

### Negating a test with `!`

`!` binds tighter than `is`, so a negated `is` test needs parentheses. It narrows like the test it
negates:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }

func area(s: Circle | Square): float64 {
  if !(s is Circle) { return s.side * s.side }
  3.0 * s.r * s.r
}

println(area(Square{ side: 2.0 }))
println(area(Circle{ r: 1.0 }))
```

```text
4.0
3.0
```

`!` applies after a dot call, since postfix binds tighter than prefix:

```ant
let xs = [1, 2]
!xs.isEmpty() // true
```

### Nullable booleans

A `bool | null` is not a `bool`. Give it a default with `??` first, in parentheses when it meets
`&&` or `||`:

```ant
let settings = Map{ "verbose": true }
let quiet = false
!(settings["verbose"] ?? false)       // false
(settings["debug"] ?? false) || quiet // false
```

## Applications

### Leap years

`&&` binds tighter than `||`, so the rule reads as written:

```ant
func isLeap(year: int): bool {
  year % 4 == 0 && year % 100 != 0 || year % 400 == 0
}

[1900, 2000, 2024, 2025].filter(isLeap) // [2000, 2024]
```

### Filtering on several conditions

```ant
type Order = { customer: string, paid: bool, total: int }

let orders = [
  Order{ customer: "Ann", paid: true, total: 120 },
  Order{ customer: "Bob", paid: false, total: 300 },
  Order{ customer: "Eve", paid: true, total: 40 },
]

orders
  | Array.filter({ x.paid && x.total >= 100 })
  | Array.map({ x.customer })
// ["Ann"]
```

### Validating input

```ant
func validUsername(name: string): bool {
  let size = name.length()
  size >= 3 && size <= 16 && !name.contains(" ")
}

println(validUsername("ant_dev"))
println(validUsername("al"))
println(validUsername("two words"))
```

```text
true
false
false
```

## Properties and relations

De Morgan's laws hold for every pair of booleans, and each operator equals an `if` expression:

```ant
for a in [true, false] {
  for b in [true, false] {
    assert(!(a && b) == (!a || !b))
    assert(!(a || b) == (!a && !b))
    assert((a && b) == (if a { b } else { false }))
    assert((a || b) == (if a { true } else { b }))
  }
}
println("ok")
```

```text
ok
```

## Possible issues

### No truthiness

A number, a string, or a nullable value is not a condition. Compare it explicitly:

```ant
let count = 3
let ready = count && true // error: cannot use int where bool is expected
```

```ant
let name: string | null = null
let shown = !name // error: cannot use string | null where bool is expected
```

```ant
let count = 3
let name: string | null = null
count > 0 && name == null // true
```

### Mixing `??` with `&&` or `||`

```ant
let flag: bool | null = null
let on = flag ?? false || true // error: '??' cannot mix with '&&' or '||' without parentheses
```

Add parentheses to say which you mean:

```ant
let flag: bool | null = null
(flag ?? false) || true // true
flag ?? (false || true) // true
```

### `!` binds tighter than `is` and `==`

`!s is Circle` negates `s`, not the test:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
let s: Circle | Square = Circle{ r: 1.0 }
let notCircle = !s is Circle
// error: cannot use Circle | Square where bool is expected
// error: bool is never a Circle, so the test is always false
```

Write `!(s is Circle)`, and `x != null` instead of `!x == null`.

### `&&` is not a statement shortcut

Both operands must be `bool`, so `&&` cannot run a call for its effect. Use `if`:

```ant
let ok = true
ok && println("yes") // error: cannot use void where bool is expected
```

```ant
let ok = true
if ok { println("yes") }
```

```text
yes
```

### No `^`, `&&=`, or `||=`

Ant has no exclusive-or operator and no `&&=` or `||=`. Two `bool`s differ exactly when `a != b`, so
`!=` is exclusive or:

```ant
let a = true
let b = false
a != b // true
```

In place of `&&=` and `||=`, write the assignment out:

```ant
var ok = true
ok = ok && 3 > 5
ok = ok || 1 < 2
println(ok)
```

```text
true
```

### A field does not narrow

Narrowing applies to locals, parameters, and top-level `let`s. A test on a field path does not
narrow the right operand. Bind the field to a local first:

```ant
type User = { nickname: string | null }
let user = User{ nickname: "ant" }
let long = user.nickname != null && user.nickname.length() > 2
// error: string | null has no field or function length on every member: narrow the value first
```

```ant
type User = { nickname: string | null }
let user = User{ nickname: "ant" }
let nickname = user.nickname
nickname != null && nickname.length() > 2 // true
```

## See also

- [bool](../types/bool.md)
- [true and false](../literals/booleans.md)
- [Comparison operators](comparison.md)
- [Equality operators](equality.md)
- [is](is.md)
- [?? (null coalescing)](null-coalescing.md)
- [if](../keywords/if.md)
- [Array.every](../../lib/Array/every.md)
- [Array.some](../../lib/Array/some.md)
- [Control flow](../../guide/control-flow.md)
- [Null](../../guide/null.md)
