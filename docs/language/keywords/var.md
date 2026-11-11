<!-- status: reviewed -->

# `var`

`var` declares a mutable binding: a name that can be assigned a new value later.

## Details

- A `var` needs an initializer. The type is inferred from it, or written after the name, and every
  later assignment must fit that type.
- A `var` is visible from its declaration to the end of its block. Its own initializer can refer to
  it only inside a nested function body.
- A closure captures a `var` by reference: it sees every later assignment, and an assignment it
  makes is seen outside.
- A narrowing of a `var`, such as after `x != null`, lasts until the next assignment. A `var` that a
  lambda, a `go` block, a `defer` block, or another function assigns does not narrow. A library's
  top-level `var`, and an imported one, never narrows; read it into a `let` first.
- An importer reads an `export var` live and cannot assign it.
- A `var` that is never reassigned is valid, but [let](let.md) says so to the reader.

## Basic examples

```ant
var count = 0
count += 1
count = count * 10
count // 10
```

Write the type after the name when the initializer does not give the one you want:

```ant
var scores: Array[int] = []
scores = [90, 75]
scores.length() // 2
```

A `var` that a loop updates:

```ant
var total = 0
for n in [4, 8, 15] {
  total += n
}
println(total)
```

```text
27
```

## Scope

### Wider types

A `var` keeps the type it is declared with, so declare a union or a nullable type when the name
holds more than one kind of value:

```ant
var state: string | int = "idle"
state = 3
state // 3
var found: string | null = "yes"
found = null
found // null
```

### Destructuring

A destructuring declares a `var` for each name it binds:

```ant
type Size = { width: int, height: int }
var { width, height } = Size{ width: 2, height: 3 }
width *= 10
width * height // 60
var [low, high] = [9, 1]
[low, high] = [high, low]
low  // 1
high // 9
```

### Blocks and shadowing

A `var` is visible to the end of its block. An inner scope, such as a function body, may declare the
same name, which hides the outer one there:

```ant
var label = "outer"
func describe(): string {
  var label = "inner"
  label += "!"
  label
}
describe() // "inner!"
label      // "outer"
```

### Closures

A closure sees later assignments to a `var` it captures, and its own assignments change the `var`:

```ant
var level = "info"
let show = func(): string { level }
level = "debug"
show() // "debug"
let reset = func() { level = "info" }
reset()
level // "info"
```

A function can keep state in a captured `var` that lives on after the function returns:

```ant
func counter(): func(): int {
  var n = 0
  func(): int {
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
3
```

### `select` heads

A `select` head can bind its result with `var`, and the block can then reassign it:

```ant
let numbers: Channel[int] = Channel.new(1)
numbers.send(3)
let next = select {
  var n = Channel.receive(numbers) {
    n = (n ?? 0) + 1
    n
  }
}
println(next)
```

```text
4
```

### Exported

A library exports a `var` with `export var`. The importer always reads its current value:

```ant [counter.ant]
export var hits = 0

export func hit(): void {
  hits += 1
}
```

```ant
import "./counter.ant" { hits, hit }

hit()
hit()
println(hits)
```

```text
2
```

The importer cannot narrow an imported `var`, even one that nothing assigns, because any call may
reassign it:

```ant [config.ant]
export var name: string | null = "ant"
```

```ant
import "./config.ant" { name }

if name != null {
  println(name.length()) // error: string | null has no field or function length on every member: narrow the value first
}
```

## Applications

Keep the best value seen so far:

```ant
var longest = ""
for word in ["ant", "beetle", "moth"] {
  if word.length() > longest.length() {
    longest = word
  }
}
println(longest)
```

```text
beetle
```

Walk a linked list with a cursor. The loop condition narrows `node` on every pass, after each
assignment:

```ant
type Node = { value: int, next: Node | null }

let list = Node{ value: 1, next: Node{ value: 2, next: Node{ value: 3, next: null } } }
var node: Node | null = list
var sum = 0
while node != null {
  sum += node.value
  node = node.next
}
println(sum)
```

```text
6
```

Count the steps of the Collatz sequence from 27:

```ant
var n = 27
var steps = 0
while n != 1 {
  n = if n % 2 == 0 { n / 2 } else { 3 * n + 1 }
  steps += 1
}
println(steps)
```

```text
111
```

## Properties and relations

A `var` and a `let` with the same initializer hold the same value. Only the `var` can be assigned
later:

```ant
var a = [1, 2]
let b = [1, 2]
assert(a == b)
a = [3]
assert(a != b)
```

A compound assignment is the same as assigning the result of the operator:

```ant
var x = 7
var y = 7
x += 5
y = y + 5
assert(x == y)
```

## Possible issues

### The type comes from the initializer

A `var` keeps the type of its first value. Write a wider type when it holds another later:

```ant
var total = 0
total = 1.5 // error: cannot use float64 where int is expected
var status = "idle"
status = null // error: cannot use null where string is expected
```

### A `var` assigned in a function does not narrow

The checker cannot tell when a function or a lambda that assigns the `var` runs, so a test does not
narrow it. Read it into a `let` first:

```ant
var current: string | null = "draft"
func clear(): void {
  current = null
}
if current != null {
  println(current.toUpper()) // error: string | null has no field or function toUpper on every member: narrow the value first
}
```

```ant
var current: string | null = "draft"
func clear(): void {
  current = null
}
let value = current
if value != null {
  println(value.toUpper())
}
```

```text
DRAFT
```

### A narrowed `var` inside a lambda

Inside a lambda, a `func`, a `go` block, or a `defer` block, an outer `var` never narrows: a call or
another goroutine may assign it between a test and a use. A narrowing outside the lambda does not
reach into it, and a test inside does not narrow it either:

```ant
var user: string | null = "ant"
if user != null {
  println([1].map({ user.length() })) // error: user is a var of an enclosing body, which does not narrow here
}
```

Bind the narrowed value to a `let` inside the `if` and use that in the lambda:

```ant
var user: string | null = "ant"
if user != null {
  let name = user
  println([1].map({ name.length() }))
}
user = null
```

```text
[3]
```

### Reading a `var` in its own initializer

Outside a nested function body, an initializer cannot read the `var` it declares:

```ant
func f(): int {
  var n = n + 1 // error: n is read in its own initializer
  n
}
```

### Redeclaring

Declaring a name twice in one scope is an error. Assign the existing `var` instead:

```ant
var count = 0
var count = 1 // error: count is already declared in this scope
```

### Assigning an imported `var`

Only the module that declares a `var` can assign it. Call one of its functions instead:

```ant [store.ant]
export var limit = 10
```

```ant
import "./store.ant" { limit }

limit = 20 // error: cannot assign to the imported variable limit
```

## See also

[let](let.md) · [for](for.md) · [select](select.md) · [export](export.md) · [func](func.md) ·
[Assignment](../operators/assignment.md) · [Variables tour](../../tour/02-variables.md) ·
[Idiomatic Ant](../../guide/idiomatic.md)
