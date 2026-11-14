<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Once

```ant signature
type Once
```

Once runs a function exactly once, however many goroutines ask. Onces compare by identity.

<!-- end generated -->

A `Once` guards work that must happen exactly once: setting up shared state, opening a connection,
closing a channel. The `Once` namespace holds a function that makes a once, one that runs work
through it, and one that wraps a loader so its value is computed on first use and cached.

## Details

- The first call runs its function. A call made while that run is going waits for it, so after any
  call returns, everything the run wrote is visible, from any goroutine.
- A panic in the function completes the once too. The function is not run again, and every call,
  waiting or later, panics with the same panic.
- A `Once` is a reference value: a copy in a record field, a parameter, or a closure is the same
  once.

## Basic examples

`Once.new` makes a once, and `Once.do` runs a function through it. Only the first call runs:

```ant
let setup = Once.new()
Once.do(setup, { println("connecting") })
Once.do(setup, { println("connecting") })
println("ready")
```

```text
connecting
ready
```

`Once.value` wraps a loader. The returned function runs it on its first call and returns the cached
value after that:

```ant
let config = Once.value({
  println("loading config")
  "debug=true"
})
println(config())
println(config())
```

```text
loading config
debug=true
debug=true
```

## Scope

A once in a record field guards each record on its own. A field default gives every record a fresh
one:

```ant
type Cache = { name: string, warmed: Once = Once.new() }

func Cache.warm(c: Cache) {
  c.warmed.do({ println("warming ${c.name}") })
}

let users = Cache{ name: "users" }
let posts = Cache{ name: "posts" }
users.warm()
posts.warm()
users.warm()
```

```text
warming users
warming posts
```

A once passed around is still one once. Running it through a copy completes it for every copy:

```ant
let shared = Once.new()
let alias = shared
alias.do({ println("first run") })
shared.do({ println("second run") })
println(shared == alias)
println(shared == Once.new())
println(shared)
```

```text
first run
true
false
Once
```

## Functions by job

### Running work once

- [Once.new](new.md): a new once that has not run yet
- [Once.do](do.md): run a function on the first call only; later calls wait for it and do nothing

### Computing a value once

- [Once.value](value.md): a function that runs the loader on its first call and caches its value

## Applications

### Load shared state on first use

Several goroutines ask for the same table at once. The loader runs once, and every goroutine gets
its result:

```ant
let loads = Atomic.new(0)
let table = Once.value({
  loads.add(1)
  Time.sleep(10ms)
  Map{ "ant": 6, "bee": 4 }
})

let legs = Array.filled(3, 0)
let wg = WaitGroup.new()
for i in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    legs[i] = table().get("ant") ?? 0
  }
}
wg.wait()
println(legs)
println(loads.load())
```

```text
[6, 6, 6]
1
```

### Close a channel exactly once

A second close of a channel panics. When any of several workers may decide to stop, a once makes the
close happen exactly once:

```ant
let stop: Channel[bool] = Channel.new()
let closeStop = Once.new()
Parallel.each(Array.range(0, 3), func(_: int, ctx: Context) {
  closeStop.do({ stop.close() })
}, limit: 3)
println(stop.isClosed())
```

```text
true
```

## Properties and relations

`Once.value(f)` does what a once, a nullable variable, and `Once.do` do by hand, without the
variable or the `??` at every read:

```ant
let once = Once.new()
var cached: string | null = null
func byHand(): string {
  once.do({ cached = "ready" })
  cached ?? panic("unreachable")
}

let lazy = Once.value({ "ready" })
assert(byHand() == lazy())
assert(lazy() == lazy())
```

Each once completes on its own:

```ant
let first = Once.new()
let second = Once.new()
var runs = 0
first.do({ runs += 1 })
first.do({ runs += 1 })
second.do({ runs += 1 })
assert(runs == 2)
assert(first != second)
```

## Possible issues

### A once made inside a function guards nothing

A once created on every call is a new once every time, so its function runs every time. Make it at
the top level or in a record, where it outlives the calls:

```ant
func connect() {
  let once = Once.new()
  once.do({ println("connecting") })
}

connect()
connect()
```

```text
connecting
connecting
```

### A failed run is not retried

After the function panics, the once is complete, and every later call panics with the same message:

```ant
let once = Once.new()
let first = Panic.catch({ once.do({ panic("config missing") }) })
first is Err                     // true
once.do({ println("retrying") }) // panic: config missing
```

`Once.value` caches a returned `Err` the same way. To retry after a failure, call the loader
directly and keep its value only after it succeeds, for example under a `Mutex`.

### Calling a once from its own function waits forever

A call made while the run is going waits for that run to finish. So the function cannot call the
same once, directly or through a helper: the inner call waits for the run it is part of, and the
program hangs. Give each step its own once, so the outer step can call the inner one:

```ant
let loadKeys = Once.new()
let loadConfig = Once.new()

func keys() {
  loadKeys.do({ println("loading keys") })
}

func config() {
  loadConfig.do({
    keys()
    println("loading config")
  })
}

config()
keys()
config()
```

```text
loading keys
loading config
```

### Once.do has no result

`Once.do` returns `void`. Use `Once.value` for work whose value you need:

```ant
let once = Once.new()
let name = once.do({ "ant" })
println(name.length()) // error: void has no field or function length
```

## See also

- [Mutex](../Mutex/index.md): guard shared state that changes more than once
- [Atomic](../Atomic/index.md): one value read and written atomically
- [WaitGroup](../WaitGroup/index.md): wait for goroutines to finish
- [Panic.catch](../Panic/catch.md): recover from the panic of a failed run
- [Concurrency](../../guide/concurrency.md): how goroutines, channels, and sync objects fit together
