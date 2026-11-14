<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Once.do

```ant signature
func Once.do[T](once: Once, f: func(): T): void
```

Once.do runs f on the first call for once and discards its value; later calls wait for that run to
finish and do nothing. If f panics, once is complete and every call, waiting or later, panics with
the same message.

<!-- end generated -->

## Details

- Only the first call runs its function. Every later call ignores the function it is given, even a
  different one.
- A call made while the first run is still going waits for it to finish, so after any call returns,
  the run's effects are visible.
- A panic in `f` completes the once too: that call and every later one panic with the same message
  and trace, and `f` does not run again.
- A call inside `f` on the same once waits for `f` to finish, so it never returns.

## Basic examples

Only the first of three calls runs its function:

```ant
let once = Once.new()
for _ in Array.range(0, 3) {
  Once.do(once, { println("setting up") })
}
println("done")
```

```text
setting up
done
```

Several goroutines need the same setup. The first one runs it, the others wait for it, and all of
them see its result:

```ant
type Config = { region: string }

let once = Once.new()
let loads = Atomic.new(0)
let config = Config{ region: "" }

func loadConfig() {
  Time.sleep(10ms)
  config.region = "eu-west"
  loads.add(1)
}

let regions = Array.filled(4, "")
let wg = WaitGroup.new()
for i in Array.range(0, 4) {
  wg.add(1)
  go {
    defer { wg.done() }
    Once.do(once, loadConfig)
    regions[i] = config.region
  }
}
wg.wait()
println(regions)
println("loaded ${loads.load()} time")
```

```text
["eu-west", "eu-west", "eu-west", "eu-west"]
loaded 1 time
```

## Scope

The function may return any type. `Once.do` runs it for its effects and drops the value:

```ant
func readGreeting(): string {
  println("reading greeting")
  "hello"
}

let once = Once.new()
once.do(readGreeting)
once.do(readGreeting)
```

```text
reading greeting
```

A `Once` in a record field runs once per record. A field default gives each record its own:

```ant
type Server = { name: string, started: Once = Once.new() }

func Server.start(s: Server) {
  s.started.do({ println("starting ${s.name}") })
}

let api = Server{ name: "api" }
let web = Server{ name: "web" }
api.start()
web.start()
api.start()
```

```text
starting api
starting web
```

## Applications

Closing a channel twice panics. Several goroutines may each try to close the same channel; the once
closes it exactly once:

```ant
let stop: Channel[bool] = Channel.new()
let closeStop = Once.new()
let wg = WaitGroup.new()
for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    closeStop.do({ stop.close() })
  }
}
wg.wait()
println(stop.isClosed())
```

```text
true
```

## Properties and relations

The function runs once, however many calls there are, and the functions of later calls never run:

```ant
let once = Once.new()
var runs = 0
var later = false
once.do({ runs += 1 })
once.do({ runs += 1 })
once.do({ later = true })
assert(runs == 1)
assert(!later)
```

`Once.value` is the same guarantee for a function whose result you need. It returns the cached value
on every call:

```ant
let theme = Once.value({
  println("reading theme")
  "dark"
})
println(theme())
println(theme())
```

```text
reading theme
dark
dark
```

## Possible issues

`Once.do` returns `void`, so it does not hand back the function's value. Use `Once.value` instead:

```ant
let once = Once.new()
let ready = once.do({ "ready" })
println(ready.length()) // error: void has no field or function length
```

A failed run is not retried. After `f` panics, every later call panics with the same message, here
caught the first time with `Panic.catch`:

```ant
let once = Once.new()
let first = Panic.catch({ once.do({ panic("config missing") }) })
first is Err                  // true
once.do({ println("retry") }) // panic: config missing
```

A `Once` created inside a function is new on every call, so it guards nothing. Create it where it
outlives the calls, at the top level or in a record:

```ant
func setup() {
  let once = Once.new()
  once.do({ println("setting up") })
}

setup()
setup()
```

```text
setting up
setting up
```

## See also

[Once.new](new.md) · [Once.value](value.md) · [Once](index.md) ·
[Mutex.withLock](../Mutex/withLock.md) · [WaitGroup](../WaitGroup/index.md) ·
[Panic.catch](../Panic/catch.md) · [Concurrency guide](../../guide/concurrency.md)
