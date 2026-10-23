# Plan 8: the review

After Plan 7 the changes of Plans 5, 6 and 7 were reviewed: four independent reads of the lexer
and `Regex`, the library additions, the sync objects with the run's lifecycle, and I/O under a
context, with every finding reproduced by a program. Plan 7 held. Plan 5 had one divergence from
`stdlib.md` and a handful of edges. Plan 6's handles, contexts and channels had bugs the corpus
could not see, because no program there puts two goroutines on one handle, cancels a read of
standard input and reads again, pipes under a deadline, or closes a channel under a blocked
sender. This plan fixes what was found, in groups, and settles the specification where the
implementation had chosen silently.

## 1. Where we start

Handles. Every `File` call takes the handle's one mutex for its whole duration, the blocking read
included, so a goroutine looping on `readLine` and another calling `writeLine` on the same
connection deadlock on a Go mutex no context can end, and `File.close` from another goroutine
cannot end a pending read. A connection read ended by its deadline returns the bytes consumed so
far with the error, and both read closures discard them, so the next line is corrupted. A
context-ended read of standard input marks the stream closed, so every later read fails, and the
abandoned goroutine keeps running: with `File.pipe` it writes into the destination while the
program does, a data race in the runtime's own state. `File.pipe` ended by a context skips the
flush, so what it copied to standard output is lost.

Contexts and channels. `Context.new` with a timeout wraps the cancel layer in the timeout layer,
so when the deadline passes only the inner layer unregisters and the outer stays in the run's
context until `Context.cancel`, which the spec's own idiom never calls: one entry per call for the
run's lifetime. A blocking call with a context that is done already proceeds about half the time,
since Go's `select` chooses at random among ready cases, where the spec says it gives up.
`Context.deadline` builds its instant from nanoseconds, which overflow past 2262, so a far
deadline reads as a date in the 1700s. `Chan.new`, `String.padStart`, `String.padEnd`,
`Crypto.randomBytes` and the older `String.repeat` pass a count straight to `make`, and a huge one
kills the process with a Go stack. A channel is closed under a blocked sender by closing the Go
channel and recovering the runtime's panic, which the race detector reports as a data race, so a
correct Ant program fails under `-race`. The host-cancel path of the run sets the stop flag after
the cancellation, the window the relay fix of Plan 7 closed on its own path. A run whose host
context expires reports `timeout` where section 9 promises `cancelled`. An interrupt that ends the
run prints Go's `context canceled` and exits 1, as a panic does.

Http and Net. `Http.request` uses Go's default client, which follows redirects and resends a POST
as a GET, where the spec says one request and any status is `Ok`. `Http.Request.headers` lacks
`Host`, which Go moves out of the header map. An empty method is sent as `GET`. A 1xx status
passes the ◆ check but cannot be sent as a final response. After the five-second drain `serve`
returns without closing the connections still open, and the shutdown closes the Go listener
without marking the `Listener` closed. Concurrent acceptors share one listener deadline, so one
caller's context ending gives the other a raw `i/o timeout`. `Url.stringify` drops a userinfo
whose escape is malformed.

Regex and the library's edges. Duplicate group names give `null` in `named` even when the group
matched. The lexer's class tracking is a flag, so the `]` of a POSIX name like `[:alpha:]` ends
the class and a following `/` ends the literal. `Regex.split` of an empty pattern on an empty
string gives `[]`. A compiled source holding a raw line break renders as a literal that reads back
to a different value, and so does the empty source. `Encoding.fromBase64` strips any number of
trailing `=`. `Math.clamp` with a NaN bound passes the ◆ check. `Csv.parse` rewrites a CRLF
inside quotes, `Url.stringify` writes `query` verbatim, `Time.sleep` of a negative count returns
at once, `Os.exec` kills the child and not what it started, `\/` inside `\Q...\E` is two
characters, `\10` and up are octal: each fine, none said.

The checker, before any of these plans: a lambda inside a generic call is unified against the
expected return with any unbound parameter in reach, the enclosing function's included, so a
lambda whose body has the wrong type passes inside a generic function and a checked program
returns a value of the wrong type.

Documentation. `WithInterrupts` was inserted between the doc comment of `WithEnvironment` and its
declaration. The Plan 6 entry of the CHANGELOG still names the `Select` namespace among the calls
that take a context. The unchecked path of a select head says "takes 1 arguments".

## 2. Definition of done

- Two goroutines may read and write one handle; a read ended by a context keeps what it consumed
  for the next read; a read of standard input ended by a context leaves the stream readable and
  the line it was waiting for arrives at the next read; `File.pipe` delivers what it copied.
- A context with a timeout holds nothing in the run after its deadline; a blocking call with a
  done context gives up at once; `Context.deadline` is right for any deadline; a count beyond what
  the runtime can hold is a ◆ with a row in the table; the race detector is quiet on a channel
  closed under a blocked sender; the run's end has no window on either path; the host's deadline
  reads `cancelled`; an interrupt that ends the run exits 130 and prints nothing.
- `Http.request` sends one request; `Host` is a header; an empty method is `Err`; the status
  bound is 200 to 599; `serve` closes what the drain left; a listener the shutdown closed reads as
  closed; acceptors with contexts of their own do not disturb each other; a userinfo is written as
  given.
- The regex and library edges above are fixed or said in the spec.
- The checker binds only the call's own type parameters when it unifies a lambda's body.
- The documentation defects are fixed, the suite is green under `-race`, and the gate holds.

## 3. The fixes

### 3.1 Handles

A handle gets three locks in place of one. A read slot, a channel of capacity one, is acquired
with the caller's context, so a second reader waiting behind a first gives up when its context
ends. A write mutex serializes writers. A state mutex guards the closed flag and the closer, and
`File.close` takes only that: it closes the file or connection, which ends a pending read or write
with an error on the other goroutine, and flushes a file's buffer first, since a file's writes
never block.

What a read consumed before its context ended stays in the handle, a pending string prepended to
the next read of any kind: `readLine` takes a line from it when it holds one, `readAll` returns it
with the rest, `pipe` writes it first. Standard input is read through a pump, a goroutine that
blocks in the real read and hands chunks over a channel; the handle's reader takes chunks from the
channel and gives up on the caller's context without taking a chunk, so nothing is lost and the
stream stays open. The pump starts at the first read and is the one goroutine a terminal can hold.
A connection is ended through its read deadline as before. `pipe` flushes the destination whatever
ended the copy.

### 3.2 Contexts, channels and the run

`Context.new` builds the timeout layer on the parent and the cancel layer on that, so the
deadline cancels both and unregisters both, and `Context.cancel` cancels the inner layer with its
cause and stops the timer. Every blocking call looks at its context before it waits, so a done
context is a give-up, never a coin. `Context.deadline` builds its instant from seconds and
nanoseconds. A count argument above 2^31 − 1 is a ◆, one row of the table for `String.repeat`,
`String.padStart`, `String.padEnd`, `Crypto.randomBytes` and `Chan.new`.

A channel keeps a closed signal beside its Go channel, a second channel closed under the mutex,
and the Go channel of values is never closed: a send selects on the value channel and the signal,
and panics on the signal; a receive selects on both and, on the signal, drains the value channel
without waiting before it answers `null`. The `select` keyword does the same with two cases per
channel clause. No runtime panic is recovered anywhere, and the race detector sees ordinary
channel operations.

A goroutine whose wait ends because the run's own context is done ends itself there, before it
returns to the program, on both paths of the run's end. `ReasonOf` maps Go's deadline error,
which only the host's context can carry, to `cancelled`. The relay records that it ended the run,
`Run` returns `ErrInterrupted`, and the command exits 130 without a message, as a shell does.

### 3.3 Http and Net

A client of its own with `CheckRedirect` returning the response as it is. `Host` goes into the
request's headers from Go's field. An empty method is `Err` before anything is sent. The status
bound becomes 200 to 599 in the spec and the check, since a 1xx is an interim response Go follows
with a final one. After the drain `serve` calls `Close` on the server, and marks the `Listener`
closed, since the shutdown closed its socket. A listener accepts through a pump of its own, one
goroutine started at the first `Net.accept` that hands connections over an unbuffered channel;
callers select on it and their context, so a context ending touches no one else, and a connection
accepted while no caller waits is held for the next. `Url.stringify` keeps a userinfo whose
escape does not decode as written, escaped as a whole.

### 3.4 Regex and the library's edges

The first participating group of a name fills `named`. The lexer tracks a POSIX name inside a
class. `Regex.split` answers `[""]` for an empty string whatever the pattern. The normalized
source writes a raw line break as `\n` and an empty source as `(?:)`, so the rendering reads back
to the value and `==` holds, said in section 11. `Encoding.fromBase64` accepts at most two `=` and
only at the end. A NaN bound in `Math.clamp` is a ◆. The spec says that a line break inside a
quoted CSV field is `\n`, that `query` is written as given, that a negative sleep sleeps nothing,
that `Os.exec` kills the child and not its descendants, that `\/` inside `\Q...\E` is literal,
that `\10` and up are octal escapes, that a program which cancels its own interrupt context has
handled the interrupt, and that the command exits 130 on an interrupt it did not ask for.

### 3.5 The checker

When a lambda's synthesized result is unified against the expected result, only the call's own
type parameters may be bound; a parameter of the enclosing function is a fixed type there, so a
body of the wrong type is an error, as a `var` of that type already was.

## 4. Phases

- Phase 56: the plan; the handles of 3.1 with corpus programs for the reader and writer, the kept
  partial line, the pipe under a deadline, and a Go test for the pump.
- Phase 57: contexts, channels and the run of 3.2, with the table's count row and its program,
  programs for the done context and the far deadline, and the command's exit.
- Phase 58: Http, Net and Url of 3.3, with programs for the redirect, the host header, the drain's
  close and two acceptors.
- Phase 59: Regex and the edges of 3.4, in the spec and the implementation, with programs.
- Phase 60: the checker of 3.5, with a check program.
- Phase 61: the documentation, the CHANGELOG, the status; the suite under `-race`; the gate.

## 5. Status

Phase 56 was done on 2026-10-23: a handle has a read slot taken with the caller's context, a
write mutex and a state mutex, and `File.close` takes the last alone, so a reader and a writer
share a connection and a close from a third goroutine ends a read in progress; what a read
consumed before its context ended it stays in the handle's pending text for the next read of any
kind; standard input is read through a pump, one goroutine that hands chunks over a channel and
whose Read gives up on the reader's context without taking one, so the stream stays open and the
late line arrives whole; `File.pipe` writes the pending text first and flushes the destination
whatever ended the copy. `stdlib.md` says so under File. Three corpus programs and a Go test
with a late standard input cover the reader and writer, the kept partial line, the pipe under a
deadline and the resumed read; the suite is green under `-race`.

Phase 57 was done on 2026-10-23: `Context.new` builds the timeout layer on the parent and the
cancel layer on it, so an expired deadline unregisters both, which a Go probe of ten thousand
expired contexts confirms; `Chan.send`, `Chan.receive`, `Chan.values`, `WaitGroup.wait` and
`Time.sleep` look at their context before they wait; `Context.deadline` builds its instant from
seconds and nanoseconds, checked by a Go test since the harness bounds every run; a count above
2^31 − 1 is the table's new row, with `boundCount` on the five functions; a channel keeps a
closed signal beside a channel of values that is never closed, so a send blocked under a close
ends on the signal and a receive drains the buffer before `null`, in `Chan` and in the `select`
keyword alike, and the race detector is quiet; `AfterWait` ends a goroutine whose wait the run's
own context released, called by every blocking call, so neither path of the run's end has a
window; Go's deadline error, which only the host's context can carry, reads `cancelled`; the
relay records that it ended the run, `Run` returns `ErrInterrupted`, and the command exits 130
in silence. `stdlib.md` says so of `Os.interrupt`, with the self-cancelled context. Four corpus
programs and the interrupt test's assertion cover it; the suite is green under `-race`.

Phase 58 was done on 2026-10-23: `Http.request` sends through a client of its own that follows no
redirect, so a 3xx is the response with its `Location`; an empty method is `Err`; `Host` is in a
handler's headers; the status bound is 200 to 599, since a status below 200 is an interim
response; after the drain `serve` closes what is left open and marks the listener closed; a
listener accepts through a pump, one goroutine from the first `Net.accept`, so acceptors with
contexts of their own do not disturb each other and a connection accepted while no one waits is
held for the next; `Url.stringify` keeps a userinfo whose escape does not decode, escaped as a
whole. `stdlib.md` says so under Http and Net. Two corpus programs cover the redirect, the
empty method, the host header, the closed listener and the two acceptors; the reviewer's
shutdown probe shows the late client cut at five seconds; the suite is green under `-race`.
