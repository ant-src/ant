<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.new

```ant signature
func RwMutex.new(): RwMutex
```

RwMutex.new returns a new, unlocked reader-writer lock.

<!-- end generated -->

## Details

- Each call makes a distinct lock: locks compare by identity, and a copy of a record shares the lock
  in its field.
- The new lock is held by no one, so the first `readLock` or `writeLock` returns at once, and an
  unlock before any lock panics.

## Basic examples

Make a lock and hold it for reading while you read shared data:

```ant
let lock = RwMutex.new()
let names = ["ann", "bob"]
let count = lock.withRead({ names.length() })
println(count)
```

```text
2
```

Hold it for writing while you change the data:

```ant
let lock = RwMutex.new()
let scores: Map[string, int] = Map{}
lock.withWrite({ scores["ann"] = 3 })
println(scores)
```

```text
Map{ "ann": 3 }
```

## Scope

A lock renders as its type's name:

```ant
RwMutex.new() // RwMutex
```

A record field can default to `RwMutex.new()`. The default runs again for each record, so each
record gets its own lock:

```ant
type Config = { lock: RwMutex = RwMutex.new(), values: Map[string, string] }
let a = Config{ values: Map{} }
let b = Config{ values: Map{} }
a.lock == b.lock // false
```

Keep the lock in a record next to the data it guards, and lock it with the explicit pair when a
function returns from the middle. `defer` releases the lock on every path:

```ant
type Config = { lock: RwMutex = RwMutex.new(), values: Map[string, string] }

func lookup(config: Config, key: string): string {
  config.lock.readLock()
  defer { config.lock.readUnlock() }
  let value = config.values[key]
  if value == null {
    return "unset"
  }
  value.toUpper()
}

let config = Config{ values: Map{ "mode": "dev" } }
println(lookup(config, "mode"))
println(lookup(config, "port"))
```

```text
DEV
unset
```

A lock captured by a closure is the same lock in every goroutine that runs the closure:

```ant
let lock = RwMutex.new()
let visits: Map[string, int] = Map{}

let pages = ["/", "/docs", "/", "/blog", "/"]
Parallel.each(pages, func(page: string, ctx: Context) {
  lock.withWrite({ visits[page] = (visits[page] ?? 0) + 1 })
}, limit: 4)
let home = lock.withRead({ visits["/"] ?? 0 })
println(home)
```

```text
3
```

## Applications

Guard a read-mostly cache. Many goroutines read at once under the read lock; a writer takes the
write lock only to add an entry:

```ant
type Cache = { lock: RwMutex = RwMutex.new(), entries: Map[string, int] }

func get(cache: Cache, key: string): int {
  let hit = cache.lock.withRead({ cache.entries[key] })
  if hit != null {
    return hit
  }
  let value = key.length()
  cache.lock.withWrite({ cache.entries[key] = value })
  value
}

let cache = Cache{ entries: Map{} }
let lengths = Parallel.map(["ant", "bee", "ant", "wasp"], func(word: string, ctx: Context) {
  get(cache, word)
}, limit: 4)
println(lengths)
println(cache.entries.size())
```

```text
[3, 3, 3, 4]
3
```

## Properties and relations

Two calls make two locks, and a lock equals only itself:

```ant
let a = RwMutex.new()
let b = RwMutex.new()
assert(a != b)
assert(a == a)
```

A new lock takes any number of read locks, and then a write lock once they are all released:

```ant
let lock = RwMutex.new()
lock.readLock()
lock.readLock()
lock.readUnlock()
lock.readUnlock()
lock.writeLock()
lock.writeUnlock()
```

Copying a record copies the reference to its lock, not the lock:

```ant
type Guarded = { lock: RwMutex, value: int }

let first = Guarded{ lock: RwMutex.new(), value: 1 }
let second = Guarded{ ...first, value: 2 }
assert(first.lock == second.lock)
```

## Possible issues

A new lock is not locked, so releasing it panics:

```ant
let lock = RwMutex.new()
lock.readUnlock() // panic: RwMutex.readUnlock: readUnlock of an RwMutex that is not read-locked
```

```ant
let lock = RwMutex.new()
lock.writeUnlock()
// panic: RwMutex.writeUnlock: writeUnlock of an RwMutex that is not write-locked
```

A lock made inside the function it guards is a new lock on every call, so it keeps no two calls
apart:

```ant
func guard(): RwMutex {
  RwMutex.new()
}

assert(guard() != guard())
```

Make the lock once, outside the function, next to the data. One shared lock serializes every call:

```ant
let lock = RwMutex.new()
var total = 0

func add(n: int): void {
  lock.withWrite({ total += n })
}

Parallel.each([1, 2, 3, 4], func(n: int, ctx: Context) { add(n) }, limit: 4)
println(total)
```

```text
10
```

## See also

[RwMutex.withRead](withRead.md) · [RwMutex.withWrite](withWrite.md) ·
[RwMutex.readLock](readLock.md) · [RwMutex.writeLock](writeLock.md) · [Mutex.new](../Mutex/new.md) ·
[Atomic.new](../Atomic/new.md)
