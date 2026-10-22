# Plan 6: the waiting

Plan 5 gave Ant the small change of every program's first hour and left out everything that
waits: a server, a socket, a deadline on a blocking call, a goroutine told to stop. They were
left out together because they share one design question. How does a goroutine wait for one of
several things, and how is it told to give up, in a language that has goroutines but neither
`select`, nor channels, nor a context?

Go answers with `select` over channels and a `context` threaded through every blocking call.
This plan takes the two halves Ant can carry without new syntax, a channel and a context as
sync objects, states one rule for blocking calls, and builds sockets, a server and a graceful
interrupt on top. It takes `select` as a library call with a clause per channel, as Kotlin's
coroutines have it, and not as a statement.

## 1. Where we start

Section 9 of `ant-lang.md` is the concurrency chapter: `go` runs a block on a new goroutine,
there is no handle to it, results come back through shared memory, completion through a
`WaitGroup`, and "there are no channels". The sync objects are `Mutex`, `RwMutex`, `WaitGroup`,
`Once`, `Cond` and `Atomic[T]`, each a reference created by `new`, each a type with functions
declared on it. A blocking library call blocks only the calling goroutine.

There is already a cancellation, but it belongs to the host. A run has a context: the `ant`
command cancels it on Ctrl-C, and a Go program cancels the one it passed to `Run`. When that
happens the run is over: `Run` returns at once, every goroutine ends at its next call or loop
iteration, the blocking calls that watch the context return (`Time.sleep`, `Http.request`,
`Os.exec`, a read of standard input), and the rest get a hundred milliseconds of grace before
the process exits. No program code runs after an interrupt. It is an end, never an event.

Within a program, nothing can be bounded or stopped. `Http.request` has a `timeout` since Plan
5 and nothing else does: `WaitGroup.wait`, `Mutex.lock` and a read of a handle block until
their condition holds. A goroutine is told to stop only through a flag it polls in shared
memory. "The first of several things" can be written with a `Cond` and a shared field, and
nobody writes that in their first hour.

The gaps, by what goes wrong in a program:

- **A queue.** A worker that takes jobs from a queue, a pipeline, a result handed back: there
  is no blocking queue. A `Mutex`, a `Cond` and an array make one in fifteen lines.
- **A bound.** A wait that gives up after a while, or when something else happens.
- **A stop.** A goroutine that should end when its work is no longer wanted.
- **A server.** `Http` is a client only, and there is no `Net`.
- **An interrupt.** Ctrl-C kills the run. A server cannot drain, a job cannot write what it has.

In the implementation, the run's context is `in.ctx`, which `g.Context()` returns; a `done`
flag is checked at calls and loop back-edges and ends a goroutine with a `cancelSignal`;
`Instance.wait` returns when main returns, `Os.exit` is called, or the context is cancelled, and
then waits out the grace period. `File.Handle` wraps an `*os.File` with buffered reader and
writer, and a read of standard input runs in a goroutine so a cancelled run is not held by a
terminal. `Time.sleep` selects on the context already. The sync objects are small state
machines in `value/sync.go`; `Atomic[T]` shows that a generic opaque sync type is declared in
the prelude and typed by the checker without special cases. The `ant` command uses
`signal.NotifyContext` with `os.Interrupt` only.

## 2. Definition of done

- Section 9 says how goroutines wait and how they stop; `Chan`, `Select` and `Context` are in
  section 11 with the other sync objects, and `Net`, `Http.serve` and `Os.interrupt` are in
  `stdlib.md`.
  Every new function has its signature, its comment and its ◆; every changed one its new shape.
- The prelude declares each, a builtin implements each, and the three coverage gates have no
  pending entry. Every ◆ is reached by a program.
- Every wait in the corpus has a bound, so no program can hang the suite, and the suite's wall
  time grows by less than two seconds.
- The suite is green under `-race`, including a program that passes ten thousand values through
  a channel between a hundred goroutines and one that cancels a context under a hundred blocked
  receivers.
- Three new examples: a worker pool, an echo server over sockets, an HTTP API that drains on
  Ctrl-C.
- `CHANGELOG.md`, the README and the package documentation in `ant.go`, with the new `ant`
  option for interrupts.

## 3. What Go provides and what Ant takes

Go waits with `select` over channels, bounds with `context`, times with `time.Timer` and
`time.Ticker`, listens with `net` and `net/http.Server`, and hears signals with `os/signal`.
Ant takes the channel and the context as sync objects, `select` as a function over clauses, the
socket as a handle, the server as a function, and the interrupt as a context. It takes no
`select` statement and no timer objects, and section 3.7 says how programs wait instead.

### 3.1 Context (core, Synchronization)

A context is a sync object that becomes done once, for one of three reasons: a call cancelled
it, its deadline passed, or its parent became done. Every context descends from the run's
context, which the host cancels, so a cancelled run still ends every wait in the program, as it
does today.

```
Context.new(timeout: Time.Duration | null = null, parent: Context | null = null): Context
                                                   // a child of parent, or of the run; ◆ a timeout with years, months or weeks, or not positive
Context.cancel(ctx: Context): void                 // done now, with the reason "cancelled"; nothing when done already
Context.isDone(ctx: Context): bool
Context.reason(ctx: Context): string | null        // null while running; "cancelled", "timeout", "interrupt", or the parent's reason
Context.deadline(ctx: Context): Time.Instant | null   // the earliest deadline of ctx and its parents
Context.wait(ctx: Context): void                   // blocks until done
```

The timeout converts as `Http.request`'s does, days at 24 hours and the rest exact, but a
calendar unit or a duration that is not positive is a ◆ rather than an `Err`, since
`Context.new` is computation, not I/O. `Context.wait` is how a main goroutine waits for the
interrupt while a server runs. A context is a reference compared by identity, renders as
`Context`, and cannot be tested by `is`, like the other sync objects.

**The rule.** Every blocking library call takes `ctx: Context | null = null` as its last
parameter and gives up when the context is done. What giving up means follows the call's
return type: a `Result` is `Err` with the reason as its message, a value-or-null is `null`, and
a `void` returns early, so the program asks `Context.isDone` when it needs to know. With `null`
the call waits as it does today, until the run is cancelled. The calls are `Chan.send`,
`Chan.receive` and `Chan.values`, `WaitGroup.wait`, `Time.sleep`, `Http.request`, `Http.get` and
`Http.post`, `Select.wait`, `Os.exec`, `File.readLine`, `File.readAll` and `File.pipe`,
`Net.accept`, `Net.dial` and `Http.serve`. Locks and conditions are not on the list (decision 7).

### 3.2 Chan and Select (core, Synchronization)

A channel carries values from goroutines to goroutines, in order, and blocks the side that is
ahead. Section 9's sentence "there are no channels" becomes "a channel is a sync object, and
`select` is a function".

```
Chan.new[T](capacity: int = 0): Chan[T]                            // ◆ capacity < 0; 0 is a rendezvous
Chan.send[T](c: Chan[T], v: T, ctx: Context | null = null): bool   // blocks until taken or buffered; false when ctx ends first; ◆ closed
Chan.receive[T](c: Chan[T], ctx: Context | null = null): T | null  // blocks; null when closed and drained, or when ctx ends first
Chan.close[T](c: Chan[T]): void                                    // ◆ closed already
Chan.isClosed[T](c: Chan[T]): bool
Chan.length[T](c: Chan[T]): int                                    // values buffered now
Chan.values[T](c: Chan[T], ctx: Context | null = null): Seq[T]     // every value until closed, or until ctx ends
```

An unbuffered channel hands a value over directly: `send` returns when a receiver has taken it.
A buffered one holds `capacity` values and blocks a sender beyond that. Closing lets receivers
drain what is buffered and then get `null`; a later `send` is a ◆, as in Go, since a value with
no receiver is a bug in the sender. `Chan.values` is the `for` form:
`for job in Chan.values(queue)` ends when the channel is closed and drained.

`Chan.new` fixes `T` from nothing, so a declaration annotates:
`var jobs: Chan[Job] = Chan.new()`. `Chan.receive` conflates a `null` value with the end of the
channel when `T` includes `null`, as `Map.get` and `Array.pop` do; a program with such a `T`
wraps the value, and `Context.isDone(ctx)` tells a cancellation from a close (decision 4).

Three orderings join section 9's list: a send happens before the receive that takes its value;
a close happens before a receive that returns `null` because of it; and a cancellation happens
before the return of every wait it ends.

**Select.** A wait on several channels at once is a call with a clause per channel, each clause
carrying the function to run when it is chosen. `Chan` is a type with functions, and section 6
lets a type carry no types, so the clauses live in a namespace of their own:

```
type Case[R]                                                       // opaque: one clause of a select, compared by identity

Select.receive[T, R](c: Chan[T], f: func(T | null): R): Select.Case[R]   // ready when c holds a value or is closed; f gets the value, or null
Select.send[T, R](c: Chan[T], v: T, f: func(): R): Select.Case[R]        // ready when c can take v; ◆ closed, when chosen
Select.otherwise[R](f: func(): R): Select.Case[R]                        // ready at once; runs when no other clause is ready
Select.wait[R](cases: Array[Select.Case[R]], ctx: Context | null = null): R | null
                                                                   // runs one ready clause and returns its result; null when ctx ends first
                                                                   // ◆ no cases, or two otherwise clauses
```

`Select.wait` takes nothing until a clause is chosen, as Go's `select` does: a value stays in
its channel until the moment it is taken, so a channel that another goroutine reads, or that a
later state will read, is not drained. Among several ready clauses one is chosen at random, so
no source starves. With an `otherwise` clause the call never blocks, which is the non-blocking
try. The clauses hide each channel's element type inside their function, so channels of
different types mix in one wait, and `R` is what the functions agree to return: a state machine
returns its next state, an event loop returns `void`, and a wait whose clauses return `void` has
the type `void | null`, which a statement ignores.

```ant
type State = Line | Tick | Sent | Closed | Stopped

var next: State = Select.wait([
  Select.receive(lines, { if x == null { Closed{} } else { Line{ text: x } } }),
  Select.receive(ticks, { Tick{} }),
  Select.send(out, batch, { Sent{} }),
], ctx) ?? Stopped{}
```

A function given to a clause returns from itself, not from the function around the wait, so a
`break` or a `return` that depends on the clause goes through the result, as `next` above. That
is why the mailbox of section 3.7, a plain loop in the function's own body, is the shape for an
event loop, and `Select.wait` is for channels that exist already.

### 3.3 The blocking calls that gain a context (core and stdlib)

```
WaitGroup.wait(wg: WaitGroup, ctx: Context | null = null): void   // returns early when ctx ends
Time.sleep(ms: int, ctx: Context | null = null): void             // likewise

Http.request(method, url, body = "", headers = Map{}, timeout: Time.Duration | null = null,
             ctx: Context | null = null): Result[Http.Response, Error]   // and get and post alike; Err with the reason
Os.exec(program, args = [], dir = null, env = null, stdin = "",
        ctx: Context | null = null): Result[Os.Output, Error]     // the child is killed when ctx ends; Err with the reason
File.readLine(f: File.Handle, ctx: Context | null = null): Result[string | null, Error]
File.readAll(f: File.Handle, ctx: Context | null = null): Result[string, Error]
File.pipe(src: File.Handle, dst: File.Handle, ctx: Context | null = null): Result[void, Error]
```

`WaitGroup.wait` and `Time.sleep` keep their `void`: a call in statement position is unchanged,
and a function whose body ends in one still has type `void` (decision 6). The ticker a program
wants is a loop:

```ant
while !Context.isDone(ctx) {
  poll()
  Time.sleep(1000, ctx)
}
```

A context interrupts a read of a connection, a pipe or a terminal, which are the handles that
wait. A read of a regular file never waits, so there the context is checked only before the
read begins. `Http.request` keeps `timeout` beside `ctx`, since a one-off bound is the common
case and a context the composable one; when both are given, the earlier wins (decision 12).

### 3.4 Net (stdlib.md)

```
type Listener                                                    // opaque: a listening TCP socket

Net.listen(addr: string): Result[Net.Listener, Error]            // "host:port"; ":0" takes any free port
Net.address(l: Net.Listener): string                             // the address it listens on, "127.0.0.1:50321"
Net.accept(l: Net.Listener, ctx: Context | null = null): Result[File.Handle, Error]   // the next connection; Err when closed, or with ctx's reason
Net.close(l: Net.Listener): void                                 // pending and later accepts return Err; nothing twice
Net.dial(addr: string, ctx: Context | null = null): Result[File.Handle, Error]        // a connection to "host:port"
Net.peer(f: File.Handle): string | null                          // the other end of a connection; null for a file or a stream
```

A connection is a `File.Handle`, so `File.readLine`, `readAll`, `writeLine`, `writeText`, `pipe`
and `close` need nothing new, and a line protocol is a loop over `readLine`. A write to a
connection is flushed as a write to a standard stream is, since the other end is waiting for
it; `File.flush` stays for files. `Net.peer` is what a chat server prints. TCP only: TLS, UDP
and Unix sockets are left out (decision 10).

### 3.5 Http.serve (stdlib.md)

```
type Request = { method: string, path: string, query: string, headers: Map[string, string],
                 body: string, remote: string }

Http.serve(l: Net.Listener, handler: func(Http.Request): Http.Response,
           ctx: Context | null = null): Result[void, Error]
```

`serve` answers every request on `l` by calling `handler` on a goroutine of its own, as section
9's goroutines, and returns `Ok` once `ctx` is done and the requests in flight have finished,
waiting at most five seconds for them before it closes their connections. It is `Err` when the
listener is closed under it. A handler's panic is reported as a goroutine's panic is, and the
client gets status 500 with an empty body; so is a response whose status is outside 100 to 599.
The handler's `Http.Response` is sent as given, with `Content-Length` from the body; `path` is
decoded and `query` is raw, for `Url.parseQuery`; the headers hold one string per name, a
repeated header joined with `, `; `remote` is `ip:port`. Bodies are read whole, as the client
reads them, so streaming is left out, and so is HTTPS.

Taking a listener rather than an address is what lets a program know its port, run the server
on a goroutine, and talk to it, which the corpus does on `:0`:

```ant
func handle(req: Http.Request): Http.Response {
  if req.path == "/health" { Http.Response{ status: 200, headers: Map{}, body: "ok" } }
  else { Http.Response{ status: 404, headers: Map{}, body: `no ${req.path}` } }
}

func main(): void {
  var listener = Result.unwrap(Net.listen(":8080"))
  print(`listening on ${Net.address(listener)}`)
  var stop = Os.interrupt()
  var served = Http.serve(listener, handle, stop)
  print(`stopped: ${Context.reason(stop) ?? "?"}`)
}
```

### 3.6 Os.interrupt (stdlib.md)

```
Os.interrupt(): Context     // done when the process is interrupted; the same context at every call
```

An interrupt is Ctrl-C, SIGINT, and SIGTERM where the host has it. Today the first one ends the
run. With this plan, a program that has called `Os.interrupt()` receives the first interrupt
instead: the context is done with the reason `interrupt`, the program's waits end, and the
program goes on, to drain a server or write what it has, and exits when it is ready. A second
interrupt, or the first when no program asked, ends the run as today (decision 11).

The `ant` package gains `ant.WithInterrupts(c <-chan struct{})`: each value the host sends is
one interrupt, delivered by the rule above. The `ant` command sends one per signal. A host that
gives no channel keeps today's behaviour, where cancelling its context is the only stop.

### 3.7 Waiting without select

A goroutine that waits for several kinds of thing owns one channel, a mailbox, whose element
type is a union of them, and every producer sends into it. The spec's section 9 gains a
subsection, Waiting, with the patterns below, and they are corpus programs.

**An event loop.** A session's state machine in Go is a `for` around a `select` over a channel
per source. In Ant the session has one `events: Chan[Event]`, and each state is a loop over it
with an `is` chain:

```ant
type SpeechStarted = {}
type UserMessage = { text: string, synthetic: bool }
type Interrupt = { offsetMs: int }
type Event = SpeechStarted | UserMessage | Interrupt

func generating(s: Session, ctx: Context): void {
  for event in Chan.values(s.events, ctx) {          // ends when ctx is done
    if event is SpeechStarted {
      flushRenderer(s)
      Context.cancel(s.generation)
    } else if event is UserMessage {
      Array.push(s.pendingMessages, event)           // deferred to the next state
    } else if event is Interrupt {
      s.turn.interrupt = event
      return                                         // a transition is ordinary control flow
    }
  }
}
```

Every event is delivered to whatever state is current, which defers what it cannot handle, in
arrival order across all sources, where Go's `select` picks at random among the ready ones and
leaves the rest in their channels. The union lists every event the loop can see in one place. A
producer that cannot share the mailbox, because its channel exists already or is read elsewhere
too, is the case for `Select.wait` (section 3.2).

**The first of several.** One goroutine per source, one channel for them all, one receive, and
a context that stops the rest:

```ant
type Line = { text: string }
type Timeout = {}
type Event = Line | Timeout

func lineOrTimeout(input: File.Handle, ms: int): Event {
  var events: Chan[Event] = Chan.new()
  var ctx = Context.new()
  go {
    var got = Result.unwrapOr(File.readLine(input, ctx), null)
    if got != null { Chan.send(events, Line{ text: got }, ctx) }
  }
  go {
    Time.sleep(ms, ctx)
    Chan.send(events, Timeout{}, ctx)
  }
  var event = Chan.receive(events) ?? Timeout{}
  Context.cancel(ctx)
  event
}
```

**A bound on one wait.** A context with a timeout, made for the call:
`Chan.receive(results, Context.new(Time.duration(seconds: 2)))`.

**A worker pool.** A channel of jobs, `Chan.values` in each worker, a channel of results, a
`WaitGroup` to know when to close it:

```ant
type Job = { id: int, url: string }

func fetchAll(jobs: Array[Job], workers: int): Map[int, Result[Http.Response, Error]] {
  var queue: Chan[Job] = Chan.new()
  var results: Chan[Entry[int, Result[Http.Response, Error]]] = Chan.new(Array.length(jobs))
  var wg = WaitGroup.new()
  for w in Array.range(0, workers) {
    WaitGroup.add(wg, 1)
    go {
      defer { WaitGroup.done(wg) }
      for job in Chan.values(queue) {
        Chan.send(results, Entry{ key: job.id, value: Http.get(job.url, timeout: Time.duration(seconds: 10)) })
      }
    }
  }
  for job in jobs { Chan.send(queue, job) }
  Chan.close(queue)
  WaitGroup.wait(wg)
  Chan.close(results)
  Map.fromEntries(Iter.toArray(Chan.values(results)))
}
```

A future is a channel of capacity one. A ticker is the loop of section 3.3. A goroutine that
should stop checks `Context.isDone` between steps and passes the context into its waits; a
goroutine that ignores its context runs until the run ends, as in Go.

### 3.8 Considered and left out

- **A `select` statement.** It needs two keywords, a statement form with a binding per clause
  and a checker rule for each: the largest change to the grammar since the literal of Plan 5.
  `Select.wait` gives the same atomic choice as a call, and the mailbox gives an event loop the
  ordinary control flow a statement's clauses would have and a call's functions do not. If the
  call proves too heavy in practice, the statement can be sugar for it later.
- **`Future` or `Promise`.** A channel of capacity one, plus `Chan.receive`.
- **Timers and tickers as objects.** `Time.sleep` on a goroutine with a context.
- **A context on locks and conditions.** Go has none either; a lock held long enough to need
  one is the bug.
- **UDP, Unix sockets, TLS and HTTPS, WebSockets, HTTP/2 settings, streaming bodies, a body
  size limit.** Each is a plan when a program needs it; HTTPS needs a design for keys and
  certificates.
- **Signals other than the interrupt**, process groups, and a child's stdin and stdout as
  streams on `Os.exec`.
- **A deadlock detector.** A program whose goroutines all wait forever is stuck. Under the
  `ant` command Go's runtime may end the process with its own report and status 2; embedded, it
  hangs until the host cancels. The spec says a wait without a bound may never return; a
  context with a deadline is the remedy, and the corpus never waits unbounded.

## 4. Mechanisms

### 4.1 Context values

A `Context` is a `value.Context` sync object over `context.WithCancelCause` and
`context.WithTimeoutCause`, with the run's `in.ctx` as the default parent, so the host's
cancellation reaches every child. `Context.reason` reads `context.Cause`; the reasons are three
fixed errors, `cancelled`, `timeout` and `interrupt`, and a child whose parent is done reports
the parent's. `Context.wait` flushes standard output and selects on `Done`, as `Time.sleep`
does. The checker sees `type Context` as an opaque prelude type, like `Mutex`.

### 4.2 Chan values

A `Chan` is a `chan value.Value` with a mutex-guarded `closed` flag. `send` checks the flag,
then selects between the send and `ctx.Done()`; a close that races the check makes Go panic on
the send, which is recovered into the same ◆ message. `receive` selects between the receive and
`ctx.Done()`, and the `ok` of a Go receive tells a close from a value. `Chan.values` is a
sequence built with the `seq` helper of `iter.go`. The orderings of section 3.2 are Go's own
channel guarantees, so the race detector checks them for free.

`Select.wait` builds a `reflect.SelectCase` per clause, a receive on the context's `Done` when
there is one, and a default case for `otherwise`, flushes standard output, and calls
`reflect.Select`; the chosen clause's function then runs on the calling goroutine with the
received value, or `null` when its channel was closed. A send clause chosen on a closed channel
makes Go panic, which is recovered into the ◆. A `Select.Case` is an opaque reference value of
the sync kind, compared by identity, so it needs no new kind. `Select` joins the hand-written
list of prelude names in `names.go`, as `Regex` did in Plan 5.

### 4.3 Blocking with a context

One helper runs a wait against a context: it flushes standard output, then selects between the
operation's channel and `ctx.Done()`, and maps the cause to the reason. `Time.sleep`'s select
gains the context. `WaitGroup` gains a channel that is closed when the counter reaches zero and
replaced at the next `add` from zero, so a wait with a context is a select and a wait that gives
up leaks nothing. `Mutex` and `Cond` are untouched.

`Os.exec` builds its command on a context derived from the run's and the given one. `Http.request`
chains them: the run, then `ctx`, then `timeout`. `File.Handle`'s `*os.File` becomes an
`io.ReadWriteCloser` beside an optional deadline setter, which `*os.File` and `net.Conn` both
are; a read with a context on a handle that has one sets the read deadline to now when `ctx`
is done and clears it after, and reports the reason. The goroutine trick for standard input
stays, since a terminal honours no deadline on every platform.

### 4.4 Net

`Net.listen` is `net.Listen("tcp", addr)`, and `Net.accept` with a context sets the listener's
deadline to now when the context is done, as the reads do. `Net.dial` is `net.Dialer.DialContext`
under the run's context and the given one. A connection handle wraps the `net.Conn` with the
buffered reader and writer every handle has, flagged so that a write flushes. `Net.peer` reads
`RemoteAddr`.

### 4.5 Http.serve

`Http.serve` runs an `http.Server` with a ten-second header timeout on the listener, and when
the context is done calls `Shutdown` with a five-second bound. Each request's handler reads the
body whole, builds the `Http.Request` record, starts an Ant goroutine as a `go` block does,
through `spawn`, and waits for the `Http.Response` on a Go channel. A panic in that goroutine is
reported by `recoverRoot` as every goroutine's panic is, the channel stays empty, and the
handler answers 500. A response's status outside 100 to 599 is a ◆ in the handler goroutine,
which takes the same path. Because every context descends from the run's, a cancelled run shuts
the server down too, and its goroutines end at their next check.

### 4.6 Interrupts

`interp.Config` gains an `Interrupts <-chan struct{}`, which `ant.WithInterrupts` sets. The
instance watches it on a goroutine: a value goes to the context `Os.interrupt()` created, if
one exists and is not done, with the cause `interrupt`; otherwise it cancels the run. The `ant`
command replaces `signal.NotifyContext` with `signal.Notify` on `os.Interrupt` and
`syscall.SIGTERM`, relayed to that channel, so the second Ctrl-C still ends a program that
ignores the first. The REPL keeps `NotifyContext`.

### 4.7 Tests

Corpus programs under `tests/run`, tagged `9 Waiting`, `11 Synchronization`, `stdlib Net`,
`stdlib Http` and `stdlib Os`. Each wait has a bound of at most two hundred milliseconds, and a
test checks that no program under `tests/run` sleeps longer. The stress programs of section 2
run under `-race`. A select program waits on two channels of different types, a closed one, a
send clause and an `otherwise`, and checks afterwards that a value in a channel whose clause was
not chosen is still there; another checks that `Select.wait` returns `null` when its context
ends, and that a send clause on a closed channel is the ◆. A `Net` program listens on `:0`,
dials itself and exchanges lines on two goroutines; an `Http.serve` program does the same with
`Http.request`, cancels its context and prints that `serve` returned `Ok`; a third makes its
handler panic and checks the 500 and the report on standard error. The `/slow` route of Plan 5
serves the client's context cases.

A Go test under `tests` drives `ant.WithInterrupts`: a program that calls `Os.interrupt` gets one
interrupt, prints its reason and exits 0; a program that never calls it is ended by the first;
a program that calls it and ignores it is ended by the second. The lexer and parser are
untouched, so their fuzz tests are not.

## 5. Phases

Phases continue from Plan 5, which ended at 43. Each is one commit or a few, with the suite
green under `-race` and the coverage gate where it was.

- **Phase 44: spec.** Section 9 loses "there are no channels" and gains Waiting, a paragraph on
  the run's cancellation as it is today, and the three orderings. Section 11's Synchronization
  gains `Context`, `Chan` and `WaitGroup.wait`'s context; section 12's sync row names the two.
  `stdlib.md` gains `Net`, `Http.Request` and `Http.serve`, `Os.interrupt`, and the contexts on
  `Time.sleep`, `Os.exec`, `Http.request` and the `File` reads; section 11's list of I/O
  namespaces names `Net`, and its Synchronization gains `Select` beside `Chan`. One commit in the spec repository. The new headings go on
  `pendingSpec` and the new ◆ functions on `pendingPanics`. Done when every decision of section
  6 is written down and the suite is green with those lists.
- **Phase 45: Context.** The value, the six functions, `Time.sleep` and `WaitGroup.wait` with a
  context, and the `WaitGroup` broadcast channel. Corpus for every reason, for a child of a
  done parent, and for the hundred blocked receivers once `Chan` exists in the next phase.
  Done when `Context` has no pending entry.
- **Phase 46: Chan and Select.** The channel value and its seven functions, `Select` with its
  four and the `Case` value over `reflect.Select`, the stress programs, and the patterns of
  section 3.7 as corpus programs. Done when `Chan`, `Select` and `9 Waiting` have no pending
  entry and the stress programs are clean under `-race`.
- **Phase 47: contexts on I/O.** `Http.request`, `get` and `post`; `Os.exec` with a killed
  child; the `File.Handle` generalization and the three reads. Corpus with the `/slow` route, a
  cancelled `sleep 10`, and a read of a pipe given up on.
- **Phase 48: Net.** The listener, the six functions and connection handles with flushed
  writes. Corpus with the echo exchange.
- **Phase 49: Http.serve and the interrupt.** `Http.Request`, `serve` with its shutdown and its
  500s, `Os.interrupt`, `ant.WithInterrupts`, the command's signal loop, and the Go test for
  interrupts. Done when the three server programs and the interrupt test pass.
- **Phase 50: examples and documentation.** `examples/pool.ant` fetches a list of URLs with a
  worker pool and a deadline; `examples/echo.ant` serves lines over a socket and talks to
  itself; `examples/api.ant` is an HTTP API with a JSON body, a health route and a drain on
  Ctrl-C. `CHANGELOG.md`, the package documentation, the README and the status here.

## 6. Decisions

Open, each with the recommendation the plan is written to:

1. **Channels as a sync object, `Chan[T]`** (recommended), rather than no channels with a
   `Mutex`-and-`Cond` queue as the documented pattern. Section 9's "there are no channels" was
   written when shared memory was the only coordination; a queue is what every worker needs,
   and a sync object adds no syntax.
2. **A `Context` as the one way to bound or stop a wait**, taken by every blocking call as its
   last parameter (recommended), rather than a `timeout` on each call, which cannot stop
   anything, or a `Future` that carries its own cancellation.
3. **The name `Context`** (recommended), Go's, which the runtime documentation uses, rather than
   `Cancel`, `Signal` or JavaScript's `AbortSignal`.
4. **`Chan.receive` returns `T | null`** (recommended), with `Context.isDone` telling a
   cancellation from a close, as `Map.get` and `Array.pop` conflate, rather than a record or a
   union with a `Closed` and a `Cancelled` member on every receive.
5. **A send on a closed channel is a ◆** (recommended), as in Go, rather than `false`, since a
   value with no receiver is the sender's bug.
6. **`Time.sleep` and `WaitGroup.wait` keep `void`** and return early, the program asking the
   context (recommended), rather than returning `bool`, which would change the type of every
   function whose body ends in one.
7. **Locks and conditions take no context** (recommended), as in Go.
8. **`Http.serve` takes a `Net.Listener` and a `func(Http.Request): Http.Response`**, answers
   each request on its own goroutine, and turns a handler's panic into a 500 plus the usual
   report (recommended), rather than an address parameter, a response-writer object, or a
   panic that ends the server.
9. **Connections are `File.Handle`s whose writes flush** (recommended), rather than a `Net.Conn`
   type with its own read and write functions.
10. **TCP only** (recommended): no TLS, UDP or Unix sockets until a program asks.
11. **The first interrupt goes to the program when it asked for it, the second ends the run**
    (recommended), rather than interrupts always ending the run, as today, or a handler
    function called on a signal.
12. **`Http.request` keeps `timeout` and gains `ctx`**, the earlier bound winning
    (recommended), rather than `ctx` alone, which would make the one-off case verbose.
13. **A wait without a bound may never return**, and the spec says no more (recommended),
    rather than a deadlock detector in the interpreter.
14. **Left out**: a `select` statement, futures, timer objects, contexts on locks, UDP, Unix
    sockets, TLS, WebSockets, streaming bodies, other signals, a child's streams (recommended,
    section 3.8).
15. **`Context` before `Chan`** in the phases (recommended), since every channel wait takes a
    context, and `Chan`'s stress programs need `Context`'s cancellation to be bounded.
16. **`select` as a call, `Select.wait` over clauses with a function each, in a namespace of
    its own** (recommended), as Kotlin's coroutines have it, rather than a statement, or no
    select at all with the mailbox as the only fan-in. The call gives Go's atomic choice and
    the non-blocking try without syntax; the mailbox stays the shape for event loops, since a
    clause's function cannot `return` or `break` for the function around the wait.
17. **`Select.wait` returns `R | null`, `null` when its context ends** (recommended), as the
    rule of section 3.1 has it for every blocking call, rather than a `Select.done(ctx, f)`
    clause, which would make the context a case rather than a bound.

## 7. Status

Written on 2026-10-22 for review, after Plan 5 was done, and revised the same day after a
discussion of `select`: it is a call over clauses, `Select.wait`, beside the mailbox, which
stays the shape for event loops, and the state machine of section 3.7 came from that
discussion.

Phase 44 was done on 2026-10-22: one commit in the spec repository, with section 9's paragraph on
the host ending a run, the Synchronization paragraph naming channels and contexts, the Waiting
subsection with the mailbox, the first-of-several and the `Select.wait` examples, the three
orderings, the `Context`, `Chan` and `Select` blocks and `WaitGroup.wait`'s context in section
11, the sync row of section 12, and in `stdlib.md` the context rule in the opening, the `File`
reads and connections, `Http.Request` and `Http.serve`, `Net`, `Time.sleep`, `Os.exec` and
`Os.interrupt`. In this repository `pendingSpec` holds `9 Waiting` and `stdlib Net`, and
`pendingPanics` the seven new ◆ functions. Decisions 1 to 17 are written into the spec as
recommended, so they are taken with this phase.

Phase 45 was done on 2026-10-22: `value.Context` over `context.WithCancelCause`, with the
deadline from `WithTimeoutCause` layered on top so a cancel keeps its own cause, the six
functions in `internal/lib/context.go` beside the helpers every blocking call will use,
`WaitGroup` rebuilt on a channel closed at zero, `Time.sleep` and `WaitGroup.wait` with a
context, and the duration conversion shared with `Http.request`. The corpus program covers every
reason, a child of a done parent, and a goroutine stopped by a cancel; it names `9 Waiting`, so
that heading left the pending list a phase early.

Phase 46 was done on 2026-10-22: `value.Chan` over a Go channel with a closed flag, the seven
`Chan` functions, `Select` with `Select.Case` as an opaque reference and `wait` over
`reflect.Select` with the context as one more receive, the stress program, the patterns of
section 3.7 as a corpus program, and a program per ◆. One thing to know: a named function
declared `void` cannot end in `Chan.send`, since its body's value would be a `bool`, which is
the rule for every non-void call and not a channel matter; a `go` block and a lambda are not
affected, and a `var` takes the value where it matters. The bound on sleeps is a test,
`TestBoundedSleeps`, that refuses a `Time.sleep` above a second without a context under
`tests/run`; the rest of the bound is kept by the trailers, which fail a program that hangs.
