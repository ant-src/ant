<!-- status: reviewed -->

# `select`

`select` waits for the first of several blocking calls that can proceed, performs it, and runs the
block written for it. Each clause is a head, such as `Channel.receive(c)`, followed by a block.

## Details

- A head is one of five calls, written in namespace form: `Channel.receive(c)`,
  `Channel.send(c, v)`, `Context.wait(ctx)`, `WaitGroup.wait(wg)`, or `Time.sleep(d)`. A head takes
  no context; a `Context.wait` clause stands for it.
- The operands of every head evaluate first, in order, once. Then `select` waits until a head can
  proceed. When several can, it picks one at random, with no order promised.
- `let x =` before a head binds its result in the block: a receive binds `T | null`, `null` when the
  channel is closed, and a send binds `bool`.
- An `else` clause, written last, runs when no head can proceed at once, so a `select` with `else`
  never waits, except that a send head on a closed channel is still chosen and panics.
- A `select` is an expression whose value is the chosen block's value. `return`, `break`, and
  `continue` in a block act on the function or loop around the `select`.

## Basic examples

Wait for a reply, but no longer than a timeout:

```ant
let replies: Channel[string] = Channel.new()
go {
  Time.sleep(1s)
  replies.send("pong")
}
select {
  let reply = Channel.receive(replies) { println("got ${reply}") }
  Time.sleep(20ms) { println("timed out") }
}
```

```text
timed out
```

With `else`, `select` takes a value only if one is already there, and never waits:

```ant
let jobs: Channel[string] = Channel.new(capacity: 1)
jobs.send("build")
for _ in [1, 2] {
  select {
    let job = Channel.receive(jobs) { println("running ${job}") }
    else { println("no job") }
  }
}
```

```text
running build
no job
```

## Scope

### Receive

A receive head binds `T | null`. It binds `null` once the channel is closed and drained:

```ant
let words: Channel[string] = Channel.new(capacity: 1)
words.send("ant")
words.close()
for _ in [1, 2] {
  select {
    let word = Channel.receive(words) { println(word ?? "closed") }
  }
}
```

```text
ant
closed
```

### Send

A send head waits until the channel can take the value. Its binding is a `bool`, and with `else` a
full channel takes the `else` branch:

```ant
let out: Channel[int] = Channel.new(capacity: 1)
for n in [1, 2] {
  select {
    let sent = Channel.send(out, n) { println("sent ${n}: ${sent}") }
    else { println("full, dropped ${n}") }
  }
}
```

```text
sent 1: true
full, dropped 2
```

### Contexts

A `Context.wait` head proceeds when the context is done, by cancellation or deadline:

```ant
let ctx = Context.new()
let jobs: Channel[int] = Channel.new()
go { ctx.cancel() }
select {
  Context.wait(ctx) { println("stopped: ${ctx.reason() ?? ""}") }
  let job = Channel.receive(jobs) { println("job ${job}") }
}
```

```text
stopped: cancelled
```

### Wait groups

A `WaitGroup.wait` head proceeds when the group's counter is zero, so a group of goroutines can race
a deadline:

```ant
let wg = WaitGroup.new()
let squares: Channel[int] = Channel.new(capacity: 3)
for n in [1, 2, 3] {
  wg.add(1)
  go {
    defer { wg.done() }
    squares.send(n * n)
  }
}
select {
  WaitGroup.wait(wg) { println("all ${squares.length()} done") }
  Time.sleep(5s) { println("too slow") }
}
```

```text
all 3 done
```

### As an expression

The value of `select` is the value of the block it runs. Without an expected type, it is the union
of the blocks' types:

```ant
let counts: Channel[int] = Channel.new(capacity: 1)
counts.send(5)
let n = select {
  let count = Channel.receive(counts) { count ?? 0 }
  else { "none" }
}
println(n)
```

```text
5
```

With an expected type, each block is checked against it. A union of records names each outcome:

```ant
type Line = { text: string }
type Closed = {}
type Timeout = {}

let lines: Channel[string] = Channel.new(capacity: 1)
lines.send("hello")
let event: Line | Closed | Timeout = select {
  let line = Channel.receive(lines) { if line == null { Closed{} } else { Line{ text: line } } }
  Time.sleep(1s) { Timeout{} }
}
println(event)
```

```text
Line{ text: "hello" }
```

### Leaving a loop or function

A block belongs to the body around the `select`, so its `break` leaves the loop and its `return`
leaves the function:

```ant
func firstLong(words: Channel[string]): string | null {
  while true {
    select {
      let word = Channel.receive(words) {
        if word == null { break }
        if word.length() > 3 { return word }
      }
    }
  }
  null
}

let words: Channel[string] = Channel.new(capacity: 3)
words.send("ant")
words.send("beetle")
words.close()
println(firstLong(words))
```

```text
beetle
```

## Applications

Merge two channels into one stream, and stop when both are closed. A closed channel is always ready,
so the loop drops it from the `select` by replacing it with a channel that never delivers:

```ant
func merge(a: Channel[int], b: Channel[int]): Array[int] {
  let silent: Channel[int] = Channel.new()
  var left = a
  var right = b
  var open = 2
  let all: Array[int] = []
  while open > 0 {
    select {
      let x = Channel.receive(left) {
        if x == null {
          left = silent
          open -= 1
        } else {
          all.push(x)
        }
      }
      let y = Channel.receive(right) {
        if y == null {
          right = silent
          open -= 1
        } else {
          all.push(y)
        }
      }
    }
  }
  all.sort()
}

let odds: Channel[int] = Channel.new()
let evens: Channel[int] = Channel.new()
go {
  defer { odds.close() }
  for n in [1, 3, 5] { odds.send(n) }
}
go {
  defer { evens.close() }
  for n in [2, 4] { evens.send(n) }
}
println(merge(odds, evens))
```

```text
[1, 2, 3, 4, 5]
```

A worker that handles jobs until its context is cancelled:

```ant
func work(jobs: Channel[string], done: Channel[string], ctx: Context) {
  while true {
    select {
      Context.wait(ctx) { return }
      let job = Channel.receive(jobs) {
        if job == null { return }
        done.send("done ${job}")
      }
    }
  }
}

let ctx = Context.new()
let jobs: Channel[string] = Channel.new()
let done: Channel[string] = Channel.new()
go { work(jobs, done, ctx) }
for job in ["lint", "test"] {
  jobs.send(job)
  println(done.receive() ?? "")
}
ctx.cancel()
```

```text
done lint
done test
```

## Properties and relations

A `select` with one receive head and no `else` waits the way `Channel.receive` does and gets the
same value:

```ant
let c: Channel[int] = Channel.new(capacity: 2)
c.send(7)
c.send(7)
let viaSelect = select {
  let v = Channel.receive(c) { v }
}
assert(viaSelect == c.receive())
```

## Possible issues

A head is written only in namespace form. The dot form is the same call elsewhere, but not a head:

```ant
let c: Channel[int] = Channel.new(capacity: 1)
select {
  c.receive() { println("got") }
  // error: a select clause's head must be a call of Channel.receive, Channel.send, Context.wait, WaitGroup.wait or Time.sleep
}
```

A head takes no context. Write the context as a clause of its own:

```ant
let c: Channel[int] = Channel.new(capacity: 1)
let ctx = Context.new()
select {
  Channel.receive(c, ctx) { println("got") }
  // error: a select clause's head takes no context: the context is a clause of its own, Context.wait(ctx) { ... }
}
```

A receive binds `T | null`, not `T`. Handle the `null` of a closed channel:

```ant
let c: Channel[int] = Channel.new(capacity: 1)
let n: int = select {
  let v = Channel.receive(c) { v } // error: cannot use int | null where int is expected
  else { 0 }
}
```

`Context.wait`, `WaitGroup.wait`, and `Time.sleep` return `void`, so their heads bind nothing:

```ant
let ctx = Context.new()
select {
  let x = Context.wait(ctx) { println("done") }
  // error: Context.wait returns void, so there is nothing to bind to x
}
```

The block starts on the head's line. A `{` on the next line is a syntax error:

```ant
let c: Channel[int] = Channel.new(capacity: 1)
select {
  let v = Channel.receive(c) // error: expected '{' after the select clause header, got newline
  { println(v) }
}
```

A send on a closed channel is always ready, so `select` picks it over `else` and panics:

```ant
let c: Channel[int] = Channel.new(capacity: 1)
c.close()
select { // panic: select: a send clause was chosen on a closed channel
  Channel.send(c, 1) { println("sent") }
  else { println("closed") }
}
```

A closed channel's receive is always ready. A loop that keeps it in its `select` and does not leave
on `null` spins, picking the closed channel again and again. Leave the loop on `null`, as in
[Leaving a loop or function](#leaving-a-loop-or-function), or swap the channel out, as in
Applications.

## See also

- [go](go.md): start the goroutines a `select` waits on
- [break](break.md), [continue](continue.md), and [return](return.md) in a `select` block
- [Channel.receive](../../lib/Channel/receive.md) and [Channel.send](../../lib/Channel/send.md)
- [Channel](../../lib/Channel/index.md)
- [Context.wait](../../lib/Context/wait.md)
- [WaitGroup.wait](../../lib/WaitGroup/wait.md)
- [Time.sleep](../../lib/Time/sleep.md)
