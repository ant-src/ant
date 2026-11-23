<!-- status: reviewed -->

# Ant

Ant is a general-purpose, garbage-collected language with JavaScript-like syntax, static types,
control flow made of expressions, and goroutines and channels for concurrency. A file of statements
runs top to bottom; there is no `main`. Data is records and tagged unions, expected failures are
`Result` values, and absence is `null`, which goes only where a type allows it. One binary, `ant`,
runs programs, checks types, formats code, runs tests, and prints docs. The standard library is
always there, with no import: arrays, strings, time, files, processes, HTTP, JSON, and more.

## A complete program

```ant
type Account = { owner: string, balance: int, locked: bool = false }
type Transfer = { from: string, to: string, amount: int }

type NotFound = { name: string }
type Locked = { owner: string }
type Short = { owner: string, missing: int }
type TransferError = NotFound | Locked | Short

func apply(accounts: Map[string, Account], t: Transfer): Result[void, TransferError] {
  let source = accounts[t.from] ?? return Result.err(NotFound{ name: t.from })
  let target = accounts[t.to] ?? return Result.err(NotFound{ name: t.to })
  if source.locked { return Result.err(Locked{ owner: t.from }) }
  if source.balance < t.amount {
    return Result.err(Short{ owner: t.from, missing: t.amount - source.balance })
  }
  source.balance -= t.amount
  target.balance += t.amount
  Result.ok(void)
}

func describe(e: TransferError): string {
  if e is NotFound { "no account named ${e.name}" }
  else if e is Locked { "${e.owner} is locked" }
  else { "${e.owner} is short by ${e.missing}" }
}

let accounts = Map{
  ada: Account{ owner: "ada", balance: 100 },
  linus: Account{ owner: "linus", balance: 20 },
  grace: Account{ owner: "grace", balance: 40, locked: true },
}

let transfers = [
  Transfer{ from: "ada", to: "linus", amount: 30 },
  Transfer{ from: "linus", to: "ken", amount: 5 },
  Transfer{ from: "grace", to: "ada", amount: 10 },
  Transfer{ from: "linus", to: "ada", amount: 80 },
]

for t in transfers {
  let result = apply(accounts, t)
  let outcome = if result is Err { describe(result.error) } else { "done" }
  println("${t.from} -> ${t.to} ${t.amount}: ${outcome}")
}

let balances = accounts.values()
  | Array.sortBy({ -x.balance })
  | Array.map({ "${x.owner} ${x.balance}" })
  | String.join(", ")
println(balances)
```

```text
ada -> linus 30: done
linus -> ken 5: no account named ken
grace -> ada 10: grace is locked
linus -> ada 80: linus is short by 30
ada 70, linus 50, grace 40
```

Save it as `main.ant`, then run `ant check main.ant` to check its types and `ant run main.ant` to
run it. The program shows most of what makes Ant code look the way it does:

- `TransferError` is a union of three records, and `is` tells them apart. After `if e is NotFound`,
  the checker knows `e` is a `NotFound`, so `e.name` is a field it can check.
- `apply` returns a `Result`: `Ok` on success, `Err` with a `TransferError` on failure. A failure is
  an ordinary value, and the caller decides what it means.
- `accounts[t.from]` is `null` when the key is missing, and `?? return` leaves the function right
  there. The guards come first, and the main path reads straight down.
- `describe` has no `return`: an `if` chain is an expression, and a function returns its last value.
  `outcome` is chosen the same way.
- `|` passes a value as the first argument of the next call, so a chain of transformations reads
  from top to bottom. `accounts.values()` is `Map.values(accounts)`, written through `.`.
- `let` names a value you never reassign. The records it holds can still change, as `apply` changes
  the balances.

## Principles

Four rules shape the language. The design note linked from each of the first three explains the rule
and what it costs.

- **[No implicit coercion.](design/no-coercion.md)** A value never changes type without a conversion
  you can see. With `n` an `int`, `ant check` rejects `"n = " + n`; write `"n = ${n}"`. `n + 1.5`
  needs `float64(n)`, and a condition must be a `bool`.
- **[Everything is an expression.](design/expressions.md)** Blocks, `if`, and loops produce values,
  so a name can be bound once, to the result of a choice, instead of assigned in two branches, and
  markup chooses what to show with an ordinary `if`.
- **[Errors are values.](design/errors-are-values.md)** `Result` carries an expected failure, such
  as a missing file, and `null` carries absence. `try` passes a failure up to the caller. `panic` is
  for bugs, such as an index out of range.
- **Minimal syntax.** The parser reads one token ahead and never asks the type checker what a name
  means, so you can see the shape of a line without knowing the types of its names.

## Where to go next

- **New to Ant.** [Install](start/install.md) `ant`, write
  [your first program](start/first-program.md), then take the [tour](tour/01-hello.md): one concept
  and one program per page, about three minutes each.
- **From JavaScript or TypeScript.** [Ant for JavaScript developers](coming-from/javascript.md) maps
  the syntax you know and covers the differences that bite: no exceptions, no truthiness, no
  `async`.
- **From Go.** [Ant for Go developers](coming-from/go.md) maps goroutines, channels, `defer`, and
  the standard library, and shows where unions and `Result` take the place of most interfaces,
  `nil`, and a second return value.
- **From Python.** [Ant for Python developers](coming-from/python.md) puts a script side by side
  with its Python version and covers static types, `Result` in place of exceptions, and goroutines
  in place of threads.
- **Scripting.** [`ant run`](tools/run.md) runs a file of statements, and `ant` alone opens the
  REPL. The how-to guides solve one task each, such as
  [reading and writing a file](howto/files/read-write.md),
  [reading standard input](howto/cli/stdin.md), [running a command](howto/processes/run.md), and
  [reading the environment](howto/cli/env.md).
- **Embedding.** The Go package
  [`github.com/ant-src/ant`](https://pkg.go.dev/github.com/ant-src/ant) loads, checks, and runs Ant
  programs inside a Go program, and calls their functions. [Embedding Ant](guide/embedding.md)
  builds a host step by step, and the package's documentation is on pkg.go.dev. A host declares its
  namespaces in a [declaration file](guide/modules.md#declaration-files), a `.d.ant` file kept
  beside its scripts, so `ant check` and the editor work on those scripts with no host running.

To look something up, the [language guide](guide/types.md) explains each area and when to pick what,
the [language reference](language/index.md) has a page for every keyword, operator, literal, and
built-in type, and the [standard library](lib/index.md) has a page for every function and type.
[Idiomatic Ant](guide/idiomatic.md) shows how Ant code is written.
