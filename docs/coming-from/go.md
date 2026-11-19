<!-- status: reviewed -->

# Ant for Go developers

Ant takes a lot from Go: goroutines, channels, `select`, `defer`, raw strings, duration literals,
interfaces that a type satisfies without saying so, and a standard library that behaves like Go's.
The syntax reads like JavaScript, and a few ideas are Ant's own: every construct is an expression,
unions take the place of most interfaces and of `nil`, and a failure is a `Result`, not a second
return value. This page maps Go to Ant, then covers the differences that change how you write code.

## Equivalents

### Declarations and types

| Go                                              | Ant                                                                                     |
| ----------------------------------------------- | --------------------------------------------------------------------------------------- |
| `x := 1`, `var x = 1`                           | `var x = 1`, or `let x = 1` when you never reassign `x`                                 |
| `const limit = 100`                             | `let limit = 100`                                                                       |
| `type Point struct { X, Y float64 }`            | `type Point = { x: float64, y: float64 }`                                               |
| `Point{X: 1, Y: 2}`                             | `Point{ x: 1.0, y: 2.0 }`                                                               |
| `func divmod(a, b int) (int, int)`              | `func divmod(a: int, b: int): DivMod` returning a record, `let { q, r } = divmod(7, 2)` |
| `func sum(xs ...int)`, `sum(1, 2)`              | `func sum(xs: Array[int])`, `sum([1, 2])`                                               |
| `&Point{...}`, `*Point`                         | `Point{ ... }`: a record is always shared                                               |
| `q := p` (a copy)                               | `let q = Point{ ...p }`                                                                 |
| `type Celsius float64`                          | `type Celsius = float64`, built as `Celsius(20.5)`                                      |
| `func (p Point) Norm() float64`                 | `func Point.norm(p: Point): float64`                                                    |
| `type Shape interface{ Area() float64 }`        | `interface Shape { area(s: Self): float64 }`                                            |
| `switch v := s.(type)`                          | `if s is Circle { ... } else if s is Rect { ... }`                                      |
| `const ( Red = iota; Green )`                   | `type Red = {}`, `type Green = {}`, `type Color = Red \| Green`                         |
| `any`, `interface{}`                            | a type parameter, `[T]`, or a union                                                     |
| `func Apply[T, U any](xs []T, f func(T) U) []U` | `func apply[T, U](xs: Array[T], f: func(T): U): Array[U]`                               |
| `var u *User = nil`                             | `var u: User \| null = null`                                                            |
| `[]int{1, 2, 3}`                                | `[1, 2, 3]`                                                                             |
| `xs = append(xs, 4)`                            | `xs.push(4)`                                                                            |
| `xs[1:3]`                                       | `xs.slice(1, 3)`                                                                        |
| `map[string]int{"a": 1}`                        | `Map{ a: 1 }`                                                                           |
| `v, ok := m[k]`                                 | `m[k]`, a `V \| null`; `m.has(k)` for a stored `null`                                   |
| `delete(m, k)`                                  | `m.remove(k)`                                                                           |
| `counts[w]++`                                   | `counts[w] = (counts[w] ?? 0) + 1`                                                      |
| `s[i]`, `s[a:b]`                                | `s.chars()[i]`, `s.slice(a, b)`, counting code points                                   |
| `fmt.Sprintf("%d items", n)`                    | `"${n} items"`                                                                          |
| `` `C:\temp` ``                                 | `` `C:\temp` ``                                                                         |
| `2 * time.Second`                               | `2s`                                                                                    |
| `time.Duration(n) * time.Second`                | `Time.duration(seconds: n)`                                                             |
| `a & b`, `a << n`                               | `Bits.and(a, b)`, `Bits.shiftLeft(a, n)`                                                |

### Control flow

| Go                                      | Ant                                    |
| --------------------------------------- | -------------------------------------- |
| `for _, x := range xs`                  | `for x in xs`                          |
| `for i, x := range xs`                  | `for { index, value } in xs.entries()` |
| `for k, v := range m`                   | `for { key, value } in m`              |
| `for i := 0; i < n; i++`                | `for i in Iter.range(0, n)`            |
| `for cond { ... }`                      | `while cond { ... }`                   |
| `switch x { case 1: ... default: ... }` | `if x == 1 { ... } else { ... }`       |
| `i++`                                   | `i += 1`                               |
| `defer f.Close()`                       | `defer { f.close() }`                  |

### Errors

| Go                                      | Ant                                          |
| --------------------------------------- | -------------------------------------------- |
| `func f() (int, error)`                 | `func f(): Result[int, Error]`               |
| `return n, nil`                         | `Result.ok(n)`                               |
| `return 0, errors.New("empty")`         | `return Result.fail("empty")`                |
| `if err != nil { return 0, err }`       | `try` in front of the call                   |
| `fmt.Errorf("parse %s: %w", path, err)` | `Result.wrapErr(r, "parse ${path}")`         |
| `errors.Is`, `errors.As`                | an error union of your own, tested with `is` |
| `panic("bug")`                          | `panic("bug")`                               |
| `recover()` in a deferred function      | `Panic.catch(f)`                             |

### Concurrency

| Go                                        | Ant                                                                                  |
| ----------------------------------------- | ------------------------------------------------------------------------------------ |
| `go f(x)`                                 | `go { f(x) }`                                                                        |
| `ch := make(chan int, 10)`                | `let ch: Channel[int] = Channel.new(capacity: 10)`                                   |
| `ch <- v`                                 | `ch.send(v)`                                                                         |
| `v, ok := <-ch`                           | `let v = ch.receive()`, `null` once `ch` is closed and drained                       |
| `for v := range ch`                       | `for v in ch.values()`                                                               |
| `close(ch)`                               | `ch.close()`                                                                         |
| `select { case v := <-ch: ... }`          | `select { let v = Channel.receive(ch) { ... } }`                                     |
| `var wg sync.WaitGroup`                   | `let wg = WaitGroup.new()`                                                           |
| `mu.Lock(); defer mu.Unlock()`            | `mu.withLock({ ... })`                                                               |
| `context.WithTimeout(ctx, 2*time.Second)` | `Context.new(2s, parent: ctx)`                                                       |
| `errgroup.SetLimit`, a pool of workers    | `Parallel.map(xs, f, limit: 8)`, every result in order; it does not stop at an `Err` |
| `runtime.GOMAXPROCS(0)`                   | `Os.cpus()`, the CPUs the program may use; the limit for CPU-bound `Parallel.map`    |

### Programs and tools

| Go                                        | Ant                                                                     |
| ----------------------------------------- | ----------------------------------------------------------------------- |
| `package main` and `func main()`          | statements at the top level of a script                                 |
| `os.Args[1:]`                             | `Os.args()`                                                             |
| `import "strings"`                        | nothing: `String`, `File`, and the other namespaces are in every module |
| `import "example.com/geo"`                | `import "./geo.ant" { Point }`                                          |
| a capitalized name exports it             | `export func distance(...)`                                             |
| `func TestSum(t *testing.T)`              | `test "sum" { assert(...) }` in `sum_test.ant`                          |
| `fmt.Println(x)`                          | `println(x)`                                                            |
| `html/template`                           | `<p>Hello, {name}</p>`                                                  |
| `go run main.go`                          | `ant run main.ant`                                                      |
| `go vet`, the type checking of `go build` | `ant check main.ant`                                                    |
| `go test`, `gofmt`, `go doc`              | `ant test`, `ant fmt`, `ant doc`                                        |

## Differences

### Everything is an expression

A block's value is its last expression, so a function needs `return` only to leave early, and `if`
chooses between values where Go needs a variable and two assignments:

```ant
func describe(n: int): string {
  let sign = if n < 0 { "negative" } else if n == 0 { "zero" } else { "positive" }
  "${n} is ${sign}"
}

println(describe(-3))
println(describe(0))
```

```text
-3 is negative
0 is zero
```

A `select` is an expression too, so its clauses can produce the value a variable takes. See
[Expressions](../tour/08-expressions.md) and
[why everything is an expression](../design/expressions.md).

### Unions instead of type switches and nil

A closed set of types is a union, and `is` narrows it, where Go uses an interface and a type switch.
A chain of `is` tests that covers every member needs no `else`, and the type checker rejects it once
a new member goes unhandled:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle {
    Math.PI * s.radius * s.radius
  } else if s is Rect {
    s.width * s.height
  }
}

let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
println(shapes.map(area))
```

```text
[3.141592653589793, 6.0]
```

Absence works the same way. There is no `nil` and no zero value: a `User` is always a `User`, and a
value that may be missing has the type `User | null`. After `if user == null { return }`, the type
checker knows `user` is a `User`. A map read is a nullable value, so it replaces the comma-ok form:

```ant
let ports = Map{ http: 80, https: 443 }
ports["https"]     // 443
ports["ssh"]       // null
ports["ssh"] ?? 22 // 22
```

Interfaces remain for open sets, satisfied without a declaration as in Go. See
[Unions](../guide/unions.md), [Null](../guide/null.md#coming-from-other-languages), and
[Interfaces](../guide/interfaces.md).

### Errors are results

A function that can fail returns a `Result`, and `try` is `if err != nil { return err }` in one
word:

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

let port = try (readPort("port.txt") | Result.wrapErr("load config"))
println("listening on ${port}")
```

```text
listening on 8080
```

A wrapped error keeps only its message, so there is no `errors.Is`: an error that code branches on
is a type of its own. [Errors](../guide/errors.md#coming-from-other-languages) lists the rest.

### Scripts instead of main

A file with statements at its top level is a script, and those statements are its body: no
`package main`, no `func main`. The script body is a function of its own, so `try` and `return` work
there. When it returns an `Err`, `ant run` writes the message to standard error and exits with
status 1, where a Go `main` would call `log.Fatal`:

<!-- test: args ant go -->
<!-- test: exit 1 -->

```ant
for arg in Os.args() {
  println(arg.toUpper())
}
let text = try File.read("missing.txt")
println(text)
```

```text
ANT
GO
```

The two arguments print, and then the script ends at `try` with the error from `File.read`.

A file with no top-level statement is a library, which a script imports by path. See
[Modules](../guide/modules.md#coming-from-other-languages).

### Types are checked, not run

`ant check` is the type checker, and `ant run` does not use it: the interpreter discards the
annotations and runs the program. A program `ant check` rejects still runs until it reaches the
wrong value, and panics there. Go's compiler refuses to build a program with a type error; in Ant,
run `ant check` in your editor and in CI to get the same guarantee.

```ant
func double(n: int): int { n * 2 }

println(double("21")) // error: cannot use string where int is expected
```

Generics are erased too: `x is Box` tests the tag, never `Box[int]`. See
[Types at run time](../guide/runtime-types.md).

### Goroutines have no handle

`go` takes a block, not a call, and returns nothing to wait on. The program exits when the script
ends, as a Go program exits when `main` returns. For work of a known shape, `Parallel` starts the
goroutines, waits for them, and keeps the order of the results:

```ant
let words = ["ant", "go", "rust"]
let lengths = Parallel.map(words, func(word: string, _: Context): int { word.length() }, limit: Os.cpus())
println(lengths)
```

```text
[3, 2, 4]
```

`go { f(x) }` reads `x` when the goroutine runs, not when `go` runs, so bind a `let` first when `x`
is a `var` that changes.

A goroutine's panic ends that goroutine and is reported on standard error; the program goes on. See
[Concurrency](../guide/concurrency.md#pitfalls).

### Markup is an expression

Where Go renders `html/template` files, Ant writes HTML in code. The markup is an `Html` value, and
a string inside it becomes escaped text:

```ant
let name = "<Gopher>"
println(<p class="greeting">Hello, {name}</p>)
```

```text
<p class="greeting">Hello, &lt;Gopher&gt;</p>
```

See [HTML](../guide/html.md).

### A library that behaves like Go's

`Path`, `File`, `Os`, `Regex`, and the rest give Go's answers to the hard cases, under lowercase
names. Every function takes its subject first, so you call it through `.` or a pipe, and strings
count code points, not bytes:

```ant
Path.clean("a/b/../c")                        // "a/c"
"a,b,c".split(",")                            // ["a", "b", "c"]
"héllo".length()                              // 5
"héllo".toBytes().length()                    // 6
["b", "a"] | Array.sort() | String.join(", ") // "a, b"
```

`Time` is the exception: it follows JavaScript's Temporal, not Go's `time`. See
[`Time`](../lib/Time/index.md).

### Smaller surprises

- **Records are shared.** Assigning or passing a record shares it, as a pointer would in Go. Copy
  with a spread, `Point{ ...p }`. See [Records](../guide/records.md).
- **Fields have no zero value.** A literal gives every field, unless the type declares a default:
  `type Server = { addr: string, workers: int = 4 }` lets `Server{ addr: ":8080" }` leave `workers`
  out. See [Records](../guide/records.md).
- **Integer overflow panics.** Go wraps around; Ant panics. `Bits.wrapAdd` wraps on purpose.
- **Literals have one type.** `5` is an `int` and `2.0` a `float64`, with no untyped constants, so
  `let small: int8 = 5` is a compile error; write `int8(5)`. See
  [Generics](../guide/generics.md#coming-from-other-languages).
- **`defer` belongs to the block.** It runs when the enclosing block exits, not the function.
- **Maps keep insertion order.** `for` visits entries in the order their keys were first inserted.
- **No redeclaring.** A name is declared once per scope, so there is no `:=` that declares one name
  and reassigns another, and no name may shadow a prelude name such as `len` or `String`.
- **Case is spelling.** `export` exports a name; a capital letter does nothing. See
  [Values](../guide/values.md#coming-from-other-languages).
- **Functions go on your own types.** `func Point.norm` needs `Point` declared in the same module,
  so the prelude's types and another module's types cannot be extended.
- **Tests take no handle.** A test has no `*testing.T`, and a failed `assert` ends it. See
  [Testing](../guide/testing.md#coming-from-other-languages).

## See also

[Tour](../tour/01-hello.md) · [Idiomatic Ant](../guide/idiomatic.md) · [Errors](../guide/errors.md)
· [Concurrency](../guide/concurrency.md) · [Coming from JavaScript](javascript.md) ·
[Coming from Python](python.md)
