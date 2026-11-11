<!-- status: reviewed -->

# Channels and select

A channel carries values from one goroutine to another. A `select` waits for whichever of several
things happens first, and a context puts a limit on how long a wait lasts.

```ant
let jobs: Channel[int] = Channel.new()
let results: Channel[string] = Channel.new(capacity: 3)

go {
  for n in Iter.range(1, 4) {
    jobs.send(n)
  }
  jobs.close()
}

go {
  for n in jobs.values() {
    results.send("${n} squared is ${n * n}")
  }
  results.close()
}

for line in results.values() {
  println(line)
}

let replies: Channel[string] = Channel.new()
let ctx = Context.new(Time.duration(milliseconds: 50))
select {
  let reply = Channel.receive(replies) { println("reply: ${reply}") }
  Context.wait(ctx) { println("gave up: ${ctx.reason() ?? "running"}") }
}

let mailbox: Channel[string] = Channel.new(capacity: 1)
for _ in Iter.range(0, 2) {
  select {
    let note = Channel.receive(mailbox) { println("note: ${note}") }
    else {
      println("mailbox empty, leaving a note")
      mailbox.send("hello")
    }
  }
}
```

```text
1 squared is 1
2 squared is 4
3 squared is 9
gave up: timeout
mailbox empty, leaving a note
note: hello
```

The program builds a small pipeline. The first goroutine sends 1, 2, and 3 into `jobs`, then closes
it. The second receives each number, sends a line into `results`, and closes `results` when `jobs`
runs dry. The main goroutine prints the lines as they arrive.

`jobs` has no capacity, so each `send` waits until the other side takes the value: the two
goroutines meet at every handoff. `results` holds up to three values, so its sender runs ahead until
the buffer is full. `Channel.new` cannot guess what a channel carries, so you annotate it, as in
`Channel[int]`.

`values` is a sequence of everything the channel receives, and it ends when the channel is closed
and drained. That is why each sender closes its channel: the `for` loop on the other side stops
there.

The first `select` waits on two things: a reply that never comes, and the context `ctx`, which is
done after 50 milliseconds. Whichever is ready first runs its block, here the context's. The
`let reply =` before the receive names the value it got, for use in that block. `reply` is a
`string | null`: a receive gives `null` once the channel is closed, and a closed channel is always
ready, so a `select` in a loop leaves the loop when it gets `null`. A `select` head is written in
namespace form, as `Channel.receive(replies)`.

The second `select` has an `else`, so it never waits. On the first lap the mailbox is empty, nothing
is ready, and `else` leaves a note. On the second lap the note is there, and the receive takes it.

Try giving `replies` a sender, such as `go { replies.send("pong") }` right after it is created, and
the reply arrives before the timeout.

A send on a closed channel is a bug in the sender, and it panics:

```ant
let done: Channel[bool] = Channel.new(capacity: 1)
done.close()
_ = done.send(true) // panic: Channel.send: the channel is closed
```

The [concurrency chapter](../guide/concurrency.md) explains when to pick a channel, a `select`, a
lock, or `Parallel`. The reference has [`select`](../language/keywords/select.md),
[Channel](../lib/Channel/index.md) with [Channel.new](../lib/Channel/new.md),
[Channel.send](../lib/Channel/send.md), [Channel.receive](../lib/Channel/receive.md),
[Channel.close](../lib/Channel/close.md), and [Channel.values](../lib/Channel/values.md), and
[Context](../lib/Context/index.md) with [Context.new](../lib/Context/new.md),
[Context.wait](../lib/Context/wait.md), and [Context.reason](../lib/Context/reason.md).

Next: [Modules](19-modules.md).
