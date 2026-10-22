# Ant

A Go implementation of [Ant](https://github.com/ant-src/spec): a tree-walking interpreter, a
type checker, and a REPL, usable as a command and as a Go module.

The interpreter implements section 12 of the specification and the whole standard library; the
checker implements the rules of sections 4 through 10. Programs under `tests/run` are the
acceptance suite for `ant run`, programs under `tests/check` for `ant check`.

## Example

```ts
type Account = { id: string, balance: int }
type Transfer = { from: string, to: string, amount: int }

type NoAccount = { id: string }
type Insufficient = { id: string, short: int }
type TransferError = NoAccount | Insufficient

func withdraw(accounts: Map[string, Account], id: string, amount: int): Result[void, TransferError] {
  var account = accounts[id] ?? return Result.err(NoAccount{ id })
  if account.balance < amount {
    return Result.err(Insufficient{ id, short: amount - account.balance })
  }
  account.balance -= amount
  Result.ok(void)
}

func transfer(accounts: Map[string, Account], t: Transfer): Result[void, TransferError] {
  var target = accounts[t.to] ?? return Result.err(NoAccount{ id: t.to })
  try withdraw(accounts, t.from, t.amount)
  target.balance += t.amount
  Result.ok(void)
}

func explain(e: TransferError): string {
  if e is NoAccount { `no account named ${e.id}` }
  else if e is Insufficient { `${e.id} is short by ${e.short}` }
}

func main(): void {
  var accounts = Map{
    ada: Account{ id: "ada", balance: 100 },
    bob: Account{ id: "bob", balance: 20 },
  }
  var transfers = [
    Transfer{ from: "ada", to: "bob", amount: 30 },
    Transfer{ from: "bob", to: "ada", amount: 500 },
    Transfer{ from: "ada", to: "eve", amount: 1 },
  ]
  for t in transfers {
    var result = transfer(accounts, t)
    if result is Err { print(explain(result.error)) }
  }
  print(Map.values(accounts) | Array.map({ `${x.id}: ${x.balance}` }))
}
```

## Command

```
go install github.com/ant-src/ant/cmd/ant@latest

ant run main.ant [args...]    # run a program; exit 0, 1 on a panic, 2 when it cannot load
ant check main.ant            # type-check it and every module it imports; exit 1 on errors
ant                           # a REPL: each input is a declaration or an expression
```

In the REPL a declaration accumulates, a name may be declared again, an import stays in force,
and an input continues on the next line while a brace, bracket, string, or template is open.

## Library

```go
prog, err := ant.Load("main.ant")             // or ant.WithSource / ant.WithFS for sources in memory
if err != nil { ... }                         // *ant.LoadError: syntax error, missing module, cycle
err = prog.Run(ctx, ant.WithArgs(args))       // nil, *ant.PanicError, or *ant.ExitError
diags := prog.Check()                         // []ant.Diagnostic
```

`WithStdin`, `WithStdout`, `WithStderr`, `WithArgs` and `WithEnv` give a run its streams, its
arguments and its environment; `Os.stdin()` and the other streams read and write them.

To call Ant from Go, start an instance instead of running main:

```go
in, err := prog.Start(ctx)
defer in.Close()
scale, _ := in.Func("scale")
result, err := scale.Call(point, 2.0)         // an Ant panic comes back as *ant.PanicError
```

To call Go from Ant, give the program a namespace declared in Ant notation:

```go
prog, err := ant.Load("main.ant", ant.WithNamespace(ant.Namespace{
    Name:       "Kv",
    Signatures: "func Kv.get(key: string): string | null",
    Funcs:      map[string]ant.HostFunc{"get": get},
}), ant.WithoutNamespaces("File", "Http"))
```

A `Session` evaluates one input at a time on a persistent module, which is what the REPL is
built on:

```go
s, err := ant.NewSession(ctx, ant.WithStdout(os.Stdout))
v, err := s.Eval("var x = 2")                 // void
v, err = s.Eval("x * 21")                     // 42, a value.Value
```

Values cross the boundary as `value.Value`: `int64`, `float64`, `string`, `bool`, `value.Null`
and `value.Void` for the primitives, and `*value.Array`, `*value.Map`, `*value.Record` and
`*value.Tagged` for the rest. The package documentation lists the mapping; `example_test.go` has
complete examples.

## License

MIT
