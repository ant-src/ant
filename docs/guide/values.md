<!-- status: reviewed -->

# Values and variables

A name in Ant is bound once, by a declaration, and lives until the end of the block that declares
it. `let` binds a name you never reassign, and `var` one you do. The prelude binds the standard
library's names in every module, and `_` takes values you do not need. This chapter shows how to
pick between `let` and `var`, where a name is visible, what shadowing allows, and in which order the
top level of a file runs.

## `let` and `var`

Declare a name with `let` unless you reassign it. Declare it with `var` when you do:

```ant
let start = 27
var n = start
var steps = 0
while n != 1 {
  n = if n % 2 == 0 { n / 2 } else { 3 * n + 1 }
  steps += 1
}
println("${start} reaches 1 in ${steps} steps")
```

```text
27 reaches 1 in 111 steps
```

`start` never changes, so it is a `let`, and the keyword tells the reader so. Only `n` and `steps`
are reassigned. Assigning a `let` is a compile error, wherever the assignment is written:

```ant
let count = 0
count = 1 // error: cannot assign to count, a let; declare it with var to reassign it
```

Every declaration has an initializer, and the type checker infers the type from it. Write the type
after the name when the initializer does not give the one you want, such as an empty array or a
nullable value:

```ant
let names: Array[string] = []
var winner: string | null = null
names.push("Ann")
winner = names.first()
println(winner)
```

```text
Ann
```

### A `let` fixes the name, not the value

`names` above is a `let`, and `names.push("Ann")` still works. A `let` stops you from pointing the
name at another value. It does not stop the value from changing: the elements of an array or
`Bytes`, the entries of a map or set, and the fields of a record can still change.

```ant
type Counter = { hits: int }
let counter = Counter{ hits: 0 }
counter.hits += 1
counter.hits // 1
```

Use `var` for the name, not for the contents. A `var` holding an array that only grows is a `var`
nobody reassigns; make it a `let`.

### Names that are already `let`s

Parameters, a lambda's `x`, and every name a `for` binds are `let` bindings. To compute a new value
from a parameter, bind a new name:

```ant
func greet(name: string): string {
  name = name.trim()
  // error: cannot assign to name, a parameter; declare a var to hold a value that changes
  "hello, ${name}"
}
```

```ant
func greet(name: string): string {
  let trimmed = name.trim()
  if trimmed.isEmpty() { return "hello" }
  "hello, ${trimmed}"
}

println(greet("  Ann "))
println(greet("   "))
```

```text
hello, Ann
hello
```

### `let` narrows where `var` does not

For a name nobody reassigns, `let` and `var` behave the same, with one exception. Inside a lambda or
a nested function, an outer `let` keeps the type a test narrowed it to, since nothing can change it.
An outer `var` does not, since the body may run after another assignment. See
[Null](null.md#what-narrows) for the details. That is one more reason to default to `let`.

## Values are shared, not copied

A name holds a value. For numbers, strings, booleans, and `null`, which cannot change, that is the
whole story. Arrays, maps, sets, records, `Bytes`, and functions are references: binding one to
another name shares it, and a change through either name shows through both.

```ant
let a = [1, 2]
let b = a
b.push(3)
a // [1, 2, 3]
```

A function that receives an array receives the same array. Make a copy when you need one, such as
`[...a]`; [Collections](collections.md) covers copies, equality, and identity.

## Scope

A block opens a scope: a function body, a branch of `if`, the body of `for` or `while`, a lambda. A
name declared in a block is visible from its declaration to the end of that block, and not outside:

```ant
func describe(n: int): string {
  if n > 100 {
    let excess = n - 100
  }
  "over by ${excess}" // error: excess is not defined
}
```

Ant has no free-standing block: bare braces make a lambda. To use a value after an `if`, make the
`if` an expression and bind its value:

```ant
func describe(n: int): string {
  let excess = if n > 100 { n - 100 } else { 0 }
  "over by ${excess}"
}

println(describe(130))
```

```text
over by 30
```

Lookup goes outward, block by block, to the top level of the module and then to the prelude. A
function body sees the names of the blocks around it.

A declaration's initializer can use its own name only inside a function body it creates, which is
how a local function calls itself:

```ant
let fact = func(n: int): int { if n <= 1 { 1 } else { n * fact(n - 1) } }
fact(5) // 120
```

### Closures keep the variable

A lambda or a nested function keeps a reference to the scope it was made in, not a copy of its
values. It sees later assignments to a `var` it uses:

```ant
var greeting = "hello"
let greet = func(name: string): string { "${greeting}, ${name}" }
greeting = "goodbye"
println(greet("Ann"))
```

```text
goodbye, Ann
```

Each iteration of a `for` binds a fresh loop variable, so a closure made in the loop body keeps the
value of its own iteration:

```ant
let actions: Array[func(): int] = []
for n in [1, 2, 3] {
  actions.push({ n * 10 })
}
actions.map({ x() }) // [10, 20, 30]
```

## Shadowing

An inner scope may declare a name that an outer scope already has. The inner name hides the outer
one until the end of its block, and the outer one is untouched:

```ant
let limit = 10

func allows(n: int): bool {
  let limit = 20
  n < limit
}

println(allows(15))
println(limit)
```

```text
true
10
```

In one scope, a name is declared once. Declaring it again, even with `let` both times, is an error.
A function's parameters are in the scope of its body, so a local cannot take a parameter's name:

```ant
let input = "42"
let input = input.toInt() // error: input is already declared in this scope
```

```ant
func double(n: int): int {
  let n = n * 2 // error: n is already declared in this scope
  n
}
```

Pick a name that says what changed, such as `text` and `number`, or `n` and `doubled`.

Types and values share one set of names. A `type` and a `let` of the same name in one scope collide,
and a parameter named like a type hides the type inside the function:

```ant
type User = { name: string }
let User = "Ann" // error: User is already declared in this scope
```

```ant
type User = { name: string }

func rename(User: string): string {
  let u: User = User // error: User is a variable, not a type; it hides the outer User here
  u
}
```

## The prelude

`println`, `panic`, `assert`, `len`, the conversions such as `int` and `string`, the types such as
`Result` and `Error`, and the namespaces such as `Array`, `String`, and `Time` are not keywords. The
prelude binds them in every module, as if each file imported them, so you use them without an
`import`.

Unlike the other names of a module, a prelude name cannot be declared again in any scope, so
`println` and `int` mean the same thing in every line of every program:

```ant
func summary(lines: Array[string]): string {
  let len = lines.length() // error: len is a prelude name and cannot be redeclared
  "${len} lines"
}
```

Choose another name, such as `count`. A parameter cannot take a prelude name either. A field can: a
record may have a field named `len`.

## `_` discards a value

`_` is a name that takes any value and keeps none. Bind it where a declaration, a destructuring, a
loop, or a parameter list wants a name and you do not need the value:

```ant
let [_, second] = ["Ann", "Bob"]
second // "Bob"
```

```ant
for _ in Iter.range(0, 3) {
  print(".")
}
println()
```

```text
...
```

You can bind or assign `_` any number of times, in one scope too. You cannot read it:

```ant
let _ = "first"
let _ = "second"
println(_) // error: _ is write-only and cannot be read
```

In the REPL, reading `_` gives the result the last input printed.

## Top-level order

The top level of a file holds declarations and statements in any order, but not all of them work in
any order:

- `func` and `type` declarations are visible throughout the module, so functions call each other and
  use types declared further down.
- `let` and `var` initializers and statements run from top to bottom. A top-level initializer or
  statement may use only the variables declared above it.

```ant
let doubled = base * 2
// error: base is declared later in the module; a top-level initializer or statement may refer only to the variables above it
let base = 10
```

Inside a function body, a later variable is fine, since the body runs only when the function is
called. The type checker does not follow calls, so a call made before the variable is initialized
passes `ant check` and panics in `ant run`:

```ant
let early = total() // panic: base is read before its initializer has run
println(early)

func total(): int {
  base + 1
}

let base = 10
```

Keep the top-level variables at the top, before the statements that use them, and the functions
anywhere.

## Coming from other languages

- **No implicit declaration.** In Python, `total = 0` creates `total`. In Ant, assigning a name that
  no `let` or `var` declares is an error: `total is not defined`.
- **`let` is JavaScript's `const`, and `var` is its `let`.** Both are block scoped. Nothing is
  function scoped, as JavaScript's `var` is.
- **No same-scope shadowing.** Rust's `let s = s.trim()` in one block is a redeclaration in Ant.
  Choose a new name.
- **Assignment is not an expression.** `a = b = 0` and `if n = 5` are errors. There is no `++` or
  `--`; write `n += 1`.
- **Case is only spelling.** `Value` and `value` are two names, and capitals do not export a name,
  as they do in Go; `export` does.

```ant
var a = 0
var b = 0
a = b = 0 // error: assignment cannot be chained
```

```ant
var n = 0
if n = 5 { println("five") } // error: cannot use void where bool is expected
```

## See also

- [`let`](../language/keywords/let.md) and [`var`](../language/keywords/var.md)
- [Identifiers](../language/lexical/identifiers.md) and
  [keywords and reserved words](../language/lexical/reserved.md)
- [Assignment](../language/operators/assignment.md)
- [Prelude functions](../lib/prelude/index.md)
- [Null](null.md), for narrowing
- [Collections](collections.md), for sharing and copying
- [Modules](modules.md), for imports and exports
- [Idiomatic Ant](idiomatic.md)
