<!-- status: reviewed -->

# Runtime panics

A panic stops the current goroutine. The type checker rules out every panic that depends on types,
so a checked program panics only in the ways this page lists, which depend on values known at run
time. Each one is a bug to fix or a state the program cannot go on from; a failure a caller is meant
to handle is a `Result`.

## Details

- When nothing catches a panic, the runtime writes `panic: `, the message, and a stack trace to
  standard error, innermost frame first. Each frame names a function and the line and column of the
  operation or call:

  ```text
  panic: index 3 out of range for an array of length 3
    at last (main.ant:2:5)
    at module main.ant (main.ant:5:13)
  ```

- A panic in the main goroutine ends the program with a non-zero exit status. A panic in any other
  goroutine ends only that goroutine, and the program goes on. A panic that leaves a `Parallel`
  callback panics again in the caller, with its message and trace, once the other callbacks have
  stopped.
- As a panic unwinds, every block it leaves runs its deferred blocks. A panic inside a deferred
  block replaces the one in progress.
- `Panic.catch` stops a panic and returns it as an `Err` holding a `Panic` record, with the message
  and the trace. Nothing is reported then.

## Basic examples

An index past the end of an array panics. The comment marks the line and gives the message:

```ant
func last(xs: Array[int]): int {
  xs[xs.length()] // panic: index 3 out of range for an array of length 3
}

println(last([1, 2, 3]))
```

Deferred blocks run on the way out, so cleanup still happens:

```ant
func load(path: string): int {
  println("opened ${path}")
  defer { println("closed ${path}") }
  10 / 0 // panic: division by zero
}

load("data.txt")
```

```text
opened data.txt
closed data.txt
```

`Panic.catch` turns a panic into a value:

```ant
let r = Panic.catch({ [1, 2, 3][5] })
if r is Err {
  println(r.error.message)
}
```

```text
index 5 out of range for an array of length 3
```

## Scope

The complete list of operations that panic in a checked program follows.

### Integer overflow

Integer `+`, `-`, `*`, and unary `-` panic when the result does not fit the kind. There is no silent
wraparound:

```ant
Math.maxInt + 1     // panic: integer overflow
-Math.minInt        // panic: integer overflow
int8(100) * int8(2) // panic: integer overflow
byte(0) - byte(1)   // panic: integer overflow
```

Compound assignment overflows the same way:

```ant
var n = Math.maxInt
n += 1 // panic: integer overflow
```

`Array.sum` panics when a partial sum does not fit, and `Atomic.add` when the new value does not:

```ant
[Math.maxInt, 1].sum() // panic: Array.sum: integer overflow
```

```ant
let hits = Atomic.new(Math.maxInt)
hits.add(1) // panic: Atomic.add: integer overflow
```

Float arithmetic never panics: it gives an infinity or `NaN`.

### Integer division by zero

Integer `/` and `%` panic on a zero divisor:

```ant
10 / 0    // panic: division by zero
10 % 0    // panic: division by zero
1.0 / 0.0 // Infinity
```

### Bit shifts

`Bits.shiftLeft` and `Bits.shiftRight` panic when the count is negative or not below the operand's
width:

```ant
Bits.shiftLeft(1, 63)       // -9223372036854775808
Bits.shiftLeft(1, 64)       // panic: Bits.shiftLeft needs a count from 0 up to but excluding 64, got 64
Bits.shiftRight(byte(1), 8) // panic: Bits.shiftRight needs a count from 0 up to but excluding 8, got 8
```

### Numeric conversion

A conversion to an integer kind panics on a fraction, a value out of range, `NaN`, or an infinity.
It never rounds or truncates on its own:

```ant
int(2.0)             // 2
int(2.5)             // panic: 2.5 has a fraction, so it does not convert to int
byte(300)            // panic: 300 out of range for byte
int(Math.NaN)        // panic: NaN does not convert to int
int(1.0 / 0.0)       // panic: Infinity does not convert to int
int(Math.trunc(2.5)) // 2
```

### Index out of range

Reading or writing an array with an index outside `0` up to but excluding the length panics. A
negative index is out of range too:

```ant
let xs = [10, 20, 30]
xs[-1]            // panic: index -1 out of range for an array of length 3
"ab".toBytes()[2] // panic: index 2 out of range for Bytes of length 2
```

```ant
let xs = [10, 20]
xs[2] = 30 // panic: index 2 out of range for an array of length 2
```

`xs.get(i)` returns `null` instead of panicking:

```ant
[10, 20, 30].get(5) // null
```

### Destructuring

Destructuring an array panics when it has fewer elements than the pattern's fixed positions:

```ant
let point = [3, 4]
let [x, y, z] = point // panic: a destructuring needs 3 elements, got 2
```

### Library functions

A library function panics in the cases its page documents, such as `Result.unwrap` on an `Err`,
`Mutex.unlock` when the mutex is not locked, or `Channel.send` on a closed channel:

```ant
let port: Result[int, Error] = Result.fail("no port")
port.unwrap() // panic: Result.unwrap on an Err: Error{ message: "no port" }
```

```ant
let lock = Mutex.new()
lock.unlock() // panic: Mutex.unlock: unlock of a Mutex that is not locked
```

```ant
func publish(c: Channel[int]) {
  c.send(1) // panic: Channel.send: the channel is closed
}

let updates: Channel[int] = Channel.new(1)
updates.close()
publish(updates)
```

### `panic` and `assert`

`panic` always panics, with the message it is given. `assert` panics when its condition is `false`,
and its message shows the values the condition compared:

```ant
panic("unreachable state") // panic: unreachable state
```

```ant
let stock = 2
let wanted = 5
assert(stock >= wanted, "not enough stock") // panic: assertion failed: stock >= wanted (2 >= 5): not enough stock
```

### Call depth

Every call panics when the goroutine's call depth exceeds the runtime's limit, as unbounded
recursion does:

```ant
func countdown(n: int): int {
  countdown(n - 1) // panic: call depth exceeded 10000
}

countdown(3)
```

### Reading a variable before its initializer

A top-level `var` or `let` panics when it is read or assigned before its initializer has run, and so
does a local read through a call made from its own initializer. Here `rate()` runs while the module
initializes `price`, before `tax` has a value:

```ant
func rate(): float64 {
  tax // panic: tax is read before its initializer has run
}

let price = 100.0 * rate()
let tax = 0.2
```

Declare `tax` before the line that uses it.

A local's initializer may name the local only inside a nested function. Calling that function while
the initializer still runs reads the local too early:

```ant
func setup() {
  let total: int = (func(): int { total + 1 })() // panic: total is read before its initializer has run
}

setup()
```

### Counts and result sizes

A count above 2147483647 panics in the functions that size what the runtime must hold, such as
`Array.filled` and `String.repeat`. So do `String.repeat`, `String.padStart`, `String.padEnd`, and
`String.replace` when the string they would return is longer than 2147483647 bytes:

```ant
Array.filled(3000000000, 0) // panic: Array.filled needs a count of at most 2147483647, got 3000000000
"ab".repeat(2000000000)     // panic: String.repeat would give a string longer than 2147483647 bytes
```

### HTML attribute spreads

A spread `{...attrs}` panics when it gives a name that is not an attribute name, an attribute that
runs code, `key`, or a `bool` to an attribute that is not boolean. The checker sees a map's type,
not its keys:

```ant
let attrs: Html.Attributes = Map{ "onclick": "go()" }
let button = <button {...attrs}>Go</button> // panic: onclick takes only literal text, so a spread cannot give it
```

```ant
let attrs: Html.Attributes = Map{ "href": true }
let link = <a {...attrs}>Home</a> // panic: href is not a boolean attribute, so it takes a string, not a bool
```

```ant
let attrs: Html.Attributes = Map{ "key": "1" }
let item = <li {...attrs}>One</li> // panic: key is reserved, so a spread cannot give it
```

```ant
let attrs: Html.Attributes = Map{ "data id": "1" }
let item = <li {...attrs}>One</li> // panic: "data id" is not an attribute name
```

### `select` on a closed channel

A `select` panics when it chooses a send clause on a closed channel:

```ant
let out: Channel[int] = Channel.new(1)
out.close()
select { // panic: select: a send clause was chosen on a closed channel
  Channel.send(out, 1) { println("sent") }
}
```

### Sequences in `for`

A `for` over a sequence function panics when the sequence calls `yield` on another goroutine, after
`yield` returned `false`, or after the loop ended:

```ant
func pair(yield: func(int): bool) {
  yield(1)
  yield(2) // panic: yield called again after it returned false; a sequence must return when yield returns false
}

for x in pair {
  break
}
```

A sequence that catches a panic of the loop body and returns normally panics too, so the body's
panic is not lost:

```ant
func guarded(yield: func(int): bool) {
  Panic.catch({ yield(1) })
}

for x in guarded { // panic: the sequence returned after a panic in the loop body: boom
  panic("boom")
}
```

## Applications

### One bad job does not stop the rest

Run each unit of work through `Panic.catch`, so a bug in one ends that unit only:

```ant
type Job = { id: int, items: Array[int] }

func average(job: Job): int {
  (job.items.sum() ?? 0) / job.items.length()
}

let jobs = [Job{ id: 1, items: [4, 6] }, Job{ id: 2, items: [] }, Job{ id: 3, items: [9] }]
for job in jobs {
  let done = Panic.catch({ average(job) })
  if done is Err {
    println("job ${job.id}: ${done.error.message}")
    continue
  }
  println("job ${job.id}: ${done.value}")
}
```

```text
job 1: 5
job 2: division by zero
job 3: 9
```

### Where the panic happened

The trace of a caught panic holds the frames the runtime would have written, innermost first:

```ant
func check(age: int): int {
  assert(age >= 0)
  age
}

let r = Panic.catch({ check(-1) })
if r is Err {
  for f in r.error.trace {
    println("${f.function} ${f.line}:${f.column}")
  }
}
```

```text
check 2:9
func 6:28
module main.ant 6:20
```

### A lock released by a panic

`Mutex.withLock` releases the lock as a panic leaves its block, so other goroutines are not stuck:

```ant
let lock = Mutex.new()
let r = Panic.catch({ lock.withLock({ panic("half done") }) })
println(r is Err)
lock.withLock({ println("lock is free") })
```

```text
true
lock is free
```

## Properties and relations

`??`, `?.`, `==`, and `!=` never panic themselves; only evaluating an operand can:

```ant
let missing: Array[int] | null = null
missing?.length() ?? 0 // 0
missing == null        // true
```

A panic in a `Parallel` callback reaches the caller, so a `Panic.catch` around the call stops it:

```ant
let r = Panic.catch({ Parallel.map([2, 0], func(n: int, work: Context) { 10 / n }, limit: 2) })
if r is Err {
  println(r.error.message)
}
```

```text
division by zero
```

`Panic.raise` passes a caught panic on with its message and trace unchanged:

```ant
let first = Panic.catch({ panic("lost") })
if first is Err {
  let again = Panic.catch({ Panic.raise(first.error) })
  if again is Err {
    assert(again.error.message == first.error.message)
    assert(again.error.trace == first.error.trace)
  }
}
```

## Possible issues

### A goroutine's panic does not stop the program

A panic in a goroutine ends that goroutine, and the main goroutine goes on. Code that waits for it
must not wait for work that never happens: defer `wg.done()` so the count drops either way.

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  let empty: Array[int] = []
  println(empty[0])
}
wg.wait()
println("main goes on")
```

```text
main goes on
```

### `Panic.catch` does not reach into goroutines

`Panic.catch` stops panics in the goroutine that calls it only. A goroutine started inside `f`
panics on its own, and `catch` still returns `Ok`:

```ant
let wg = WaitGroup.new()
wg.add(1)
let r = Panic.catch({
  go {
    defer { wg.done() }
    Panic.catch({ panic("job failed") })
  }
  1
})
wg.wait()
println(r)
```

```text
Ok{ value: 1 }
```

The goroutine needs its own `Panic.catch`, inside the `go` block, around its unit of work, as here;
without it, the goroutine's panic is reported and ends that goroutine.

### A panic in a deferred block hides the first one

The panic from a deferred block replaces the one in progress, so only the last is reported:

```ant
let r = Panic.catch({
  defer { panic("cleanup failed") }
  panic("first problem")
})
if r is Err {
  println(r.error.message)
}
```

```text
cleanup failed
```

### Catching a panic is not error handling

A panic carries only a message, so a caller has nothing to branch on. Return a `Result` for a
failure the caller is meant to handle, and check the input before the operation that would panic:

```ant
func ratio(a: int, b: int): Result[int, Error] {
  if b == 0 {
    return Result.fail("ratio of ${a} to zero")
  }
  Result.ok(a / b)
}

println(ratio(6, 3))
println(ratio(6, 0))
```

```text
Ok{ value: 2 }
Err{ error: Error{ message: "ratio of 6 to zero" } }
```

### Caught data may be half changed

A panic can leave a value half updated. What a block wrote before the panic stays written:

```ant
let totals = [0, 0]
let r = Panic.catch({
  totals[0] = 5
  totals[2] = 7
})
println(r is Err)
println(totals)
```

```text
true
[5, 0]
```

## See also

[panic](../lib/prelude/panic.md) · [assert](../lib/prelude/assert.md) ·
[Panic.catch](../lib/Panic/catch.md) · [Panic.raise](../lib/Panic/raise.md) ·
[defer](keywords/defer.md) · [go](keywords/go.md) · [select](keywords/select.md) ·
[Arithmetic operators](operators/arithmetic.md) · [Result](../lib/Result/index.md)
