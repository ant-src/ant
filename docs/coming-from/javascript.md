<!-- status: reviewed -->

# Ant for JavaScript and TypeScript developers

If you write TypeScript, Ant looks familiar from the first line: braces, `let`, `${...}` in strings,
`?.` and `??`, and types after a colon. A few ideas underneath work differently, and this page is
about those. It does not try to cover the whole language. The [Tour](../tour/01-hello.md) does that.

## A first program

Here is a small TypeScript program:

```ts
type User = { name: string; age: number };

const users: User[] = [
  { name: "Ann", age: 31 },
  { name: "Bob", age: 17 },
];

for (const user of users.filter((u) => u.age >= 18)) {
  console.log(`Hello, ${user.name}`);
}
```

The same program in Ant:

```ant
type User = { name: string, age: int }

let users = [User{ name: "Ann", age: 31 }, User{ name: "Bob", age: 17 }]

for user in users.filter({ x.age >= 18 }) {
  println("Hello, ${user.name}")
}
```

```text
Hello, Ann
```

A few things to notice:

- A record literal names its type: `User{ ... }`, never a bare `{ ... }`.
- `{ x.age >= 18 }` is a short function. Its one parameter is always named `x`.
- Double-quoted strings interpolate `${...}`, as template literals do.
- Numbers are `int` or `float64`, not `number`.
- There are no semicolons.

Save it as `main.ant` and run `ant run main.ant`. `ant run` does not check types, the way `node`
does not check a `.ts` file when it strips the types. Run `ant check main.ant` for that, or let your
editor do it with the [Ant extension](/vscode).

## What carries over

| JavaScript or TypeScript             | Ant                             |
| ------------------------------------ | ------------------------------- |
| `const limit = 10`                   | `let limit = 10`                |
| `let count = 0`                      | `var count = 0`                 |
| `count++`                            | `count += 1`                    |
| `a === b`                            | `a == b`                        |
| `ok ? a : b`                         | `if ok { a } else { b }`        |
| `(x) => x * 2`                       | `{ x * 2 }`                     |
| `function add(a: number, b: number)` | `func add(a: int, b: int): int` |
| `xs.length`, `xs.includes(v)`        | `xs.length()`, `xs.contains(v)` |
| `for (const x of xs)`                | `for x in xs`                   |
| `const { x, y } = point`             | `let { x, y } = point`          |
| `new Map([["a", 1]])`                | `Map{ a: 1 }`                   |
| `undefined`                          | `null`                          |
| `console.log(x)`                     | `println(x)`                    |
| `import { a } from "./m.js"`         | `import "./m.ant" { a }`        |
| `export function f()`                | `export func f()`               |

## Five things that work differently

### Errors are values

Ant has no `throw` and no `catch`. A function that can fail says so in its return type, and returns
a `Result`:

<!-- test: file port.txt
8080
-->

```ant
func readPort(path: string): Result[int, Error] {
  let text = try File.read(path)
  let port = text.trim().toInt()
  if port == null { return Result.fail("${path}: not a number") }
  Result.ok(port)
}

let port = readPort("port.txt")
if port is Ok {
  println("listening on ${port.value}")
} else {
  println("error: ${port.error.message}")
}
```

```text
listening on 8080
```

`try` goes in front of one call, the way `await` goes in front of one promise. When the call fails,
`try` returns the error from the current function. Otherwise it gives you the value. The caller
checks the result with `is`, as above, or passes it on with another `try`.

`panic` still exists, but for bugs, such as an index past the end of an array. See
[Errors](../guide/errors.md).

### No `async`, no `await`

Any function can wait. A call that sleeps, reads a file, or talks to the network pauses only its own
goroutine, so it needs no `async` and no `await`:

```ant
func fetchName(id: int): string {
  Time.sleep(10ms) // stands in for a network call
  "user ${id}"
}

println(fetchName(1))
```

```text
user 1
```

`fetchName` returns a `string`, not a promise. When you want several calls at once, the way
`Promise.all` runs them, use [`Parallel.map`](../lib/Parallel/map.md). See
[Why goroutines and not async](../design/no-async.md).

### Nothing is truthy

A condition must be a `bool`, and nothing converts on its own:

```ant
let count = 0
if count { println("some") }  // error: cannot use int where bool is expected
let label = "items: " + count // error: + needs both operands of one type, got string and int
```

Ask the question you mean:

```ant
let count = 0
count > 0         // false
"items: ${count}" // "items: 0"
"42".toInt()      // 42
"4x".toInt()      // null
```

### `null` goes only where the type allows it

A `string` is never `null`. A `string | null` may be, and after you check it, Ant knows it is a
`string`:

```ant
func greet(name: string | null): string {
  if name == null { return "Hello, stranger" }
  "Hello, ${name.toUpper()}"
}

println(greet("ann"))
println(greet(null))
```

```text
Hello, ANN
Hello, stranger
```

`??` replaces only `null`, so `0 ?? 8080` stays `0`, where `0 || 8080` in JavaScript gives `8080`.
See [Null](../guide/null.md).

### Records instead of classes

There is no `class` and no `this`. A record holds the data, and functions on its type hold the
behavior:

```ant
type Point = { x: float64, y: float64 }

func Point.norm(p: Point): float64 {
  Math.sqrt(p.x * p.x + p.y * p.y)
}

let p = Point{ x: 3.0, y: 4.0 }
println(p.norm())
```

```text
5.0
```

`p.norm()` calls `Point.norm(p)`: the receiver is the first parameter, so there is no `this` to
lose. See [Records instead of classes](../design/records-not-classes.md).

## Where to go next

- The [Tour](../tour/01-hello.md) walks through the language one small program at a time.
- [Idiomatic Ant](../guide/idiomatic.md) shows how Ant code usually looks.
- [Concurrency](../guide/concurrency.md) covers goroutines and channels, Ant's answer to the event
  loop.
- [Coming from Go](go.md) and [Coming from Python](python.md).
