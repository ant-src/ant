<!-- status: reviewed -->

# `defer`

`defer { ... }` registers a block to run when the enclosing block exits.

## Details

- The deferred block belongs to the innermost enclosing block, not to the function: in a loop body
  it runs at the end of every iteration, and in an `if` branch when that branch ends.
- It runs however the block exits: at its end, by `return`, `break`, or `continue`, or by a panic.
  Several deferred blocks run in reverse order of registration.
- It runs in the scope it was registered in, so it reads variables as they are when it runs, not
  when `defer` registers it.
- Its value is discarded. A `return` inside it leaves the deferred block only, and takes no value.
- `Os.exit` ends the program without running any deferred block.

## Basic examples

The deferred block runs after the rest of the function body:

```ant
func greet() {
  defer { println("bye") }
  println("hello")
}

greet()
```

```text
hello
bye
```

Put `defer` right after acquiring a resource, so that every way out of the function releases it:

<!-- test: file notes.txt
first note
second note
-->

```ant
func firstLine(path: string): Result[string | null, Error] {
  let file = try File.open(path)
  defer { file.close() }
  file.readLine()
}

println(firstLine("notes.txt"))
```

```text
Ok{ value: "first note" }
```

## Scope

### Order

Deferred blocks run last registered, first run:

```ant
func steps() {
  defer { println("registered first, runs last") }
  defer { println("registered second, runs first") }
  println("body")
}

steps()
```

```text
body
registered second, runs first
registered first, runs last
```

### Early exits

A `return` runs the deferred blocks registered before it:

```ant
func check(n: int): string {
  defer { println("checked ${n}") }
  if n < 0 { return "negative" }
  "fine"
}

println(check(-1))
println(check(5))
```

```text
checked -1
negative
checked 5
fine
```

### Loop bodies

In a loop body, the deferred block runs at the end of each iteration, including one that `continue`
cuts short or `break` ends:

```ant
for i in [1, 2, 3, 4] {
  defer { println("done with ${i}") }
  if i == 2 { continue }
  if i == 3 { break }
  println("work on ${i}")
}
```

```text
work on 1
done with 1
done with 2
done with 3
```

### Any block

An `if` branch, or any other block, owns the deferred blocks registered in it:

```ant
let verbose = true
if verbose {
  defer { println("leaving the branch") }
  println("in the branch")
}
println("after the branch")
```

```text
in the branch
leaving the branch
after the branch
```

### Scripts

In a script's top-level code, a deferred block runs when the script ends:

```ant
defer { println("script done") }
println("last line")
```

```text
last line
script done
```

### Panics

A panic runs the deferred blocks of every block it leaves. Here `Panic.catch` stops it, after the
deferred block has run:

```ant
let result = Panic.catch({
  defer { println("cleaning up") }
  panic("boom")
})
if result is Err {
  println("caught: ${result.error.message}")
}
```

```text
cleaning up
caught: boom
```

### Variables

The deferred block reads a variable when it runs, so it sees the last value assigned before the
block exits:

```ant
func attempts() {
  var tries = 1
  defer { println("tries: ${tries}") }
  tries += 2
}

attempts()
```

```text
tries: 3
```

## Applications

### Release a lock

Unlock a `Mutex` on every path out of the function:

```ant
type Account = { balance: int, lock: Mutex }

func withdraw(a: Account, amount: int): bool {
  a.lock.lock()
  defer { a.lock.unlock() }
  if a.balance < amount { return false }
  a.balance -= amount
  true
}

let account = Account{ balance: 100, lock: Mutex.new() }
println(withdraw(account, 30))
println(withdraw(account, 500))
println(account.balance)
```

```text
true
false
70
```

### Signal that a goroutine finished

`defer { wg.done() }` at the top of a `go` block marks the work done even when the goroutine ends
early:

```ant
let squares = Array.filled(4, 0)
let wg = WaitGroup.new()
for i in [0, 1, 2, 3] {
  wg.add(1)
  go {
    defer { wg.done() }
    squares[i] = i * i
  }
}
wg.wait()
println(squares)
```

```text
[0, 1, 4, 9]
```

### Trace entry and exit

Print a line when a function starts and another when it ends, whichever way it ends:

```ant
func traced(name: string, depth: int): int {
  let indent = "  ".repeat(depth)
  println("${indent}enter ${name}")
  defer { println("${indent}leave ${name}") }
  if depth == 2 { return depth }
  traced("${name}.child", depth + 1) + 1
}

println(traced("root", 0))
```

```text
enter root
  enter root.child
    enter root.child.child
    leave root.child.child
  leave root.child
leave root
4
```

## Properties and relations

A function's result is evaluated before its deferred blocks run, so assigning the returned variable
in a deferred block does not change the result. A change to an array the result holds does show:

```ant
func count(): int {
  var n = 1
  defer { n = 10 }
  n
}

func items(): Array[int] {
  let xs = [1]
  defer { xs.push(2) }
  xs
}

assert(count() == 1, "the result is taken before the deferred block runs")
assert(items() == [1, 2], "the deferred block changes the array the result holds")
```

## Possible issues

### The block is required

`defer` takes a block, not an expression:

```ant
func f() {
  defer println("done") // error: expected '{', got identifier "println"
}
```

### A deferred block has no result

A `return` in a deferred block takes no value, and `try` cannot return from it:

```ant
func f(): int {
  defer { return 1 } // error: 'return' takes no value in a defer block
  2
}
```

```ant
func save(path: string): Result[void, Error] {
  let file = try File.create(path)
  defer { try file.close() } // error: try needs an enclosing function that returns a Result
  file.writeText("saved")
}
```

### A deferred block cannot leave the loop

`break` and `continue` in a deferred block target only a loop inside that block:

```ant
for i in [1, 2, 3] {
  defer { if i == 2 { break } } // error: 'break' outside a loop
  println(i)
}
```

### A deferred close drops its error

The value of a deferred block is discarded, so an error from a deferred `close` is lost. Close an
output explicitly at the end of the function to report it; the deferred close then covers the early
exits and does nothing more:

```ant
func save(path: string, lines: Array[string]): Result[void, Error] {
  let file = try File.create(path)
  defer { file.close() }
  for line in lines {
    try file.writeLine(line)
  }
  file.close()
}

println(save("out.txt", ["one", "two"]))
println(File.read("out.txt"))
```

```text
Ok{ value: void }
Ok{ value: "one\ntwo\n" }
```

### Cleanup in a branch runs too early

A `defer` inside an `if` runs when the branch ends, not when the function returns. Here the log is
closed before anything is written to it, and the write fails:

```ant
func run(log: bool): Result[void, Error] {
  var out: File.Handle | null = null
  if log {
    out = try File.create("run.log")
    defer { out?.close() }
  }
  if out != null {
    try out.writeLine("started")
  }
  Result.ok(void)
}

println(run(true) is Err)
```

```text
true
```

### `Os.exit` skips deferred blocks

`Os.exit` ends the process at once, without running any deferred block:

<!-- test: exit 3 -->

```ant
defer { println("never printed") }
println("exiting")
Os.exit(3)
```

```text
exiting
```

### A panic in a deferred block replaces the first

When a deferred block panics while a panic is in progress, the new panic takes its place:

```ant
let result = Panic.catch({
  defer { panic("second") }
  panic("first")
})
if result is Err {
  println(result.error.message)
}
```

```text
second
```

## See also

[go](go.md) · [return](return.md) · [break](break.md) · [continue](continue.md) · [try](try.md) ·
[Panic.catch](../../lib/Panic/catch.md) · [Os.exit](../../lib/Os/exit.md) ·
[File.close](../../lib/File/close.md) · [Mutex.withLock](../../lib/Mutex/withLock.md) ·
[WaitGroup.done](../../lib/WaitGroup/done.md) · [Errors](../../guide/errors.md)
