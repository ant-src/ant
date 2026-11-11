<!-- status: reviewed -->

# void

`void` is a type with one value, also written `void`, which carries no information. It is the result
of code that runs for its effect: a function with nothing to return, an assignment, a loop.

## Details

- A function with nothing to return, a bare `return`, a block that ends in a declaration or an
  assignment, a loop, and an `if` without `else` whose condition is false all give `void`.
- `void` is assignable only to `void`. It is not `null`, and `??` does not apply to it.
- A function type with a `void` result, such as `func(int): void`, accepts only a function whose
  result is `void`. A callback whose last expression has a value discards it with `_ = ...`.
- `void` renders as `void`, and all `void` values are equal. Inside `Html`, it renders nothing.

## Basic examples

A function with nothing to return gives `void`:

```ant
func greet(name: string) {
  println("Hello, ${name}")
}

let result = greet("Ann")
println(result)
println(result == void)
```

```text
Hello, Ann
void
true
```

Write `void` where a value is required, as in a `Result` that only reports success or failure:

```ant
let saved: Result[void, Error] = Result.ok(void)
saved // Ok{ value: void }
```

## Scope

### Functions that mutate

Many functions that change their argument, such as `Array.push` and `Map.set`, return `void`:

```ant
let xs = [1, 2]
let pushed = xs.push(3)
pushed == void // true
xs             // [1, 2, 3]
```

### Early `return`

A `return` with nothing after it on its line returns `void`, so it exits early from a function that
has nothing to return:

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

### A `void` return type

A function may state `void` as its return type. Its body then must end in a `void` expression:

```ant
func log(message: string): void {
  println("[log] ${message}")
}

log("started")
```

```text
[log] started
```

### `Result[void, E]`

A function that can fail but has nothing to return gives `Result[void, E]`. `try` passes on the
error and discards the `void`:

```ant
func validate(age: int): Result[void, Error] {
  if age < 0 { return Result.fail("age is negative") }
  Result.ok(void)
}

func register(age: int): Result[string, Error] {
  try validate(age)
  Result.ok("registered")
}

println(register(30))
println(register(-1))
```

```text
Ok{ value: "registered" }
Err{ error: Error{ message: "age is negative" } }
```

### `if` without `else`

An `if` without `else` gives `void` when its condition is false, so its type is a union with `void`.
`is void` tells the two apart:

```ant
let n = 3
let size = if n > 5 { "big" }
size is void // true
```

### Callbacks

A parameter of type `func(T): void` takes a lambda whose body ends in a statement, such as an
assignment or a call that returns `void`:

```ant
func each(xs: Array[int], f: func(int): void) {
  for x in xs { f(x) }
}

var total = 0
let doubled: Array[int] = []
each([1, 2, 3], { total += x })
each([1, 2, 3], { doubled.push(x * 2) })
println(total)
println(doubled)
```

```text
6
[2, 4, 6]
```

### Channels as signals

A `Channel[void]` carries no data, only the fact that something happened:

```ant
let done: Channel[void] = Channel.new()
go {
  println("working")
  done.send(void)
}
done.receive()
println("finished")
```

```text
working
finished
```

### In `Html`

`void` renders nothing, so an `if` without `else` inside markup shows its content only when the
condition holds:

```ant
func list(items: Array[string]): Html {
  <ul>{items.map({ <li>{x}</li> })}{if items.isEmpty() { "nothing yet" }}</ul>
}

println(list([]))
println(list(["milk"]))
```

```text
<ul>nothing yet</ul>
<ul><li>milk</li></ul>
```

## Properties and relations

`void` equals itself, renders as its name, and is the value of every assignment:

```ant
var n = 0
let assigned = func() { n = 1 }
assert(void == void)
assert(String.from(void) == "void")
assert(assigned() == void)
println("ok")
```

```text
ok
```

## Possible issues

### Using the result of a `void` function

```ant
func log(message: string) {
  println(message)
}

let n: int = log("x") // error: cannot use void where int is expected
```

### A callback that ends in a value

`Set.add` returns a `bool`, so a lambda that ends in it does not fit `func(int): void`:

```ant
func each(xs: Array[int], f: func(int): void) {
  for x in xs { f(x) }
}

let seen: Set[int] = Set.of([])
each([1, 2, 2], { seen.add(x) }) // error: cannot use bool where void is expected
```

Discard the value with `_ =`, which is an assignment and so gives `void`:

```ant
func each(xs: Array[int], f: func(int): void) {
  for x in xs { f(x) }
}

let seen: Set[int] = Set.of([])
each([1, 2, 2], { _ = seen.add(x) })
println(seen)
```

```text
Set.of([1, 2])
```

### An `if` without `else` where a value is required

```ant
let n = 3
let label: string = if n > 5 { "big" }
// error: an if without else has type void when the condition is false, which cannot be used where string is expected
```

Add the `else` branch:

```ant
let n = 3
let label: string = if n > 5 { "big" } else { "small" }
label // "small"
```

### `void` is not `null`

`??` needs a nullable value, and `void` is not one:

```ant
func log(message: string) {
  println(message)
}

let shown = log("x") ?? "none" // error: ?? needs a nullable left side, got void
```

### Mapping with a `void` function

`map` keeps every result, so a callback that only prints gives an array of `void`:

```ant
let printed = [1, 2].map({ println(x) })
println(printed)
```

```text
1
2
[void, void]
```

Loop with `for` instead when you want only the effect:

```ant
for x in [1, 2] {
  println(x)
}
```

```text
1
2
```

## See also

- [never](never.md)
- [null](../literals/null.md)
- [Function types](functions.md)
- [return](../keywords/return.md)
- [if](../keywords/if.md)
- [Assignment](../operators/assignment.md)
- [Result](../../lib/Result/index.md)
