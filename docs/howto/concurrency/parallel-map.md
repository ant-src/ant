<!-- status: reviewed -->

# Process a list in parallel with a limit

Run one piece of work per element at the same time, but never more than a fixed number at once, and
get the results back in input order. The examples fetch pages from a web server whose address is in
`base`, such as `http://127.0.0.1:8080`.

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path.startsWith("/slow/") {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "late" }
  }
  if req.path.startsWith("/pages/") {
    Time.sleep(20ms)
    return Http.Reply{ status: 200, headers: none, body: "contents of ${req.path}" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
-->

```ant
func fetch(path: string, work: Context): Result[string, Error] {
  let response = try Http.get("${base}${path}", ctx: work)
  if response.status != 200 {
    return Result.fail("${path}: status ${response.status}")
  }
  Result.ok(response.text())
}

let paths = ["/pages/home", "/pages/about", "/pages/blog", "/pages/contact", "/pages/faq"]
let deadline = Context.new(10s)
let results = Parallel.map(paths, fetch, limit: 3, ctx: deadline)
if results == null {
  return Result.fail("gave up: ${deadline.reason() ?? "cancelled"}")
}
for page in try (results | Result.all()) {
  println(page)
}
```

```text
contents of /pages/home
contents of /pages/about
contents of /pages/blog
contents of /pages/contact
contents of /pages/faq
```

1. Write the work for one element as a function that takes the element and a
   [Context](../../lib/Context/index.md). Pass the context to every wait inside it, here
   [Http.get](../../lib/Http/get.md), so the work stops when the batch is cancelled. Pass the
   callback's own context, `work`, not `deadline`: only `work` ends when a sibling panics. Return
   [Result.fail](../../lib/Result/fail.md) for a failure and [Result.ok](../../lib/Result/ok.md) of
   [Http.text](../../lib/Http/text.md) for a page.
2. Call [Parallel.map](../../lib/Parallel/map.md) with `limit:`, the most callbacks that run at
   once. Here at most three requests are in flight; the next path starts when one of them ends.
   `limit` must be at least 1; a smaller value panics.
3. Bound the whole batch with a deadline from [Context.new](../../lib/Context/new.md), passed as
   `ctx:`. Parallel.map returns `null` when the context ends before every element has started, and
   [Context.reason](../../lib/Context/reason.md) says why.
4. Turn the array of results into one `Result` with [Result.all](../../lib/Result/all.md): the first
   `Err` in input order, or `Ok` of every page.

Parallel.map returns only after every callback has ended, and the results keep the order of `paths`,
whatever order the requests finish in. If a callback panics, the others are cancelled and the call
panics with the same message.

## Give up after a deadline

When the deadline passes while elements are still waiting for a slot, those elements never start,
the running callbacks see their context end, and the call returns `null`:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path.startsWith("/slow/") {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "late" }
  }
  if req.path.startsWith("/pages/") {
    Time.sleep(20ms)
    return Http.Reply{ status: 200, headers: none, body: "contents of ${req.path}" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
func fetch(path: string, work: Context): Result[string, Error] {
  let response = try Http.get("${base}${path}", ctx: work)
  if response.status != 200 {
    return Result.fail("${path}: status ${response.status}")
  }
  Result.ok(response.text())
}
-->

```ant
let paths = ["/pages/home", "/slow/archive", "/pages/about", "/slow/search", "/pages/faq"]
let deadline = Context.new(100ms)
let results = Parallel.map(paths, fetch, limit: 2, ctx: deadline)
if results == null {
  println("gave up: ${deadline.reason() ?? "cancelled"}")
}
```

```text
gave up: timeout
```

A deadline that passes after the last element has started does not give `null`: the waits in the
running callbacks give up, so Http.get returns `Err` with the message `timeout`, and the array holds
that `Err`.

## Keep the results that finished

Parallel.map gives all results or `null`. To keep what finished before the deadline, use
[Parallel.each](../../lib/Parallel/each.md) over [Array.entries](../../lib/Array/entries.md) and
write each result into its own slot of an array made with [Array.filled](../../lib/Array/filled.md).
The call returns after every callback has ended, so you can read the slots without a lock. A slot
still `null` is an element that never started:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path.startsWith("/slow/") {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "late" }
  }
  if req.path.startsWith("/pages/") {
    Time.sleep(20ms)
    return Http.Reply{ status: 200, headers: none, body: "contents of ${req.path}" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
func fetch(path: string, work: Context): Result[string, Error] {
  let response = try Http.get("${base}${path}", ctx: work)
  if response.status != 200 {
    return Result.fail("${path}: status ${response.status}")
  }
  Result.ok(response.text())
}
-->

```ant
let paths = ["/pages/home", "/slow/archive", "/pages/about", "/slow/search", "/pages/faq"]
let deadline = Context.new(100ms)
let pages: Array[Result[string, Error] | null] = Array.filled(paths.length(), null)
Parallel.each(paths.entries(), func(entry: Indexed[string], work: Context) {
  pages[entry.index] = fetch(entry.value, work)
}, limit: 2, ctx: deadline)
for entry in pages.entries() {
  let path = paths[entry.index]
  let page = entry.value
  if page == null {
    println("${path}: not started")
  } else if page is Ok {
    println("${path}: ${page.value}")
  } else {
    println("${path}: ${page.error.message}")
  }
}
```

```text
/pages/home: contents of /pages/home
/slow/archive: timeout
/pages/about: contents of /pages/about
/slow/search: timeout
/pages/faq: not started
```

## Stop at the first failure

An `Err` from one callback stops nothing: the other elements still run. To stop the batch at the
first failure, pass a context of your own as `ctx:` and cancel it with
[Context.cancel](../../lib/Context/cancel.md) when a callback fails. Record the error in an
[Atomic](../../lib/Atomic/index.md) made with [Atomic.new](../../lib/Atomic/new.md), since several
callbacks can fail at once. [Atomic.compareAndSwap](../../lib/Atomic/compareAndSwap.md) keeps only
the first error, and [Atomic.load](../../lib/Atomic/load.md) reads it after the call:

<!-- test: before
func handle(req: Http.Request): Http.Reply {
  let none: Map[string, string] = Map{}
  if req.path.startsWith("/slow/") {
    Time.sleep(1s)
    return Http.Reply{ status: 200, headers: none, body: "late" }
  }
  if req.path.startsWith("/pages/") {
    Time.sleep(20ms)
    return Http.Reply{ status: 200, headers: none, body: "contents of ${req.path}" }
  }
  Http.Reply{ status: 404, headers: none, body: "not found" }
}
let listener = try Net.listen("127.0.0.1:0")
let base = "http://${Net.address(listener)}"
go { Http.serve(listener, handle) }
func fetch(path: string, work: Context): Result[string, Error] {
  let response = try Http.get("${base}${path}", ctx: work)
  if response.status != 200 {
    return Result.fail("${path}: status ${response.status}")
  }
  Result.ok(response.text())
}
-->

```ant
let paths = ["/pages/home", "/missing", "/pages/about", "/pages/faq"]
let stop = Context.new()
let firstError: Atomic[Error | null] = Atomic.new(null)
let results = Parallel.map(paths, func(path: string, work: Context): Result[string, Error] {
  let page = fetch(path, work)
  if page is Err {
    firstError.compareAndSwap(null, page.error)
    stop.cancel()
  }
  page
}, limit: 2, ctx: stop)
if results == null {
  println("stopped: ${firstError.load()?.message}")
}
```

```text
stopped: /missing: status 404
```

The request for `/pages/home` was running when `/missing` failed, so its context ended too.
`/pages/about` and `/pages/faq` never started.

## See also

[Parallel.map](../../lib/Parallel/map.md) · [Parallel.each](../../lib/Parallel/each.md) ·
[Context.new](../../lib/Context/new.md) · [Run a worker pool over a channel](worker-pool.md) ·
[Give up after a timeout or deadline](timeout.md)
