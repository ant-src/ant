<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Once.value

```ant signature
func Once.value[T](f: func(): T): func(): T
```

Once.value returns a function that runs f on its first call and returns the cached value on every
call. If f panics, every call, waiting or later, panics with the same message.

<!-- end generated -->

## Details

- `Once.value` itself runs nothing. `f` runs on the first call of the returned function, and never
  when that function is not called.
- Every call returns the value of that one run, even `null` or an `Err`. A mutable value, such as an
  array, is the same value at every call, not a copy.
- Calls from several goroutines at once run `f` once: the others wait for it to finish, and then see
  everything `f` wrote.
- If `f` panics, `f` is not run again. Every call, waiting or later, panics with the same panic.

## Basic examples

The first call runs the block. Later calls return the cached value:

```ant
let greeting = Once.value({
  println("computing")
  "hello"
})
println(greeting())
println(greeting())
```

```text
computing
hello
hello
```

Nothing runs until the first call:

```ant
var runs = 0
let answer = Once.value({ runs += 1; 42 })
runs     // 0
answer() // 42
answer() // 42
runs     // 1
```

## Scope

A function with no parameters can be passed by name:

```ant
type Config = { host: string, port: int }

func loadConfig(): Config {
  println("loading")
  Config{ host: "localhost", port: 8080 }
}

let config = Once.value(loadConfig)
println(config().host)
println(config().port)
```

```text
loading
localhost
8080
```

A block that returns `void` gives a function that does its work once, a setup step that any caller
can trigger:

```ant
let setup = Once.value({ println("setting up") })
setup()
setup()
println("done")
```

```text
setting up
done
```

A `null` result is cached like any other value, so `f` does not run again to look for a better one:

```ant
var lookups = 0
let token = Once.value(func(): string | null {
  lookups += 1
  null
})
token() // null
token() // null
lookups // 1
```

A mutable result is shared, not copied. A change through one call shows in the next:

```ant
let cache = Once.value(func(): Array[string] { [] })
cache().push("a")
cache().push("b")
cache() // ["a", "b"]
```

Each call of `Once.value` makes a new cache, and each runs `f` on its own:

```ant
var runs = 0
func count(): int {
  runs += 1
  runs
}

let first = Once.value(count)
let second = Once.value(count)
first()  // 1
second() // 2
first()  // 1
```

## Applications

Build a lookup table the first time a function needs it, and not at all in a run that never does:

```ant
let squares = Once.value({
  println("building table")
  Array.range(0, 10).map({ x * x })
})

func square(n: int): int {
  squares()[n]
}

println(square(3))
println(square(9))
```

```text
building table
9
81
```

Share one loaded value among goroutines. However many ask at once, the loader runs once:

```ant
let loads = Atomic.new(0)
let load = Once.value({
  loads.add(1)
  Time.sleep(10ms)
  "config"
})

let wg = WaitGroup.new()
let got = Array.filled(4, "")
for i in Iter.range(0, 4) {
  wg.add(1)
  go {
    defer { wg.done() }
    got[i] = load()
  }
}
wg.wait()
println(got)
println(loads.load())
```

```text
["config", "config", "config", "config"]
1
```

Read a file once and keep the `Result`. A failure is cached as well, so the file is read once even
when it is missing:

<!-- test: file motd.txt
Welcome back
-->

```ant
var reads = 0
let motd = Once.value({ File.read("motd.txt") })
let missing = Once.value({
  reads += 1
  File.read("missing.txt")
})

let text = motd()
if text is Ok {
  println(text.value.trim())
}
println(missing() is Err)
println(missing() is Err)
println(reads)
```

```text
Welcome back
true
true
1
```

## Properties and relations

Every call returns the same value, and that value is what `f` returns:

```ant
func build(): Array[int] {
  [3, 1, 2].sort()
}

let sorted = Once.value(build)
assert(sorted() == build())
assert(sorted() == sorted())
```

`Once.value(f)` does what a `Once` and a nullable variable do with `Once.do`, without the variable
and the `??` at every read:

```ant
let once = Once.new()
var table: Array[int] | null = null
func tableByHand(): Array[int] {
  once.do({ table = [1, 2, 3] })
  table ?? panic("unreachable")
}

let tableLazy = Once.value({ [1, 2, 3] })
assert(tableByHand() == tableLazy())
```

## Possible issues

A panic is cached too. `f` does not run again, and every later call panics with the same message:

```ant
var tries = 0
let connect = Once.value(func(): string {
  tries += 1
  panic("connection refused")
})

let first = Panic.catch(connect) | Result.mapErr({ x.message })
let second = Panic.catch(connect) | Result.mapErr({ x.message })
println(first)
println(second)
println(tries)
```

```text
Err{ error: "connection refused" }
Err{ error: "connection refused" }
1
```

`Once.value` never runs `f` a second time, whether `f` panics or returns an `Err`. To retry after a
failure, call the loader directly and keep a value only after it succeeds, for example under a
`Mutex`.

Calling `Once.value` inside a function makes a new cache on every call, so nothing is cached. Create
it once, outside the function:

```ant
var loads = 0
func load(): string {
  loads += 1
  "data"
}

func dataWrong(): string {
  Once.value(load)()
}

let cached = Once.value(load)
func dataRight(): string {
  cached()
}

dataWrong()
dataWrong()
dataRight()
dataRight()
println(loads)
```

```text
3
```

## See also

[Once.do](do.md) · [Once.new](new.md) · [Once](index.md) · [Mutex.withLock](../Mutex/withLock.md) ·
[Atomic](../Atomic/index.md) · [WaitGroup](../WaitGroup/index.md) · [Panic.catch](../Panic/catch.md)
· [go](../../language/keywords/go.md)
