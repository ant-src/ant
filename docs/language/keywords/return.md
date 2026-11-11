<!-- status: reviewed -->

# `return`

`return` leaves a function early with a value. The last expression of a function body is already its
result, so `return` is for the paths that end before it.

## Details

- `return` leaves the innermost function body, script body, lambda, `go` block, or `defer` block. In
  a lambda it returns from the lambda only, not from the function around it. In a `test` body,
  `return` ends the test; like a script, a test returns only `void` or an `Err`.
- `return` takes the whole expression to its right. A `return` at the end of a line returns `void`.
- `return` has type `never`, so it can stand where a value is expected, as in `x ?? return y` or a
  branch of an `if` expression.
- After `if x == null { return }`, the checker treats `x` as not `null` for the rest of the block.
- Deferred blocks run when `return` leaves the function.

## Basic examples

Return early from a guard, and let the last expression give the ordinary result:

```ant
func clamp(value: int, low: int, high: int): int {
  if value < low { return low }
  if value > high { return high }
  value
}

println(clamp(-5, 0, 10))
println(clamp(7, 0, 10))
println(clamp(42, 0, 10))
```

```text
0
7
10
```

A function that returns `void` uses `return` with no value:

```ant
func greet(name: string) {
  if name.isEmpty() { return }
  println("Hello, ${name}")
}

greet("")
greet("Ann")
```

```text
Hello, Ann
```

## Scope

### Leaving a loop

`return` inside a `for` or `while` leaves the whole function, not only the loop:

```ant
func firstNegative(xs: Array[int]): int | null {
  for x in xs {
    if x < 0 { return x }
  }
  null
}

println(firstNegative([4, -2, -7]))
println(firstNegative([1, 2]))
```

```text
-2
null
```

### Inside a lambda

`return` in a lambda gives the lambda's value for that call. The function around it keeps running:

```ant
func scores(xs: Array[int]): Array[int] {
  xs.map({
    if x < 0 { return 0 }
    x * 10
  })
}

println(scores([1, -2, 3]))
```

```text
[10, 0, 30]
```

An anonymous function is a boundary too:

```ant
let sign = func(n: int): string {
  if n < 0 { return "-" }
  if n > 0 { return "+" }
  "0"
}
sign(-3) // "-"
sign(0)  // "0"
```

### After `??`

`return` has type `never`, so it can be the right side of `??`. When the left side is `null`, the
function returns; otherwise the value is unwrapped:

```ant
type User = { name: string }
let users = Map{ 1: User{ name: "Ann" } }

func greeting(id: int): Result[string, Error] {
  let user = users.get(id) ?? return Result.fail("no user ${id}")
  Result.ok("Hello, ${user.name}")
}

println(greeting(1))
println(greeting(2))
```

```text
Ok{ value: "Hello, Ann" }
Err{ error: Error{ message: "no user 2" } }
```

### In an `if` expression

One branch of an `if` expression can return while the other gives the value:

```ant
func label(n: int): string {
  let size = if n < 0 { return "invalid" } else if n < 10 { "small" } else { "large" }
  "${size} (${n})"
}

println(label(-1))
println(label(3))
println(label(250))
```

```text
invalid
small (3)
large (250)
```

This form is shown to demonstrate the feature. When the returning branch comes first, a guard before
the main path usually reads better: `if n < 0 { return "invalid" }`, followed by
`let size = if n < 10 { "small" } else { "large" }`.

### Narrowing after a guard

An `if` whose branch ends in `return` narrows the rest of the block. Below the guard, `port` is an
`int`, not an `int | null`:

```ant
func parsePort(text: string): Result[int, Error] {
  let port = text.toInt()
  if port == null { return Result.fail("not a number: ${text}") }
  if port < 1 || port > 65535 { return Result.fail("out of range: ${port}") }
  Result.ok(port)
}

println(parsePort("8080"))
println(parsePort("http"))
println(parsePort("70000"))
```

```text
Ok{ value: 8080 }
Err{ error: Error{ message: "not a number: http" } }
Err{ error: Error{ message: "out of range: 70000" } }
```

### With `defer`

The deferred blocks run when `return` leaves the function:

```ant
func load(): string {
  defer { println("closing") }
  println("reading")
  return "data"
}

println(load())
```

```text
reading
closing
data
```

A `defer` block is a boundary of its own. A `return` in it ends that block only, and takes no value:

```ant
func work(verbose: bool) {
  defer {
    if !verbose { return }
    println("done")
  }
  println("working")
}

work(false)
work(true)
```

```text
working
working
done
```

### In a script

The top-level code of a script is a body too. `return` ends the script early:

```ant
let args = ["--help"]
if args.contains("--help") {
  println("usage: tool [file]")
  return
}
println("running")
```

```text
usage: tool [file]
```

A script that returns an `Err` writes its message to standard error and exits with status 1:

<!-- test: exit 1 -->

```ant
let input = ""
if input.isEmpty() {
  return Result.fail("no input")
}
println("processing ${input}")
```

## Applications

Validate a record with a guard per rule, and keep the successful path at the end:

```ant
type Signup = { name: string, age: int }

func validate(s: Signup): Result[Signup, Error] {
  if s.name.trim().isEmpty() { return Result.fail("name is required") }
  if s.age < 13 { return Result.fail("${s.name}: too young") }
  Result.ok(s)
}

println(validate(Signup{ name: "Ann", age: 30 }))
println(validate(Signup{ name: " ", age: 30 }))
println(validate(Signup{ name: "Bob", age: 9 }))
```

```text
Ok{ value: Signup{ name: "Ann", age: 30 } }
Err{ error: Error{ message: "name is required" } }
Err{ error: Error{ message: "Bob: too young" } }
```

Stop a search at the first hit in a nested loop:

```ant
func findPair(xs: Array[int], target: int): Array[int] | null {
  for i in Array.range(0, xs.length()) {
    for j in Array.range(i + 1, xs.length()) {
      if xs[i] + xs[j] == target { return [xs[i], xs[j]] }
    }
  }
  null
}

println(findPair([3, 9, 4, 6], 10))
println(findPair([1, 2], 10))
```

```text
[4, 6]
null
```

## Properties and relations

In a function body, a `return` of the last expression gives the same result as the expression alone:

```ant
func square(n: int): int { n * n }
func squareReturn(n: int): int { return n * n }

assert(square(7) == squareReturn(7))
```

## Possible issues

A `return` at the end of a line returns `void`. The expression on the next line is not part of it:

```ant
func double(n: int): int {
  return
  // error: cannot use void where int is expected
  n * 2
}
```

Keep the value on the same line as `return`, or wrap a long one in parentheses after it:

```ant
func double(n: int): int {
  return (
    n * 2
  )
}

println(double(21))
```

```text
42
```

`return` takes everything to its right, so `x ?? return a + b` returns `a + b`, not `a`:

```ant
func orDefault(n: int | null): int {
  let value = n ?? return 1 + 100
  value
}

println(orDefault(null))
```

```text
101
```

`return` in a lambda does not leave the function around it. To stop at the first match, use
[Array.find](../../lib/Array/find.md) or a `for` loop instead of returning from a callback.

A script returns only `void` or an `Err`:

```ant
println("hi")
return 5 // error: a script returns void or an Err, not int
```

A `go` block has no result, so its `return` takes no value:

```ant
func start() {
  go { return 1 } // error: 'return' takes no value in a go block
}
```

A field default is not inside any function, so `return` there is an error:

```ant
type Config = { port: int = return 80 } // error: 'return' in a field default
```

## See also

- [func](func.md)
- [try](try.md)
- [defer](defer.md)
- [break](break.md)
- [continue](continue.md)
- [Null coalescing](../operators/null-coalescing.md)
- [Lambdas](../literals/lambdas.md)
- [never](../types/never.md)
