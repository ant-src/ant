<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Mutex

```ant signature
type Mutex
```

Mutex is a mutual-exclusion lock: while it is locked, no other lock of it succeeds. Any goroutine may unlock it. Mutexes compare by identity.

<!-- end generated -->

A mutex guards data that several goroutines share. A goroutine locks the mutex before it touches the
data and unlocks it afterwards, so only one goroutine works on the data at a time. The `Mutex`
namespace holds the functions that make, lock, and unlock a mutex; `Mutex` is also the type, for an
annotation or a record field.

## Details

- A new mutex is unlocked. [`Mutex.lock`](lock.md) waits while the mutex is locked and takes no
  context, so it cannot give up.
- [`Mutex.unlock`](unlock.md) panics when the mutex is not locked.
- Everything written before an unlock is visible to the next goroutine that locks the same mutex.
  The mutex orders only the code that takes it, so reads of the guarded data need the lock too.

## Basic examples

`Mutex.withLock` locks the mutex, runs the block, and unlocks it. Eight goroutines at a time add to
one counter, and no increment is lost:

```ant
let m = Mutex.new()
var count = 0
Parallel.each(Array.range(0, 100), func(i: int, ctx: Context) {
  Mutex.withLock(m, { count += 1 })
}, limit: 8)
println(count)
```

```text
100
```

`Mutex.lock` and `Mutex.unlock` take and release the lock by hand. A `defer` right after the lock
releases it on every way out of the block:

```ant
let m = Mutex.new()
let names: Array[string] = []
Parallel.each(["Ann", "Bob", "Cy"], func(name: string, ctx: Context) {
  Mutex.lock(m)
  defer { Mutex.unlock(m) }
  names.push(name)
}, limit: 3)
println(names.sort())
```

```text
["Ann", "Bob", "Cy"]
```

## Scope

A record keeps the mutex next to the data it guards, and the functions on the record take the lock:

```ant
type Tally = { lock: Mutex, votes: Map[string, int] }

func Tally.vote(t: Tally, name: string) {
  t.lock.withLock({ t.votes[name] = (t.votes[name] ?? 0) + 1 })
}

func Tally.count(t: Tally, name: string): int {
  t.lock.withLock({ t.votes[name] ?? 0 })
}

let tally = Tally{ lock: Mutex.new(), votes: Map{} }
Parallel.each(["ann", "bob", "ann"], func(name: string, ctx: Context) { tally.vote(name) }, limit: 3)
println(tally.count("ann"))
println(tally.count("cy"))
```

```text
2
0
```

A field default makes a new mutex for each record:

```ant
type Account = { lock: Mutex = Mutex.new(), balance: int = 0 }

let a = Account{}
let b = Account{}
println(a.lock == b.lock)
```

```text
false
```

Mutexes compare by identity, and print as their type's name:

```ant
let a = Mutex.new()
let b = Mutex.new()
let alias = a
println(a == b)
println(a == alias)
println(a)
```

```text
false
true
Mutex
```

## Functions by job

### Making a mutex

- [Mutex.new](new.md): a new, unlocked mutex

### Locking

- [Mutex.withLock](withLock.md): lock, run a function, unlock even if it panics, and return its
  result
- [Mutex.lock](lock.md): lock, waiting while the mutex is locked
- [Mutex.unlock](unlock.md): unlock; panics if the mutex is not locked

## Applications

### Read a snapshot under the lock

`withLock` returns the block's value, so a `let` takes a snapshot under the lock and the code after
it works with the snapshot, not the shared map:

```ant
let m = Mutex.new()
let scores: Map[string, int] = Map{ ann: 7, bob: 9 }

func report(): string {
  let best = m.withLock({ scores.entries().maxBy({ x.value }) })
  if best == null { return "no scores" }
  "${best.key} leads with ${best.value}"
}

println(report())
```

```text
bob leads with 9
```

### Wait for a condition

A [`Cond`](../Cond/index.md) over the mutex lets a goroutine wait, holding no lock, until another
goroutine changes the guarded data:

```ant
let m = Mutex.new()
let changed = Cond.new(m)
let queue: Array[string] = []

go {
  for job in ["build", "test"] {
    m.withLock({
      queue.push(job)
      changed.signal()
    })
  }
}

for _ in Array.range(0, 2) {
  let job = m.withLock({
    while queue.isEmpty() {
      changed.wait()
    }
    queue.removeAt(0)
  })
  println(job)
}
```

```text
build
test
```

## Properties and relations

A change made under the lock is never seen half done by another goroutine that takes the same lock.
Each transfer keeps the total at 200, and every check under the lock sees that:

```ant
let m = Mutex.new()
let balances = Map{ ann: 100, bob: 100 }
Parallel.each(Array.range(0, 40), func(i: int, ctx: Context) {
  m.withLock({
    balances["ann"] = (balances["ann"] ?? 0) - 1
    balances["bob"] = (balances["bob"] ?? 0) + 1
  })
  assert(m.withLock({ balances.values().sum() }) == 200)
}, limit: 4)
assert(balances == Map{ ann: 60, bob: 140 })
```

`withLock` releases the mutex on every way out of the block, a panic included, so the mutex is free
again afterwards:

```ant
let m = Mutex.new()
let r = Panic.catch({ m.withLock({ panic("boom") }) })
assert(r is Err)
m.lock()
m.unlock()
```

## Possible issues

### Unlocking twice panics

A deferred unlock and an explicit one unlock the mutex twice. The deferred one runs second and
panics:

```ant
let m = Mutex.new()
m.lock()
defer { m.unlock() } // panic: Mutex.unlock: the Mutex is not locked
m.unlock()
```

Keep one of them: the `defer`, or `Mutex.withLock` in place of both.

### A panic can leave the guarded data half changed

`withLock` releases the lock when the block panics, but what the block wrote before the panic stays
written. The next goroutine that takes the lock sees the half-done transfer:

```ant
let m = Mutex.new()
let balances = Map{ ann: 100, bob: 100 }
let r = Panic.catch({
  m.withLock({
    balances["ann"] = (balances["ann"] ?? 0) - 1
    panic("boom")
  })
})
assert(r is Err)
println(m.withLock({ balances.values().sum() }))
```

```text
199
```

The code that catches the panic decides whether the data can still be used.

### An outer `var` does not narrow inside the block

The block is a lambda, and another goroutine may assign an outer `var` between a test and a use, so
the test does not narrow it there:

```ant
func find(id: int): string | null {
  if id == 1 { "Ann" } else { null }
}

let m = Mutex.new()
var current = find(1)
m.withLock({
  if current != null { println(current.length()) }
  // error: current is a var of an enclosing body, which does not narrow here, so bind it to a let first
})
```

Bind the value to a `let` first. An outer `let` keeps its narrowing inside the block, as in
[Read a snapshot under the lock](#read-a-snapshot-under-the-lock):

```ant
func find(id: int): string | null {
  if id == 1 { "Ann" } else { null }
}

let m = Mutex.new()
var current = find(1)
let snapshot = current
m.withLock({
  if snapshot != null { println(snapshot.length()) }
})
```

```text
3
```

## See also

- [RwMutex](../RwMutex/index.md): many readers or one writer at a time
- [Atomic](../Atomic/index.md): one value read and changed without a lock
- [Cond](../Cond/index.md): wait under a mutex until a condition holds
- [Once](../Once/index.md): run initialization once, however many goroutines ask
- [WaitGroup](../WaitGroup/index.md): wait for goroutines to finish
- [Channel](../Channel/index.md): pass values between goroutines instead of sharing them
- [Parallel](../Parallel/index.md): run callbacks on goroutines and wait for them
- [Concurrency](../../guide/concurrency.md): how goroutines, channels, and sync objects fit together
