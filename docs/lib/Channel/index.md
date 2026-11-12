<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel

```ant signature
type Channel[T]
```

Channel carries values of T from goroutine to goroutine, blocking the side that is ahead. Channels
compare by identity.

<!-- end generated -->

A channel connects goroutines: one sends a value, another receives it. The `Channel` namespace holds
the functions that make a channel, send and receive on it, close it, and ask how many values it
holds. A `select` waits on several channels at once.

## Details

- [`Channel.new`](new.md) fixes `T` from nothing, so the channel needs an expected type: an
  annotated `let`, a parameter, a record field, or a return type.
- A channel of capacity 0, the default, hands each value over directly: `send` waits for a receiver.
  A buffered channel holds up to `capacity` values and blocks a sender only when it is full.
- After [`Channel.close`](close.md), receivers drain what is buffered and then get `null`. A send on
  a closed channel panics, and so does a second close, so the sending side closes. `receive` with a
  `ctx` also returns `null` when the context ends; `Context.isDone(ctx)` tells that from a close.

## Basic examples

A goroutine sends a value with `Channel.send`, and the main goroutine waits for it with
`Channel.receive`:

```ant
let results: Channel[int] = Channel.new()
go { Channel.send(results, 6 * 7) }
println(Channel.receive(results))
```

```text
42
```

A buffered channel holds values without a receiver. After a close, a `for` loop over
`Channel.values` drains them and ends:

```ant
let words: Channel[string] = Channel.new(capacity: 3)
words.send("ant")
words.send("bee")
words.close()
for word in Channel.values(words) {
  println(word)
}
```

```text
ant
bee
```

## Scope

A channel of a union carries several kinds of message. An `is` check tells them apart:

```ant
type Job = { name: string }
type Stop = {}

let inbox: Channel[Job | Stop] = Channel.new()
go {
  inbox.send(Job{ name: "build" })
  inbox.send(Job{ name: "test" })
  inbox.send(Stop{})
}
for message in inbox.values() {
  if message is Stop { break }
  println(message.name)
}
```

```text
build
test
```

A record field or a return type gives `Channel.new` its element type, as an annotation does:

```ant
type Worker = { name: string, inbox: Channel[string] }

func mailbox(): Channel[int] {
  Channel.new(2)
}

let w = Worker{ name: "w1", inbox: Channel.new(4) }
w.inbox.send("hello")
println(w.inbox.length())
println(mailbox().length())
```

```text
1
0
```

Channels compare by identity, and print as their type's name:

```ant
let a: Channel[int] = Channel.new()
let b: Channel[int] = Channel.new()
let alias = a
println(a == b)
println(a == alias)
println(a)
```

```text
false
true
Channel
```

## Functions by job

### Making a channel

- [Channel.new](new.md): an open channel that buffers up to `capacity` values; 0 is a rendezvous

### Sending and receiving

- [Channel.send](send.md): send a value, waiting until it is taken or buffered
- [Channel.receive](receive.md): the next value, or `null` once the channel is closed and drained
- [Channel.values](values.md): a sequence of every value until the channel is closed and drained

### Closing

- [Channel.close](close.md): close the channel; receivers drain it and then get `null`
- [Channel.isClosed](isClosed.md): whether the channel has been closed

### Inspecting

- [Channel.length](length.md): the number of values buffered now; the same as `len(c)`

## Applications

### A pipeline of stages

Each stage is a goroutine that receives from one channel and sends on the next. Each closes its
output when its input ends, so the close travels down the pipeline:

```ant
func generate(n: int): Channel[int] {
  let out: Channel[int] = Channel.new()
  go {
    defer { out.close() }
    for i in Iter.range(1, n + 1) {
      out.send(i)
    }
  }
  out
}

func square(input: Channel[int]): Channel[int] {
  let out: Channel[int] = Channel.new()
  go {
    defer { out.close() }
    for n in input.values() {
      out.send(n * n)
    }
  }
  out
}

for n in generate(4) | square() | Channel.values() {
  println(n)
}
```

```text
1
4
9
16
```

### Wait for a goroutine to finish

A receive on a channel of capacity 0 waits until the goroutine sends:

```ant
let done: Channel[bool] = Channel.new()
go {
  println("working")
  done.send(true)
}
done.receive()
println("finished")
```

```text
working
finished
```

### A worker pool

`Parallel.each` takes jobs from a channel, running at most `limit` callbacks at once, until the
producer closes it:

```ant
let jobs: Channel[string] = Channel.new()
go {
  defer { jobs.close() }
  for name in ["a.txt", "bb.txt", "ccc.txt"] {
    jobs.send(name)
  }
}

let total = Atomic.new(0)
Parallel.each(jobs, func(name: string, ctx: Context) { total.add(name.length()) }, limit: 2)
println(total.load())
```

```text
18
```

### Give up after a timeout

A `select` waits for whichever comes first, the reply or the timer:

```ant
let replies: Channel[string] = Channel.new()
go {
  Time.sleep(Time.duration(milliseconds: 200))
  replies.send("late")
}
let reply = select {
  let r = Channel.receive(replies) { r }
  Time.sleep(Time.duration(milliseconds: 20)) { "timed out" }
}
println(reply)
```

```text
timed out
```

## Properties and relations

`len(c)` and `Channel.length` agree, and a close keeps the buffered values: the length stays 2 until
a receiver drains them:

```ant
let c: Channel[int] = Channel.new(2)
c.send(1)
c.send(2)
assert(c.length() == 2)
assert(len(c) == c.length())
c.close()
assert(c.isClosed())
assert(c.length() == 2)
```

A channel of capacity 0 buffers nothing, so its length is always `0`:

```ant
let c: Channel[int] = Channel.new()
go { c.send(1) }
assert(c.length() == 0)
assert(c.receive() == 1)
```

## Possible issues

### A channel needs an element type

Without an expected type, `Channel.new` has nothing to fix `T`:

```ant
let c = Channel.new() // error: cannot infer the type parameter T of Channel.new; annotate the result
```

### Sending on a closed channel panics

A value sent after the close has no receiver, so the send panics:

```ant
let c: Channel[int] = Channel.new(1)
c.close()
let sent = c.send(1) // panic: Channel.send: the channel is closed
```

Closing twice panics too, so close a channel in one place, on the sending side:

```ant
func finish(c: Channel[int]): void {
  c.close() // panic: Channel.close: the channel is closed already
}

let c: Channel[int] = Channel.new()
finish(c)
finish(c)
```

### A send with no receiver waits

On a channel of capacity 0, a send waits until another goroutine receives. With no receiver, it
waits until the run is cancelled. A context bounds the wait, and the send returns `false`:

```ant
let c: Channel[int] = Channel.new()
let ctx = Context.new(timeout: Time.duration(milliseconds: 20))
println(c.send(1, ctx: ctx))
```

```text
false
```

A receive with a context ends the same way. It returns `null` as after a close, and `Context.isDone`
tells the two apart:

```ant
let c: Channel[int] = Channel.new()
let ctx = Context.new(timeout: Time.duration(milliseconds: 20))
println(c.receive(ctx: ctx))
println(ctx.isDone())
println(c.isClosed())
```

```text
null
true
false
```

### A null value looks like the end

When `T` includes `null`, a sent `null` makes `receive` return `null` exactly as the end of the
channel does:

```ant
let c: Channel[int | null] = Channel.new(1)
c.send(null)
println(c.receive())
c.close()
println(c.receive())
```

```text
null
null
```

Wrap the value in a record, so `null` from `receive` means only the end:

```ant
type Msg = { value: int | null }

let c: Channel[Msg] = Channel.new(2)
c.send(Msg{ value: null })
c.send(Msg{ value: 7 })
c.close()
for msg in c.values() {
  println(msg.value)
}
println("closed")
```

```text
null
7
closed
```

## See also

- [go](../../language/keywords/go.md), [select](../../language/keywords/select.md): start a
  goroutine, and wait on several channels at once
- [Context](../Context/index.md): bound a wait with a deadline or a cancellation
- [Parallel](../Parallel/index.md): run work from an array or a channel on goroutines
- [WaitGroup](../WaitGroup/index.md): wait for goroutines to finish
- [Time.ticker](../Time/ticker.md): a channel that receives the time at intervals
- [Concurrency](../../guide/concurrency.md): how goroutines, channels, and sync objects fit together
