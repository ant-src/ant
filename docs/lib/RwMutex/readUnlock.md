<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.readUnlock

```ant signature
func RwMutex.readUnlock(m: RwMutex): void
```

RwMutex.readUnlock releases one read lock on m. It panics if m is not read-locked.

<!-- end generated -->

## Details

- Read locks count: every [RwMutex.readLock](readLock.md) needs its own `readUnlock`, and a writer
  waits until the last reader has released m.
- It panics when no reader holds m, which includes when m is locked for writing; release a write
  lock with [RwMutex.writeUnlock](writeUnlock.md).
- Call it in a `defer` right after the read lock, so that a `return` or a panic still releases m.
  [RwMutex.withRead](withRead.md) does both in one call.

## Basic examples

Lock for reading, read, then release the lock:

```ant
let m = RwMutex.new()
let scores = Map{ "ann": 3, "bob": 5 }
RwMutex.readLock(m)
let ann = scores["ann"]
RwMutex.readUnlock(m)
println(ann)
```

```text
3
```

Release the lock in a `defer` right after taking it, so every way out of the function releases it:

```ant
type Scores = { lock: RwMutex, byName: Map[string, int] }

func Scores.get(s: Scores, name: string): int {
  s.lock.readLock()
  defer { s.lock.readUnlock() }
  s.byName[name] ?? 0
}

let scores = Scores{ lock: RwMutex.new(), byName: Map{ "ann": 3 } }
println(scores.get("ann"))
println(scores.get("cy"))
```

```text
3
0
```

## Scope

### Several read locks

Each read lock is released on its own. While one remains, m is still read-locked:

```ant
let m = RwMutex.new()
m.readLock()
m.readLock()
m.readUnlock()
m.readUnlock()
m.writeLock()
println("no readers left")
m.writeUnlock()
```

```text
no readers left
```

### A waiting writer

A writer gets m only after the reader releases its lock:

```ant
let m = RwMutex.new()
let done: Channel[bool] = Channel.new()
m.readLock()
go {
  m.writeLock()
  println("writer has the lock")
  m.writeUnlock()
  Channel.send(done, true)
}
println("reader releases the lock")
m.readUnlock()
Channel.receive(done)
```

```text
reader releases the lock
writer has the lock
```

### From many goroutines

Readers run side by side, and each releases its own lock:

```ant
let m = RwMutex.new()
let prices = Map{ "tea": 4, "cake": 6 }
let items = ["tea", "cake", "tea", "jam"]
let found = Parallel.map(items, func(item: string, _: Context): int {
  m.readLock()
  defer { m.readUnlock() }
  prices[item] ?? 0
}, limit: 4)
println(found)
```

```text
[4, 6, 4, 0]
```

## Applications

### A read-mostly cache

Lookups share the lock for reading; a store takes it for writing. `Cache.find` releases its read
lock before `Cache.store` takes the write lock:

```ant
type Cache = { lock: RwMutex, entries: Map[string, string] }

func Cache.find(c: Cache, key: string): string | null {
  c.lock.readLock()
  defer { c.lock.readUnlock() }
  c.entries[key]
}

func Cache.store(c: Cache, key: string, value: string) {
  c.lock.writeLock()
  defer { c.lock.writeUnlock() }
  c.entries[key] = value
}

func Cache.load(c: Cache, key: string, compute: func(string): string): string {
  let cached = c.find(key)
  if cached != null { return cached }
  let value = compute(key)
  c.store(key, value)
  value
}

let cache = Cache{ lock: RwMutex.new(), entries: Map{} }
println(cache.load("ant", { x.toUpper() }))
println(cache.find("ant"))
```

```text
ANT
ANT
```

## Properties and relations

A read lock released in a `defer` is what [RwMutex.withRead](withRead.md) does:

```ant
let m = RwMutex.new()
let xs = [1, 2, 3]

func countLocked(): int {
  m.readLock()
  defer { m.readUnlock() }
  xs.length()
}

assert(countLocked() == m.withRead({ xs.length() }))
```

After as many unlocks as locks, m is unlocked, and one more unlock panics:

```ant
let m = RwMutex.new()
for _ in Iter.range(0, 3) {
  m.readLock()
}
for _ in Iter.range(0, 3) {
  m.readUnlock()
}
assert(Panic.catch({ m.readUnlock() }) is Err)
```

## Possible issues

### Unlocking a lock you do not hold

A `readUnlock` with no read lock to release panics:

```ant
let m = RwMutex.new()
m.readUnlock() // panic: RwMutex.readUnlock: the RwMutex is not read-locked
```

### Releasing a write lock

`readUnlock` does not release a write lock; [RwMutex.writeUnlock](writeUnlock.md) does:

```ant
let m = RwMutex.new()
m.writeLock()
m.readUnlock() // panic: RwMutex.readUnlock: the RwMutex is not read-locked
```

### Taking the write lock before the read unlock

A write lock waits for every reader, including the goroutine that asks for it. Release the read lock
with `readUnlock` before you call [RwMutex.writeLock](writeLock.md):

```ant
let m = RwMutex.new()
let done: Channel[bool] = Channel.new()
go {
  m.readLock()
  m.writeLock()
  m.writeUnlock()
  m.readUnlock()
  Channel.send(done, true)
}
println(Channel.receive(done, Context.new(timeout: 100ms)))
```

```text
null
```

### A path that skips the unlock

A `return` before a `readUnlock` that is not deferred leaves m read-locked, and every writer waits
for it forever. Here the writer is still waiting when a bounded receive gives up:

```ant
let m = RwMutex.new()
let names = ["ann"]

func has(name: string): bool {
  m.readLock()
  if names.contains(name) { return true }
  m.readUnlock()
  false
}

println(has("ann"))
let written: Channel[bool] = Channel.new()
go {
  m.writeLock()
  Channel.send(written, true)
}
println(Channel.receive(written, Context.new(timeout: 100ms)))
```

```text
true
null
```

A panic skips it in the same way. Put the unlock in a `defer`, whose block runs on a `return` and on
a panic alike:

```ant
let m = RwMutex.new()

func risky() {
  m.readLock()
  defer { m.readUnlock() }
  panic("bad input")
}

println(Panic.catch(risky) is Err)
m.writeLock()
println("the read lock was released")
m.writeUnlock()
```

```text
true
the read lock was released
```

## See also

[RwMutex.readLock](readLock.md) · [RwMutex.withRead](withRead.md) ·
[RwMutex.writeUnlock](writeUnlock.md) · [RwMutex.writeLock](writeLock.md) · [RwMutex.new](new.md) ·
[Mutex.unlock](../Mutex/unlock.md) · [RwMutex](index.md)
