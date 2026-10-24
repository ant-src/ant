# Ant

## Features

- **HTTP server and client**, JSON, CSV, crypto, files, processes.
- **Errors as values.** `Result` plus `try` to bail out early.
- **Pipes.** `list | Array.map({ x * 2 })` and lambdas.
- **Real concurrency.** `go { ... }`, channels, `select`, and contexts with deadlines.
- **Time done right.** Time zones, dates, and durations modeled after JavaScript's Temporal.
- **Union types.** `type Shape = Circle | Square`, then `if s is Circle { s.radius }`.

## Examples

```ts
var result = [5, 2, 8, 1]
  | Array.filter({ x > 2 })
  | Array.map({ x * 10 })
  | Array.sort()
println(result) // [50, 80]
```

```ts
var messages: Channel[string] = Channel.new()
go { Channel.send(messages, "Hello from another goroutine!") }
println(Channel.receive(messages))
```

```ts
func loadConfig(path: string): Result[Json.Value, Error] {
  var text = try File.read(path)
  Json.parse(text)
}
```

```ts
func serve(): Result[void, Error] {
  var listener = try Net.listen(":8080")
  var reply = Http.Reply{ status: 200, headers: Map{}, body: "Hello!" }
  Http.serve(listener, { reply })
}
```

```ts
type Ready = { url: string }
type Pending = {}
func status(job: Ready | Pending): string {
  if job is Ready { job.url } else { "pending" }
}
```

```ts
func firstLine(path: string): Result[string | null, Error] {
  var file = try File.open(path)
  defer { File.close(file) }
  File.readLine(file)
}
```

```ts
var messages: Channel[string] = Channel.new()
select {
  var message = Channel.receive(messages) { println(message) }
  Time.sleep(Time.duration(seconds: 1)) { println("Timeout") }
}
```

```ts
var date = Time.plainDate(2024, 8, 31)
var next = Time.add(date, Time.duration(months: 1))
println(next) // 2024-09-30
```

```ts
println(String.graphemes("🇨🇭🇩🇪")) // ["🇨🇭", "🇩🇪"]
println(len(String.graphemes("👨‍👩‍👧"))) // 1
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

Format, or check the layout in CI:

```sh
ant fmt .
ant fmt --check .
```

Read the docs of the standard library, or of a module's exports:

```sh
ant doc Array.map
ant doc Time
ant doc geometry.ant distance
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
