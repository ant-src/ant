<!-- status: reviewed -->

# Ant for JavaScript and TypeScript developers

Ant reads like TypeScript: braces, `let`, `${...}` interpolation, `?.` and `??`, regular expression
literals, and types after a colon. It behaves differently underneath. A failure is a value, not an
exception. Any function can wait, so there is no `async`. Nothing is truthy, and nothing converts on
its own. `null` goes only where a type allows it. Data is a record, not an object of a class. And
concurrent code runs on goroutines, not on an event loop.

This page maps the syntax you know, then covers the differences that bite. Each section links the
page that goes deeper.

## Equivalents

| JavaScript or TypeScript                       | Ant                                                          |
| ---------------------------------------------- | ------------------------------------------------------------ |
| `const limit = 10`                             | `let limit = 10`                                             |
| `let count = 0`                                | `var count = 0`                                              |
| `count++`                                      | `count += 1`                                                 |
| `a === b`                                      | `a == b`, which compares structure                           |
| `ok ? a : b`                                   | `if ok { a } else { b }`                                     |
| `` `total: ${n}` ``                            | `"total: ${n}"`                                              |
| `` String.raw`C:\temp` ``                      | `` `C:\temp` ``                                              |
| `undefined`                                    | `null`                                                       |
| `function add(a: number, b: number): number`   | `func add(a: int, b: int): int`                              |
| `(x) => x * 2`                                 | `{ x * 2 }`                                                  |
| `(a, b) => a + b`                              | `func(a: int, b: int): int { a + b }`                        |
| `connect("db", { port: 80 })`                  | `connect("db", port: 80)`                                    |
| `type Point = { x: number, y: number }`        | `type Point = { x: float64, y: float64 }`                    |
| `{ x: 1, y: 2 }`                               | `Point{ x: 1.0, y: 2.0 }`                                    |
| `{ ...p, x: 0 }`                               | `Point{ ...p, x: 0.0 }`                                      |
| `class Point { norm() { ... } }`               | `func Point.norm(p: Point): float64 { ... }`                 |
| `typeof x === "string"`, `x instanceof Circle` | `x is string`, `x is Circle`                                 |
| `switch (shape.kind)`                          | `if shape is Circle { ... } else { ... }`                    |
| `new Map([["a", 1]])`                          | `Map{ a: 1 }`                                                |
| `const ages = { ann: 31 }`, `ages[name]`       | `let ages = Map{ ann: 31 }`, `ages[name]`, an `int \| null`  |
| `Object.keys(o)`                               | `m.keys()`                                                   |
| `new Set(xs)`                                  | `Set.of(xs)`                                                 |
| `xs.length`, `xs.includes(v)`                  | `xs.length()`, `xs.contains(v)`                              |
| `xs.join(", ")`                                | `xs \| String.join(", ")`                                    |
| ``xs.map((x, i) => `${i}: ${x}`)``             | `xs.entries().map({ "${x.index}: ${x.value}" })`             |
| `for (const x of xs)`                          | `for x in xs`                                                |
| `for (let i = 0; i < n; i++)`                  | `for i in Iter.range(0, n)`                                  |
| `for (const [k, v] of m)`                      | `for { key, value } in m`                                    |
| `const { x, y } = p`                           | `let { x, y } = p`                                           |
| `parseInt(s, 10)`                              | `s.toInt()`, an `int \| null`                                |
| `String(n)`                                    | `String.from(n)`, or `"${n}"`                                |
| `flags \| WRITE`                               | `Bits.or(flags, WRITE)`                                      |
| `console.log(a, b)`                            | `println("${a} ${b}")`, one value per call                   |
| `JSON.parse(text)`, `JSON.stringify(v)`        | `Json.parse(text)`, a `Result`; `Json.stringify(v)`          |
| `throw new Error("bad")`                       | `return Result.fail("bad")`                                  |
| `try { ... } catch (e) { ... }`                | `try` on one call, or an `is` test on the `Result`           |
| `async function`, `await`                      | an ordinary function; a call that waits blocks its goroutine |
| `await Promise.all(xs.map(f))`                 | `Parallel.map(xs, f, limit: n)`, an `Array[U] \| null`       |
| `import { a } from "./m.js"`                   | `import "./m.ant" { a }`                                     |
| `export function`                              | `export func`                                                |

## Failures are values

Ant has no `throw` and no `catch` for expected failure. A function that can fail returns a `Result`,
and says so in its return type:

<!-- test: file port.txt
8080
-->
<!-- test: file bad.txt
80x
-->

```ant
func readPort(path: string): Result[int, Error] {
  let text = try File.read(path)
  let port = text.trim().toInt()
  if port == null { return Result.fail("${path}: not a number") }
  Result.ok(port)
}

for path in ["port.txt", "bad.txt", "missing.txt"] {
  let port = readPort(path)
  println(if port is Ok { "port ${port.value}" } else { "skip: ${port.error.message}" })
}
```

```text
port 8080
skip: bad.txt: not a number
skip: open missing.txt: no such file or directory
```

`try` is not a block. It goes in front of one call and returns its error to the caller, as `await`
goes in front of one promise. The return type says the call can fail. The caller tests the `Result`
with `is` or passes its error on with `try`. `panic` is for bugs, and `Panic.catch` belongs at the
edge of a job, not around each call. See [Errors](../guide/errors.md#coming-from-other-languages).

## No async

Any function can wait. A call that reads a file, sleeps, or receives from a channel blocks only its
own goroutine, so nothing marks a function as `async` and nothing needs an `await`:

```ant
func slowSquare(n: int, ctx: Context | null = null): int {
  Time.sleep(10ms, ctx: ctx) // stands in for a network call
  n * n
}

println([1, 2, 3].map(slowSquare))
println(Parallel.map([1, 2, 3], slowSquare, limit: 3))
```

```text
[1, 4, 9]
[1, 4, 9]
```

`map` takes a function that waits as it takes any other, and returns values, not promises. The
caller decides what runs at once: `Parallel.map` runs the calls on goroutines, as `Promise.all`
would, and returns the results in input order. It passes each call its own `Context`, which
`slowSquare` passes on to `Time.sleep`, so the sleep stops when `Parallel.map` cancels the rest, as
it does when one call panics. It returns `null` when its context ends before every call has started,
so a caller unwraps it, as in `Parallel.map(xs, f, limit: 8) ?? return Result.fail("cancelled")`.
`async` and `await` are reserved words, so a habit stops at `ant check`. See
[Why goroutines and not async](../design/no-async.md).

## No truthiness, no coercion

A condition must be a `bool`, and `+` takes two strings or two numbers of one kind:

```ant
let count = 0
if count { println("some") }  // error: cannot use int where bool is expected
let label = "items: " + count // error: + needs both operands of one type, got string and int
```

Ask the question you mean, and convert where you mean it:

```ant
let count = 0
count > 0         // false
"items: ${count}" // "items: 0"
"4x".toInt()      // null
7 / 2             // 3
float64(7) / 2.0  // 3.5
let saved: int | null = 0
saved ?? 8080 // 0
```

`??` replaces only `null`, so a saved `0` stays `0`, where `saved || 8080` would give `8080` in
JavaScript. `int` and `float64` never mix, and integer division truncates. See
[Mistakes to avoid](../guide/types.md#mistakes-to-avoid) and
[No implicit coercion, no truthiness](../design/no-coercion.md).

## Null only where the type says

There is no `undefined`, and `null` fits only a type that includes it. A `string` is never `null`; a
`string | null` may be, and a test narrows it:

```ant
func greet(name: string | null): string {
  if name == null { return "hello, stranger" }
  "hello, ${name.toUpper()}"
}

println(greet("ann"))
println(greet(null))
```

```text
hello, ANN
hello, stranger
```

Where JavaScript returns `undefined`, Ant either says so in the type or panics:

```ant
let names = ["Ann", "Bob"]
let ages = Map{ ann: 31 }
names.get(5) // null
ages["bob"]  // null
names[5]     // panic: index 5 out of range for an array of length 2
```

`get` and a map read return `null` for a missing element. Indexing an array past its end is a bug,
and panics. See [Null](../guide/null.md#coming-from-other-languages).

## Records, not classes

Ant has no `class`, no `this`, and no constructor method. A record literal builds the value, and a
function on the type, such as `Point.origin()`, builds one that needs logic. A record holds the
data, and functions on its type hold the behavior:

```ant
type Point = { x: float64, y: float64 }

func Point.norm(p: Point): float64 { Math.sqrt(p.x * p.x + p.y * p.y) }

let p = Point{ x: 3.0, y: 4.0 }
let q = p
q.x = 0.0
println(p.norm())
println(p == Point{ x: 0.0, y: 4.0 })
```

```text
4.0
true
```

The receiver is the first parameter, so there is no `this` to bind or to lose, and `p.norm()` is
`Point.norm(p)`. Every record literal names its type: there is no bare `{ x: 1 }`. Assigning a
record shares it, as assigning an object does in JavaScript, so `q.x = 0.0` changes `p` too. But
`==` compares structure, not identity.

A named record type is nominal, unlike in TypeScript. A `Vec` with the same fields is not a `Point`.
To take any record with a `name`, write the parameter as an anonymous record type:
`func greet(n: { name: string })`. A record has no dynamic access: when the keys are data, use a
`Map`. See [Records instead of classes](../design/records-not-classes.md) and
[Records](../guide/records.md).

## Goroutines, not an event loop

`go` starts a goroutine, and a channel carries its result back:

```ant
let done: Channel[string] = Channel.new()
go {
  Time.sleep(10ms)
  done.send("report sent")
}
println(done.receive() ?? "closed")
```

```text
report sent
```

Goroutines run concurrently and are preemptive, unlike callbacks on an event loop. A long
computation does not block the others, but two goroutines that share a variable, when one of them
writes it, need a `Mutex`, an `Atomic`, or a channel. The program exits when the script ends, with
goroutines still running: wait on a channel or a `WaitGroup`, or use `Parallel`. See
[Concurrency pitfalls](../guide/concurrency.md#pitfalls).

## Other habits to drop

- **`ant run` does not check types.** It discards annotations, as a TypeScript runner that strips
  types does, and panics where the checker would have reported an error. Run `ant check` in your
  editor and in CI. See [Types](../guide/types.md).
- **A lambda has one parameter, `x`.** `{ x * 2 }` is `(x) => x * 2`; with two parameters, write
  `func`. `return` in a lambda leaves the lambda, as in a `forEach` callback. See
  [Functions](../guide/functions.md).
- **`sort` and `reverse` return a new array.** They leave their argument alone. See
  [Collections](../guide/collections.md).
- **Backticks make a raw string.** `${...}` interpolates between double quotes; between backticks it
  is literal text. A string's length counts code points, not UTF-16 units, and a string cannot be
  indexed. See [Strings](../guide/strings.md).
- **An import names a file.** `"./m.ant"` with its extension, never a package, and no cycles. See
  [Modules](../guide/modules.md#coming-from-other-languages).
- **Markup is not JSX.** Attribute names are HTML's, such as `class`, and `&&` is not a conditional.
  See [HTML](../guide/html.md#coming-from-jsx-and-templates).

## See also

- [Tour](../tour/01-hello.md), from the first program to goroutines
- [Idiomatic Ant](../guide/idiomatic.md)
- [Values](../guide/values.md#coming-from-other-languages), for `let`, `var`, and assignment
- [Coming from Go](go.md) and [Coming from Python](python.md)
- [Why everything is an expression](../design/expressions.md) and
  [Why no exceptions](../design/errors-are-values.md)
- [Keywords and reserved words](../language/lexical/reserved.md)
