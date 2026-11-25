<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex

```ant signature
type RwMutex
```

RwMutex is a reader-writer lock: many goroutines may hold it for reading, or one for writing. Any
goroutine may release it. RwMutexes compare by identity.

<!-- end generated -->

An `RwMutex` guards shared data that many goroutines read and few change. Readers share the lock, so
they run side by side; a writer holds it alone. The `RwMutex` namespace holds a function that makes
a lock, a pair that lock and unlock it for reading, a pair for writing, and two that run a function
under a read or a write lock.

## Details

- While a writer holds the lock, readers and other writers wait. While any reader holds it, a writer
  waits until the last one releases it.
- A new lock is unlocked. `readLock` and `writeLock` take no context, so a goroutine waiting for the
  lock cannot give up.
- An unlock of a lock that is not held, or of the other kind, panics.
- A read lock does not turn into a write lock. A goroutine that asks for the write lock while it
  holds a read lock waits for itself, forever.
- Changes made under the write lock are visible to every later reader and writer. A read lock only
  keeps writers out, so a change made under it races with the other readers.
- An `RwMutex` is a reference value: a copy in a record field, a parameter, or a closure is the same
  lock.

## Basic examples

`RwMutex.new` makes a lock. `RwMutex.withRead` runs a function under the read lock and returns its
result:

```ant
let lock = RwMutex.new()
let prices = Map{ "tea": 3, "cake": 5 }
println(RwMutex.withRead(lock, { prices["tea"] }))
```

```text
3
```

`RwMutex.withWrite` runs a function under the write lock, for a change:

```ant
let lock = RwMutex.new()
let prices = Map{ "tea": 3 }
RwMutex.withWrite(lock, { prices["cake"] = 5 })
println(prices)
```

```text
Map{ "tea": 3, "cake": 5 }
```

`RwMutex.readLock` and `RwMutex.readUnlock` take and release the read lock by hand. Release it in a
`defer` right after taking it:

```ant
let lock = RwMutex.new()
let names = ["Ann", "Bob"]

func count(): int {
  RwMutex.readLock(lock)
  defer { RwMutex.readUnlock(lock) }
  names.length()
}

println(count())
```

```text
2
```

## Scope

A lock in a record field guards the record's other fields. A field default gives every record a
fresh lock:

```ant
type Counter = { lock: RwMutex = RwMutex.new(), hits: Map[string, int] }

func Counter.hit(c: Counter, page: string) {
  c.lock.withWrite({ c.hits[page] = (c.hits[page] ?? 0) + 1 })
}

func Counter.get(c: Counter, page: string): int {
  c.lock.withRead({ c.hits[page] ?? 0 })
}

let counter = Counter{ hits: Map{} }
counter.hit("/")
counter.hit("/")
counter.hit("/about")
println([counter.get("/"), counter.get("/about"), counter.get("/docs")])
```

```text
[2, 1, 0]
```

Locks compare by identity, and print as their type's name:

```ant
let a = RwMutex.new()
let b = RwMutex.new()
let alias = a
println(a == b)
println(a == alias)
println(a)
```

```text
false
true
RwMutex
```

## Functions by job

### Making a lock

- [RwMutex.new](new.md): a new, unlocked reader-writer lock

### Reading

- [RwMutex.withRead](withRead.md): run a function under the read lock and return its result
- [RwMutex.readLock](readLock.md): lock for reading, waiting while a writer holds the lock
- [RwMutex.readUnlock](readUnlock.md): release one read lock; panics if the lock is not read-locked

### Writing

- [RwMutex.withWrite](withWrite.md): run a function under the write lock and return its result
- [RwMutex.writeLock](writeLock.md): lock for writing, waiting while any reader or writer holds it
- [RwMutex.writeUnlock](writeUnlock.md): release the write lock; panics if it is not write-locked

## Applications

### A read-mostly settings store

Many goroutines read the settings at once, and an update takes the write lock for the moment it
changes them:

```ant
type Settings = { lock: RwMutex = RwMutex.new(), values: Map[string, string] }

func Settings.get(s: Settings, key: string): string {
  s.lock.withRead({ s.values[key] ?? "unset" })
}

func Settings.set(s: Settings, key: string, value: string) {
  s.lock.withWrite({ s.values[key] = value })
}

let settings = Settings{ values: Map{ "theme": "dark" } }
settings.set("lang", "en")
let keys = ["theme", "lang", "font"]
let found = Parallel.map(keys, func(key: string, ctx: Context) { settings.get(key) }, limit: 3)
println(found)
```

```text
["dark", "en", "unset"]
```

### Read several fields together

A read under one lock sees the fields as one writer left them, never half of an update:

```ant
type Range = { lock: RwMutex = RwMutex.new(), low: int, high: int }

func Range.move(r: Range, by: int) {
  r.lock.withWrite({
    r.low += by
    r.high += by
  })
}

func Range.width(r: Range): int {
  r.lock.withRead({ r.high - r.low })
}

let r = Range{ low: 0, high: 10 }
Parallel.each(Array.filled(50, 1), func(by: int, ctx: Context) {
  r.move(by)
  assert(r.width() == 10)
}, limit: 8)
println([r.low, r.high])
```

```text
[50, 60]
```

## Properties and relations

Readers share the lock: a second reader enters while the first still holds it:

```ant
let lock = RwMutex.new()
let entered: Channel[string] = Channel.new()
let message = lock.withRead({
  go { lock.withRead({ entered.send("second reader is in") }) }
  entered.receive()
})
println(message)
```

```text
second reader is in
```

A writer holds it alone: a reader waits until the write lock is released, and then sees the change:

```ant
let lock = RwMutex.new()
let names = ["Ann"]
let seen: Channel[int] = Channel.new()
lock.writeLock()
go { seen.send(lock.withRead({ names.length() })) }
names.push("Bob")
lock.writeUnlock()
println(seen.receive())
```

```text
2
```

`withRead` and `withWrite` are a lock, a `defer`red unlock, and a call, and they release the lock
even after a panic:

```ant
let lock = RwMutex.new()
let xs = [1, 2, 3]

func byHand(): int {
  lock.readLock()
  defer { lock.readUnlock() }
  xs.length()
}

assert(byHand() == lock.withRead({ xs.length() }))
assert(Panic.catch({ lock.withWrite({ panic("bad write") }) }) is Err)
lock.writeLock()
lock.writeUnlock()
```

## Possible issues

### A read lock does not upgrade

Inside `withRead`, a write lock waits for every reader to leave, its own goroutine included, so it
waits forever. Here a bounded receive gives up on it:

```ant
let lock = RwMutex.new()
let upgraded: Channel[bool] = Channel.new()
go {
  lock.withRead({
    // Blocks for good: this goroutine still holds the read lock.
    lock.writeLock()
    upgraded.send(true)
  })
}
println(upgraded.receive(ctx: Context.new(100ms)))
```

```text
null
```

Release the read lock first, then take the write lock, and check again under it, since another
writer may have run in between:

```ant
let lock = RwMutex.new()
let cache = Map{ "a": 1 }

func ensure(key: string, value: int) {
  if lock.withRead({ cache.has(key) }) { return }
  lock.withWrite({
    if !cache.has(key) { cache[key] = value }
  })
}

ensure("a", 10)
ensure("b", 2)
println(cache)
```

```text
Map{ "a": 1, "b": 2 }
```

### A change under the read lock races

Readers do not keep each other out, so increments made under the read lock can be lost, and the
count can end below 100:

<!-- test: check-only the lost increments depend on goroutine order -->

```ant
let lock = RwMutex.new()
var hits = 0
Parallel.each(Array.filled(100, 1), func(n: int, ctx: Context) { lock.withRead({ hits += n }) }, limit: 8)
println(hits)
```

Change shared data under the write lock:

```ant
let lock = RwMutex.new()
var hits = 0
Parallel.each(Array.filled(100, 1), func(n: int, ctx: Context) { lock.withWrite({ hits += n }) }, limit: 8)
println(hits)
```

```text
100
```

### Unlocking the wrong kind

Each unlock releases only its own kind of lock. A `writeUnlock` of a read lock panics:

```ant
let lock = RwMutex.new()
lock.readLock()
lock.writeUnlock() // panic: RwMutex.writeUnlock: the RwMutex is not write-locked
```

## See also

- [Mutex](../Mutex/index.md): one lock for readers and writers alike
- [Atomic](../Atomic/index.md): one value read and written atomically
- [Once](../Once/index.md): set up shared state exactly once
- [Parallel](../Parallel/index.md): run work on goroutines and wait for it
- [Concurrency](../../guide/concurrency.md): how goroutines, channels, and sync objects fit together
