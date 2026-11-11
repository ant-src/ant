<!-- status: reviewed -->

# Assignment and compound assignment

`=` stores a new value in a place: a `var`, a field, an array element, or a map entry. The compound
operators `+=`, `-=`, `*=`, `/=`, and `%=` combine the place's value with the right side and store
the result.

## Details

- An assignment is an expression of type `void`. It cannot be chained or used as a condition.
- A target is a `var`, a field such as `p.x`, an element such as `xs[i]` or `m[k]`, or an array
  destructuring such as `[a, b]`. Assigning a `let` name, a parameter, or a name a `for` binds is a
  compile error.
- The value must fit the target's declared type. `x += y` follows the rules of `x + y`: both
  operands have one type, and integer overflow panics.
- The target's sub-expressions, such as the index in `xs[i] = v`, evaluate before the right side. An
  array destructuring assignment then stores left to right.
- A destructuring assignment allows only `=`. There is no record destructuring assignment, and no
  `++` or `--`.

## Basic examples

```ant
var count = 0
count = 5
count += 2
count // 7
```

Each compound operator applies its arithmetic operator:

```ant
var n = 20
n -= 2
n *= 3
n /= 4
n %= 5
n // 3
```

`+=` on a string appends:

```ant
var greeting = "Hello"
greeting += ", Ant"
greeting // "Hello, Ant"
```

## Scope

### Fields and elements

A field or an element is a place too. A `let` fixes only the name, so the value's fields and
elements can still be assigned:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
p.x += 10
p.y = 0
p // Point{ x: 11, y: 0 }
let xs = [1, 2, 3]
xs[0] = 9
xs[2] *= 10
xs // [9, 2, 30]
```

A path of fields and indexes ends in one place:

```ant
type Point = { x: int, y: int }
let path = [Point{ x: 0, y: 0 }, Point{ x: 3, y: 4 }]
path[1].x *= 2
path // [Point{ x: 0, y: 0 }, Point{ x: 6, y: 4 }]
```

### Map entries

`m[k] = v` inserts a new entry or replaces the value of an existing one:

```ant
let ages = Map{ "Ann": 31 }
ages["Ann"] = 32
ages["Bob"] = 27
ages // Map{ "Ann": 32, "Bob": 27 }
```

### Destructuring

An array destructuring assigns several places at once. The right side is built before any position
is stored, so `[a, b] = [b, a]` swaps:

```ant
var a = 1
var b = 2
[a, b] = [b, a]
a // 2
b // 1
```

The positions can be any places, `_` to skip a value, or `...` for the rest:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 0, y: 0 }
let xs = [0, 0]
var rest: Array[int] = []
[p.x, _, xs[1], ...rest] = [1, 2, 3, 4, 5]
p    // Point{ x: 1, y: 0 }
xs   // [0, 3]
rest // [4, 5]
```

A destructuring nests when the elements are arrays:

```ant
var top = 0
var low = 0
[[top], [_, low]] = [[1, 2], [3, 4]]
top // 1
low // 4
```

### Discarding a value

`_ = expr` evaluates `expr` and discards its value:

```ant
func save(): int {
  println("saved")
  1
}
_ = save()
```

```text
saved
```

### Closures

A closure can assign a `var` it captures, and the change is seen outside:

```ant
var clicks = 0
let click = func() { clicks += 1 }
click()
click()
clicks // 2
```

### Evaluation order

The target's index runs before the right side:

```ant
func slot(): int {
  println("index")
  0
}
func value(): int {
  println("value")
  9
}
let xs = [1]
xs[slot()] = value()
println(xs)
```

```text
index
value
[9]
```

## Applications

Advance a pair in one step, as in a Fibonacci loop:

```ant
var a = 0
var b = 1
for _ in Iter.range(0, 10) {
  [a, b] = [b, a + b]
}
println(a)
```

```text
55
```

Count words in a map. Reading `counts[word]` gives `null` for a missing key, so add a default with
`??`:

```ant
let counts: Map[string, int] = Map{}
for word in "the ant and the bee".split(" ") {
  counts[word] = (counts[word] ?? 0) + 1
}
println(counts)
```

```text
Map{ "the": 2, "ant": 1, "and": 1, "bee": 1 }
```

Total a cart:

```ant
type Item = { price: int, qty: int }
let cart = [Item{ price: 3, qty: 2 }, Item{ price: 5, qty: 1 }]
var total = 0
for item in cart {
  total += item.price * item.qty
}
println(total)
```

```text
11
```

## Properties and relations

A compound assignment is the same as assigning the result of its operator:

```ant
var x = 7
var y = 7
x *= 6
y = y * 6
assert(x == y)
```

A destructuring assignment is the same as building the right side, then assigning each position in
turn:

```ant
var a = 1
var b = 2
[a, b] = [b, a]
var c = 1
var d = 2
let tmp = [d, c]
c = tmp[0]
d = tmp[1]
assert([a, b] == [c, d])
```

## Possible issues

### Assigning a `let`

A `let` name cannot be assigned, with `=` or a compound operator. Declare it with `var`:

```ant
let total = 0
total += 5 // error: cannot assign to total, a let; declare it with var to reassign it
```

Parameters are `let` bindings too. Copy one into a `var` to change it:

```ant
func countdown(n: int): int {
  n -= 1 // error: cannot assign to n, a parameter; declare a var to hold a value that changes
  n
}
```

```ant
func countdown(n: int): int {
  var left = n
  left -= 1
  left
}
countdown(3) // 2
```

### Assignment is not a value

An assignment has type `void`, so `if` cannot test it. Compare with `==`:

```ant
var count = 0
if count = 5 { println("five") } // error: cannot use void where bool is expected
```

Assignments cannot be chained. Write one per target:

```ant
var a = 0
var b = 0
a = b = 3 // error: assignment cannot be chained
```

### Assigning through a union

A field cannot be assigned through a union type, such as a nullable record. Narrow first:

```ant
type User = { name: string }
func rename(u: User | null) {
  u.name = "Ann" // error: cannot assign a field through the union User | null; narrow first
}
```

```ant
type User = { name: string }
func rename(u: User | null) {
  if u == null {
    return
  }
  u.name = "Ann"
}
let u = User{ name: "Bob" }
rename(u)
u.name // "Ann"
```

### No `++` or `--`

Write `+= 1` and `-= 1` instead:

```ant
var i = 0
i++ // error: unexpected '+'
```

### Mixed types

Both sides of a compound operator have one type. An `int` does not mix with a `float64`:

```ant
var total = 0.0
total += 1 // error: + needs both operands of one type, got float64 and int
```

```ant
var total = 0.0
total += 1.0
total // 1.0
```

The value of `=` must fit the target's declared type. Declare a wider type up front:

```ant
var status = "idle"
status = null // error: cannot use null where string is expected
```

### Compound destructuring

A destructuring assignment takes only `=`:

```ant
var a = 1
var b = 2
[a, b] += [10, 20] // error: a destructuring assignment allows only '='
```

### Out of range

Assigning past the end of an array panics; it does not grow the array. Use `push` to append:

```ant
let xs = [1, 2, 3]
xs[3] = 4 // panic: index 3 out of range for an array of length 3
```

A destructuring with more positions than the array has elements panics:

```ant
var a = 0
var b = 0
[a, b] = [1] // panic: a destructuring needs 2 elements, got 1
```

### Overflow

A compound operator on integers panics on overflow, as the operator does:

```ant
var n = 9223372036854775807
n += 1 // panic: integer overflow
```

## See also

[var](../keywords/var.md) · [let](../keywords/let.md) · [Arithmetic operators](arithmetic.md) ·
[Indexing](indexing.md) · [?? operator](null-coalescing.md) · [void](../types/void.md) ·
[Variables tour](../../tour/02-variables.md)
