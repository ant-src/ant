<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Once.new

```ant signature
func Once.new(): Once
```

Once.new returns a new Once that has not run yet.

<!-- end generated -->

## Details

- Each call makes a distinct Once. Onces compare by identity, so two new Onces are never equal.
- A Once is a reference value: a copy in a record field, a parameter, or a closure is the same Once,
  and running it through any copy completes it for all of them.
- To compute a value once and reuse it, use `Once.value`, which makes its own Once; `Once.new` is
  for a `do` whose work has no result.

## Basic examples

Make a Once and run a function through it twice. Only the first call runs the function:

```ant
let once = Once.new()
once.do({ println("loading") })
once.do({ println("loading") })
```

```text
loading
```

A new Once renders as its type's name:

```ant
Once.new() // Once
```

## Scope

Keep a Once in a record field to set the record up the first time it is used. Each record gets its
own Once:

```ant
type Table = { name: string, ready: Once, rows: Array[string] }

func Table.open(name: string): Table {
  Table{ name, ready: Once.new(), rows: [] }
}

func Table.load(t: Table): Array[string] {
  t.ready.do({
    println("reading ${t.name}")
    t.rows.push("${t.name}-1")
  })
  t.rows
}

let users = Table.open("users")
let posts = Table.open("posts")
println(users.load())
println(users.load())
println(posts.load())
```

```text
reading users
["users-1"]
["users-1"]
reading posts
["posts-1"]
```

A Once at the top of a module guards work that every function of the module shares:

```ant
let setup = Once.new()

func greet(name: string): string {
  setup.do({ println("registering handlers") })
  "Hello, ${name}"
}

func leave(name: string): string {
  setup.do({ println("registering handlers") })
  "Bye, ${name}"
}

println(greet("Ann"))
println(leave("Bob"))
```

```text
registering handlers
Hello, Ann
Bye, Bob
```

## Applications

Initialize shared state once, however many goroutines ask at the same time. Every `do` returns only
after the one run has finished, so each goroutine sees the initialized value:

```ant
let init = Once.new()
let config: Map[string, string] = Map{}
let seen = Atomic.new(0)

Parallel.each(Array.range(0, 4), func(_: int, work: Context) {
  init.do({
    println("initializing")
    config["mode"] = "fast"
  })
  if config.get("mode") == "fast" {
    seen.add(1)
  }
}, limit: 4)
println(seen.load())
```

```text
initializing
4
```

## Properties and relations

Two new Onces are distinct, and a Once equals itself:

```ant
let a = Once.new()
let b = Once.new()
assert(a != b)
assert(a == a)
```

Completing one Once leaves another untouched:

```ant
let first = Once.new()
let second = Once.new()
var runs = 0
first.do({ runs += 1 })
first.do({ runs += 1 })
second.do({ runs += 1 })
assert(runs == 2)
```

## Possible issues

A Once made inside the function it guards is new on every call, so it runs every time. Make it once,
outside:

```ant
func connect(): void {
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

When the first run panics, the Once is complete all the same, and every later `do` panics with the
same message instead of trying again:

```ant
let once = Once.new()
let first = Panic.catch({ once.do({ panic("no config") }) })
first is Err                     // true
once.do({ println("retrying") }) // panic: no config
```

## See also

[Once.do](do.md) · [Once.value](value.md) · [Once](index.md) · [Mutex.new](../Mutex/new.md) ·
[WaitGroup.new](../WaitGroup/new.md)
