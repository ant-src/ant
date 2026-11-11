<!-- status: reviewed -->

# never

`never` is the type of an expression that does not finish: `return`, `break`, `continue`, a call of
`panic`, and a call of any function declared to return `never`. `never` is the bottom type: it is
assignable to every type.

## Details

- A `never` expression fits wherever a value is expected: in an `if` branch, after `??`, or as a
  function's last expression.
- An `if` drops a `never` branch from its type: `if ok { 1 } else { panic("no") }` is an `int`.
- When an `if` without `else` has a `never` branch, the rest of the block gets the opposite of its
  condition, so `if user == null { fail("...") }` leaves `user` non-null below it.
- `while true` with no `break` in its body has type `never`.

## Basic examples

A function that never returns declares `never` as its result:

```ant
func fail(message: string): never {
  panic("fatal: ${message}")
}

fail("no config") // panic: fatal: no config
```

A `never` expression takes the type the context needs. Here the `if` is an `int`:

```ant
let ready = true
let count = if ready { 3 } else { panic("not ready") }
count + 1 // 4
```

## Scope

### Alongside `??`

`return`, `panic`, and a `never` function fit on the right of `??`. The result has the type of the
left side with `null` removed:

```ant
func fail(message: string): never {
  panic(message)
}

func port(text: string): int {
  let n = text.toInt() ?? fail("not a number: ${text}")
  if n < 1 || n > 65535 { fail("out of range: ${n}") }
  n
}

port("8080") // 8080
port("http") // panic: not a number: http
port("0")    // panic: out of range: 0
```

### Guards

A guard whose body is `never` narrows the value below it, the same as one that ends in `return`:

```ant
func fail(message: string): never {
  panic(message)
}

func lookup(env: Map[string, string], key: string): string {
  let value = env[key]
  if value == null { fail("${key} is not set") }
  value.trim()
}

let env = Map{ "HOME": " /home/ann " }
lookup(env, "HOME") // "/home/ann"
lookup(env, "USER") // panic: USER is not set
```

### `return`, `break`, and `continue`

Each has type `never`, so each can stand where a value is expected:

```ant
func firstWord(line: string): string {
  let words = line.fields()
  let first = words.first() ?? return "(empty)"
  first.toUpper()
}

firstWord("hello world") // "HELLO"
firstWord("   ")         // "(empty)"
```

### Inferred `never`

A function without a result annotation whose body ends in `panic` gets the inferred result type
`never`, so a quick local helper works on the right of `??`:

```ant
func fail(message: string) {
  panic(message)
}

func parse(text: string): int {
  let n: int = text.toInt() ?? fail("bad number: ${text}")
  n
}

parse("42")  // 42
parse("abc") // panic: bad number: abc
```

An exported function states `: never` itself, because exported functions state their result type.

### Endless loops

A `while true` loop with no `break` never ends on its own, so its type is `never`. A function can
end on one and still declare a result type, as long as every way out is a `return`:

```ant
func firstSquareAbove(limit: int): int {
  var n = 1
  while true {
    if n * n > limit { return n }
    n += 1
  }
}

firstSquareAbove(50) // 8
```

### Function values

A function value returning `never` is assignable to a function type with any result, since `never`
is assignable to that result:

```ant
func reject(input: string): never {
  panic("unexpected input: ${input}")
}

let measures: Array[func(string): int] = [{ x.length() }, reject]
measures[0]("ant") // 3
measures[1]("ant") // panic: unexpected input: ant
```

### Complete chains of `is`

When the tests of an `if` chain cover every member of a union, the last test leaves the subject
`never`, and the `if` needs no `else`. Its type is the union of the branches, without `void`:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square

func area(s: Shape): float64 {
  if s is Circle { 3.0 * s.radius * s.radius } else if s is Square { s.side * s.side }
}

area(Square{ side: 2.0 }) // 4.0
```

## Applications

### Exiting with a usage message

`Os.exit` returns `never`, so a guard that calls it narrows like a `return`:

<!-- test: exit 2 -->

```ant
let name = Os.args().first()
if name == null {
  eprintln("usage: greet <name>")
  Os.exit(2)
}
println("Hello, ${name}")
```

### An unreachable case

A helper that panics with a fixed message marks a path the caller rules out; as a guard it narrows
like `return`, and inside an expression it takes the type the context needs:

```ant
func unreachable(what: string): never {
  panic("unreachable: ${what}")
}

func weekday(n: int): string {
  if n < 0 || n > 6 { unreachable("weekday ${n}") }
  ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"][n]
}

weekday(1) // "Mon"
weekday(9) // panic: unreachable: weekday 9
```

## Properties and relations

A `never` branch does not change the type of an `if`, and a `never` expression is assignable to any
declared type:

```ant
func fail(message: string): never {
  panic(message)
}

let ok = true
let n: int = if ok { 1 } else { fail("no") }
let s: string = if ok { "a" } else { fail("no") }
let m: int | null = if ok { null } else { fail("no") }
assert(n == 1 && s == "a" && m == null)
println("ok")
```

```text
ok
```

## Possible issues

### A `never` function that finishes

A function declared `: never` may not end with a value, an empty body, or a bare `return`:

```ant
func fail(message: string): never {
  println(message) // error: cannot use void where never is expected
}
```

End it with `panic`, `Os.exit`, or a call of another `never` function.

### A loop with `break` is not `never`

A `break` anywhere in a `while true` body makes the loop `void`, so a function that ends on it needs
a value after the loop:

```ant
func firstNegative(xs: Array[int]): int {
  var i = 0
  while true { // error: cannot use void where int is expected
    if i >= xs.length() { break }
    if xs[i] < 0 { return xs[i] }
    i += 1
  }
}
```

### An `else` after a complete chain

Once the tests cover every member of a union, the subject is `never`, and an extra `else` cannot
run:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square

func name(s: Shape): string {
  if s is Circle { "circle" } else if s is Square { "square" } else { "other" }
  // error: this else cannot run: the tests before it cover every case
}
```

## See also

- [void](void.md)
- [return](../keywords/return.md)
- [break](../keywords/break.md)
- [continue](../keywords/continue.md)
- [while](../keywords/while.md)
- [if](../keywords/if.md)
- [is](../operators/is.md)
- [??](../operators/null-coalescing.md)
- [panic](../../lib/prelude/panic.md)
- [Os.exit](../../lib/Os/exit.md)
- [Panic.raise](../../lib/Panic/raise.md)
- [Unions](../../guide/unions.md)
