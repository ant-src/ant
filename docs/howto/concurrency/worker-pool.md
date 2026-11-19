<!-- status: reviewed -->

# Run a worker pool over a channel

Hand jobs to a fixed number of workers as a producer finds them. This script reads file paths from
standard input, as `find logs -name '*.log' | ant run count.ant` gives them, and counts the lines of
each file with at most four files open at once:

<!-- test: file logs/api.log
GET /
GET /users
POST /users
-->
<!-- test: file logs/db.log
connect
query
-->
<!-- test: file logs/web.log
start
stop
-->
<!-- test: stdin
logs/web.log
logs/api.log
logs/missing.log
logs/db.log
-->

```ant
func countLines(path: string): string {
  let text = File.read(path)
  if text is Err { return text.error.message }
  "${text.value.lines().length()} lines"
}

let jobs: Channel[string] = Channel.new()
go {
  defer { jobs.close() }
  let input = Os.stdin()
  while true {
    let path = input.readLine().unwrapOr(null) ?? break
    jobs.send(path)
  }
}

let m = Mutex.new()
let reports: Map[string, string] = Map{}
Parallel.each(jobs, func(path: string, work: Context) {
  let report = countLines(path)
  m.withLock({ reports[path] = report })
}, limit: 4)

for path in reports.keys() | Array.sort() {
  println("${path}: ${reports[path]}")
}
```

```text
logs/api.log: 3 lines
logs/db.log: 2 lines
logs/missing.log: open logs/missing.log: no such file or directory
logs/web.log: 2 lines
```

1. Make the job channel with [Channel.new](../../lib/Channel/new.md). Annotate its type, since
   nothing else fixes it.
2. Start the producer with `go`. Send each job with [Channel.send](../../lib/Channel/send.md), and
   close the channel with [Channel.close](../../lib/Channel/close.md) in a `defer`, so the workers
   end even when the producer stops early. Here the producer reads lines with
   [File.readLine](../../lib/File/readLine.md) and stops at the end of the input or at a read error.
3. Run the workers with [Parallel.each](../../lib/Parallel/each.md) over the channel. `limit:` is
   the number of workers. The call takes a job only when a worker is free, and returns once the
   channel is closed and empty and every job has finished. Here four bounds the files open at once.
   For jobs that keep a CPU busy, use [`Os.cpus()`](../../lib/Os/cpus.md) workers.
4. Collect the results in a map guarded by [Mutex.withLock](../../lib/Mutex/withLock.md), since
   several workers write it at once. Read it after `Parallel.each` returns, with no lock.

The jobs finish in no fixed order, so sort the results with [Array.sort](../../lib/Array/sort.md)
before you print them. A panic in a job ends the pool and is raised again where `Parallel.each` was
called.

## Stop the pool at the first failure

Make a context with [Context.new](../../lib/Context/new.md) and give it to both sides: to
`Parallel.each` as `ctx:`, and to each `Channel.send` of the producer. A job that fails calls
[Context.cancel](../../lib/Context/cancel.md). The pool then takes no more jobs, and the producer's
send returns `false` instead of waiting for a worker that never comes:

```ant
let ctx = Context.new()
let jobs: Channel[string] = Channel.new()
go {
  defer { jobs.close() }
  for path in ["logs/api.log", "logs/missing.log", "logs/db.log"] {
    if !jobs.send(path, ctx: ctx) { break }
  }
}

let failure: Atomic[Error | null] = Atomic.new(null)
Parallel.each(jobs, func(path: string, work: Context) {
  let text = File.read(path)
  if text is Err {
    failure.compareAndSwap(null, text.error)
    ctx.cancel()
  }
}, limit: 4, ctx: ctx)

let err = failure.load()
if err != null {
  println("stopped: ${err.message}")
}
```

```text
stopped: open logs/missing.log: no such file or directory
```

[Atomic.compareAndSwap](../../lib/Atomic/compareAndSwap.md) keeps the first error when several jobs
fail at once. A job that has already started runs to its end; pass `work` to its own waits so they
end when the pool stops.

## Write the pool with go and WaitGroup

Start the workers yourself when they send their results on to another channel. Each worker loops
over [Channel.values](../../lib/Channel/values.md) until the job channel is closed and empty. A
[WaitGroup](../../lib/WaitGroup/index.md) counts the workers, and a last goroutine closes the
results channel once they have all ended. A file that fails to read adds nothing to the total:

```ant
let jobs: Channel[string] = Channel.new()
go {
  defer { jobs.close() }
  for path in ["logs/api.log", "logs/db.log", "logs/web.log"] {
    jobs.send(path)
  }
}

let results: Channel[int] = Channel.new()
let workers = 4
let wg = WaitGroup.new()
wg.add(workers)
for _ in Iter.range(0, workers) {
  go {
    defer { wg.done() }
    for path in jobs.values() {
      let text = File.read(path)
      if text is Err { continue }
      results.send(text.value.lines().length())
    }
  }
}
go {
  wg.wait()
  results.close()
}

var total = 0
for n in results.values() {
  total += n
}
println("${total} lines")
```

```text
7 lines
```

1. Call [WaitGroup.add](../../lib/WaitGroup/add.md) with the number of workers before you start
   them.
2. Call [WaitGroup.done](../../lib/WaitGroup/done.md) in a `defer` at the top of each worker, so a
   worker that panics still counts as ended.
3. Close the results channel from a goroutine that calls
   [WaitGroup.wait](../../lib/WaitGroup/wait.md). The main goroutine reads the results meanwhile, so
   no worker waits on a send that nobody takes.

Unlike `Parallel.each`, a panic here ends only the worker it occurs in, and the others go on. To
keep that worker taking jobs, run each job through [Panic.catch](../../lib/Panic/catch.md).

## See also

[Channel.new](../../lib/Channel/new.md) · [Channel.send](../../lib/Channel/send.md) ·
[Channel.close](../../lib/Channel/close.md) · [Channel.values](../../lib/Channel/values.md) ·
[Channel.receive](../../lib/Channel/receive.md) · [Parallel.each](../../lib/Parallel/each.md) ·
[WaitGroup](../../lib/WaitGroup/index.md) · [Context.new](../../lib/Context/new.md) ·
[Context.cancel](../../lib/Context/cancel.md) · [Mutex.withLock](../../lib/Mutex/withLock.md) ·
[Atomic.compareAndSwap](../../lib/Atomic/compareAndSwap.md) · [Os.stdin](../../lib/Os/stdin.md) ·
[Os.cpus](../../lib/Os/cpus.md) · [File.readLine](../../lib/File/readLine.md) ·
[File.read](../../lib/File/read.md) · [Process a list in parallel](parallel-map.md) ·
[Cancel a group of goroutines](cancel.md) · [Concurrency](../../guide/concurrency.md)
