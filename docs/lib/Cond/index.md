<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cond

```ant signature
type Cond
```

Cond is a condition variable: goroutines wait on it, under its mutex, until another signals. Conds
compare by identity.

<!-- end generated -->

A `Cond` lets goroutines wait until shared state reaches a condition, such as a queue that is no
longer empty or a gate that is open. The state lives in ordinary variables guarded by a `Mutex`; the
cond only puts goroutines to sleep and wakes them. The `Cond` namespace holds a function that makes
a cond over a mutex, one that waits on it, and two that wake waiters.

## Details

- A cond belongs to the mutex it is made over. Lock that mutex, check the condition in a `while`
  loop, and call `wait` inside the loop. `wait` panics when the mutex is not locked.
- A signal is not remembered: `signal` and `broadcast` wake only goroutines waiting at that moment,
  and do nothing when none waits. Change the state under the mutex before you signal.

## Basic examples

`Cond.new` makes a cond over a mutex. The main goroutine waits under the mutex until another
goroutine sets the flag and calls `Cond.signal`:

```ant
let mu = Mutex.new()
let ready = Cond.new(mu)
var loaded = false

go {
  mu.withLock({
    loaded = true
    Cond.signal(ready)
  })
}

mu.withLock({
  while !loaded {
    Cond.wait(ready)
  }
})
println("loaded")
```

```text
loaded
```

`Cond.broadcast` wakes every waiter at once. Each one reacquires the mutex in turn before it goes
on:

```ant
let mu = Mutex.new()
let opened = Cond.new(mu)
var open = false
var waiting = 3
let wg = WaitGroup.new()

for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    mu.withLock({
      while !open {
        Cond.wait(opened)
      }
      waiting -= 1
    })
  }
}

mu.withLock({
  open = true
  Cond.broadcast(opened)
})
wg.wait()
println("${waiting} still waiting")
```

```text
0 still waiting
```

## Scope

A cond renders as its type's name:

```ant
Cond.new(Mutex.new()) // Cond
```

Several conds can share one mutex, one for each condition. Here two goroutines take turns, and each
waits on its own cond for its turn:

```ant
let mu = Mutex.new()
let pingTurn = Cond.new(mu)
let pongTurn = Cond.new(mu)
var turn = "ping"
let wg = WaitGroup.new()

func player(name: string, mine: Cond, other: Cond, next: string) {
  defer { wg.done() }
  for _ in Array.range(0, 2) {
    mu.withLock({
      while turn != name {
        mine.wait()
      }
      println(name)
      turn = next
      other.signal()
    })
  }
}

wg.add(2)
go { player("pong", pongTurn, pingTurn, "ping") }
go { player("ping", pingTurn, pongTurn, "pong") }
wg.wait()
```

```text
ping
pong
ping
pong
```

## Functions by job

### Making a cond

- [Cond.new](new.md): a new cond over a mutex

### Waiting

- [Cond.wait](wait.md): release the mutex, wait for a wake-up, and lock the mutex again

### Waking waiters

- [Cond.signal](signal.md): wake one waiting goroutine, if any
- [Cond.broadcast](broadcast.md): wake every waiting goroutine

## Applications

### A blocking queue

`take` waits while the queue is empty, and `put` wakes one waiting taker after it adds an item. One
producer feeds two consumers:

```ant
type Queue[T] = { mu: Mutex, notEmpty: Cond, items: Array[T] }

func Queue.new[T](): Queue[T] {
  let mu = Mutex.new()
  Queue{ mu, notEmpty: Cond.new(mu), items: [] }
}

func Queue.put[T](q: Queue[T], item: T) {
  q.mu.withLock({
    q.items.push(item)
    q.notEmpty.signal()
  })
}

func Queue.take[T](q: Queue[T]): T {
  q.mu.withLock({
    while q.items.isEmpty() {
      q.notEmpty.wait()
    }
    q.items.removeAt(0)
  })
}

let jobs: Queue[int] = Queue.new()
let takenMu = Mutex.new()
let taken: Array[int] = []
let wg = WaitGroup.new()

for _ in Array.range(0, 2) {
  wg.add(1)
  go {
    defer { wg.done() }
    for _ in Array.range(0, 3) {
      let job = jobs.take()
      takenMu.withLock({ taken.push(job) })
    }
  }
}

for job in Array.range(1, 7) {
  jobs.put(job)
}
wg.wait()
println(taken.sort())
```

```text
[1, 2, 3, 4, 5, 6]
```

### A barrier

Each worker loads its part, then waits at the barrier until every worker has arrived. The last one
to arrive wakes the rest, so no worker goes on before all parts are loaded:

```ant
type Barrier = { mu: Mutex, allIn: Cond, size: int, arrived: int }

func Barrier.new(size: int): Barrier {
  let mu = Mutex.new()
  Barrier{ mu, allIn: Cond.new(mu), size, arrived: 0 }
}

func Barrier.arrive(b: Barrier) {
  b.mu.withLock({
    b.arrived += 1
    if b.arrived == b.size { b.allIn.broadcast() }
    while b.arrived < b.size {
      b.allIn.wait()
    }
  })
}

let barrier = Barrier.new(3)
let loaded = Atomic.new(0)
let seen = Array.filled(3, 0)
let wg = WaitGroup.new()

for i in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    loaded.add(1)
    barrier.arrive()
    seen[i] = loaded.load()
  }
}
wg.wait()
println(seen)
```

```text
[3, 3, 3]
```

### Limit how many goroutines use a resource

A pool hands out a fixed number of permits. `acquire` waits while none is free, and `release`
returns one and wakes one waiter:

```ant
type Pool = { mu: Mutex, freed: Cond, free: int, inUse: int, peak: int }

func Pool.new(size: int): Pool {
  let mu = Mutex.new()
  Pool{ mu, freed: Cond.new(mu), free: size, inUse: 0, peak: 0 }
}

func Pool.acquire(p: Pool) {
  p.mu.withLock({
    while p.free == 0 {
      p.freed.wait()
    }
    p.free -= 1
    p.inUse += 1
    p.peak = Math.max(p.peak, p.inUse)
  })
}

func Pool.release(p: Pool) {
  p.mu.withLock({
    p.free += 1
    p.inUse -= 1
    p.freed.signal()
  })
}

let pool = Pool.new(2)
let wg = WaitGroup.new()
for _ in Array.range(0, 6) {
  wg.add(1)
  go {
    defer { wg.done() }
    pool.acquire()
    defer { pool.release() }
    Time.sleep(5ms)
  }
}
wg.wait()
println("at most ${pool.peak} at once")
```

```text
at most 2 at once
```

## Properties and relations

A cond equals itself and its copies, and no other cond, even one over the same mutex:

```ant
let mu = Mutex.new()
let a = Cond.new(mu)
let alias = a
assert(a == alias)
assert(a != Cond.new(mu))
```

With nobody waiting, `signal` and `broadcast` return at once and change nothing:

```ant
let c = Cond.new(Mutex.new())
c.signal()
c.broadcast()
println("nobody was waiting")
```

```text
nobody was waiting
```

## Possible issues

### Waiting without the mutex

`wait` panics when the cond's mutex is not locked:

```ant
let mu = Mutex.new()
let ready = Cond.new(mu)
ready.wait() // panic: Cond.wait: the Mutex is not locked
```

### Only a Mutex

A cond works over a `Mutex` only. `ant check` rejects an `RwMutex`:

```ant
let rw = RwMutex.new()
let changed = Cond.new(rw) // error: cannot use RwMutex where Mutex is expected
```

### Waiting without checking the state first

A signal sent before a goroutine waits wakes nobody. Check the state before you wait, so a goroutine
that comes late does not wait at all:

```ant
let mu = Mutex.new()
let ready = Cond.new(mu)
var done = false

mu.withLock({
  done = true
  ready.signal()
})

mu.withLock({
  while !done {
    ready.wait()
  }
})
println("no wait needed")
```

```text
no wait needed
```

Check with `while`, not `if`: a woken goroutine reacquires the mutex after others may have changed
the state again.

### No deadline

`wait` cannot give up. For a wait with a deadline, hand the event over on a channel, whose `receive`
takes a context and returns `null` when the deadline passes first:

```ant
let ready: Channel[bool] = Channel.new(capacity: 1)

go {
  Time.sleep(1s)
  ready.send(true)
}

let got = ready.receive(ctx: Context.new(50ms))
println(if got == null { "gave up" } else { "ready" })
```

```text
gave up
```

## See also

- [Mutex](../Mutex/index.md): the lock a cond waits under
- [WaitGroup](../WaitGroup/index.md): wait for goroutines to finish
- [Channel](../Channel/index.md): hand values between goroutines, with deadlines
- [Once](../Once/index.md): run work exactly once
- [Atomic](../Atomic/index.md): one value read and written atomically
- [Concurrency](../../guide/concurrency.md): how goroutines, channels, and sync objects fit together
