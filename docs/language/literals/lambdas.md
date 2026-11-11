<!-- status: reviewed -->

# Lambdas

A lambda is a bare block, `{ ... }`, used as a function value. It has one implicit parameter, `x`,
and its last expression is its result. An anonymous function, `func` without a name, is the general
form: it names and annotates its parameters.

## Details

- A lambda takes its parameter type and arity from the function type the context expects: a
  parameter, an annotated binding, a record field, a return type. With no expected type, a lambda is
  a compile error.
- A lambda has zero or one parameter. Where the expected function takes none, `x` is not in scope.
  For two or more parameters, write `func` with named parameters.
- `x` and the parameters of an anonymous function are `let` bindings: assigning one is a compile
  error.
- `return` inside a lambda leaves the lambda only. A `break` or `continue` inside a lambda cannot
  reach a loop outside it.
- A lambda or anonymous function captures variables by reference: it sees later assignments to them,
  and its own assignments are visible outside.

## Basic examples

Pass a lambda where a function is expected; `x` is the element:

```ant
Array.map([1, 2, 3], { x * x })               // [1, 4, 9]
Array.filter(["ant", "", "bee"], { x != "" }) // ["ant", "bee"]
```

Bind a lambda to a name with an annotated type:

```ant
let double: func(int): int = { x * 2 }
double(21) // 42
```

An anonymous function annotates its parameters, so it needs no expected type:

```ant
let add = func(a: int, b: int): int { a + b }
add(2, 3) // 5
```

## Scope

### Where the type comes from

The expected type may come from a function's parameter, an annotated binding, a record field, an
array or map element, or a return type:

```ant
type Button = { label: string, onClick: func(): string }

func adder(n: int): func(int): int { { x + n } }

let save = Button{ label: "Save", onClick: { "saved" } }
let steps: Array[func(int): int] = [{ x + 1 }, { x * 2 }]
let ops: Map[string, func(int): int] = Map{ "double": { x * 2 }, "negate": { -x } }

println(save.onClick())
println(adder(3)(4))
println(steps.map({ x(10) }))
println(ops["negate"]?.(5))
```

```text
saved
7
[11, 20]
-5
```

### The parameter `x`

`x` has the parameter type, so dot calls and field access work on it:

```ant
type User = { name: string, age: int }
let users = [User{ name: "Bob", age: 17 }, User{ name: "Ann", age: 31 }]
users.map({ x.name })                    // ["Bob", "Ann"]
users.some({ x.age < 18 })               // true
[" a ", "b "].map({ x.trim() })          // ["a", "b"]
users.sortBy({ x.name }).map({ x.name }) // ["Ann", "Bob"]
```

On a union, test the member with `is`:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
let shapes: Array[Shape] = [Circle{ r: 1.0 }, Square{ side: 2.0 }]
shapes.filter({ x is Circle }) // [Circle{ r: 1.0 }]
```

### No parameter

Where the expected function takes no argument, the lambda is a plain block of code:

```ant
let greet: func(): string = { "hello" }
greet() // "hello"
```

```ant
let lock = Mutex.new()
var hits = 0
let now = lock.withLock({
  hits += 1
  hits
})
println(now)
```

```text
1
```

### The result

The last expression is the result. When the expected result is a type parameter, the body fixes it:

```ant
[1, 2].map({ "#${x}" }) // ["#1", "#2"]
[1, 2].map({ x > 1 })   // [false, true]
```

`return` leaves the lambda early with its value:

```ant
let labels = [3, -1].map({
  if x < 0 { return "negative" }
  "ok"
})
labels // ["ok", "negative"]
```

A lambda whose expected result is `void` ends in a statement, or discards its last value with
`_ = ...`:

```ant
let seen = Set.of(["a"])
let remember: func(string): void = { _ = seen.add(x) }
remember("b")
seen // Set.of(["a", "b"])
```

### Nested lambdas

Each lambda binds its own `x`, so an inner lambda hides the outer `x`:

```ant
[[1, 2], [3]].map({ x.map({ x * 10 }) }) // [[10, 20], [30]]
```

When the inner body needs the outer value too, name it with `func`:

```ant
let rows = [[1, 2], [3]]
rows.map(func(row: Array[int]): Array[int] { row.map({ x * row.length() }) }) // [[2, 4], [3]]
```

### Several parameters

A lambda has at most one parameter. A callback of two or more takes an anonymous function:

```ant
[1, 2, 3].reduce(0, func(sum: int, n: int): int { sum + n }) // 6
[3, 1, 2].sortWith(func(a: int, b: int): int { b - a })      // [3, 2, 1]
```

### Closures

A lambda reads and assigns the variables around it. They are captured by reference:

```ant
var calls = 0
let bump: func(): void = { calls += 1 }
bump()
bump()
println(calls)

var base = 1
let addBase: func(int): int = { x + base }
base = 100
println(addBase(1))
```

```text
2
101
```

### Inside strings

Braces nest inside `${...}`, so a lambda works in an interpolation:

```ant
let xs = [1, 2]
"doubled: ${xs.map({ x * 2 })}" // "doubled: [2, 4]"
```

## Applications

Chain lambdas through a pipeline:

```ant
type User = { name: string, age: int }
let users = [User{ name: "Bob", age: 17 }, User{ name: "Ann", age: 31 }]
users
  | Array.filter({ x.age >= 18 })
  | Array.map({ x.name.toUpper() })
  | String.join(", ") // "ANN"
```

Look up an operation by name in a table of lambdas:

```ant
let ops: Map[string, func(int): int] = Map{ "double": { x * 2 }, "square": { x * x } }

func apply(name: string, n: int): int | null { ops[name]?.(n) }

println(apply("square", 7))
println(apply("cube", 7))
```

```text
49
null
```

## Properties and relations

A lambda and an anonymous function with one parameter of the same type do the same job:

```ant
let xs = [1, 2, 3]
assert(xs.map({ x * 2 }) == xs.map(func(n: int): int { n * 2 }))
```

Passing a lambda that only calls a function is the same as passing the function:

```ant
func square(n: int): int { n * n }
let xs = [1, 2, 3]
assert(xs.map({ square(x) }) == xs.map(square))
```

## Possible issues

A lambda with no expected type is a compile error. Annotate the binding, or write `func`:

```ant
let triple = { x * 3 }
// error: a lambda needs an expected function type; annotate the variable or use func with typed parameters
```

A `{` at the start of a statement starts a lambda, not a block:

```ant
{ println("hi") }
// error: a lambda needs an expected function type; annotate the variable or use func with typed parameters
```

Braces without a type name are a lambda, not a record. A record literal names its type:

```ant
let p = { x: 1 }
// error: a record or map literal names its type, as in Point{ x: 1 } or Map{ "a": 1 }
```

A lambda cannot take two parameters:

```ant
let total = [1, 2, 3].reduce(0, { x + 1 })
// error: a lambda has at most one parameter, x; func(int, int): int expects 2: write func with named parameters
```

Where the expected function takes no argument, there is no `x`:

```ant
let g: func(): string = { x } // error: x is not defined: a lambda that takes no argument has no x
```

`x` is a `let` binding:

```ant
let next: func(int): int = {
  x = x + 1 // error: cannot assign to x, a lambda's parameter; declare a var to hold a value that changes
  x
}
```

A lambda that must return `void` cannot end in a value. Discard it with `_ = ...`:

```ant
let seen = Set.of(["a"])
let remember: func(string): void = { seen.add(x) } // error: cannot use bool where void is expected
```

`return` in a lambda leaves the lambda, not the enclosing function. Here `firstNegative` always
reaches `null`:

```ant
func firstNegative(xs: Array[int]): int | null {
  _ = xs.map({
    if x < 0 { return x }
    0
  })
  null
}

println(firstNegative([3, -2]))
println([3, -2].find({ x < 0 }))
```

```text
null
-2
```

Use `find`, which returns the first match. A `for` loop also works, since its `return` leaves the
function.

A lambda's `x` hides an outer `x`. Name the parameter with `func`, or rename the outer variable:

```ant
let x = 5
[1, 2].map({ x + 1 }) // [2, 3]
```

A `break` in a lambda cannot reach a loop outside it:

```ant
for row in [[1, 2], [3]] {
  _ = row.map({
    if x == 2 { break } // error: 'break' outside a loop
    x
  })
}
```

## See also

[func](../keywords/func.md), [return](../keywords/return.md),
[function types](../types/functions.md), [pipe](../operators/pipe.md), [records](records.md),
[Array.map](../../lib/Array/map.md), [Array.reduce](../../lib/Array/reduce.md),
[Mutex.withLock](../../lib/Mutex/withLock.md), [Functions](../../guide/functions.md)
