<!-- status: reviewed -->

# Functions

A function takes values, runs a block, and gives back the block's last expression. This program
declares a few and calls them in different ways:

```ant
func greet(name: string, greeting: string = "Hello", mark: string = "!"): string {
  "${greeting}, ${name}${mark}"
}

println(greet("Ann"))
println(greet("Bob", "Hi"))
println(greet("Cy", mark: "?"))
println(greet(mark: ".", name: "Dee"))

func twice(f: func(int): int, n: int): int {
  f(f(n))
}

let square = func(n: int): int { n * n }
println(twice(square, 3))
println(twice({ x + 10 }, 1))
println(Array.map([1, 2, 3], { x * x }))

func counter(): func(): int {
  var n = 0
  func() {
    n += 1
    n
  }
}

let next = counter()
next()
next()
println(next())
```

```text
Hello, Ann!
Hi, Bob!
Hello, Cy?
Hello, Dee.
81
21
[1, 4, 9]
3
```

## Declaring and calling

`func greet` declares a function. Each parameter has a type, and `: string` after the parentheses is
the return type. The body has no `return`: its last expression is the result. Use `return` only to
leave early. You can leave out the return type, as the function inside `counter` does; Ant infers it
from the body. An exported or recursive function states it.

`greeting` and `mark` have default values, so a call can leave them out. Arguments go in order, or
by name with `mark: "?"`. Named arguments let you skip a parameter in the middle, or write them in
any order, as the call for `"Dee"` does.

## Functions are values

`twice` takes a function as its first parameter. Its type, `func(int): int`, says what the function
accepts and returns. Any function of that type fits: a declared function by its name, `square`,
which binds a `func` without a name, or a lambda.

A lambda is a bare block, `{ x + 10 }`. It takes at most one parameter, always named `x`, and its
type comes from where you use it: `twice` expects `func(int): int`, so `x` is an `int`. `Array.map`
passes each element as `x` in the same way.

`counter` returns a function that keeps its own `n`. The returned function captures the variable,
not its value, so each call sees the count the last call left.

## Where it goes wrong

A lambda needs a place that says its type. Bound to a name with no type, it does not compile:

```ant
let double = { x * 2 }
// error: a lambda needs an expected function type; annotate the variable or use func with typed parameters
```

Write `let double: func(int): int = { x * 2 }`, or use `func(n: int): int { n * 2 }`.

Named arguments work on a declared function, not on a function value. A parameter is a `let`
binding, so assigning one does not compile either:

```ant
let square = func(n: int): int { n * n }
square(n: 3)
// error: a call through a function value is positional; named arguments apply only to a declared function

func bump(n: int): int {
  n = n + 1 // error: cannot assign to n, a parameter; declare a var to hold a value that changes
  n
}
```

Named arguments come after the positional ones:

```ant
greet(greeting: "Hi", "Ann") // error: a positional argument cannot follow a named argument
```

## Try it

- Greet one more person with `greeting: "Welcome"`.
- Call `twice` with a lambda that halves its argument.
- Make two counters and check that they count separately.

## Next

[Calls](07-calls.md) shows `xs.map(...)` and the pipe. The [functions guide](../guide/functions.md)
goes further, and the reference has [func](../language/keywords/func.md),
[lambdas](../language/literals/lambdas.md), [function types](../language/types/functions.md), and
[Array.map](../lib/Array/map.md).
