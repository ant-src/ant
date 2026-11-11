<!-- status: reviewed -->

# Operator precedence

Precedence decides which operator takes an operand first when an expression has several. From
tightest to loosest binding:

| Level          | Operators                    |
| -------------- | ---------------------------- |
| postfix        | `a.b` `a?.b` `a[i]` `f(x)`   |
| prefix         | `-a` `!a` `try a`            |
| multiplicative | `*` `/` `%`                  |
| additive       | `+` `-`                      |
| comparison     | `<` `<=` `>` `>=` `is`       |
| equality       | `==` `!=`                    |
| logical        | `&&`, then `\|\|`            |
| null           | `??`                         |
| pipe           | `\|`                         |
| assignment     | `=` `+=` `-=` `*=` `/=` `%=` |

Parentheses override the table: `(a + b) * c` adds first.

## Details

- `&&` binds tighter than `||`, so `a || b && c` is `a || (b && c)`.
- `??` cannot mix with `&&`, `||`, a comparison, or an equality test without parentheses. It is a
  compile error.
- A binary operator after a pipe stage, as in `s | String.trim() == ""`, is a compile error.
  Parenthesize the pipeline: `(s | String.trim()) == ""`.
- `|` is left associative: `a | f() | g()` is `g(f(a))`.
- Operands evaluate left to right whatever the grouping. `&&`, `||`, and `??` skip their right
  operand when the left one decides the result.

## Basic examples

Multiplication binds tighter than addition. Parentheses change the order:

```ant
2 + 3 * 4   // 14
(2 + 3) * 4 // 20
```

A comparison binds tighter than `&&`, so a condition needs no parentheses:

```ant
let age = 30
age >= 18 && age < 65 // true
```

A pipe binds looser than the arithmetic in its left side:

```ant
let prices = [3, 1, 2]
prices.length() * 10 | Math.max(25) // 30
```

## Scope

### Postfix and prefix

Postfix operators bind tightest, so a prefix operator applies to the whole chain after it:

```ant
let name = "Ant"
let xs = [4, 5, 6]
-name.length()  // -3
!name.isEmpty() // true
-xs[0] * 2      // -8
```

`try` is a prefix operator too. In `try parse(s).map(f)`, it unwraps the result of the whole chain,
and `try a + 1` adds 1 to the unwrapped value:

```ant
func parse(s: string): Result[int, Error] {
  let n = s.toInt()
  if n == null {
    return Result.fail("not a number: ${s}")
  }
  Result.ok(n)
}

func next(s: string): Result[int, Error] {
  let doubled = try parse(s).map({ x * 2 })
  Result.ok(try parse(s) + doubled)
}

println(next("7"))
```

```text
Ok{ value: 21 }
```

### Multiplicative and additive

`*`, `/`, and `%` bind tighter than `+` and `-`:

```ant
1 + 10 % 4 // 3
-2 * 3 + 4 // -2
```

`+` on strings sits at the same level as on numbers:

```ant
let first = "Ada"
"Hello, " + first + "!" // "Hello, Ada!"
```

### Comparison and equality

Arithmetic binds tighter than comparison, and comparison binds tighter than equality. So `==` can
compare the results of two comparisons:

```ant
let a = [1, 2]
let b: Array[int] = []
a.length() + 1 > 2               // true
a.length() > 0 == b.length() > 0 // false
```

`is` sits at the comparison level, so it combines with `&&` without parentheses, and the right side
of `&&` sees the narrowed value:

```ant
let v: int | string = 3
v is int && v > 2 // true
```

### Logical

`&&` binds tighter than `||`:

```ant
let a = true
let b = false
let c = false
a || b && c   // true
(a || b) && c // false
```

### Null

`??` binds looser than comparison and arithmetic, so its right side takes the whole operand after
it:

```ant
let stock: int | null = 5
stock ?? 1 + 2   // 5
(stock ?? 1) + 2 // 7
```

A chain of `??` takes the first value that is not `null`:

```ant
let saved: string | null = null
let typed: string | null = "dark"
saved ?? typed ?? "light" // "dark"
```

`return` ends a guard after `??`. It takes the whole expression to its right, so `x ?? return y + 1`
returns `y + 1`. Parenthesize the `return` to end it earlier:

```ant
func count(input: string): int {
  let n = input.toInt() ?? return -1
  n * 2
}

println(count("21"))
println(count("many"))
```

```text
42
-1
```

### Pipe

The pipe binds looser than every other operator but assignment. Its left side can be any expression,
including `??`:

```ant
let limit: int | null = null
limit ?? -5 | Math.abs() // 5
```

Stages read left to right, each passing its value on as the first argument:

```ant
[1, 2, 3] | Array.map({ x * 2 }) | Array.filter({ x > 2 }) // [4, 6]
```

### Assignment

Assignment binds loosest, so the whole right side, a pipeline included, evaluates before the store:

```ant
var total = 10
total += [3, 1, 2] | Array.filter({ x > 1 }) | Array.length()
total // 12
var label = ""
label = ["a", "b"] | String.join(", ")
label // "a, b"
```

## Applications

A price rule mixes every arithmetic level and a comparison, with no parentheses needed:

```ant
type Item = { price: int, qty: int }

let cart = [Item{ price: 250, qty: 2 }, Item{ price: 1200, qty: 1 }]
let subtotal = cart.map({ x.price * x.qty }).sum() ?? 0
let shipping = if subtotal >= 1500 { 0 } else { 300 }
println(subtotal + shipping)
```

```text
1700
```

A validation reads as one condition, since `is` and comparison bind tighter than `&&` and `||`:

```ant
type Field = { name: string, value: int | string }

func valid(f: Field): bool {
  let v = f.value
  v is int && v >= 0 || v is string && !v.isEmpty()
}

println(valid(Field{ name: "age", value: 31 }))
println(valid(Field{ name: "nick", value: "" }))
```

```text
true
false
```

## Properties and relations

Each level groups as the parenthesized form shows:

```ant
let a = 7
let b = 3
let c = 2
assert(a + b * c == a + (b * c))
assert(-a * b == (-a) * b)
assert((a < b || b < c && c < a) == (a < b || (b < c && c < a)))
```

A pipe is the call it stands for, so it can replace a nested call:

```ant
let words = ["b", "a"]
assert((words | Array.sort() | String.join("")) == String.join(Array.sort(words), ""))
```

## Possible issues

### A binary operator after a pipe

The pipe binds looser than `==`, so the comparison would apply to the stage, not the pipeline. The
checker rejects it:

```ant
let input = "  "
let blank = input | String.trim() == ""
// error: '|' binds looser than '==', so parenthesize the pipe, as in (a | f()) == b
```

```ant
let input = "  "
(input | String.trim()) == "" // true
```

The same holds for `??` after a pipe stage:

```ant
let scores = [3, 1, 2]
let total = scores | Array.sum() ?? 0
// error: '|' binds looser than '??', so parenthesize the pipe, as in (a | f()) ?? b
```

```ant
let scores = [3, 1, 2]
(scores | Array.sum()) ?? 0 // 6
```

### `try` before a pipeline

`try` is a prefix operator, so `try f(x) | g()` is `(try f(x)) | g()`. It unwraps the first stage
and pipes the plain value on, so a `Result` stage after it does not check:

```ant
func f(s: string): Result[int, Error] {
  Result.ok(s.length())
}

func g(s: string): Result[int, Error] {
  let n = try f(s) | Result.map({ x + 1 }) // error: cannot use int where Result[T, E] is expected
  Result.ok(n)
}
```

To propagate the error of the whole pipeline, parenthesize it:

```ant
func f(s: string): Result[int, Error] {
  Result.ok(s.length())
}

func g(s: string): Result[int, Error] {
  let n = try (f(s) | Result.map({ x + 1 }))
  Result.ok(n)
}

println(g("abc"))
```

```text
Ok{ value: 4 }
```

### `!` before `is`

Prefix `!` binds tighter than `is`, so `!v is int` is `(!v) is int`, not a test that `v` is not an
`int`. Parenthesize the test:

```ant
let v: int | string = 3
!v is int // error: cannot use int | string where bool is expected
```

```ant
let v: int | string = 3
!(v is int) // false
```

### `??` with `&&` or `||`

`??` cannot mix with `&&` or `||` without parentheses:

```ant
let ready: bool | null = null
let proceed = ready ?? false && true
// error: '??' cannot mix with '&&' or '||' without parentheses
```

```ant
let ready: bool | null = null
(ready ?? false) && true // false
```

### `??` before a comparison

`??` binds looser than a comparison, so `count ?? 0 > 0` would read as `count ?? (0 > 0)`, not a
test of the count. Mixing `??` with a comparison or an equality test without parentheses is a
compile error:

```ant
let count: int | null = 4
if count ?? 0 > 0 { println("some") } // error: '??' cannot mix with '>' without parentheses
```

```ant
let count: int | null = 4
if (count ?? 0) > 0 { println("some") }
```

```text
some
```

### Chained comparisons

`a < b < c` compares the `bool` of `a < b` with `c`. Write two comparisons joined by `&&`:

```ant
let low = 1
let n = 2
let high = 3
let inside = low < n < high // error: < needs both operands of one type, got bool and int
```

```ant
let low = 1
let n = 2
let high = 3
low < n && n < high // true
```

## See also

[Arithmetic operators](operators/arithmetic.md) · [Comparison operators](operators/comparison.md) ·
[Equality](operators/equality.md) · [Logical operators](operators/logical.md) ·
[?? operator](operators/null-coalescing.md) · [The pipe](operators/pipe.md) ·
[Assignment](operators/assignment.md) · [is](operators/is.md) · [try](keywords/try.md) ·
[Calls through .](operators/dot-calls.md)
