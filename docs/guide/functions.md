<!-- status: reviewed -->

# Functions

Ant has one kind of function. A declared function, a function on a type, an anonymous `func`, and a
lambda are all values of a function type, and you can call, pass, store, and return any of them.
What differs is how you write them and how you call them. This chapter shows each form, when to pick
it, and the mistakes people bring from other languages.

## Declaring a function

```ant
func clamp(value: int, low: int, high: int): int {
  if value < low { return low }
  if value > high { return high }
  value
}

func label(n: int) {
  "${n} of 10"
}

println(clamp(15, 0, 10))
println(label(clamp(-3, 0, 10)))
```

```text
10
0 of 10
```

A body is a block, and its last expression is the result. Use `return` only to leave early, as the
guards in `clamp` do. Put those guards first and keep the ordinary result at the end.

Every parameter states its type. The return type after the parentheses is optional: `label` leaves
it out, and Ant infers `string` from the body. State it when the function is exported or calls
itself, where the checker needs it, and wherever it helps the reader.

`func` declares a function at the top level of a module. Inside a block, you bind a function value
to a name instead; see [Recursion and local functions](#recursion-and-local-functions).

## Parameters, defaults, and named arguments

```ant
func connect(host: string, port: int = 443, tls: bool = true): string {
  let scheme = if tls { "https" } else { "http" }
  "${scheme}://${host}:${port}"
}

println(connect("example.com"))
println(connect("example.com", 8080))
println(connect("example.com", tls: false))
println(connect(port: 80, host: "example.com", tls: false))
```

```text
https://example.com:443
https://example.com:8080
http://example.com:443
http://example.com:80
```

A parameter with a default value may be left out of a call. Arguments go in order, by name, or in
order followed by names. A name lets you skip a parameter in the middle, as `tls: false` skips
`port`.

A default is evaluated afresh on each call that leaves it out, so each call below gets a new array.
A default can read the parameters before it.

```ant
func tag(name: string, labels: Array[string] = []): Array[string] {
  labels.push(name)
  labels
}

tag("a") // ["a"]
tag("b") // ["b"]
```

Name an argument when its value alone does not say what it means.
`connect("example.com", 443, false)` leaves the reader guessing, and `tls: false` does not. A
`bool`, a bare number, and an optional setting are the usual candidates. The standard library
follows the same pattern, as in `Time.sleep(1s, ctx: ctx)`. A call through `.` takes names as the
namespace form does: `p.distance(b: origin)`.

Ant has no variadic parameters. A function that takes any number of values takes an array:

```ant
func total(prices: Array[int]): int {
  prices.sum() ?? 0
}

total([3, 4, 5]) // 12
total([])        // 0
```

A parameter is a `let` binding. The body may change the array or record a parameter holds, but it
cannot assign the parameter itself. To normalize an argument, bind a new name:

```ant
func greet(name: string): string {
  let clean = name.trim()
  if clean.isEmpty() { return "Hello, stranger" }
  "Hello, ${clean}"
}

greet("  Ann ") // "Hello, Ann"
greet("   ")    // "Hello, stranger"
```

## Functions are values

A function type, such as `func(string): bool`, lists the parameter types and the result type. Any
function of that type fits where one is expected, whichever way you wrote it:

```ant
func isShort(s: string): bool {
  s.length() <= 3
}

let words = ["ant", "beetle", "bee", "moth"]

println(words.filter(isShort))
println(words.filter({ x.length() <= 3 }))
println(words.filter(func(s: string): bool { s.length() <= 3 }))

let byLength = func(a: string, b: string): int { a.length() - b.length() }
println(words.sortWith(byLength))
println(words.map(String.toUpper))
```

```text
["ant", "bee"]
["ant", "bee"]
["ant", "bee"]
["ant", "bee", "moth", "beetle"]
["ANT", "BEETLE", "BEE", "MOTH"]
```

- A **declared function** passes by its name, `isShort`, and so does a function of a namespace,
  `String.toUpper`. Declare one when its job deserves a name or you use it in several places.
- A **lambda**, `{ x.length() <= 3 }`, is a bare block. It is the short form for a one-line
  callback.
- An **anonymous function**, `func(a: string, b: string): int { ... }`, is a declaration without a
  name. Use it when the function takes two or more parameters, or when no function type is expected.

Your own function can take a function too. A parameter of function type is called like any function,
and a lambda passed to it takes its type from that parameter:

```ant
func retry(times: int, attempt: func(): bool): bool {
  for _ in Iter.range(0, times) {
    if attempt() { return true }
  }
  false
}

var tries = 0
let ok = retry(5, {
  tries += 1
  tries == 3
})
println("${ok} after ${tries} tries")
```

```text
true after 3 tries
```

A name in a function type documents intent only, as in `func(request: Request): Response`. Calls
through a function value are positional.

## Lambdas and `x`

A lambda takes at most one parameter, and it is always named `x`. Its type comes from the place you
use it: `words.filter` expects `func(string): bool`, so `x` is a `string`. With no expected type
there is nothing to infer from, so a lambda bound to a plain `let` needs a type on the binding:

```ant
let double: func(int): int = { x * 2 }
double(21) // 42
```

A lambda also fits a function type with no parameters. There it has no `x`, and its block is the
body of the function:

```ant
var calls = 0
let tick: func(): int = {
  calls += 1
  calls
}
tick()
println(tick())
```

```text
2
```

`return` inside a lambda leaves the lambda, not the function around it. Here it gives the lambda's
result for one element, and `map` goes on to the next:

```ant
let xs = [3, -1, 2]
xs.map({
  if x < 0 { return 0 }
  x * 2
}) // [6, 0, 4]
```

Each lambda binds its own `x`, so a lambda inside another one hides the outer `x`. When the inner
callback needs the outer value, give that value a name with `func`:

```ant
type Order = { customer: string, items: Array[string] }

let orders = [
  Order{ customer: "Ann", items: ["tea", "jam"] },
  Order{ customer: "Bob", items: ["milk"] },
]

let lines = orders.flatMap(func(order: Order): Array[string] {
  order.items.map({ "${order.customer}: ${x}" })
})
println(lines)
```

```text
["Ann: tea", "Ann: jam", "Bob: milk"]
```

## Closures

A function captures the variables around it by reference, not by value. It sees every later
assignment to them:

```ant
var prefix = "a"
let label: func(int): string = { "${prefix}${x}" }
println(label(1))
prefix = "b"
println(label(1))
```

```text
a1
b1
```

That makes a function a place to keep state. Each call of `counter` below makes a fresh `n`, and the
function it returns is the only code that can reach it:

```ant
func counter(): func(): int {
  var n = 0
  func() {
    n += 1
    n
  }
}

let first = counter()
let second = counter()
first()
first()
println(first())
println(second())
```

```text
3
1
```

Each iteration of a `for` loop has its own loop variable, so closures made in a loop do not share
one. A JavaScript `var` in a loop would give `[3, 3, 3]` here:

```ant
let getters: Array[func(): int] = []
for i in Array.range(0, 3) {
  getters.push({ i })
}
println(getters.map({ x() }))
```

```text
[0, 1, 2]
```

## Recursion and local functions

A function may call itself, and top-level functions may call each other in any order. A recursive
function states its return type, because the checker cannot infer a type from a body that uses it:

```ant
type Node = { name: string, children: Array[Node] }

func depth(node: Node): int {
  1 + (node.children.map(depth).max() ?? 0)
}

let tree = Node{
  name: "root",
  children: [
    Node{ name: "a", children: [Node{ name: "a1", children: [] }] },
    Node{ name: "b", children: [] },
  ],
}
println(depth(tree))
```

```text
3
```

Inside a block, `func name` is not allowed. Bind the function to a `let` instead. A local function
that calls itself states its type on the binding. Here `walk` is local to `names` and collects into
`out`, which it captures:

```ant
type Node = { name: string, children: Array[Node] }

func names(root: Node): Array[string] {
  let out: Array[string] = []
  let walk: func(Node): void = func(node: Node) {
    out.push(node.name)
    for child in node.children { walk(child) }
  }
  walk(root)
  out
}

let tree = Node{
  name: "root",
  children: [Node{ name: "a", children: [] }, Node{ name: "b", children: [] }],
}
println(names(tree))
```

```text
["root", "a", "b"]
```

Keep a helper local when only one function uses it and it needs that function's variables. Move it
to the top level when it has a job of its own.

## Functions on types

`func Point.norm` declares `norm` in the namespace of the type `Point`. A type's name is a
namespace, and this is how you put functions in it:

```ant
type Point = { x: float64, y: float64 }

func Point.origin(): Point {
  Point{ x: 0.0, y: 0.0 }
}

func Point.distance(a: Point, b: Point): float64 {
  Math.sqrt((a.x - b.x) * (a.x - b.x) + (a.y - b.y) * (a.y - b.y))
}

func Point.norm(p: Point): float64 {
  p.distance(Point.origin())
}

let p = Point{ x: 3.0, y: 4.0 }
println(Point.norm(p))
println(p.norm())
println(p | Point.distance(Point{ x: 3.0, y: 0.0 }))
println([p, Point.origin()].map(Point.norm))
```

```text
5.0
5.0
4.0
[5.0, 0.0]
```

A function on a type is an ordinary function that lives in the type's namespace. It is not a method:
the value it works on is a parameter you declare and name. When that parameter comes first, the
function has three call forms, `Point.norm(p)`, `p.norm()`, and `p | Point.norm()`, which all call
the same function. It need not take the type at all: `Point.origin()` takes nothing and is called
through the type's name.

You can declare functions only on a type declared at the top level of your own module: a record, a
named primitive type such as `type UserId = string`, or an alias. The types of the standard library
are closed, so to add an operation on strings, write an ordinary function and call it with a pipe:

```ant
func shout(s: string): string {
  "${s.toUpper()}!"
}

let greeting = " hello " | String.trim() | shout()
println(greeting)
```

```text
HELLO!
```

## Calls through `.`, pipes, and namespaces

Every call of a namespace function can be written in namespace form, such as
`Array.filter(words, f)`. Two shorter forms read better in most code, and each fits a different job.

When `x` has no field `f`, `x.f(a)` calls `f` from the namespace of `x`'s own type, with `x` first;
a field `f` holding a function is called as it is. On an array it looks in `Array`, on a string in
`String`, on a `Point` in `Point`. Use it for local operations on a value: `name.trim()`,
`xs.isEmpty()`, `p.norm()`.

`a | f(b)` is `f(a, b)` for any function `f`. Use it for a chain of transformations, and wherever
the next function lives in another namespace or is your own top-level function:

```ant
let words = ["moth", "Ant", "bee", "Ant"]
let line = words
  | Array.map({ x.toLower() })
  | Array.unique()
  | Array.sort()
  | String.join(", ")
println(line)
```

```text
ant, bee, moth
```

`String.join` takes an array but lives in `String`, so `.` on the array cannot reach it, and the
pipe can. Keep the namespace form where it explains the operation or treats its inputs as equals,
such as `Point.distance(start, end)`. [Idiomatic Ant](idiomatic.md) has more on choosing between the
three.

## Mistakes to avoid

A lambda with no expected type does not compile. Annotate the binding, or write `func` with typed
parameters:

```ant
let double = { x * 2 }
// error: a lambda needs an expected function type; annotate the variable or use func with typed parameters
```

Named arguments work on a declared function, not on a function value. A value may be a lambda whose
only parameter is `x`, so its calls are positional:

```ant
func area(width: int, height: int): int {
  width * height
}

let f = area
area(height: 2, width: 3)
f(height: 2, width: 3)
// error: a call through a function value is positional; named arguments apply only to a declared function
```

A parameter cannot be assigned, a local function cannot use `func name`, and a recursive function
must state its return type:

```ant
func bump(n: int): int {
  n += 1 // error: cannot assign to n, a parameter; declare a var to hold a value that changes
  n
}
```

```ant
func outer(): int {
  func inner(): int { 1 }
  // error: a named function is only allowed at the top level; bind it with let or var
  inner()
}
```

```ant
func fact(n: int) {
  if n <= 1 { 1 } else { n * fact(n - 1) } // error: fact is recursive; state its return type
}
```

`p.norm` without a call is a field read, and `Point` has no field `norm`. To get the function as a
value, name it through the type, `Point.norm`, or wrap the call in a lambda:

```ant
type Point = { x: float64, y: float64 }

func Point.norm(p: Point): float64 {
  Math.sqrt(p.x * p.x + p.y * p.y)
}

let p = Point{ x: 3.0, y: 4.0 }
let norm = p.norm // error: Point has no field norm
```

A type of the standard library takes no new functions:

```ant
func String.shout(s: string): string { s.toUpper() }
// error: func String.shout: String is not a type declared at the top level of this module
```

## Learn more

The reference covers each part in full: [func](../language/keywords/func.md),
[lambdas](../language/literals/lambdas.md), [function types](../language/types/functions.md),
[calls through `.`](../language/operators/dot-calls.md), and the
[pipe](../language/operators/pipe.md). [Generics](generics.md) explains type parameters on
functions, and [Interfaces](interfaces.md) how several types share one set of functions.
