<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# assert

```ant signature
func assert(cond: bool, message: string = ""): void
```

assert panics when cond is false, with a message that quotes the condition, shows the values it was
computed from, and ends with message when it is not empty. message is evaluated only then. assert is
not a value: it can only be called by name.

<!-- end generated -->

## Details

- `cond` is evaluated once.
- The panic message starts with `assertion failed: ` and the condition as `ant fmt` prints it on one
  line. When the condition is a comparison or an `is` test, the two values follow in parentheses.
  The names, field accesses, indexes, and calls inside the condition that ran follow, one per line,
  with their values; the two compared values are not repeated there.
- `assert` is always on: no flag or build mode turns it off.
- After `assert(cond)`, the rest of the block is narrowed as after `if !cond { panic(...) }`.

## Basic examples

A condition that holds does nothing:

```ant
let total = 2 + 2
assert(total == 4)
println("ok")
```

```text
ok
```

A condition that fails panics. The message shows the condition and the values it compared:

```ant
let prices = [1, 2]
let total = prices.sum() ?? 0
assert(total == 4) // panic: assertion failed: total == 4 (3 == 4)
```

Add a message that says why the condition should hold. It goes at the end:

```ant
let prices = [1, 2]
let total = prices.sum() ?? 0
assert(total == 3, "the cart adds every price")
println(total)
```

```text
3
```

## Scope

A condition with no comparison and no names shows only itself:

```ant
assert(false) // panic: assertion failed: false
```

Strings are quoted, so a missing or extra character shows:

```ant
let name = "Ann"
let r = Panic.catch({ assert(name == "Anne", "the name is copied") })
if r is Err {
  println(r.error.message)
}
```

```text
assertion failed: name == "Anne" ("Ann" == "Anne"): the name is copied
```

For `is`, the right side shows the tag the value carries, or its kind when it has none:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square

let shape: Shape = Square{ side: 2.0 }
let r = Panic.catch({ assert(shape is Circle) })
if r is Err {
  println(r.error.message)
}
```

```text
assertion failed: shape is Circle (Square{ side: 2.0 } is Square)
```

A failed `!= null` names the side that is `null`:

```ant
let user: string | null = null
let r = Panic.catch({ assert(user != null) })
if r is Err {
  println(r.error.message)
}
```

```text
assertion failed: user != null (user is null)
```

The names, field accesses, indexes, and calls inside the condition follow one per line, in the order
they finished:

```ant
type Point = { x: int }
let p = Point{ x: 1 }
let xs = [1, 2]
let r = Panic.catch({ assert(p.x + xs[0] == len(xs) + 1) })
if r is Err {
  println(r.error.message)
}
```

```text
assertion failed: p.x + xs[0] == len(xs) + 1 (2 == 3)
  p.x = 1
  xs[0] = 1
  xs = [1, 2]
  len(xs) = 2
```

The arguments of a call show too, so a failed check on a helper shows its input:

```ant
func isSorted(xs: Array[int]): bool {
  xs == xs.sort()
}

let xs = [3, 1, 2]
let r = Panic.catch({ assert(isSorted(xs)) })
if r is Err {
  println(r.error.message)
}
```

```text
assertion failed: isSorted(xs)
  xs = [3, 1, 2]
```

A part that `&&`, `||`, `??`, or `?.` skipped did not run and is not shown:

```ant
let lo = 0
let mid = 5
let hi = 5
let r = Panic.catch({ assert(lo > 0 && mid < hi) })
if r is Err {
  println(r.error.message)
}
```

```text
assertion failed: lo > 0 && mid < hi
  lo = 0
```

## Named arguments

`message:` is evaluated only when the condition fails, so it may do work a passing check should not
pay for:

```ant
var built = 0
func describe(): string {
  built += 1
  "built ${built} time(s)"
}

assert(true, describe())
println(built)
let r = Panic.catch({ assert(false, message: describe()) })
if r is Err {
  println(r.error.message)
}
```

```text
0
assertion failed: false: built 1 time(s)
```

## Applications

Check a documented precondition at the top of a function, so that a bad call fails where it is made:

```ant
func average(xs: Array[float64]): float64 {
  assert(len(xs) > 0, "average of no numbers")
  (xs.sum() ?? 0.0) / float64(len(xs))
}

println(average([1.0, 2.0, 6.0]))
let none = average([]) // panic: assertion failed: len(xs) > 0 (0 > 0): average of no numbers
```

```text
3.0
```

State that a value cannot be `null` here. The rest of the block sees it narrowed:

```ant
type User = { name: string }

func find(id: int): User | null {
  if id == 1 { User{ name: "Ann" } } else { null }
}

let user = find(1)
assert(user != null, "user 1 is created at startup")
println(user.name)
```

```text
Ann
```

Check a loop invariant:

```ant
var lo = 0
var hi = 100
let target = 37
while lo < hi {
  let mid = (lo + hi) / 2
  assert(lo <= mid && mid < hi)
  if mid < target {
    lo = mid + 1
  } else {
    hi = mid
  }
}
println(lo)
```

```text
37
```

State the expectations of a test. A failed `assert` fails the test with the same message:

```ant [stack.ant]
export func top(xs: Array[int]): int {
  assert(len(xs) > 0, "top of an empty stack")
  xs[len(xs) - 1]
}
```

<!-- test: cmd ant test . -->

```ant [stack_test.ant]
import "./stack.ant" { top }

test "top returns the last element" {
  assert(top([1, 2, 3]) == 3)
}

test "top of an empty stack panics" {
  let r = Panic.catch({ top([]) })
  assert(r is Err)
  assert(r.error.message.contains("top of an empty stack"))
}
```

## Properties and relations

`assert(cond, message)` panics as `if !cond { panic(...) }` does, and both can be caught with
`Panic.catch`:

```ant
let a = Panic.catch({ assert(1 + 1 == 3) })
let b = Panic.catch({
  if !(1 + 1 == 3) {
    panic("assertion failed")
  }
})
assert(a is Err && b is Err)
```

A failed `assert` with a message ends its panic message with `: ` and the message:

```ant
let r = Panic.catch({ assert(1 + 1 == 3, "arithmetic works") })
assert(r is Err && r.error.message.endsWith(": arithmetic works"))
```

## Possible issues

`assert` is not a value, so it cannot be passed to a function:

```ant
let flags = [true, false]
let checked = flags.map(assert)
// error: assert is not a value; it can only be called, as assert(cond) or assert(cond, message)
```

The condition must be a `bool`, and the message a `string`. Ant has no truthy values:

```ant
let n = 3
assert(n)        // error: cannot use int where bool is expected
assert(n > 0, n) // error: cannot use int where string is expected
```

Do not assert input from a user, a file, or the network. A bad input is not a bug in the program;
return a `Result` instead:

```ant
func parseAge(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.err(Error{ message: "not a number: ${text}" })
  if n < 0 {
    return Result.err(Error{ message: "age is negative: ${n}" })
  }
  Result.ok(n)
}

println(parseAge("-4"))
println(parseAge("31"))
```

```text
Err{ error: Error{ message: "age is negative: -4" } }
Ok{ value: 31 }
```

## See also

[panic](panic.md) · [Panic.catch](../Panic/catch.md) · [Panic.raise](../Panic/raise.md) ·
[test](../../language/keywords/test.md) · [Result](../Result/index.md)
