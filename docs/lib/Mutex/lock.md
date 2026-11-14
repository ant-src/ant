<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Mutex.lock

```ant signature
func Mutex.lock(m: Mutex): void
```

Mutex.lock locks m, blocking while m is locked, by any goroutine, the caller included: a mutex is not reentrant. It takes no context.

<!-- end generated -->

## Details

- `lock` blocks while `m` is locked, whoever locked it, so a second `lock` in the same goroutine
  never returns.
- Everything written before the previous `Mutex.unlock` of `m` is visible after `lock` returns.
- Follow a `lock` with `defer { m.unlock() }`, or use `Mutex.withLock`, so a `return` or a panic
  releases `m`.

## Basic examples

Lock, change the shared map, unlock:

```ant
let m = Mutex.new()
let ages: Map[string, int] = Map{}
Mutex.lock(m)
ages["Ann"] = 31
m.unlock()
println(ages)
```

```text
Map{ "Ann": 31 }
```

Several goroutines push to one array. Each takes the lock before it touches the array, and the
deferred unlock releases it:

```ant
let m = Mutex.new()
let squares: Array[int] = []
let wg = WaitGroup.new()
for n in [1, 2, 3, 4] {
  wg.add(1)
  go {
    defer { wg.done() }
    Mutex.lock(m)
    defer { m.unlock() }
    squares.push(n * n)
  }
}
wg.wait()
println(squares.sort())
```

```text
[1, 4, 9, 16]
```

## Scope

`lock` waits while another goroutine holds the mutex. The worker takes the lock and reports it on a
channel, so the main goroutine's `lock` returns only after the worker has printed and unlocked:

```ant
let m = Mutex.new()
let held: Channel[bool] = Channel.new()
go {
  m.lock()
  defer { m.unlock() }
  held.send(true)
  println("worker unlocks")
}
held.receive()
m.lock()
println("main locked")
m.unlock()
```

```text
worker unlocks
main locked
```

Reads need the lock too, not only writes. Read the guarded value into a fresh local under the lock;
the local is a snapshot, so it narrows like any other:

```ant
type Latest = { lock: Mutex, value: string | null }

func Latest.set(l: Latest, value: string) {
  l.lock.lock()
  defer { l.lock.unlock() }
  l.value = value
}

func Latest.describe(l: Latest): string {
  l.lock.lock()
  let value = l.value
  l.lock.unlock()
  if value == null { return "nothing yet" }
  "latest: ${value.toUpper()}"
}

let latest = Latest{ lock: Mutex.new(), value: null }
println(latest.describe())
latest.set("ant")
println(latest.describe())
```

```text
nothing yet
latest: ANT
```

## Applications

Count words with `Parallel.each`. The callbacks run at the same time, so each one locks the shared
map before it updates it:

```ant
let words = ["ant", "bee", "ant", "cat", "ant", "bee"]
let m = Mutex.new()
let counts: Map[string, int] = Map{}
Parallel.each(words, func(word: string, work: Context) {
  m.lock()
  defer { m.unlock() }
  counts[word] = (counts.get(word) ?? 0) + 1
}, limit: 3)
println(counts["ant"])
println(counts["bee"])
println(counts["cat"])
```

```text
3
2
1
```

A transfer between two accounts holds both locks. Every transfer takes them in the same order, by
account id, so two transfers in opposite directions cannot each hold one lock and wait for the
other:

```ant
type Account = { id: int, lock: Mutex, balance: int }

func transfer(from: Account, to: Account, amount: int) {
  let first = if from.id < to.id { from } else { to }
  let second = if from.id < to.id { to } else { from }
  first.lock.lock()
  defer { first.lock.unlock() }
  second.lock.lock()
  defer { second.lock.unlock() }
  from.balance -= amount
  to.balance += amount
}

let a = Account{ id: 1, lock: Mutex.new(), balance: 100 }
let b = Account{ id: 2, lock: Mutex.new(), balance: 100 }
let wg = WaitGroup.new()
for _ in Array.range(0, 50) {
  wg.add(2)
  go {
    defer { wg.done() }
    transfer(a, b, 3)
  }
  go {
    defer { wg.done() }
    transfer(b, a, 1)
  }
}
wg.wait()
println("${a.balance} ${b.balance}")
```

```text
0 200
```

## Properties and relations

A write made under the lock is visible to the next goroutine that locks the mutex. After the
`WaitGroup` wait, every increment is counted:

```ant
type Counter = { lock: Mutex, n: int }

let counter = Counter{ lock: Mutex.new(), n: 0 }
let wg = WaitGroup.new()
for _ in Array.range(0, 100) {
  wg.add(1)
  go {
    defer { wg.done() }
    counter.lock.lock()
    defer { counter.lock.unlock() }
    counter.n += 1
  }
}
wg.wait()
counter.lock.lock()
assert(counter.n == 100)
counter.lock.unlock()
```

`lock` followed by a deferred `unlock` does what `Mutex.withLock` does: the work runs under the
lock, and the mutex is free again afterwards:

```ant
let m = Mutex.new()
let xs = [1, 2]

func pushLocked(x: int) {
  m.lock()
  defer { m.unlock() }
  xs.push(x)
}

pushLocked(3)
Mutex.withLock(m, { xs.push(4) })
assert(xs == [1, 2, 3, 4])
m.lock()
m.unlock()
```

## Possible issues

A function that locks must not call another function that locks the same mutex: the second `lock`
blocks while the mutex is locked, the mutex stays locked until the first function returns, and the
program hangs. Split the work into a function that expects the lock held and a public one that takes
it:

```ant
type Cart = { lock: Mutex, items: Array[string] }

// addLocked expects c.lock to be held by the caller.
func addLocked(c: Cart, item: string) {
  c.items.push(item)
}

func Cart.add(c: Cart, item: string) {
  c.lock.lock()
  defer { c.lock.unlock() }
  addLocked(c, item)
}

func Cart.addPair(c: Cart, a: string, b: string) {
  c.lock.lock()
  defer { c.lock.unlock() }
  // Calling c.add(a) here would lock c.lock a second time, and that lock never returns.
  addLocked(c, a)
  addLocked(c, b)
}

let cart = Cart{ lock: Mutex.new(), items: [] }
cart.add("milk")
cart.addPair("bread", "eggs")
println(cart.items)
```

```text
["milk", "bread", "eggs"]
```

`lock` cannot give up. To wait for a lock with a deadline, use a channel of capacity 1 as the lock:
a `Channel.send` with a `ctx` takes it, and returns `false` when the context ends first:

```ant
let slot: Channel[bool] = Channel.new(1)
slot.send(true)

let ctx = Context.new(timeout: 20ms)
if slot.send(true, ctx: ctx) {
  println("got the lock")
  slot.receive()
} else {
  println("timed out")
}
```

```text
timed out
```

## See also

[Mutex.unlock](unlock.md) · [Mutex.withLock](withLock.md) · [Mutex.new](new.md) · [Mutex](index.md)
· [RwMutex.writeLock](../RwMutex/writeLock.md) · [RwMutex.readLock](../RwMutex/readLock.md) ·
[Cond.wait](../Cond/wait.md) · [Parallel.each](../Parallel/each.md) ·
[defer](../../language/keywords/defer.md) · [Concurrency guide](../../guide/concurrency.md)
