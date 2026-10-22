# Plan 7: select

Plan 6 made waiting on several channels a call: `Select.wait` over clauses, each clause a
channel and a function, in the shape Kotlin's coroutines use. It works, and it has the cost the
plan named when it chose it: a clause's function returns from itself, so a `return` or a `break`
that depends on which clause ran has to travel through the call's result, and an event loop in
that form is a loop around a value. Go's `select` and the state machines written with it have no
such detour, because each case is a block in the function's own body.

This plan makes `select` a keyword. A clause is the blocking call a program would have written,
followed by a block; the `select` performs the one that can proceed first and runs its block,
which returns, breaks and continues like any other. The call form goes away, so there is one way.

## 1. Where we start

Section 9 of `ant-lang.md` has Waiting: every blocking call takes a context, a mailbox is the
shape for an event loop, and `Select.wait` is the call for channels that exist already. Section
11 declares `Select.receive`, `Select.send`, `Select.otherwise` and `Select.wait` with the opaque
`Select.Case[R]`. In the implementation `internal/lib/chan.go` builds the clauses as
`value.SelectCase` values and `Select.wait` runs one `reflect.Select` over them and the context,
then calls the chosen clause's function on the calling goroutine. The corpus covers it in
`select_lib.ant` and two panic programs, and the spec's Waiting example writes a loop with a
`Continue | Stop` result.

The lexer has the keywords of section 2 in one table, and the parser reads a header, the
expression before a block in an `if`, `while` or `for`, up to the first `{` that is not an
operand: one that follows an identifier path, a literal, or a closing `)` or `]`. Braces mean a
block after `else`, `go`, `defer`, `func(...)` and at the end of a header, a value after a type
name, and a lambda anywhere else. The checker types an `if` as the union of its branches or the
expected type. The interpreter's `Chan`, `Context`, `WaitGroup` and `Time.sleep` all wait on Go
channels: a channel's own, a context's `Done`, the channel a `WaitGroup` closes at zero, a timer.
That is what makes each of them a possible clause, and what keeps an I/O call from being one.

No program in the corpus or the examples uses `select` as a name.

## 2. Definition of done

- `select` is a keyword in section 2, a subsection of section 5 defines it, Waiting in section 9
  uses it, and the `Select` namespace is gone from section 11 and the Values table of section 12.
- The parser reads every form of section 3, the checker types it, the interpreter runs it, and the
  three gates have no pending entry; every syntax error of section 3 has a program under
  `tests/syntax`, every checker error one under `tests/check`.
- The corpus programs and the spec examples that used `Select.wait` are rewritten with the
  keyword, and the namespace, its values and its tests are removed.
- The suite is green under `-race`, the coverage gate where it was, and the parser's fuzz seeds
  include selects.
- `CHANGELOG.md`, the README and the package documentation.

## 3. The design

### 3.1 Syntax

```ant
select {
  var line = Chan.receive(lines) { print(line ?? "closed") }
  Chan.send(out, batch) { batch = [] }
  Time.sleep(1000) { flush() }
  Context.wait(ctx) { return }
  else { idle += 1 }
}
```

A `select` is the keyword, then clauses between braces, one per line or separated by `;`. A
clause is a head and a block. A head is one of five blocking calls, written as the call would be
and not performed: `Chan.receive(c)`, `Chan.send(c, v)`, `Context.wait(ctx)`, `WaitGroup.wait(wg)`
or `Time.sleep(ms)`. `var x =` before a head binds the call's result in the block, so a receive
binds `T | null`, `null` when the channel is closed. An `else` block, written last, runs when no
head can proceed at once. The block of a clause begins at the head's `{`, which the header rule
of section 5 finds already: a `{` after a `)` ends a header.

The event loop of section 9, which Plan 6 wrote as a mailbox, reads as Go's `for` and `select`
with the keyword, one channel per source:

```ant
func generating(s: Session, ctx: Context): void {
  while true {
    select {
      Context.wait(ctx) { return }
      Chan.receive(s.speechStarted) {
        flushRenderer(s)
        Context.cancel(s.generation)
      }
      var message = Chan.receive(s.userMessages) {
        if message != null { Array.push(s.pendingMessages, message) }
      }
      var interrupt = Chan.receive(s.interrupts) {
        if interrupt != null { s.turn.interrupt = interrupt }
        return
      }
    }
  }
}
```

A `select` is an expression, as an `if` is:

```ant
var event: Event = select {
  var line = Chan.receive(lines) { if line == null { Closed{} } else { Line{ text: line } } }
  Time.sleep(1000) { Timeout{} }
}
```

### 3.2 Semantics

The heads' operands evaluate first, in order, once: the channels, the value to send, the
context, the group, the count. Then the `select` waits until one head can proceed, performs it,
binds its result, and runs its block; the value of the `select` is the block's value. A head can
proceed when its channel has a value or is closed, when its channel can take the value, when
its context is done, when its group's counter is zero, or when its time has passed. When
several can proceed at once, one is chosen at random, so no source starves. With an `else`
clause the `select` never waits: when no head can proceed at once, `else` runs. A `select` with
no `else` and heads that can never proceed waits until the run is cancelled, as any wait does.

A head takes no context: the context is its own clause, `Context.wait(ctx) { ... }`, and the
run's cancellation ends a `select` as it ends every wait. A closed channel's receive is always
ready, so a loop that keeps one in its `select` spins; the block leaves the loop on `null`. A
send chosen on a closed channel is a ◆, as `Chan.send` is. The block's `return`, `break` and
`continue` act on the function or loop around the `select`, since the block is a block of that
body.

### 3.3 Typing

A head is typed as the call it names: a receive binds `T | null`, a send `bool`, the rest `void`,
and a binding on a head that returns `void` is a compile error. The head must be one of the five
calls, with exactly its required arguments: a context argument on a head is a compile error that
names the `Context.wait` clause. The blocks are checked against the expected type when there is
one, and otherwise the `select` has the union of their types with `never` dropped, as an `if`;
without `else`, the union is of the heads' blocks alone, since one of them runs. A bound variable
is a local of its block and narrows there like any local.

### 3.4 Syntax errors

A `select` needs at least one head: a body with no clause, or with only `else`, is a syntax
error. Two `else` clauses, an `else` that is not last, a binding on `else`, a head that is not a
call of one of the five, and a clause whose `{` is not on the head's line are each a syntax error
with a message naming the rule.

### 3.5 What goes away

`Select.receive`, `Select.send`, `Select.otherwise`, `Select.wait` and `Select.Case` leave
section 11, the sync row of section 12, the prelude, the name list and the library. The keyword
covers every use they had but one: a `select` over a number of channels known only at run time,
which the call took as an array. That use is left out (section 3.6).

Waiting in section 9 keeps the mailbox as a pattern, since producers that can share a channel
still make the simplest loop, and shows the keyword for the loop over a session's own channels.
The paragraph on channels that exist already, and the `Continue | Stop` example, go.

### 3.6 Considered and left out

- **A clause syntax of its own**, `receive x from c` or Go's `<-`. The call a program would have
  written is already the name of the wait, and nothing new has to be learned to read a clause.
- **I/O calls as heads**: `Net.accept`, `File.readLine`, `Http.request`. They cannot be begun
  without committing to them, so a `select` over them would consume what it did not choose. They
  stay what they are, calls with a context.
- **A dynamic `select`**, over an array of channels. Go has it only through `reflect`; a
  `Chan.any(cs, ctx)` over channels of one type is the candidate when a program asks.
- **Priority among ready clauses**, Kotlin's bias to the first. A program that wants one checks
  it first with a `select` whose `else` falls through to the real wait.
- **`default` and `case` as keywords.** `else` already means "when nothing else", and the heads
  need no introducer.

## 4. Mechanisms

### 4.1 Lexer and parser

`select` joins the keyword table; it is not an operand, so a `/` after it begins a pattern, and
it cannot end a line. In `primary`, the keyword begins a `Select` node. The parser reads the body
as it reads a block, clause by clause: `else` and a block; or `var`, a name and `=`, then a
header, which the existing header reader ends at the first `{` that is not an operand, and the
block; or a header and the block. The head must parse as a call whose callee is an identifier
path of two parts naming one of the five; since no declaration may shadow a prelude name, the
path is the function. The parser counts heads and `else` clauses and reports the errors of
section 3.4 at the clause's position.

### 4.2 AST and resolver

`ast.Select{Pos, Clauses}` with `ast.SelectClause{Pos, Name, Slot, Head *Call, Body *Block,
Else bool}`. The resolver opens the clause's scope with the bound name, as it does for the
variable of a `for`, so the interpreter stores the result in a slot and the checker sees a local.

### 4.3 Checker

A `Select` case in `synth` and `check`: each head is typed through the call checker as the call it
is, which gives the binding's type and rejects a wrong argument; the head's callee and argument
count are checked against the five forms; each block is checked against the expected type or
synthesized; the result is the union with `never` dropped. The bound variable enters the block's
scope with the head's type.

### 4.4 Interpreter

A `Select` evaluates each head's operands in order, then builds one `reflect.SelectCase` per head
from what Phase 46 built them from, a receive on the channel, a send with the value, a receive
on the context's `Done`, on the group's zero channel, or on a timer, plus a default for `else`
and a receive on the run's context. `reflect.Select` chooses; the run's case takes the
cancellation path every wait takes; a send chosen on a closed channel is recovered into the ◆;
the chosen clause's result is stored in its slot and its block evaluated. `doSelect` and the
closed-send recovery move from `chan.go` to the interpreter. The library loses the `Select`
namespace and `value.SelectCase`.

### 4.5 Tests

Syntax programs for each error of section 3.4; check programs for a binding on a `void` head, a
context argument on a head, and a block whose type does not fit the expected one. Run programs
for the five heads, a receive that binds `null` on a closed channel, `else` with nothing ready,
the random choice leaving the unchosen value in its channel, a `select` as an expression with an
annotation and without, `break` and `return` from a block, and the state machine of section
3.1 driven by goroutines; a panic program for the closed send, under a new row of the panics
table. The parser's fuzz seeds gain selects, and `TestSpecBlocks` reads the new examples.

## 5. Phases

Phases continue from Plan 6, which ended at 50.

- **Phase 51: spec.** `select` in section 2's keywords and in the brace rule of section 5; a
  `select` subsection in section 5 after Loops; Waiting in section 9 rewritten; the `Select`
  namespace out of section 11 and the Values row; the panics table's row for the closed send.
  One commit in the spec repository; the new heading on `pendingSpec`, the row on
  `pendingPanics`, and the `Select` corpus programs kept until Phase 54.
- **Phase 52: lexer, parser, AST and resolver**, with the syntax programs and the fuzz seeds.
- **Phase 53: checker**, with the check programs.
- **Phase 54: interpreter**, the run programs, the migration of the corpus and the examples, and
  the removal of the `Select` namespace. Done when the pending lists are empty.
- **Phase 55: documentation.** `CHANGELOG.md`, the README, the package documentation and the
  status here.

## 6. Decisions

Open, each with the recommendation the plan is written to:

1. **`select` as a keyword with `else` as the clause for nothing ready** (recommended), rather
   than Go's `case` and `default`, which would be two more keywords for no more meaning.
2. **A head is the blocking call as written, not performed** (recommended), rather than a clause
   syntax of its own or Go's `<-`.
3. **Five heads**: `Chan.receive`, `Chan.send`, `Context.wait`, `WaitGroup.wait` and
   `Time.sleep` (recommended), which are Kotlin's `onReceive`, `onSend`, cancellation, `onJoin`
   and `onTimeout`, rather than channels only, or I/O calls that cannot be begun without being
   taken.
4. **`var x =` binds a head's result** (recommended), rather than `as x` or a bare `x =`.
5. **A `select` is an expression typed like an `if`** (recommended), rather than a statement.
6. **Random choice among ready clauses** (recommended), as Go's and as `Select.wait` had, rather
   than Kotlin's bias to the first clause, which its own documentation shows starving the second.
7. **Heads' operands evaluate once, in order, before the wait** (recommended), as Go's do.
8. **The context is a clause and heads take none** (recommended), rather than a context on the
   `select` itself or on each head.
9. **The `Select` namespace is removed** (recommended), so there is one way, rather than kept for
   a dynamic clause list, which is left out with `Chan.any` as the candidate.
10. **A closed channel's receive is ready with `null`, and a send chosen on a closed channel is
    a ◆** (recommended), as the calls themselves behave.

## 7. Status

Written on 2026-10-22 for review, after Plan 6 was done and a look at Kotlin's `select`
expression.

Phase 51 was done on 2026-10-23: one commit in the spec repository, with `select` among the
keywords and in the brace rule, the `select` subsection of section 5, Waiting rewritten around
the keyword with the mailbox kept as a pattern, the `Select` namespace out of sections 11 and 12,
a `select` bullet in section 12's expressions and a row in its panics table. Here `5 select` is on
`pendingSpec`, the row's key `select-closed` on `pendingPanics` and `panicRows`, the three new
spec blocks on the parser's `pendingBlocks`, and the two panic programs of the namespace keep
only the table's key until Phase 54 removes them. Decisions 1 to 10 are written into the spec as
recommended, so they are taken with this phase.

Phase 52 was done on 2026-10-23: `select` among the keywords, the `Select` and `SelectClause`
nodes, the parser reading a clause as a header and a block with the five heads checked by their
path, the resolver giving each clause's body a scope of its own with the binding, as a loop's
iteration has, the printer, seven syntax programs and the parser's cases for the errors of section
3.4, and fuzz seeds with selects. One thing the parser had to know: a line that starts with `else`
continues the line before it, so no terminator separates a clause's block from the `else` clause.
The three spec blocks parse and left `pendingBlocks`, and `5 select` left `pendingSpec`.
