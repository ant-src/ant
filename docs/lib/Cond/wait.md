<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cond.wait

```ant signature
func Cond.wait(c: Cond): void
```

Cond.wait releases the mutex of c, waits for a signal or broadcast, and reacquires the mutex before
returning. Another goroutine may take the mutex and change the state first, so a waiter tests its
condition in a loop. It panics if the mutex is not locked. It takes no context.

<!-- end generated -->

## Details

- Call `wait` with the mutex of `c` locked.
- A [`Cond.signal`](signal.md) or [`Cond.broadcast`](broadcast.md) wakes only goroutines waiting at
  that moment, and another goroutine can lock the mutex and change the state before `wait` returns,
  so check the state in a `while` loop, before the first wait and after each one.

## Basic examples

Wait, under the mutex, until another goroutine sets a flag and signals:

```ant
let m = Mutex.new()
let ready = Cond.new(m)
var done = false

go {
  m.withLock({
    done = true
    ready.signal()
  })
}

m.lock()
defer { m.unlock() }
while !done {
  Cond.wait(ready)
}
println("done")
```

```text
done
```

`Mutex.withLock` holds the mutex around the loop and releases it at the end:

```ant
let m = Mutex.new()
let changed = Cond.new(m)
var count = 0

go {
  for _ in [1, 2, 3] {
    m.withLock({
      count += 1
      changed.signal()
    })
  }
}

m.withLock({
  while count < 3 {
    Cond.wait(changed)
  }
})
println(count)
```

```text
3
```

## Scope

### Several waiters

Several goroutines can wait on one cond. [`Cond.broadcast`](broadcast.md) wakes them all, and each
reacquires the mutex in turn:

```ant
let m = Mutex.new()
let opened = Cond.new(m)
let wg = WaitGroup.new()
let entered = Atomic.new(0)
var isOpen = false

for _ in ["a", "b", "c"] {
  wg.add(1)
  go {
    defer { wg.done() }
    m.withLock({
      while !isOpen {
        opened.wait()
      }
    })
    entered.add(1)
  }
}

m.withLock({
  isOpen = true
  opened.broadcast()
})
wg.wait()
println(entered.load())
```

```text
3
```

### A cond in a record

A record can keep the mutex, the cond, and the state they guard together. Here a counter lets a
goroutine wait until it reaches a value:

```ant
type Counter = { value: int, lock: Mutex, changed: Cond }

func Counter.new(): Counter {
  let lock = Mutex.new()
  Counter{ value: 0, lock, changed: Cond.new(lock) }
}

func Counter.increment(c: Counter) {
  c.lock.withLock({
    c.value += 1
    c.changed.broadcast()
  })
}

func Counter.waitFor(c: Counter, target: int) {
  c.lock.withLock({
    while c.value < target {
      c.changed.wait()
    }
  })
}

let hits = Counter.new()
for _ in ["a", "b", "c", "d"] {
  go { hits.increment() }
}
hits.waitFor(4)
println(hits.value)
```

```text
4
```

## Applications

### Bounded queue

Two conds over one mutex: `take` waits while the queue is empty, and `put` waits while it is full.
Each wakes the other side after it changes the queue:

```ant
type Queue = {
  items: Array[string],
  limit: int,
  lock: Mutex,
  notEmpty: Cond,
  notFull: Cond,
}

func Queue.new(limit: int): Queue {
  let lock = Mutex.new()
  Queue{ items: [], limit, lock, notEmpty: Cond.new(lock), notFull: Cond.new(lock) }
}

func Queue.put(q: Queue, item: string) {
  q.lock.withLock({
    while q.items.length() == q.limit {
      q.notFull.wait()
    }
    q.items.push(item)
    q.notEmpty.signal()
  })
}

func Queue.take(q: Queue): string {
  q.lock.withLock({
    while q.items.isEmpty() {
      q.notEmpty.wait()
    }
    let item = q.items.removeAt(0)
    q.notFull.signal()
    item
  })
}

let jobs = ["build", "test", "lint", "ship"]
let q = Queue.new(2)
go {
  for job in jobs {
    q.put(job)
  }
}
for _ in jobs {
  println(q.take())
}
```

```text
build
test
lint
ship
```

## Properties and relations

`wait` unlocks the mutex while it waits: the waiter starts the other goroutine with the mutex
locked, and that goroutine can lock it only once the waiter waits. When `wait` returns, the waiter
holds the mutex again, so it can unlock it:

```ant
let m = Mutex.new()
let c = Cond.new(m)
var turn = "waiter"

m.lock()
turn = "signaller"
go {
  m.withLock({
    assert(turn == "signaller")
    turn = "waiter again"
    c.signal()
  })
}
while turn == "signaller" {
  c.wait()
}
assert(turn == "waiter again")
m.unlock()
println(turn)
```

```text
waiter again
```

## Possible issues

### The mutex must be locked

`wait` panics when the mutex of the cond is not locked:

```ant
let m = Mutex.new()
let c = Cond.new(m)
c.wait() // panic: Cond.wait: the Mutex is not locked
```

### A signal is not remembered

A signal sent while no goroutine waits wakes nobody. A waiter that waits without checking the state
first can miss it and wait forever. Check the state in a `while` loop: when the signal came first,
the loop does not wait at all:

```ant
let m = Mutex.new()
let ready = Cond.new(m)
var done = false

m.withLock({
  done = true
  ready.signal()
})

m.withLock({
  while !done {
    ready.wait()
  }
})
println("no wait needed")
```

```text
no wait needed
```

### A goroutine that panics never signals

A goroutine that panics ends, and `wait` takes no context, so a waiter for its signal waits forever.
Set the state and wake the waiters in a `defer`, which runs even when the goroutine panics:

```ant
let m = Mutex.new()
let changed = Cond.new(m)
var finished = false

go {
  defer {
    m.withLock({
      finished = true
      changed.broadcast()
    })
  }
  panic("worker failed")
}

m.withLock({
  while !finished {
    changed.wait()
  }
})
println("worker gone")
```

```text
worker gone
```

The panic of the goroutine is reported on standard error, and the program goes on.

### Waiting with a deadline

`Cond.wait` cannot time out or be cancelled. To give up after a deadline, hand the event over on a
channel instead: [`Channel.receive`](../Channel/receive.md) takes a context and returns `null` when
the deadline passes first:

```ant
let ready: Channel[bool] = Channel.new(capacity: 1)

go {
  Time.sleep(1s)
  ready.send(true)
}

let got = Channel.receive(ready, ctx: Context.new(50ms))
println(if got == null { "gave up" } else { "ready" })
```

```text
gave up
```

## See also

- [Cond.new](new.md)
- [Cond.signal](signal.md)
- [Cond.broadcast](broadcast.md)
- [Mutex.withLock](../Mutex/withLock.md)
- [Mutex.lock](../Mutex/lock.md)
- [WaitGroup.wait](../WaitGroup/wait.md)
- [Channel.receive](../Channel/receive.md)
