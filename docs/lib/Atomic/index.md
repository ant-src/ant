<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Atomic

```ant signature
type Atomic[T]
```

Atomic is a cell holding one value of T, read and written atomically. All operations on all Atomic
cells form one total order. Atomics compare by identity.

<!-- end generated -->

An atomic cell holds one value that several goroutines read and change without a lock. Each
operation on the cell happens as one step, so a goroutine never sees a change half made and two adds
never lose one of them. The `Atomic` namespace holds the functions that make a cell and read, write,
add to, and swap its value; `Atomic[T]` is also the type, for an annotation or a record field.

## Details

- The type of the cell is fixed when it is made. `Atomic.new(null)` makes a cell that holds only
  `null`, so annotate a cell that starts empty:
  `let latest: Atomic[string | null] = Atomic.new(null)`.
- A `load` that sees a `store` from another goroutine also sees everything that goroutine wrote
  before the `store`.
- The cell holds the value itself, not a copy. Changing a field of a loaded record changes the
  record in the cell, with no atomic step. Store a new record instead.
- `compareAndSwap` compares with `==`, so a record or array matches an equal one built anew, and a
  cell holding `NaN` never swaps. See [Atomic.compareAndSwap](compareAndSwap.md) for the details.

## Basic examples

`Atomic.new` makes a cell, `Atomic.add` adds to it, and `Atomic.load` reads it. A hundred callbacks
count into one cell, and no increment is lost:

```ant
let hits = Atomic.new(0)
Parallel.each(Array.range(0, 100), func(i: int, ctx: Context) {
  Atomic.add(hits, 1)
}, limit: 8)
println(Atomic.load(hits))
```

```text
100
```

`Atomic.store` sets the value. A worker runs until the main goroutine stores `true` into a stop
flag:

```ant
let stop = Atomic.new(false)
let rounds = Atomic.new(0)
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  while !stop.load() {
    rounds.add(1)
    Time.sleep(1ms)
  }
}
Time.sleep(20ms)
Atomic.store(stop, true)
wg.wait()
println(rounds.load() > 0)
```

```text
true
```

## Scope

A cell holds a value of any type, such as a string, a float, or a record:

```ant
type Point = { x: int, y: int }
let status = Atomic.new("starting")
status.store("ready")
status.load()                          // "ready"
Atomic.new(0.5).add(0.25)              // 0.75
Atomic.new(Point{ x: 1, y: 2 }).load() // Point{ x: 1, y: 2 }
```

A nullable cell needs its type written out, and `null` stays possible after a load:

```ant
let owner: Atomic[string | null] = Atomic.new(null)
owner.compareAndSwap(null, "ann") // true
owner.load()                      // "ann"
```

A field default gives each record its own cell:

```ant
type Server = { name: string, requests: Atomic[int] = Atomic.new(0) }

let api = Server{ name: "api" }
let admin = Server{ name: "admin" }
api.requests.add(3)
println(api.requests.load())
println(admin.requests.load())
println(api)
```

```text
3
0
Server{ name: "api", requests: Atomic }
```

Atomics compare by identity: a cell equals only itself, under any name, never another cell with the
same value:

```ant
let a = Atomic.new(1)
let alias = a
a == Atomic.new(1) // false
a == alias         // true
a                  // Atomic
```

The functions are functions on the type, so they pass as function values once the call fixes `T`:

```ant
let cells = [Atomic.new(1), Atomic.new(2)]
cells.map(Atomic.load) // [1, 2]
```

## Functions by job

### Making a cell

- [Atomic.new](new.md): a new cell holding a value

### Reading and writing

- [Atomic.load](load.md): the value in the cell
- [Atomic.store](store.md): set the value in the cell

### Changing the value from the one before

- [Atomic.add](add.md): add to a number and return the new value; panics on overflow
- [Atomic.compareAndSwap](compareAndSwap.md): set the value only if it equals an expected one

## Applications

### Hand a result back from a goroutine

A goroutine cannot reassign a local of its parent and have it narrow there. Store the result in a
cell, and after the wait load it into a fresh `let` and narrow that:

```ant
func fetch(): Result[string, Error] {
  Result.ok("pong")
}

let latest: Atomic[Result[string, Error] | null] = Atomic.new(null)
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  latest.store(fetch())
}
wg.wait()
let reply = latest.load()
if reply != null && reply is Ok { println(reply.value) }
```

```text
pong
```

### Publish a new configuration

Readers load the current configuration while a reload stores a new one whole. A reader sees the old
record or the new one, never a mix:

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

### Give out unique numbers

`add` returns the new value, so each goroutine gets its own number:

```ant
let next = Atomic.new(0)
let ids = Parallel.map(["ann", "bob", "cy"], func(name: string, ctx: Context) {
  next.add(1)
}, limit: 3)
println(ids?.sort())
```

```text
[1, 2, 3]
```

## Properties and relations

A `store` happens before every `load` that sees it, so data written before the `store` is visible
after the `load`:

```ant
let data: Array[string] = []
let ready = Atomic.new(false)
go {
  data.push("loaded")
  ready.store(true)
}
while !ready.load() {
  Time.sleep(1ms)
}
assert(data == ["loaded"])
```

`add` is a compare-and-swap loop done in one step. Both give the same value:

```ant
let a = Atomic.new(10)
let b = Atomic.new(10)
while true {
  let old = a.load()
  if a.compareAndSwap(old, old + 5) { break }
}
b.add(5)
assert(a.load() == b.load())
```

## Possible issues

### A load and a store are two steps

Another goroutine can write between a `load` and the `store` that follows it, and its update is
lost. Here two loads read 0 before either store, as two goroutines can:

```ant
let count = Atomic.new(0)
let first = count.load()
let second = count.load()
count.store(first + 1)
count.store(second + 1)
count.load() // 1
```

Use `add` for a number, or `compareAndSwap` in a loop for any other change.

### Each load is a new read

A test on one `load` does not narrow the next one, since another goroutine may store in between:

```ant
let latest: Atomic[string | null] = Atomic.new("v2")
if latest.load() != null {
  println(latest.load().length())
  // error: string | null has no field or function length on every member: narrow the value first
}
```

Load once into a `let` and test that:

```ant
let latest: Atomic[string | null] = Atomic.new("v2")
let version = latest.load()
if version != null { println(version.length()) }
```

```text
2
```

### A cell made from `null` holds only `null`

```ant
let latest = Atomic.new(null)
latest.store("ready") // error: cannot use string where null is expected
```

Annotate the cell, as in [Scope](#scope).

### `add` needs a number

```ant
let log = Atomic.new("a")
log.add("b") // error: T would be string, which is not a member of its bound Number
```

### Changing a loaded record

A field set on a loaded record changes the record in the cell, outside any atomic step:

```ant
type Point = { x: int, y: int }
let p = Atomic.new(Point{ x: 1, y: 2 })
let loaded = p.load()
loaded.x = 9
p.load() // Point{ x: 9, y: 2 }
```

Build a new record and `store` it, or swap it in with `compareAndSwap`. For several values that
change together, use a [Mutex](../Mutex/index.md).

## See also

- [Mutex](../Mutex/index.md): guard several values that change together
- [WaitGroup](../WaitGroup/index.md): wait for goroutines to finish
- [Once](../Once/index.md): run initialization once
- [Channel](../Channel/index.md): pass values between goroutines instead of sharing them
- [Parallel](../Parallel/index.md): run callbacks on goroutines and wait for them
- [Concurrency](../../guide/concurrency.md): how goroutines, channels, and sync objects fit together
