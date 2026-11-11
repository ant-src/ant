<!-- status: reviewed -->

# The pipe |

`a | f(b, c)` calls `f(a, b, c)`: the value on the left becomes the first argument of the call on
the right. A chain of pipes reads top to bottom, one transformation per stage, and each stage may
call a function of any namespace.

## Details

- The right side must be a call, written with its parentheses.
- The left side evaluates first, then the stage's own arguments, left to right.
- `|` binds looser than every other operator but assignment. A pipeline used as an operand of `==`,
  `+`, `??`, or any other operator needs parentheses.
- `try` binds tighter than `|`: `try f(x) | g()` unwraps `f(x)` and pipes the value on. To propagate
  the error of a whole pipeline, write `try (… | …)`.
- A line that starts with `|` continues the expression of the line before.

## Basic examples

```ant
"  Ant  " | String.trim()           // "Ant"
[3, 1, 2] | Array.sort()            // [1, 2, 3]
"ant" | String.repeat(3)            // "antantant"
["a", "b", "c"] | String.join(", ") // "a, b, c"
```

A longer pipeline puts each stage on its own line:

```ant
let words = "the quick brown ant"
  | String.split(" ")
  | Array.filter({ x.length() > 3 })
  | Array.map({ x.toUpper() })
println(words)
```

```text
["QUICK", "BROWN"]
```

## Scope

### Your own functions

A stage calls any function, not only those of the prelude. A function value bound to a name works
too:

```ant
func double(n: int): int { n * 2 }
let add = func(a: int, b: int): int { a + b }
5 | double()          // 10
5 | double() | add(3) // 13
```

### Crossing namespaces

A call through `.` reaches only the functions of the value's own namespace. A pipe calls the
function of another namespace, such as `String.join` on an array:

```ant
let line = "b,a,c".split(",")
  | Array.sort()
  | String.join(" ")
println(line)
```

```text
a b c
```

### Named arguments

The piped value fills the first parameter, and named arguments follow as in any call:

```ant
"ff" | String.toInt(base: 16)      // 255
"7" | String.padStart(3, pad: "0") // "007"
```

### Constructors

A named primitive type is called like a function, so a pipe can wrap a value in it:

```ant
type Port = int
80 | Port() // Port(80)
```

### Results

A pipe adds context to a `Result` in one step:

```ant
func parsePort(text: string): Result[int, Error] {
  let n = text.trim().toInt()
  if n == null || n < 1 || n > 65535 {
    return Result.fail("want a port from 1 to 65535, got ${text}")
  }
  Result.ok(n)
}
println(parsePort("http") | Result.wrapErr("config line 3"))
```

```text
Err{ error: Error{ message: "config line 3: want a port from 1 to 65535, got http" } }
```

### With `try`

`try` takes the first stage only. Parenthesize the pipeline to propagate the error of its last
stage:

<!-- test: before
func parsePort(text: string): Result[int, Error] {
  let n = text.trim().toInt()
  if n == null || n < 1 || n > 65535 {
    return Result.fail("want a port from 1 to 65535, got ${text}")
  }
  Result.ok(n)
}
-->

```ant
func countPorts(texts: Array[string]): Result[int, Error] {
  let all = try (texts | Array.map(parsePort) | Result.all())
  Result.ok(all.length())
}
println(countPorts(["80", "443"]))
println(countPorts(["80", "http"]))
```

```text
Ok{ value: 2 }
Err{ error: Error{ message: "want a port from 1 to 65535, got http" } }
```

### Ending with output

A pipeline may end in a function that prints its result:

```ant
[3, 1, 2]
  | Array.sort()
  | println()
```

```text
[1, 2, 3]
```

### Evaluation order

The left side runs before the arguments of the stage:

```ant
func word(): string {
  println("left")
  "ab"
}
func count(): int {
  println("argument")
  2
}
println(word() | String.repeat(count()))
```

```text
left
argument
abab
```

## Applications

Total the paid orders. `Array.sum` gives `null` for an empty array, so the pipeline is parenthesized
before `??`:

```ant
type Order = { amount: int, paid: bool }
let orders = [
  Order{ amount: 30, paid: true },
  Order{ amount: 12, paid: false },
  Order{ amount: 8, paid: true },
]
let total = (
  orders
    | Array.filter({ x.paid })
    | Array.map({ x.amount })
    | Array.sum()
) ?? 0
println(total)
```

```text
38
```

Turn a title into a URL slug:

```ant
let slug = "  Pipes in Ant  "
  | String.trim()
  | String.toLower()
  | String.split(" ")
  | Array.filter({ x != "" })
  | String.join("-")
println(slug)
```

```text
pipes-in-ant
```

Count the distinct words of a text:

```ant
let distinct = "the ant and the bee and the ant"
  | String.split(" ")
  | Set.of()
println(distinct.size())
```

```text
4
```

## Properties and relations

A pipe is the call with the left value as its first argument:

```ant
let xs = [3, 1, 2]
assert((xs | Array.map({ x * 2 })) == Array.map(xs, { x * 2 }))
assert(("ant" | String.repeat(2)) == String.repeat("ant", 2))
```

Pipes associate to the left, so each stage takes the result of the one before:

```ant
let s = "  b a  "
assert((s | String.trim() | String.split(" ")) == String.split(String.trim(s), " "))
```

For a function of the value's own namespace, a pipe and a call through `.` are the same call:

```ant
let xs = [3, 1, 2]
assert((xs | Array.sort()) == xs.sort())
```

## Possible issues

### The right side is not a call

A stage needs its parentheses, even when it takes no other argument:

```ant
let name = "  Ant  " | String.trim // error: the right side of a pipe must be a call
```

A lambda is not a call either. Call a named function, or use `.` inside a lambda passed to a stage:

```ant
let n = "Ant" | { x.length() } // error: the right side of a pipe must be a call
```

### An operator after a stage

`|` binds looser than `==`, `+`, and `??`, so an operator after a stage is a compile error.
Parenthesize the pipeline:

```ant
let input = "  "
let blank = input | String.trim() == ""
// error: '|' binds looser than '==', so parenthesize the pipe, as in (a | f()) == b
```

```ant
let input = "  "
(input | String.trim()) == "" // true
```

### Passing the value twice

The pipe passes the left value already. Writing it again in the call adds one argument too many:

```ant
let xs = [1, 2]
let doubled = xs | Array.map(xs, { x * 2 }) // error: Array.map takes 2 arguments, got 3
```

### `try` takes only the first stage

Without parentheses, `try` applies to the value before the first `|`:

<!-- test: before
func parsePort(text: string): Result[int, Error] {
  Result.ok(80)
}
-->

```ant
func ports(texts: Array[string]): Result[Array[int], Error] {
  let all = try texts | Array.map(parsePort) | Result.all()
  // error: try needs a Result, got Array[string]
  Result.ok(all) // error: cannot use Result[Array[int], Error] where Array[int] is expected
}
```

### A dot call as a stage

When a stage is a call through `.`, its receiver stays the first argument and the piped value comes
second:

```ant
[1, 2] | [3].concat() // [3, 1, 2]
```

### A nullable value

A pipe does not skip `null`. A nullable value must be narrowed first, or use `?.` for a function of
its own namespace:

```ant
let name: string | null = null
let n = name | String.length() // error: cannot use string | null where string is expected
```

```ant
let name: string | null = null
name?.length() // null
```

## See also

[Calls through .](dot-calls.md) · [?. operator](optional-chaining.md) ·
[?? operator](null-coalescing.md) · [try](../keywords/try.md) ·
[Result.wrapErr](../../lib/Result/wrapErr.md) · [String.join](../../lib/String/join.md) ·
[Idiomatic Ant](../../guide/idiomatic.md) · [Calls tour](../../tour/07-calls.md)
