<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Atomic.load

```ant signature
func Atomic.load[T](a: Atomic[T]): T
```

Atomic.load returns the value in a.

<!-- end generated -->

## Details

- A `load` that observes a `store` from another goroutine sees everything that goroutine wrote
  before the `store`.
- The value is returned as it is, not copied: a loaded record or array is the same value that the
  cell holds.
- A `load` followed by a `store` is two operations, and another goroutine can write between them.
  Use `Atomic.add` or `Atomic.compareAndSwap` to change a value based on the one before it.

## Basic examples

Read the value a cell was created with:

```ant
let hits = Atomic.new(0)
Atomic.load(hits) // 0
```

Read it again after a `store` and an `add`:

```ant
let hits = Atomic.new(0)
hits.store(5)
Atomic.load(hits) // 5
hits.add(2)       // 7
Atomic.load(hits) // 7
```

## Scope

A cell holds a value of any type, such as a string or a float:

```ant
let status = Atomic.new("starting")
status.store("ready")
status.load() // "ready"
let ratio = Atomic.new(0.5)
ratio.load() // 0.5
```

A nullable cell returns `null` until something is stored:

```ant
let latest: Atomic[string | null] = Atomic.new(null)
latest.load() // null
latest.store("v2")
latest.load() // "v2"
```

A cell of a union returns whichever member was stored last:

```ant
type Idle = { since: int }
type Busy = { job: string }
type State = Idle | Busy

let state: Atomic[State] = Atomic.new(Idle{ since: 0 })
state.load() // Idle{ since: 0 }
state.store(Busy{ job: "backup" })
state.load() // Busy{ job: "backup" }
```

A record comes back as the stored value itself, so in one goroutine a field set through one `load`
shows in the next:

```ant
type Point = { x: int, y: int }

let cell = Atomic.new(Point{ x: 1, y: 2 })
let p = cell.load()
p.x = 10
cell.load() // Point{ x: 10, y: 2 }
```

The cell can be the start of a pipe:

```ant
let count = Atomic.new(41)
count | Atomic.load() // 41
```

## Applications

Count events from many goroutines, then read the total after they finish:

```ant
let handled = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Iter.range(0, 100) {
  wg.add(1)
  go {
    defer { wg.done() }
    handled.add(1)
  }
}
wg.wait()
println(handled.load())
```

```text
100
```

Hand a result back from a goroutine. Load it into a local after the wait, and narrow the local:

```ant
let answer: Atomic[string | null] = Atomic.new(null)
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  answer.store("forty-two")
}
wg.wait()
let got = answer.load()
if got != null {
  println(got.toUpper())
}
```

```text
FORTY-TWO
```

Stop a worker with a flag. The worker loads the flag on every round and returns once it is set:

```ant
let stop = Atomic.new(false)
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  while !stop.load() {
    Time.sleep(1ms)
  }
  println("worker stopped")
}
Time.sleep(10ms)
stop.store(true)
wg.wait()
```

```text
worker stopped
```

Publish a new configuration by storing a whole new record. Each reader loads the current one once
and uses that snapshot, so it never sees half of one version and half of another:

```ant
type Config = { host: string, port: int }

let config = Atomic.new(Config{ host: "localhost", port: 8080 })

func address(): string {
  let c = config.load()
  "${c.host}:${c.port}"
}

println(address())
config.store(Config{ host: "example.com", port: 443 })
println(address())
```

```text
localhost:8080
example.com:443
```

## Properties and relations

A `load` returns what the last `store` put in the cell, and loading changes nothing:

```ant
let cell = Atomic.new("a")
cell.store("b")
assert(cell.load() == "b")
assert(cell.load() == cell.load())
```

`Atomic.add` returns the new value, the same value a `load` right after it returns:

```ant
let n = Atomic.new(10)
let after = n.add(5)
assert(after == n.load())
```

A `compareAndSwap` that expects the loaded value succeeds when nothing wrote in between:

```ant
let n = Atomic.new(3)
let seen = n.load()
assert(n.compareAndSwap(seen, seen * 2))
assert(n.load() == 6)
```

## Possible issues

The checker does not narrow a `load` call, since the next call can return something else:

```ant
let name: Atomic[string | null] = Atomic.new("ant")
if name.load() != null {
  println(name.load().length())
  // error: string | null has no field or function length on every member: narrow the value first
}
```

Load the value into a local once, and narrow the local:

```ant
let name: Atomic[string | null] = Atomic.new("ant")
let current = name.load()
if current != null {
  println(current.length())
}
```

```text
3
```

A `load` and a `store` of its result do not add up to one atomic update. Here two goroutines' steps
are written out in the order they can interleave: both load 0, both store 1, and one increment is
lost:

```ant
let count = Atomic.new(0)
let a = count.load()
let b = count.load()
count.store(a + 1)
count.store(b + 1)
count.load() // 1
```

`Atomic.add` does the read and the write as one operation, so no increment is lost:

```ant
let count = Atomic.new(0)
count.add(1) // 1
count.add(1) // 2
count.load() // 2
```

For an update other than addition, retry `compareAndSwap` until no other write came between the
`load` and the swap:

```ant
let longest = Atomic.new("")

func offer(word: string) {
  while true {
    let current = longest.load()
    if word.length() <= current.length() { return }
    if longest.compareAndSwap(current, word) { return }
  }
}

let wg = WaitGroup.new()
for word in ["ant", "spider", "bee"] {
  wg.add(1)
  go {
    defer { wg.done() }
    offer(word)
  }
}
wg.wait()
println(longest.load())
```

```text
spider
```

A field write on a loaded record is not atomic. The `Atomic` orders only the cell, so another
goroutine that loads the same record can read it while it changes, which is a data race:

```ant
type Config = { host: string, port: int }

let config = Atomic.new(Config{ host: "localhost", port: 8080 })
let c = config.load()
c.port = 443 // races with any goroutine that loads and reads config
```

Store a new record instead, built from the loaded one:

```ant
type Config = { host: string, port: int }

let config = Atomic.new(Config{ host: "localhost", port: 8080 })
let c = config.load()
config.store(Config{ ...c, port: 443 })
config.load() // Config{ host: "localhost", port: 443 }
```

## See also

[Atomic.store](store.md) · [Atomic.add](add.md) · [Atomic.compareAndSwap](compareAndSwap.md) ·
[Atomic.new](new.md) · [Atomic](index.md) · [Mutex.withLock](../Mutex/withLock.md) ·
[WaitGroup.wait](../WaitGroup/wait.md) · [Once.value](../Once/value.md)
