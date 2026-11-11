<!-- status: reviewed -->

# Function types

`func(A, B): R` is the type of a function that takes an `A` and a `B` and returns an `R`. Functions
are values: a variable, a parameter, a record field, or a map can hold one.

## Details

- Parameter names in a function type are optional and only document intent. A call through a
  function value is positional, so named arguments do not apply to it.
- The return type extends as far right as it can: `func(): int | null` returns `int | null`. Write
  `(func(): int) | null` for a function that may be missing.
- A function is assignable to a function type with the same number of parameters when it accepts
  every argument the type promises and returns only what the type allows. A function with default
  parameters also fits a type that lists its required parameters and any prefix of the optional
  ones.
- A `void` result gets no exception: a function that returns a value is not a `func(T): void`.
- `==` compares functions by identity. `is func` matches any function, whatever its signature.

## Basic examples

Annotate a variable with a function type, and a lambda takes its parameter type from it:

```ant
let double: func(int): int = { x * 2 }
double(21) // 42
```

A parameter of function type takes a lambda or a named function:

```ant
func applyTwice(f: func(int): int, n: int): int { f(f(n)) }
applyTwice({ x + 3 }, 1) // 7
func square(n: int): int { n * n }
applyTwice(square, 3) // 81
```

A type alias names a function type:

```ant
type Predicate[T] = func(T): bool
let isEven: Predicate[int] = { x % 2 == 0 }
[1, 2, 3, 4].filter(isEven) // [2, 4]
```

## Scope

### Parameter lists

A function type lists any number of parameters. A lambda fits a type with zero or one parameter; for
more, write an anonymous `func`:

```ant
let tick: func(): string = { "tick" }
tick() // "tick"
let add: func(int, int): int = func(a: int, b: int): int { a + b }
add(2, 3) // 5
```

### Parameter names

Names document what each parameter means. They do not change the type:

```ant
type Handler = func(path: string, attempt: int): string
let fetch: Handler = func(p: string, n: int): string { "${p} #${n}" }
fetch("/users", 2) // "/users #2"
```

### Named functions and functions on types

A declared function is a value of its function type, and so is a function on a type:

```ant
type Point = { x: float64, y: float64 }
func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }
let measure: func(Point): float64 = Point.norm
measure(Point{ x: 3.0, y: 4.0 }) // 5.0
```

### Constructors

A named primitive type's constructor is a function from the base type to the named type:

```ant
type UserId = string
let toId: func(string): UserId = UserId
toId("u-1")            // UserId("u-1")
["a", "b"].map(UserId) // [UserId("a"), UserId("b")]
```

### Functions that return functions

A function type can be the result of another. The returned lambda takes its type from the declared
result, and captures `n`:

```ant
func adder(n: int): func(int): int {
  { x + n }
}
let add10 = adder(10)
[1, 2, 3].map(add10) // [11, 12, 13]
```

### Generic function types

A function type can mention the type parameters of the function or alias that declares it:

```ant
func compose[A, B, C](f: func(A): B, g: func(B): C): func(A): C {
  { g(f(x)) }
}
let clean = compose(String.trim, String.toUpper)
clean("  ant ") // "ANT"
```

### Record fields and maps

A record field of function type is called like a function:

```ant
type Button = { label: string, onClick: func(): void }
var clicks = 0
let ok = Button{ label: "OK", onClick: { clicks += 1 } }
ok.onClick()
ok.onClick()
println(clicks)
```

```text
2
```

A map value of function type is read by key; the key may be missing, so call it with `?.(...)`:

```ant
let ops: Map[string, func(int): int] = Map{ "inc": { x + 1 } }
ops["inc"]?.(1) // 2
```

### Nullable functions

Parenthesize the function type to put it in a union with `null`, and call it with `?.(...)`:

```ant
let onError: (func(string): string) | null = null
onError?.("disk full") // null
let log: (func(string): string) | null = { "logged: ${x}" }
log?.("disk full") // "logged: disk full"
```

### Unions with other types

`is func` tests for a function of any signature:

```ant
type Lazy = int | (func(): int)
func force(v: Lazy): int {
  if v is func { v() } else { v }
}
force(4)      // 4
force({ 40 }) // 40
```

### Rendering

`String.from` renders a function value as `func`, and a constructor as its type's name:

```ant
let inc: func(int): int = { x + 1 }
String.from(inc) // "func"
```

```ant
type UserId = string
String.from(UserId) // "UserId"
```

## Applications

### A table of operations

Keep functions in a map and pick one by key:

```ant
let ops: Map[string, func(int, int): int] = Map{
  "+": func(a: int, b: int): int { a + b },
  "-": func(a: int, b: int): int { a - b },
  "*": func(a: int, b: int): int { a * b },
}

func calc(op: string, a: int, b: int): int | null {
  let f = ops[op] ?? return null
  f(a, b)
}

println(calc("*", 6, 7))
println(calc("/", 6, 7))
```

```text
42
null
```

### Validation rules

Keep named checks in records, each a function from the input to `bool`:

```ant
type Rule = { name: string, check: func(string): bool }

let rules = [
  Rule{ name: "not empty", check: { !x.isEmpty() } },
  Rule{ name: "at most 8 characters", check: { x.length() <= 8 } },
  Rule{ name: "lowercase", check: { x == x.toLower() } },
]

func failures(input: string): Array[string] {
  rules
    | Array.filter({ !x.check(input) })
    | Array.map({ x.name })
}

println(failures("ant"))
println(failures("Antelopes"))
```

```text
[]
["at most 8 characters", "lowercase"]
```

## Properties and relations

A function type accepts a function whose parameter is wider and whose result is narrower:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle { 3.0 * s.radius * s.radius } else { s.width * s.height }
}
func unitCircle(): Circle { Circle{ radius: 1.0 } }

let circleArea: func(Circle): float64 = area
let makeShape: func(): Shape = unitCircle
assert(circleArea(Circle{ radius: 1.0 }) == area(Circle{ radius: 1.0 }))
assert(makeShape() is Circle)
```

A function with default parameters fits a type for each prefix of its optional ones:

```ant
func greet(name: string, greeting: string = "Hello", mark: string = "!"): string {
  "${greeting}, ${name}${mark}"
}

let short: func(string): string = greet
let medium: func(string, string): string = greet
let full: func(string, string, string): string = greet
assert(short("Ann") == "Hello, Ann!")
assert(medium("Ann", "Hi") == "Hi, Ann!")
assert(full("Ann", "Hi", "?") == "Hi, Ann?")
```

`==` is identity: the same function is equal to itself, and two equal-looking lambdas are not equal:

```ant
let f: func(int): int = { x + 1 }
let g = f
let h: func(int): int = { x + 1 }
assert(f == g)
assert(f != h)
```

## Possible issues

### A narrower parameter does not fit

A function that takes only a `Circle` cannot stand where any `Shape` may be passed:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func circleArea(c: Circle): float64 { 3.0 * c.radius * c.radius }
let area: func(Shape): float64 = circleArea
// error: cannot use func(Circle): float64 where func(Shape): float64 is expected
```

### The parameter count must match

```ant
let add: func(int): int = func(a: int, b: int): int { a + b }
// error: cannot use func(int, int): int where func(int): int is expected
```

### A lambda has at most one parameter

```ant
let add: func(int, int): int = { x } // error: a lambda has at most one parameter, x; func(int, int): int expects 2
```

### A lambda needs a function type

Without an expected type, a lambda has no parameter type:

```ant
let triple = { x * 3 } // error: a lambda needs an expected function type; annotate the variable or use func with typed parameters
```

### A value where `void` is expected

A callback typed `func(T): void` must not return a value. Discard the value with `_ = ...`:

```ant
let lines: Array[string] = []
func record(line: string): int {
  lines.push(line)
  lines.length()
}
let onLine: func(string): void = { record(x) } // error: cannot use int where void is expected
```

```ant
let lines: Array[string] = []
func record(line: string): int {
  lines.push(line)
  lines.length()
}
let onLine: func(string): void = { _ = record(x) }
onLine("started")
println(lines)
```

```text
["started"]
```

### The return type takes the whole union

`func(): int | null` is a function that returns `int | null`, not a nullable function. Its result
needs a `null` check:

```ant
let next: func(): int | null = { 1 }
let n = next() + 1 // error: + needs both operands of one type, got int | null and int
```

### Named arguments through a value

```ant
let add = func(a: int, b: int): int { a + b }
let sum = add(a: 1, b: 2) // error: a call through a function value is positional; named arguments apply only to a declared function
```

### A recursive function value needs its type

A function value that calls itself needs its type stated before its body: a return type on the
`func`, or a type on the binding.

```ant
let fact = func(n: int) { if n <= 1 { 1 } else { n * fact(n - 1) } } // error: fact is used by a function in its own initializer; state its type
```

```ant
let fact = func(n: int): int { if n <= 1 { 1 } else { n * fact(n - 1) } }
println(fact(5))
```

```text
120
```

### A function on a type is not a field

`p.norm` without a call is a field read, and `Point` has no field `norm`. Use `Point.norm`, or wrap
the call in a lambda:

```ant
type Point = { x: float64, y: float64 }
func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }
let p = Point{ x: 3.0, y: 4.0 }
let m = p.norm // error: Point has no field norm
```

## See also

- [func](../keywords/func.md)
- [Lambdas](../literals/lambdas.md)
- [Dot calls](../operators/dot-calls.md)
- [?. (optional chaining)](../operators/optional-chaining.md)
- [is](../operators/is.md)
- [void](void.md)
- [Functions](../../guide/functions.md)
- [Generics](../../guide/generics.md)
