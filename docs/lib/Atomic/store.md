<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Atomic.store

```ant signature
func Atomic.store[T](a: Atomic[T], v: T): void
```

Atomic.store sets the value in a to v.

<!-- end generated -->

## Details

- To change the value only when it still holds an expected value, use `Atomic.compareAndSwap`; to
  add to a number, use `Atomic.add`.
- Everything a goroutine writes before a store is visible to a goroutine whose `load` sees the
  stored value.
- A record, array, map, set, or `Bytes` is stored by reference. The cell holds the value you passed,
  so a later change to that value shows in every `load`, and the change is not atomic.

## Basic examples

Store a new value and load it back:

```ant
let a = Atomic.new(0)
Atomic.store(a, 5)
a.load() // 5
```

A goroutine stores its result, and the main goroutine loads it after the wait:

```ant
let answer: Atomic[int | null] = Atomic.new(null)
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  Atomic.store(answer, 6 * 7)
}
wg.wait()
let v = answer.load()
if v != null { println(v + 1) }
```

```text
43
```

## Scope

A cell of a nullable type can store a value and `null` again:

```ant
let name: Atomic[string | null] = Atomic.new(null)
name.store("Ann")
name.load() // "Ann"
name.store(null)
name.load() // null
```

A cell can hold a union; each store picks any of its members:

```ant
type Idle = {}
type Running = { since: int }
type State = Idle | Running
let state: Atomic[State] = Atomic.new(Idle{})
state.store(Running{ since: 3 })
state.load() // Running{ since: 3 }
state.store(Idle{})
state.load() // Idle{}
```

An atomic is a reference value. Two names of one cell see each other's stores:

```ant
let a = Atomic.new(1)
let b = a
b.store(2)
a.load() // 2
```

## Applications

Signal a worker to stop. The worker checks the flag between rounds, and the main goroutine stores
`true` when it is time to finish:

```ant
let stop = Atomic.new(false)
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  while !stop.load() {
    Time.sleep(1ms)
  }
}
Time.sleep(10ms)
stop.store(true)
wg.wait()
println("stopped")
```

```text
stopped
```

Swap a whole configuration at once. Each reader loads one record and uses only it, so it never sees
half of an old configuration and half of a new one:

```ant
type Config = { greeting: string, punctuation: string }

func greet(config: Atomic[Config], name: string): string {
  let c = config.load()
  "${c.greeting}, ${name}${c.punctuation}"
}

let config = Atomic.new(Config{ greeting: "Hello", punctuation: "." })
println(greet(config, "Ann"))
config.store(Config{ greeting: "Hi", punctuation: "!" })
println(greet(config, "Bob"))
```

```text
Hello, Ann.
Hi, Bob!
```

## Properties and relations

A load after a store on the same goroutine returns the stored value, and a `compareAndSwap` that
expects it succeeds:

```ant
let a = Atomic.new("draft")
a.store("final")
assert(a.load() == "final")
assert(a.compareAndSwap("final", "archived"))
assert(a.load() == "archived")
```

A store publishes the writes made before it. Once the main goroutine loads `true`, it sees every
slot the worker filled, with no lock:

```ant
let results = Array.filled(3, 0)
let ready = Atomic.new(false)
go {
  results[0] = 1
  results[1] = 2
  results[2] = 3
  ready.store(true)
}
while !ready.load() {
  Time.sleep(1ms)
}
assert(results == [1, 2, 3])
println(results)
```

```text
[1, 2, 3]
```

## Possible issues

`Atomic.new` infers the type of the cell from its argument, so `Atomic.new(null)` makes a cell of
type `Atomic[null]`, which can store nothing else. Annotate the type when later stores need a wider
one, as in `let name: Atomic[string | null] = Atomic.new(null)` or
`let total: Atomic[float64] = Atomic.new(0.0)`:

```ant
let name = Atomic.new(null)
name.store("Ann") // error: cannot use string where null is expected
```

A load followed by a store is two steps, not one. Another goroutine can store between them, and its
update is lost. This counter can end below 100:

<!-- test: check-only the total depends on goroutine order -->

```ant
let count = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Array.range(0, 100) {
  wg.add(1)
  go {
    defer { wg.done() }
    count.store(count.load() + 1)
  }
}
wg.wait()
println(count.load())
```

`Atomic.add` does the read and the write as one operation, so no update is lost:

```ant
let count = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Array.range(0, 100) {
  wg.add(1)
  go {
    defer { wg.done() }
    count.add(1)
  }
}
wg.wait()
println(count.load())
```

```text
100
```

A stored record is shared, not copied. Changing it after the store changes what `load` returns, and
that change is not atomic. Store a new record instead of changing a stored one:

```ant
type Config = { host: string, port: int }
let config = Config{ host: "localhost", port: 80 }
let current = Atomic.new(Config{ host: "", port: 0 })
current.store(config)
config.port = 9
current.load() // Config{ host: "localhost", port: 9 }
```

## See also

[Atomic.load](load.md) · [Atomic.compareAndSwap](compareAndSwap.md) · [Atomic.add](add.md) ·
[Atomic.new](new.md) · [Atomic](index.md) · [Mutex.withLock](../Mutex/withLock.md) ·
[WaitGroup.wait](../WaitGroup/wait.md) · [Concurrency guide](../../guide/concurrency.md)
