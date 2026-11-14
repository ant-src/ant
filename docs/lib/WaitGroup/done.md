<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# WaitGroup.done

```ant signature
func WaitGroup.done(wg: WaitGroup): void
```

WaitGroup.done decrements the counter of wg by one. It panics if the counter is already zero.

<!-- end generated -->

## Details

- Each call takes one off the counter: one `done` for each unit an `add` counted.
- Everything a goroutine writes before its `done` is visible after the `wait` that the `done`
  unblocks.
- A `done` with no `add` before it, or one more than `add` counted, panics.
- A panic in a goroutine skips a plain `done` at its end, and `wait` then never sees the counter
  reach zero. Call it in a `defer` at the top of the goroutine.

## Basic examples

Count one goroutine, mark it done when it finishes, and wait for it:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  println("working")
  WaitGroup.done(wg)
}
wg.wait()
println("finished")
```

```text
working
finished
```

Put the `done` in a `defer` at the top of each goroutine, so it runs on every way out. Each
goroutine writes its own slot, and the main goroutine reads them after the wait:

```ant
let wg = WaitGroup.new()
let squares = Array.filled(4, 0)
for i in Array.range(0, 4) {
  wg.add(1)
  go {
    defer { WaitGroup.done(wg) }
    squares[i] = i * i
  }
}
wg.wait()
println(squares)
```

```text
[0, 1, 4, 9]
```

## Scope

One `add` can count several goroutines at once; each of them calls `done` once:

```ant
let wg = WaitGroup.new()
let names = ["Ann", "Bob", "Cid"]
let greetings = Array.filled(names.length(), "")
wg.add(names.length())
for entry in names.entries() {
  go {
    defer { wg.done() }
    greetings[entry.index] = "hi, ${entry.value}"
  }
}
wg.wait()
println(greetings)
```

```text
["hi, Ann", "hi, Bob", "hi, Cid"]
```

`done` does not have to run on another goroutine. When the counter reaches zero on the goroutine
that waits, `wait` returns at once:

```ant
let wg = WaitGroup.new()
wg.add(2)
wg.done()
wg.done()
wg.wait()
println("counter is zero")
```

```text
counter is zero
```

A wait group is a reference value. Pass it to a function, and that function's `done` counts down the
same counter:

```ant
func work(id: int, wg: WaitGroup, out: Array[string]) {
  defer { wg.done() }
  out[id] = "job ${id} done"
}

let wg = WaitGroup.new()
let out = Array.filled(3, "")
for id in Array.range(0, 3) {
  wg.add(1)
  go { work(id, wg, out) }
}
wg.wait()
println(out | String.join("\n"))
```

```text
job 0 done
job 1 done
job 2 done
```

## Applications

Close a results channel once every worker is done. A separate goroutine waits for the workers and
then closes the channel, which ends the `for` loop over its values:

```ant
let jobs = [3, 1, 4, 1, 5]
let results: Channel[int] = Channel.new()
let wg = WaitGroup.new()
for job in jobs {
  wg.add(1)
  go {
    defer { wg.done() }
    results.send(job * 10)
  }
}
go {
  wg.wait()
  results.close()
}
var total = 0
for r in results.values() {
  total += r
}
println(total)
```

```text
140
```

## Properties and relations

Everything a goroutine writes before its `done` is visible after the `wait` that the `done`
unblocks:

```ant
type Tally = { a: int, b: int }

let tally = Tally{ a: 0, b: 0 }
let wg = WaitGroup.new()
wg.add(2)
go {
  tally.a += 1
  wg.done()
}
go {
  tally.b += 1
  wg.done()
}
wg.wait()
assert(tally.a == 1 && tally.b == 1)
```

## Possible issues

A `done` with no `add` before it, or one more than `add` counted, panics:

```ant
WaitGroup.done(WaitGroup.new()) // panic: WaitGroup.done: WaitGroup.done when the counter is already zero
```

```ant
let wg = WaitGroup.new()
wg.add(1)
wg.done()
wg.done() // panic: WaitGroup.done: WaitGroup.done when the counter is already zero
```

A panic before a plain `done` skips it, so the counter never reaches zero and `wait` blocks forever.
Here a context with a timeout ends the wait instead, and `Context.isDone` shows that the wait timed
out:

```ant
let wg = WaitGroup.new()
let ctx = Context.new(timeout: 50ms)
wg.add(1)
go {
  let empty: Array[int] = []
  println(empty[0])
  wg.done()
}
wg.wait(ctx: ctx)
println(ctx.isDone())
```

```text
true
```

With `defer { wg.done() }` at the top of the goroutine, the panic still counts it down, and the wait
returns without the context:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  let empty: Array[int] = []
  println(empty[0])
}
wg.wait()
println("not stuck")
```

```text
not stuck
```

## See also

[WaitGroup.add](add.md) · [WaitGroup.wait](wait.md) · [WaitGroup.new](new.md) ·
[WaitGroup](index.md) · [Parallel.each](../Parallel/each.md) · [Mutex.unlock](../Mutex/unlock.md) ·
[Channel.close](../Channel/close.md) · [defer](../../language/keywords/defer.md) ·
[Concurrency guide](../../guide/concurrency.md)
