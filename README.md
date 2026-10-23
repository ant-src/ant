# Ant

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

## Install

```sh
go install github.com/ant-src/ant/cmd/ant@latest
```

Type-check and run:

```sh
ant check main.ant
ant run main.ant
```

## API

Ant can be used as a lib too: 

```go
prog, err := ant.Load("main.ant")
diags := prog.Check()
err = prog.Run(ctx, ant.WithArgs(args))
```

```go
in, err := prog.Start(ctx)
defer in.Close()
scale, _ := in.Func("scale")
result, err := scale.Call(point, 2.0)
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
v, err := s.Eval("var x = 2")
v, err = s.Eval("x * 21")
```

## License

[MIT](LICENSE)
