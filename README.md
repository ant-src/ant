# Ant

Static types, _goroutines_, and nothing left `undefined`.

## Example

```ts
type Found = { path: string, line: string }
type Missing = { path: string }

func firstLine(path: string): Result[string | null, Error] {
  var file = try File.open(path)
  defer { File.close(file) }
  File.readLine(file)
}

func check(path: string): Found | Missing {
  var line = firstLine(path)
  if line is Ok && line.value != null { Found{ path, line: line.value } } else { Missing{ path } }
}

var paths = ["a.txt", "b.txt", "c.txt", "notes.md"] | Array.filter({ String.endsWith(x, ".txt") })
var results: Channel[Found | Missing] = Channel.new()
for path in paths { go { Channel.send(results, check(path)) } }

for _ in paths {
  select {
    var r = Channel.receive(results) {
      if r is Found { println("${r.path}: ${r.line}") } else if r is Missing { println("${r.path}: missing") }
    }
    Time.sleep(Time.duration(seconds: 1)) { println("timeout") }
  }
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

Try things out in the REPL, which highlights the code as it is typed, completes names, members, and
arguments with Tab, and suggests the first completion after the cursor:

```sh
ant
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

Complete commands, flags, and files in the shell:

```sh
source <(ant --comp zsh)   # in ~/.zshrc, after compinit
source <(ant --comp bash)  # in ~/.bashrc
ant --comp fish | source   # in ~/.config/fish/config.fish
```

## Editor Support

Ant provides code completion, diagnostics, and navigation through the Language Server Protocol
(LSP), and debugging through the Debug Adapter Protocol (DAP).

- **[Visual Studio Code](https://github.com/ant-src/vscode-ant)**
- **[IntelliJ IDEA](https://github.com/ant-src/intellij-ant)**
- **[Neovim](https://github.com/ant-src/nvim-ant)**

## Go API

Ant can be embedded in Go applications for executing scripts, calling Ant functions, and exposing Go
functions to Ant.

**Load, check, and run a program:**

```go
prog, err := ant.Load("main.ant")
diags := prog.Check()
err = prog.Run(ctx, ant.WithArgs(args))
```

**Call Ant functions from Go:**

```go
in, err := prog.Start(ctx)
defer in.Close()

scale, _ := in.Func("scale")
result, err := scale.Call(point, 2.0)
```

**Expose Go functions to Ant** using namespaces and Ant type signatures:

```go
prog, err := ant.Load("main.ant", ant.WithNamespace(ant.Namespace{
    Name:       "Kv",
    Signatures: "func Kv.get(key: string): string | null",
    Funcs:      map[string]ant.HostFunc{"get": get},
}), ant.WithoutNamespaces("File", "Http"))
```

**Evaluate code interactively** with persistent state, using the same API as the Ant REPL:

```go
s, err := ant.NewSession(ctx, ant.WithStdout(os.Stdout))
v, err := s.Eval("var x = 2")
v, err = s.Eval("x * 21")
```

## License

[MIT](LICENSE)
