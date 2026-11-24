<!-- status: reviewed -->

# Ant for Go developers

Ant borrows a lot from Go: goroutines, channels, `select`, `defer`, and a standard library that
behaves like Go's. The syntax is closer to TypeScript, and a few ideas are Ant's own. This page
shows one program in both languages, then the handful of differences you meet first. The
[Tour](../tour/01-hello.md) covers the rest.

## A first program

Here is a Go program that reads a port number from a file:

```go
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readPort(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func main() {
	port, err := readPort("port.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("listening on", port)
}
```

The same program in Ant:

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

let port = try readPort("port.txt")
println("listening on ${port}")
```

```text
listening on 8080
```

A few things to notice:

- There is no `package main` and no `func main`. Statements at the top of a file make it a script,
  and they run from top to bottom.
- There are no imports. `File`, `String`, `Time`, and the rest are in every file.
- A function that can fail returns a `Result` instead of a second `error` value.
- `try` is `if err != nil { return err }` in one word. At the top of a script, a failed `try` prints
  the error and exits with status 1.
- `toInt` returns an `int | null`: `null` when the text is not a number.
- The last expression of a block is its value, so `Result.ok(port)` needs no `return`.

## What carries over

| Go                                   | Ant                                       |
| ------------------------------------ | ----------------------------------------- |
| `x := 1`                             | `var x = 1`, or `let x = 1` if it stays   |
| `type Point struct { X, Y float64 }` | `type Point = { x: float64, y: float64 }` |
| `func (p Point) Norm() float64`      | `func Point.norm(p: Point): float64`      |
| `[]int{1, 2, 3}`                     | `[1, 2, 3]`                               |
| `xs = append(xs, 4)`                 | `xs.push(4)`                              |
| `map[string]int{"a": 1}`             | `Map{ a: 1 }`                             |
| `for _, x := range xs`               | `for x in xs`                             |
| `for cond { ... }`                   | `while cond { ... }`                      |
| `fmt.Sprintf("%d items", n)`         | `"${n} items"`                            |
| `2 * time.Second`                    | `2s`                                      |
| `go f(x)`                            | `go { f(x) }`                             |
| `ch := make(chan int)`               | `let ch: Channel[int] = Channel.new()`    |
| `ch <- v`, `v := <-ch`               | `ch.send(v)`, `let v = ch.receive()`      |
| `defer f.Close()`                    | `defer { f.close() }`                     |
| `go run`, `go test`, `gofmt`         | `ant run`, `ant test`, `ant fmt`          |

## Concurrency feels like home

Goroutines and channels work as you expect. `go` takes a block, and a channel carries results back:

```ant
let results: Channel[int] = Channel.new()
for n in [1, 2, 3] {
  go { results.send(n * n) }
}

var total = 0
for _ in Iter.range(0, 3) {
  total += results.receive() ?? 0
}
println(total)
```

```text
14
```

`receive` returns `null` once the channel is closed, which replaces Go's `v, ok := <-ch`. For the
common worker-pool shape, [`Parallel.map`](../lib/Parallel/map.md) starts the goroutines, waits for
them, and keeps the results in order. See [Concurrency](../guide/concurrency.md).

## What works differently

### Unions instead of `nil` and type switches

A value that may be missing has a type that says so, such as `User | null`. A plain `User` is never
`nil`. A map read gives `null` for a missing key, so `??` replaces the comma-ok form:

```ant
let ports = Map{ http: 80, https: 443 }
ports["https"]     // 443
ports["ssh"]       // null
ports["ssh"] ?? 22 // 22
```

The same idea replaces most type switches. List the types a value can be, and test with `is`:

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

println(area(Rect{ width: 2.0, height: 3.0 }))
```

```text
6.0
```

The `if` needs no `else` because it covers every member of `Shape`. Add a `Triangle` to the union,
and `ant check` points at this function until you handle it. Interfaces still exist for open sets,
and a type satisfies one without saying so, as in Go. See [Unions](../guide/unions.md).

### Records are shared

Assigning or passing a record shares it, as a pointer does in Go. There is no `&` and no `*`. To
copy, spread it into a new literal:

```ant
type Point = { x: int, y: int }

let p = Point{ x: 1, y: 2 }
let shared = p
let copy = Point{ ...p }
shared.x = 10
println(p.x)
println(copy.x)
```

```text
10
1
```

### `ant run` does not check types

`ant check` is the type checker, and `ant run` does not call it. A program with a type error still
starts, and panics when it reaches the wrong value. To get what `go build` gives you, run
`ant check` in your editor and in CI:

```ant
func double(n: int): int { n * 2 }

println(double("21")) // error: cannot use string where int is expected
```

### Smaller things

- `export func` exports a name. A capital letter means nothing.
- Fields have no zero value. A record literal gives every field, unless the type declares a default.
- Integer overflow panics instead of wrapping around.
- `defer` runs when its block ends, not when the function returns.
- A map keeps the order its keys were inserted in.

## Where to go next

- The [Tour](../tour/01-hello.md) walks through the language one small program at a time.
- [Errors](../guide/errors.md) covers wrapping errors and errors of your own type.
- [Idiomatic Ant](../guide/idiomatic.md) shows how Ant code usually looks.
- [Coming from JavaScript](javascript.md) and [Coming from Python](python.md).
